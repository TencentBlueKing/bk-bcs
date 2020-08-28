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
	maxLatency  = 120 * time.Millisecond
	maxRetry    = 25
	throttleQPS = 300
	bucketSize  = 300

	waitPeriodLBDealing = 2
)

// SdkWrapper wrapper for tencentcloud sdk
type SdkWrapper struct {
	domain string

	secretID string

	secretKey string

	cpf        *tprofile.ClientProfile
	credential *tcommon.Credential

	throttler throttle.RateLimiter
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
	var resp *tclb.DescribeLoadBalancersResponse
	counter := 1
	for ; counter <= maxRetry; counter++ {
		blog.V(3).Infof("DescribeLoadBalancers try %d/%d", counter, maxRetry)
		sw.tryThrottle()
		clbCli, err := sw.getRegionClient(region)
		if err != nil {
			return nil, err
		}
		resp, err = clbCli.DescribeLoadBalancers(req)
		if err != nil {
			if terr, ok := err.(*terrors.TencentCloudSDKError); ok {
				sw.checkErrCode(terr)
				if terr.Code == RequestLimitExceededCode || terr.Code == WrongStatusCode {
					continue
				}
			}
			blog.Errorf("DescribeLoadBalancers failed, err %s", err.Error())
			return nil, fmt.Errorf("DescribeLoadBalancers failed, err %s", err.Error())
		}
		blog.V(3).Infof("DescribeLoadBalancers response: %s", resp.ToJsonString())
		break
	}
	if counter > maxRetry {
		blog.Errorf("DescribeLoadBalancers out of maxRetry %d", maxRetry)
		return nil, fmt.Errorf("DescribeLoadBalancers out of maxRetry %d", maxRetry)
	}
	return resp, nil
}

// CreateListener wrap CreateListener
func (sw *SdkWrapper) CreateListener(region string, req *tclb.CreateListenerRequest) (string, error) {
	blog.V(3).Infof("CreateListener request: %s", req.ToJsonString())
	var err error
	var resp *tclb.CreateListenerResponse
	counter := 1
	for ; counter <= maxRetry; counter++ {
		blog.V(3).Infof("CreateListener try %d/%d", counter, maxRetry)
		sw.tryThrottle()
		clbCli, err := sw.getRegionClient(region)
		if err != nil {
			return "", err
		}
		resp, err = clbCli.CreateListener(req)
		if err != nil {
			if terr, ok := err.(*terrors.TencentCloudSDKError); ok {
				sw.checkErrCode(terr)
				if terr.Code == RequestLimitExceededCode || terr.Code == WrongStatusCode {
					continue
				}
			}
			blog.Errorf("CreateListener failed, err %s", err.Error())
			return "", fmt.Errorf("CreateListener failed, err %s", err.Error())
		}
		if len(resp.Response.ListenerIds) == 0 {
			blog.Errorf("create listener return zero length ids")
			return "", fmt.Errorf("create listener return zero length ids")
		}
		blog.V(3).Infof("CreateListener response: %s", resp.ToJsonString())
		break
	}
	if counter > maxRetry {
		blog.Errorf("CreateListener out of maxRetry %d", maxRetry)
		return "", fmt.Errorf("CreateListener out of maxRetry %d", maxRetry)
	}
	err = sw.waitTaskDone(region, *resp.Response.RequestId)
	if err != nil {
		return "", err
	}
	return *resp.Response.ListenerIds[0], nil
}

// DescribeListeners wrap DescribeListeners
func (sw *SdkWrapper) DescribeListeners(region string, req *tclb.DescribeListenersRequest) (
	*tclb.DescribeListenersResponse, error) {

	blog.V(3).Infof("DescribeListeners request: %s", req.ToJsonString())
	var resp *tclb.DescribeListenersResponse
	counter := 1
	for ; counter <= maxRetry; counter++ {
		blog.V(3).Infof("DescribeListeners try %d/%d", counter, maxRetry)
		sw.tryThrottle()
		clbCli, err := sw.getRegionClient(region)
		if err != nil {
			return nil, err
		}
		resp, err = clbCli.DescribeListeners(req)
		if err != nil {
			if terr, ok := err.(*terrors.TencentCloudSDKError); ok {
				sw.checkErrCode(terr)
				if terr.Code == RequestLimitExceededCode || terr.Code == WrongStatusCode {
					continue
				}
			}
			blog.Errorf("DescribeListeners failed, err %s", err.Error())
			return nil, fmt.Errorf("DescribeListeners failed, err %s", err.Error())
		}
		blog.V(3).Infof("DescribeListeners response: %s", resp.ToJsonString())
		break
	}
	if counter > maxRetry {
		blog.Errorf("DescribeListeners out of maxRetry %d", maxRetry)
		return nil, fmt.Errorf("DescribeListeners out of maxRetry %d", maxRetry)
	}
	return resp, nil
}

