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

package tencentcloud

import (
	"os"
	"strconv"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/throttle"
)

var (
	sharedRateLimiterOnce sync.Once
	sharedRateLimiter     throttle.RateLimiter
	trySharedThrottle     = func() {
		GetSharedRateLimiter().Accept()
	}
)

// GetSharedRateLimiter returns the process-wide Tencent Cloud API rate limiter.
// SdkWrapper, APIWrapper and sslClientImpl share the same token bucket instance.
func GetSharedRateLimiter() throttle.RateLimiter {
	sharedRateLimiterOnce.Do(func() {
		qps, bucketSize := parseRateLimitConfig()
		sharedRateLimiter = throttle.NewTokenBucket(qps, bucketSize)
	})
	return sharedRateLimiter
}

func parseRateLimitConfig() (int64, int64) {
	qps := int64(defaultThrottleQPS)
	bucketSize := int64(defaultBucketSize)

	qpsStr := os.Getenv(EnvNameTencentCloudRateLimitQPS)
	if len(qpsStr) != 0 {
		parsed, err := strconv.ParseInt(qpsStr, 10, 64)
		if err != nil {
			blog.Warnf("parse rate limit qps %s failed, err %s, use default %d",
				qpsStr, err.Error(), defaultThrottleQPS)
		} else {
			qps = parsed
		}
	}

	bucketSizeStr := os.Getenv(EnvNameTencentCloudRateLimitBucketSize)
	if len(bucketSizeStr) != 0 {
		parsed, err := strconv.ParseInt(bucketSizeStr, 10, 64)
		if err != nil {
			blog.Warnf("parse rate limit bucket size %s failed, err %s, use default %d",
				bucketSizeStr, err.Error(), defaultBucketSize)
		} else {
			bucketSize = parsed
		}
	}
	return qps, bucketSize
}
