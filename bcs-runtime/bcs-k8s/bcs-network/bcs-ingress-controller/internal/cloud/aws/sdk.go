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
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws/ratelimit"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	elbv2 "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	gocache "github.com/patrickmn/go-cache"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog/glog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/throttle"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/cloud"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/metrics"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/pkg/common"
)

// SdkWrapper wrapper for aws sdk
type SdkWrapper struct {
	// secret id for the aws account
	secretID string
	// secret key for the aws account
	secretKey string
	// map of client for the different region elb
	elbClientMap map[string]*elbv2.Client

	// ratelimit
	ratelimitqps        int64
	ratelimitbucketSize int64
	// rate limiter for calling sdk
	throttler throttle.RateLimiter

	// key: genRegionArn(region, lbArn string) string
	// value: types.LoadBalancer
	lbCache *gocache.Cache
}

const (
	// EnvNameAWSRegion env name of aws region
	EnvNameAWSRegion = "AWS_REGION"
	// EnvNameAWSAccessKeyID env name of aws access key id
	EnvNameAWSAccessKeyID = "AWS_ACCESS_KEY_ID"
	// EnvNameAWSAccessKey env name of aws secret key
	EnvNameAWSAccessKey = "AWS_ACCESS_KEY"

	// EnvNameAWSRateLimitQPS env name for aws api rate limit qps
	EnvNameAWSRateLimitQPS = "AWS_RATELIMIT_QPS"
	// EnvNameAWSRateLimitBucketSize env name for aws api rate limit bucket size
	EnvNameAWSRateLimitBucketSize = "AWS_RATELIMIT_BUCKET_SIZE"
)

var (
	// If the delay caused by the frequency limit exceeds this value, it is recorded in the log
	maxLatency = 120 * time.Millisecond
	// the maximum number of retries caused by server error or API overrun
	maxRetry = 5
	// qps for rate limit
	defaultThrottleQPS = 50
	// bucket size for rate limit
	defaultBucketSize = 50
	// wait seconds when cloud api is busy
	waitPeriodLBDealing = 2

	lbCacheExpire        = time.Hour * 24 * 30
	lbCacheCleanInterval = time.Hour * 24
)

// NewSdkWrapper create a new aws sdk wrapper
func NewSdkWrapper() (*SdkWrapper, error) {
	sw := &SdkWrapper{}
	err := sw.loadEnv()
	if err != nil {
		return nil, err
	}
	sw.elbClientMap = make(map[string]*elbv2.Client)
	sw.throttler = throttle.NewTokenBucket(sw.ratelimitqps, sw.ratelimitbucketSize)

	// lbCache仅用于存放lb的一些固定信息，缓存时间可以设的比较久
	sw.lbCache = gocache.New(lbCacheExpire, lbCacheCleanInterval)
	return sw, nil
}

// RetryerWithDefaultOptions returns a retryer with default options
func RetryerWithDefaultOptions(o *retry.StandardOptions) {
	o.MaxAttempts = maxRetry
	retryAbles := retry.RetryableHTTPStatusCode{}
	o.Retryables = append(o.Retryables, retryAbles)
	o.RateLimiter = ratelimit.NewTokenRateLimit(uint(defaultBucketSize))
}

// NewSdkWrapperWithSecretIDKey create a new aws sdk wrapper with secret id and key
func NewSdkWrapperWithSecretIDKey(secretID, secretKey string) (*SdkWrapper, error) {
	sw := &SdkWrapper{}
	sw.secretID = secretID
	sw.secretKey = secretKey
	return NewSdkWrapper()
}

