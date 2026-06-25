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
	"testing"
)

func TestSharedRateLimiterSameInst(t *testing.T) {
	first := GetSharedRateLimiter()
	second := GetSharedRateLimiter()
	if first != second {
		t.Fatal("GetSharedRateLimiter should return the same instance")
	}
}

func TestParseRateLimitConfigFromEnv(t *testing.T) {
	t.Setenv(EnvNameTencentCloudRateLimitQPS, "30")
	t.Setenv(EnvNameTencentCloudRateLimitBucketSize, "40")

	qps, bucket := parseRateLimitConfig()
	if qps != 30 {
		t.Fatalf("expected qps 30, got %d", qps)
	}
	if bucket != 40 {
		t.Fatalf("expected bucket 40, got %d", bucket)
	}
}

func TestParseRateLimitConfigDefaults(t *testing.T) {
	os.Unsetenv(EnvNameTencentCloudRateLimitQPS)
	os.Unsetenv(EnvNameTencentCloudRateLimitBucketSize)

	qps, bucket := parseRateLimitConfig()
	if qps != int64(defaultThrottleQPS) {
		t.Fatalf("expected default qps %d, got %d", defaultThrottleQPS, qps)
	}
	if bucket != int64(defaultBucketSize) {
		t.Fatalf("expected default bucket %d, got %d", defaultBucketSize, bucket)
	}
}
