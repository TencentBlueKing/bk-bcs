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
	"context"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	elbv2 "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
)

// getRegionClient create region client
func (sw *SdkWrapper) getRegionClient(region string) *elbv2.Client {
	cli, ok := sw.elbClientMap[region]
	if !ok {
		credentials := aws.CredentialsProviderFunc(func(ctx context.Context) (aws.Credentials, error) {
			return aws.Credentials{
				AccessKeyID:     sw.secretID,
				SecretAccessKey: sw.secretKey,
			}, nil
		})
		options := elbv2.Options{Region: region, Credentials: credentials}
		options.Retryer = retry.NewStandard(RetryerWithDefaultOptions)
		newCli := elbv2.New(options)
		sw.elbClientMap[region] = newCli
		return newCli
	}
	return cli
}

// loadEnv load config from environment
func (sw *SdkWrapper) loadEnv() error {
	if len(sw.secretID) == 0 {
		sw.secretID = os.Getenv(EnvNameAWSAccessKeyID)
	}
	if len(sw.secretKey) == 0 {
		sw.secretKey = os.Getenv(EnvNameAWSAccessKey)
	}

	qpsStr := os.Getenv(EnvNameAWSRateLimitQPS)
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

	bucketSizeStr := os.Getenv(EnvNameAWSRateLimitBucketSize)
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