// DescribeTargets wrap DescribeTargets
func (sw *SdkWrapper) DescribeTargets(region string, req *tclb.DescribeTargetsRequest) (
	*tclb.DescribeTargetsResponse, error) {

	blog.V(3).Infof("DescribeTargets request: %s", req.ToJsonString())
	var resp *tclb.DescribeTargetsResponse
	counter := 1
	for ; counter <= maxRetry; counter++ {
		blog.V(3).Infof("DescribeTargets try %d/%d", counter, maxRetry)
		sw.tryThrottle()
		clbCli, err := sw.getRegionClient(region)
		if err != nil {
			return nil, err
		}
		resp, err = clbCli.DescribeTargets(req)
		if err != nil {
			if terr, ok := err.(*terrors.TencentCloudSDKError); ok {
				sw.checkErrCode(terr)
				if terr.Code == RequestLimitExceededCode || terr.Code == WrongStatusCode {
					continue
				}
			}
			blog.Errorf("DescribeTargets failed, err %s", err.Error())
			return nil, fmt.Errorf("DescribeTargets failed, err %s", err.Error())
		}
		blog.V(3).Infof("DescribeTargets response: %s", resp.ToJsonString())
		break
	}
	if counter > maxRetry {
		blog.Errorf("DescribeTargets out of maxRetry %d", maxRetry)
		return nil, fmt.Errorf("DescribeTargets out of maxRetry %d", maxRetry)
	}
	return resp, nil
}

// DeleteListener wrap DeleteListener
func (sw *SdkWrapper) DeleteListener(region string, req *tclb.DeleteListenerRequest) error {
	blog.V(3).Infof("DeleteListener request: %s", req.ToJsonString())
	var err error
	var resp *tclb.DeleteListenerResponse
	counter := 1
	for ; counter <= maxRetry; counter++ {
		blog.V(3).Infof("DeleteListener try %d/%d", counter, maxRetry)
		sw.tryThrottle()
		clbCli, err := sw.getRegionClient(region)
		if err != nil {
			return err
		}
		resp, err = clbCli.DeleteListener(req)
		if err != nil {
			if terr, ok := err.(*terrors.TencentCloudSDKError); ok {
				sw.checkErrCode(terr)
				if terr.Code == RequestLimitExceededCode || terr.Code == WrongStatusCode {
					continue
				}
			}
			blog.Errorf("DeleteListener failed, err %s", err.Error())
			return fmt.Errorf("DeleteListener failed, err %s", err.Error())
		}
		blog.V(3).Infof("DeleteListener response: %s", resp.ToJsonString())
		break
	}
	if counter > maxRetry {
		blog.Errorf("DeleteListener out of maxRetry %d", maxRetry)
		return fmt.Errorf("DeleteListener out of maxRetry %d", maxRetry)
	}
	err = sw.waitTaskDone(region, *resp.Response.RequestId)
	if err != nil {
		return err
	}
	return nil
}

// CreateRule wrap CreateRule
func (sw *SdkWrapper) CreateRule(region string, req *tclb.CreateRuleRequest) error {
	blog.V(3).Infof("CreateRule request: %s", req.ToJsonString())
	var err error
	var resp *tclb.CreateRuleResponse
	counter := 1
	for ; counter <= maxRetry; counter++ {
		blog.V(3).Infof("CreateRule try %d/%d", counter, maxRetry)
		sw.tryThrottle()
		clbCli, err := sw.getRegionClient(region)
		if err != nil {
			return err
		}
		resp, err = clbCli.CreateRule(req)
		if err != nil {
			if terr, ok := err.(*terrors.TencentCloudSDKError); ok {
				sw.checkErrCode(terr)
				if terr.Code == RequestLimitExceededCode || terr.Code == WrongStatusCode {
					continue
				}
			}
			blog.Errorf("CreateRule failed, err %s", err.Error())
			return fmt.Errorf("CreateRule failed, err %s", err.Error())
		}
		blog.V(3).Infof("CreateRule response: %s", resp.ToJsonString())
		break
	}
	if counter > maxRetry {
		blog.Errorf("CreateRule out of maxRetry %d", maxRetry)
		return fmt.Errorf("CreateRule out of maxRetry %d", maxRetry)
	}
	err = sw.waitTaskDone(region, *resp.Response.RequestId)
	if err != nil {
		return err
	}
	return nil
}

