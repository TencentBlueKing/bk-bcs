/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package v1beta1

import (
	"fmt"
	"strings"

	apiextv1b1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-k8s-watch/pkg/kubefed/apis/core/common"
)

// FederatedTypeConfigSpec defines the desired state of FederatedTypeConfig.
type FederatedTypeConfigSpec struct {
	// The configuration of the target type. If not set, the pluralName and
	// groupName fields will be set from the metadata.name of this resource. The
	// kind field must be set.
	TargetType APIResource `json:"targetType"`
	// Whether or not propagation to member clusters should be enabled.
	Propagation PropagationMode `json:"propagation"`
	// Configuration for the federated type that defines (via
	// template, placement and overrides fields) how the target type
	// should appear in multiple cluster.
	FederatedType APIResource `json:"federatedType"`
	// Configuration for the status type that holds information about which type
	// holds the status of the federated resource. If not provided, the group
	// and version will default to those provided for the federated type api
	// resource.
	// +optional
	StatusType *APIResource `json:"statusType,omitempty"`
	// Whether or not Status object should be populated.
	// +optional
	StatusCollection *StatusCollectionMode `json:"statusCollection,omitempty"`
}

// APIResource defines how to configure the dynamic client for an API resource.
type APIResource struct {
	// metav1.GroupVersion is not used since the json annotation of
	// the fields enforces them as mandatory.

	// Group of the resource.
	// +optional
	Group string `json:"group,omitempty"`
	// Version of the resource.
	Version string `json:"version"`
	// Camel-cased singular name of the resource (e.g. ConfigMap)
	Kind string `json:"kind"`
	// Lower-cased plural name of the resource (e.g. configmaps).  If
	// not provided, it will be computed by lower-casing the kind and
	// suffixing an 's'.
	PluralName string `json:"pluralName"`
	// Scope of the resource.
	Scope apiextv1b1.ResourceScope `json:"scope"`
}

// PropagationMode defines the state of propagation to member clusters.
type PropagationMode string

const (
	// PropagationEnabled xxx
	PropagationEnabled PropagationMode = "Enabled"
	// PropagationDisabled xxx
	PropagationDisabled PropagationMode = "Disabled"
)

// StatusCollectionMode defines the state of status collection.
type StatusCollectionMode string

const (
	// StatusCollectionEnabled xxx
	StatusCollectionEnabled StatusCollectionMode = "Enabled"
	// StatusCollectionDisabled xxx
	StatusCollectionDisabled StatusCollectionMode = "Disabled"
)

// ControllerStatus defines the current state of the controller
type ControllerStatus string

const (
	// ControllerStatusRunning means controller is in "running" state
	ControllerStatusRunning ControllerStatus = "Running"
	// ControllerStatusNotRunning means controller is in "notrunning" state
	ControllerStatusNotRunning ControllerStatus = "NotRunning"
)

