package controller

import (
	"context"

	docsv1alpha1 "github.com/akyriako/docurator/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *SpaceReconciler) patchStatus(
	ctx context.Context,
	space *docsv1alpha1.Space,
	patcher func(status *docsv1alpha1.SpaceStatus),
) error {
	patch := client.MergeFrom(space.DeepCopy())
	patcher(&space.Status)

	err := r.Status().Patch(ctx, space, patch)
	if err != nil {
		r.logger.Error(err, "unable to patch space status")
		return err
	}

	return nil
}
