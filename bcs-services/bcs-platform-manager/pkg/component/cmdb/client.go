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

// Package cmdb xxx
package cmdb

import (
	"encoding/json"
	"fmt"
	"time"

	bkcmdbkube "configcenter/src/kube/types" // nolint
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/parnurzeal/gorequest"
)

var (
	// cmdbClient global cmdb client
	cmdbClient     *Client
	defaultTimeOut = time.Second * 60
)

// Options for cmdb client
type Options struct {
	AppCode    string
	AppSecret  string
	BKUserName string
	Server     string
	Debug      bool
}

// Client for cc
type Client struct {
	config   *Options
	userAuth string
}

// AuthInfo auth user
type AuthInfo struct {
	BkAppCode   string `json:"bk_app_code"`
	BkAppSecret string `json:"bk_app_secret"`
	BkUserName  string `json:"bk_username"`
}

// SetCmdbClient set cmdb client
func SetCmdbClient(options Options) {
	// init client
	cmdbClient = NewCmdbClient(options)
}

// GetCmdbClient get cmdb client
func GetCmdbClient() *Client {
	return cmdbClient
}

// NewCmdbClient create cmdb client
func NewCmdbClient(options Options) *Client {
	c := &Client{
		config: &Options{
			AppCode:    options.AppCode,
			AppSecret:  options.AppSecret,
			BKUserName: options.BKUserName,
			Server:     options.Server,
			Debug:      options.Debug,
		},
	}

	auth, err := c.generateGateWayAuth()
	if err != nil {
		return nil
	}
	c.userAuth = auth
	return c
}

func (c *Client) generateGateWayAuth() (string, error) {
	auth := &AuthInfo{
		BkAppCode:   c.config.AppCode,
		BkAppSecret: c.config.AppSecret,
		BkUserName:  c.config.BKUserName,
	}

	userAuth, err := json.Marshal(auth)
	if err != nil {
		return "", err
	}

	return string(userAuth), nil
}