// DeleteRule wrap DeleteRule
func (sw *SdkWrapper) DeleteRule(region string, req *tclb.DeleteRuleRequest) error {
	blog.V(3).Infof("DeleteRule request: %s", req.ToJsonString())
	var err error
	var resp *tclb.DeleteRuleResponse
	counter := 1
	for ; counter <= maxRetry; counter++ {
		blog.V(3).Infof("DeleteRule try %d/%d", counter, maxRetry)
		sw.tryThrottle()
		clbCli, err := sw.getRegionClient(region)
		if err != nil {
			return err
		}
		resp, err = clbCli.DeleteRule(req)
		if err != nil {
			if terr, ok := err.(*terrors.TencentCloudSDKError); ok {
				sw.checkErrCode(terr)
				if terr.Code == RequestLimitExceededCode || terr.Code == WrongStatusCode {
					continue
				}
			}
			blog.Errorf("DeleteRule failed, err %s", err.Error())
			return fmt.Errorf("DeleteRule failed, err %s", err.Error())
		}
		blog.V(3).Infof("DeleteRule response: %s", resp.ToJsonString())
		break
	}
	if counter > maxRetry {
		blog.Errorf("DeleteRule out of maxRetry %d", maxRetry)
		return fmt.Errorf("DeleteRule out of maxRetry %d", maxRetry)
	}
	err = sw.waitTaskDone(region, *resp.Response.RequestId)
	if err != nil {
		return err
	}
	return nil
}

// ModifyRule wrap ModifyRule
func (sw *SdkWrapper) ModifyRule(region string, req *tclb.ModifyRuleRequest) error {
	blog.V(3).Infof("ModifyRule request: %s", req.ToJsonString())
	var err error
	var resp *tclb.ModifyRuleResponse
	counter := 1
	for ; counter <= maxRetry; counter++ {
		blog.V(3).Infof("ModifyRule try %d/%d", counter, maxRetry)
		sw.tryThrottle()
		clbCli, err := sw.getRegionClient(region)
		if err != nil {
			return err
		}
		resp, err = clbCli.ModifyRule(req)
		if err != nil {
			if terr, ok := err.(*terrors.TencentCloudSDKError); ok {
				sw.checkErrCode(terr)
				if terr.Code == RequestLimitExceededCode || terr.Code == WrongStatusCode {
					continue
				}
			}
			blog.Errorf("ModifyRule failed, err %s", err.Error())
			return fmt.Errorf("ModifyRule failed, err %s", err.Error())
		}
		blog.V(3).Infof("ModifyRule response: %s", resp.ToJsonString())
		break
	}
	if counter > maxRetry {
		blog.Errorf("ModifyRule out of maxRetry %d", maxRetry)
		return fmt.Errorf("ModifyRule out of maxRetry %d", maxRetry)
	}
	err = sw.waitTaskDone(region, *resp.Response.RequestId)
	if err != nil {
		return err
	}
	return nil
}

// ModifyListener wrap ModifyListener
func (sw *SdkWrapper) ModifyListener(region string, req *tclb.ModifyListenerRequest) error {
	blog.V(3).Infof("ModifyListener request: %s", req.ToJsonString())
	var err error
	var resp *tclb.ModifyListenerResponse
	counter := 1
	for ; counter <= maxRetry; counter++ {
		blog.V(3).Infof("ModifyListener try %d/%d", counter, maxRetry)
		sw.tryThrottle()
		clbCli, err := sw.getRegionClient(region)
		if err != nil {
			return err
		}
		resp, err = clbCli.ModifyListener(req)
		if err != nil {
			if terr, ok := err.(*terrors.TencentCloudSDKError); ok {
				sw.checkErrCode(terr)
				if terr.Code == RequestLimitExceededCode || terr.Code == WrongStatusCode {
					continue
				}
			}
			blog.Errorf("ModifyListener failed, err %s", err.Error())
			return fmt.Errorf("ModifyListener failed, err %s", err.Error())
		}
		blog.V(3).Infof("ModifyListener response: %s", resp.ToJsonString())
		break
	}
	if counter > maxRetry {
		blog.Errorf("ModifyListener out of maxRetry %d", maxRetry)
		return fmt.Errorf("ModifyListener out of maxRetry %d", maxRetry)
	}
	err = sw.waitTaskDone(region, *resp.Response.RequestId)
	if err != nil {
		return err
	}
	return nil
}