// DescribeLoadBalancers describe load balancers
func (sw *SdkWrapper) DescribeLoadBalancers(region string, input *elbv2.DescribeLoadBalancersInput) (
	*elbv2.DescribeLoadBalancersOutput, error) {
	blog.V(3).Infof("DescribeLoadBalancers input: %s", common.ToJsonString(input))

	startTime := time.Now()
	// 统计API调用延时/状态
	mf := func(ret string) {
		metrics.ReportLibRequestMetric(
			SystemNameInMetricAWS,
			HandlerNameInMetricAWSSDK,
			"DescribeLoadBalancers", ret, startTime)
	}
	sw.tryThrottle()
	out, err := sw.getRegionClient(region).DescribeLoadBalancers(context.TODO(), input)
	if err != nil {
		rerr := ResolveError(err)
		if rerr.IsExceededAttemptError() {
			mf(metrics.LibCallStatusTimeout)
			errMsg := fmt.Sprintf("DescribeLoadBalancers req[%s] out of maxRetry %d", common.ToJsonString(input),
				maxRetry)
			blog.Errorf(errMsg)
			return nil, fmt.Errorf(errMsg)
		}
		if rerr.IsOperationError() {
			mf(metrics.LibCallStatusErr)
			errMsg := fmt.Sprintf("DescribeLoadBalancers req[%s] failed, err %s", common.ToJsonString(input),
				err.Error())
			blog.Errorf(errMsg)
			return nil, fmt.Errorf(errMsg)
		}
		blog.Errorf("DescribeLoadBalancers req[%s] failed, err: %s", common.ToJsonString(input), err.Error())
		return nil, rerr.Unwrap()
	}
	blog.V(3).Infof("DescribeLoadBalancers response: %s", common.ToJsonString(out))
	mf(metrics.LibCallStatusOK)
	for _, lb := range out.LoadBalancers {
		sw.lbCache.SetDefault(genRegionArn(region, *lb.LoadBalancerArn), lb)
	}
	return out, nil
}

// CreateListener create listener
func (sw *SdkWrapper) CreateListener(region string, input *elbv2.CreateListenerInput) (
	*elbv2.CreateListenerOutput, error) {
	blog.V(3).Infof("CreateListener input: %s", common.ToJsonString(input))

	startTime := time.Now()
	// 统计API调用延时/状态
	mf := func(ret string) {
		defer metrics.ReportLibRequestMetric(
			SystemNameInMetricAWS,
			HandlerNameInMetricAWSSDK,
			"CreateListener", ret, startTime)
	}

	sw.tryThrottle()
	out, err := sw.getRegionClient(region).CreateListener(context.TODO(), input)
	if err != nil {
		rerr := ResolveError(err)
		if rerr.IsExceededAttemptError() {
			mf(metrics.LibCallStatusTimeout)
			errMsg := fmt.Sprintf("CreateListener out of maxRetry %d", maxRetry)
			blog.Errorf(errMsg)
			return nil, fmt.Errorf(errMsg)
		}
		if rerr.IsOperationError() {
			mf(metrics.LibCallStatusErr)
			errMsg := fmt.Sprintf("CreateListener failed, err %s", err.Error())
			blog.Errorf(errMsg)
			return nil, fmt.Errorf(errMsg)
		}
		blog.Errorf("CreateListener failed, err: %s", err.Error())
		return nil, rerr.Unwrap()
	}
	blog.V(3).Infof("CreateListener response: %s", common.ToJsonString(out))
	mf(metrics.LibCallStatusOK)
	return out, nil
}

// DescribeListeners describe listeners
func (sw *SdkWrapper) DescribeListeners(region string, input *elbv2.DescribeListenersInput, logV int32) (
	*elbv2.DescribeListenersOutput, error) {
	blog.V(glog.Level(logV)).Infof("DescribeListeners input: %s", common.ToJsonString(input))

	startTime := time.Now()
	// 统计API调用延时/状态
	mf := func(ret string) {
		defer metrics.ReportLibRequestMetric(
			SystemNameInMetricAWS,
			HandlerNameInMetricAWSSDK,
			"DescribeListeners", ret, startTime)
	}

	sw.tryThrottle()
	out, err := sw.getRegionClient(region).DescribeListeners(context.TODO(), input)
	if err != nil {
		rerr := ResolveError(err)
		if rerr.IsExceededAttemptError() {
			mf(metrics.LibCallStatusTimeout)
			errMsg := fmt.Sprintf("DescribeListeners out of maxRetry %d", maxRetry)
			blog.Errorf(errMsg)
			return nil, fmt.Errorf(errMsg)
		}
		if rerr.IsOperationError() {
			mf(metrics.LibCallStatusErr)
			errMsg := fmt.Sprintf("DescribeListeners failed, err %s", err.Error())
			blog.Errorf(errMsg)
			return nil, fmt.Errorf(errMsg)
		}
		blog.Errorf("DescribeListeners failed, err: %s", err.Error())
		return nil, rerr.Unwrap()
	}
	blog.V(glog.Level(logV)).Infof("DescribeListeners response: %s", common.ToJsonString(out))
	mf(metrics.LibCallStatusOK)
	return out, nil
}

