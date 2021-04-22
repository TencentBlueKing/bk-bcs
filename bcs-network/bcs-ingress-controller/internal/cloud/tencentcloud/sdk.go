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

package tencentcloud

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"time"

	tclb "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/clb/v20180317"
	tcommon "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	terrors "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	tprofile "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/throttle"
	"github.com/Tencent/bk-bcs/bcs-network/bcs-ingress-controller/internal/metrics"
)

const (
	// RequestLimitExceededCode code for request exceeded limit
	RequestLimitExceededCode = "RequestLimitExceeded"
	// WrongStatusCode code for incorrect status
	WrongStatusCode = "InternalError"
	// TaskStatusDealing  task is dealing
	TaskStatusDealing = 2
	// TaskStatusFailed task is failed
	TaskStatusFailed = 1
	// TaskStatusSucceed task is successful
	TaskStatusSucceed = 0

	// EnvNameTencentCloudClbDomain env name of tencent cloud clb domain
	EnvNameTencentCloudClbDomain = "TENCENTCLOUD_CLB_DOMAIN"
	// EnvNameTencentCloudRegion env name of tencent cloud region
	EnvNameTencentCloudRegion = "TENCENTCLOUD_REGION"
	// EnvNameTencentCloudAccessKeyID env name of tencent cloud access key id
	EnvNameTencentCloudAccessKeyID = "TENCENTCLOUD_ACCESS_KEY_ID"
	// EnvNameTencentCloudAccessKey env name of tencent cloud secret key
	EnvNameTencentCloudAccessKey = "TENCENTCLOUD_ACESS_KEY"

	// EnvNameTencentCloudRateLimitQPS env name for tencent cloud api rate limit qps
	EnvNameTencentCloudRateLimitQPS = "TENCENTCLOUD_RATELIMIT_QPS"
	// EnvNameTencentCloudRateLimitBucketSize env name for tencent cloud api rate limit bucket size
	EnvNameTencentCloudRateLimitBucketSize = "TENCENTCLOUD_RATELIMIT_BUCKET_SIZE"
)

var (
	// If the delay caused by the frequency limit exceeds this value, it is recorded in the log
	maxLatency = 120 * time.Millisecond
	// the maximum number of retries caused by server error or API overrun
	maxRetry = 25
	// qps for rate limit
	defaultThrottleQPS = 50
	// bucket size for rate limit
	defaultBucketSize = 50
	// wait seconds when tencent cloud api is busy
	waitPeriodLBDealing = 2
)

// SdkWrapper wrapper for tencentcloud sdk
type SdkWrapper struct {
	// domain for tencent cloud clb service
	domain string
	// secret id for tencent cloud account
	secretID string
	// secret key for tencent cloud account
	secretKey string
	// client profile for tencent cloud sdk
	cpf *tprofile.ClientProfile
	// credential for tencent cloud sdk
	credential *tcommon.Credential

	ratelimitqps        int64
	ratelimitbucketSize int64
	// rate limiter for calling sdk
	throttler throttle.RateLimiter
	// map of client for different region
	// for ingress controller, it may control different cloud loadbalancer in different regions
	clbCliMap map[string]*tclb.Client
}

// NewSdkWrapper create sdk wrapper
func NewSdkWrapper() (*SdkWrapper, error) {
	sw := &SdkWrapper{}
	err := sw.loadEnv()
	if err != nil {
		return nil, err
	}

	credential := tcommon.NewCredential(
		sw.secretID,
		sw.secretKey,
	)
	cpf := tprofile.NewClientProfile()
	if len(sw.domain) != 0 {
		cpf.HttpProfile.Endpoint = sw.domain
	}
	sw.credential = credential
	sw.cpf = cpf
	sw.clbCliMap = make(map[string]*tclb.Client)

	sw.throttler = throttle.NewTokenBucket(sw.ratelimitqps, sw.ratelimitbucketSize)
	return sw, nil
}

// NewSdkWrapperWithSecretIDKey create sdk wrapper with secret id and secret key
func NewSdkWrapperWithSecretIDKey(id, key string) (*SdkWrapper, error) {
	sw := &SdkWrapper{}
	err := sw.loadEnv()
	if err != nil {
		return nil, err
	}

	sw.secretID = id
	sw.secretKey = key

	credential := tcommon.NewCredential(
		sw.secretID,
		sw.secretKey,
	)
	cpf := tprofile.NewClientProfile()
	if len(sw.domain) != 0 {
		cpf.HttpProfile.Endpoint = sw.domain
	}
	sw.credential = credential
	sw.cpf = cpf
	sw.clbCliMap = make(map[string]*tclb.Client)

	sw.throttler = throttle.NewTokenBucket(sw.ratelimitqps, sw.ratelimitbucketSize)
	return sw, nil
}