// GetBcsPod get pod
func (c *Client) GetBcsPod(req *GetBcsPodReq) (*[]bkcmdbkube.Pod, error) {
	reqURL := fmt.Sprintf("%s/api/v3/findmany/kube/pod", c.config.Server)
	respData := &GetBcsPodResp{}

	resp, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Post(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", c.userAuth).
		SetDebug(c.config.Debug).
		Send(req).
		Retry(3, 3*time.Second, 429).
		EndStruct(&respData)
	if len(errs) > 0 {
		blog.Errorf("call api list_pod failed: %v", errs[0])
		return nil, errs[0]
	}

	if !respData.Result {
		blog.Errorf("call api list_pod failed: %v, rid: %s",
			respData.Message, resp.Header.Get("X-Request-Id"))
		return nil, fmt.Errorf(respData.Message)
	}

	blog.Infof("call api list_pod with url(%s) (%v) successfully, X-Request-Id: %s",
		reqURL, req, resp.Header.Get("X-Request-Id"))

	return respData.Data.Info, nil
}

// DeleteBcsPod delete pod
func (c *Client) DeleteBcsPod(req *DeleteBcsPodReq) error {
	reqURL := fmt.Sprintf("%s/api/v3/deletemany/kube/pod", c.config.Server)
	respData := &DeleteBcsPodResp{}

	resp, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Delete(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", c.userAuth).
		SetDebug(c.config.Debug).
		Send(req).
		Retry(3, 3*time.Second, 429).
		EndStruct(&respData)
	if len(errs) > 0 {
		blog.Errorf("call api batch_delete_kube_pod failed: %v", errs[0])
		return errs[0]
	}

	if !respData.Result {
		blog.Errorf("call api batch_delete_kube_pod failed: %v, rid: %s",
			respData.Message, resp.Header.Get("X-Request-Id"))
		return fmt.Errorf(respData.Message)
	}

	blog.Infof("call api batch_delete_kube_pod with url(%s) (%v) successfully, X-Request-Id: %s",
		reqURL, req, resp.Header.Get("X-Request-Id"))

	return nil
}

// GetBcsWorkload get workload
func (c *Client) GetBcsWorkload(req *GetBcsWorkloadReq) (*[]interface{}, error) {
	reqURL := fmt.Sprintf("%s/api/v3/findmany/kube/workload/%s", c.config.Server, req.Kind)
	respData := &GetBcsWorkloadResp{}

	resp, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Post(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", c.userAuth).
		SetDebug(c.config.Debug).
		Send(req).
		Retry(3, 3*time.Second, 429).
		EndStruct(&respData)
	if len(errs) > 0 {
		blog.Errorf("call api list_workload failed: %v, rid: %s", errs[0], respData.RequestID)
		return nil, errs[0]
	}

	if !respData.Result {
		blog.Errorf("call api list_workload failed: %v, rid: %s",
			respData.Message, resp.Header.Get("X-Request-Id"))
		return nil, fmt.Errorf(respData.Message)
	}

	blog.Infof("call api list_workload with url(%s) (%v) successfully, X-Request-Id: %s",
		reqURL, req, resp.Header.Get("X-Request-Id"))

	return &respData.Data.Info, nil
}

// DeleteBcsWorkload delete workload
func (c *Client) DeleteBcsWorkload(req *DeleteBcsWorkloadReq) error {
	reqURL := fmt.Sprintf("%s/api/v3/deletemany/kube/workload/%s", c.config.Server, *req.Kind)
	respData := &DeleteBcsWorkloadResp{}

	resp, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Delete(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", c.userAuth).
		SetDebug(c.config.Debug).
		Send(req).
		Retry(3, 3*time.Second, 429).
		EndStruct(&respData)
	if len(errs) > 0 {
		blog.Errorf("call api batch_delete_workload failed: %v", errs[0])
		return errs[0]
	}

	if !respData.Result {
		blog.Errorf(
			"call api batch_delete_workload failed: %v, rid: %s, request: bkbizid: %d, kind: %s, ids: %v",
			respData.Message, resp.Header.Get("X-Request-Id"), *req.BKBizID, *req.Kind, *req.IDs)
		return fmt.Errorf(respData.Message)
	}

	blog.Infof("call api batch_delete_workload with url(%s) (%v) successfully, X-Request-Id: %s",
		reqURL, req, resp.Header.Get("X-Request-Id"))

	return nil
}

// GetBcsNamespace get namespace
func (c *Client) GetBcsNamespace(req *GetBcsNamespaceReq) (*[]bkcmdbkube.Namespace, error) {
	reqURL := fmt.Sprintf("%s/api/v3/findmany/kube/namespace", c.config.Server)
	respData := &GetBcsNamespaceResp{}

	resp, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Post(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", c.userAuth).
		SetDebug(c.config.Debug).
		Send(req).
		Retry(3, 3*time.Second, 429).
		EndStruct(&respData)
	if len(errs) > 0 {
		blog.Errorf("call api list_namespace failed: %v", errs[0])
		return nil, errs[0]
	}

	if !respData.Result {
		blog.Errorf("call api list_namespace failed: %v, rid: %s",
			respData.Message, resp.Header.Get("X-Request-Id"))
		return nil, fmt.Errorf(respData.Message)
	}

	blog.Infof("call api list_namespace with url(%s) (%v) successfully, X-Request-Id: %s",
		reqURL, req, resp.Header.Get("X-Request-Id"))

	return respData.Data.Info, nil
}

// DeleteBcsNamespace delete namespace
func (c *Client) DeleteBcsNamespace(req *DeleteBcsNamespaceReq) error {
	reqURL := fmt.Sprintf("%s/api/v3/deletemany/kube/namespace", c.config.Server)
	respData := &DeleteBcsNamespaceResp{}

	resp, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Delete(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", c.userAuth).
		SetDebug(c.config.Debug).
		Send(req).
		Retry(3, 3*time.Second, 429).
		EndStruct(&respData)
	if len(errs) > 0 {
		blog.Errorf("call api batch_delete_namespace failed: %v", errs[0])
		return errs[0]
	}

	if !respData.Result {
		blog.Errorf("call api batch_delete_namespace failed: %v, rid: %s",
			respData.Message, resp.Header.Get("X-Request-Id"))
		return fmt.Errorf(respData.Message)
	}

	blog.Infof("call api batch_delete_namespace with url(%s) (%v) successfully, X-Request-Id: %s",
		reqURL, req, resp.Header.Get("X-Request-Id"))

	return nil
}

// GetBcsNode get node
func (c *Client) GetBcsNode(req *GetBcsNodeReq) (*[]bkcmdbkube.Node, error) {
	reqURL := fmt.Sprintf("%s/api/v3/findmany/kube/node", c.config.Server)
	respData := &GetBcsNodeResp{}

	resp, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Post(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", c.userAuth).
		SetDebug(c.config.Debug).
		Send(req).
		Retry(3, 3*time.Second, 429).
		EndStruct(&respData)
	if len(errs) > 0 {
		blog.Errorf("call api list_kube_node failed: %v", errs[0])
		return nil, errs[0]
	}

	if !respData.Result {
		blog.Errorf("call api list_kube_node failed: %v, rid: %s",
			respData.Message, resp.Header.Get("X-Request-Id"))
		return nil, fmt.Errorf(respData.Message)
	}

	blog.Infof("call api list_kube_node with url(%s) (%v) successfully, X-Request-Id: %s",
		reqURL, req, resp.Header.Get("X-Request-Id"))

	return respData.Data.Info, nil
}

// DeleteBcsNode delete node
func (c *Client) DeleteBcsNode(req *DeleteBcsNodeReq) error {
	reqURL := fmt.Sprintf("%s/api/v3/deletemany/kube/node", c.config.Server)
	respData := &DeleteBcsNodeResp{}

	resp, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Delete(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", c.userAuth).
		SetDebug(c.config.Debug).
		Send(req).
		Retry(3, 3*time.Second, 429).
		EndStruct(&respData)
	if len(errs) > 0 {
		blog.Errorf("call api batch_delete_kube_node failed: %v", errs[0])
		return errs[0]
	}

	if !respData.Result {
		blog.Errorf("call api batch_delete_kube_node failed: %v, rid: %s",
			respData.Message, resp.Header.Get("X-Request-Id"))
		return fmt.Errorf(respData.Message)
	}

	blog.Infof("call api batch_delete_kube_node with url(%s) (%v) successfully, X-Request-Id: %s",
		reqURL, req, resp.Header.Get("X-Request-Id"))

	return nil
}

// GetBcsCluster get cluster
func (c *Client) GetBcsCluster(req *GetBcsClusterReq) (*[]bkcmdbkube.Cluster, error) {
	// 如果没有通过数据库查询，则通过API调用获取集群信息
	reqURL := fmt.Sprintf("%s/api/v3/findmany/kube/cluster", c.config.Server)
	respData := &GetBcsClusterResp{}
	// 使用gorequest库发送POST请求，并处理响应
	resp, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Post(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", c.userAuth).
		SetDebug(c.config.Debug).
		Send(req).
		Retry(3, 3*time.Second, 429).
		EndStruct(&respData)
	// 检查是否有错误发生
	if len(errs) > 0 {
		blog.Errorf("call api list_kube_cluster failed: %v", errs[0])
		return nil, errs[0]
	}

	// 检查API响应是否成功
	if !respData.Result {
		blog.Errorf("call api list_kube_cluster failed: %v, rid: %s",
			respData.Message, resp.Header.Get("X-Request-Id"))
		return nil, fmt.Errorf(respData.Message)
	}

	blog.Infof("call api list_kube_cluster with url(%s) (%v) successfully, X-Request-Id: %s",
		reqURL, req, resp.Header.Get("X-Request-Id"))

	return &respData.Data.Info, nil
}

// DeleteBcsCluster delete bcs cluster
func (c *Client) DeleteBcsCluster(req *DeleteBcsClusterReq) error {
	// 构造请求的 URL
	reqURL := fmt.Sprintf("%s/api/v3/delete/kube/cluster", c.config.Server)
	// 初始化响应数据结构
	respData := &DeleteBcsClusterResp{}

	// 使用 gorequest 库发送 HTTP DELETE 请求
	// 设置请求超时时间、内容类型、接受类型、授权头等信息
	// 发送请求体并尝试重试，最多重试3次，每次间隔3秒，如果遇到429状态码也会重试
	resp, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Delete(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", c.userAuth).
		SetDebug(c.config.Debug).
		Send(req).
		Retry(3, 3*time.Second, 429).
		EndStruct(&respData)

	// 如果请求过程中出现错误，则记录错误并返回
	if len(errs) > 0 {
		blog.Errorf("call api batch_delete_kube_cluster failed: %v", errs[0])
		return errs[0]
	}

	// 如果响应结果指示操作失败，则记录错误信息并返回错误
	if !respData.Result {
		blog.Errorf("call api batch_delete_kube_cluster failed: %v, rid: %s",
			respData.Message, resp.Header.Get("X-Request-Id"))
		return fmt.Errorf(respData.Message)
	}

	// 如果操作成功，则记录成功的日志信息
	blog.Infof("call api batch_delete_kube_cluster with url(%s) (%v) successfully, X-Request-Id: %s",
		reqURL, req, resp.Header.Get("X-Request-Id"))

	return nil
}

// GetBusiness get business
func (c *Client) GetBusiness() (*[]GetBusinessRespDataInfo, error) {
	// 构造请求的 URL
	reqURL := fmt.Sprintf("%s/api/c/compapi/v2/cc/search_business", c.config.Server)
	// 初始化响应数据结构
	respData := &GetBusinessResp{}

	// 使用 gorequest 库发送 HTTP DELETE 请求
	// 设置请求超时时间、内容类型、接受类型、授权头等信息
	// 发送请求体并尝试重试，最多重试3次，每次间隔3秒，如果遇到429状态码也会重试
	resp, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Post(reqURL).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", c.userAuth).
		SetDebug(c.config.Debug).
		Retry(3, 3*time.Second, 429).
		EndStruct(&respData)

	// 如果请求过程中出现错误，则记录错误并返回
	if len(errs) > 0 {
		blog.Errorf("call api get business failed: %v", errs[0])
		return nil, errs[0]
	}

	// 如果响应结果指示操作失败，则记录错误信息并返回错误
	if !respData.Result {
		blog.Errorf("call api get business failed: %v, rid: %s",
			respData.Message, resp.Header.Get("X-Request-Id"))
		return nil, fmt.Errorf(respData.Message)
	}

	// 如果操作成功，则记录成功的日志信息
	blog.Infof("call api get business with url(%s) successfully, X-Request-Id: %s",
		reqURL, resp.Header.Get("X-Request-Id"))

	return respData.Data.Info, nil
}
