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

package generator

import (
	"testing"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/constant"
)

func TestMatchLbStrWithID(t *testing.T) {
	testCases := []struct {
		cloud   string
		lbID    string
		isValid bool
	}{
		{
			cloud:   constant.CloudTencent,
			lbID:    "lb-123",
			isValid: true,
		},
		{
			cloud:   constant.CloudTencent,
			lbID:    "123",
			isValid: false,
		},
		{
			cloud:   constant.CloudTencent,
			lbID:    "ap-guangzhou:lb-213",
			isValid: true,
		},
		{
			cloud:   constant.CloudTencent,
			lbID:    "ap-guangzhou:123",
			isValid: false,
		},
		{
			cloud:   constant.CloudTencent,
			lbID:    "1:1:lb-123",
			isValid: false,
		},
		{
			cloud:   constant.CloudTencent,
			lbID:    "ap-shenzhen:ap-shenzhen:lb-123",
			isValid: false,
		},
	}
	for i, c := range testCases {
		if MatchLbStrWithID(c.cloud, c.lbID) != c.isValid {
			t.Errorf("idx: %d result is error", i)
		}
	}

}
