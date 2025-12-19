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
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/throttle"
	"github.com/pkg/errors"
	tclb "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/clb/v20180317"
	tcommon "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	terrors "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	tprofile "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/constant"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/metrics"
)

const (
	// RequestLimitExceededCode code for request exceeded limit
	RequestLimitExceededCode = "RequestLimitExceeded"
	// WrongStatusCode code for incorrect status
	WrongStatusCode = "InternalError"
	// ResourceInOperation code for resource in operation
	ResourceInOperation = "FailedOperation.ResourceInOperating"
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
	clbCliMap sync.Map
	// 加锁避免同时处理多个同Region的CLB时，触发concurrent write map错误
	mu sync.Mutex

	listenerNameValidateMode string // if in strict Mode, create ListenerName with clusterID
	bcsClusterID             string
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

	sw.throttler = throttle.NewTokenBucket(sw.ratelimitqps, sw.ratelimitbucketSize)
	sw.bcsClusterID = os.Getenv(constant.EnvNameBkBCSClusterID)
	sw.listenerNameValidateMode = os.Getenv(constant.EnvNameListenerNameValidateMode)
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

	sw.throttler = throttle.NewTokenBucket(sw.ratelimitqps, sw.ratelimitbucketSize)
	sw.bcsClusterID = os.Getenv(constant.EnvNameBkBCSClusterID)
	sw.listenerNameValidateMode = os.Getenv(constant.EnvNameListenerNameValidateMode)
	return sw, nil
}

