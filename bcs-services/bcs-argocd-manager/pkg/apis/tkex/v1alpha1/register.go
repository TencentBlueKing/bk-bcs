package v1alpha1

import (
	"github.com/Tencent/bk-bcs/bcs-services/bcs-argocd-manager/pkg/apis/tkex"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var SchemeGroupVersion = schema.GroupVersion{
	Group:   tkex.GroupName,
	Version: tkex.Version,
}

var (
	SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)
	AddToScheme   = SchemeBuilder.AddToScheme
)

// Resource returns GroupResource of a resource
func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}

// Kind returns GroupKind of a kind
func Kind(kind string) schema.GroupKind {
	return SchemeGroupVersion.WithKind(kind).GroupKind()
}

func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(
		SchemeGroupVersion,
		&ArgocdInstance{},
		&ArgocdInstanceList{},
		&ArgocdProject{},
		&ArgocdProjectList{},
		&ArgocdPlugin{},
		&ArgocdPluginList{},
	)
	return nil
}
