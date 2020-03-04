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

package k8s

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"bk-bcs/bcs-common/common/blog"
	"bk-bcs/bcs-services/bcs-webhook-server/pkg/inject/common/bscp"

	corev1 "k8s.io/api/core/v1"
)

// BscpInject implements BscpInject
type BscpInject struct {
	//template containers
	temContainers []corev1.Container
}

// NewBscpInject new BscpInject object
func NewBscpInject() *BscpInject {
	return &BscpInject{}
}

// InitTemplate load template from file
func (bi *BscpInject) InitTemplate(templatePath string) error {
	by, err := ioutil.ReadFile(templatePath)
	if err != nil {
		blog.Errorf("bscp load template file %s failed, err %s", templatePath, err.Error())
		return fmt.Errorf("bscp load template file %s failed, err %s", templatePath, err.Error())
	}

	err = json.Unmarshal(by, &bi.temContainers)
	if err != nil {
		//template format err, then exit
		blog.Errorf("bscp Unmarshal template file %s error %s", templatePath, err.Error())
		return fmt.Errorf("bscp Unmarshal template file %s error %s", templatePath, err.Error())
	}

	if len(bi.temContainers) == 0 {
		blog.Errorf("bscp init template %s failed, No template information found", templatePath)
		return fmt.Errorf("bscp init template %s failed, No template information found", templatePath)
	}
	return nil
}

//injectRequired check if pod injection needed
//* check annotation k/v
func (bi *BscpInject) injectRequired(pod *corev1.Pod) bool {
	if pod.Annotations == nil {
		blog.Warnf("Pod %s/%s has no annotation information.", pod.Namespace, pod.Name)
		return false
	}
	if value, ok := pod.Annotations[bscp.AnnotationKey]; !ok || value != bscp.AnnotationValue {
		blog.Warnf("Pod %s/%s has no expected annoation key & value", pod.Namespace, pod.Name)
		return false
	}
	return true
}

// InjectContent implements k8s inject interface
func (bi *BscpInject) InjectContent(pod *corev1.Pod) ([]PatchOperation, error) {

	if !bi.injectRequired(pod) {
		blog.Infof("webhookController skip Pod %s/%s sidecar injection.", pod.GetNamespace(), pod.GetName())
		return nil, nil
	}

	var patches []PatchOperation
	//check sidecar environments
	envs, patchesReplace, err := bi.retrieveEnvFromPod(pod)
	if err != nil {
		blog.Errorf("webhookController get %s/%s bscp-system Environments err, %s", pod.GetNamespace(), pod.GetName(), err.Error())
		return nil, fmt.Errorf("webhookController get %s/%s bscp-system Environments err, %s", pod.GetNamespace(), pod.GetName(), err.Error())
	}
	patchesAdd := bi.injectToPod(pod, envs)
	patches = append(patches, patchesReplace...)
	patches = append(patches, patchesAdd...)

	return patches, nil
}

func (bi *BscpInject) retrieveEnvFromPod(pod *corev1.Pod) (map[string]string, []PatchOperation, error) {
	envMap := make(map[string]string)
	var patches []PatchOperation
	for index, c := range pod.Spec.Containers {
		for _, env := range c.Env {
			// record all env with sidecar prefix
			if strings.Contains(env.Name, bscp.SideCarPrefix) {
				envMap[env.Name] = env.Value
				blog.Infof("Injection for %s [%s=%s]", pod.GetName()+"/"+pod.GetNamespace(), env.Name, env.Value)
			}
			// check sidecar config path for sharing files between containers
			if env.Name == bscp.SideCarCfgPath {
				// inject emptydir
				emptydir := corev1.Volume{
					Name: bscp.SideCarVolumeName,
				}
				patches = append(patches, PatchOperation{
					Op:    bscp.PatchOperationAdd,
					Path:  fmt.Sprintf(bscp.PatchPathVolumes, 0),
					Value: emptydir,
				})
				// inject volumeMount
				volumeMount := corev1.VolumeMount{
					Name:      bscp.SideCarVolumeName,
					ReadOnly:  false,
					MountPath: env.Value,
				}
				c.VolumeMounts = append(c.VolumeMounts, volumeMount)
				patches = append(patches, PatchOperation{
					Op:    bscp.PatchOperationReplace,
					Path:  fmt.Sprintf(bscp.PatchPathContainers, index),
					Value: c,
				})
			}
		}
	}

	cfgPath, ok := envMap[bscp.SideCarCfgPath]
	if !ok {
		return nil, nil, fmt.Errorf("bscp SideCar environment lost BSCP_BCSSIDECAR_APPCFG_PATH")
	}
	if len(cfgPath) == 0 {
		return nil, nil, fmt.Errorf("bscp SideCar environment BSCP_BCSSIDECAR_APPCFG_PATH is empty")
	}
	if modValue, ok := envMap[bscp.SideCarMod]; !ok {
		// for single app
		if !bscp.ValidateEnvs(envMap) {
			return nil, nil, fmt.Errorf("bscp sidecar envs are invalid")
		}
	} else {
		// for multiple app
		// if BSCP_BCSSIDECAR_APPINFO_MOD's value is invlaid, cannot do inject
		modValue, err := bscp.AddPathIntoAppInfoMode(modValue, cfgPath)
		if err != nil {
			return nil, nil, fmt.Errorf("env %s:%s is invalid", bscp.SideCarMod, modValue)
		}
		envMap[bscp.SideCarMod] = modValue
	}

	return envMap, patches, nil
}

// inject inject env and volume mounts into template containers
func (bi *BscpInject) injectToPod(pod *corev1.Pod, envs map[string]string) []PatchOperation {
	var patches []PatchOperation
	for _, container := range bi.temContainers {
		// inject envs
		for key, value := range envs {
			env := corev1.EnvVar{
				Name:  key,
				Value: value,
			}
			container.Env = append(container.Env, env)
			// inject volumeMount for template containers
			if key == bscp.SideCarCfgPath {
				volumeMount := corev1.VolumeMount{
					Name:      bscp.SideCarVolumeName,
					ReadOnly:  false,
					MountPath: value,
				}
				container.VolumeMounts = append(container.VolumeMounts, volumeMount)
			}
		}
		patches = append(patches, PatchOperation{
			Op:    bscp.PatchOperationAdd,
			Path:  fmt.Sprintf(bscp.PatchPathContainers, 0),
			Value: container,
		})
	}

	return patches
}