// DeleteListener delete listener
func (sw *SdkWrapper) DeleteListener(region string, input *elbv2.DeleteListenerInput) (
	*elbv2.DeleteListenerOutput, error) {
	blog.V(3).Infof("DeleteListener input: %s", common.ToJsonString(input))

	startTime := time.Now()
	// 统计API调用延时/状态
	mf := func(ret string) {
		defer metrics.ReportLibRequestMetric(
			SystemNameInMetricAWS,
			HandlerNameInMetricAWSSDK,
			"DeleteListener", ret, startTime)
	}

	sw.tryThrottle()
	out, err := sw.getRegionClient(region).DeleteListener(context.TODO(), input)
	if err != nil {
		rerr := ResolveError(err)
		if rerr.IsExceededAttemptError() {
			mf(metrics.LibCallStatusTimeout)
			errMsg := fmt.Sprintf("DeleteListener out of maxRetry %d", maxRetry)
			blog.Errorf(errMsg)
			return nil, fmt.Errorf(errMsg)
		}
		if rerr.IsOperationError() {
			mf(metrics.LibCallStatusErr)
			errMsg := fmt.Sprintf("DeleteListener failed, err %s", err.Error())
			blog.Errorf(errMsg)
			return nil, fmt.Errorf(errMsg)
		}
		blog.Errorf("DeleteListener failed, err: %s", err.Error())
		return nil, rerr.Unwrap()
	}
	blog.V(3).Infof("DeleteListener response: %s", common.ToJsonString(out))
	mf(metrics.LibCallStatusOK)
	return out, nil
}

// ModifyListener modify listener
func (sw *SdkWrapper) ModifyListener(region string, input *elbv2.ModifyListenerInput) (
	*elbv2.ModifyListenerOutput, error) {
	blog.V(3).Infof("ModifyListener input: %s", common.ToJsonString(input))

	startTime := time.Now()
	// 统计API调用延时/状态
	mf := func(ret string) {
		defer metrics.ReportLibRequestMetric(
			SystemNameInMetricAWS,
			HandlerNameInMetricAWSSDK,
			"ModifyListener", ret, startTime)
	}

	sw.tryThrottle()
	out, err := sw.getRegionClient(region).ModifyListener(context.TODO(), input)
	if err != nil {
		rerr := ResolveError(err)
		if rerr.IsExceededAttemptError() {
			mf(metrics.LibCallStatusTimeout)
			errMsg := fmt.Sprintf("ModifyListener out of maxRetry %d", maxRetry)
			blog.Errorf(errMsg)
			return nil, fmt.Errorf(errMsg)
		}
		if rerr.IsOperationError() {
			mf(metrics.LibCallStatusErr)
			errMsg := fmt.Sprintf("ModifyListener failed, err %s", err.Error())
			blog.Errorf(errMsg)
			return nil, fmt.Errorf(errMsg)
		}
		blog.Errorf("ModifyListener failed, err: %s", err.Error())
		return nil, rerr.Unwrap()
	}
	blog.V(3).Infof("ModifyListener response: %s", common.ToJsonString(out))
	mf(metrics.LibCallStatusOK)
	return out, nil
}

// CreateRule create rule
func (sw *SdkWrapper) CreateRule(region string, input *elbv2.CreateRuleInput) (
	*elbv2.CreateRuleOutput, error) {
	blog.V(3).Infof("CreateRule input: %s", common.ToJsonString(input))

	startTime := time.Now()
	// 统计API调用延时/状态
	mf := func(ret string) {
		defer metrics.ReportLibRequestMetric(
			SystemNameInMetricAWS,
			HandlerNameInMetricAWSSDK,
			"CreateRule", ret, startTime)
	}

	sw.tryThrottle()
	out, err := sw.getRegionClient(region).CreateRule(context.TODO(), input)
	if err != nil {
		rerr := ResolveError(err)
		if rerr.IsExceededAttemptError() {
			mf(metrics.LibCallStatusTimeout)
			errMsg := fmt.Sprintf("CreateRule req[%s] out of maxRetry %d", common.ToJsonString(input), maxRetry)
			blog.Errorf(errMsg)
			return nil, fmt.Errorf(errMsg)
		}
		if rerr.IsOperationError() {
			mf(metrics.LibCallStatusErr)
			errMsg := fmt.Sprintf("CreateRule req[%s] failed, err %s", common.ToJsonString(input), err.Error())
			blog.Errorf(errMsg)
			return nil, fmt.Errorf(errMsg)
		}
		blog.Errorf("CreateRule req[%s] failed, err: %s", common.ToJsonString(input), err.Error())
		return nil, rerr.Unwrap()
	}
	blog.V(3).Infof("CreateRule response: %s", common.ToJsonString(out))
	mf(metrics.LibCallStatusOK)
	return out, nil
}

