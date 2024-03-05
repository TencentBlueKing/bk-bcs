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

package expectations

import (
	"testing"
)

func TestScale(t *testing.T) {
	e := NewScaleExpectations()
	controllerKey01 := "default/cs01"
	controllerKey02 := "default/cs02"
	pod01 := "pod01"
	pod02 := "pod02"

	e.ExpectScale(controllerKey01, Create, pod01)
	e.ExpectScale(controllerKey01, Create, pod02)
	e.ExpectScale(controllerKey01, Delete, pod01)
	if ok, _ := e.SatisfiedExpectations(controllerKey01); ok {
		t.Fatalf("expected not satisfied")
	}

	e.ObserveScale(controllerKey01, Create, pod02)
	e.ObserveScale(controllerKey01, Create, pod01)
	if ok, _ := e.SatisfiedExpectations(controllerKey01); ok {
		t.Fatalf("expected not satisfied")
	}

	e.ObserveScale(controllerKey02, Delete, pod01)
	if ok, _ := e.SatisfiedExpectations(controllerKey01); ok {
		t.Fatalf("expected not satisfied")
	}

	e.ObserveScale(controllerKey01, Delete, pod01)
	if ok, _ := e.SatisfiedExpectations(controllerKey01); !ok {
		t.Fatalf("expected satisfied")
	}
}
