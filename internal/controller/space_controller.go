/*
Copyright 2026.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	docsv1alpha1 "github.com/akyriako/docurator/api/v1alpha1"
)

// SpaceReconciler reconciles a Space object
type SpaceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	logger logr.Logger
}

// +kubebuilder:rbac:groups=docs.opentelekomcloud.com,resources=spaces,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=docs.opentelekomcloud.com,resources=spaces/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=docs.opentelekomcloud.com,resources=spaces/finalizers,verbs=update
// +kubebuilder:rbac:groups=batch,resources=cronjobs,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Space object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.21.0/pkg/reconcile
func (r *SpaceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.logger = logf.Log.WithValues("namespace", req.Namespace, "space", req.Name)

	var space docsv1alpha1.Space
	if err := r.Get(ctx, req.NamespacedName, &space); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	r.logger.Info("reconciling space")

	// If Gitea repo already exists mark the space as rejected

	giteaSecretName := space.Spec.GiteaSecretRef.Name
	giteaSecretNamespace := space.Spec.GiteaSecretRef.Namespace
	if giteaSecretNamespace == "" {
		giteaSecretNamespace = space.Namespace
	}

	giteaSecretObjectKey := client.ObjectKey{Namespace: giteaSecretNamespace, Name: giteaSecretName}
	var giteaSecret corev1.Secret
	if err := r.Get(ctx, giteaSecretObjectKey, &giteaSecret); err != nil {
		r.logger.Error(err, "cannot find gitea secret")
		return ctrl.Result{}, err
	}

	completed, err := r.ReconcileBootstrap(ctx, &space)
	if err != nil {
		return ctrl.Result{}, err
	}

	//
	if completed && ((space.Status.Ready != nil && *space.Status.Ready == false) || space.Status.Ready == nil) {
		if err := r.patchStatus(ctx, &space, func(status *docsv1alpha1.SpaceStatus) {
			status.Ready = ptr.To(completed)

			giteaProtocol := string(giteaSecret.Data["GITEA_PROTOCOL"])
			giteaHost := string(giteaSecret.Data["GITEA_HOST"])
			gitOwner := string(giteaSecret.Data["GIT_OWNER"])
			repoName := fmt.Sprintf(SpaceRepoName, space.Name)
			status.RepoURL = fmt.Sprintf("%s://%s/%s/%s", giteaProtocol, giteaHost, gitOwner, repoName)
		}); err != nil {
			r.logger.Error(err, "updating space status failed")
			return ctrl.Result{RequeueAfter: 15 * time.Second}, nil
		}
	}

	return ctrl.Result{RequeueAfter: 1 * time.Minute}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SpaceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&docsv1alpha1.Space{}).
		Named("space").
		Complete(r)
}
