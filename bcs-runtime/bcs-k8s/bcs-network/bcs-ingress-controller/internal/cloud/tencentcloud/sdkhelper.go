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
	"runtime"
	"strconv"
	"time"

	tclb "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/clb/v20180317"
	tcommon "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	terrors "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/metrics"
)

// load config for environment variables
func (sw *SdkWrapper) loadEnv() error {
	// 请求腾讯云API使用的域名
	clbDomain := os.Getenv(EnvNameTencentCloudClbDomain)
	// 腾讯云密钥ID
	secretID := os.Getenv(EnvNameTencentCloudAccessKeyID)
	// 腾讯云密钥Key
	secretKey := os.Getenv(EnvNameTencentCloudAccessKey)
	sw.domain = clbDomain
	sw.secretID = secretID
	sw.secretKey = secretKey

	// 本地限制的请求API QPS
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
	iCli, ok := sw.clbCliMap.Load(region)
	if !ok {
		newCli, err := tclb.NewClient(sw.credential, region, sw.cpf)
		if err != nil {
			blog.Errorf("create clb client for region %s failed, err %s", region, err.Error())
			return nil, fmt.Errorf("create clb client for region %s failed, err %s", region, err.Error())
		}
		sw.clbCliMap.Store(region, newCli)
		return newCli, nil
	}
	cli, ok := iCli.(*tclb.Client)
	if !ok {
		blog.Errorf("unknown type store in clbCliMap, value: %v", iCli)
		return nil, fmt.Errorf("unknown type store in clbCliMap, value: %v", iCli)
	}

	return cli, nil
}

// checkErrCode common method for check tencent cloud sdk err
func (sw *SdkWrapper) checkErrCode(err *terrors.TencentCloudSDKError, metricFunc func(ret string)) {
	if err.Code == RequestLimitExceededCode { // API请求速率超过QPS
		blog.Warnf("request exceed limit, have a rest for %d second, err: %s", waitPeriodLBDealing, err.Error())
		metricFunc(metrics.LibCallStatusExceedLimit)
		time.Sleep(time.Duration(waitPeriodLBDealing) * time.Second)
	} else if err.Code == WrongStatusCode { // 通常是由于有多个请求同时操作LB（如同时创建/删除监听器）
		blog.Warnf("clb is dealing another action, have a rest for %d second, err: %s", waitPeriodLBDealing, err.Error())
		time.Sleep(time.Duration(waitPeriodLBDealing) * time.Second)
	}
}

// call tryThrottle before each api call
func (sw *SdkWrapper) tryThrottle() {
	now := time.Now()
	sw.throttler.Accept()
	if latency := time.Since(now); latency > maxLatency {
		pc, _, _, _ := runtime.Caller(2)
		callerName := runtime.FuncForPC(pc).Name() // 通过调用栈获取方法名称
		blog.Infof("Throttling request took %d ms, function: %s", latency, callerName)
	}
}

