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

// Package resourceratio filters invalid resource ratio
package resourceratio

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-webhook-server/internal/metrics"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-webhook-server/internal/pluginmanager"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-webhook-server/internal/pluginutil"
)

const (
	pluginName = "resourceratio"
)

func init() {
	p := &ResourceRatio{}
	pluginmanager.Register(pluginName, p)
}

// ResourceRatio is a plugin for resource ration
type ResourceRatio struct {
	// min Ratio is the ratio of cpu/memory
	MinRatio float64 `json:"minRatio"`
	// max Ratio is the ratio of cpu/memory
	MaxRatio float64 `json:"maxRatio"`
	// SkipNamespace is the namespace list that the plugin will skip
	SkipNamespace []string `json:"skipNamespace"`
	// 如果cpu 和meory 为0 是否忽略
	SkipZero bool `json:"skipZero"`
	// 如果cpu小于多少核，则忽略
	SkipCPUTotal float64 `json:"skipCPUTotal"`
	// 如果memory小于多少Gi，则忽略
	SkipMemoryTotal float64 `json:"skipMemTotal"`
}

// AnnotationKey return the annotation key
func (r *ResourceRatio) AnnotationKey() string {
	return ""
}

// Init init the plugin
func (r *ResourceRatio) Init(configFilePath string) error {
	configData, err := os.ReadFile(configFilePath)
	if err != nil {
		return err
	}
	return json.Unmarshal(configData, r)
}

// Handle handle the pod
func (r *ResourceRatio) Handle(review v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
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

	if !r.handleRequired(pod) {
		return &v1beta1.AdmissionResponse{
			Allowed: true,
			PatchType: func() *v1beta1.PatchType {
				pt := v1beta1.PatchTypeJSONPatch
				return &pt
			}(),
		}
	}
	if r.isBlock(pod) {
		return pluginutil.ToAdmissionResponse(
			fmt.Errorf("pod %s/%s 资源配比(cpu/memory) 不在区间 %.2f - %.2f", pod.GetName(), pod.GetNamespace(), r.MinRatio, r.MaxRatio))
	}

	return &v1beta1.AdmissionResponse{
		Allowed: true,
		PatchType: func() *v1beta1.PatchType {
			pt := v1beta1.PatchTypeJSONPatch
			return &pt
		}(),
	}

}

func (r *ResourceRatio) handleRequired(pod *corev1.Pod) bool {
	// pod 所属的 namespace 在 skipNamespace 中，不处理
	for _, ns := range r.SkipNamespace {
		if strings.TrimSpace(ns) == pod.Namespace {
			return false
		}
	}
	return true
}

func (r *ResourceRatio) isBlock(pod *corev1.Pod) bool {
	if pod.Spec.Containers == nil {
		return false
	}
	cpuRequest := resource.NewQuantity(0, resource.DecimalSI)
	memoryRequest := resource.NewQuantity(0, resource.BinarySI)

	allContainers := append(pod.Spec.Containers, pod.Spec.InitContainers...) // nolint
	for _, container := range allContainers {
		if container.Resources.Requests != nil {
			if cpu, ok := container.Resources.Requests[corev1.ResourceCPU]; ok {
				cpuRequest.Add(cpu)
			} else if container.Resources.Limits != nil {
				// 没有设置 request，但是设置了 limit，使用 limit 作为 request
				if cpu, ok := container.Resources.Limits[corev1.ResourceCPU]; ok {
					cpuRequest.Add(cpu)
				}
			}

			if memory, ok := container.Resources.Requests[corev1.ResourceMemory]; ok {
				memoryRequest.Add(memory)
			} else if container.Resources.Limits != nil {
				// 没有设置 request，但是设置了 limit，使用 limit 作为 request
				if memory, ok := container.Resources.Limits[corev1.ResourceMemory]; ok {
					memoryRequest.Add(memory)
				}
			}
		} else if container.Resources.Limits != nil {
			// 没有设置 request，但是设置了 limit，使用 limit 作为 request
			if cpu, ok := container.Resources.Limits[corev1.ResourceCPU]; ok {
				cpuRequest.Add(cpu)
			}
			// 没有设置 request，但是设置了 limit，使用 limit 作为 request
			if memory, ok := container.Resources.Limits[corev1.ResourceMemory]; ok {
				memoryRequest.Add(memory)
			}
		}

	}
	if cpuRequest.IsZero() || memoryRequest.IsZero() {
		if r.SkipZero {
			return false
		}
		blog.Warnf("pod %s/%s cpuRequest or memoryRequest is zero, blocked", pod.Namespace, pod.Name)
		return true
	}

	// 转化为核和Gi，再计算比例
	cpuCoreValue := float64(cpuRequest.MilliValue()) / float64(1000)
	memoryGiValue := float64(memoryRequest.Value()) / float64(1024*1024*1024)
	ratio := cpuCoreValue / memoryGiValue

	if r.SkipCPUTotal > 0 && cpuCoreValue < r.SkipCPUTotal {
		return false
	}
	if r.SkipMemoryTotal > 0 && memoryGiValue < r.SkipMemoryTotal {
		return false
	}
	if ratio > r.MaxRatio || ratio < r.MinRatio {
		blog.Infof("pod %s/%s cpuRequest %s memoryRequest %s ratio %.2f not in %.2f-%.2f, blocked",
			pod.Namespace, pod.Name, cpuRequest.String(), memoryRequest.String(),
			ratio, r.MinRatio, r.MaxRatio)
		return true
	}

	return false
}

// Close close the plugin
func (r *ResourceRatio) Close() error {
	return nil
}
