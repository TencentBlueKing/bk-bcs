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

// Package loop xxx
package loop

import (
	"context"
	"errors"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
)

// RunFunc is task concreate action
type RunFunc func(taskID, stepName string)

// Action is the interface that user action must implement
type Action interface {
	Name() string
	Run(taskID, stepName string) error
}

var (
	// EndLoop xxx
	EndLoop = errors.New("end loop") // nolint
)

// LoopOption init LoopOptions
type LoopOption func(loop *LoopOptions) // nolint

// LoopInterval set LoopOptions interval parameter
func LoopInterval(duration time.Duration) LoopOption { // nolint
	return func(loop *LoopOptions) {
		if duration != 0 {
			loop.interval = duration
		}
	}
}

// LoopOptions loop parameter
type LoopOptions struct { // nolint
	interval time.Duration
}

// LoopDoFunc execute func do for interval
func LoopDoFunc(ctx context.Context, do func() error, ops ...LoopOption) error { // nolint
	opt := &LoopOptions{interval: time.Second}

	for _, o := range ops {
		o(opt)
	}

	coldStart := make(chan struct{}, 1)
	coldStart <- struct{}{}

	tick := time.Tick(opt.interval) // nolint
	for {
		select {
		case <-coldStart:
		case <-tick:
		case <-ctx.Done():
			if errors.Is(ctx.Err(), context.Canceled) || errors.Is(ctx.Err(), context.DeadlineExceeded) {
				blog.Errorf("LoopDoFunc is canceled or timeOut")
			}
			return ctx.Err()
		}

		if err := do(); err != nil {
			if errors.Is(err, EndLoop) {
				return nil
			}
			return err
		}
	}
}
