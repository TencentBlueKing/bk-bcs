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

// Package job xxx
package job

import (
	"encoding/json"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-powertrading/pkg/apis"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-powertrading/pkg/apis/requester"
)

// Client interface
type Client interface {
	GetJobStatus(scope, jobID, bizID string) (*StatusResponse, error)
	GetBatchJobLog(scope, jobID, bizID, stepID string, ipList []BatchLogIPRequest) (*BatchLogResponse, error)
}

const (
	getJobStatus = "%s/api/v3/get_job_instance_status/?bk_scope_type=%s&bk_scope_id=%s&job_instance_id=%s" +
		"&return_ip_result=true"
	getBatchLog = "%s/api/v3/batch_get_job_instance_ip_log/"
)

// New client
func New(opt *apis.ClientOptions, r requester.Requester) Client {
	cli := &client{
		requestClient: r,
		opt:           opt,
	}
	err := cli.SetDefaultHeader()
	if err != nil {
		return nil
	}
	return cli
}

type client struct {
	defaultHeader map[string]string
	requestClient requester.Requester
	opt           *apis.ClientOptions
}

// SetDefaultHeader set default header
func (c *client) SetDefaultHeader() error {
	header := make(map[string]string)
	auth := &apis.BkAuthOpts{
		BkAppCode:   c.opt.AppCode,
		BkAppSecret: c.opt.AppSecret,
		AccessToken: c.opt.AccessToken,
	}
	str, err := json.Marshal(auth)
	if err != nil {
		blog.Errorf("marshal header error:%s", err.Error())
		return err
	}
	header["X-Bkapi-Authorization"] = string(str)
	c.defaultHeader = header
	return nil
}

// GetJobStatus get job status
func (c *client) GetJobStatus(scope, jobID, bizID string) (*StatusResponse, error) {
	url := fmt.Sprintf(getJobStatus, c.opt.Endpoint, scope, bizID, jobID)
	rsp, requestErr := c.requestClient.DoGetRequest(url, c.defaultHeader)
	if requestErr != nil {
		return nil, fmt.Errorf("do GetJobStatus request error:%s", requestErr.Error())
	}
	result := &StatusResponse{}
	unMarshalErr := json.Unmarshal(rsp, result)
	if unMarshalErr != nil {
		return nil, fmt.Errorf("do unmarshal error, url: %s, error:%v", url, unMarshalErr.Error())
	}
	return result, nil
}

// GetBatchJobLog get batch job log
func (c *client) GetBatchJobLog(scope, jobID, scopeID, stepID string,
	ipList []BatchLogIPRequest) (*BatchLogResponse, error) {
	url := fmt.Sprintf(getBatchLog, c.opt.Endpoint)
	createBody := &BatchLogRequest{
		BkScopeType:    scope,
		BkScopeID:      scopeID,
		JobInstanceID:  jobID,
		StepInstanceID: stepID,
		IPList:         ipList,
	}
	data, err := json.Marshal(createBody)
	if err != nil {
		blog.Errorf("Error encoding JSON: %v", err)
	}
	rsp, requestErr := c.requestClient.DoPostRequest(url, c.defaultHeader, data)
	if requestErr != nil {
		return nil, fmt.Errorf("do GetBatchJobLog request error:%s", requestErr.Error())
	}
	result := &BatchLogResponse{}
	unMarshalErr := json.Unmarshal(rsp, result)
	if unMarshalErr != nil {
		return nil, fmt.Errorf("do unmarshal error, url: %s, error:%v", url, unMarshalErr.Error())
	}
	if result.Code != 0 {
		return nil, fmt.Errorf("GetBatchJobLog failed, code:%d, url: %s, requestid:%s, message:%s",
			result.Code, url, result.JobRequestID, result.Message)
	}
	return result, nil
}
