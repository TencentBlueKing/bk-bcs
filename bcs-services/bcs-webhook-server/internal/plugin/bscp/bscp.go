/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package bscp

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webhook-server/internal/metrics"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webhook-server/internal/pluginutil"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webhook-server/internal/types"
)

// Hooker webhook for bscp
type Hooker struct {
	// template containers
	temContainers []corev1.Container
}

// AnnotationKey get annotation key for webhook plugin
func (h *Hooker) AnnotationKey() string {
	return AnnotationKey
}

// Init init webhook plugin, plugin should read config from file configFilePath
func (h *Hooker) Init(configFilePath string) error {
	fileBytes, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		blog.Errorf("load template file %s failed, err %s", configFilePath, err.Error())
		return fmt.Errorf("load template file %s failed, err %s", configFilePath, err.Error())
	}
	h.temContainers = make([]corev1.Container, 0)
	err = json.Unmarshal(fileBytes, &h.temContainers)
	if err != nil {
		blog.Errorf("decode template file %s failed, err %s", configFilePath, err.Error())
		return fmt.Errorf("decode template file %s failed, err %s", configFilePath, err.Error())
	}
	return nil
}

// Handle do hook function
func (h *Hooker) Handle(ar v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
	req := ar.Request
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
		metrics.ReportBcsWebhookServerPluginLantency(BscpPluginName, metrics.StatusFailure, started)
		return pluginutil.ToAdmissionResponse(err)
	}

	// Deal with potential empty fields, e.g., when the pod is created by a deployment
	if pod.ObjectMeta.Namespace == "" {
		pod.ObjectMeta.Namespace = req.Namespace
	}

	// do inject
	patches, err := h.createPatch(req.Name, req.Namespace, pod)
	if err != nil {
		blog.Errorf("create path failed, err %s", err.Error())
		metrics.ReportBcsWebhookServerPluginLantency(BscpPluginName, metrics.StatusFailure, started)
		return pluginutil.ToAdmissionResponse(err)
	}
	patchesBytes, err := json.Marshal(patches)
	if err != nil {
		blog.Errorf("encoding patches faile, err %s", err.Error())
		metrics.ReportBcsWebhookServerPluginLantency(BscpPluginName, metrics.StatusFailure, started)
		return pluginutil.ToAdmissionResponse(err)
	}
	reviewResponse := v1beta1.AdmissionResponse{
		Allowed: true,
		Patch:   patchesBytes,
		PatchType: func() *v1beta1.PatchType {
			pt := v1beta1.PatchTypeJSONPatch
			return &pt
		}(),
	}
	metrics.ReportBcsWebhookServerPluginLantency(BscpPluginName, metrics.StatusSuccess, started)

	return &reviewResponse
}

// check if pod injection needed
func (h *Hooker) injectRequired(pod *corev1.Pod) bool {
	if value, ok := pod.Annotations[AnnotationKey]; !ok || value != AnnotationValue {
		blog.Warnf("Pod %s/%s has no expected annoation key & value", pod.Name, pod.Namespace)
		return false
	}
	return true
}

func (h *Hooker) createPatch(reqName, reqNs string, pod *corev1.Pod) ([]types.PatchOperation, error) {
	if !h.injectRequired(pod) {
		blog.Infof("bscp hooker | skip Pod from req %s/%s sidecar injection.", reqName, reqNs)
		return nil, nil
	}

	var patches []types.PatchOperation
	// check sidecar environments
	envs, patchesReplace, err := h.retrieveEnvFromPod(pod)
	if err != nil {
		blog.Errorf("bscp hooker | get pod from req %s/%s environments failed, err %s",
			reqName, reqNs, err.Error())
		return nil, fmt.Errorf("bscp hooker | get pod from req %s/%s environments failed, err %s",
			reqName, reqNs, err.Error())
	}
	patchesAdd := h.injectToPod(pod, envs)
	patches = append(patches, patchesReplace...)
	patches = append(patches, patchesAdd...)

	return patches, nil
}

func (h *Hooker) retrieveEnvFromPod(pod *corev1.Pod) (map[string]string, []types.PatchOperation, error) {
	envMap := make(map[string]string)
	var patches []types.PatchOperation
	for index, c := range pod.Spec.Containers {
		for _, env := range c.Env {
			// record all env with sidecar prefix
			if strings.Contains(env.Name, SideCarPrefix) {
				envMap[env.Name] = env.Value
				blog.Infof("Injection [%s=%s]", env.Name, env.Value)
			}
			// check sidecar config path for sharing files between containers
			if env.Name == SideCarCfgPath {
				// inject emptydir
				emptydir := corev1.Volume{
					Name: SideCarVolumeName,
				}
				patches = append(patches, types.PatchOperation{
					Op:    PatchOperationAdd,
					Path:  fmt.Sprintf(PatchPathVolumes, 0),
					Value: emptydir,
				})
				// inject volumeMount
				volumeMount := corev1.VolumeMount{
					Name:      SideCarVolumeName,
					ReadOnly:  false,
					MountPath: env.Value,
				}
				c.VolumeMounts = append(c.VolumeMounts, volumeMount)
				patches = append(patches, types.PatchOperation{
					Op:    PatchOperationReplace,
					Path:  fmt.Sprintf(PatchPathContainers, index),
					Value: c,
				})
			}
		}
	}

	cfgPath, ok := envMap[SideCarCfgPath]
	if !ok {
		return nil, nil, fmt.Errorf("bscp SideCar environment lost BSCP_BCSSIDECAR_APPCFG_PATH")
	}
	if len(cfgPath) == 0 {
		return nil, nil, fmt.Errorf("bscp SideCar environment BSCP_BCSSIDECAR_APPCFG_PATH is empty")
	}
	if modValue, ok := envMap[SideCarMod]; !ok {
		// for single app
		if !ValidateEnvs(envMap) {
			return nil, nil, fmt.Errorf("bscp sidecar envs are invalid")
		}
	} else {
		// for multiple app
		// if BSCP_BCSSIDECAR_APPINFO_MOD's value is invlaid, cannot do inject
		modValue, err := AddPathIntoAppInfoMode(modValue, cfgPath)
		if err != nil {
			return nil, nil, fmt.Errorf("env %s:%s is invalid", SideCarMod, modValue)
		}
		envMap[SideCarMod] = modValue
	}

	return envMap, patches, nil
}

// inject inject env and volume mounts into template containers
func (h *Hooker) injectToPod(pod *corev1.Pod, envs map[string]string) []types.PatchOperation {
	var patches []types.PatchOperation
	for _, container := range h.temContainers {
		// inject envs
		for key, value := range envs {
			env := corev1.EnvVar{
				Name:  key,
				Value: value,
			}
			container.Env = append(container.Env, env)
			// inject volumeMount for template containers
			if key == SideCarCfgPath {
				volumeMount := corev1.VolumeMount{
					Name:      SideCarVolumeName,
					ReadOnly:  false,
					MountPath: value,
				}
				container.VolumeMounts = append(container.VolumeMounts, volumeMount)
			}
		}
		patches = append(patches, types.PatchOperation{
			Op:    PatchOperationAdd,
			Path:  fmt.Sprintf(PatchPathContainers, 0),
			Value: container,
		})
	}

	return patches
}

// Close called when webhook plugin exit
func (h *Hooker) Close() error {
	return nil
}
