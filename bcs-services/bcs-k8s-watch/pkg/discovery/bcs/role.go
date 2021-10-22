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
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
)

type Role int

const (
	RoleNobody Role = iota
	RoleLeader
	RoleSlave
)

type ActionType int

const (
	ActionTypeMakeNobody ActionType = iota
	ActionTypeMakeLeader
	ActionTypeMakeSlave
)

type RoleState interface {
	OnMakeNobody(elector *LeaderElector)
	OnMakeLeader(elector *LeaderElector)
	OnMakeSlave(elector *LeaderElector)
}

func NewRoleState(role Role) RoleState {
	switch role {
	case RoleNobody:
		return NobodyState{}
	case RoleLeader:
		return LeaderState{}
	case RoleSlave:
		return SlaveState{}
	}
	return nil
}

//
// State transition for Nobody State
//
type NobodyState struct{}

func (s NobodyState) OnMakeNobody(elector *LeaderElector) {}

func (s NobodyState) OnMakeLeader(elector *LeaderElector) {
	elector.SetRole(RoleLeader)
	blog.Info("Role changed: nobody -> leader")
	go elector.callbacks.OnStartedLeading(elector.StopChan)
}

func (s NobodyState) OnMakeSlave(elector *LeaderElector) {
	elector.SetRole(RoleSlave)
	blog.Info("Role changed: nobody -> slave")
	go elector.callbacks.OnStoppedLeading()
}

//
// State transition for Leader State
//
type LeaderState struct{}

func (s LeaderState) OnMakeNobody(elector *LeaderElector) {
	blog.Info("Role changed: leader -> nobody")
	s.stopLeading(elector, RoleNobody)
}

func (s LeaderState) OnMakeLeader(elector *LeaderElector) {}

func (s LeaderState) OnMakeSlave(elector *LeaderElector) {
	blog.Info("Role changed: leader -> slave")
	s.stopLeading(elector, RoleSlave)
}

func (s LeaderState) stopLeading(elector *LeaderElector, targetRole Role) {
	elector.SetRole(targetRole)
	// Try to stop old master goroutine
	// This should not block
	select {
	case elector.StopChan <- struct{}{}:
	default:
	}
	go elector.callbacks.OnStoppedLeading()
}

//
// State transition for Slave State
//
type SlaveState struct{}

func (s SlaveState) OnMakeNobody(elector *LeaderElector) {
	elector.SetRole(RoleNobody)
	blog.Info("Role changed: slave -> nobody")
}

func (s SlaveState) OnMakeLeader(elector *LeaderElector) {
	elector.SetRole(RoleLeader)
	blog.Info("Role changed: slave -> leader")
	go elector.callbacks.OnStartedLeading(elector.StopChan)
}

func (s SlaveState) OnMakeSlave(elector *LeaderElector) {}
