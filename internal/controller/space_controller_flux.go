package controller

import (
	"context"
	"fmt"
	"time"

	docsv1alpha1 "github.com/akyriako/docurator/api/v1alpha1"
	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1"
	sourcev1 "github.com/fluxcd/source-controller/api/v1"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *SpaceReconciler) ReconcileFluxArtifacts(ctx context.Context, space *docsv1alpha1.Space, gitRepoUrl string) (*sourcev1.GitRepository, error) {
	r.logger.V(debugLevel).Info("reconciling flux artifacts")
	gitRepository, err := r.reconcileGitRepository(ctx, space, gitRepoUrl)
	if err != nil {
		return nil, err
	}

	_, err = r.reconcileKustomization(ctx, space, gitRepository)
	if err != nil {
		return nil, err
	}

	return gitRepository, nil
}

func (r *SpaceReconciler) reconcileGitRepository(ctx context.Context, space *docsv1alpha1.Space, gitRepoUrl string) (*sourcev1.GitRepository, error) {
	gitRepositoryName := fmt.Sprintf(SpaceLabel, space.Name)
	gitRepositoryExists := true
	gitRepositoryObjectKey := client.ObjectKey{Namespace: space.Namespace, Name: gitRepositoryName}

	r.logger.V(debugLevel).Info("reconciling git repository", "job", gitRepositoryName)

	gitRepository := &sourcev1.GitRepository{}
	if err := r.Get(ctx, gitRepositoryObjectKey, gitRepository); err != nil {
		if apierrors.IsNotFound(err) {
			gitRepositoryExists = false
		} else {
			r.logger.Error(err, "unable to git repository")
			return nil, err
		}
	}

	if gitRepositoryExists {
		return gitRepository, nil
	}

	gitRepository, err := r.createGitRepository(ctx, gitRepositoryObjectKey, space, gitRepoUrl)
	if err != nil {
		return nil, err
	}

	return gitRepository, nil
}

func (r *SpaceReconciler) createGitRepository(ctx context.Context, key client.ObjectKey, space *docsv1alpha1.Space, gitRepoUrl string) (*sourcev1.GitRepository, error) {
	gitRepository := &sourcev1.GitRepository{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: getObjectMeta(space, &key.Name, nil),
		Spec: sourcev1.GitRepositorySpec{
			Interval: metav1.Duration{
				Duration: 1 * time.Minute,
			},
			URL: gitRepoUrl,
			Reference: &sourcev1.GitRepositoryRef{
				Branch: "main",
			},
		},
	}

	err := ctrl.SetControllerReference(space, gitRepository, r.Scheme)
	if err != nil {
		return nil, err
	}

	err = r.Create(ctx, gitRepository)
	if err != nil {
		return nil, err
	}

	return gitRepository, nil
}

func (r *SpaceReconciler) reconcileKustomization(ctx context.Context, space *docsv1alpha1.Space, gitRepository *sourcev1.GitRepository) (*kustomizev1.Kustomization, error) {
	kustomizationName := fmt.Sprintf(SpaceLabel, space.Name)
	kustomizationExists := true
	kustomizationObjectKey := client.ObjectKey{Namespace: space.Namespace, Name: kustomizationName}

	r.logger.V(debugLevel).Info("reconciling kustomization", "job", kustomizationName)

	kustomization := &kustomizev1.Kustomization{}
	if err := r.Get(ctx, kustomizationObjectKey, kustomization); err != nil {
		if apierrors.IsNotFound(err) {
			kustomizationExists = false
		} else {
			r.logger.Error(err, "unable to get kustomization")
			return nil, err
		}
	}

	if kustomizationExists {
		return kustomization, nil
	}

	kustomization, err := r.createKustomization(ctx, kustomizationObjectKey, space, gitRepository)
	if err != nil {
		return nil, err
	}

	return kustomization, nil
}

func (r *SpaceReconciler) createKustomization(ctx context.Context, key client.ObjectKey, space *docsv1alpha1.Space, gitRepository *sourcev1.GitRepository) (*kustomizev1.Kustomization, error) {
	kustomization := &kustomizev1.Kustomization{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: getObjectMeta(space, &key.Name, nil),
		Spec: kustomizev1.KustomizationSpec{
			SourceRef: kustomizev1.CrossNamespaceSourceReference{
				Kind: "GitRepository",
				Name: gitRepository.Name,
			},
			Path: "./deploy/manifests",
		},
	}

	err := ctrl.SetControllerReference(gitRepository, kustomization, r.Scheme)
	if err != nil {
		return nil, err
	}

	err = r.Create(ctx, kustomization)
	if err != nil {
		return nil, err
	}

	return kustomization, nil
}
