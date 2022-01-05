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

package kube

import (
	"context"
	"fmt"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/bcs-hook/apis/tkex/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/restmapper"
)

func (p *Provider) getGroupVersionResource(args []v1alpha1.Argument) (*schema.GroupVersionResource, error) {
	var group, version, kind string
	for _, arg := range args {
		switch arg.Name {
		case "Group":
			group = *arg.Value
		case "Version":
			version = *arg.Value
		case "Kind":
			kind = *arg.Value
		}
	}
	// convert GVK to GVR
	var gvk schema.GroupVersionKind
	if group == "" && version == "" && kind == "" {
		// GVK are all nil, the resource is pod
		gvk = schema.GroupVersionKind{
			Group:   "",
			Version: "v1",
			Kind:    "Pod",
		}

	} else {
		gvk = schema.GroupVersionKind{
			Group:   group,
			Version: version,
			Kind:    kind,
		}
	}
	if !p.cachedClient.Fresh() {
		p.cachedClient.Invalidate()
	}
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(p.cachedClient)
	mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return nil, fmt.Errorf(
			"Failed to get GVR from Group:%s, Version:%s, Kind:%s. Error: %v",
			gvk.Group, gvk.Version, gvk.Kind, err)
	}
	return &mapping.Resource, nil
}

func (p *Provider) getDynamicResource(gvr *schema.GroupVersionResource,
	args []v1alpha1.Argument) (dynamic.ResourceInterface, string, error) {
	var ns, name string
	for _, arg := range args {
		switch arg.Name {
		case "PodName":
			name = *arg.Value
		case "PodNamespace":
			ns = *arg.Value
		}
	}
	if ns == "" {
		return nil, "", fmt.Errorf("Namespace must be provided")
	}
	if name == "" {
		return nil, "", fmt.Errorf("Name must be provided")
	}
	dr := p.dynamicClient.Resource(*gvr).Namespace(ns)
	return dr, name, nil
}

func (p *Provider) handleFunction(dr dynamic.ResourceInterface, name string, metric *v1alpha1.KubernetesMetric) error {
	fields := metric.Fields
	function := metric.Function
	cnt := 0
	for _, field := range fields {
		err := p.handle(dr, name, function, field)
		if err != nil {
			return err
		}
		cnt++
	}
	if cnt == 0 {
		return fmt.Errorf("Fields are empty")
	}
	return nil
}

func (p *Provider) handle(dr dynamic.ResourceInterface, name, function string, field v1alpha1.Field) error {
	res, err := dr.Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("Failed to get Kubernetes object %s. Error: %v", name, err)
	}
	if res == nil {
		return fmt.Errorf("Kubernetes object %s is not exist", name)
	}

	paths := strings.Split(field.Path, ".")
	if len(paths) < 3 {
		return fmt.Errorf("Path only supports format: <metadata.annotations.xxx>. Now is %s", field.Path)
	}
	if paths[0] != "metadata" || paths[1] != "annotations" {
		return fmt.Errorf("Path only supports format: <metadata.annotations.xxx>. Now is %s", field.Path)
	}

	switch function {
	case FunctionTypeGet:
		data, found, getErr := unstructured.NestedString(res.Object, paths[0], paths[1], strings.Join(paths[2:], "."))
		if getErr != nil {
			return fmt.Errorf("Get failed. Path: %s. Error: %v", field.Path, getErr)
		}
		if !found {
			return fmt.Errorf("Get Failed. Path: %s is not exist", field.Path)
		}
		if data != field.Value {
			return fmt.Errorf("Get failed. Path: %s, want: %s, now: %s", field.Path, field.Value, data)
		}
		return nil
	case FunctionTypePatch:
		key := strings.Replace(strings.Join(paths[2:], "."), "/", "~1", -1)
		patchData := []byte(fmt.Sprintf("[{ \"op\": \"add\", \"path\": \"/%s/%s/%s\", \"value\": \"%s\"}]",
			paths[0], paths[1], key, field.Value))
		_, err = dr.Patch(context.TODO(), name, types.JSONPatchType, patchData, metav1.PatchOptions{})
		if err != nil {
			return fmt.Errorf("Failed to patch %s. PatchData: %s, Error: %v", name, string(patchData), err)
		}
		return nil
	default:
		return fmt.Errorf("Function %s is not supported by kubernetes provider", function)
	}
}
