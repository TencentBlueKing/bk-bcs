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

package task

import (
	"context"
	"errors"
	"runtime/debug"
	"time"

	"github.com/RichardKnop/machinery/v2/log"
	"github.com/RichardKnop/machinery/v2/retry"
)

// RecoverPrintStack capture panic and print stack
func RecoverPrintStack(proc string) {
	if r := recover(); r != nil {
		log.ERROR.Printf("[%s][recover] panic: %v, stack %s", proc, r, debug.Stack())
		return
	}
}

// GetTimeOutCtx get timeout context
func GetTimeOutCtx(ctx context.Context, seconds uint32) (context.Context, context.CancelFunc) {
	if seconds > 0 {
		return context.WithTimeout(ctx, time.Duration(seconds)*time.Second)
	}

	return context.WithCancel(ctx)
}

// GetDeadlineCtx get daedline context
func GetDeadlineCtx(ctx context.Context, t *time.Time, seconds uint32) (context.Context, context.CancelFunc) {
	if t == nil || seconds <= 0 {
		return context.WithCancel(ctx)
	}

	return context.WithDeadline(context.Background(), t.Add(time.Duration(seconds)*time.Second))
}

var (
	// ErrEndLoop xxx
	ErrEndLoop = errors.New("end loop")
)

// LoopOption init LoopOptions
type LoopOption func(loop *LoopOptions)

// LoopInterval set LoopOptions interval parameter
func LoopInterval(duration time.Duration) LoopOption {
	return func(loop *LoopOptions) {
		if duration != 0 {
			loop.interval = duration
		}
	}
}

// LoopOptions loop parameter
type LoopOptions struct {
	interval time.Duration
}

// LoopDoFunc execute func do for interval
func LoopDoFunc(ctx context.Context, do func() error, ops ...LoopOption) error {
	opt := &LoopOptions{interval: time.Second}

	for _, o := range ops {
		o(opt)
	}

	coldStart := make(chan struct{}, 1)
	coldStart <- struct{}{}

	tick := time.NewTicker(opt.interval)
	defer tick.Stop()
	for {
		select {
		case <-coldStart:
		case <-tick.C:
		case <-ctx.Done():
			if errors.Is(ctx.Err(), context.Canceled) {
				log.ERROR.Printf("LoopDoFunc is canceled")
			}
			return ctx.Err()
		}

		if err := do(); err != nil {
			if errors.Is(err, ErrEndLoop) {
				return nil
			}
			return err
		}
	}
}

// retryNext 计算重试时间, 基于Fibonacci
func retryNext(count int) int {
	start := 1
	for i := 0; i < count; i++ {
		start = retry.FibonacciNext(start)
	}
	return start
}
