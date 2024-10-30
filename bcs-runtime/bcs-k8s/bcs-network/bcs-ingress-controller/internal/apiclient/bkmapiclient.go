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
 *
 */

package apiclient

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/apiclient/xrequests"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/constant"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/metrics"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/pkg/common"
	"os"
	"strconv"
)

// IMonitorApiClient monitor client interface
type IMonitorApiClient interface {
	ListNode(ctx context.Context) (*ListNodeResponse, error)
	ListUptimeCheckTask(ctx context.Context, request *ListUptimeCheckRequest) (*ListUptimeCheckResponse, error)
	CreateUptimeCheckTask(ctx context.Context,
		req *CreateOrUpdateUptimeCheckTaskRequest) (*CreateOrUpdateUptimeCheckTaskResponse, error)
	UpdateUptimeCheckTask(ctx context.Context,
		req *CreateOrUpdateUptimeCheckTaskRequest) (*CreateOrUpdateUptimeCheckTaskResponse, error)
	DeployUptimeCheckTask(ctx context.Context, taskID int64) error
	DeleteUptimeCheckTask(ctx context.Context, taskID int64) error
}

// BkmApiClient api client to call bk monitor
type BkmApiClient struct {
	BizID       int64
	BkAppCode   string
	BkAppSecret string
	*BaseRequest
}

// NewBkmApiClient return new bkm apli client
func NewBkmApiClient() *BkmApiClient {
	bizID, _ := strconv.ParseInt(os.Getenv(constant.EnvNameBkBizID), 10, 64)
	cli := &BkmApiClient{
		BizID:       bizID,
		BkAppSecret: os.Getenv(constant.EnvNameBkAppSecret),
		BkAppCode:   os.Getenv(constant.EnvNameBkAppCode),
	}

	cli.BaseRequest = NewBaseRequest(cli.BizID, cli.BkAppCode, cli.BkAppSecret, "")
	return cli
}

// SendRequest send request
func (b *BkmApiClient) SendRequest(ctx context.Context, method string, uri string, params any, responseData any,
	opts ...*xrequests.RequestOptions) (*BaseResponse, error) {
	startTime := time.Now()
	mf := func(ret string) {
		defer metrics.ReportLibRequestMetric(
			SystemNameInMetricBlueKingMonitor,
			HandlerNameInMetricBkmAPI,
			uri, ret, startTime)
	}
	url := GetApigwApiUrl(serviceName, urlPrefix, uri)
	baseResp, err := SendRequest(ctx, method, url, params, responseData, opts...)
	if err != nil {
		mf(metrics.LibCallStatusErr)
	} else {
		mf(metrics.LibCallStatusOK)
	}
	return baseResp, err
}

// ListUptimeCheckTask list uptime check task
func (b *BkmApiClient) ListUptimeCheckTask(ctx context.Context, req *ListUptimeCheckRequest) (*ListUptimeCheckResponse,
	error) {
	if req == nil {
		return nil, fmt.Errorf("nil request")
	}
	req.BkBizID = b.BizID

	param := struct {
		*BaseRequest
		*ListUptimeCheckRequest
	}{
		BaseRequest:            b.BaseRequest,
		ListUptimeCheckRequest: req,
	}

	data := make([]*UptimeCheckTask, 0)
	baseResp, err := b.SendRequest(ctx, http.MethodGet, "/get_uptime_check_task_list", param, &data)
	if err != nil {
		return nil, fmt.Errorf("get_uptime_check_task_list failed, req: %s, err: %v", common.ToJsonString(req), err)
	}

	return &ListUptimeCheckResponse{
		BaseResponse: baseResp,
		Data:         data,
	}, nil
}

// ListNode list node
func (b *BkmApiClient) ListNode(ctx context.Context) (*ListNodeResponse, error) {
	req := &ListUptimeCheckRequest{BkBizID: b.BizID}

	param := struct {
		*BaseRequest
		*ListUptimeCheckRequest
	}{
		BaseRequest:            b.BaseRequest,
		ListUptimeCheckRequest: req,
	}

	data := make([]*Node, 0)
	baseResp, err := b.SendRequest(ctx, http.MethodGet, "/get_uptime_check_node_list", param, &data)
	if err != nil {
		return nil, fmt.Errorf("get_uptime_check_node_list failed, req: %s, err: %v", common.ToJsonString(req), err)
	}

	return &ListNodeResponse{
		BaseResponse: baseResp,
		Data:         data,
	}, nil
}

