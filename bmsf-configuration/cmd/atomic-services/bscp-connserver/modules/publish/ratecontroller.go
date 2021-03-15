/*
Tencent is pleased to support the open source community by making Blueking Container Service available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package publish

import (
	"math"
	"time"

	"bk-bscp/cmd/atomic-services/bscp-connserver/modules/session"
)

// RateController is publishing rate controller define.
type RateController interface {

	// Arrange arranges the target nodes for step publishing.
	Arrange(targets []*session.Session)

	// Next return the next targets slice to publish.
	Next() []*session.Session
}

// StepPubUnit is step-publishing unit.
type StepPubUnit struct {
	targets []*session.Session
	wait    time.Duration
}

// SimpleRateController is a simple rate controller.
type SimpleRateController struct {
	// steps slice for publishing.
	steps []*StepPubUnit

	// step count.
	stepCount int

	// min unit size of step.
	minStepUnitSize int

	// time duration to wait for next slice.
	stepWait time.Duration

	// steps slice index.
	index int
}

// NewSimpleRateController creates new SimpleRateController.
func NewSimpleRateController(stepCount, minStepUnitSize int, stepWait time.Duration) *SimpleRateController {
	return &SimpleRateController{stepCount: stepCount, minStepUnitSize: minStepUnitSize, stepWait: stepWait}
}

func (s *SimpleRateController) arrange(targets []*session.Session, unitSize int) {
	if len(targets) == 0 {
		return
	}

	if len(targets) <= unitSize {
		s.steps = append(s.steps, &StepPubUnit{targets: targets})
	} else {
		s.steps = append(s.steps, &StepPubUnit{targets: targets[0:unitSize], wait: s.stepWait})
		s.arrange(targets[unitSize:], unitSize)
	}
}

// Arrange arranges the targets with simple rate controller mode.
func (s *SimpleRateController) Arrange(targets []*session.Session) {
	if len(targets) == 0 {
		return
	}

	unitSize := int(math.Ceil(float64(len(targets)) / float64(s.stepCount)))
	if unitSize < s.minStepUnitSize {
		unitSize = s.minStepUnitSize
	}
	s.arrange(targets, unitSize)
}

// Next returns teh next targets slice for publishing.
func (s *SimpleRateController) Next() []*session.Session {
	if len(s.steps) == 0 {
		return nil
	}

	if s.index >= len(s.steps) {
		// no more steps.
		return nil
	}

	step := s.steps[s.index]
	if s.index == 0 {
		s.index++
		return step.targets
	}

	time.Sleep(s.steps[s.index-1].wait)

	s.index++
	return step.targets
}

// you can implement your own rate controller base on load information here...
