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
 */

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	// Kind xxx
	Kind = "GameDeployment"
	// Plural xxx
	Plural = "gamedeployments"
	// Singular xxx
	Singular = "gamedeployment"

	// GroupName xxx
	GroupName = "tkex.tencent.com"
	// Version xxx
	Version = "v1alpha1"
)

var (
	// SchemeBuilder xxx
	SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)
	// AddToScheme xxx
	AddToScheme = SchemeBuilder.AddToScheme

	// SchemeGroupVersion xxx
	SchemeGroupVersion = schema.GroupVersion{Group: GroupName, Version: Version}
)

// Resource takes an unqualified resource and returns back a Group qualified GroupResource
func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}

// addKnownTypes adds the set of types defined in this package to the supplied scheme.
func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion,
		&HookTemplate{},
		&HookTemplateList{},
		&HookRun{},
		&HookRunList{},
	)
	metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
}