// load config for environment variables
func (sw *SdkWrapper) loadEnv() error {
	clbDomain := os.Getenv(EnvNameTencentCloudClbDomain)
	secretID := os.Getenv(EnvNameTencentCloudAccessKeyID)
	secretKey := os.Getenv(EnvNameTencentCloudAccessKey)
	sw.domain = clbDomain
	sw.secretID = secretID
	sw.secretKey = secretKey

	qpsStr := os.Getenv(EnvNameTencentCloudRateLimitQPS)
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

	bucketSizeStr := os.Getenv(EnvNameTencentCloudRateLimitBucketSize)
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

// getRegionClient create region client
func (sw *SdkWrapper) getRegionClient(region string) (*tclb.Client, error) {
	cli, ok := sw.clbCliMap[region]
	if !ok {
		newCli, err := tclb.NewClient(sw.credential, region, sw.cpf)
		if err != nil {
			blog.Errorf("create clb client for region %s failed, err %s", region, err.Error())
			return nil, fmt.Errorf("create clb client for region %s failed, err %s", region, err.Error())
		}
		sw.clbCliMap[region] = newCli
		return newCli, nil
	}
	return cli, nil
}

// checkErrCode common method for check tencent cloud sdk err
func (sw *SdkWrapper) checkErrCode(err *terrors.TencentCloudSDKError) {
	if err.Code == RequestLimitExceededCode {
		blog.Warnf("request exceed limit, have a rest for %d second", waitPeriodLBDealing)
		time.Sleep(time.Duration(waitPeriodLBDealing) * time.Second)
	} else if err.Code == WrongStatusCode {
		blog.Warnf("clb is dealing another action, have a rest for %d second", waitPeriodLBDealing)
		time.Sleep(time.Duration(waitPeriodLBDealing) * time.Second)
	}
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

// waitTaskDone wait asynchronous task done
func (sw *SdkWrapper) waitTaskDone(region, taskID string) error {
	blog.V(3).Infof("start waiting for task %s", taskID)
	request := tclb.NewDescribeTaskStatusRequest()
	request.TaskId = tcommon.StringPtr(taskID)
	blog.Infof("describe task status request:\n%s", request.ToJsonString())
	for counter := 0; counter < maxRetry; counter++ {
		// it may exceed limit when describe task result
		sw.tryThrottle()
		clbCli, err := sw.getRegionClient(region)
		if err != nil {
			return err
		}
		response, err := clbCli.DescribeTaskStatus(request)
		if err != nil {
			if terr, ok := err.(*terrors.TencentCloudSDKError); ok {
				sw.checkErrCode(terr)
				if terr.Code == RequestLimitExceededCode || terr.Code == WrongStatusCode {
					continue
				}
			}
			blog.Errorf("describe task status failed, err %s", err.Error())
			return fmt.Errorf("describe task status failed, err %s", err.Error())
		}
		blog.Infof("describe task status response:\n%s", response.ToJsonString())
		// dealing
		if *response.Response.Status == TaskStatusDealing {
			blog.Infof("task %s is dealing", taskID)
			time.Sleep(time.Duration(waitPeriodLBDealing) * time.Second)
			continue
			// failed
		} else if *response.Response.Status == TaskStatusFailed {
			blog.Errorf("task %s is failed", taskID)
			return fmt.Errorf("task %s is failed", taskID)
			// succeed
		} else if *response.Response.Status == TaskStatusSucceed {
			blog.Infof("task %s is done", taskID)
			return nil
		}
		return fmt.Errorf("error status of task %d", *response.Response.Status)
	}
	blog.Errorf("describe task status with request %s timeout", request.ToJsonString())
	return fmt.Errorf("describe task status with request %s timeout", request.ToJsonString())
}

// DescribeLoadBalancers wrap DescribeLoadBalancers
func (sw *SdkWrapper) DescribeLoadBalancers(region string, req *tclb.DescribeLoadBalancersRequest) (
	*tclb.DescribeLoadBalancersResponse, error) {

	blog.V(3).Infof("DescribeLoadBalancers request: %s", req.ToJsonString())
	var err error
	var resp *tclb.DescribeLoadBalancersResponse

	startTime := time.Now()
	mf := func(ret string) {
		metrics.ReportLibRequestMetric(
			SystemNameInMetricTencentCloud,
			HandlerNameInMetricTencentCloudSDK,
			"DescribeLoadBalancers", ret, startTime)
	}

	counter := 1
	for ; counter <= maxRetry; counter++ {
		blog.V(3).Infof("DescribeLoadBalancers try %d/%d", counter, maxRetry)
		sw.tryThrottle()
		// get client by region
		clbCli, inErr := sw.getRegionClient(region)
		if inErr != nil {
			mf(metrics.LibCallStatusErr)
			return nil, inErr
		}
		resp, err = clbCli.DescribeLoadBalancers(req)
		if err != nil {
			if terr, ok := err.(*terrors.TencentCloudSDKError); ok {
				sw.checkErrCode(terr)
				if terr.Code == RequestLimitExceededCode || terr.Code == WrongStatusCode {
					continue
				}
			}
			mf(metrics.LibCallStatusErr)
			blog.Errorf("DescribeLoadBalancers failed, err %s", err.Error())
			return nil, fmt.Errorf("DescribeLoadBalancers failed, err %s", err.Error())
		}
		blog.V(3).Infof("DescribeLoadBalancers response: %s", resp.ToJsonString())
		break
	}
	if counter > maxRetry {
		mf(metrics.LibCallStatusTimeout)
		blog.Errorf("DescribeLoadBalancers out of maxRetry %d", maxRetry)
		return nil, fmt.Errorf("DescribeLoadBalancers out of maxRetry %d", maxRetry)
	}
	mf(metrics.LibCallStatusOK)
	return resp, nil
}

// CreateListener wrap CreateListener, length of Ports should be less than 50
func (sw *SdkWrapper) CreateListener(region string, req *tclb.CreateListenerRequest) ([]string, error) {
	blog.V(3).Infof("CreateListener request: %s", req.ToJsonString())
	var err error
	var resp *tclb.CreateListenerResponse

	startTime := time.Now()
	mf := func(ret string) {
		defer metrics.ReportLibRequestMetric(
			SystemNameInMetricTencentCloud,
			HandlerNameInMetricTencentCloudSDK,
			"CreateListener", ret, startTime)
	}

	counter := 1
	for ; counter <= maxRetry; counter++ {
		blog.V(3).Infof("CreateListener try %d/%d", counter, maxRetry)
		sw.tryThrottle()
		// get client by region
		clbCli, inErr := sw.getRegionClient(region)
		if inErr != nil {
			mf(metrics.LibCallStatusErr)
			return nil, inErr
		}
		resp, err = clbCli.CreateListener(req)
		if err != nil {
			if terr, ok := err.(*terrors.TencentCloudSDKError); ok {
				sw.checkErrCode(terr)
				if terr.Code == RequestLimitExceededCode || terr.Code == WrongStatusCode {
					continue
				}
			}
			mf(metrics.LibCallStatusErr)
			blog.Errorf("CreateListener failed, err %s", err.Error())
			return nil, fmt.Errorf("CreateListener failed, err %s", err.Error())
		}
		if len(resp.Response.ListenerIds) == 0 {
			mf(metrics.LibCallStatusErr)
			blog.Errorf("create listener return zero length ids")
			return nil, fmt.Errorf("create listener return zero length ids")
		}
		blog.V(3).Infof("CreateListener response: %s", resp.ToJsonString())
		break
	}
	if counter > maxRetry {
		mf(metrics.LibCallStatusTimeout)
		blog.Errorf("CreateListener out of maxRetry %d", maxRetry)
		return nil, fmt.Errorf("CreateListener out of maxRetry %d", maxRetry)
	}
	err = sw.waitTaskDone(region, *resp.Response.RequestId)
	if err != nil {
		mf(metrics.LibCallStatusErr)
		return nil, err
	}
	mf(metrics.LibCallStatusOK)
	var retIDs []string
	for _, id := range resp.Response.ListenerIds {
		retIDs = append(retIDs, *id)
	}
	return retIDs, nil
}

// DescribeListeners wrap DescribeListeners
func (sw *SdkWrapper) DescribeListeners(region string, req *tclb.DescribeListenersRequest) (
	*tclb.DescribeListenersResponse, error) {

	blog.V(3).Infof("DescribeListeners request: %s", req.ToJsonString())
	var err error
	var resp *tclb.DescribeListenersResponse

	startTime := time.Now()
	mf := func(ret string) {
		metrics.ReportLibRequestMetric(
			SystemNameInMetricTencentCloud,
			HandlerNameInMetricTencentCloudSDK,
			"DescribeListeners", ret, startTime)
	}

	counter := 1
	for ; counter <= maxRetry; counter++ {
		blog.V(3).Infof("DescribeListeners try %d/%d", counter, maxRetry)
		sw.tryThrottle()
		// get client by region
		clbCli, inErr := sw.getRegionClient(region)
		if inErr != nil {
			mf(metrics.LibCallStatusErr)
			return nil, inErr
		}
		resp, err = clbCli.DescribeListeners(req)
		if err != nil {
			if terr, ok := err.(*terrors.TencentCloudSDKError); ok {
				sw.checkErrCode(terr)
				if terr.Code == RequestLimitExceededCode || terr.Code == WrongStatusCode {
					continue
				}
			}
			mf(metrics.LibCallStatusErr)
			blog.Errorf("DescribeListeners failed, err %s", err.Error())
			return nil, fmt.Errorf("DescribeListeners failed, err %s", err.Error())
		}
		blog.V(3).Infof("DescribeListeners response: %s", resp.ToJsonString())
		break
	}
	if counter > maxRetry {
		mf(metrics.LibCallStatusTimeout)
		blog.Errorf("DescribeListeners out of maxRetry %d", maxRetry)
		return nil, fmt.Errorf("DescribeListeners out of maxRetry %d", maxRetry)
	}
	mf(metrics.LibCallStatusOK)
	return resp, nil
}

// DescribeTargets wrap DescribeTargets
func (sw *SdkWrapper) DescribeTargets(region string, req *tclb.DescribeTargetsRequest) (
	*tclb.DescribeTargetsResponse, error) {

	blog.V(3).Infof("DescribeTargets request: %s", req.ToJsonString())
	var err error
	var resp *tclb.DescribeTargetsResponse

	startTime := time.Now()
	mf := func(ret string) {
		metrics.ReportLibRequestMetric(
			SystemNameInMetricTencentCloud,
			HandlerNameInMetricTencentCloudSDK,
			"DescribeTargets", ret, startTime)
	}

	counter := 1
	for ; counter <= maxRetry; counter++ {
		blog.V(3).Infof("DescribeTargets try %d/%d", counter, maxRetry)
		sw.tryThrottle()
		// get client by region
		clbCli, inErr := sw.getRegionClient(region)
		if inErr != nil {
			mf(metrics.LibCallStatusErr)
			return nil, inErr
		}
		resp, err = clbCli.DescribeTargets(req)
		if err != nil {
			if terr, ok := err.(*terrors.TencentCloudSDKError); ok {
				sw.checkErrCode(terr)
				if terr.Code == RequestLimitExceededCode || terr.Code == WrongStatusCode {
					continue
				}
			}
			mf(metrics.LibCallStatusErr)
			blog.Errorf("DescribeTargets failed, err %s", err.Error())
			return nil, fmt.Errorf("DescribeTargets failed, err %s", err.Error())
		}
		blog.V(3).Infof("DescribeTargets response: %s", resp.ToJsonString())
		break
	}
	if counter > maxRetry {
		mf(metrics.LibCallStatusErr)
		blog.Errorf("DescribeTargets out of maxRetry %d", maxRetry)
		return nil, fmt.Errorf("DescribeTargets out of maxRetry %d", maxRetry)
	}
	mf(metrics.LibCallStatusOK)
	return resp, nil
}

// DeleteListener wrap DeleteListener
func (sw *SdkWrapper) DeleteListener(region string, req *tclb.DeleteListenerRequest) error {
	blog.V(3).Infof("DeleteListener request: %s", req.ToJsonString())
	var err error
	var resp *tclb.DeleteListenerResponse

	startTime := time.Now()
	mf := func(ret string) {
		metrics.ReportLibRequestMetric(
			SystemNameInMetricTencentCloud,
			HandlerNameInMetricTencentCloudSDK,
			"DeleteListener", ret, startTime)
	}

	counter := 1
	for ; counter <= maxRetry; counter++ {
		blog.V(3).Infof("DeleteListener try %d/%d", counter, maxRetry)
		sw.tryThrottle()
		// get client by region
		clbCli, inErr := sw.getRegionClient(region)
		if inErr != nil {
			mf(metrics.LibCallStatusErr)
			return inErr
		}
		resp, err = clbCli.DeleteListener(req)
		if err != nil {
			if terr, ok := err.(*terrors.TencentCloudSDKError); ok {
				sw.checkErrCode(terr)
				if terr.Code == RequestLimitExceededCode || terr.Code == WrongStatusCode {
					continue
				}
			}
			mf(metrics.LibCallStatusErr)
			blog.Errorf("DeleteListener failed, err %s", err.Error())
			return fmt.Errorf("DeleteListener failed, err %s", err.Error())
		}
		blog.V(3).Infof("DeleteListener response: %s", resp.ToJsonString())
		break
	}
	if counter > maxRetry {
		mf(metrics.LibCallStatusTimeout)
		blog.Errorf("DeleteListener out of maxRetry %d", maxRetry)
		return fmt.Errorf("DeleteListener out of maxRetry %d", maxRetry)
	}
	err = sw.waitTaskDone(region, *resp.Response.RequestId)
	if err != nil {
		mf(metrics.LibCallStatusErr)
		return err
	}
	mf(metrics.LibCallStatusOK)
	return nil
}

// DeleteLoadbalanceListenners delete multiple listener
func (sw *SdkWrapper) DeleteLoadbalanceListenners(region string, req *tclb.DeleteLoadBalancerListenersRequest) error {
	rounds := len(req.ListenerIds) / MaxListenersForDeleteEachTime
	remains := len(req.ListenerIds) % MaxListenersForDeleteEachTime

	index := 0
	for ; index <= rounds; index++ {
		start := index * MaxListenersForDeleteEachTime
		end := (index + 1) * MaxListenersForDeleteEachTime
		if index == rounds {
			end = start + remains
		}
		newReq := tclb.NewDeleteLoadBalancerListenersRequest()
		newReq.LoadBalancerId = req.LoadBalancerId
		newReq.ListenerIds = req.ListenerIds[start:end]
		if err := sw.doDeleteLoadbalanceListenners(region, newReq); err != nil {
			return err
		}
	}
	return nil
}

// do delete multiple listener
func (sw *SdkWrapper) doDeleteLoadbalanceListenners(region string, req *tclb.DeleteLoadBalancerListenersRequest) error {
	blog.V(3).Infof("DeleteLoadbalanceListenners request: %s", req.ToJsonString())
	var err error
	var resp *tclb.DeleteLoadBalancerListenersResponse

	startTime := time.Now()
	mf := func(ret string) {
		metrics.ReportLibRequestMetric(
			SystemNameInMetricTencentCloud,
			HandlerNameInMetricTencentCloudSDK,
			"DeleteLoadbalanceListenners", ret, startTime)
	}

	counter := 1
	for ; counter <= maxRetry; counter++ {
		blog.V(3).Infof("DeleteLoadbalanceListenners try %d/%d", counter, maxRetry)
		sw.tryThrottle()
		clbCli, inErr := sw.getRegionClient(region)
		if inErr != nil {
			mf(metrics.LibCallStatusErr)
			return inErr
		}
		resp, err = clbCli.DeleteLoadBalancerListeners(req)
		if err != nil {
			if terr, ok := err.(*terrors.TencentCloudSDKError); ok {
				sw.checkErrCode(terr)
				if terr.Code == RequestLimitExceededCode || terr.Code == WrongStatusCode {
					continue
				}
			}
			mf(metrics.LibCallStatusErr)
			blog.Errorf("DeleteLoadbalanceListenners failed, err %s", err.Error())
			return fmt.Errorf("DeleteLoadbalanceListenners failed, err %s", err.Error())
		}
		blog.V(3).Infof("DeleteLoadbalanceListenners response: %s", resp.ToJsonString())
		break
	}
	if counter > maxRetry {
		mf(metrics.LibCallStatusErr)
		blog.Errorf("DeleteLoadbalanceListenners out of maxRetry %d", maxRetry)
		return fmt.Errorf("DeleteLoadbalanceListenners out of maxRetry %d", maxRetry)
	}
	err = sw.waitTaskDone(region, *resp.Response.RequestId)
	if err != nil {
		mf(metrics.LibCallStatusErr)
		return err
	}
	mf(metrics.LibCallStatusOK)
	return nil
}

// CreateRule wrap CreateRule
func (sw *SdkWrapper) CreateRule(region string, req *tclb.CreateRuleRequest) ([]string, error) {
	blog.V(3).Infof("CreateRule request: %s", req.ToJsonString())
	var err error
	var resp *tclb.CreateRuleResponse

	startTime := time.Now()
	mf := func(ret string) {
		metrics.ReportLibRequestMetric(
			SystemNameInMetricTencentCloud,
			HandlerNameInMetricTencentCloudSDK,
			"CreateRule", ret, startTime)
	}

	counter := 1
	for ; counter <= maxRetry; counter++ {
		blog.V(3).Infof("CreateRule try %d/%d", counter, maxRetry)
		sw.tryThrottle()
		// get client by region
		clbCli, inErr := sw.getRegionClient(region)
		if inErr != nil {
			mf(metrics.LibCallStatusErr)
			return nil, inErr
		}
		resp, err = clbCli.CreateRule(req)
		if err != nil {
			if terr, ok := err.(*terrors.TencentCloudSDKError); ok {
				sw.checkErrCode(terr)
				if terr.Code == RequestLimitExceededCode || terr.Code == WrongStatusCode {
					continue
				}
			}
			mf(metrics.LibCallStatusErr)
			blog.Errorf("CreateRule failed, err %s", err.Error())
			return nil, fmt.Errorf("CreateRule failed, err %s", err.Error())
		}
		blog.V(3).Infof("CreateRule response: %s", resp.ToJsonString())
		break
	}
	if counter > maxRetry {
		mf(metrics.LibCallStatusErr)
		blog.Errorf("CreateRule out of maxRetry %d", maxRetry)
		return nil, fmt.Errorf("CreateRule out of maxRetry %d", maxRetry)
	}
	err = sw.waitTaskDone(region, *resp.Response.RequestId)
	if err != nil {
		mf(metrics.LibCallStatusErr)
		return nil, err
	}
	mf(metrics.LibCallStatusOK)
	return tcommon.StringValues(resp.Response.LocationIds), nil
}

// DeleteRule wrap DeleteRule
func (sw *SdkWrapper) DeleteRule(region string, req *tclb.DeleteRuleRequest) error {
	blog.V(3).Infof("DeleteRule request: %s", req.ToJsonString())
	var err error
	var resp *tclb.DeleteRuleResponse

	startTime := time.Now()
	mf := func(ret string) {
		metrics.ReportLibRequestMetric(
			SystemNameInMetricTencentCloud,
			HandlerNameInMetricTencentCloudSDK,
			"DeleteRule", ret, startTime)
	}

	counter := 1
	for ; counter <= maxRetry; counter++ {
		blog.V(3).Infof("DeleteRule try %d/%d", counter, maxRetry)
		sw.tryThrottle()
		// get client by region
		clbCli, inErr := sw.getRegionClient(region)
		if inErr != nil {
			mf(metrics.LibCallStatusErr)
			return inErr
		}
		resp, err = clbCli.DeleteRule(req)
		if err != nil {
			if terr, ok := err.(*terrors.TencentCloudSDKError); ok {
				sw.checkErrCode(terr)
				if terr.Code == RequestLimitExceededCode || terr.Code == WrongStatusCode {
					continue
				}
			}
			mf(metrics.LibCallStatusErr)
			blog.Errorf("DeleteRule failed, err %s", err.Error())
			return fmt.Errorf("DeleteRule failed, err %s", err.Error())
		}
		blog.V(3).Infof("DeleteRule response: %s", resp.ToJsonString())
		break
	}
	if counter > maxRetry {
		mf(metrics.LibCallStatusTimeout)
		blog.Errorf("DeleteRule out of maxRetry %d", maxRetry)
		return fmt.Errorf("DeleteRule out of maxRetry %d", maxRetry)
	}
	err = sw.waitTaskDone(region, *resp.Response.RequestId)
	if err != nil {
		mf(metrics.LibCallStatusErr)
		return err
	}
	mf(metrics.LibCallStatusOK)
	return nil
}

// ModifyRule wrap ModifyRule
func (sw *SdkWrapper) ModifyRule(region string, req *tclb.ModifyRuleRequest) error {
	blog.V(3).Infof("ModifyRule request: %s", req.ToJsonString())
	var err error
	var resp *tclb.ModifyRuleResponse

	startTime := time.Now()
	mf := func(ret string) {
		metrics.ReportLibRequestMetric(
			SystemNameInMetricTencentCloud,
			HandlerNameInMetricTencentCloudSDK,
			"ModifyRule", ret, startTime)
	}

	counter := 1
	for ; counter <= maxRetry; counter++ {
		blog.V(3).Infof("ModifyRule try %d/%d", counter, maxRetry)
		sw.tryThrottle()
		// get client by region
		clbCli, inErr := sw.getRegionClient(region)
		if inErr != nil {
			mf(metrics.LibCallStatusErr)
			return inErr
		}
		resp, err = clbCli.ModifyRule(req)
		if err != nil {
			if terr, ok := err.(*terrors.TencentCloudSDKError); ok {
				sw.checkErrCode(terr)
				if terr.Code == RequestLimitExceededCode || terr.Code == WrongStatusCode {
					continue
				}
			}
			mf(metrics.LibCallStatusErr)
			blog.Errorf("ModifyRule failed, err %s", err.Error())
			return fmt.Errorf("ModifyRule failed, err %s", err.Error())
		}
		blog.V(3).Infof("ModifyRule response: %s", resp.ToJsonString())
		break
	}
	if counter > maxRetry {
		mf(metrics.LibCallStatusTimeout)
		blog.Errorf("ModifyRule out of maxRetry %d", maxRetry)
		return fmt.Errorf("ModifyRule out of maxRetry %d", maxRetry)
	}
	err = sw.waitTaskDone(region, *resp.Response.RequestId)
	if err != nil {
		mf(metrics.LibCallStatusErr)
		return err
	}
	mf(metrics.LibCallStatusOK)
	return nil
}

// ModifyListener wrap ModifyListener
func (sw *SdkWrapper) ModifyListener(region string, req *tclb.ModifyListenerRequest) error {
	blog.V(3).Infof("ModifyListener request: %s", req.ToJsonString())
	var err error
	var resp *tclb.ModifyListenerResponse

	startTime := time.Now()
	mf := func(ret string) {
		metrics.ReportLibRequestMetric(
			SystemNameInMetricTencentCloud,
			HandlerNameInMetricTencentCloudSDK,
			"ModifyListener", ret, startTime)
	}

	counter := 1
	for ; counter <= maxRetry; counter++ {
		blog.V(3).Infof("ModifyListener try %d/%d", counter, maxRetry)
		sw.tryThrottle()
		// get client by region
		clbCli, inErr := sw.getRegionClient(region)
		if inErr != nil {
			mf(metrics.LibCallStatusErr)
			return inErr
		}
		resp, err = clbCli.ModifyListener(req)
		if err != nil {
			if terr, ok := err.(*terrors.TencentCloudSDKError); ok {
				sw.checkErrCode(terr)
				if terr.Code == RequestLimitExceededCode || terr.Code == WrongStatusCode {
					continue
				}
			}
			mf(metrics.LibCallStatusErr)
			blog.Errorf("ModifyListener failed, err %s", err.Error())
			return fmt.Errorf("ModifyListener failed, err %s", err.Error())
		}
		blog.V(3).Infof("ModifyListener response: %s", resp.ToJsonString())
		break
	}
	if counter > maxRetry {
		mf(metrics.LibCallStatusErr)
		blog.Errorf("ModifyListener out of maxRetry %d", maxRetry)
		return fmt.Errorf("ModifyListener out of maxRetry %d", maxRetry)
	}
	err = sw.waitTaskDone(region, *resp.Response.RequestId)
	if err != nil {
		mf(metrics.LibCallStatusErr)
		return err
	}
	mf(metrics.LibCallStatusOK)
	return nil
}

// DeregisterTargets deregister targets
// the max number target for each operation is 20
// when receiving a big request, just split it to multiple small request
func (sw *SdkWrapper) DeregisterTargets(region string, req *tclb.DeregisterTargetsRequest) error {
	rounds := len(req.Targets) / MaxTargetForRegisterEachTime
	remains := len(req.Targets) % MaxTargetForRegisterEachTime

	index := 0
	for ; index <= rounds; index++ {
		start := index * MaxTargetForRegisterEachTime
		end := (index + 1) * MaxTargetForRegisterEachTime
		if index == rounds {
			end = start + remains
		}
		blog.V(3).Infof("DeregisterTargets (%d,%d)/%d", start, end-1, len(req.Targets))
		newReq := tclb.NewDeregisterTargetsRequest()
		newReq.LoadBalancerId = req.LoadBalancerId
		newReq.ListenerId = req.ListenerId
		newReq.Domain = req.Domain
		newReq.Url = req.Url
		newReq.Targets = req.Targets[start:end]
		if err := sw.doDeregisterTargets(region, newReq); err != nil {
			return err
		}
	}
	return nil
}

// doDeregisterTargets wrap DeregisterTargets
func (sw *SdkWrapper) doDeregisterTargets(region string, req *tclb.DeregisterTargetsRequest) error {
	blog.V(3).Infof("DeregisterTargets request: %s", req.ToJsonString())
	var err error
	var resp *tclb.DeregisterTargetsResponse

	startTime := time.Now()
	mf := func(ret string) {
		metrics.ReportLibRequestMetric(
			SystemNameInMetricTencentCloud,
			HandlerNameInMetricTencentCloudSDK,
			"DeregisterTargets", ret, startTime)
	}

	counter := 1
	for ; counter <= maxRetry; counter++ {
		blog.V(3).Infof("DeregisterTargets try %d/%d", counter, maxRetry)
		sw.tryThrottle()
		// get client by region
		clbCli, inErr := sw.getRegionClient(region)
		if inErr != nil {
			mf(metrics.LibCallStatusErr)
			return inErr
		}
		resp, err = clbCli.DeregisterTargets(req)
		if err != nil {
			if terr, ok := err.(*terrors.TencentCloudSDKError); ok {
				sw.checkErrCode(terr)
				if terr.Code == RequestLimitExceededCode || terr.Code == WrongStatusCode {
					continue
				}
			}
			mf(metrics.LibCallStatusErr)
			blog.Errorf("DeregisterTargets failed, err %s", err.Error())
			return fmt.Errorf("DeregisterTargets failed, err %s", err.Error())
		}
		blog.V(3).Infof("DeregisterTargets response: %s", resp.ToJsonString())
		break
	}
	if counter > maxRetry {
		mf(metrics.LibCallStatusTimeout)
		blog.Errorf("DeregisterTargets out of maxRetry %d", maxRetry)
		return fmt.Errorf("DeregisterTargets out of maxRetry %d", maxRetry)
	}
	err = sw.waitTaskDone(region, *resp.Response.RequestId)
	if err != nil {
		mf(metrics.LibCallStatusErr)
		return err
	}
	mf(metrics.LibCallStatusOK)
	return nil
}

// RegisterTargets register targets
// the max number target for each operation is 20
// when receiving a big request, just split it to multiple small request
func (sw *SdkWrapper) RegisterTargets(region string, req *tclb.RegisterTargetsRequest) error {
	rounds := len(req.Targets) / MaxTargetForRegisterEachTime
	remains := len(req.Targets) % MaxTargetForRegisterEachTime

	index := 0
	for ; index <= rounds; index++ {
		start := index * MaxTargetForRegisterEachTime
		end := (index + 1) * MaxTargetForRegisterEachTime
		if index == rounds {
			end = start + remains
		}
		blog.V(3).Infof("RegisterTargets (%d,%d)/%d", start, end-1, len(req.Targets))
		newReq := tclb.NewRegisterTargetsRequest()
		newReq.LoadBalancerId = req.LoadBalancerId
		newReq.ListenerId = req.ListenerId
		newReq.Domain = req.Domain
		newReq.Url = req.Url
		newReq.Targets = req.Targets[start:end]
		if err := sw.doRegisterTargets(region, newReq); err != nil {
			return err
		}
	}
	return nil
}

// doRegisterTargets wrap RegisterTargets
func (sw *SdkWrapper) doRegisterTargets(region string, req *tclb.RegisterTargetsRequest) error {
	blog.V(3).Infof("RegisterTargets request: %s", req.ToJsonString())
	var err error
	var resp *tclb.RegisterTargetsResponse

	startTime := time.Now()
	mf := func(ret string) {
		metrics.ReportLibRequestMetric(
			SystemNameInMetricTencentCloud,
			HandlerNameInMetricTencentCloudSDK,
			"RegisterTargets", ret, startTime)
	}

	counter := 1
	for ; counter <= maxRetry; counter++ {
		blog.V(3).Infof("RegisterTargets try %d/%d", counter, maxRetry)
		sw.tryThrottle()
		clbCli, inErr := sw.getRegionClient(region)
		if inErr != nil {
			mf(metrics.LibCallStatusErr)
			return inErr
		}
		resp, err = clbCli.RegisterTargets(req)
		if err != nil {
			if terr, ok := err.(*terrors.TencentCloudSDKError); ok {
				sw.checkErrCode(terr)
				if terr.Code == RequestLimitExceededCode || terr.Code == WrongStatusCode {
					continue
				}
			}
			mf(metrics.LibCallStatusErr)
			blog.Errorf("RegisterTargets failed, err %s", err.Error())
			return fmt.Errorf("RegisterTargets failed, err %s", err.Error())
		}
		blog.V(3).Infof("RegisterTargets response: %s", resp.ToJsonString())
		break
	}
	if counter > maxRetry {
		mf(metrics.LibCallStatusTimeout)
		blog.Errorf("RegisterTargets out of maxRetry %d", maxRetry)
		return fmt.Errorf("RegisterTargets out of maxRetry %d", maxRetry)
	}
	err = sw.waitTaskDone(region, *resp.Response.RequestId)
	if err != nil {
		mf(metrics.LibCallStatusErr)
		return err
	}
	mf(metrics.LibCallStatusOK)
	return nil
}

// ModifyTargetWeight wrap ModifyTargetWeight
func (sw *SdkWrapper) ModifyTargetWeight(region string, req *tclb.ModifyTargetWeightRequest) error {
	blog.V(3).Infof("ModifyTargetWeight request: %s", req.ToJsonString())
	var err error
	var resp *tclb.ModifyTargetWeightResponse

	startTime := time.Now()
	mf := func(ret string) {
		metrics.ReportLibRequestMetric(
			SystemNameInMetricTencentCloud,
			HandlerNameInMetricTencentCloudSDK,
			"ModifyTargetWeight", ret, startTime)
	}

	counter := 1
	for ; counter <= maxRetry; counter++ {
		blog.V(3).Infof("ModifyTargetWeight try %d/%d", counter, maxRetry)
		sw.tryThrottle()
		clbCli, inErr := sw.getRegionClient(region)
		if inErr != nil {
			mf(metrics.LibCallStatusErr)
			return inErr
		}
		resp, err = clbCli.ModifyTargetWeight(req)
		if err != nil {
			if terr, ok := err.(*terrors.TencentCloudSDKError); ok {
				sw.checkErrCode(terr)
				if terr.Code == RequestLimitExceededCode || terr.Code == WrongStatusCode {
					continue
				}
			}
			mf(metrics.LibCallStatusErr)
			blog.Errorf("ModifyTargetWeight failed, err %s", err.Error())
			return fmt.Errorf("ModifyTargetWeight failed, err %s", err.Error())
		}
		blog.V(3).Infof("ModifyTargetWeight response: %s", resp.ToJsonString())
		break
	}
	if counter > maxRetry {
		mf(metrics.LibCallStatusErr)
		blog.Errorf("ModifyTargetWeight out of maxRetry %d", maxRetry)
		return fmt.Errorf("ModifyTargetWeight out of maxRetry %d", maxRetry)
	}
	err = sw.waitTaskDone(region, *resp.Response.RequestId)
	if err != nil {
		mf(metrics.LibCallStatusErr)
		return err
	}
	mf(metrics.LibCallStatusOK)
	return nil
}

// BatchRegisterTargets batch register clb targets
func (sw *SdkWrapper) BatchRegisterTargets(region string, req *tclb.BatchRegisterTargetsRequest) []string {
	rounds := len(req.Targets) / MaxTargetForBatchRegisterEachTime
	remains := len(req.Targets) % MaxTargetForBatchRegisterEachTime

	failedListenerIDMap := make(map[string]struct{})
	index := 0
	for ; index <= rounds; index++ {
		start := index * MaxTargetForBatchRegisterEachTime
		end := (index + 1) * MaxTargetForBatchRegisterEachTime
		if index == rounds {
			end = start + remains
		}
		blog.V(3).Infof("BatchRegisterTargets (%d,%d)/%d", start, end-1, len(req.Targets))
		newReq := tclb.NewBatchRegisterTargetsRequest()
		newReq.LoadBalancerId = req.LoadBalancerId
		newReq.Targets = req.Targets[start:end]
		tmpFailedIDs, err := sw.doBatchRegisterTargets(region, newReq)
		if err != nil {
			blog.Warnf("do batch register targets failed, err %s", err.Error())
			for _, tg := range newReq.Targets {
				failedListenerIDMap[*tg.ListenerId] = struct{}{}
			}
			continue
		}
		for _, id := range tmpFailedIDs {
			failedListenerIDMap[id] = struct{}{}
		}
	}
	var retList []string
	for id := range failedListenerIDMap {
		retList = append(retList, id)
	}
	return retList
}

// doBatchRegisterTargets batch register clb targets
func (sw *SdkWrapper) doBatchRegisterTargets(region string, req *tclb.BatchRegisterTargetsRequest) ([]string, error) {
	blog.V(3).Infof("BatchRegisterTargets request: %s", req.ToJsonString())
	var err error
	var resp *tclb.BatchRegisterTargetsResponse
	startTime := time.Now()
	mf := func(ret string) {
		metrics.ReportLibRequestMetric(
			SystemNameInMetricTencentCloud,
			HandlerNameInMetricTencentCloudSDK,
			"BatchRegisterTargets", ret, startTime)
	}

	counter := 1
	for ; counter <= maxRetry; counter++ {
		blog.V(3).Infof("BatchRegisterTargets try %d/%d", counter, maxRetry)
		sw.tryThrottle()
		clbCli, inErr := sw.getRegionClient(region)
		if inErr != nil {
			mf(metrics.LibCallStatusErr)
			return nil, inErr
		}
		resp, err = clbCli.BatchRegisterTargets(req)
		if err != nil {
			if terr, ok := err.(*terrors.TencentCloudSDKError); ok {
				sw.checkErrCode(terr)
				if terr.Code == RequestLimitExceededCode || terr.Code == WrongStatusCode {
					continue
				}
			}
			mf(metrics.LibCallStatusErr)
			blog.Errorf("BatchRegisterTargets failed, err %s", err.Error())
			return nil, fmt.Errorf("BatchRegisterTargets failed, err %s", err.Error())
		}
		blog.V(3).Infof("BatchRegisterTargets response: %s", resp.ToJsonString())
		break
	}
	if counter > maxRetry {
		mf(metrics.LibCallStatusTimeout)
		blog.Errorf("BatchRegisterTargets out of maxRetry %d", maxRetry)
		return nil, fmt.Errorf("BatchRegisterTargets out of maxRetry %d", maxRetry)
	}
	err = sw.waitTaskDone(region, *resp.Response.RequestId)
	if err != nil {
		mf(metrics.LibCallStatusErr)
		return nil, err
	}
	mf(metrics.LibCallStatusOK)

	var failedListenerIDs []string
	if len(resp.Response.FailListenerIdSet) != 0 {
		failedListenerIDs = tcommon.StringValues(resp.Response.FailListenerIdSet)
	}
	return failedListenerIDs, nil
}

// BatchDeregisterTargets batch deregister clb targets
func (sw *SdkWrapper) BatchDeregisterTargets(region string, req *tclb.BatchDeregisterTargetsRequest) []string {
	rounds := len(req.Targets) / MaxTargetForBatchRegisterEachTime
	remains := len(req.Targets) % MaxTargetForBatchRegisterEachTime

	var failedListenerIDs []string
	index := 0
	for ; index <= rounds; index++ {
		start := index * MaxTargetForBatchRegisterEachTime
		end := (index + 1) * MaxTargetForBatchRegisterEachTime
		if index == rounds {
			end = start + remains
		}
		blog.V(3).Infof("BatchDeregisterTargets (%d,%d)/%d", start, end-1, len(req.Targets))
		newReq := tclb.NewBatchDeregisterTargetsRequest()
		newReq.LoadBalancerId = req.LoadBalancerId
		newReq.Targets = req.Targets[start:end]
		tmpFailedIDs, err := sw.doBatchDeregisterTargets(region, newReq)
		if err != nil {
			blog.Warnf("do batch de register targets failed, err %s", err.Error())
			for _, tg := range newReq.Targets {
				failedListenerIDs = append(failedListenerIDs, *tg.ListenerId)
			}
			continue
		}
		if len(tmpFailedIDs) != 0 {
			failedListenerIDs = append(failedListenerIDs, tmpFailedIDs...)
		}
	}
	return failedListenerIDs
}

// batch deregister clb targets
func (sw *SdkWrapper) doBatchDeregisterTargets(region string, req *tclb.BatchDeregisterTargetsRequest) (
	[]string, error) {
	blog.V(3).Infof("BatchDeregisterTargets request: %s", req.ToJsonString())
	var err error
	var resp *tclb.BatchDeregisterTargetsResponse
	startTime := time.Now()
	mf := func(ret string) {
		metrics.ReportLibRequestMetric(
			SystemNameInMetricTencentCloud,
			HandlerNameInMetricTencentCloudSDK,
			"BatchDeregisterTargets", ret, startTime)
	}

	counter := 1
	for ; counter <= maxRetry; counter++ {
		blog.V(3).Infof("BatchDeregisterTargets try %d/%d", counter, maxRetry)
		sw.tryThrottle()
		clbCli, inErr := sw.getRegionClient(region)
		if inErr != nil {
			mf(metrics.LibCallStatusErr)
			return nil, inErr
		}
		resp, err = clbCli.BatchDeregisterTargets(req)
		if err != nil {
			if terr, ok := err.(*terrors.TencentCloudSDKError); ok {
				sw.checkErrCode(terr)
				if terr.Code == RequestLimitExceededCode || terr.Code == WrongStatusCode {
					continue
				}
			}
			mf(metrics.LibCallStatusErr)
			blog.Errorf("BatchDeregisterTargets failed, err %s", err.Error())
			return nil, fmt.Errorf("BatchDeregisterTargets failed, err %s", err.Error())
		}
		blog.V(3).Infof("BatchDeregisterTargets response: %s", resp.ToJsonString())
		break
	}
	if counter > maxRetry {
		mf(metrics.LibCallStatusTimeout)
		blog.Errorf("BatchDeregisterTargets out of maxRetry %d", maxRetry)
		return nil, fmt.Errorf("BatchDeregisterTargets out of maxRetry %d", maxRetry)
	}
	err = sw.waitTaskDone(region, *resp.Response.RequestId)
	if err != nil {
		mf(metrics.LibCallStatusErr)
		return nil, err
	}
	mf(metrics.LibCallStatusOK)

	var failedListenerIDs []string
	if len(resp.Response.FailListenerIdSet) != 0 {
		failedListenerIDs = tcommon.StringValues(resp.Response.FailListenerIdSet)
	}
	return failedListenerIDs, nil
}

// BatchModifyTargetWeight batch modify target weight
func (sw *SdkWrapper) BatchModifyTargetWeight(region string, req *tclb.BatchModifyTargetWeightRequest) error {
	rounds := len(req.ModifyList) / MaxTargetForBatchRegisterEachTime
	remains := len(req.ModifyList) % MaxTargetForBatchRegisterEachTime

	index := 0
	for ; index <= rounds; index++ {
		start := index * MaxTargetForBatchRegisterEachTime
		end := (index + 1) * MaxTargetForBatchRegisterEachTime
		if index == rounds {
			end = start + remains
		}
		blog.V(3).Infof("BatchModifyTargetWeight (%d,%d)/%d", start, end-1, len(req.ModifyList))
		newReq := tclb.NewBatchModifyTargetWeightRequest()
		newReq.LoadBalancerId = req.LoadBalancerId
		newReq.ModifyList = req.ModifyList[start:end]
		if err := sw.doBatchModifyTargetWeight(region, newReq); err != nil {
			return err
		}
	}
	return nil
}

// batch modify target weight
func (sw *SdkWrapper) doBatchModifyTargetWeight(region string, req *tclb.BatchModifyTargetWeightRequest) error {
	blog.V(3).Infof("BatchModifyTargetWeight request: %s", req.ToJsonString())
	var err error
	var resp *tclb.BatchModifyTargetWeightResponse
	startTime := time.Now()
	mf := func(ret string) {
		metrics.ReportLibRequestMetric(
			SystemNameInMetricTencentCloud,
			HandlerNameInMetricTencentCloudSDK,
			"BatchModifyTargetWeight", ret, startTime)
	}

	counter := 1
	for ; counter <= maxRetry; counter++ {
		blog.V(3).Infof("BatchModifyTargetWeight try %d/%d", counter, maxRetry)
		sw.tryThrottle()
		clbCli, inErr := sw.getRegionClient(region)
		if inErr != nil {
			mf(metrics.LibCallStatusErr)
			return inErr
		}
		resp, err = clbCli.BatchModifyTargetWeight(req)
		if err != nil {
			if terr, ok := err.(*terrors.TencentCloudSDKError); ok {
				sw.checkErrCode(terr)
				if terr.Code == RequestLimitExceededCode || terr.Code == WrongStatusCode {
					continue
				}
			}
			mf(metrics.LibCallStatusErr)
			blog.Errorf("BatchModifyTargetWeight failed, err %s", err.Error())
			return fmt.Errorf("BatchModifyTargetWeight failed, err %s", err.Error())
		}
		blog.V(3).Infof("BatchModifyTargetWeight response: %s", resp.ToJsonString())
		break
	}
	if counter > maxRetry {
		mf(metrics.LibCallStatusTimeout)
		blog.Errorf("BatchModifyTargetWeight out of maxRetry %d", maxRetry)
		return fmt.Errorf("BatchModifyTargetWeight out of maxRetry %d", maxRetry)
	}
	err = sw.waitTaskDone(region, *resp.Response.RequestId)
	if err != nil {
		mf(metrics.LibCallStatusErr)
		return err
	}
	mf(metrics.LibCallStatusOK)
	return nil
}
