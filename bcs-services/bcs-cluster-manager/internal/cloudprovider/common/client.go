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

package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	"github.com/parnurzeal/gorequest"
)

var (
	defaultTimeOut = time.Second * 60
	// ErrServerNotInit server notInit
	ErrServerNotInit = errors.New("server not inited")
)

// TaskState bkops task status
type TaskState string

// String() to string
func (ts TaskState) String() string {
	return string(ts)
}

const (
	// CREATED status
	CREATED TaskState = "CREATED"
	// RUNNING status
	RUNNING TaskState = "RUNNING"
	// FAILED status
	FAILED TaskState = "FAILED"
	// SUSPENDED status
	SUSPENDED TaskState = "SUSPENDED"
	// REVOKED status
	REVOKED TaskState = "REVOKED"
	// FINISHED status
	FINISHED TaskState = "FINISHED"
)

// Options bkops options
type Options struct {
	AppCode   string
	AppSecret string
	External  bool
	Debug     bool

	TaskStatusURL string
	StartTaskURL  string
	CreateTaskURL string
}

// AuthInfo auth info
type AuthInfo struct {
	BkAppCode   string `json:"bk_app_code"`
	BkAppSecret string `json:"bk_app_secret"`
	BkUserName  string `json:"bk_username"`
}

// BKOpsClient global bkops client
var BKOpsClient *Client

// SetBKOpsClient set bkops client
func SetBKOpsClient(options Options) error {
	cli, err := NewClient(options)
	if err != nil {
		return err
	}

	BKOpsClient = cli
	return nil
}

// GetBKOpsClient get bkops client
func GetBKOpsClient() *Client {
	return BKOpsClient
}

// NewClient create bksops client
func NewClient(options Options) (*Client, error) {
	c := &Client{
		appCode:     options.AppCode,
		appSecret:   options.AppSecret,
		serverDebug: options.Debug,
		external:    options.External,
	}

	c.urls = DependURLs{
		createTaskURL: options.CreateTaskURL,
		startTaskURL:  options.StartTaskURL,
		getTaskStatus: options.TaskStatusURL,
	}

	return c, nil
}

// Client for bksops
type Client struct {
	appCode     string
	appSecret   string
	external    bool
	serverDebug bool

	urls DependURLs
}

// DependURLs depend bkops urls
type DependURLs struct {
	createTaskURL string
	startTaskURL  string
	getTaskStatus string
}

func (c *Client) generateGateWayAuth(user string) (string, error) {
	if c == nil {
		return "", ErrServerNotInit
	}

	auth := &AuthInfo{
		BkAppCode:   c.appCode,
		BkAppSecret: c.appSecret,
		BkUserName:  user,
	}

	userAuth, err := json.Marshal(auth)
	if err != nil {
		return "", err
	}

	return string(userAuth), nil
}

