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

// Package ratelimiter provides custom rate limiter
package ratelimiter

import (
	"sync/atomic"
	"time"

	"golang.org/x/time/rate"
)

// RateLimiter is interface for rate limiter
type RateLimiter interface {
	// WaitTimeMil returns the wait time(milliseconds) according to the rate limiter
	// 用于引导调用方等待相应时间后再访问服务，从而达到流控效果，避免服务过载
	WaitTimeMil(size int) int64
	// Stats returns the statistics of rate limiter
	Stats() *StatsData
}

// New news a rate limiter
// it is intended for direct use for other package
// limit为流量速率限制，单位为MB/s，burst为允许处理的突发流量上限，单位为MB（允许系统在短时间内处理比速率限制更多的流量）
func New(limit, burst int) RateLimiter {
	return &RL{
		globalRL: NewGlobalRL(limit, burst),
	}
}

// RL is rate limiter for unified use
// NOTE: maybe add other rate limiter for biz, app or instance dimension as need
type RL struct {
	*globalRL
}

// globalRL is rate limiter for global dimension
type globalRL struct {
	*baseRL
}

// NewGlobalRL news a global rate limiter
func NewGlobalRL(limit, burst int) *globalRL {
	return &globalRL{
		baseRL: newBaseRL(limit, burst),
	}
}

// baseRL is base rate limiter
type baseRL struct {
	limiter           *rate.Limiter
	totalByteSize     int64
	delayCnt          int64
	delayMilliseconds int64
}

// StatsData is stats data
type StatsData struct {
	TotalByteSize     int64
	DelayCnt          int64
	DelayMilliseconds int64
}

// MB means byte size of 1MB
var MB = 1024 * 1024

// newBaseRL news a base rate limiter
// limit为流量速率限制，单位为MB/s，burst为允许处理的突发流量上限，单位为MB（允许系统在短时间内处理比速率限制更多的流量）
// 内部实现使用令牌桶算法，令牌恢复速率为limit，在令牌被消耗完且不再有任何令牌消耗时，令牌数恢复至burst需要burst/limit秒
// 举例说明：limit为100，burst为200，则将创建一个每秒生成100MB令牌、容量为200MB的限流器
func newBaseRL(limit, burst int) *baseRL {
	return &baseRL{
		limiter: rate.NewLimiter(rate.Limit(limit*MB), burst*MB),
	}
}

// WaitTimeMil returns the wait time(milliseconds) according to the rate limiter
func (r *baseRL) WaitTimeMil(size int) int64 {
	atomic.AddInt64(&r.totalByteSize, int64(size))
	reservation := r.limiter.ReserveN(time.Now(), size)
	delay := reservation.Delay()
	atomic.StoreInt64(&r.delayMilliseconds, delay.Milliseconds())
	if delay > 0 {
		atomic.AddInt64(&r.delayCnt, 1)
	}
	return delay.Milliseconds()
}

// Stats returns the statistics of rate limiter
func (r *baseRL) Stats() *StatsData {
	return &StatsData{
		TotalByteSize:     atomic.LoadInt64(&r.totalByteSize),
		DelayCnt:          atomic.LoadInt64(&r.delayCnt),
		DelayMilliseconds: atomic.LoadInt64(&r.delayMilliseconds),
	}
}