// CreateUptimeCheckTask create uptime check task
func (b *BkmApiClient) CreateUptimeCheckTask(ctx context.Context,
	req *CreateOrUpdateUptimeCheckTaskRequest) (*CreateOrUpdateUptimeCheckTaskResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("nil request")
	}

	param := struct {
		*BaseRequest
		*CreateOrUpdateUptimeCheckTaskRequest
	}{
		BaseRequest:                          b.BaseRequest,
		CreateOrUpdateUptimeCheckTaskRequest: req,
	}

	data := &UptimeCheckTask{}
	baseResp, err := b.SendRequest(ctx, http.MethodPost, "/create_uptime_check_task", param, &data)
	if err != nil {
		return nil, fmt.Errorf("create_uptime_check_task failed, req: %s, err: %v", common.ToJsonString(req), err)
	}

	blog.V(3).Infof("create_uptime_check_task success, req: %s, resp[%d]", common.ToJsonString(req), data.ID)
	return &CreateOrUpdateUptimeCheckTaskResponse{
		BaseResponse: baseResp,
		Data:         data,
	}, nil
}

// UpdateUptimeCheckTask update uptime check task
func (b *BkmApiClient) UpdateUptimeCheckTask(ctx context.Context,
	req *CreateOrUpdateUptimeCheckTaskRequest) (*CreateOrUpdateUptimeCheckTaskResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("nil request")
	}

	param := struct {
		*BaseRequest
		*CreateOrUpdateUptimeCheckTaskRequest
	}{
		BaseRequest:                          b.BaseRequest,
		CreateOrUpdateUptimeCheckTaskRequest: req,
	}

	data := &UptimeCheckTask{}
	baseResp, err := b.SendRequest(ctx, http.MethodPost, "/update_uptime_check_task", param, &data)
	if err != nil {
		return nil, fmt.Errorf("update_uptime_check_task failed, req: %s, err: %v", common.ToJsonString(req), err)
	}

	blog.V(3).Infof("update_uptime_check_task success, req: %s, resp[%d]", common.ToJsonString(req), data.ID)
	return &CreateOrUpdateUptimeCheckTaskResponse{
		BaseResponse: baseResp,
		Data:         data,
	}, nil
}

// DeployUptimeCheckTask deploy uptime check task
func (b BkmApiClient) DeployUptimeCheckTask(ctx context.Context, taskID int64) error {
	req := &DeployUptimeCheckRequest{TaskID: taskID}

	param := struct {
		*BaseRequest
		*DeployUptimeCheckRequest
	}{
		BaseRequest:              b.BaseRequest,
		DeployUptimeCheckRequest: req,
	}

	_, err := b.SendRequest(ctx, http.MethodPost, "/deploy_uptime_check_task", param, nil,
		&xrequests.RequestOptions{
			RequestTimeout: time.Minute,
		})
	if err != nil {
		return fmt.Errorf("deploy_uptime_check_task failed, req: %s, err: %v", common.ToJsonString(req), err)
	}

	return nil
}

// DeleteUptimeCheckTask delete uptime check task
func (b *BkmApiClient) DeleteUptimeCheckTask(ctx context.Context, taskID int64) error {
	req := &DeleteUptimeCheckRequest{TaskID: taskID}

	param := struct {
		*BaseRequest
		*DeleteUptimeCheckRequest
	}{
		BaseRequest:              b.BaseRequest,
		DeleteUptimeCheckRequest: req,
	}

	_, err := b.SendRequest(ctx, http.MethodPost, "/delete_uptime_check_task", param, &struct{}{})
	if err != nil {
		return fmt.Errorf("delete_uptime_check_task failed, req: %s, err: %v", common.ToJsonString(req), err)
	}
	blog.V(3).Infof("delete_uptime_check_task success, req: %s", common.ToJsonString(req))

	return nil
}

// SendRequest send request
func SendRequest(ctx context.Context, method string, url string, params any, responseData any, opts ...*xrequests.RequestOptions,
) (*BaseResponse, error) {
	var rawResponseData any
	baseResponse := &BaseResponse{}
	apiResponse := &ApiResponse{
		BaseResponse: baseResponse,
		Data:         &rawResponseData,
	}
	f := httpMethodMep[method]
	trace, _, err := f(ctx, url, params, apiResponse, opts...)
	if err != nil {
		return baseResponse, err
	}
	if !apiResponse.Result {
		blog.Errorf("api failed, trace: %s", common.ToJsonString(trace))
		return baseResponse, fmt.Errorf("api failed, trace: %s", common.ToJsonString(trace))
	}
	if responseData != nil {
		err = common.JsonConvert(rawResponseData, responseData)
		if err != nil {
			return baseResponse, fmt.Errorf("api '%s' response json error %s", url, err.Error())
		}
	}
	return baseResponse, nil
}
