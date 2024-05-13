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

// Package action for grpc action
package action

import (
	pbcommon "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/api/protocol/common"
)

// Action is comon business action interface.
type Action interface {
	// Input handles the input messages.
	Input() error

	// Do makes the workflows of this action base on input messages.
	Do() error

	// Output handles the output messages.
	Output() error

	// Err setup error code message in response and return the error.
	Err(errCode pbcommon.ErrCode, errMsg string) error
}

// Executor is business action executor.
type Executor struct{}

// NewExecutor creates a new Executor.
func NewExecutor() *Executor {
	return &Executor{}
}

// Execute executes the action.
func (e *Executor) Execute(action Action) error {
	if err := action.Input(); err != nil {
		return err
	}

	if err := action.Do(); err != nil {
		return err
	}

	return action.Output()
}