// waitTaskDone wait asynchronous task done
func (sw *SdkWrapper) waitTaskDone(region, taskID string) error {
	blog.V(3).Infof("start waiting for task %s", taskID)
	request := tclb.NewDescribeTaskStatusRequest()
	request.TaskId = tcommon.StringPtr(taskID)
	blog.Infof("describe task status request:\n%s", request.ToJsonString())

	startTime := time.Now()
	// 统计API调用延时/状态
	mf := func(ret string) {
		metrics.ReportLibRequestMetric(
			SystemNameInMetricTencentCloud,
			HandlerNameInMetricTencentCloudSDK,
			"waitTaskDone", ret, startTime)
	}

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
				sw.checkErrCode(terr, mf)
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

// doDeregisterTargets wrap DeregisterTargets
func (sw *SdkWrapper) doDeregisterTargets(region string, req *tclb.DeregisterTargetsRequest) error {
	blog.V(3).Infof("DeregisterTargets request: %s", req.ToJsonString())
	var err error
	var resp *tclb.DeregisterTargetsResponse

	startTime := time.Now()
	// 统计API调用延时/状态
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
				sw.checkErrCode(terr, mf)
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
	// 等待异步请求执行完毕
	err = sw.waitTaskDone(region, *resp.Response.RequestId)
	if err != nil {
		mf(metrics.LibCallStatusErr)
		return err
	}
	mf(metrics.LibCallStatusOK)
	return nil
}

// doRegisterTargets wrap RegisterTargets
func (sw *SdkWrapper) doRegisterTargets(region string, req *tclb.RegisterTargetsRequest) error {
	blog.V(3).Infof("RegisterTargets request: %s", req.ToJsonString())
	var err error
	var resp *tclb.RegisterTargetsResponse

	startTime := time.Now()
	// 统计API调用延时/状态
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
				sw.checkErrCode(terr, mf)
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
	// 等待异步请求执行完毕
	err = sw.waitTaskDone(region, *resp.Response.RequestId)
	if err != nil {
		mf(metrics.LibCallStatusErr)
		return err
	}
	mf(metrics.LibCallStatusOK)
	return nil
}

// doBatchRegisterTargets batch register clb targets
func (sw *SdkWrapper) doBatchRegisterTargets(region string, req *tclb.BatchRegisterTargetsRequest) ([]string, error) {
	blog.V(3).Infof("BatchRegisterTargets request: %s", req.ToJsonString())
	var err error
	var resp *tclb.BatchRegisterTargetsResponse
	startTime := time.Now()
	// 统计API调用延时/状态
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
				sw.checkErrCode(terr, mf)
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
	// 等待异步请求执行完毕
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

// batch deregister clb targets
func (sw *SdkWrapper) doBatchDeregisterTargets(region string, req *tclb.BatchDeregisterTargetsRequest) (
	[]string, error) {
	blog.V(3).Infof("BatchDeregisterTargets request: %s", req.ToJsonString())
	var err error
	var resp *tclb.BatchDeregisterTargetsResponse
	startTime := time.Now()
	// 统计API调用延时/状态
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
				sw.checkErrCode(terr, mf)
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
	// 等待异步请求执行完毕
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

// batch modify target weight
func (sw *SdkWrapper) doBatchModifyTargetWeight(region string, req *tclb.BatchModifyTargetWeightRequest) error {
	blog.V(3).Infof("BatchModifyTargetWeight request: %s", req.ToJsonString())
	var err error
	var resp *tclb.BatchModifyTargetWeightResponse
	startTime := time.Now()
	// 统计API调用延时/状态
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
				sw.checkErrCode(terr, mf)
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
	// 等待异步请求执行完毕
	err = sw.waitTaskDone(region, *resp.Response.RequestId)
	if err != nil {
		mf(metrics.LibCallStatusErr)
		return err
	}
	mf(metrics.LibCallStatusOK)
	return nil
}

func (sw *SdkWrapper) doDescribeTargetHealth(region string,
	req *tclb.DescribeTargetHealthRequest) (*tclb.DescribeTargetHealthResponse, error) {
	blog.V(5).Infof("DescribeTargetHealth request: %s", req.ToJsonString())
	var err error
	var resp *tclb.DescribeTargetHealthResponse

	startTime := time.Now()
	// 统计API调用延时/状态
	mf := func(ret string) {
		metrics.ReportLibRequestMetric(
			SystemNameInMetricTencentCloud,
			HandlerNameInMetricTencentCloudSDK,
			"DescribeTargetHealth", ret, startTime)
	}

	counter := 1
	for ; counter <= maxRetry; counter++ {
		blog.V(5).Infof("DescribeTargetHealth try %d/%d", counter, maxRetry)
		sw.tryThrottle()
		// get client by region
		clbCli, inErr := sw.getRegionClient(region)
		if inErr != nil {
			mf(metrics.LibCallStatusErr)
			return nil, inErr
		}
		resp, err = clbCli.DescribeTargetHealth(req)
		if err != nil {
			if terr, ok := err.(*terrors.TencentCloudSDKError); ok {
				sw.checkErrCode(terr, mf)
				if terr.Code == RequestLimitExceededCode || terr.Code == WrongStatusCode {
					continue
				}
			}
			mf(metrics.LibCallStatusErr)
			blog.Errorf("DescribeTargetHealth failed, err %s", err.Error())
			return nil, fmt.Errorf("DescribeTargetHealth failed, err %s", err.Error())
		}
		blog.V(5).Infof("DescribeTargetHealth response: %s", resp.ToJsonString())
		break
	}
	if counter > maxRetry {
		mf(metrics.LibCallStatusErr)
		blog.Errorf("DescribeTargetHealth out of maxRetry %d", maxRetry)
		return nil, fmt.Errorf("DescribeTargetHealth out of maxRetry %d", maxRetry)
	}
	mf(metrics.LibCallStatusOK)
	return resp, nil
}
