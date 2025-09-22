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

// Package quota xxx
package quota

import (
	"strconv"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	nsm "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/namespace"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/errorx"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/proto/bcsproject"
)

// ValidateResourceQuota validate proto ResourceQuota
func ValidateResourceQuota(quota *proto.ResourceQuota) error {
	if quota == nil {
		return nil
	}
	if _, err := resource.ParseQuantity(quota.CpuLimits); err != nil {
		return errorx.NewParamErr("invalid cpu limits")
	}
	if _, err := resource.ParseQuantity(quota.CpuRequests); err != nil {
		return errorx.NewParamErr("invalid cpu requests")
	}
	if _, err := resource.ParseQuantity(quota.MemoryLimits); err != nil {
		return errorx.NewParamErr("invalid memory limits")
	}
	if _, err := resource.ParseQuantity(quota.MemoryRequests); err != nil {
		return errorx.NewParamErr("invalid memory requests")
	}
	return nil
}

// ResourceMemoryCompute resource mem calculate
func ResourceMemoryCompute(inc bool, origin string, num string) (string, error) {
	memOriginQuantity, err := resource.ParseQuantity(origin)
	if err != nil {
		return "", errorx.NewParamErr("failed to parse origin memory quantity")
	}

	memQuantity, err := resource.ParseQuantity(num)
	if err != nil {
		return "", errorx.NewParamErr("failed to parse add memory quantity")
	}

	// origin add num
	if inc {
		memOriginQuantity.Add(memQuantity)
		memInt64, _ := memOriginQuantity.AsInt64()

		return strconv.FormatInt(memInt64, 10), nil
	}

	cmp := memOriginQuantity.Cmp(memQuantity)
	if cmp == -1 {
		return "0", nil
	}
	memOriginQuantity.Sub(memQuantity)
	memInt64, _ := memOriginQuantity.AsInt64()

	return strconv.FormatInt(memInt64, 10), nil
}

// ResourceCpuCompute resource cpu calculate
func ResourceCpuCompute(inc bool, origin string, num string) (string, error) {
	cpuOriginQuantity, err := resource.ParseQuantity(origin)
	if err != nil {
		return "", errorx.NewParamErr("failed to parse origin cpu quantity")
	}

	cpuQuantity, err := resource.ParseQuantity(num)
	if err != nil {
		return "", errorx.NewParamErr("failed to parse add cpu quantity")
	}

	// origin add num
	if inc {
		cpuOriginQuantity.Add(cpuQuantity)
		cpuInt64, _ := cpuOriginQuantity.AsInt64()

		return strconv.FormatInt(cpuInt64, 10), nil
	}

	cmp := cpuOriginQuantity.Cmp(cpuQuantity)
	if cmp == -1 {
		return "0", nil
	}
	cpuOriginQuantity.Sub(cpuQuantity)
	cpuInt64, _ := cpuOriginQuantity.AsInt64()

	return strconv.FormatInt(cpuInt64, 10), nil
}

// TransferToProto transfer k8s ResourceQuota to proto ResourceQuota
func TransferToProto(q *corev1.ResourceQuota) (
	quota *proto.ResourceQuota, used *proto.ResourceQuota, cpuUseRate float32, memoryUseRate float32) {
	quota = &proto.ResourceQuota{}
	cpuLimitsQuota := q.Status.Hard[corev1.ResourceLimitsCPU]
	quota.CpuLimits = cpuLimitsQuota.String()
	cpuRequestQuota := q.Status.Hard[corev1.ResourceRequestsCPU]
	quota.CpuRequests = cpuRequestQuota.String()
	memoryLimitsQuota := q.Status.Hard[corev1.ResourceLimitsMemory]
	quota.MemoryLimits = memoryLimitsQuota.String()
	memoryRequestsQuota := q.Status.Hard[corev1.ResourceRequestsMemory]
	quota.MemoryRequests = memoryRequestsQuota.String()
	used = &proto.ResourceQuota{}
	cpuLimitsUsed := q.Status.Used[corev1.ResourceLimitsCPU]
	used.CpuLimits = cpuLimitsUsed.String()
	cpuRequestsUsed := q.Status.Used[corev1.ResourceRequestsCPU]
	used.CpuRequests = cpuRequestsUsed.String()
	memoryLimitsUsed := q.Status.Used[corev1.ResourceLimitsMemory]
	used.MemoryLimits = memoryLimitsUsed.String()
	memoryRequestsUsed := q.Status.Used[corev1.ResourceRequestsMemory]
	used.MemoryRequests = memoryRequestsUsed.String()
	if cpuLimitsQuota.AsApproximateFloat64() != 0 {
		cpuUseRate = float32(cpuLimitsUsed.AsApproximateFloat64() / cpuLimitsQuota.AsApproximateFloat64())
	}
	if memoryLimitsQuota.AsApproximateFloat64() != 0 {
		memoryUseRate = float32(memoryLimitsUsed.AsApproximateFloat64() / memoryLimitsQuota.AsApproximateFloat64())
	}
	return quota, used, cpuUseRate, memoryUseRate
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
	if quota.Spec.Hard == nil {
		quota.Spec.Hard = corev1.ResourceList{}
	}
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
