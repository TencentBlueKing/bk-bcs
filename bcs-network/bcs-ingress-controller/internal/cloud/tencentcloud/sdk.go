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
	RequestLimitExceededCode = "4400"
	// WrongStatusCode code for incorrect status
	WrongStatusCode = "4000"
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
)

var (
	// If the delay caused by the frequency limit exceeds this value, it is recorded in the log
	maxLatency = 120 * time.Millisecond
	// the maximum number of retries caused by server error or API overrun
	maxRetry = 25
	// qps for rate limit
	throttleQPS = 300
	// bucket size for rate limit
	bucketSize = 300
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

	sw.throttler = throttle.NewTokenBucket(int64(throttleQPS), int64(bucketSize))
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

// CreateListener wrap CreateListener
func (sw *SdkWrapper) CreateListener(region string, req *tclb.CreateListenerRequest) (string, error) {
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
			return "", inErr
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
			return "", fmt.Errorf("CreateListener failed, err %s", err.Error())
		}
		if len(resp.Response.ListenerIds) == 0 {
			mf(metrics.LibCallStatusErr)
			blog.Errorf("create listener return zero length ids")
			return "", fmt.Errorf("create listener return zero length ids")
		}
		blog.V(3).Infof("CreateListener response: %s", resp.ToJsonString())
		break
	}
	if counter > maxRetry {
		mf(metrics.LibCallStatusTimeout)
		blog.Errorf("CreateListener out of maxRetry %d", maxRetry)
		return "", fmt.Errorf("CreateListener out of maxRetry %d", maxRetry)
	}
	err = sw.waitTaskDone(region, *resp.Response.RequestId)
	if err != nil {
		mf(metrics.LibCallStatusErr)
		return "", err
	}
	mf(metrics.LibCallStatusOK)
	return *resp.Response.ListenerIds[0], nil
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

// CreateRule wrap CreateRule
func (sw *SdkWrapper) CreateRule(region string, req *tclb.CreateRuleRequest) error {
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
			return inErr
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
			return fmt.Errorf("CreateRule failed, err %s", err.Error())
		}
		blog.V(3).Infof("CreateRule response: %s", resp.ToJsonString())
		break
	}
	if counter > maxRetry {
		mf(metrics.LibCallStatusErr)
		blog.Errorf("CreateRule out of maxRetry %d", maxRetry)
		return fmt.Errorf("CreateRule out of maxRetry %d", maxRetry)
	}
	err = sw.waitTaskDone(region, *resp.Response.RequestId)
	if err != nil {
		mf(metrics.LibCallStatusErr)
		return err
	}
	mf(metrics.LibCallStatusOK)
	return nil
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
