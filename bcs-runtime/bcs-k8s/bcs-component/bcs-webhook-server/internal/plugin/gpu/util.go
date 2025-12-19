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

package gpu

import (
	"fmt"
	"math"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func getContainerGPUNum(c *corev1.Pod, index int, gpuResourceName corev1.ResourceName) int64 {
	if index >= len(c.Spec.Containers) {
		return 0
	}
	con := c.Spec.Containers[index]
	gpuNum := int64(0)
	for k, r := range con.Resources.Requests {
		if k == gpuResourceName {
			gpuNum = r.Value()
			break
		}
	}
	return gpuNum
}

func getGPUNum(c *corev1.Pod, gpuResourceName corev1.ResourceName) int64 {
	gpuNum := int64(0)
	for index, con := range c.Spec.Containers {
		if !isGPUContainer(con, gpuResourceName) {
			continue
		}
		gpuNum += getContainerGPUNum(c, index, gpuResourceName)
	}

	return gpuNum
}

func isGPUContainer(c corev1.Container, gpuResourceName corev1.ResourceName) bool {
	for name := range c.Resources.Requests {
		if name == gpuResourceName {
			return true
		}
	}
	return false
}

func (gi *Injector) getGPUTypeAndResourceName(pod *corev1.Pod) (string, corev1.ResourceName, error) {
	gpuType := pod.Annotations[pluginAnnotationKey]
	foundGPUType := false
	for t := range gi.conf.GPUResourceMap {
		if t == gpuType {
			foundGPUType = true
			break
		}
	}
	if !foundGPUType {
		return "", "", nil
	}

	gpuResourceKey := corev1.ResourceName("")
	gpuResourceMap := make(map[corev1.ResourceName]bool)
	for _, container := range pod.Spec.Containers {
		for reqKey := range container.Resources.Requests {
			found := false
			for _, gpuResource := range gi.conf.GPUResourceMap {
				for k := range gpuResource {
					if reqKey == k {
						gpuResourceMap[k] = true
						gpuResourceKey = k
						found = true
						break
					}
				}
				if found {
					break
				}
			}
		}
	}

	if len(gpuResourceMap) == 0 {
		return "", "", nil
	}
	if len(gpuResourceMap) > 1 {
		blog.Warnf("pod %s/%s has different gpu resource %v", pod.Namespace, pod.Name, gpuResourceMap)
		return "", "", fmt.Errorf("pod %s/%s has different gpu resource %v", pod.Namespace, pod.Name, gpuResourceMap)
	}
	return gpuType, gpuResourceKey, nil
}

func (gi *Injector) getPerGPUResource(
	pod *corev1.Pod, gpuResourceName corev1.ResourceName, gpuInfo InjectInfo) (
	map[corev1.ResourceName]resource.Quantity, map[corev1.ResourceName]ResourceCoefficient, error) {

	gpuNum := getGPUNum(pod, gpuResourceName)
	if gpuNum == 0 {
		return nil, nil, fmt.Errorf("pod %s/%s has no gpu", pod.Namespace, pod.Name)
	}
	extendedResourceMap := make(map[corev1.ResourceName]ResourceCoefficient)
	remainResourceMap := make(map[corev1.ResourceName]resource.Quantity)
	for _, resCoefficient := range gpuInfo.ResourceList {
		if _, ok := defaultResourceNameMap[resCoefficient.Name]; !ok {
			extendedResourceMap[resCoefficient.Name] = resCoefficient
			continue
		}
		quantityStr := fmt.Sprintf(
			"%d%s", int(math.Floor(resCoefficient.Coefficient*float64(gpuNum))), resCoefficient.Unit)
		quantity, err := resource.ParseQuantity(quantityStr)
		if err != nil {
			blog.Errorf("failed to parse quantity %s, err %s", quantityStr, err.Error())
			return nil, nil, err
		}
		remainResourceMap[resCoefficient.Name] = quantity
	}
	for _, container := range pod.Spec.Containers {
		if isGPUContainer(container, gpuResourceName) {
			continue
		}
		for resName, res := range container.Resources.Requests {
			remainResource, ok := remainResourceMap[resName]
			if !ok {
				continue
			}
			remainResource.Sub(res)
			remainResourceMap[resName] = remainResource
		}
	}
	for k, v := range remainResourceMap {
		// k8s 资源数值除以gpu个数
		value := v.MilliValue()
		if value < 0 {
			return nil, nil, fmt.Errorf("pod %s/%s resource %s too large", pod.Namespace, pod.Name, k)
		}
		realValue := value / gpuNum
		remainResourceMap[k] = *resource.NewMilliQuantity(realValue, v.Format)
	}
	return remainResourceMap, extendedResourceMap, nil
}
