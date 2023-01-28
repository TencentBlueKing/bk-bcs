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

package gse

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	"github.com/parnurzeal/gorequest"
)

// Interface for gse api
type Interface interface {
	// GetAgentStatus get agent status
	GetAgentStatus(req *GetAgentStatusReq) (*GetAgentStatusResp, error)
}

// GseClient global gse client
var GseClient *Client

// SetGseClient set gse client
func SetGseClient(options Options) error {
	cli, err := NewGseClient(options)
	if err != nil {
		return err
	}

	GseClient = cli
	return nil
}

// GetGseClient get gse client
func GetGseClient() *Client {
	return GseClient
}

// NewGseClient create gse client
func NewGseClient(options Options) (*Client, error) {
	c := &Client{
		appCode:     options.AppCode,
		appSecret:   options.AppSecret,
		bkUserName:  options.BKUserName,
		server:      options.Server,
		serverDebug: options.Debug,
	}

	if !options.Enable {
		return nil, nil
	}

	auth, err := c.generateGateWayAuth()
	if err != nil {
		return nil, err
	}
	c.userAuth = auth
	return c, nil
}

var (
	defaultTimeOut = time.Second * 60
	// ErrServerNotInit server not init
	ErrServerNotInit = errors.New("server not inited")
)

// Options for gse client
type Options struct {
	Enable     bool
	AppCode    string
	AppSecret  string
	BKUserName string
	Server     string
	Debug      bool
}

// AuthInfo auth user
type AuthInfo struct {
	BkAppCode   string `json:"bk_app_code"`
	BkAppSecret string `json:"bk_app_secret"`
	BkUserName  string `json:"bk_username"`
}

// Client for gse
type Client struct {
	appCode     string
	appSecret   string
	bkUserName  string
	server      string
	serverDebug bool
	userAuth    string
}

func (c *Client) generateGateWayAuth() (string, error) {
	if c == nil {
		return "", ErrServerNotInit
	}

	auth := &AuthInfo{
		BkAppCode:   c.appCode,
		BkAppSecret: c.appSecret,
		BkUserName:  c.bkUserName,
	}

	userAuth, err := json.Marshal(auth)
	if err != nil {
		return "", err
	}

	return string(userAuth), nil
}

// GetAgentStatus get host agent status
func (c *Client) GetAgentStatus(req *GetAgentStatusReq) (*GetAgentStatusResp, error) {
	if c == nil {
		return nil, ErrServerNotInit
	}

	var (
		reqURL   = fmt.Sprintf("%s/get_agent_status/", c.server)
		respData = &GetAgentStatusResp{}
	)

	_, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Post(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", c.userAuth).
		SetDebug(c.serverDebug).
		Send(req).
		EndStruct(&respData)
	if len(errs) > 0 {
		blog.Errorf("call api GetAgentStatus failed: %v", errs[0])
		return nil, errs[0]
	}

	if respData.Code != 0 {
		blog.Errorf("call api GetAgentStatus failed: %s, request_id: %s", respData.Message,
			respData.RequestID)
		return nil, fmt.Errorf("%s", respData.Message)
	}

	if !respData.Result {
		blog.Errorf("call api GetAgentStatus failed: %v, request_id: %s", respData.Message,
			respData.RequestID)
		return nil, fmt.Errorf(respData.Message)
	}

	if len(respData.Data) == 0 {
		blog.Errorf("call api GetAgentStatus failed: %v, request_id: %s", respData.Message,
			respData.RequestID)
		return nil, fmt.Errorf("no agent found")
	}

	blog.Infof("call api GetAgentStatus with url(%s) successfully", reqURL)
	return respData, nil
}
