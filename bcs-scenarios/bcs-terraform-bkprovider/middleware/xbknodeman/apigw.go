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

// Package xbknodeman xxx
package xbknodeman

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-bkprovider/common"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-bkprovider/middleware/xrequests"
)

var (
	httpMethodMep = map[string]xrequests.SendFunc{
		http.MethodGet:    xrequests.Get,
		http.MethodPost:   xrequests.Post,
		http.MethodDelete: xrequests.Delete,
		http.MethodPut:    xrequests.Put,
	}
)

// BaseRequest contains the common information of all requests
type BaseRequest struct {
	BkAppCode   string `json:"bk_app_code,omitempty"`
	BkAppSecret string `json:"bk_app_secret,omitempty"`
	AccessToken string `json:"access_token,omitempty"`
	BkUsername  string `json:"bk_username,omitempty"`
}

// BaseResponse contains the common information of all responses
type BaseResponse struct {
	Result     bool   `json:"result"`
	Code       any    `json:"code"`
	Message    string `json:"message"`
	Permission any    `json:"permission"`
	RequestId  string `json:"request_id"`
}

// ApiResponse contains the common information and data
type ApiResponse struct {
	*BaseResponse
	Data any `json:"data"`
}

// NewBaseRequest create a request object
func NewBaseRequest(bkAppCode, bkAppSecret, accessToken, bkUsername string) *BaseRequest {
	return &BaseRequest{
		BkAppCode:   bkAppCode,
		BkAppSecret: bkAppSecret,
		AccessToken: accessToken,
		BkUsername:  bkUsername,
	}
}

// GetApigwApiUrl apigw api
func GetApigwApiUrl(host string, path string) string {
	return fmt.Sprintf("%s%s", host, path)
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
		blog.Errorf("api failed, trace: %s", common.JsonMarshal(trace))
		return baseResponse, fmt.Errorf("api failed, trace: %s", common.JsonMarshal(trace))
	}
	if responseData != nil {
		err = common.JsonConvert(rawResponseData, responseData)
		if err != nil {
			return baseResponse, fmt.Errorf("api '%s' response json error %s", url, err.Error())
		}
	}
	return baseResponse, nil
}
