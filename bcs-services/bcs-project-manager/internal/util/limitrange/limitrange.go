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

// Package limitrange provides conversions and validation for Pod LimitRange items.
package limitrange

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/errorx"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/proto/bcsproject"
)

// TransferToProto converts the Pod item in a Kubernetes LimitRange to the API representation.
// A LimitRange without a Pod item is not exposed by this feature.
func TransferToProto(limitRange *corev1.LimitRange) *proto.PodLimitRange {
	if limitRange == nil {
		return nil
	}
	for _, item := range limitRange.Spec.Limits {
		if item.Type != corev1.LimitTypePod {
			continue
		}
		return &proto.PodLimitRange{
			Name: limitRange.GetName(),
			Min:  resourceListToProto(item.Min),
			Max:  resourceListToProto(item.Max),
		}
	}
	return nil
}

// UpsertPodItem validates and replaces or appends the Pod item while preserving all other items.
func UpsertPodItem(limitRange *corev1.LimitRange, min, max *proto.PodLimitRangeResource) error {
	item, err := buildPodItem(min, max)
	if err != nil {
		return err
	}
	for index := range limitRange.Spec.Limits {
		if limitRange.Spec.Limits[index].Type == corev1.LimitTypePod {
			limitRange.Spec.Limits[index].Min = item.Min
			limitRange.Spec.Limits[index].Max = item.Max
			return nil
		}
	}
	limitRange.Spec.Limits = append(limitRange.Spec.Limits, item)
	return nil
}

// RemovePodItem removes the Pod item and reports whether one was present.
func RemovePodItem(limitRange *corev1.LimitRange) bool {
	for index := range limitRange.Spec.Limits {
		if limitRange.Spec.Limits[index].Type != corev1.LimitTypePod {
			continue
		}
		limitRange.Spec.Limits = append(limitRange.Spec.Limits[:index], limitRange.Spec.Limits[index+1:]...)
		return true
	}
	return false
}

func buildPodItem(min, max *proto.PodLimitRangeResource) (corev1.LimitRangeItem, error) {
	minResources, err := protoToResourceList("min", min)
	if err != nil {
		return corev1.LimitRangeItem{}, err
	}
	maxResources, err := protoToResourceList("max", max)
	if err != nil {
		return corev1.LimitRangeItem{}, err
	}
	if len(minResources) == 0 && len(maxResources) == 0 {
		return corev1.LimitRangeItem{}, errorx.NewParamErr("pod limit range resources cannot all be empty")
	}
	if err = validateMinMax(corev1.ResourceCPU, minResources, maxResources); err != nil {
		return corev1.LimitRangeItem{}, err
	}
	if err = validateMinMax(corev1.ResourceMemory, minResources, maxResources); err != nil {
		return corev1.LimitRangeItem{}, err
	}
	return corev1.LimitRangeItem{Type: corev1.LimitTypePod, Min: minResources, Max: maxResources}, nil
}

func protoToResourceList(field string, value *proto.PodLimitRangeResource) (corev1.ResourceList, error) {
	result := corev1.ResourceList{}
	if value == nil {
		return result, nil
	}
	if err := addQuantity(result, corev1.ResourceCPU, field+".cpu", value.GetCpu()); err != nil {
		return nil, err
	}
	if err := addQuantity(result, corev1.ResourceMemory, field+".memory", value.GetMemory()); err != nil {
		return nil, err
	}
	return result, nil
}

func addQuantity(resources corev1.ResourceList, name corev1.ResourceName, field, value string) error {
	if value == "" {
		return nil
	}
	quantity, err := resource.ParseQuantity(value)
	if err != nil {
		return errorx.NewParamErr(fmt.Sprintf("invalid %s", field))
	}
	if quantity.Sign() < 0 {
		return errorx.NewParamErr(fmt.Sprintf("%s cannot be negative", field))
	}
	resources[name] = quantity
	return nil
}

func validateMinMax(name corev1.ResourceName, min, max corev1.ResourceList) error {
	minQuantity, hasMin := min[name]
	maxQuantity, hasMax := max[name]
	if hasMin && hasMax && minQuantity.Cmp(maxQuantity) > 0 {
		return errorx.NewParamErr(fmt.Sprintf("min.%s cannot be greater than max.%s", name, name))
	}
	return nil
}

func resourceListToProto(resources corev1.ResourceList) *proto.PodLimitRangeResource {
	result := &proto.PodLimitRangeResource{}
	if cpu, ok := resources[corev1.ResourceCPU]; ok {
		result.Cpu = cpu.String()
	}
	if memory, ok := resources[corev1.ResourceMemory]; ok {
		result.Memory = memory.String()
	}
	return result
}
