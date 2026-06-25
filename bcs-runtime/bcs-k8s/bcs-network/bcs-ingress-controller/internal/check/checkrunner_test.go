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

package check

import (
	"context"
	"testing"
)

type fakeIntervalChecker struct {
	name string
}

func (f *fakeIntervalChecker) Run() {}

func TestCheckRunnerRegisterIntervals(t *testing.T) {
	ctx := context.Background()
	runner := NewCheckRunner(ctx)

	minChecker := &fakeIntervalChecker{name: "per-min"}
	hourChecker := &fakeIntervalChecker{name: "per-hour"}
	tenMinChecker := &fakeIntervalChecker{name: "per-10min"}

	runner.Register(minChecker, CheckPerMin).
		Register(hourChecker, CheckPer60Min).
		Register(tenMinChecker, CheckPer10Min)

	if len(runner.checkPerMin) != 1 || runner.checkPerMin[0] != minChecker {
		t.Fatalf("checkPerMin = %#v, want [%#v]", runner.checkPerMin, minChecker)
	}
	if len(runner.checkPer60Min) != 1 || runner.checkPer60Min[0] != hourChecker {
		t.Fatalf("checkPer60Min = %#v, want [%#v]", runner.checkPer60Min, hourChecker)
	}
	if len(runner.checkPer10Min) != 1 || runner.checkPer10Min[0] != tenMinChecker {
		t.Fatalf("checkPer10Min = %#v, want [%#v]", runner.checkPer10Min, tenMinChecker)
	}
}
