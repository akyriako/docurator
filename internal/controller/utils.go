package controller

import (
	"fmt"

	docsv1alpha1 "github.com/akyriako/docurator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	debugLevel = 1

	SpaceLabel                     = "space-%s"
	SpaceBootstrapJobName          = "space-%s-bootstrap"
	SpaceBootstrapJobContainerName = "space-%s-bootstrap-job"
	SpaceRepoName                  = "docs-%s"
)

func getLabels(space *docsv1alpha1.Space) map[string]string {
	return map[string]string{
		"space": fmt.Sprintf(space.Name),
	}
}

func getObjectMeta(space *docsv1alpha1.Space, name *string, annotations map[string]string) metav1.ObjectMeta {
	if name == nil {
		name = &space.Name
	}

	return metav1.ObjectMeta{
		Name:        *name,
		Namespace:   space.Namespace,
		Labels:      getLabels(space),
		Annotations: annotations,
	}
}
