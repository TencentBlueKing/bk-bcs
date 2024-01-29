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

package tools

import (
	"math/rand"
	"time"

	"go.uber.org/atomic"
)

var (
	// default immune count for retry policy
	defaultImmuneCount = uint(5)
	// default retry sleep random milliseconds range.
	defaultRangeMillSeconds = [2]uint{1000, 15000}
)

// NewRetryPolicy create a new retry policy.
// Note:
//  1. immuneCount is the count which the sleep time is constant at: retryCount * sleepRangeMS[0] * time.Millisecond.
//  2. if the retry times > immuneCount, then the sleep time will be the value calculated bellow in milliseconds.
//     sleepTime = sleepRangeMS[0] + randomValueBetween(sleepRangeMS[0], sleepRangeMS[1])
//  3. both immuneCount and sleepRangeMS value should all be > 0, if not, the default value will be used.
func NewRetryPolicy(immuneCount uint, sleepRangeMS [2]uint) *RetryPolicy {
	immune := immuneCount
	if immune == 0 {
		immune = defaultImmuneCount
	}

	rangeMS := sleepRangeMS
	if rangeMS[0] == 0 || rangeMS[1] == 0 {
		rangeMS = defaultRangeMillSeconds
	}

	return &RetryPolicy{
		immuneCount:      uint32(immune),
		rangeMillSeconds: rangeMS,
		retryCount:       atomic.NewUint32(0),
	}
}

// RetryPolicy defines how to retry the failed jobs
type RetryPolicy struct {
	// when retry count less than this, retry with a immune policy.
	// default value is defaultImmuneCount
	immuneCount uint32
	// to generate a random time between this range in million seconds
	// default value is defaultRangeMillSeconds
	rangeMillSeconds [2]uint
	retryCount       *atomic.Uint32
}

// Sleep for a while which time range with different retry count.
// When sleep is called, the retry count will be increased by 1.
func (r *RetryPolicy) Sleep() {
	defer r.retryCount.Inc()

	if r.retryCount.Load() == 0 {
		time.Sleep(time.Duration(uint32(r.rangeMillSeconds[0])/2) * time.Millisecond)
		return
	}

	if r.retryCount.Load() <= r.immuneCount {
		duration := r.retryCount.Load() * uint32(r.rangeMillSeconds[0])
		time.Sleep(time.Duration(duration) * time.Millisecond)
		return
	}

	// no matter retry how many times, sleep a const time and with an extra rand time.
	rd := rand.New(rand.NewSource(time.Now().UnixNano())) //nolint:gosec
	randTime := rd.Intn(int(r.rangeMillSeconds[1])-int(r.rangeMillSeconds[0])) + int(r.rangeMillSeconds[0])
	duration := r.rangeMillSeconds[0] + uint(randTime)
	time.Sleep(time.Duration(duration) * time.Millisecond)
}

// RetryCount return the already retried count
func (r *RetryPolicy) RetryCount() uint32 {
	return r.retryCount.Load()
}

// Reset the retry policy counter to 0
func (r *RetryPolicy) Reset() {
	r.retryCount = atomic.NewUint32(0)
}