// FederatedTypeConfigStatus defines the observed state of FederatedTypeConfig
type FederatedTypeConfigStatus struct {
	// ObservedGeneration is the generation as observed by the controller consuming the FederatedTypeConfig.
	ObservedGeneration int64 `json:"observedGeneration"`
	// PropagationController tracks the status of the sync controller.
	PropagationController ControllerStatus `json:"propagationController"`
	// StatusController tracks the status of the status controller.
	// +optional
	StatusController *ControllerStatus `json:"statusController,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// FederatedTypeConfig programs KubeFed to know about a single API type - the
// "target type" - that a user wants to federate. For each target type, there is
// a corresponding FederatedType that has the following fields:
//
// - The "template" field specifies the basic definition of a federated resource
// - The "placement" field specifies the placement information for the federated
//   resource
// - The "overrides" field specifies how the target resource should vary across
//   clusters.
//
// +k8s:openapi-gen=true
// +kubebuilder:resource:path=federatedtypeconfigs,shortName=ftc
// +kubebuilder:subresource:status
// FederatedTypeConfig is the Schema for the federatedtypeconfigs API
type FederatedTypeConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec FederatedTypeConfigSpec `json:"spec"`
	// +optional
	Status FederatedTypeConfigStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// FederatedTypeConfigList contains a list of FederatedTypeConfig
type FederatedTypeConfigList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []FederatedTypeConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&FederatedTypeConfig{}, &FederatedTypeConfigList{})
}

// SetFederatedTypeConfigDefaults sets default values for a FederatedTypeConfig.
func SetFederatedTypeConfigDefaults(obj *FederatedTypeConfig) {
	// TODO(marun) will name always be populated?
	nameParts := strings.SplitN(obj.Name, ".", 2)
	targetPluralName := nameParts[0]
	setStringDefault(&obj.Spec.TargetType.PluralName, targetPluralName)
	if len(nameParts) > 1 {
		group := nameParts[1]
		setStringDefault(&obj.Spec.TargetType.Group, group)
	}
	setStringDefault(&obj.Spec.FederatedType.PluralName, PluralName(obj.Spec.FederatedType.Kind))
	if obj.Spec.StatusType != nil {
		setStringDefault(&obj.Spec.StatusType.PluralName, PluralName(obj.Spec.StatusType.Kind))
		setStringDefault(&obj.Spec.StatusType.Group, obj.Spec.FederatedType.Group)
		setStringDefault(&obj.Spec.StatusType.Version, obj.Spec.FederatedType.Version)
	}
}

// setStringDefault xxx
// GetDefaultedString returns the value if provided, and otherwise
// returns the provided default.
func setStringDefault(value *string, defaultValue string) {
	if value == nil || len(*value) > 0 {
		return
	}
	*value = defaultValue
}

// PluralName computes the plural name from the kind by
// lowercasing and suffixing with 's' or `es`.
// PluralName is used to compute the plural name of the target type.
func PluralName(kind string) string {
	lowerKind := strings.ToLower(kind)
	if strings.HasSuffix(lowerKind, "s") || strings.HasSuffix(lowerKind, "x") ||
		strings.HasSuffix(lowerKind, "ch") || strings.HasSuffix(lowerKind, "sh") ||
		strings.HasSuffix(lowerKind, "z") || strings.HasSuffix(lowerKind, "o") {
		return fmt.Sprintf("%ses", lowerKind)
	}
	if strings.HasSuffix(lowerKind, "y") {
		lowerKind = strings.TrimSuffix(lowerKind, "y")
		return fmt.Sprintf("%sies", lowerKind)
	}
	return fmt.Sprintf("%ss", lowerKind)
}

// GetObjectMeta returns the ObjectMeta for the given object.
func (f *FederatedTypeConfig) GetObjectMeta() metav1.ObjectMeta {
	return f.ObjectMeta
}

// GetTargetType returns the target type.
func (f *FederatedTypeConfig) GetTargetType() metav1.APIResource {
	return apiResourceToMeta(f.Spec.TargetType, f.GetNamespaced())
}

// GetNamespaced xxx
// TODO(font): This method should be removed from the interface in favor of
// checking the namespaced property of the appropriate APIResource (TargetType,
// FederatedType) depending on context.
// GetNamespaced returns true if the target type is namespaced.
func (f *FederatedTypeConfig) GetNamespaced() bool {
	return f.Spec.TargetType.Namespaced()
}

// GetPropagationEnabled returns whether propagation is enabled for the
func (f *FederatedTypeConfig) GetPropagationEnabled() bool {
	return f.Spec.Propagation == PropagationEnabled
}

// GetFederatedType returns the federated type.
func (f *FederatedTypeConfig) GetFederatedType() metav1.APIResource {
	return apiResourceToMeta(f.Spec.FederatedType, f.GetFederatedNamespaced())
}

// GetStatusType returns the status type for this federated type.
func (f *FederatedTypeConfig) GetStatusType() *metav1.APIResource {
	if f.Spec.StatusType == nil {
		return nil
	}
	metaAPIResource := apiResourceToMeta(*f.Spec.StatusType, f.Spec.StatusType.Namespaced())
	return &metaAPIResource
}

// GetStatusEnabled returns true if the status controller is enabled.
func (f *FederatedTypeConfig) GetStatusEnabled() bool {
	return f.Spec.StatusCollection != nil &&
		*f.Spec.StatusCollection == StatusCollectionEnabled &&
		f.Name == "services"
}

// GetFederatedNamespaced xxx
// TODO(font): This method should be removed from the interface i.e. remove
// special-case handling for namespaces, in favor of checking the namespaced
// property of the appropriate APIResource (TargetType, FederatedType)
// depending on context.
// GetFederatedNamespaced returns true if the federated type is namespaced.
func (f *FederatedTypeConfig) GetFederatedNamespaced() bool {
	// Special-case the scope of federated namespace since it will
	// hopefully be the only instance of the scope of a federated
	// type differing from the scope of its target.

	if f.IsNamespace() {
		// FederatedNamespace is namespaced to allow the control plane to run
		// with only namespace-scoped permissions e.g. to determine placement.
		return true
	}
	return f.GetNamespaced()
}

// IsNamespace returns true if the FederatedTypeConfig is for a namespace.
func (f *FederatedTypeConfig) IsNamespace() bool {
	return f.Name == common.NamespaceName
}

// Namespaced returns true if the resource is namespaced.
func (a *APIResource) Namespaced() bool {
	return a.Scope == apiextv1b1.NamespaceScoped
}

func apiResourceToMeta(apiResource APIResource, namespaced bool) metav1.APIResource {
	return metav1.APIResource{
		Group:      apiResource.Group,
		Version:    apiResource.Version,
		Kind:       apiResource.Kind,
		Name:       apiResource.PluralName,
		Namespaced: namespaced,
	}
}
