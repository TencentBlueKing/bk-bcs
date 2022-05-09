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

package utils

import (
	bcsmcsv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-mcs/pkg/apis/mcs/v1alpha1"
	discoveryv1beta1 "k8s.io/api/discovery/v1beta1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
)

// GenManifestName generates manifest name
func GenManifestName(resourceType string, namespace, name string) string {
	if namespace != "" {
		return resourceType + "." + namespace + "." + name
	}
	return resourceType + "." + name
}

// GenManifestNamespace	 generates manifest namespace
func GenManifestNamespace(agentID string) string {
	return "bcs-mcs-" + agentID
}

// FindNeedDeleteManifest finds need delete manifest
func FindNeedDeleteManifest(manifestList *bcsmcsv1alpha1.ManifestList, endpointSliceList *discoveryv1beta1.EndpointSliceList) []*bcsmcsv1alpha1.Manifest {
	if manifestList == nil {
		return nil
	}
	if len(manifestList.Items) == 0 {
		return nil
	}
	toDeleteManifest := make([]*bcsmcsv1alpha1.Manifest, 0)
	for _, manifest := range manifestList.Items {
		find := false
	innerLoop:
		for _, endpointSlice := range endpointSliceList.Items {
			manifestName := GenManifestName(EndpointsSliceResourceName, endpointSlice.Namespace, endpointSlice.Name)
			if manifest.Name == manifestName {
				find = true
				break innerLoop
			}
		}
		if !find {
			toDeleteManifest = append(toDeleteManifest, &manifest)
		}
	}
	return toDeleteManifest
}

//UnmarshalEndpointSlice 解析出EndpointSlice
func UnmarshalEndpointSlice(manifest *bcsmcsv1alpha1.Manifest) (*discoveryv1beta1.EndpointSlice, error) {
	unstructObj := &unstructured.Unstructured{}
	if err := unstructObj.UnmarshalJSON(manifest.Template.Raw); err != nil {
		klog.ErrorS(err, "Failed to unmarshal manifest", "namespace", manifest.Namespace, "name", manifest.Name)
		return nil, err
	}
	unstructObj.SetLabels(manifest.Labels)
	unstructObj.SetAnnotations(manifest.Annotations)

	typedObj := &discoveryv1beta1.EndpointSlice{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstructObj.UnstructuredContent(), typedObj); err != nil {
		klog.ErrorS(err, "Failed to convert unstructured to EndpointSlice", "namespace", manifest.Namespace, "name", manifest.Name)
		return nil, err
	}
	return typedObj, nil
}
