/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package throttle

import (
	"github.com/juju/ratelimit"
)

// RateLimiter ratelimiter interface
type RateLimiter interface {
	// TryAccept returns true if a token is taken immediately. Otherwise,
	// it returns false.
	TryAccept() bool

	// Accept will wait and not return unless a token becomes available.
	Accept()

	// QPS returns QPS of this rate limiter
	QPS() int64

	// Burst returns the burst of this rate limiter
	Burst() int64
}

// NewTokenBucket create new token bucket ratelimiter
func NewTokenBucket(qps, burst int64) *TokenBucket {
	limiter := ratelimit.NewBucketWithRate(float64(qps), burst)
	return &TokenBucket{
		limiter: limiter,
		qps:     qps,
		burst:   burst,
	}
}

// TokenBucket implements RateLimiter
type TokenBucket struct {
	limiter *ratelimit.Bucket
	qps     int64
	burst   int64
}

// TryAccept try to accept one token
func (t *TokenBucket) TryAccept() bool {
	return t.limiter.TakeAvailable(1) == 1
}

// Accept accept one token
func (t *TokenBucket) Accept() {
	t.limiter.Wait(1)
}

// QPS get qps
func (t *TokenBucket) QPS() int64 {
	return t.qps
}

// Burst get burst
func (t *TokenBucket) Burst() int64 {
	return t.burst
}

// NewMockRateLimiter create mock rate limiter
func NewMockRateLimiter() RateLimiter {
	return &mockRatelimiter{}
}

type mockRatelimiter struct{}

// TryAccept try to accept one token
func (*mockRatelimiter) TryAccept() bool {
	return true
}

// Accept accept one token
func (*mockRatelimiter) Accept() {

}

// QPS get qps
func (*mockRatelimiter) QPS() int64 {
	return 0
}

// Burst get burst
func (*mockRatelimiter) Burst() int64 {
	return 0
}
