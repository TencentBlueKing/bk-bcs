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

package utils

import (
	"context"
	"errors"
	"time"

	"github.com/avast/retry-go"
)

var (
	// ErrContextTimeout err context timeout
	ErrContextTimeout = errors.New("operation timeout")
)

// RetryOption define retry option func
type RetryOption func(loop *RetryOptions)

// RetryAttempts set RetryOption attempts
func RetryAttempts(attempts uint) RetryOption {
	return func(op *RetryOptions) {
		op.attempts = attempts
	}
}

// RetryTimeout set RetryOption timeout
func RetryTimeout(timeout time.Duration) RetryOption {
	return func(op *RetryOptions) {
		op.timeout = timeout
	}
}

// RetryOptions retry options
type RetryOptions struct {
	attempts uint
	timeout  time.Duration
}

// RetryWithTimeout func and
func RetryWithTimeout(fn func() error, opts ...RetryOption) error {
	opt := &RetryOptions{attempts: 3, timeout: time.Second}
	for _, o := range opts {
		o(opt)
	}
	return retry.Do(
		func() error {
			ctx, cancel := context.WithTimeout(context.Background(), opt.timeout)

			var err error
			go func() {
				defer cancel()
				err = fn()
			}()

			select {
			case <-ctx.Done():
				if err = ctx.Err(); errors.Is(err, context.DeadlineExceeded) {
					return ErrContextTimeout
				}
				return err
			}
		},
		retry.Attempts(opt.attempts),
		retry.RetryIf(func(err error) bool {
			return errors.Is(err, ErrContextTimeout)
		}),
	)
}
