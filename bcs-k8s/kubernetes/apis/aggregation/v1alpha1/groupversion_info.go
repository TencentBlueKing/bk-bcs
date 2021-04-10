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

// Package v1 contains API Schema definitions for the monitor v1 API group
// +kubebuilder:object:generate=true
// +groupName=aggregation.federated.bkbcs.tencent.com
package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	// GroupVersion is group version used to register these objects
	GroupVersion = schema.GroupVersion{Group: "aggregation.federated.bkbcs.tencent.com", Version: "v1alpha1"}

	// AddToScheme adds the types in this group-version to the given scheme.
	AddToScheme = (&runtime.SchemeBuilder{
		func(scheme *runtime.Scheme) error {
			metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
			return nil
		},
	}).AddToScheme
)

var (
	SchemeGroupVersion = GroupVersion
)

// Kind takes an unqualified kind and returns a Group qualified GroupKind
func Kind(kind string) schema.GroupKind {
	return SchemeGroupVersion.WithKind(kind).GroupKind()
}

// Resource is required by pkg/client/listers/...
func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}