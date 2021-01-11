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

package executor

import (
	"go.uber.org/ratelimit"
)

// Action is comon business action interface.
type Action interface {
	// Input handles the input messages.
	Input() error

	// Do makes the workflows of this action base on input messages.
	Do() error

	// Output handles the output messages.
	Output() error
}

// ActionWithAuth is comon business action interface with authorization.
type ActionWithAuth interface {
	// Input handles the input messages.
	Input() error

	// Authorize checks the action authorization.
	Authorize() error

	// Do makes the workflows of this action base on input messages.
	Do() error

	// Output handles the output messages.
	Output() error
}

// Executor is business action executor.
type Executor struct {
	// limiter is rate limiter that ctrl each action execute limit, it
	// limits in per-second level.
	limiter ratelimit.Limiter
}

// NewExecutor creates a new Executor.
func NewExecutor() *Executor {
	return &Executor{limiter: nil}
}

// NewRateLimitExecutor creates a new Executor with rate limit.
func NewRateLimitExecutor(rate int) *Executor {
	if rate <= 0 {
		// not limit.
		return &Executor{limiter: nil}
	}
	return &Executor{limiter: ratelimit.New(rate)}
}

// Execute executes the action.
func (e *Executor) Execute(action Action) error {
	// rate limit.
	if e.limiter != nil {
		e.limiter.Take()
	}

	// executes.
	if err := action.Input(); err != nil {
		return err
	}
	if err := action.Do(); err != nil {
		return err
	}
	return action.Output()
}

// ExecuteWithAuth executes the action with auth.
func (e *Executor) ExecuteWithAuth(action ActionWithAuth) error {
	// rate limit.
	if e.limiter != nil {
		e.limiter.Take()
	}

	// executes.
	if err := action.Input(); err != nil {
		return err
	}
	if err := action.Authorize(); err != nil {
		return err
	}
	if err := action.Do(); err != nil {
		return err
	}
	return action.Output()
}
