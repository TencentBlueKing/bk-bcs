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

// Package patchmount xxx
package patchmount

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-webhook-server/internal/metrics"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-webhook-server/internal/pluginmanager"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-webhook-server/internal/pluginutil"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-webhook-server/internal/types"
)

func init() {
	p := &PatchMount{}
	pluginmanager.Register(pluginName, p)
}

// PatchMount is a plugin for patching pod volume mount
type PatchMount struct {
}

// AnnotationKey is the key of annotation for patching
func (p *PatchMount) AnnotationKey() string {
	return pluginAnnotationKey
}

// Init initialize the plugin
func (p *PatchMount) Init(configFilePath string) error {
	return nil
}

// Handle handles the admission request
func (p *PatchMount) Handle(review v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
	req := review.Request
	// when the kind is not Pod, ignore hook
	if req.Kind.Kind != "Pod" {
		return &v1beta1.AdmissionResponse{Allowed: true}
	}
	if req.Operation != v1beta1.Create {
		return &v1beta1.AdmissionResponse{Allowed: true}
	}
	started := time.Now()
	pod := &corev1.Pod{}
	if err := json.Unmarshal(req.Object.Raw, pod); err != nil {
		blog.Errorf("cannot decode raw object %s to pod, err %s", string(req.Object.Raw), err.Error())
		metrics.ReportBcsWebhookServerPluginLantency(pluginName, metrics.StatusFailure, started)
		return pluginutil.ToAdmissionResponse(err)
	}
	// Deal with potential empty fileds, e.g., when the pod is created by a deployment
	if pod.ObjectMeta.Namespace == "" {
		pod.ObjectMeta.Namespace = req.Namespace
	}
	if !p.injectRequired(pod) {
		return &v1beta1.AdmissionResponse{
			Allowed: true,
			PatchType: func() *v1beta1.PatchType {
				pt := v1beta1.PatchTypeJSONPatch
				return &pt
			}(),
		}
	}
	patches, err := p.injectToPod(pod, pod.Annotations[pluginAnnotationKey])
	if err != nil {
		blog.Errorf("inject to pod %s/%s failed, err %s", pod.GetName(), pod.GetNamespace(), err.Error())
		metrics.ReportBcsWebhookServerPluginLantency(pluginName, metrics.StatusFailure, started)
		return pluginutil.ToAdmissionResponse(err)
	}
	patchesBytes, err := json.Marshal(patches)
	if err != nil {
		blog.Errorf("encoding patches failed, err %s", err.Error())
		metrics.ReportBcsWebhookServerPluginLantency(pluginName, metrics.StatusFailure, started)
		return pluginutil.ToAdmissionResponse(err)
	}

	metrics.ReportBcsWebhookServerPluginLantency(pluginName, metrics.StatusSuccess, started)
	return &v1beta1.AdmissionResponse{
		Allowed: true,
		Patch:   patchesBytes,
		PatchType: func() *v1beta1.PatchType {
			pt := v1beta1.PatchTypeJSONPatch
			return &pt
		}(),
	}
}

// injectRequired xxx
// check if pod injection needed
func (p *PatchMount) injectRequired(pod *corev1.Pod) bool {
	value := pod.Annotations[pluginAnnotationKey]
	for _, v := range []string{patchMountCgroupfs, patchMountLxcfs} {
		if v == value {
			return true
		}
	}
	blog.Warnf("Pod %s/%s has no expected annoation key & value", pod.Name, pod.Namespace)
	return false
}

func (p *PatchMount) injectToPod(pod *corev1.Pod, patchMountType string) ([]types.PatchOperation, error) {
	volumeMountsTemplate, volumesTemplate, err := getMountTypesTemplate(pod, patchMountType)
	if err != nil {
		return nil, err
	}
	var patches []types.PatchOperation

	containers := pod.Spec.Containers
	for i := range containers {
		if containers[i].VolumeMounts == nil {
			path := fmt.Sprintf("/spec/containers/%d/volumeMounts", i)
			op := types.PatchOperation{
				Op:    "add",
				Path:  path,
				Value: volumeMountsTemplate,
			}
			patches = append(patches, op)
		} else {
			path := fmt.Sprintf("/spec/containers/%d/volumeMounts/-", i)
			for _, volumeMount := range volumeMountsTemplate {
				op := types.PatchOperation{
					Op:    "add",
					Path:  path,
					Value: volumeMount,
				}
				patches = append(patches, op)
			}
		}
	}

	if pod.Spec.Volumes == nil {
		op := types.PatchOperation{
			Op:    "add",
			Path:  "/spec/volumes",
			Value: volumesTemplate,
		}
		patches = append(patches, op)
	} else {
		for _, volume := range volumesTemplate {
			op := types.PatchOperation{
				Op:    "add",
				Path:  "/spec/volumes/-",
				Value: volume,
			}
			patches = append(patches, op)
		}
	}
	return patches, nil
}

// Close closes the plugin
func (p *PatchMount) Close() error {
	return nil
}

func getMountTypesTemplate(pod *corev1.Pod, patchMountType string) ([]corev1.VolumeMount, []corev1.Volume, error) {
	var volumeMountsTemplate []corev1.VolumeMount
	var volumesTemplate []corev1.Volume

	disableMountSysDevices := pod.Annotations[disableMountSysDevicesAnnotationKey] == "true"
	// nolint
	switch patchMountType {
	case patchMountLxcfs:
		if disableMountSysDevices {
			volumeMountsTemplate = lxcfsVolumeMountsTemplate
			volumesTemplate = lxcfsVolumesTemplate
		} else {
			volumeMountsTemplate = append(lxcfsVolumeMountsTemplate, lxcfsVolumeMountsTemplateSysDevices)
			volumesTemplate = append(lxcfsVolumesTemplate, lxcfsVolumesTemplateSysDevices)
		}
	case patchMountCgroupfs:
		if disableMountSysDevices {
			volumeMountsTemplate = cgroupfsVolumeMountsTemplate
			volumesTemplate = cgroupfsVolumesTemplate
		} else {
			volumeMountsTemplate = append(cgroupfsVolumeMountsTemplate, cgroupfsVolumeMountsTemplateSysDevices)
			fmt.Println(len(volumeMountsTemplate))
			volumesTemplate = append(cgroupfsVolumesTemplate, cgroupfsVolumesTemplateSysDevices)
		}
	default:
		return nil, nil, fmt.Errorf("unknown patch mount type %s", patchMountType)
	}
	return volumeMountsTemplate, volumesTemplate, nil
}
