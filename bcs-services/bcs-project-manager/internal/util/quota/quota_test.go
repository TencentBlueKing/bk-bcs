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

	"k8s.io/apimachinery/pkg/api/resource"
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
