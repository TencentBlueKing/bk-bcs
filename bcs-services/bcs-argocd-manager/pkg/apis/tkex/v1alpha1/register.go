/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package v1alpha1

import (
	"github.com/Tencent/bk-bcs/bcs-services/bcs-argocd-manager/pkg/apis/tkex"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// SchemeGroupVersion xxx
var SchemeGroupVersion = schema.GroupVersion{
	Group:   tkex.GroupName,
	Version: tkex.Version,
}

var (
	// SchemeBuilder xxx
	SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)
	// AddToScheme xxx
	AddToScheme = SchemeBuilder.AddToScheme
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
	metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
}