// DescribeRules describe rules
func (sw *SdkWrapper) DescribeRules(region string, input *elbv2.DescribeRulesInput, logV int32) (
	*elbv2.DescribeRulesOutput, error) {
	blog.V(glog.Level(logV)).Infof("DescribeRules input: %s", common.ToJsonString(input))

	startTime := time.Now()
	// 统计API调用延时/状态
	mf := func(ret string) {
		defer metrics.ReportLibRequestMetric(
			SystemNameInMetricAWS,
			HandlerNameInMetricAWSSDK,
			"DescribeRules", ret, startTime)
	}

	sw.tryThrottle()
	out, err := sw.getRegionClient(region).DescribeRules(context.TODO(), input)
	if err != nil {
		rerr := ResolveError(err)
		if rerr.IsExceededAttemptError() {
			mf(metrics.LibCallStatusTimeout)
			errMsg := fmt.Sprintf("DescribeRules req[%s] out of maxRetry %d", common.ToJsonString(input), maxRetry)
			blog.Errorf(errMsg)
			return nil, fmt.Errorf(errMsg)
		}
		if rerr.IsOperationError() {
			mf(metrics.LibCallStatusErr)
			errMsg := fmt.Sprintf("DescribeRules req[%s] failed, err %s", common.ToJsonString(input), err.Error())
			blog.Errorf(errMsg)
			return nil, fmt.Errorf(errMsg)
		}
		blog.Errorf("DescribeRules req[%s] failed, err: %s", common.ToJsonString(input), err.Error())
		return nil, rerr.Unwrap()
	}
	blog.V(glog.Level(logV)).Infof("DescribeRules response: %s", common.ToJsonString(out))
	mf(metrics.LibCallStatusOK)
	return out, nil
}

// DeleteRule delete rule
func (sw *SdkWrapper) DeleteRule(region string, input *elbv2.DeleteRuleInput) (
	*elbv2.DeleteRuleOutput, error) {
	blog.V(3).Infof("DeleteRule input: %s", common.ToJsonString(input))

	startTime := time.Now()
	// 统计API调用延时/状态
	mf := func(ret string) {
		defer metrics.ReportLibRequestMetric(
			SystemNameInMetricAWS,
			HandlerNameInMetricAWSSDK,
			"DeleteRule", ret, startTime)
	}

	sw.tryThrottle()
	out, err := sw.getRegionClient(region).DeleteRule(context.TODO(), input)
	if err != nil {
		rerr := ResolveError(err)
		if rerr.IsExceededAttemptError() {
			mf(metrics.LibCallStatusTimeout)
			errMsg := fmt.Sprintf("DeleteRule out of maxRetry %d", maxRetry)
			blog.Errorf(errMsg)
			return nil, fmt.Errorf(errMsg)
		}
		if rerr.IsOperationError() {
			mf(metrics.LibCallStatusErr)
			errMsg := fmt.Sprintf("DeleteRule failed, err %s", err.Error())
			blog.Errorf(errMsg)
			return nil, fmt.Errorf(errMsg)
		}
		blog.Errorf("DeleteRule failed, err: %s", err.Error())
		return nil, rerr.Unwrap()
	}
	blog.V(3).Infof("DeleteRule response: %s", common.ToJsonString(out))
	mf(metrics.LibCallStatusOK)
	return out, nil
}

