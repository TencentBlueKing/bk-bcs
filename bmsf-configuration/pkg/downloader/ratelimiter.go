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

package downloader

import (
	"time"
)

// Limiter is rate limiter interface.
type Limiter interface {
	// Wait handles the limit logics.
	Wait(readNum int64)

	// LimitNum returns the limit num.
	LimitNum() int64

	// Reset resets limits.
	Reset(newLimitNum int64)
}

// SimpleRateLimiter is simple rate limiter.
type SimpleRateLimiter struct {
	// limitNum is target num of limit.
	limitNum int64

	// readNum is num of read datas in each check.
	readNum int64

	// lastTS is last check timestamp.
	lastTS time.Time
}

// NewSimpleRateLimiter creates a new SimpleRateLimiter object base on target limit num.
func NewSimpleRateLimiter(limitNum int64) *SimpleRateLimiter {
	return &SimpleRateLimiter{limitNum: limitNum}
}

// Wait handles simple rate limiter limit logics.
func (l *SimpleRateLimiter) Wait(readNum int64) {
	if (time.Now().UnixNano() - l.lastTS.UnixNano()) <= int64(time.Second) {
		num := readNum - l.readNum

		// check limit num.
		if num > l.limitNum {
			wait := time.Second.Nanoseconds() - (time.Now().UnixNano() - l.lastTS.UnixNano())
			time.Sleep(time.Duration(wait))

			// update after limit action.
			l.readNum = readNum
			l.lastTS = time.Now()
		}
	} else {
		// not limit, update.
		l.readNum = readNum
		l.lastTS = time.Now()
	}
}

// LimitNum returns the limit num.
func (l *SimpleRateLimiter) LimitNum() int64 {
	return l.limitNum
}

// Reset resets a new limit num.
func (l *SimpleRateLimiter) Reset(newLimitNum int64) {
	l.limitNum = newLimitNum
}