// CreateBkOpsTask create bkops task
func (c *Client) CreateBkOpsTask(url string, paras *CreateTaskPathParas, request *CreateTaskRequest) (*CreateTaskResponse, error) {
	if c == nil {
		return nil, ErrServerNotInit
	}
	if url == "" {
		url = c.urls.createTaskURL
	}

	var (
		reqURL   string
		respData = &CreateTaskResponse{}
	)

	if c.external {
		reqURL = url
		request.BusinessID = paras.BkBizID
		request.TemplateID = paras.TemplateID
	} else {
		reqURL = fmt.Sprintf(url, paras.TemplateID, paras.BkBizID)
	}

	userAuth, err := c.generateGateWayAuth(paras.Operator)
	if err != nil {
		return nil, fmt.Errorf("bksops CreateBkOpsTask generateGateWayAuth failed: %v", err)
	}

	request.FlowType = "common"
	request.TemplateSource = "business"
	if c.external {
		request.TemplateSource = "common"
	}

	_, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Post(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", userAuth).
		SetDebug(c.serverDebug).
		Send(request).
		EndStruct(&respData)
	if len(errs) > 0 {
		blog.Errorf("call api CreateBkOpsTask failed: %v", errs[0])
		return nil, errs[0]
	}

	if !respData.Result {
		blog.Errorf("call api CreateBkOpsTask failed: %v", respData.Message)
		return nil, fmt.Errorf(respData.Message)
	}
	//successfully request
	blog.Infof("call api CreateBkOpsTask with url(%s) successfully", reqURL)
	return respData, nil
}

// StartBkOpsTask start bkops task
func (c *Client) StartBkOpsTask(url string, paras *TaskPathParas, request *StartTaskRequest) (*StartTaskResponse, error) {
	if c == nil {
		return nil, ErrServerNotInit
	}

	if url == "" {
		url = c.urls.startTaskURL
	}

	var (
		reqURL   string
		reqData  interface{}
		respData = &StartTaskResponse{}
	)

	if c.external {
		reqURL = url
		reqData = &TaskReqParas{
			BkBizID: paras.BkBizID,
			TaskID:  paras.TaskID,
		}
	} else {
		reqURL = fmt.Sprintf(url, paras.TaskID, paras.BkBizID)
		request.Scope = "cmdb_biz"
		reqData = request
	}

	userAuth, err := c.generateGateWayAuth(paras.Operator)
	if err != nil {
		return nil, fmt.Errorf("bksops StartBkOpsTask generateGateWayAuth failed: %v", err)
	}

	_, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Post(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", userAuth).
		SetDebug(c.serverDebug).
		Send(reqData).
		EndStruct(&respData)
	if len(errs) > 0 {
		blog.Errorf("call api StartBkOpsTask failed: %v", errs[0])
		return nil, errs[0]
	}

	if !respData.Result {
		blog.Errorf("call api StartBkOpsTask failed: %v", respData.Message)
		return nil, fmt.Errorf(respData.Message)
	}

	//successfully request
	blog.Infof("call api StartBkOpsTask with url(%s) successfully", reqURL)
	return respData, nil
}

// GetTaskStatus get bkops task status
func (c *Client) GetTaskStatus(url string, paras *TaskPathParas, request *StartTaskRequest) (*TaskStatusResponse, error) {
	if c == nil {
		return nil, ErrServerNotInit
	}

	userAuth, err := c.generateGateWayAuth(paras.Operator)
	if err != nil {
		return nil, fmt.Errorf("bksops StartBkOpsTask generateGateWayAuth failed: %v", err)
	}

	if url == "" {
		url = c.urls.getTaskStatus
	}

	var (
		reqURL   string
		respData = &TaskStatusResponse{}
	)

	if c.external {
		reqURL = url
	} else {
		reqURL = fmt.Sprintf(url, paras.TaskID, paras.BkBizID)
		request.Scope = "cmdb_biz"
	}

	agent := gorequest.New().Timeout(defaultTimeOut).Get(reqURL)
	if c.external {
		agent = agent.Query(fmt.Sprintf("bk_biz_id=%s&task_id=%s", paras.BkBizID, paras.TaskID))
	}

	_, _, errs := agent.
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", userAuth).
		SetDebug(c.serverDebug).
		Send(request).
		EndStruct(&respData)
	if len(errs) > 0 {
		blog.Errorf("call api GetTaskStatus failed: %v", errs[0])
		return nil, errs[0]
	}

	if !respData.Result {
		blog.Errorf("call api GetTaskStatus failed: %v", respData.Message)
		return nil, fmt.Errorf(respData.Message)
	}

	//successfully request
	blog.Infof("call api GetTaskStatus with url(%s) successfully", reqURL)
	return respData, nil
}

// CreateTaskPathParas task path paras
type CreateTaskPathParas struct {
	// BkBizID template bizID
	BkBizID string `json:"bk_biz_id"`
	// TemplateID
	TemplateID string `json:"template_id"`
	// Operator template perm user
	Operator string `json:"operator"`
}

// CreateTaskRequest create task req
type CreateTaskRequest struct {
	BusinessID string `json:"bk_biz_id"`
	TemplateID string `json:"template_id"`
	// TemplateSource 模版来源(business/common)
	TemplateSource string `json:"template_source"`
	// TaskName 任务名称
	TaskName string `json:"name"`
	// FlowType 任务流程类型 (默认 common即可)
	FlowType  string            `json:"flow_type"`
	Constants map[string]string `json:"constants"`
}

// CreateTaskResponse create task resp
type CreateTaskResponse struct {
	Result  bool     `json:"result"`
	Data    *ResData `json:"data"`
	Message string   `json:"message"`
}

// ResData resp data
type ResData struct {
	TaskID  int    `json:"task_id"`
	TaskURL string `json:"task_url"`
}

// TaskReqParas task request body
type TaskReqParas struct {
	BkBizID string `json:"bk_biz_id"`
	TaskID  string `json:"task_id"`
}

// TaskPathParas task path paras
type TaskPathParas struct {
	BkBizID  string `json:"bk_biz_id"`
	TaskID   string `json:"task_id"`
	Operator string `json:"operator"`
}

// StartTaskRequest request
type StartTaskRequest struct {
	Scope string `json:"scope"`
}

// StartTaskResponse start task response
type StartTaskResponse struct {
	Result  bool   `json:"result"`
	Message string `json:"message"`
}

// TaskStatusResponse task status response
type TaskStatusResponse struct {
	Result  bool        `json:"result"`
	Data    *StatusData `json:"data"`
	Message string      `json:"message"`
}

// StatusData status
type StatusData struct {
	State string `json:"state"`
}
