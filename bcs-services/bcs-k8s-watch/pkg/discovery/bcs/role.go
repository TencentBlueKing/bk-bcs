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

// Role int
type Role int

const (
	// RoleNobody xxx
	RoleNobody Role = iota
	// RoleLeader xxx
	RoleLeader
	// RoleSlave xxx
	RoleSlave
)

// ActionType int
type ActionType int

const (
	// ActionTypeMakeNobody xxx
	ActionTypeMakeNobody ActionType = iota
	// ActionTypeMakeLeader xxx
	ActionTypeMakeLeader
	// ActionTypeMakeSlave xxx
	ActionTypeMakeSlave
)

// RoleState interface
type RoleState interface {
	OnMakeNobody(elector *LeaderElector)
	OnMakeLeader(elector *LeaderElector)
	OnMakeSlave(elector *LeaderElector)
}

// NewRoleState return role state
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

// NobodyState xxx
// State transition for Nobody State
// NobodyState
type NobodyState struct{}

// OnMakeNobody for nobody state
func (s NobodyState) OnMakeNobody(elector *LeaderElector) {}

// OnMakeLeader for nobody state
func (s NobodyState) OnMakeLeader(elector *LeaderElector) {
	elector.SetRole(RoleLeader)
	blog.Info("Role changed: nobody -> leader")
	go elector.callbacks.OnStartedLeading(elector.StopChan)
}

// OnMakeSlave for nobody state
func (s NobodyState) OnMakeSlave(elector *LeaderElector) {
	elector.SetRole(RoleSlave)
	blog.Info("Role changed: nobody -> slave")
	go elector.callbacks.OnStoppedLeading()
}

// LeaderState xxx
// State transition for Leader State
// LeaderState
type LeaderState struct{}

// OnMakeNobody for leader state
func (s LeaderState) OnMakeNobody(elector *LeaderElector) {
	blog.Info("Role changed: leader -> nobody")
	s.stopLeading(elector, RoleNobody)
}

// OnMakeLeader for leader state
func (s LeaderState) OnMakeLeader(elector *LeaderElector) {}

// OnMakeSlave for leader state
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

// SlaveState xxx
// State transition for Slave State
// SlaveState
type SlaveState struct{}

// OnMakeNobody for slave state
func (s SlaveState) OnMakeNobody(elector *LeaderElector) {
	elector.SetRole(RoleNobody)
	blog.Info("Role changed: slave -> nobody")
}

// OnMakeLeader for slave state
func (s SlaveState) OnMakeLeader(elector *LeaderElector) {
	elector.SetRole(RoleLeader)
	blog.Info("Role changed: slave -> leader")
	go elector.callbacks.OnStartedLeading(elector.StopChan)
}

// OnMakeSlave for slave state
func (s SlaveState) OnMakeSlave(elector *LeaderElector) {}