// ModifyRule modify rule
func (sw *SdkWrapper) ModifyRule(region string, input *elbv2.ModifyRuleInput) (
	*elbv2.ModifyRuleOutput, error) {
	blog.V(3).Infof("ModifyRule input: %s", common.ToJsonString(input))

	startTime := time.Now()
	// 统计API调用延时/状态
	mf := func(ret string) {
		defer metrics.ReportLibRequestMetric(
			SystemNameInMetricAWS,
			HandlerNameInMetricAWSSDK,
			"ModifyRule", ret, startTime)
	}

	sw.tryThrottle()
	for i := range input.Conditions {
		input.Conditions[i].Values = nil
	}
	out, err := sw.getRegionClient(region).ModifyRule(context.TODO(), input)
	if err != nil {
		rerr := ResolveError(err)
		if rerr.IsExceededAttemptError() {
			mf(metrics.LibCallStatusTimeout)
			errMsg := fmt.Sprintf("ModifyRule out of maxRetry %d", maxRetry)
			blog.Errorf(errMsg)
			return nil, fmt.Errorf(errMsg)
		}
		if rerr.IsOperationError() {
			mf(metrics.LibCallStatusErr)
			errMsg := fmt.Sprintf("ModifyRule failed, err %s", err.Error())
			blog.Errorf(errMsg)
			return nil, fmt.Errorf(errMsg)
		}
		blog.Errorf("ModifyRule failed, err: %s", err.Error())
		return nil, rerr.Unwrap()
	}
	blog.V(3).Infof("ModifyRule response: %s", common.ToJsonString(out))
	mf(metrics.LibCallStatusOK)
	return out, nil
}

// CreateTargetGroup create target group
func (sw *SdkWrapper) CreateTargetGroup(region string, input *elbv2.CreateTargetGroupInput) (
	*elbv2.CreateTargetGroupOutput, error) {
	blog.V(3).Infof("CreateTargetGroup input: %s", common.ToJsonString(input))

	startTime := time.Now()
	// 统计API调用延时/状态
	mf := func(ret string) {
		defer metrics.ReportLibRequestMetric(
			SystemNameInMetricAWS,
			HandlerNameInMetricAWSSDK,
			"CreateTargetGroup", ret, startTime)
	}

	sw.tryThrottle()
	out, err := sw.getRegionClient(region).CreateTargetGroup(context.TODO(), input)
	if err != nil {
		rerr := ResolveError(err)
		if rerr.IsExceededAttemptError() {
			mf(metrics.LibCallStatusTimeout)
			errMsg := fmt.Sprintf("CreateTargetGroup out of maxRetry %d", maxRetry)
			blog.Errorf(errMsg)
			return nil, fmt.Errorf(errMsg)
		}
		if rerr.IsOperationError() {
			mf(metrics.LibCallStatusErr)
			errMsg := fmt.Sprintf("CreateTargetGroup failed, err %s", err.Error())
			blog.Errorf(errMsg)
			return nil, fmt.Errorf(errMsg)
		}
		blog.Errorf("CreateTargetGroup failed, err: %s", err.Error())
		return nil, rerr.Unwrap()
	}
	blog.V(3).Infof("CreateTargetGroup response: %s", common.ToJsonString(out))
	mf(metrics.LibCallStatusOK)
	return out, nil
}

// RegisterTargets register targets
func (sw *SdkWrapper) RegisterTargets(region string, input *elbv2.RegisterTargetsInput) (
	*elbv2.RegisterTargetsOutput, error) {
	blog.V(3).Infof("RegisterTargets input: %s", common.ToJsonString(input))

	startTime := time.Now()
	// 统计API调用延时/状态
	mf := func(ret string) {
		defer metrics.ReportLibRequestMetric(
			SystemNameInMetricAWS,
			HandlerNameInMetricAWSSDK,
			"RegisterTargets", ret, startTime)
	}

	sw.tryThrottle()
	out, err := sw.getRegionClient(region).RegisterTargets(context.TODO(), input)
	if err != nil {
		rerr := ResolveError(err)
		if rerr.IsExceededAttemptError() {
			mf(metrics.LibCallStatusTimeout)
			errMsg := fmt.Sprintf("RegisterTargets out of maxRetry %d", maxRetry)
			blog.Errorf(errMsg)
			return nil, fmt.Errorf(errMsg)
		}
		if rerr.IsOperationError() {
			mf(metrics.LibCallStatusErr)
			errMsg := fmt.Sprintf("RegisterTargets failed, err %s", err.Error())
			blog.Errorf(errMsg)
			return nil, fmt.Errorf(errMsg)
		}
		blog.Errorf("RegisterTargets failed, err: %s", err.Error())
		return nil, rerr.Unwrap()
	}
	blog.V(3).Infof("RegisterTargets response: %s", common.ToJsonString(out))
	mf(metrics.LibCallStatusOK)
	return out, nil
}

