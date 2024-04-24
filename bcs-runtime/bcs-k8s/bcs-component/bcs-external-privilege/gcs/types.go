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

package gcs

import "encoding/json"

type PrivAPIResponse struct {
	Result  bool          `json:"result"`
	Code    int           `json:"code"`
	Data    PrivilegeData `json:"data"`
	Message string        `json:"message"`
}

type CDPAPIResponse struct {
	Message string `json:"msg"`
	JobId   string `json:"jobid"`
	Url     string `json:"url"`
	Code    int    `json:"code"`
}

type CheckAPIResponse struct {
	Result  bool            `json:"result"`
	Code    int             `json:"code"`
	Data    CheckStatusData `json:"data"`
	Message string          `json:"message"`
}

type PrivilegeData struct {
	Url   string `json:"url"`
	Pwd   string `json:"pwd"`
	JobId string `json:"job_id"`
	Usr   string `json:"usr"`
}

type CheckStatusData struct {
	Message string `json:"msg"`
	Url     string `json:"url"`
	Code    int    `json:"code"`
	JobId   string `json:"jobid"`
}

func bodyToPrivAPIResponse(body []byte) (PrivAPIResponse, error) {
	resp := PrivAPIResponse{}
	cdpresp := CDPAPIResponse{}
	err := json.Unmarshal(body, &cdpresp)
	if err != nil {
		return resp, err
	}
	if len(cdpresp.Message) == 0 && len(cdpresp.JobId) == 0 && len(cdpresp.Url) == 0 {
		err := json.Unmarshal(body, &resp)
		if err != nil {
			return resp, err
		}
		return resp, nil
	}
	resp.Code = cdpresp.Code
	resp.Message = cdpresp.Message
	resp.Result = cdpresp.Code == 0
	resp.Data = PrivilegeData{
		Url:   cdpresp.Url,
		JobId: cdpresp.JobId,
	}
	return resp, nil
}

func bodyToCheckAPIResponse(body []byte) (CheckAPIResponse, error) {
	resp := CheckAPIResponse{}
	cdpresp := CDPAPIResponse{}
	err := json.Unmarshal(body, &cdpresp)
	if err != nil {
		return resp, err
	}
	if len(cdpresp.Message) == 0 && len(cdpresp.JobId) == 0 && len(cdpresp.Url) == 0 {
		err := json.Unmarshal(body, &resp)
		if err != nil {
			return resp, err
		}
		return resp, nil
	}
	resp.Code = cdpresp.Code
	resp.Message = cdpresp.Message
	resp.Result = true
	resp.Data = CheckStatusData{
		Message: cdpresp.Message,
		Url:     cdpresp.Url,
		JobId:   cdpresp.JobId,
	}
	return resp, nil
}
