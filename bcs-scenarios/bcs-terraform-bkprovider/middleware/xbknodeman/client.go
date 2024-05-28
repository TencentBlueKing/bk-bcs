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

package xbknodeman

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-bkprovider/common"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-bkprovider/middleware/xrequests"
)

// Client is a client for bk API
type Client struct {
	*BaseRequest
	Env  string `json:"env"`
	Host string `json:"host"`
}

// NewClient  return new client
func NewClient(env, bkAppCode, bkAppSecret, accessToken, username string) *Client {
	return &Client{
		BaseRequest: NewBaseRequest(bkAppCode, bkAppSecret, accessToken, username),
		Env:         env,
		Host:        os.Getenv(EnvBkNodeManHost),
	}
}

// SendRequest send request and get baseResponse or error
func (t *Client) SendRequest(ctx context.Context, method string, uri string, params any, responseData any) (*BaseResponse, error) {
	url := GetApigwApiUrl(t.Host, uri)
	authMap := map[string]string{
		"bk_app_code":   t.BkAppCode,
		"bk_app_secret": t.BkAppSecret,
	}
	if t.BkUsername != "" {
		authMap["bk_username"] = t.BkUsername
	}
	if t.AccessToken != "" {
		authMap["access_token"] = t.AccessToken
	}
	return SendRequest(ctx, method, url, params, responseData, &xrequests.RequestOptions{
		Headers: map[string]string{
			"X-Bkapi-Authorization": common.JsonMarshal(authMap),
		},
	})
}

// GetProxyHost get proxy host
func (t *Client) GetProxyHost(ctx context.Context, request *GetProxyHostRequest) (*GetProxyHostResponse, error) {
	if request == nil {
		request = &GetProxyHostRequest{}
	}

	param := struct {
		*BaseRequest
		*GetProxyHostRequest
	}{
		BaseRequest:         t.BaseRequest,
		GetProxyHostRequest: request,
	}
	data := make([]*ProxyHost, 0)
	uri := "/host/proxies/"
	baseResponse, err := t.SendRequest(ctx, http.MethodGet, uri, param, &data)
	if err != nil {
		return nil, err
	}
	response := &GetProxyHostResponse{
		BaseResponse: baseResponse,
		Data:         data,
	}
	return response, nil
}

// InstallJob install job
func (t *Client) InstallJob(ctx context.Context, request *InstallJobRequest) (*InstallJobResponse, error) {
	if request == nil {
		request = &InstallJobRequest{}
	}

	param := struct {
		*BaseRequest
		*InstallJobRequest
	}{
		BaseRequest:       t.BaseRequest,
		InstallJobRequest: request,
	}
	data := &Job{}
	uri := "/job/install/"
	baseResponse, err := t.SendRequest(ctx, http.MethodPost, uri, param, &data)
	if err != nil {
		return nil, err
	}
	response := &InstallJobResponse{
		BaseResponse: baseResponse,
		Data:         data,
	}
	return response, nil
}

// ListCloud list cloud
func (t *Client) ListCloud(ctx context.Context, request *ListCloudRequest) (*ListCloudResponse, error) {
	if request == nil {
		request = &ListCloudRequest{}
	}

	param := struct {
		*BaseRequest
		*ListCloudRequest
	}{
		BaseRequest:      t.BaseRequest,
		ListCloudRequest: request,
	}
	data := make([]*Cloud, 0)
	uri := "/cloud/"
	baseResponse, err := t.SendRequest(ctx, http.MethodGet, uri, param, &data)
	if err != nil {
		return nil, err
	}
	response := &ListCloudResponse{
		BaseResponse: baseResponse,
		Data:         data,
	}
	return response, nil
}

