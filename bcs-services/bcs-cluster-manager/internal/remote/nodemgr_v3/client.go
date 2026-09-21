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

// Package nodemgr_v3 nodeman v3 接口客户端实现
package nodemgr_v3

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/types"
	rutils "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/utils"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/i18n"
	"github.com/parnurzeal/gorequest"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/metrics"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

// NodeManV3Client global nodeman v3 client
var NodeManV3Client *Client

// enabled 版本开关，为 true 时调用方走 v3 接口
var enabled bool

// SetEnabled 设置版本开关
func SetEnabled(v bool) {
	enabled = v
}

// IsEnabled 返回是否启用 v3 接口
func IsEnabled() bool {
	return enabled
}

// SetNodeManV3Client set nodeman v3 client
func SetNodeManV3Client(options Options) error {
	cli, err := NewNodeManV3Client(options)
	if err != nil {
		return err
	}

	NodeManV3Client = cli
	return nil
}

// GetNodeManV3Client get nodeman v3 client
func GetNodeManV3Client() *Client {
	return NodeManV3Client
}

// NewNodeManV3Client create nodeman v3 client
func NewNodeManV3Client(options Options) (*Client, error) {
	c := &Client{
		appCode:     options.AppCode,
		appSecret:   options.AppSecret,
		bkUserName:  options.BKUserName,
		server:      options.V3Server,
		serverDebug: options.Debug,
	}

	auth, err := c.generateGateWayAuth()
	if err != nil {
		return nil, err
	}
	c.userAuth = auth
	blog.Infof("nodemgr_v3 NewNodeManV3Client server=%s, appCode=%s, bkUserName=%s", c.server, c.appCode, c.bkUserName)
	return c, nil
}

var (
	defaultTimeOut = time.Second * 60
	defaultLimit   = 200
)

// Options for client
type Options struct {
	Enable     bool
	AppCode    string
	AppSecret  string
	BKUserName string
	Server     string
	V3Server   string
	Debug      bool
}

// AuthInfo auth user
type AuthInfo struct {
	BkAppCode   string `json:"bk_app_code"`
	BkAppSecret string `json:"bk_app_secret"`
	BkUserName  string `json:"bk_username"`
}

// Client for nodeman v3
type Client struct {
	appCode     string
	appSecret   string
	bkUserName  string
	server      string
	serverDebug bool
	userAuth    string
}