// NewSdkWrapperWithParams create sdk wrapper with secret id and secret key and domain
func NewSdkWrapperWithParams(id, key, domain string) (*SdkWrapper, error) {
	sw := &SdkWrapper{}
	err := sw.loadEnv()
	if err != nil {
		return nil, err
	}

	sw.secretID = id
	sw.secretKey = key
	sw.domain = domain

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

	sw.throttler = throttle.NewTokenBucket(sw.ratelimitqps, sw.ratelimitbucketSize)
	sw.bcsClusterID = os.Getenv(constant.EnvNameBkBCSClusterID)
	sw.listenerNameValidateMode = os.Getenv(constant.EnvNameListenerNameValidateMode)
	return sw, nil
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
				if sw.isRetryableErr(terr, mf) {
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
	if sw.listenerNameValidateMode == constant.ListenerNameValidateModeStrict {
		liNames := make([]*string, 0, len(req.ListenerNames))
		for _, name := range req.ListenerNames {
			liNames = append(liNames, tcommon.StringPtr(sw.bcsClusterID+"-"+*name))
		}

		req.ListenerNames = liNames
	}
	rounds := len(req.ListenerNames) / MaxListenersForCreateEachTime
	remains := len(req.ListenerNames) % MaxListenersForCreateEachTime

	index := 0
	var respList []string
	for ; index <= rounds; index++ {
		start := index * MaxListenersForCreateEachTime
		end := (index + 1) * MaxListenersForCreateEachTime
		if index == rounds {
			end = start + remains
		}
		if start == end {
			break
		}
		blog.V(3).Infof("CreateListener (%d,%d)/%d", start, end-1, len(req.ListenerNames))
		newReq := tclb.NewCreateListenerRequest()
		newReq.LoadBalancerId = req.LoadBalancerId
		newReq.ListenerNames = req.ListenerNames[start:end]
		newReq.Ports = req.Ports[start:end]
		newReq.Protocol = req.Protocol
		newReq.SessionExpireTime = req.SessionExpireTime
		newReq.Scheduler = req.Scheduler
		newReq.HealthCheck = req.HealthCheck
		newReq.Certificate = req.Certificate
		newReq.SniSwitch = req.SniSwitch
		newReq.KeepaliveEnable = req.KeepaliveEnable
		newReq.SessionType = req.SessionType
		if req.EndPort != nil {
			newReq.EndPort = req.EndPort
		}
		listenerIDs, err := sw.doCreateListener(region, newReq)
		if err != nil {
			return nil, err
		}
		respList = append(respList, listenerIDs...)
	}
	return respList, nil
}

// wrap CreateListener, length of Ports should be less than 50
func (sw *SdkWrapper) doCreateListener(region string, req *tclb.CreateListenerRequest) ([]string, error) {
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
				if sw.isRetryableErr(terr, mf) {
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

// DescribeListeners wraps DescribeListeners
func (sw *SdkWrapper) DescribeListeners(region string, req *tclb.DescribeListenersRequest) (
	*tclb.DescribeListenersResponse, error) {
	// when length of listenerIDs is zero, describe all listeners of a loadbalancer
	if len(req.ListenerIds) == 0 {
		return sw.doDescribeListeners(region, req)
	}

	rounds := len(req.ListenerIds) / MaxListenersForDescribeEachTime
	remains := len(req.ListenerIds) % MaxListenersForDescribeEachTime
	index := 0
	var resp *tclb.DescribeListenersResponse
	for ; index <= rounds; index++ {
		start := index * MaxListenersForDescribeEachTime
		end := (index + 1) * MaxListenersForDescribeEachTime
		if index == rounds {
			end = start + remains
		}
		if start == end {
			break
		}
		blog.V(3).Infof("DescribeListeners (%d,%d)/%d", start, end-1, len(req.ListenerIds))
		newReq := tclb.NewDescribeListenersRequest()
		newReq.LoadBalancerId = req.LoadBalancerId
		newReq.ListenerIds = req.ListenerIds[start:end]
		newReq.Port = req.Port
		newReq.Protocol = req.Protocol
		tmpResp, err := sw.doDescribeListeners(region, newReq)
		if err != nil {
			return nil, err
		}
		if index == 0 {
			resp = &tclb.DescribeListenersResponse{
				BaseResponse: tmpResp.BaseResponse,
				Response:     tmpResp.Response,
			}
		} else {
			resp.Response.Listeners = append(resp.Response.Listeners, tmpResp.Response.Listeners...)
		}
	}
	return resp, nil
}

// wrap DescribeListeners
func (sw *SdkWrapper) doDescribeListeners(region string, req *tclb.DescribeListenersRequest) (
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
				if sw.isRetryableErr(terr, mf) {
					continue
				}
			}
			mf(metrics.LibCallStatusErr)
			blog.Errorf("DescribeListeners failed, err %s", err.Error())
			return nil, fmt.Errorf("DescribeListeners failed, err %s", err.Error())
		}
		blog.V(3).Infof("DescribeListeners response: %d listeners", len(resp.Response.Listeners))
		blog.V(4).Infof("DescribeListeners response: %s", resp.ToJsonString())
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
	// when length of listenerIDs is zero, describe targets by port
	if len(req.ListenerIds) == 0 {
		return sw.doDescribeTargets(region, req)
	}

	rounds := len(req.ListenerIds) / MaxListenerForDescribeTargetEachTime
	remains := len(req.ListenerIds) % MaxListenerForDescribeTargetEachTime
	index := 0
	var resp *tclb.DescribeTargetsResponse
	for ; index <= rounds; index++ {
		start := index * MaxListenerForDescribeTargetEachTime
		end := (index + 1) * MaxListenerForDescribeTargetEachTime
		if index == rounds {
			end = start + remains
		}
		if start == end {
			break
		}
		blog.V(3).Infof("DescribeTargets (%d,%d)/%d", start, end-1, len(req.ListenerIds))
		newReq := tclb.NewDescribeTargetsRequest()
		newReq.LoadBalancerId = req.LoadBalancerId
		newReq.Protocol = req.Protocol
		newReq.Port = req.Port
		newReq.ListenerIds = req.ListenerIds[start:end]
		tmpResp, err := sw.doDescribeTargets(region, newReq)
		if err != nil {
			return nil, err
		}
		if index == 0 {
			resp = &tclb.DescribeTargetsResponse{
				BaseResponse: tmpResp.BaseResponse,
				Response:     tmpResp.Response,
			}
		} else {
			resp.Response.Listeners = append(resp.Response.Listeners, tmpResp.Response.Listeners...)
		}
	}
	return resp, nil
}

// doDescribeTargets wrap DescribeTargets
func (sw *SdkWrapper) doDescribeTargets(region string, req *tclb.DescribeTargetsRequest) (
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
				if sw.isRetryableErr(terr, mf) {
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
				if sw.isRetryableErr(terr, mf) {
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
	// It's possible delete all listeners when listenerIds empty
	if len(req.ListenerIds) == 0 {
		return fmt.Errorf("req.ListenerIds should not be empty when Delete Loadbalance Listenners")
	}
	// deletions is limited to MaxListenersForDeleteEachTime for each time
	rounds := len(req.ListenerIds) / MaxListenersForDeleteEachTime
	remains := len(req.ListenerIds) % MaxListenersForDeleteEachTime
	index := 0
	for ; index <= rounds; index++ {
		start := index * MaxListenersForDeleteEachTime
		end := (index + 1) * MaxListenersForDeleteEachTime
		if index == rounds {
			end = start + remains
		}
		if start == end {
			break
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
				if sw.isRetryableErr(terr, mf) {
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
				if sw.isRetryableErr(terr, mf) {
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
				if sw.isRetryableErr(terr, mf) {
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
				if sw.isRetryableErr(terr, mf) {
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
				if sw.isRetryableErr(terr, mf) {
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

// ModifyDomainAttributes wrap ModifyDomainAttributes
func (sw *SdkWrapper) ModifyDomainAttributes(region string, req *tclb.ModifyDomainAttributesRequest) error {
	blog.V(3).Infof("ModifyDomainAttributes request: %s", req.ToJsonString())
	var err error
	var resp *tclb.ModifyDomainAttributesResponse

	startTime := time.Now()
	mf := func(ret string) {
		metrics.ReportLibRequestMetric(
			SystemNameInMetricTencentCloud,
			HandlerNameInMetricTencentCloudSDK,
			"ModifyDomainAttributes", ret, startTime)
	}

	counter := 1
	for ; counter <= maxRetry; counter++ {
		blog.V(3).Infof("ModifyDomainAttributes try %d/%d", counter, maxRetry)
		sw.tryThrottle()
		// get client by region
		clbCli, inErr := sw.getRegionClient(region)
		if inErr != nil {
			mf(metrics.LibCallStatusErr)
			return inErr
		}
		resp, err = clbCli.ModifyDomainAttributes(req)
		if err != nil {
			if terr, ok := err.(*terrors.TencentCloudSDKError); ok {
				if sw.isRetryableErr(terr, mf) {
					continue
				}
			}
			mf(metrics.LibCallStatusErr)
			blog.Errorf("ModifyDomainAttributes failed, err %s", err.Error())
			return fmt.Errorf("ModifyDomainAttributes failed, err %s", err.Error())
		}
		blog.V(3).Infof("ModifyDomainAttributes response: %s", resp.ToJsonString())
		break
	}
	if counter > maxRetry {
		mf(metrics.LibCallStatusErr)
		blog.Errorf("ModifyDomainAttributes out of maxRetry %d", maxRetry)
		return fmt.Errorf("ModifyDomainAttributes out of maxRetry %d", maxRetry)
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
		if start == end {
			break
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
		if start == end {
			break
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
				if sw.isRetryableErr(terr, mf) {
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
func (sw *SdkWrapper) BatchRegisterTargets(region string, req *tclb.BatchRegisterTargetsRequest) map[string]error {
	rounds := len(req.Targets) / MaxTargetForBatchRegisterEachTime
	remains := len(req.Targets) % MaxTargetForBatchRegisterEachTime

	failedListenerIDMap := make(map[string]error)
	index := 0
	for ; index <= rounds; index++ {
		start := index * MaxTargetForBatchRegisterEachTime
		end := (index + 1) * MaxTargetForBatchRegisterEachTime
		if index == rounds {
			end = start + remains
		}
		if start == end {
			break
		}
		blog.V(3).Infof("BatchRegisterTargets (%d,%d)/%d", start, end-1, len(req.Targets))
		newReq := tclb.NewBatchRegisterTargetsRequest()
		newReq.LoadBalancerId = req.LoadBalancerId
		newReq.Targets = req.Targets[start:end]
		tmpFailedIDs, err := sw.doBatchRegisterTargets(region, newReq)
		if err != nil {
			err = errors.Wrapf(err, "do batch register targets failed, return err")
			blog.Warnf(err.Error())
			for _, tg := range newReq.Targets {
				failedListenerIDMap[*tg.ListenerId] = err
			}
			continue
		}
		for _, id := range tmpFailedIDs {
			failedListenerIDMap[id] = errors.New("do batch register targets failed, return failedID")
		}
	}
	return failedListenerIDMap
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
		if start == end {
			break
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
		if start == end {
			break
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

// DescribeTargetHealth describe health status of loadbalance rs
func (sw *SdkWrapper) DescribeTargetHealth(region string,
	req *tclb.DescribeTargetHealthRequest) (*tclb.DescribeTargetHealthResponse, error) {
	rounds := len(req.LoadBalancerIds) / MaxLoadBalancersForDescribeHealthStatus
	remains := len(req.LoadBalancerIds) % MaxLoadBalancersForDescribeHealthStatus

	index := 0
	var resp *tclb.DescribeTargetHealthResponse
	for ; index <= rounds; index++ {
		start := index * MaxLoadBalancersForDescribeHealthStatus
		end := (index + 1) * MaxLoadBalancersForDescribeHealthStatus
		if index == rounds {
			end = start + remains
		}
		if start == end {
			break
		}
		blog.V(3).Infof("DescribeTargetHealth (%d,%d)/%d", start, end-1, len(req.LoadBalancerIds))
		newReq := tclb.NewDescribeTargetHealthRequest()
		newReq.LoadBalancerIds = req.LoadBalancerIds[start:end]
		tmpResp, err := sw.doDescribeTargetHealth(region, newReq)
		if err != nil {
			return nil, err
		}
		if index == 0 {
			resp = &tclb.DescribeTargetHealthResponse{
				BaseResponse: tmpResp.BaseResponse,
				Response:     tmpResp.Response,
			}
		} else {
			resp.Response.LoadBalancers = append(resp.Response.LoadBalancers, tmpResp.Response.LoadBalancers...)
		}
	}
	return resp, nil
}
