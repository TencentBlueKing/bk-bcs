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

package quota

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/api/resource"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/proto/bcsproject"
)

func TestQuota(t *testing.T) {
	q1, err := resource.ParseQuantity("10m")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(q1.AsApproximateFloat64())

	q2, err := resource.ParseQuantity("1")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(q2.AsApproximateFloat64())
}

func TestValidateQuotaEquality(t *testing.T) {
	// Case 1: Nil quota
	err := ValidateQuotaEquality(nil)
	assert.Nil(t, err)

	// Case 2: Equal CPU and Memory
	qEqual := &proto.ResourceQuota{
		CpuLimits:      "2",
		CpuRequests:    "2000m",
		MemoryLimits:   "4Gi",
		MemoryRequests: "4096Mi",
	}
	err = ValidateQuotaEquality(qEqual)
	assert.Nil(t, err)

	// Case 3: Unequal CPU
	qUnequalCPU := &proto.ResourceQuota{
		CpuLimits:      "2",
		CpuRequests:    "1000m",
		MemoryLimits:   "4Gi",
		MemoryRequests: "4Gi",
	}
	err = ValidateQuotaEquality(qUnequalCPU)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "cpu limits and requests must be consistent under shared cluster")

	// Case 4: Unequal Memory
	qUnequalMem := &proto.ResourceQuota{
		CpuLimits:      "2",
		CpuRequests:    "2",
		MemoryLimits:   "4Gi",
		MemoryRequests: "2Gi",
	}
	err = ValidateQuotaEquality(qUnequalMem)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "memory limits and requests must be consistent under shared cluster")
}