// CreateCloud create cloud
func (t *Client) CreateCloud(ctx context.Context, request *CreateCloudRequest) (*CreateCloudResponse, error) {
	if request == nil {
		request = &CreateCloudRequest{}
	}

	param := struct {
		*BaseRequest
		*CreateCloudRequest
	}{
		BaseRequest:        t.BaseRequest,
		CreateCloudRequest: request,
	}

	data := CloudID{}
	// var uri string
	// switch t.Env {
	// case EnvSg:
	// 	uri = "/cloud/create_cloud_area/"
	// default:
	// 	uri = "/cloud/"
	// }
	uri := "/cloud/"
	baseResponse, err := t.SendRequest(ctx, http.MethodPost, uri, param, &data)
	if err != nil {
		return nil, err
	}
	response := &CreateCloudResponse{
		BaseResponse: baseResponse,
		Data:         data,
	}
	return response, nil
}

// DeleteCloud delete cloud
func (t *Client) DeleteCloud(ctx context.Context, request *DeleteCloudRequest) (*BaseResponse, error) {
	if request == nil {
		request = &DeleteCloudRequest{}
	}

	param := struct {
		*BaseRequest
		// *DeleteCloudRequest
	}{
		BaseRequest: t.BaseRequest,
		// DeleteCloudRequest: request,
	}
	uri := fmt.Sprintf("/cloud/%d/", request.BkCloudID)
	response, err := t.SendRequest(ctx, http.MethodDelete, uri, param, nil)
	if err != nil {
		return nil, err
	}
	return response, nil
}

// UpdateCloud update cloud
func (t *Client) UpdateCloud(ctx context.Context, request *UpdateCloudRequest) (*BaseResponse, error) {
	if request == nil {
		request = &UpdateCloudRequest{}
	}

	param := struct {
		*BaseRequest
		*UpdateCloudRequest
	}{
		BaseRequest:        t.BaseRequest,
		UpdateCloudRequest: request,
	}
	uri := fmt.Sprintf("/cloud/%d/", request.BkCloudID)
	response, err := t.SendRequest(ctx, http.MethodPut, uri, param, nil)
	if err != nil {
		return nil, err
	}
	return response, nil
}

// GetBizProxyHost get biz proxy host
func (t *Client) GetBizProxyHost(ctx context.Context, request *GetBizProxyHostRequest) (*GetBizProxyHostResponse, error) {
	if request == nil {
		request = &GetBizProxyHostRequest{}
	}

	param := struct {
		*BaseRequest
		*GetBizProxyHostRequest
	}{
		BaseRequest:            t.BaseRequest,
		GetBizProxyHostRequest: request,
	}
	data := make([]*BizProxyHost, 0)
	uri := "/host/biz_proxies/"
	baseResponse, err := t.SendRequest(ctx, http.MethodGet, uri, param, &data)
	if err != nil {
		return nil, err
	}
	response := &GetBizProxyHostResponse{
		BaseResponse: baseResponse,
		Data:         data,
	}
	return response, nil
}

// ListHosts list hosts
func (t *Client) ListHosts(ctx context.Context, request *ListHostRequest) (*ListHostResponse, error) {
	if request == nil {
		request = &ListHostRequest{}
	}

	param := struct {
		*BaseRequest
		*ListHostRequest
	}{
		BaseRequest:     t.BaseRequest,
		ListHostRequest: request,
	}
	data := &ListHostData{}
	uri := "/host/search/"
	baseResponse, err := t.SendRequest(ctx, http.MethodPost, uri, param, &data)
	if err != nil {
		return nil, err
	}
	response := &ListHostResponse{
		BaseResponse: baseResponse,
		Data:         data,
	}
	return response, nil
}

// GetJobDetails get job details
func (t *Client) GetJobDetails(ctx context.Context, request *GetJobDetailRequest) (*GetJobDetailResponse, error) {
	if request == nil {
		request = &GetJobDetailRequest{}
	}

	param := struct {
		*BaseRequest
		*GetJobDetailRequest
	}{
		BaseRequest:         t.BaseRequest,
		GetJobDetailRequest: request,
	}
	data := &GetJobDetailData{}
	uri := "/job/details/"
	baseResponse, err := t.SendRequest(ctx, http.MethodPost, uri, param, &data)
	if err != nil {
		return nil, err
	}
	response := &GetJobDetailResponse{
		BaseResponse: baseResponse,
		Data:         data,
	}
	return response, nil
}