// DeregisterTargets deregister targets
func (sw *SdkWrapper) DeregisterTargets(region string, input *elbv2.DeregisterTargetsInput) (
	*elbv2.DeregisterTargetsOutput, error) {
	blog.V(3).Infof("DeregisterTargets input: %s", common.ToJsonString(input))

	startTime := time.Now()
	// 统计API调用延时/状态
	mf := func(ret string) {
		defer metrics.ReportLibRequestMetric(
			SystemNameInMetricAWS,
			HandlerNameInMetricAWSSDK,
			"DeregisterTargets", ret, startTime)
	}

	sw.tryThrottle()
	out, err := sw.getRegionClient(region).DeregisterTargets(context.TODO(), input)
	if err != nil {
		rerr := ResolveError(err)
		if rerr.IsExceededAttemptError() {
			mf(metrics.LibCallStatusTimeout)
			errMsg := fmt.Sprintf("DeregisterTargets out of maxRetry %d", maxRetry)
			blog.Errorf(errMsg)
			return nil, fmt.Errorf(errMsg)
		}
		if rerr.IsOperationError() {
			mf(metrics.LibCallStatusErr)
			errMsg := fmt.Sprintf("DeregisterTargets failed, err %s", err.Error())
			blog.Errorf(errMsg)
			return nil, fmt.Errorf(errMsg)
		}
		blog.Errorf("DeregisterTargets failed, err: %s", err.Error())
		return nil, rerr.Unwrap()
	}
	blog.V(3).Infof("DeregisterTargets response: %s", common.ToJsonString(out))
	mf(metrics.LibCallStatusOK)
	return out, nil
}

// DescribeTargetGroups describe target groups
func (sw *SdkWrapper) DescribeTargetGroups(region string, input *elbv2.DescribeTargetGroupsInput) (
	*elbv2.DescribeTargetGroupsOutput, error) {
	blog.V(3).Infof("DescribeTargetGroups input: %s", common.ToJsonString(input))

	startTime := time.Now()
	// 统计API调用延时/状态
	mf := func(ret string) {
		defer metrics.ReportLibRequestMetric(
			SystemNameInMetricAWS,
			HandlerNameInMetricAWSSDK,
			"DescribeTargetGroups", ret, startTime)
	}

	sw.tryThrottle()
	out, err := sw.getRegionClient(region).DescribeTargetGroups(context.TODO(), input)
	if err != nil {
		rerr := ResolveError(err)
		if rerr.IsExceededAttemptError() {
			mf(metrics.LibCallStatusTimeout)
			errMsg := fmt.Sprintf("DescribeTargetGroups out of maxRetry %d", maxRetry)
			blog.Errorf(errMsg)
			return nil, fmt.Errorf(errMsg)
		}
		if strings.Contains(err.Error(), "response error StatusCode: 400") {
			blog.Warnf("DescribeTargetGroups not found: %v, raw error: %s", input.Names, err.Error())
			return &elbv2.DescribeTargetGroupsOutput{}, nil
		}
		if rerr.IsOperationError() {
			mf(metrics.LibCallStatusErr)
			errMsg := fmt.Sprintf("DescribeTargetGroups failed, err %s", err.Error())
			blog.Errorf(errMsg)
			return nil, fmt.Errorf(errMsg)
		}
		blog.Errorf("DescribeTargetGroups failed, err: %s", err.Error())
		return nil, rerr.Unwrap()
	}
	blog.V(3).Infof("DescribeTargetGroups response: %s", common.ToJsonString(out))
	mf(metrics.LibCallStatusOK)
	return out, nil
}

