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

package scaler

import (
	"time"

	"k8s.io/client-go/util/workqueue"
)

// FixedItemIntervalRateLimiter limits items to a fixed-rate interval
type FixedItemIntervalRateLimiter struct {
	interval time.Duration
}

var _ workqueue.RateLimiter = &FixedItemIntervalRateLimiter{}

// NewFixedItemIntervalRateLimiter creates a new instance of a RateLimiter using a fixed interval
func NewFixedItemIntervalRateLimiter(interval time.Duration) workqueue.RateLimiter {
	return &FixedItemIntervalRateLimiter{
		interval: interval,
	}
}

// When returns the interval of the rate limiter
func (r *FixedItemIntervalRateLimiter) When(item interface{}) time.Duration {
	return r.interval
}

// NumRequeues returns back how many failures the item has had
func (r *FixedItemIntervalRateLimiter) NumRequeues(item interface{}) int {
	return 1
}

// Forget indicates that an item is finished being retried.
func (r *FixedItemIntervalRateLimiter) Forget(item interface{}) {
}

// NewDefaultGPARateLimiter creates a rate limiter which limits overall (as per the
// default controller rate limiter), as well as per the resync interval
func NewDefaultGPARateLimiter(interval time.Duration) workqueue.RateLimiter {
	return NewFixedItemIntervalRateLimiter(interval)
}