// DeregisterTargets wrap DeregisterTargets
func (sw *SdkWrapper) DeregisterTargets(region string, req *tclb.DeregisterTargetsRequest) error {
	blog.V(3).Infof("DeregisterTargets request: %s", req.ToJsonString())
	var err error
	var resp *tclb.DeregisterTargetsResponse
	counter := 1
	for ; counter <= maxRetry; counter++ {
		blog.V(3).Infof("DeregisterTargets try %d/%d", counter, maxRetry)
		sw.tryThrottle()
		clbCli, err := sw.getRegionClient(region)
		if err != nil {
			return err
		}
		resp, err = clbCli.DeregisterTargets(req)
		if err != nil {
			if terr, ok := err.(*terrors.TencentCloudSDKError); ok {
				sw.checkErrCode(terr)
				if terr.Code == RequestLimitExceededCode || terr.Code == WrongStatusCode {
					continue
				}
			}
			blog.Errorf("DeregisterTargets failed, err %s", err.Error())
			return fmt.Errorf("DeregisterTargets failed, err %s", err.Error())
		}
		blog.V(3).Infof("DeregisterTargets response: %s", resp.ToJsonString())
		break
	}
	if counter > maxRetry {
		blog.Errorf("DeregisterTargets out of maxRetry %d", maxRetry)
		return fmt.Errorf("DeregisterTargets out of maxRetry %d", maxRetry)
	}
	err = sw.waitTaskDone(region, *resp.Response.RequestId)
	if err != nil {
		return err
	}
	return nil
}

// RegisterTargets wrap RegisterTargets
func (sw *SdkWrapper) RegisterTargets(region string, req *tclb.RegisterTargetsRequest) error {
	blog.V(3).Infof("RegisterTargets request: %s", req.ToJsonString())
	var err error
	var resp *tclb.RegisterTargetsResponse
	counter := 1
	for ; counter <= maxRetry; counter++ {
		blog.V(3).Infof("RegisterTargets try %d/%d", counter, maxRetry)
		sw.tryThrottle()
		clbCli, err := sw.getRegionClient(region)
		if err != nil {
			return err
		}
		resp, err = clbCli.RegisterTargets(req)
		if err != nil {
			if terr, ok := err.(*terrors.TencentCloudSDKError); ok {
				sw.checkErrCode(terr)
				if terr.Code == RequestLimitExceededCode || terr.Code == WrongStatusCode {
					continue
				}
			}
			blog.Errorf("RegisterTargets failed, err %s", err.Error())
			return fmt.Errorf("RegisterTargets failed, err %s", err.Error())
		}
		blog.V(3).Infof("RegisterTargets response: %s", resp.ToJsonString())
		break
	}
	if counter > maxRetry {
		blog.Errorf("RegisterTargets out of maxRetry %d", maxRetry)
		return fmt.Errorf("RegisterTargets out of maxRetry %d", maxRetry)
	}
	err = sw.waitTaskDone(region, *resp.Response.RequestId)
	if err != nil {
		return err
	}
	return nil
}

// ModifyTargetWeight wrap ModifyTargetWeight
func (sw *SdkWrapper) ModifyTargetWeight(region string, req *tclb.ModifyTargetWeightRequest) error {
	blog.V(3).Infof("ModifyTargetWeight request: %s", req.ToJsonString())
	var err error
	var resp *tclb.ModifyTargetWeightResponse
	counter := 1
	for ; counter <= maxRetry; counter++ {
		blog.V(3).Infof("ModifyTargetWeight try %d/%d", counter, maxRetry)
		sw.tryThrottle()
		clbCli, err := sw.getRegionClient(region)
		if err != nil {
			return err
		}
		resp, err = clbCli.ModifyTargetWeight(req)
		if err != nil {
			if terr, ok := err.(*terrors.TencentCloudSDKError); ok {
				sw.checkErrCode(terr)
				if terr.Code == RequestLimitExceededCode || terr.Code == WrongStatusCode {
					continue
				}
			}
			blog.Errorf("ModifyTargetWeight failed, err %s", err.Error())
			return fmt.Errorf("ModifyTargetWeight failed, err %s", err.Error())
		}
		blog.V(3).Infof("ModifyTargetWeight response: %s", resp.ToJsonString())
		break
	}
	if counter > maxRetry {
		blog.Errorf("ModifyTargetWeight out of maxRetry %d", maxRetry)
		return fmt.Errorf("ModifyTargetWeight out of maxRetry %d", maxRetry)
	}
	err = sw.waitTaskDone(region, *resp.Response.RequestId)
	if err != nil {
		return err
	}
	return nil
}