// DeleteTargetGroup delete target group
func (sw *SdkWrapper) DeleteTargetGroup(region string, input *elbv2.DeleteTargetGroupInput) (
	*elbv2.DeleteTargetGroupOutput, error) {
	blog.V(3).Infof("DeleteTargetGroup input: %s", common.ToJsonString(input))

	startTime := time.Now()
	// 统计API调用延时/状态
	mf := func(ret string) {
		defer metrics.ReportLibRequestMetric(
			SystemNameInMetricAWS,
			HandlerNameInMetricAWSSDK,
			"DeleteTargetGroup", ret, startTime)
	}

	sw.tryThrottle()
	out, err := sw.getRegionClient(region).DeleteTargetGroup(context.TODO(), input)
	if err != nil {
		rerr := ResolveError(err)
		if rerr.IsExceededAttemptError() {
			mf(metrics.LibCallStatusTimeout)
			errMsg := fmt.Sprintf("DeleteTargetGroup out of maxRetry %d", maxRetry)
			blog.Errorf(errMsg)
			return nil, fmt.Errorf(errMsg)
		}
		if rerr.IsOperationError() {
			mf(metrics.LibCallStatusErr)
			errMsg := fmt.Sprintf("DeleteTargetGroup failed, err %s", err.Error())
			blog.Errorf(errMsg)
			return nil, fmt.Errorf(errMsg)
		}
		blog.Errorf("DeleteTargetGroup failed, err: %s", err.Error())
		return nil, rerr.Unwrap()
	}
	blog.V(3).Infof("DeleteTargetGroup response: %s", common.ToJsonString(out))
	mf(metrics.LibCallStatusOK)
	return out, nil
}

// DescribeTargetGroupAttributes describe target group attributes
func (sw *SdkWrapper) DescribeTargetGroupAttributes(region string, input *elbv2.DescribeTargetGroupAttributesInput) (
	*elbv2.DescribeTargetGroupAttributesOutput, error) {
	blog.V(3).Infof("DescribeTargetGroupAttributes input: %s", common.ToJsonString(input))

	startTime := time.Now()
	// 统计API调用延时/状态
	mf := func(ret string) {
		defer metrics.ReportLibRequestMetric(
			SystemNameInMetricAWS,
			HandlerNameInMetricAWSSDK,
			"DescribeTargetGroupAttributes", ret, startTime)
	}

	sw.tryThrottle()
	out, err := sw.getRegionClient(region).DescribeTargetGroupAttributes(context.TODO(), input)
	if err != nil {
		rerr := ResolveError(err)
		if rerr.IsExceededAttemptError() {
			mf(metrics.LibCallStatusTimeout)
			errMsg := fmt.Sprintf("DescribeTargetGroupAttributes out of maxRetry %d", maxRetry)
			blog.Errorf(errMsg)
			return nil, fmt.Errorf(errMsg)
		}
		if rerr.IsOperationError() {
			mf(metrics.LibCallStatusErr)
			errMsg := fmt.Sprintf("DescribeTargetGroupAttributes failed, err %s", err.Error())
			blog.Errorf(errMsg)
			return nil, fmt.Errorf(errMsg)
		}
		blog.Errorf("DescribeTargetGroupAttributes failed, err: %s", err.Error())
		return nil, rerr.Unwrap()
	}
	blog.V(3).Infof("DescribeTargetGroupAttributes response: %s", common.ToJsonString(out))
	mf(metrics.LibCallStatusOK)
	return out, nil
}

// ModifyTargetGroup modify target group
func (sw *SdkWrapper) ModifyTargetGroup(region string, input *elbv2.ModifyTargetGroupInput) (
	*elbv2.ModifyTargetGroupOutput, error) {
	blog.V(3).Infof("ModifyTargetGroup input: %s", common.ToJsonString(input))

	startTime := time.Now()
	// 统计API调用延时/状态
	mf := func(ret string) {
		defer metrics.ReportLibRequestMetric(
			SystemNameInMetricAWS,
			HandlerNameInMetricAWSSDK,
			"ModifyTargetGroup", ret, startTime)
	}

	sw.tryThrottle()
	out, err := sw.getRegionClient(region).ModifyTargetGroup(context.TODO(), input)
	if err != nil {
		rerr := ResolveError(err)
		if rerr.IsExceededAttemptError() {
			mf(metrics.LibCallStatusTimeout)
			errMsg := fmt.Sprintf("ModifyTargetGroup out of maxRetry %d", maxRetry)
			blog.Errorf(errMsg)
			return nil, fmt.Errorf(errMsg)
		}
		if rerr.IsOperationError() {
			mf(metrics.LibCallStatusErr)
			errMsg := fmt.Sprintf("ModifyTargetGroup failed, err %s", err.Error())
			blog.Errorf(errMsg)
			return nil, fmt.Errorf(errMsg)
		}
		blog.Errorf("ModifyTargetGroup failed, err: %s", err.Error())
		return nil, rerr.Unwrap()
	}
	blog.V(3).Infof("ModifyTargetGroup response: %s", common.ToJsonString(out))
	mf(metrics.LibCallStatusOK)
	return out, nil
}

