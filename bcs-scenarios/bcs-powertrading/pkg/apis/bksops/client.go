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

// Package bksops xxx
package bksops

import (
	"encoding/json"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-powertrading/pkg/apis"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-powertrading/pkg/apis/requester"
)

// Client interface
type Client interface {
	CreateTask(templateId, businessId, templateName string, constants map[string]string) (*CreateTaskRsp, error)
	StartTask(taskId, businessId string) (*StartTaskRsp, error)
	GetTaskStatus(taskId, businessId string) (*GetTaskStatusRsp, error)
	GetTaskNodeDetail(taskId, businessId, nodeId string) (*GetTaskNodeDetailRsp, error)
}

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

// CreateTask create task
func (c *client) CreateTask(templateId, businessId, templateName string,
	constants map[string]string) (*CreateTaskRsp, error) {
	url := fmt.Sprintf(CreateTaskUrl, c.opt.Endpoint, templateId, businessId)
	createBody := CreateTaskReq{
		TemplateSource: "common",
		Name:           templateName,
		FlowType:       "common",
		Constants:      constants,
	}
	data, err := json.Marshal(createBody)
	if err != nil {
		blog.Errorf("Error encoding JSON: %v", err)
	}
	blog.Infof("create req:%s", data)
	rsp, requestErr := c.requestClient.DoPostRequest(url, c.defaultHeader, data)
	if requestErr != nil {
		return nil, fmt.Errorf("do CreateTask request error:%s", requestErr.Error())
	}
	blog.Infof("create rsp:%s", rsp)
	result := &CreateTaskRsp{}
	unMarshalErr := json.Unmarshal(rsp, result)
	if unMarshalErr != nil {
		return nil, fmt.Errorf("do unmarshal error, url: %s, error:%v", url, unMarshalErr.Error())
	}
	if !result.Result {
		return nil, fmt.Errorf("create task error:%s", result.Message)
	}
	return result, nil
}

// StartTask start task
func (c *client) StartTask(taskId, businessId string) (*StartTaskRsp, error) {
	url := fmt.Sprintf(StartTaskUrl, c.opt.Endpoint, taskId, businessId)
	req, _ := json.Marshal(struct{}{})
	rsp, requestErr := c.requestClient.DoPostRequest(url, c.defaultHeader, req)
	if requestErr != nil {
		return nil, fmt.Errorf("do StartTask request error:%s", requestErr.Error())
	}
	result := &StartTaskRsp{}
	unMarshalErr := json.Unmarshal(rsp, result)
	if unMarshalErr != nil {
		return nil, fmt.Errorf("do unmarshal error, url: %s, error:%v", url, unMarshalErr.Error())
	}
	if !result.Result {
		return nil, fmt.Errorf("start task %s failed:%s", taskId, result.Message)
	}
	return result, nil
}

// GetTaskStatus get task status
func (c *client) GetTaskStatus(taskId, businessId string) (*GetTaskStatusRsp, error) {
	url := fmt.Sprintf(GetTaskStatusUrl, c.opt.Endpoint, taskId, businessId)
	rsp, requestErr := c.requestClient.DoGetRequest(url, c.defaultHeader)
	if requestErr != nil {
		return nil, fmt.Errorf("do GetTaskStatus request error:%s", requestErr.Error())
	}
	result := &GetTaskStatusRsp{}
	unMarshalErr := json.Unmarshal(rsp, result)
	if unMarshalErr != nil {
		return nil, fmt.Errorf("do unmarshal error, url: %s, error:%v", url, unMarshalErr.Error())
	}
	return result, nil
}

// GetTaskNodeDetail get task node detail
func (c *client) GetTaskNodeDetail(taskId, businessId, nodeId string) (*GetTaskNodeDetailRsp, error) {
	url := fmt.Sprintf(GetTaskNodeDetailUrl, c.opt.Endpoint, taskId, businessId, nodeId)
	rsp, requestErr := c.requestClient.DoGetRequest(url, c.defaultHeader)
	if requestErr != nil {
		return nil, fmt.Errorf("do GetTaskNodeDetail request error:%s", requestErr.Error())
	}
	result := &GetTaskNodeDetailRsp{}
	unMarshalErr := json.Unmarshal(rsp, result)
	if unMarshalErr != nil {
		return nil, fmt.Errorf("do unmarshal error, url: %s, error:%v", url, unMarshalErr.Error())
	}
	return result, nil
}
