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

package utils

import (
	"fmt"
	"testing"

	bcsmcsv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-mcs/pkg/apis/mcs/v1alpha1"
	discoveryv1beta1 "k8s.io/api/discovery/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestFindNeedDeleteManifest(t *testing.T) {
	manifestList := &bcsmcsv1alpha1.ManifestList{
		Items: []bcsmcsv1alpha1.Manifest{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "endpointslices.test.test1",
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test2",
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test3",
				},
			},
		},
	}
	endpointSliceList := &discoveryv1beta1.EndpointSliceList{
		Items: []discoveryv1beta1.EndpointSlice{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test1",
					Namespace: "test",
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test2",
					Namespace: "test",
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test3",
					Namespace: "test",
				},
			},
		},
	}
	manifestListNeedDelete := FindNeedDeleteManifest(manifestList, endpointSliceList)
	fmt.Print(manifestListNeedDelete)
}
