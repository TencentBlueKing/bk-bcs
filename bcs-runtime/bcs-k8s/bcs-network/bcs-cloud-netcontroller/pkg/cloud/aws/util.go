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

package aws

import (
	"math/rand"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
)

// BackoffTimeSeries time series for back off
type BackoffTimeSeries interface {
	Reset()
	Next() time.Duration
}

// IncreaseSeries increased time series
type IncreaseSeries struct {
	base    time.Duration
	current time.Duration
	factor  float64
	jitter  float64
}

// NewIncreseSeries new increased time series
func NewIncreseSeries(base time.Duration, factor, jitter float64) *IncreaseSeries {
	return &IncreaseSeries{
		base:    base,
		current: base,
		factor:  factor,
		jitter:  jitter,
	}
}

// Reset reset time series
func (is *IncreaseSeries) Reset() {
	is.current = is.base
}

// Next get next time series
func (is *IncreaseSeries) Next() time.Duration {
	ret := is.current
	is.current = is.current +
		time.Duration(
			int64(float64(is.current)*is.factor)+
				int64(float64(is.current)*(float64(rand.Intn(2000))/1000-1)*is.jitter))
	return ret
}

// RetryWithBackoffTime retry with back off time
func RetryWithBackoffTime(maxRetries int64, timeSeries BackoffTimeSeries, fn func() bool) {
	retries := int64(0)
	for {
		blog.V(2).Infof("do func (%d/%d)", retries, maxRetries)
		flag := fn()
		if flag {
			return
		}
		time.Sleep(timeSeries.Next())
		retries++
		if retries >= maxRetries {
			blog.Errorf("do func exceed max retries")
			return
		}
	}
}
