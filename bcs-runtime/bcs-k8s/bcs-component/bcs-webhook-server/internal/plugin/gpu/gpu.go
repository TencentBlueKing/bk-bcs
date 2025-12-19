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

// Package gpu is a plugin to inject gpu resource
package gpu

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-webhook-server/internal/pluginmanager"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-webhook-server/internal/pluginutil"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-webhook-server/internal/types"
)

const (
	pluginName          = "gpuinjector"
	pluginAnnotationKey = "task.bkbcs.tencent.com/gpu-type"
)

var (
	defaultResourceNameMap = map[corev1.ResourceName]struct{}{
		corev1.ResourceCPU:              {},
		corev1.ResourceMemory:           {},
		corev1.ResourceEphemeralStorage: {},
	}
)

func init() {
	p := &Injector{}
	pluginmanager.Register(pluginName, p)
}

// ResourceCoefficient defines gpu resource coefficient
type ResourceCoefficient struct {
	Name        corev1.ResourceName `json:"name"`
	Coefficient float64             `json:"coefficient,omitempty"`
	Value       string              `json:"value,omitempty"`
	Unit        string              `json:"unit,omitempty"`
}

// InjectInfo defines gpu inject info
type InjectInfo struct {
	ResourceList []ResourceCoefficient `json:"resourceList,omitempty"`
	Annotations  map[string]string     `json:"annotations,omitempty"`
}

// InjectorConfig defines config of gpuinjector plugin
type InjectorConfig struct {
	// map[GPUType]map[ResourceName]GPUInjectInfo
	GPUResourceMap map[string]map[corev1.ResourceName]InjectInfo `json:"resourceMap,omitempty"`
}

// Injector defines gpuinjector plugin
type Injector struct {
	conf *InjectorConfig
}

// AnnotationKey returns key of the gpuinjector plugin for hook server to identify
func (gi *Injector) AnnotationKey() string {
	return pluginAnnotationKey
}

// Init do init action, load config
func (gi *Injector) Init(configFilePath string) error {
	fileBytes, err := os.ReadFile(configFilePath)
	if err != nil {
		blog.Errorf("load config file %s failed, err %s", configFilePath, err.Error())
		return fmt.Errorf("load config file %s failed, err %s", configFilePath, err.Error())
	}
	conf := &InjectorConfig{}
	if err := json.Unmarshal(fileBytes, conf); err != nil {
		blog.Errorf("decode config %s failed, err %s", string(fileBytes), err.Error())
		return fmt.Errorf("decode config %s failed, err %s", string(fileBytes), err.Error())
	}
	gi.conf = conf
	return nil
}

// Handle do handle action, check and modify
func (gi *Injector) Handle(ar v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
	req := ar.Request
	// when the kind is not Pod, ignore hook
	if req.Kind.Kind != "Pod" {
		return &v1beta1.AdmissionResponse{Allowed: true}
	}
	if req.Operation != v1beta1.Create {
		return &v1beta1.AdmissionResponse{Allowed: true}
	}
	pod := &corev1.Pod{}
	if err := json.Unmarshal(req.Object.Raw, pod); err != nil {
		blog.Errorf("cannot decode raw object %s to pod, err %s", string(req.Object.Raw), err.Error())
		return pluginutil.ToAdmissionResponse(err)
	}

	patches, err := gi.doInject(pod)
	if err != nil {
		blog.Errorf("inject gpu failed, err %s", err.Error())
		return pluginutil.ToAdmissionResponse(err)
	}
	if len(patches) == 0 {
		return &v1beta1.AdmissionResponse{Allowed: true}
	}

	patchesBytes, err := json.Marshal(patches)
	if err != nil {
		blog.Errorf("encoding patches failed, err %s", err.Error())
		return pluginutil.ToAdmissionResponse(err)
	}
	return &v1beta1.AdmissionResponse{
		Allowed: true,
		Patch:   patchesBytes,
		PatchType: func() *v1beta1.PatchType {
			pt := v1beta1.PatchTypeJSONPatch
			return &pt
		}(),
	}
}

func (gi *Injector) doInject(pod *corev1.Pod) ([]types.PatchOperation, error) {
	gpuType, resourceName, err := gi.getGPUTypeAndResourceName(pod)
	if err != nil {
		return nil, err
	}
	if gpuType == "" || resourceName == "" {
		return nil, nil
	}

	gpuInfo, ok := gi.conf.GPUResourceMap[gpuType][resourceName]
	if !ok {
		return nil, nil
	}

	defaultResMap, extResMap, err := gi.getPerGPUResource(
		pod, resourceName, gpuInfo)
	if err != nil {
		return nil, err
	}

	patches, err := gi.generateResourcePatch(pod, resourceName, defaultResMap, extResMap)
	if err != nil {
		return nil, err
	}

	annoPatches := gi.generateAnnotationPatch(gpuInfo)
	return append(patches, annoPatches...), nil
}

func (gi *Injector) generateResourcePatch(
	pod *corev1.Pod, resourceName corev1.ResourceName,
	defaultResMap map[corev1.ResourceName]resource.Quantity,
	extResMap map[corev1.ResourceName]ResourceCoefficient) ([]types.PatchOperation, error) {

	var patches []types.PatchOperation
	for index, container := range pod.Spec.Containers {
		if !isGPUContainer(container, resourceName) {
			continue
		}
		conGPUNum := getContainerGPUNum(pod, index, resourceName)
		for k, res := range defaultResMap {
			tmpV := res.MilliValue() * conGPUNum
			tmpRes := resource.NewMilliQuantity(tmpV, res.Format)
			patches = append(patches, types.PatchOperation{
				Op: types.PatchOperationReplace,
				Path: fmt.Sprintf("/spec/containers/%d/resources/requests/%s",
					index, pluginutil.JSONKeyEscape(string(k))),
				Value: tmpRes.String(),
			})
			patches = append(patches, types.PatchOperation{
				Op: types.PatchOperationReplace,
				Path: fmt.Sprintf("/spec/containers/%d/resources/limits/%s",
					index, pluginutil.JSONKeyEscape(string(k))),
				Value: tmpRes.String(),
			})
		}
		for k, res := range extResMap {
			quantity, err := resource.ParseQuantity(
				fmt.Sprintf("%d%s", int64(res.Coefficient*float64(conGPUNum)), res.Unit))
			if err != nil {
				blog.Infof("parse quantity of %v failed, err %s", res, err.Error())
				return nil, err
			}
			patches = append(patches, types.PatchOperation{
				Op: types.PatchOperationReplace,
				Path: fmt.Sprintf("/spec/containers/%d/resources/requests/%s",
					index, pluginutil.JSONKeyEscape(string(k))),
				Value: quantity.String(),
			})
			patches = append(patches, types.PatchOperation{
				Op: types.PatchOperationReplace,
				Path: fmt.Sprintf("/spec/containers/%d/resources/limits/%s",
					index, pluginutil.JSONKeyEscape(string(k))),
				Value: quantity.String(),
			})
		}
	}

	return patches, nil
}

func (gi *Injector) generateAnnotationPatch(gpuInfo InjectInfo) []types.PatchOperation {
	patches := []types.PatchOperation{}
	for k, v := range gpuInfo.Annotations {
		patches = append(patches, types.PatchOperation{
			Op:    types.PatchOperationAdd,
			Path:  fmt.Sprintf("/metadata/annotations/%s", pluginutil.JSONKeyEscape(k)),
			Value: v,
		})
	}
	return patches
}

// Close do close action, clean
func (gi *Injector) Close() error {
	return nil
}
