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

// Options bkops options
type Options struct {
	EsbServer  string
	Server     string
	AppCode    string
	AppSecret  string
	BKUserName string
	Debug      bool
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
		esbServer:   options.EsbServer,
		server:      options.Server,
		appCode:     options.AppCode,
		appSecret:   options.AppSecret,
		bkUserName:  options.BKUserName,
		serverDebug: options.Debug,
	}

	return c, nil
}

// Client for bksops
type Client struct {
	esbServer   string
	server      string
	appCode     string
	appSecret   string
	bkUserName  string
	serverDebug bool
}

// DependURLs depend bkops urls
type DependURLs struct {
	createTaskURL string // nolint
	startTaskURL  string // nolint
	getTaskStatus string // nolint
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
func (c *Client) CreateBkOpsTask(paras *CreateTaskPathParas,
	request *CreateTaskRequest) (*CreateTaskResponse, error) {
	if c == nil {
		return nil, ErrServerNotInit
	}

	var (
		reqUrl   string
		respData = &CreateTaskResponse{}
	)

	// esbServer & apigw server
	if len(c.esbServer) > 0 {
		reqUrl = c.esbServer + "/api/c/compapi/v2/sops/create_task/"
		request.BusinessID = paras.BkBizID
		request.TemplateID = paras.TemplateID
	} else {
		reqUrl = c.server + fmt.Sprintf("/create_task/%s/%s/", paras.TemplateID, paras.BkBizID)
	}

	userAuth, err := c.generateGateWayAuth(paras.Operator)
	if err != nil {
		return nil, fmt.Errorf("bksops CreateBkOpsTask generateGateWayAuth failed: %v", err)
	}

	request.FlowType = string(CommonFlow)
	// TemplateSource 模版来源, 默认是业务流程; 可由用户自定义
	if request.TemplateSource == "" {
		request.TemplateSource = string(BusinessTpl)
	}

	_, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Post(reqUrl).
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
	// successfully request
	blog.Infof("call api CreateBkOpsTask with url(%s) successfully", reqUrl)
	return respData, nil
}

// StartBkOpsTask start bkops task
func (c *Client) StartBkOpsTask(paras *TaskPathParas,
	request *StartTaskRequest) (*StartTaskResponse, error) {
	if c == nil {
		return nil, ErrServerNotInit
	}

	var (
		reqUrl   string
		reqData  interface{}
		respData = &StartTaskResponse{}
	)

	if len(c.esbServer) > 0 {
		reqUrl = c.esbServer + "/api/c/compapi/v2/sops/start_task/"
		reqData = &TaskReqParas{
			BkBizID: paras.BkBizID,
			TaskID:  paras.TaskID,
		}
	} else {
		reqUrl = c.server + fmt.Sprintf("/start_task/%s/%s/", paras.TaskID, paras.BkBizID)
		request.Scope = string(CmdbBizScope)
		reqData = request
	}

	userAuth, err := c.generateGateWayAuth(paras.Operator)
	if err != nil {
		return nil, fmt.Errorf("bksops StartBkOpsTask generateGateWayAuth failed: %v", err)
	}

	_, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Post(reqUrl).
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

	// successfully request
	blog.Infof("call api StartBkOpsTask with url(%s) successfully", reqUrl)
	return respData, nil
}

// GetTaskStatus get bkops task status
func (c *Client) GetTaskStatus(paras *TaskPathParas,
	request *StartTaskRequest) (*TaskStatusResponse, error) {
	if c == nil {
		return nil, ErrServerNotInit
	}

	userAuth, err := c.generateGateWayAuth(paras.Operator)
	if err != nil {
		return nil, fmt.Errorf("bksops StartBkOpsTask generateGateWayAuth failed: %v", err)
	}

	var (
		reqUrl   string
		respData = &TaskStatusResponse{}
	)

	if request.Scope == "" {
		request.Scope = string(CmdbBizScope)
	}

	if len(c.esbServer) > 0 {
		reqUrl = c.esbServer + "/api/c/compapi/v2/sops/get_task_status/"
	} else {
		reqUrl = c.server + fmt.Sprintf("/get_task_status/%s/%s/", paras.TaskID, paras.BkBizID)
	}

	agent := gorequest.New().Timeout(defaultTimeOut).Get(reqUrl)
	if len(c.esbServer) > 0 {
		agent = agent.Query(fmt.Sprintf("bk_biz_id=%s&task_id=%s&scope=%s",
			paras.BkBizID, paras.TaskID, request.Scope))
	} else {
		agent = agent.Query(fmt.Sprintf("scope=%s", request.Scope))
	}

	_, _, errs := agent.
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", userAuth).
		SetDebug(c.serverDebug).
		EndStruct(&respData)
	if len(errs) > 0 {
		blog.Errorf("call api GetTaskStatus failed: %v", errs[0])
		return nil, errs[0]
	}

	if !respData.Result {
		blog.Errorf("call api GetTaskStatus failed: %v", respData.Message)
		return nil, fmt.Errorf(respData.Message)
	}

	// successfully request
	blog.Infof("call api GetTaskStatus with url(%s) successfully", reqUrl)
	return respData, nil
}

// GetBusinessTemplateList 查询业务下的模板列表
func (c *Client) GetBusinessTemplateList(path *TemplateListPathPara,
	templateReq *TemplateRequest) ([]*TemplateData, error) {
	if c == nil {
		return nil, ErrServerNotInit
	}

	var (
		_   = "GetBusinessTemplateList"
		url string
	)

	templateReq.SetDefaultTemplateBody()
	if len(c.esbServer) > 0 {
		url = c.esbServer + "/api/c/compapi/v2/sops/get_template_list/"
	} else {
		url = c.server + fmt.Sprintf("/get_template_list/%s/", path.BkBizID)
	}

	agent := gorequest.New().Timeout(defaultTimeOut).Get(url)
	if len(c.esbServer) > 0 {
		agent = agent.Query(fmt.Sprintf("bk_biz_id=%s&template_source=%s&scope=%s",
			path.BkBizID, templateReq.TemplateSource, string(templateReq.Scope)))
	} else {
		agent = agent.Query(fmt.Sprintf("template_source=%s&scope=%s",
			templateReq.TemplateSource, string(templateReq.Scope)))
	}

	userAuth, err := c.generateGateWayAuth(c.bkUserName)
	if err != nil {
		return nil, fmt.Errorf("bksops GetBusinessTemplateList generateGateWayAuth failed: %v", err)
	}

	resp := &TemplateListResponse{}
	_, _, errs := agent.
		Set("X-Bkapi-Authorization", userAuth).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		SetDebug(c.serverDebug).
		EndStruct(&resp)
	if len(errs) > 0 {
		blog.Errorf("call api GetBusinessTemplateList failed: %v", errs[0])
		return nil, errs[0]
	}

	if !resp.Result {
		blog.Errorf("call api GetBusinessTemplateList failed: %v", resp.Message)
		return nil, fmt.Errorf(resp.Message)
	}

	// successfully request
	blog.Infof("call api GetBusinessTemplateList with url(%s) successfully", url)

	return resp.Data, nil
}

// GetUserProjectDetailInfo get user project detailed info
func (c *Client) GetUserProjectDetailInfo(bizID string) (*ProjectInfo, error) {
	var (
		_   = "GetUserProjectDetailInfo"
		url string
	)

	if len(c.esbServer) > 0 {
		url = c.esbServer + "/api/c/compapi/v2/sops/get_user_project_detail/"
	} else {
		url = c.server + fmt.Sprintf("/get_user_project_detail/%s/", bizID)
	}

	userAuth, err := c.generateGateWayAuth(c.bkUserName)
	if err != nil {
		return nil, fmt.Errorf("bksops GetUserProjectDetailInfo generateGateWayAuth failed: %v", err)
	}

	agent := gorequest.New().Timeout(defaultTimeOut).Get(url)
	if len(c.esbServer) > 0 {
		agent = agent.Query(fmt.Sprintf("bk_biz_id=%s", bizID))
	}

	resp := &UserProjectResponse{}
	_, _, errs := agent.
		Set("X-Bkapi-Authorization", userAuth).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		SetDebug(c.serverDebug).
		EndStruct(&resp)
	if len(errs) > 0 {
		blog.Errorf("call api GetUserProjectDetailInfo failed: %v", errs[0])
		return nil, errs[0]
	}

	if !resp.Result {
		blog.Errorf("call api GetUserProjectDetailInfo failed: %v", resp.Message)
		return nil, fmt.Errorf(resp.Message)
	}

	// successfully request
	blog.Infof("call api GetUserProjectDetailInfo with url(%s) successfully", url)

	return &resp.Data, nil
}

// GetBusinessTemplateInfo 查询业务下的模板详情
func (c *Client) GetBusinessTemplateInfo(path *TemplateDetailPathPara,
	templateReq *TemplateRequest) ([]ConstantValue, error) {
	if c == nil {
		return nil, ErrServerNotInit
	}

	var (
		_   = "GetBusinessTemplateInfo"
		url string
	)

	templateReq.SetDefaultTemplateBody()
	if len(c.esbServer) > 0 {
		url = c.esbServer + "/api/c/compapi/v2/sops/get_template_info/"
	} else {
		url = c.server + fmt.Sprintf("/get_template_info/%s/%s/", path.TemplateID, path.BkBizID)
	}

	userAuth, err := c.generateGateWayAuth(c.bkUserName)
	if err != nil {
		return nil, fmt.Errorf("bksops GetBusinessTemplateInfo generateGateWayAuth failed: %v", err)
	}

	agent := gorequest.New().Timeout(defaultTimeOut).Get(url)
	if len(c.esbServer) > 0 {
		agent = agent.Query(fmt.Sprintf("bk_biz_id=%s&template_id=%s", path.BkBizID, path.TemplateID))
	}

	resp := &TemplateDetailResponse{}
	_, _, errs := agent.
		Set("X-Bkapi-Authorization", userAuth).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		SetDebug(c.serverDebug).
		EndStruct(&resp)
	if len(errs) > 0 {
		blog.Errorf("call api GetBusinessTemplateInfo failed: %v", errs[0])
		return nil, errs[0]
	}

	if !resp.Result {
		blog.Errorf("call api GetBusinessTemplateInfo failed: %v", resp.Message)
		return nil, fmt.Errorf(resp.Message)
	}

	// successfully request
	blog.Infof("call api GetBusinessTemplateInfo with url(%s) successfully", url)

	globalCustomVars := make([]ConstantValue, 0)
	for i := range resp.Data.PipeTree.Constants {
		if resp.Data.PipeTree.Constants[i].SourceType == custom {
			globalCustomVars = append(globalCustomVars, resp.Data.PipeTree.Constants[i])
		}
	}

	return globalCustomVars, nil
}

// OperateBkOpsTask operate bkops task
func (c *Client) OperateBkOpsTask(paras *TaskPathParas, request *OperateTaskRequest) error {
	if c == nil {
		return ErrServerNotInit
	}

	var (
		reqUrl   string
		respData = &OperateTaskResponse{}
	)

	request.Scope = CmdbBizScope
	if len(c.esbServer) > 0 {
		reqUrl = c.esbServer + "/api/c/compapi/v2/sops/operate_task/"
		request.BizId = paras.BkBizID
		request.TaskId = paras.TaskID
	} else {
		reqUrl = c.server + fmt.Sprintf("/operate_task/%s/%s/", paras.TaskID, paras.BkBizID)
	}

	userAuth, err := c.generateGateWayAuth(paras.Operator)
	if err != nil {
		return fmt.Errorf("bksops OperateBkOpsTask generateGateWayAuth failed: %v", err)
	}

	_, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Post(reqUrl).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", userAuth).
		SetDebug(c.serverDebug).
		Send(request).
		EndStruct(&respData)
	if len(errs) > 0 {
		blog.Errorf("call api OperateBkOpsTask failed: %v", errs[0])
		return errs[0]
	}

	if !respData.Result {
		blog.Errorf("call api OperateBkOpsTask failed: %v", respData.Message)
		return fmt.Errorf(respData.Message)
	}

	// successfully request
	blog.Infof("call api OperateBkOpsTask with url(%s) successfully", reqUrl)
	return nil
}
