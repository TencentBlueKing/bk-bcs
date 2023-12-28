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

// Package gse xxx
package gse

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/kirito41dd/xslice"
	"github.com/parnurzeal/gorequest"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

// Interface for gse api
type Interface interface {
	// GetAgentStatusV1 get agent status for version 1
	GetAgentStatusV1(req *GetAgentStatusReq) (*GetAgentStatusResp, error)
	// GetAgentStatusV2 get agent status for version 2
	GetAgentStatusV2(req *GetAgentStatusReqV2) (*GetAgentStatusRespV2, error)
	// GetHostsGseAgentStatus get hosts agent status
	GetHostsGseAgentStatus(supplyAccount string, hosts []Host) ([]HostAgentStatus, error)
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
		appCode:       options.AppCode,
		appSecret:     options.AppSecret,
		bkUserName:    options.BKUserName,
		EsbServer:     options.EsbServer,
		GatewayServer: options.GatewayServer,
		serverDebug:   options.Debug,
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
	Enable        bool
	AppCode       string
	AppSecret     string
	BKUserName    string
	EsbServer     string
	GatewayServer string
	Debug         bool
}

// AuthInfo auth user
type AuthInfo struct {
	BkAppCode   string `json:"bk_app_code"`
	BkAppSecret string `json:"bk_app_secret"`
	BkUserName  string `json:"bk_username"`
}

// Client for gse
type Client struct {
	appCode       string
	appSecret     string
	bkUserName    string
	EsbServer     string
	GatewayServer string
	serverDebug   bool
	userAuth      string
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

// GetHostsGseAgentStatus get host agent status
// nolint
func (c *Client) GetHostsGseAgentStatus(supplyAccount string, hosts []Host) ([]HostAgentStatus, error) {
	if c == nil {
		return nil, ErrServerNotInit
	}

	var (
		agentHost = make([]Host, 0)
		cloudHost = make([]Host, 0)
	)

	for i := range hosts {
		if len(hosts[i].AgentID) > 0 {
			agentHost = append(agentHost, hosts[i])
			continue
		}

		cloudHost = append(cloudHost, hosts[i])
	}

	var (
		hostAgentStatus = make([]HostAgentStatus, 0)
		agentLock       = &sync.RWMutex{}
	)

	// handle exist agentId hosts
	chunksAgentHost := xslice.SplitToChunks(agentHost, limit)
	agentHostList, ok := chunksAgentHost.([][]Host)
	if !ok {
		return nil, fmt.Errorf("GetHostsGseAgentStatus SplitToChunks failed")
	}

	con := utils.NewRoutinePool(20)
	defer con.Close()

	for i := range agentHostList {
		con.Add(1)
		go func(hosts []Host) {
			defer con.Done()

			agentIDs := make([]string, 0)
			for i := range hosts {
				agentIDs = append(agentIDs, hosts[i].AgentID)
			}

			resp, err := c.GetAgentStatusV2(&GetAgentStatusReqV2{AgentIDList: agentIDs})
			if err != nil {
				blog.Errorf("GetHostsGseAgentStatus %v failed, %s", supplyAccount, err.Error())
				return
			}
			for _, agent := range resp.Data {
				agentLock.Lock()
				hostAgentStatus = append(hostAgentStatus, HostAgentStatus{
					Host: Host{
						AgentID:   agent.BkAgentID,
						BKCloudID: agent.BKCloudID,
					},
					Alive: func() int {
						if agent.Alive() {
							return 1
						}
						return 0
					}(),
				})
				agentLock.Unlock()
			}
		}(agentHostList[i])
	}
	con.Wait()

	// handle exist cloud hosts
	chunksCloudHost := xslice.SplitToChunks(cloudHost, limit)
	cloudHostList, ok := chunksCloudHost.([][]Host)
	if !ok {
		return nil, fmt.Errorf("GetHostsGseAgentStatus SplitToChunks failed")
	}

	for i := range cloudHostList {
		con.Add(1)
		go func(hosts []Host) {
			defer con.Done()

			agentIDs := make([]string, 0)
			for i := range hosts {
				agentIDs = append(agentIDs, hosts[i].AgentID) // nolint
			}

			resp, err := c.GetAgentStatusV1(&GetAgentStatusReq{
				BKSupplierAccount: supplyAccount,
				Hosts:             hosts,
			})
			if err != nil {
				blog.Errorf("GetHostsGseAgentStatus %v failed, %s", supplyAccount, err.Error())
				return
			}
			for _, agent := range resp.Data {
				agentLock.Lock()
				hostAgentStatus = append(hostAgentStatus, HostAgentStatus{
					Host: Host{
						IP:        agent.IP,
						BKCloudID: agent.BKCloudID,
					},
					Alive: agent.BKAgentAlive,
				})
				agentLock.Unlock()
			}

		}(cloudHostList[i])
	}
	con.Wait()

	return hostAgentStatus, nil
}

// GetAgentStatusV2 get host agent status by agentID
func (c *Client) GetAgentStatusV2(req *GetAgentStatusReqV2) (*GetAgentStatusRespV2, error) {
	if c == nil {
		return nil, ErrServerNotInit
	}

	var (
		reqURL   = fmt.Sprintf("%s/cluster/list_agent_state", c.GatewayServer)
		respData = &GetAgentStatusRespV2{}
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

	if len(respData.Data) == 0 {
		blog.Errorf("call api GetAgentStatus failed: %v, request_id: %s", respData.Message,
			respData.RequestID)
		return nil, fmt.Errorf("no agent found")
	}

	blog.Infof("call api GetAgentStatus with url(%s) successfully", reqURL)
	return respData, nil
}

// GetAgentStatusV1 get host agent status by cloud:ip
func (c *Client) GetAgentStatusV1(req *GetAgentStatusReq) (*GetAgentStatusResp, error) {
	if c == nil {
		return nil, ErrServerNotInit
	}

	var (
		reqURL   = fmt.Sprintf("%s/gse/get_agent_status", c.EsbServer)
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
