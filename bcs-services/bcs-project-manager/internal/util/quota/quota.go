/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * 	http://opensource.org/licenses/MIT
 *
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package quota

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	nsm "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/namespace"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/proto/bcsproject"
)

// TransferToProto transfer k8s ResourceQuota to proto ResourceQuota
func TransferToProto(q *corev1.ResourceQuota) (*proto.ResourceQuota, *proto.ResourceQuota) {
	quota := &proto.ResourceQuota{}
	if quantity, ok := q.Status.Hard[corev1.ResourceLimitsCPU]; ok {
		quota.CpuLimits = quantity.String()
	}
	if quantity, ok := q.Status.Hard[corev1.ResourceRequestsCPU]; ok {
		quota.CpuRequests = quantity.String()
	}
	if quantity, ok := q.Status.Hard[corev1.ResourceLimitsMemory]; ok {
		quota.MemoryLimits = quantity.String()
	}
	if quantity, ok := q.Status.Hard[corev1.ResourceRequestsMemory]; ok {
		quota.MemoryRequests = quantity.String()
	}
	used := &proto.ResourceQuota{}
	if quantity, ok := q.Status.Used[corev1.ResourceLimitsCPU]; ok {
		used.CpuLimits = quantity.String()
	}
	if quantity, ok := q.Status.Used[corev1.ResourceRequestsCPU]; ok {
		used.CpuRequests = quantity.String()
	}
	if quantity, ok := q.Status.Used[corev1.ResourceLimitsMemory]; ok {
		used.MemoryLimits = quantity.String()
	}
	if quantity, ok := q.Status.Used[corev1.ResourceRequestsMemory]; ok {
		used.MemoryRequests = quantity.String()
	}
	return quota, used
}

// LoadFromProto load k8s ResourceQuota from proto ResourceQuota
func LoadFromProto(k8sQuota *corev1.ResourceQuota, protoQuota *proto.ResourceQuota) error {
	return load(k8sQuota, protoQuota.GetCpuLimits(), protoQuota.GetCpuRequests(),
		protoQuota.GetMemoryLimits(), protoQuota.GetMemoryRequests())
}

// LoadFromModel load k8s ResourceQuota from model ResourceQuota
func LoadFromModel(k8sQuota *corev1.ResourceQuota, modelQuota *nsm.Quota) error {
	return load(k8sQuota, modelQuota.CPULimits, modelQuota.CPURequests,
		modelQuota.MemoryLimits, modelQuota.MemoryRequests)
}

func load(quota *corev1.ResourceQuota, cpuLimits, cpuRequests, memoryLimits, memoryRequests string) error {
	if cpuLimits != "" {
		cpuLimits, err := resource.ParseQuantity(cpuLimits)
		if err != nil {
			return err
		}
		quota.Spec.Hard[corev1.ResourceLimitsCPU] = cpuLimits
	}

	if cpuRequests != "" {
		cpuRequests, err := resource.ParseQuantity(cpuRequests)
		if err != nil {
			return err
		}
		quota.Spec.Hard[corev1.ResourceRequestsCPU] = cpuRequests
	}

	if memoryLimits != "" {
		memoryLimits, err := resource.ParseQuantity(memoryLimits)
		if err != nil {
			return err
		}
		quota.Spec.Hard[corev1.ResourceLimitsMemory] = memoryLimits
	}

	if memoryRequests != "" {
		memoryRequests, err := resource.ParseQuantity(memoryRequests)
		if err != nil {
			return err
		}
		quota.Spec.Hard[corev1.ResourceRequestsMemory] = memoryRequests
	}
	return nil
}