// ModifyTargetGroupAttributes modify target group attributes
func (sw *SdkWrapper) ModifyTargetGroupAttributes(region string, input *elbv2.ModifyTargetGroupAttributesInput) (
	*elbv2.ModifyTargetGroupAttributesOutput, error) {
	blog.V(3).Infof("ModifyTargetGroupAttributes input: %s", common.ToJsonString(input))

	startTime := time.Now()
	// 统计API调用延时/状态
	mf := func(ret string) {
		defer metrics.ReportLibRequestMetric(
			SystemNameInMetricAWS,
			HandlerNameInMetricAWSSDK,
			"ModifyTargetGroupAttributes", ret, startTime)
	}

	sw.tryThrottle()
	out, err := sw.getRegionClient(region).ModifyTargetGroupAttributes(context.TODO(), input)
	if err != nil {
		rerr := ResolveError(err)
		if rerr.IsExceededAttemptError() {
			mf(metrics.LibCallStatusTimeout)
			errMsg := fmt.Sprintf("ModifyTargetGroupAttributes out of maxRetry %d", maxRetry)
			blog.Errorf(errMsg)
			return nil, fmt.Errorf(errMsg)
		}
		if rerr.IsOperationError() {
			mf(metrics.LibCallStatusErr)
			errMsg := fmt.Sprintf("ModifyTargetGroupAttributes failed, err %s", err.Error())
			blog.Errorf(errMsg)
			return nil, fmt.Errorf(errMsg)
		}
		blog.Errorf("ModifyTargetGroupAttributes failed, err: %s", err.Error())
		return nil, rerr.Unwrap()
	}
	blog.V(3).Infof("ModifyTargetGroupAttributes response: %s", common.ToJsonString(out))
	mf(metrics.LibCallStatusOK)
	return out, nil
}

// DescribeTargetHealth describe target health
func (sw *SdkWrapper) DescribeTargetHealth(region string, input *elbv2.DescribeTargetHealthInput) (
	*elbv2.DescribeTargetHealthOutput, error) {
	// 定时调用，log v4避免日志量过大
	blog.V(4).Infof("DescribeTargetHealth input: %s", common.ToJsonString(input))

	startTime := time.Now()
	// 统计API调用延时/状态
	mf := func(ret string) {
		defer metrics.ReportLibRequestMetric(
			SystemNameInMetricAWS,
			HandlerNameInMetricAWSSDK,
			"DescribeTargetHealth", ret, startTime)
	}

	sw.tryThrottle()
	out, err := sw.getRegionClient(region).DescribeTargetHealth(context.TODO(), input)
	if err != nil {
		rerr := ResolveError(err)
		if rerr.IsExceededAttemptError() {
			mf(metrics.LibCallStatusTimeout)
			errMsg := fmt.Sprintf("DescribeTargetHealth out of maxRetry %d", maxRetry)
			blog.Errorf(errMsg)
			return nil, fmt.Errorf(errMsg)
		}
		if rerr.IsOperationError() {
			mf(metrics.LibCallStatusErr)
			errMsg := fmt.Sprintf("DescribeTargetHealth failed, err %s", err.Error())
			blog.Errorf(errMsg)
			return nil, fmt.Errorf(errMsg)
		}
		blog.Errorf("DescribeTargetHealth failed, err: %s", err.Error())
		return nil, rerr.Unwrap()
	}
	blog.V(4).Infof("DescribeTargetHealth response: %s", common.ToJsonString(out))
	mf(metrics.LibCallStatusOK)
	return out, nil
}

func (sw *SdkWrapper) getLbFromCache(region, lbArn string) (types.LoadBalancer, error) {
	var lb types.LoadBalancer
	lbI, found := sw.lbCache.Get(genRegionArn(region, lbArn))
	if !found {
		lbs, err := sw.DescribeLoadBalancers(region,
			&elbv2.DescribeLoadBalancersInput{LoadBalancerArns: []string{lbArn}})
		if err != nil {
			return lb, fmt.Errorf("DescribeLoadBalancers failed, %s", err.Error())
		}
		if len(lbs.LoadBalancers) == 0 {
			return lb, cloud.ErrLoadbalancerNotFound
		}
		lb = lbs.LoadBalancers[0]
	} else {
		lb = lbI.(types.LoadBalancer)
	}
	return lb, nil
}
