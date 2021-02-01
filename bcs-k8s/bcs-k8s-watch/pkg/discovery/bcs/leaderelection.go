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

package bcs

import (
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-k8s/bcs-k8s-watch/pkg/discovery/register"

	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// LeaderElector takes a NodeRegister and consumes its ServerEvents
type LeaderElector struct {
	nreg         *register.NodeRegister
	callbacks    *LeaderCallbacks
	currentState register.ServiceState
	role         Role

	StopChan chan struct{}
}

// LeaderCallbacks if the callback functions for LeaderElector
type LeaderCallbacks struct {
	OnStartedLeading func(stop <-chan struct{})
	OnStoppedLeading func()
}

func NewLeaderElector(nreg *register.NodeRegister, callbacks *LeaderCallbacks) *LeaderElector {
	return &LeaderElector{
		nreg:      nreg,
		callbacks: callbacks,
		role:      RoleNobody,
		currentState: register.ServiceState{
			MyPostion: -1,
		},
		StopChan: make(chan struct{}),
	}
}

func (le *LeaderElector) IsLeader() bool {
	return le.role == RoleLeader
}

// IsLeader determine is curent node is leader by telling if it's the first registered node in sequence
func (le *LeaderElector) getActionType(state register.ServiceState) ActionType {
	if state.MyPostion == 0 {
		return ActionTypeMakeLeader
	} else if state.MyPostion > 0 {
		return ActionTypeMakeSlave
	} else {
		return ActionTypeMakeNobody
	}
}

func (le *LeaderElector) SetRole(role Role) {
	le.role = role
}

// Run starts a loop reads events from nreg.StateChan, when any ServiceEvent causes current role to be changed,
// it will use goroutine to run callback functions
func (le *LeaderElector) Run() {
	var state register.ServiceState
	for {
		state = <-le.nreg.StateChan

		actionType := le.getActionType(state)
		blog.Debug(fmt.Sprintf("New ServiceState got: %+v, actionType=%d", state, actionType))

		roleState := NewRoleState(le.role)
		switch actionType {
		case ActionTypeMakeNobody:
			roleState.OnMakeNobody(le)
		case ActionTypeMakeLeader:
			roleState.OnMakeLeader(le)
		case ActionTypeMakeSlave:
			roleState.OnMakeSlave(le)
		}

		le.currentState = state
	}
}
