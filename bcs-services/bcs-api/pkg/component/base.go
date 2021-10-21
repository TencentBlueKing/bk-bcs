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

package component

import (
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/parnurzeal/gorequest"
)

// Response common response for bk
type Response struct {
	Code    int                    `json:"code"`
	Data    map[string]interface{} `json:"data"`
	Message string                 `json:"message"`
}

// QueryResp response from bk iam
type QueryResp struct {
	RequestID  string    `json:"request_id"`
	Result     bool      `json:"result"`
	ErrCode    int       `json:"bk_error_code"`
	ErrMessage string    `json:"bk_error_msg"`
	Data       QueryData `json:"data"`
}

// QueryData query data
type QueryData struct {
	Identity map[string]interface{} `json:"identity"`
}

// HTTPGet for bk iam get request
func HTTPGet(url string, params map[string]string) (Response, error) {
	var result Response

	req := gorequest.New().Get(url)

	for k, v := range params {
		req = req.Param(k, v)
	}

	resp, _, errs := req.Timeout(5 * time.Second).EndStruct(&result)

	if len(errs) != 0 {
		blog.Debug(fmt.Sprintf("http get fail, [url=%s, errs=%v]", url, errs))
		err := fmt.Errorf("http get fail, [url=%s, errs=%v]", url, errs)
		return result, err
	}
	if resp.StatusCode != 200 {
		blog.Debug(fmt.Sprintf("http get fail, [url=%s, status=%d]", url, resp.StatusCode))
		err := fmt.Errorf("http get fail, [url=%s, status=%d]", url, resp.StatusCode)
		return result, err
	}

	if result.Code != 0 {
		err := fmt.Errorf("http get success, [url=%s, status=%d, code=%d], return code !=0. %s",
			url, resp.StatusCode, result.Code, result.Message)
		blog.Debug(err)
		return result, err
	}

	if result.Data == nil {
		err := fmt.Errorf("http get success, [url=%s, status=%d, code=%d], return data is nil. %s",
			url, resp.StatusCode, result.Code, result.Message)
		blog.Debug(err.Error())
		return result, err
	}

	return result, nil
}

//HTTPPostToBkIamAuth bk iam post
func HTTPPostToBkIamAuth(url string, data map[string]interface{}, header map[string]string) (QueryResp, error) {
	var result QueryResp

	req := gorequest.New().Post(url)

	for k, v := range header {
		req = req.Set(k, v)
	}

	resp, _, errs := req.SendStruct(data).Timeout(5 * time.Second).EndStruct(&result)

	if len(errs) != 0 {
		blog.Debug(fmt.Sprintf("http post fail, [url=%s, errs=%v]", url, errs))
		err := fmt.Errorf("http post fail, [url=%s, errs=%v]", url, errs)
		return result, err
	}
	if resp.StatusCode != 200 {
		blog.Debug(fmt.Sprintf("http post fail, [url=%s, status=%d]", url, resp.StatusCode))
		err := fmt.Errorf("http post fail, [url=%s, status=%d]", url, resp.StatusCode)
		return result, err
	}

	blog.Infof("%d", result.ErrCode)
	// NOTE: check the code here or not?
	if result.ErrCode != 0 {
		err := fmt.Errorf("http post success, [url=%s, status=%d, errorcode=%d], return code !=0. %s",
			url, resp.StatusCode, result.ErrCode, result.ErrMessage)
		blog.Debug(err.Error())
		return result, err
	}

	if result.Data.Identity == nil {
		err := fmt.Errorf("http post success, [url=%s, status=%d, code=%d], return data is nil. %s",
			url, resp.StatusCode, result.ErrCode, result.ErrMessage)
		blog.Debug(err.Error())
		return result, err
	}

	return result, nil
}
