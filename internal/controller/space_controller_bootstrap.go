package controller

import (
	"context"
	"fmt"

	docsv1alpha1 "github.com/akyriako/docurator/api/v1alpha1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *SpaceReconciler) ReconcileBootstrap(ctx context.Context, space *docsv1alpha1.Space) (bool, error) {
	bootstrapJobName := fmt.Sprintf(SpaceBootstrapJobName, space.Name)
	bootstrapJobExists := true
	bootstrapJobObjectKey := client.ObjectKey{Namespace: space.Namespace, Name: bootstrapJobName}

	r.logger.V(debugLevel).Info("reconciling bootstrap job", "job", bootstrapJobName)

	bootstrapJob := &batchv1.Job{}
	if err := r.Get(ctx, bootstrapJobObjectKey, bootstrapJob); err != nil {
		if apierrors.IsNotFound(err) {
			bootstrapJobExists = false
		} else {
			r.logger.Error(err, "unable to get bootstrap job")
			return false, err
		}
	}

	if bootstrapJobExists {
		completed := r.isJobComplete(bootstrapJob)
		return completed, nil
	}

	err := r.createJob(ctx, bootstrapJobObjectKey, space)
	if err != nil {
		return false, err
	}

	return false, nil
}

func (r *SpaceReconciler) createJob(ctx context.Context, key client.ObjectKey, space *docsv1alpha1.Space) error {
	bootstrapper := &batchv1.Job{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "batch/v1",
			Kind:       "Job",
		},
		ObjectMeta: getObjectMeta(space, &key.Name, nil),
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					ImagePullSecrets: space.Spec.ImagePullSecrets,
					RestartPolicy:    corev1.RestartPolicyNever,
					Containers: []corev1.Container{
						{
							Name:  fmt.Sprintf(SpaceBootstrapJobContainerName, space.Name),
							Image: space.Spec.BootstrapImage,
							Env: []corev1.EnvVar{
								{Name: "TEMPLATE", Value: "classic"},
								{Name: "WORKDIR", Value: "/tmp/work"},
								{Name: "DEFAULT_BRANCH", Value: "main"},
								{Name: "CREATE_UNDER_ORG", Value: "true"},
								{Name: "SITE_NAME", Value: space.Name},
								{Name: "REPO_NAME", Value: fmt.Sprintf(SpaceRepoName, space.Name)},
								{Name: "META_NAMESPACE", Value: space.Namespace},
								{Name: "REGISTRY", Value: "docker.io"},
								{Name: "REGISTRY_ORG", Value: "akyriako78"},
								{Name: "EXPOSE_URL", Value: space.Spec.Ingress.Host},
								{Name: "INGRESS_CLASS_NAME", Value: space.Spec.Ingress.IngressClassName},
							},
							EnvFrom: []corev1.EnvFromSource{
								{
									SecretRef: &corev1.SecretEnvSource{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: space.Spec.GiteaSecretRef.Name,
										},
									},
								},
							},
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("1000m"),
									corev1.ResourceMemory: resource.MustParse("1024Mi"),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("128m"),
									corev1.ResourceMemory: resource.MustParse("64Mi"),
								},
							},
						},
					},
				},
			},
		},
	}

	err := ctrl.SetControllerReference(space, bootstrapper, r.Scheme)
	if err != nil {
		return err
	}

	err = r.Create(ctx, bootstrapper)
	if err != nil {
		return err
	}

	return nil
}

func (r *SpaceReconciler) isJobComplete(job *batchv1.Job) bool {
	for _, cond := range job.Status.Conditions {
		if cond.Type == batchv1.JobComplete && cond.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}
