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

package gcp

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"google.golang.org/api/compute/v1"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
)

func (sw *SdkWrapper) loadEnv() error {
	if len(sw.credentials) == 0 {
		sw.credentials = []byte(os.Getenv(EnvNameGCPCredentials))
	}

	qpsStr := os.Getenv(EnvNameGCPRateLimitQPS)
	if len(qpsStr) != 0 {
		qps, err := strconv.ParseInt(qpsStr, 10, 64)
		if err != nil {
			blog.Warnf("parse rate limit qps %s failed, err %s, use default %d",
				qpsStr, err.Error(), defaultThrottleQPS)
			sw.ratelimitqps = int64(defaultThrottleQPS)
		} else {
			sw.ratelimitqps = qps
		}
	} else {
		sw.ratelimitqps = int64(defaultThrottleQPS)
	}

	bucketSizeStr := os.Getenv(EnvNameGCPRateLimitBucketSize)
	if len(bucketSizeStr) != 0 {
		bucketSize, err := strconv.ParseInt(bucketSizeStr, 10, 64)
		if err != nil {
			blog.Warnf("parse rate limit bucket size %s failed, err %s, use default %d",
				bucketSizeStr, err.Error(), defaultBucketSize)
			sw.ratelimitbucketSize = int64(defaultBucketSize)
		} else {
			sw.ratelimitbucketSize = bucketSize
		}
	} else {
		sw.ratelimitbucketSize = int64(defaultBucketSize)
	}
	return nil
}

// call tryThrottle before each api call
func (sw *SdkWrapper) tryThrottle() {
	now := time.Now()
	sw.throttler.Accept()
	if latency := time.Since(now); latency > maxLatency {
		pc, _, _, _ := runtime.Caller(2)
		callerName := runtime.FuncForPC(pc).Name()
		blog.Infof("Throttling request took %d ms, function: %s", latency, callerName)
	}
}

// Wait wait for cloud api async
func (sw *SdkWrapper) Wait(ctx context.Context, project string, op *compute.Operation) error {
	tk := time.NewTicker(defaultPollingInterval)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-tk.C:
			var resp *compute.Operation
			var err error
			blog.Infof("wait for operation %s", op.Name)
			if op.Region != "" {
				regionStrs := strings.Split(op.Region, "/")
				region := regionStrs[len(regionStrs)-1]
				resp, err = sw.computeService.RegionOperations.Get(project, region, op.Name).Context(ctx).Do()
			}
			if op.Zone != "" {
				zoneStrs := strings.Split(op.Zone, "/")
				zone := zoneStrs[len(zoneStrs)-1]
				resp, err = sw.computeService.ZoneOperations.Get(project, zone, op.Name).Context(ctx).Do()
			} else {
				resp, err = sw.computeService.GlobalOperations.Get(project, op.Name).Context(ctx).Do()
			}
			if err != nil {
				blog.Errorf("wait for operation %s failed, err %s", op.Name, err.Error())
				return err
			}
			if resp == nil {
				return fmt.Errorf("operation %s not found", op.Name)
			}
			if resp.Status == "DONE" {
				if resp.Error != nil {
					e, err := resp.Error.MarshalJSON()
					if err != nil {
						return err
					}
					return fmt.Errorf("operation %s failed, error: %s", resp.Name, string(e))
				}
				return nil
			}
		}
	}
}
