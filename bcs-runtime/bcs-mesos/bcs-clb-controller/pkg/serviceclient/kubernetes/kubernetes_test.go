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

package kubernetes

import (
	"testing"

	k8scorev1 "k8s.io/api/core/v1"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetIndexFromStatefulSetName(t *testing.T) {
	tests := []struct {
		Name  string
		IsErr bool
		Index int
	}{
		{
			Name:  "pvpserver-0",
			IsErr: false,
			Index: 0,
		},
		{
			Name:  "pvpserver",
			IsErr: true,
			Index: -1,
		},
	}
	for _, test := range tests {
		outIndex, err := getIndexFromStatefulSetName(test.Name)
		if (test.IsErr && err == nil) || (!test.IsErr && err != nil) {
			t.Errorf("IsErr %v, but err %s", test.IsErr, err.Error())
		}
		if !test.IsErr {
			if outIndex != test.Index {
				t.Errorf("expect index %d, but get %d", test.Index, outIndex)
			}
		}
	}
}

func TestSortStatefulSetPod(t *testing.T) {
	tests := []struct {
		podsBefore []*k8scorev1.Pod
		podsAfter  []*k8scorev1.Pod
	}{
		{
			podsBefore: []*k8scorev1.Pod{
				{
					ObjectMeta: k8smetav1.ObjectMeta{
						Name: "pvp-3",
					},
				},
				{
					ObjectMeta: k8smetav1.ObjectMeta{
						Name: "pvp-2",
					},
				},
				{
					ObjectMeta: k8smetav1.ObjectMeta{
						Name: "pvp-9",
					},
				},
				{
					ObjectMeta: k8smetav1.ObjectMeta{
						Name: "pvp-110",
					},
				},
			},
			podsAfter: []*k8scorev1.Pod{
				{
					ObjectMeta: k8smetav1.ObjectMeta{
						Name: "pvp-2",
					},
				},
				{
					ObjectMeta: k8smetav1.ObjectMeta{
						Name: "pvp-3",
					},
				},
				{
					ObjectMeta: k8smetav1.ObjectMeta{
						Name: "pvp-9",
					},
				},
				{
					ObjectMeta: k8smetav1.ObjectMeta{
						Name: "pvp-110",
					},
				},
			},
		},
	}
	for _, test := range tests {
		sortStatefulSetPod(test.podsBefore)
		for index, pod := range test.podsBefore {
			if pod.Name != test.podsAfter[index].Name {
				t.Errorf("expect %v but get %v", test.podsAfter[index], pod)
			}
		}
	}
}