func (c *Client) generateGateWayAuth() (string, error) {
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

// ======================== Agent 安装 ========================

// AgentInstall 批量安装 Agent
// POST /api/v3/node/agent/install
func (c *Client) AgentInstall(ctx context.Context, req *AgentInstallRequest) (*AgentInstallRespData, error) {
	var (
		reqURL   = fmt.Sprintf("%s/api/v3/node/agent/install", c.server)
		respData = &AgentInstallResponse{}
	)

	language := i18n.LanguageFromCtx(ctx)
	userAuth, tenant, err := rutils.GetGatewayAuthAndTenantInfo(ctx, &types.AuthInfo{
		BkAppUser: types.BkAppUser{
			BkAppCode:   c.appCode,
			BkAppSecret: c.appSecret,
		},
		BkUserName: c.bkUserName,
	}, "")
	if err != nil {
		return nil, err
	}

	start := time.Now()
	_, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Post(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", userAuth).
		Set("X-Bk-Tenant-Id", tenant).
		Set("Blueking-Language", language).
		SetDebug(c.serverDebug).
		Send(req).
		EndStruct(&respData)
	if len(errs) > 0 {
		metrics.ReportLibRequestMetric("nodemgr_v3", "AgentInstall", "http", metrics.LibCallStatusErr, start)
		blog.Errorf("call api AgentInstall failed: %v", errs[0])
		return nil, errs[0]
	}
	metrics.ReportLibRequestMetric("nodemgr_v3", "AgentInstall", "http", metrics.LibCallStatusOK, start)

	if respData.Code != 0 {
		respBytes, _ := json.Marshal(respData)
		blog.Errorf("call api AgentInstall failed: %s", string(respBytes))
		return nil, fmt.Errorf("call api AgentInstall failed: %s", string(respBytes))
	}
	blog.Infof("call api AgentInstall with url(%s) successfully, workflow_id: %s", reqURL, respData.Data.WorkflowID)

	return &respData.Data, nil
}

// ======================== Proxy 安装 ========================

// ProxyInstall 批量安装 Proxy
// POST /api/v3/node/proxy/install
func (c *Client) ProxyInstall(ctx context.Context, req *ProxyInstallRequest) (*ProxyInstallRespData, error) {
	var (
		reqURL   = fmt.Sprintf("%s/api/v3/node/proxy/install", c.server)
		respData = &ProxyInstallResponse{}
	)

	language := i18n.LanguageFromCtx(ctx)
	userAuth, tenant, err := rutils.GetGatewayAuthAndTenantInfo(ctx, &types.AuthInfo{
		BkAppUser: types.BkAppUser{
			BkAppCode:   c.appCode,
			BkAppSecret: c.appSecret,
		},
		BkUserName: c.bkUserName,
	}, "")
	if err != nil {
		return nil, err
	}

	start := time.Now()
	_, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Post(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", userAuth).
		Set("X-Bk-Tenant-Id", tenant).
		Set("Blueking-Language", language).
		SetDebug(c.serverDebug).
		Send(req).
		EndStruct(&respData)
	if len(errs) > 0 {
		metrics.ReportLibRequestMetric("nodemgr_v3", "ProxyInstall", "http", metrics.LibCallStatusErr, start)
		blog.Errorf("call api ProxyInstall failed: %v", errs[0])
		return nil, errs[0]
	}
	metrics.ReportLibRequestMetric("nodemgr_v3", "ProxyInstall", "http", metrics.LibCallStatusOK, start)

	if respData.Code != 0 {
		respBytes, _ := json.Marshal(respData)
		blog.Errorf("call api ProxyInstall failed: %s", string(respBytes))
		return nil, fmt.Errorf("call api ProxyInstall failed: %s", string(respBytes))
	}
	blog.Infof("call api ProxyInstall with url(%s) successfully, workflow_id: %s", reqURL, respData.Data.WorkflowID)

	return &respData.Data, nil
}

// ======================== 工作流列表 ========================

// WorkflowList 查询任务流列表
// POST /api/v3/node/workflow/list
func (c *Client) WorkflowList(ctx context.Context, req *WorkflowListRequest) (*WorkflowListData, error) {
	var (
		reqURL   = fmt.Sprintf("%s/api/v3/node/workflow/list", c.server)
		respData = &WorkflowListResponse{}
	)

	language := i18n.LanguageFromCtx(ctx)
	userAuth, tenant, err := rutils.GetGatewayAuthAndTenantInfo(ctx, &types.AuthInfo{
		BkAppUser: types.BkAppUser{
			BkAppCode:   c.appCode,
			BkAppSecret: c.appSecret,
		},
		BkUserName: c.bkUserName,
	}, "")
	if err != nil {
		return nil, err
	}

	start := time.Now()
	_, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Post(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", userAuth).
		Set("X-Bk-Tenant-Id", tenant).
		Set("Blueking-Language", language).
		SetDebug(c.serverDebug).
		Send(req).
		EndStruct(&respData)
	if len(errs) > 0 {
		metrics.ReportLibRequestMetric("nodemgr_v3", "WorkflowList", "http", metrics.LibCallStatusErr, start)
		blog.Errorf("call api WorkflowList failed: %v", errs[0])
		return nil, errs[0]
	}
	metrics.ReportLibRequestMetric("nodemgr_v3", "WorkflowList", "http", metrics.LibCallStatusOK, start)

	if respData.Code != 0 {
		respBytes, _ := json.Marshal(respData)
		blog.Errorf("call api WorkflowList failed: %s", string(respBytes))
		return nil, fmt.Errorf("call api WorkflowList failed: %s", string(respBytes))
	}
	blog.Infof("call api WorkflowList with url(%s) successfully", reqURL)

	return &respData.Data, nil
}

// ======================== 工作流操作列表 ========================

// WorkflowOperationList 查询工作流操作列表
// POST /api/v3/node/workflow/operation/list
func (c *Client) WorkflowOperationList(ctx context.Context, req *WorkflowOperationListRequest) (*WorkflowOperationListData, error) {
	var (
		reqURL   = fmt.Sprintf("%s/api/v3/node/workflow/operation/list", c.server)
		respData = &WorkflowOperationListResponse{}
	)

	language := i18n.LanguageFromCtx(ctx)
	userAuth, tenant, err := rutils.GetGatewayAuthAndTenantInfo(ctx, &types.AuthInfo{
		BkAppUser: types.BkAppUser{
			BkAppCode:   c.appCode,
			BkAppSecret: c.appSecret,
		},
		BkUserName: c.bkUserName,
	}, "")
	if err != nil {
		return nil, err
	}

	start := time.Now()
	_, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Post(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", userAuth).
		Set("X-Bk-Tenant-Id", tenant).
		Set("Blueking-Language", language).
		SetDebug(c.serverDebug).
		Send(req).
		EndStruct(&respData)
	if len(errs) > 0 {
		metrics.ReportLibRequestMetric("nodemgr_v3", "WorkflowOperationList", "http", metrics.LibCallStatusErr, start)
		blog.Errorf("call api WorkflowOperationList failed: %v", errs[0])
		return nil, errs[0]
	}
	metrics.ReportLibRequestMetric("nodemgr_v3", "WorkflowOperationList", "http", metrics.LibCallStatusOK, start)

	if respData.Code != 0 {
		respBytes, _ := json.Marshal(respData)
		blog.Errorf("call api WorkflowOperationList failed: %s", string(respBytes))
		return nil, fmt.Errorf("call api WorkflowOperationList failed: %s", string(respBytes))
	}
	blog.Infof("call api WorkflowOperationList with url(%s) successfully", reqURL)

	return &respData.Data, nil
}

// ======================== 工作流操作实例列表 ========================

// OperationInstanceList 查询操作实例列表
// POST /api/v3/node/workflow/operation/instance/list
func (c *Client) OperationInstanceList(ctx context.Context, req *OperationInstanceListRequest) (*OperationInstanceListData, error) {
	var (
		reqURL   = fmt.Sprintf("%s/api/v3/node/workflow/operation/instance/list", c.server)
		respData = &OperationInstanceListResponse{}
	)

	language := i18n.LanguageFromCtx(ctx)
	userAuth, tenant, err := rutils.GetGatewayAuthAndTenantInfo(ctx, &types.AuthInfo{
		BkAppUser: types.BkAppUser{
			BkAppCode:   c.appCode,
			BkAppSecret: c.appSecret,
		},
		BkUserName: c.bkUserName,
	}, "")
	if err != nil {
		return nil, err
	}

	start := time.Now()
	_, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Post(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", userAuth).
		Set("X-Bk-Tenant-Id", tenant).
		Set("Blueking-Language", language).
		SetDebug(c.serverDebug).
		Send(req).
		EndStruct(&respData)
	if len(errs) > 0 {
		metrics.ReportLibRequestMetric("nodemgr_v3", "OperationInstanceList", "http", metrics.LibCallStatusErr, start)
		blog.Errorf("call api OperationInstanceList failed: %v", errs[0])
		return nil, errs[0]
	}
	metrics.ReportLibRequestMetric("nodemgr_v3", "OperationInstanceList", "http", metrics.LibCallStatusOK, start)

	if respData.Code != 0 {
		respBytes, _ := json.Marshal(respData)
		blog.Errorf("call api OperationInstanceList failed: %s", string(respBytes))
		return nil, fmt.Errorf("call api OperationInstanceList failed: %s", string(respBytes))
	}
	blog.Infof("call api OperationInstanceList with url(%s) successfully", reqURL)

	return &respData.Data, nil
}

// ======================== 主机列表 ========================

// HostList 查询主机列表
// POST /api/v3/topo/host/list
func (c *Client) HostList(ctx context.Context, req *HostListRequest) (*HostListData, error) {
	var (
		reqURL   = fmt.Sprintf("%s/api/v3/topo/host/list", c.server)
		respData = &HostListResponse{}
	)

	language := i18n.LanguageFromCtx(ctx)
	userAuth, tenant, err := rutils.GetGatewayAuthAndTenantInfo(ctx, &types.AuthInfo{
		BkAppUser: types.BkAppUser{
			BkAppCode:   c.appCode,
			BkAppSecret: c.appSecret,
		},
		BkUserName: c.bkUserName,
	}, "")
	if err != nil {
		return nil, err
	}

	start := time.Now()
	_, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Post(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", userAuth).
		Set("X-Bk-Tenant-Id", tenant).
		Set("Blueking-Language", language).
		SetDebug(c.serverDebug).
		Send(req).
		EndStruct(&respData)
	if len(errs) > 0 {
		metrics.ReportLibRequestMetric("nodemgr_v3", "HostList", "http", metrics.LibCallStatusErr, start)
		blog.Errorf("call api HostList failed: %v", errs[0])
		return nil, errs[0]
	}
	metrics.ReportLibRequestMetric("nodemgr_v3", "HostList", "http", metrics.LibCallStatusOK, start)

	if respData.Code != 0 {
		respBytes, _ := json.Marshal(respData)
		blog.Errorf("call api HostList failed: %s", string(respBytes))
		return nil, fmt.Errorf("call api HostList failed: %s", string(respBytes))
	}
	blog.Infof("call api HostList with url(%s) successfully", reqURL)

	// debug 输出完整响应结果，便于排查接口返回数据
	respBytes, _ := json.Marshal(respData)
	blog.Debug(fmt.Sprintf("call api HostList with url(%s) response: %s", reqURL, string(respBytes)))

	return &respData.Data, nil
}

// ListAllHosts 查询全部主机（分页并发拉取）
func (c *Client) ListAllHosts(ctx context.Context, bkBizID int) ([]HostItem, error) {
	// 获取总数
	req := &HostListRequest{
		OnlyCount: true,
		ExactIncludeConditions: &HostListExactCondition{
			BkBizID: []int{bkBizID},
		},
	}
	result, err := c.HostList(ctx, req)
	if err != nil {
		return nil, err
	}

	blog.Infof("ListAllHosts count %d by bizID %d", result.Total, bkBizID)
	var (
		hostList = make([]HostItem, 0)
		hostLock = &sync.RWMutex{}
	)

	con := utils.NewRoutinePool(20)
	defer con.Close()

	page := int(result.Total-1)/defaultLimit + 1
	for i := 0; i < page; i++ {
		con.Add(1)
		go func(offset int) {
			defer con.Done()
			req := &HostListRequest{
				Page: &V3Page{
					Offset: offset,
					Limit:  defaultLimit,
				},
				ExactIncludeConditions: &HostListExactCondition{
					BkBizID: []int{bkBizID},
				},
			}
			hosts, err := c.HostList(ctx, req)
			if err != nil {
				blog.Errorf("ListAllHosts %v failed, %s", bkBizID, err.Error())
				return
			}
			hostLock.Lock()
			hostList = append(hostList, hosts.Items...)
			hostLock.Unlock()
		}(i * defaultLimit)
	}
	con.Wait()

	blog.Infof("ListAllHosts successful %v", bkBizID)
	return hostList, nil
}

// GetHostIDByIPs 根据 IP 列表获取主机 ID
func (c *Client) GetHostIDByIPs(ctx context.Context, bkBizID int, ips []string) ([]int, error) {
	hostIDs := make([]int, 0)
	hosts, err := c.ListAllHosts(ctx, bkBizID)
	if err != nil {
		return nil, fmt.Errorf("list nodeman v3 hosts err %s", err.Error())
	}

	ipMap := make(map[string]bool)
	for _, ip := range ips {
		ipMap[ip] = true
	}

	for _, v := range hosts {
		for _, innerIP := range v.Info.BkHostInnerIPList {
			if ipMap[innerIP] {
				hostIDs = append(hostIDs, v.BkHostID)
				break
			}
		}
	}
	return hostIDs, nil
}

// ======================== 管控区域列表 ========================

// NetworkAreaList 查询管控区域列表
// POST /api/v3/topo/networkarea/list
func (c *Client) NetworkAreaList(ctx context.Context, req *NetworkAreaListRequest) (*NetworkAreaListData, error) {
	var (
		reqURL   = fmt.Sprintf("%s/api/v3/topo/networkarea/list", c.server)
		respData = &NetworkAreaListResponse{}
	)

	language := i18n.LanguageFromCtx(ctx)
	userAuth, tenant, err := rutils.GetGatewayAuthAndTenantInfo(ctx, &types.AuthInfo{
		BkAppUser: types.BkAppUser{
			BkAppCode:   c.appCode,
			BkAppSecret: c.appSecret,
		},
		BkUserName: c.bkUserName,
	}, "")
	if err != nil {
		return nil, err
	}

	start := time.Now()
	_, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Post(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", userAuth).
		Set("X-Bk-Tenant-Id", tenant).
		Set("Blueking-Language", language).
		SetDebug(c.serverDebug).
		Send(req).
		EndStruct(&respData)
	if len(errs) > 0 {
		metrics.ReportLibRequestMetric("nodemgr_v3", "NetworkAreaList", "http", metrics.LibCallStatusErr, start)
		blog.Errorf("call api NetworkAreaList failed: %v", errs[0])
		return nil, errs[0]
	}
	metrics.ReportLibRequestMetric("nodemgr_v3", "NetworkAreaList", "http", metrics.LibCallStatusOK, start)

	if respData.Code != 0 {
		respBytes, _ := json.Marshal(respData)
		blog.Errorf("call api NetworkAreaList failed: %s", string(respBytes))
		return nil, fmt.Errorf("call api NetworkAreaList failed: %s", string(respBytes))
	}
	blog.Infof("call api NetworkAreaList with url(%s) successfully", reqURL)

	return &respData.Data, nil
}

// ======================== 管控单元列表 ========================

// NetworkUnitList 查询管控单元列表
// POST /api/v3/topo/networkunit/list
func (c *Client) NetworkUnitList(ctx context.Context, req *NetworkUnitListRequest) (*NetworkUnitListData, error) {
	var (
		reqURL   = fmt.Sprintf("%s/api/v3/topo/networkunit/list", c.server)
		respData = &NetworkUnitListResponse{}
	)

	language := i18n.LanguageFromCtx(ctx)
	userAuth, tenant, err := rutils.GetGatewayAuthAndTenantInfo(ctx, &types.AuthInfo{
		BkAppUser: types.BkAppUser{
			BkAppCode:   c.appCode,
			BkAppSecret: c.appSecret,
		},
		BkUserName: c.bkUserName,
	}, "")
	if err != nil {
		return nil, err
	}

	start := time.Now()
	_, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Post(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", userAuth).
		Set("X-Bk-Tenant-Id", tenant).
		Set("Blueking-Language", language).
		SetDebug(c.serverDebug).
		Send(req).
		EndStruct(&respData)
	if len(errs) > 0 {
		metrics.ReportLibRequestMetric("nodemgr_v3", "NetworkUnitList", "http", metrics.LibCallStatusErr, start)
		blog.Errorf("call api NetworkUnitList failed: %v", errs[0])
		return nil, errs[0]
	}
	metrics.ReportLibRequestMetric("nodemgr_v3", "NetworkUnitList", "http", metrics.LibCallStatusOK, start)

	if respData.Code != 0 {
		respBytes, _ := json.Marshal(respData)
		blog.Errorf("call api NetworkUnitList failed: %s", string(respBytes))
		return nil, fmt.Errorf("call api NetworkUnitList failed: %s", string(respBytes))
	}
	blog.Infof("call api NetworkUnitList with url(%s) successfully", reqURL)

	return &respData.Data, nil
}

// ======================== 公开密钥获取 ========================

// PublicKeyGet 获取节点管理 RSA 公钥，用于加密登录凭据
// POST /api/v3/cipher/rsa/get_public_key
func (c *Client) PublicKeyGet(ctx context.Context) (string, error) {
	var (
		reqURL   = fmt.Sprintf("%s/api/v3/cipher/rsa/get_public_key", c.server)
		respData = &PublicKeyGetResponse{}
	)

	language := i18n.LanguageFromCtx(ctx)
	userAuth, tenant, err := rutils.GetGatewayAuthAndTenantInfo(ctx, &types.AuthInfo{
		BkAppUser: types.BkAppUser{
			BkAppCode:   c.appCode,
			BkAppSecret: c.appSecret,
		},
		BkUserName: c.bkUserName,
	}, "")
	if err != nil {
		return "", err
	}

	// gorequest v0.2.16 会把空对象 {} 解析为空 map，导致请求体为空（服务端报 EOF），
	// 占位字段，保证请求体为非空 JSON
	placeholder := map[string]interface{}{"ping": "pong"}
	start := time.Now()
	_, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Post(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", userAuth).
		Set("X-Bk-Tenant-Id", tenant).
		Set("Blueking-Language", language).
		SetDebug(c.serverDebug).
		Send(placeholder).
		EndStruct(&respData)
	if len(errs) > 0 {
		metrics.ReportLibRequestMetric("nodemgr_v3", "PublicKeyGet", "http", metrics.LibCallStatusErr, start)
		blog.Errorf("call api PublicKeyGet failed: %v", errs[0])
		return "", errs[0]
	}
	metrics.ReportLibRequestMetric("nodemgr_v3", "PublicKeyGet", "http", metrics.LibCallStatusOK, start)

	if respData.Code != 0 {
		respBytes, _ := json.Marshal(respData)
		blog.Errorf("call api PublicKeyGet failed: %s", string(respBytes))
		return "", fmt.Errorf("call api PublicKeyGet failed: %s", string(respBytes))
	}

	publicKey := respData.Data.PublicKey
	if publicKey == "" {
		blog.Errorf("call api PublicKeyGet failed: empty public_key")
		return "", fmt.Errorf("call api PublicKeyGet failed: empty public_key")
	}

	return publicKey, nil
}

// ListAllWorkflows 查询全部任务流（分页并发拉取）
func (c *Client) ListAllWorkflows(ctx context.Context, req *WorkflowListRequest) ([]WorkflowItem, error) {
	countReq := *req
	countReq.Page = &V3Page{
		Offset: 0,
		Limit:  defaultLimit,
	}
	countReq.OnlyCount = true
	result, err := c.WorkflowList(ctx, &countReq)
	if err != nil {
		return nil, err
	}

	blog.Infof("ListAllWorkflows count %d", result.Total)
	var (
		workflowList = make([]WorkflowItem, 0)
		workflowLock = &sync.RWMutex{}
	)

	con := utils.NewRoutinePool(20)
	defer con.Close()

	page := int(result.Total-1)/defaultLimit + 1
	for i := 0; i < page; i++ {
		con.Add(1)
		go func(offset int) {
			defer con.Done()
			pageReq := *req
			pageReq.Page = &V3Page{
				Offset: offset,
				Limit:  defaultLimit,
			}
			pageReq.OnlyCount = false
			workflows, err := c.WorkflowList(ctx, &pageReq)
			if err != nil {
				blog.Errorf("ListAllWorkflows failed, %s", err.Error())
				return
			}
			workflowLock.Lock()
			workflowList = append(workflowList, workflows.Items...)
			workflowLock.Unlock()
		}(i * defaultLimit)
	}
	con.Wait()

	blog.Infof("ListAllWorkflows successful")
	return workflowList, nil
}

// ListAllWorkflowOperations 查询全部工作流操作（分页并发拉取）
func (c *Client) ListAllWorkflowOperations(ctx context.Context, req *WorkflowOperationListRequest) ([]Operation, error) {
	countReq := *req
	countReq.Page = &V3Page{
		Offset: 0,
		Limit:  defaultLimit,
	}
	countReq.OnlyCount = true
	result, err := c.WorkflowOperationList(ctx, &countReq)
	if err != nil {
		return nil, err
	}

	blog.Infof("ListAllWorkflowOperations count %d", result.Total)
	var (
		operationList = make([]Operation, 0)
		operationLock = &sync.RWMutex{}
	)

	con := utils.NewRoutinePool(20)
	defer con.Close()

	page := int(result.Total-1)/defaultLimit + 1
	for i := 0; i < page; i++ {
		con.Add(1)
		go func(offset int) {
			defer con.Done()
			pageReq := *req
			pageReq.Page = &V3Page{
				Offset: offset,
				Limit:  defaultLimit,
			}
			pageReq.OnlyCount = false
			operations, err := c.WorkflowOperationList(ctx, &pageReq)
			if err != nil {
				blog.Errorf("ListAllWorkflowOperations failed, %s", err.Error())
				return
			}
			operationLock.Lock()
			operationList = append(operationList, operations.Operations...)
			operationLock.Unlock()
		}(i * defaultLimit)
	}
	con.Wait()

	blog.Infof("ListAllWorkflowOperations successful")
	return operationList, nil
}

// ListAllNetworkUnits 查询全部管控单元（分页并发拉取）
func (c *Client) ListAllNetworkUnits(ctx context.Context, req *NetworkUnitListRequest) ([]NetworkUnitItem, error) {
	countReq := *req
	countReq.Page = &NetworkUnitPage{
		Offset: 0,
		Limit:  int32(defaultLimit),
	}
	countReq.OnlyCount = true
	result, err := c.NetworkUnitList(ctx, &countReq)
	if err != nil {
		return nil, err
	}

	blog.Infof("ListAllNetworkUnits count %d", result.Total)
	var (
		unitList = make([]NetworkUnitItem, 0)
		unitLock = &sync.RWMutex{}
	)

	con := utils.NewRoutinePool(20)
	defer con.Close()

	page := int(result.Total-1)/defaultLimit + 1
	for i := 0; i < page; i++ {
		con.Add(1)
		go func(offset int) {
			defer con.Done()
			pageReq := *req
			pageReq.Page = &NetworkUnitPage{
				Offset: int32(offset),
				Limit:  int32(defaultLimit),
			}
			pageReq.OnlyCount = false
			units, err := c.NetworkUnitList(ctx, &pageReq)
			if err != nil {
				blog.Errorf("ListAllNetworkUnits failed, %s", err.Error())
				return
			}
			unitLock.Lock()
			unitList = append(unitList, units.Items...)
			unitLock.Unlock()
		}(i * defaultLimit)
	}
	con.Wait()

	blog.Infof("ListAllNetworkUnits successful")
	return unitList, nil
}
