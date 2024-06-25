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

package ratelimiter

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewRateLimiter(t *testing.T) {
	rl := New(10, 20)
	assert.NotNil(t, rl)
}

func TestWaitTime(t *testing.T) {
	rl := New(10, 20)

	// Add a request that fits within the burst capacity
	waitTime := rl.WaitTimeMil(MB * 20)
	assert.Equal(t, int64(0), waitTime)

	// Add another request that should be rate-limited
	waitTime = rl.WaitTimeMil(MB * 5)
	assert.True(t, waitTime > 0)

	// Add another request that should be rate-limited too and wait time should be increased based on last time
	waitTime = rl.WaitTimeMil(MB * 10)
	assert.True(t, waitTime > 1000) // 1000 milliseconds
	assert.True(t, waitTime < 2000) // 2000 milliseconds
	t.Logf("waitTime: %d milliseconds", waitTime)

	// After wait for corresponding time, it will be not rate-limited when size is less than the number of tokens
	// Note that at this point, the number of tokens has not yet recovered to burst, recover speed is rate limit(10MB/s)
	time.Sleep(time.Millisecond * time.Duration(waitTime))
	waitTime = rl.WaitTimeMil(1)
	assert.Equal(t, int64(0), waitTime)
	time.Sleep(time.Second) // which will be recovered to rate limit
	waitTime = rl.WaitTimeMil(MB * 10)
	assert.Equal(t, int64(0), waitTime)
	time.Sleep(time.Second * 2) // which will be recovered to burst(20MB)
	waitTime = rl.WaitTimeMil(MB * 20)
	assert.Equal(t, int64(0), waitTime)
}

func TestStats(t *testing.T) {
	rl := New(10, 10)

	// Test initial stats
	stats := rl.Stats()
	assert.NotNil(t, stats)
	assert.Equal(t, int64(0), stats.TotalByteSize)
	assert.Equal(t, int64(0), stats.DelayCnt)
	assert.Equal(t, int64(0), stats.DelayMilliseconds)

	// Simulate some traffic
	rl.WaitTimeMil(MB * 10) // which will not be rate-limited
	rl.WaitTimeMil(MB * 10) // which will be rate-limited
	rl.WaitTimeMil(MB * 10) // which will be rate-limited

	stats = rl.Stats()
	assert.NotNil(t, stats)
	assert.True(t, stats.TotalByteSize == int64(MB*30))
	assert.True(t, stats.DelayCnt == 2)
	assert.True(t, stats.DelayMilliseconds > 1000)
	assert.True(t, stats.DelayMilliseconds <= 2000)
	t.Logf("stats.DelayMilliseconds: %d", stats.DelayMilliseconds)
}
