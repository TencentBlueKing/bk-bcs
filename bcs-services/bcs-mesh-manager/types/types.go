package types

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type IstioOperator struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   IstioOperatorSpec   `json:"spec,omitempty"`
}

type IstioOperatorSpec struct {
	Profile ProfileType `json:"profile,omitempty"`
}

type ProfileType string

const (
	ProfileTypeDefault ProfileType = "default"
)

const (
	IstioOperatorKind string = "IstioOperator"
	IstioOperatorGroup string = "install.istio.io"
	IstioOperatorVersion string = "v1alpha1"
	IstioOperatorName string = "istiocontrolplane"
	IstioOperatorNamespace string = "istio-system"
	IstioOperatorPlural string = "istiooperators"
	IstioOperatorListKind string = "IstioOperatorList"
)