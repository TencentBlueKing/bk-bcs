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

package limitrange

import (
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/proto/bcsproject"
)

func TestTransferToProto(t *testing.T) {
	limitRange := &corev1.LimitRange{
		ObjectMeta: metav1.ObjectMeta{Name: "pod-resources"},
		Spec: corev1.LimitRangeSpec{Limits: []corev1.LimitRangeItem{
			{Type: corev1.LimitTypeContainer, Max: corev1.ResourceList{
				corev1.ResourceCPU: resource.MustParse("1"),
			}},
			{Type: corev1.LimitTypePod,
				Min: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("500m"),
					corev1.ResourceMemory: resource.MustParse("1Gi"),
				},
				Max: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("4"),
					corev1.ResourceMemory: resource.MustParse("8Gi"),
				}},
		}},
	}

	got := TransferToProto(limitRange)
	require.Equal(t, "pod-resources", got.GetName())
	require.Equal(t, "500m", got.GetMin().GetCpu())
	require.Equal(t, "1Gi", got.GetMin().GetMemory())
	require.Equal(t, "4", got.GetMax().GetCpu())
	require.Equal(t, "8Gi", got.GetMax().GetMemory())
	require.Nil(t, TransferToProto(&corev1.LimitRange{}))
}

func TestUpsertAndRemovePodItem(t *testing.T) {
	limitRange := &corev1.LimitRange{Spec: corev1.LimitRangeSpec{Limits: []corev1.LimitRangeItem{{
		Type: corev1.LimitTypeContainer,
		Max:  corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("1")},
	}}}}
	require.NoError(t, UpsertPodItem(limitRange,
		&proto.PodLimitRangeResource{Cpu: "500m"},
		&proto.PodLimitRangeResource{Cpu: "2", Memory: "4Gi"}))
	require.Len(t, limitRange.Spec.Limits, 2)
	require.Equal(t, corev1.LimitTypeContainer, limitRange.Spec.Limits[0].Type)
	require.Equal(t, corev1.LimitTypePod, limitRange.Spec.Limits[1].Type)

	limitRange.Spec.Limits[1].MaxLimitRequestRatio = corev1.ResourceList{
		corev1.ResourceCPU: resource.MustParse("2"),
	}
	require.NoError(t, UpsertPodItem(limitRange, nil, &proto.PodLimitRangeResource{Cpu: "3"}))
	require.Len(t, limitRange.Spec.Limits, 2)
	require.Equal(t, "3", limitRange.Spec.Limits[1].Max.Cpu().String())
	require.Equal(t, "2", limitRange.Spec.Limits[1].MaxLimitRequestRatio.Cpu().String())

	require.True(t, RemovePodItem(limitRange))
	require.Len(t, limitRange.Spec.Limits, 1)
	require.Equal(t, corev1.LimitTypeContainer, limitRange.Spec.Limits[0].Type)
	require.False(t, RemovePodItem(limitRange))
}

func TestUpsertPodItemValidation(t *testing.T) {
	tests := []struct {
		name string
		min  *proto.PodLimitRangeResource
		max  *proto.PodLimitRangeResource
	}{
		{name: "empty"},
		{name: "invalid quantity", max: &proto.PodLimitRangeResource{Cpu: "wrong"}},
		{name: "negative quantity", min: &proto.PodLimitRangeResource{Memory: "-1Gi"}},
		{name: "cpu min greater than max", min: &proto.PodLimitRangeResource{Cpu: "2"},
			max: &proto.PodLimitRangeResource{Cpu: "1"}},
		{name: "memory min greater than max", min: &proto.PodLimitRangeResource{Memory: "2Gi"},
			max: &proto.PodLimitRangeResource{Memory: "1Gi"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Error(t, UpsertPodItem(&corev1.LimitRange{}, test.min, test.max))
		})
	}
}
