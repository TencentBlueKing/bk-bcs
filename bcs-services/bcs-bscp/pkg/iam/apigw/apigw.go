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

// Package apigw document sync gateway
package apigw

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/cc"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/rest/client"
)

var (
	endpoint                    = "http://bkapi.sit.bktencent.com/api/bk-apigateway/prod/api/v1/apis/"
	syncApi                     = "%s/%s/sync/"
	syncStage                   = "%s/%s/stages/sync/"
	syncResources               = "%s/%s/resources/sync/"
	importResourceDocsBySwagger = "%s/%s/resource-docs/import/by-swagger/"
	createResourceVersion       = "%s/%s/resource_versions/"
	getLatestResourceVersion    = "%s/%s/resource_versions/latest/"
	release                     = "%s/%s/resource_versions/release/"
	applyPermissions            = "%s/%s/permissions/apply/"
	getApigwPublicKey           = "%s/%s/public_key/"
)

// ApiGw document sync gateway interface
type ApiGw interface {
	// SyncApi 同步网关，如果网关不存在，创建网关，如果网关已存在，更新网关
	SyncApi(gwName string, req *SyncApiReq) (*SyncApiResp, error)
	// SyncStage 同步网关环境，如果环境不存在，创建环境，如果已存在，则更新
	SyncStage(gwName string, req *SyncStageReq) (*SyncStageResp, error)
	// SyncResources 同步资源
	SyncResources(gwName string, req *SyncResourcesReq) (*SyncResourcesResp, error)
	// ImportResourceDocsBySwagger 根据 swagger 描述文件，导入资源文档
	ImportResourceDocsBySwagger(gwName string, req *ImportResourceDocsBySwaggerReq) (
		*ImportResourceDocsBySwaggerResp, error)
	// CreateResourceVersion 创建资源版本
	CreateResourceVersion(gwName string, req *CreateResourceVersionReq) (*CreateResourceVersionResp, error)
	// GetLatestResourceVersion 获取网关最新的资源版本
	GetLatestResourceVersion(gwName string) (*GetLatestResourceVersionResp, error)
	// Release 发布版本
	Release(gwName string, req *ReleaseReq) (*ReleaseResp, error)
	// ApplyPermissions 申请网关API访问权限
	ApplyPermissions(gwName string, req *ApplyPermissionsReq) (*ApplyPermissionsResp, error)
	// GetApigwPublicKey 获取网关公钥
	GetApigwPublicKey(gwName string) (*GetApigwPublicKeyResp, error)
}

// NewApiGw 初始化网关
func NewApiGw(opt cc.ApiServerSetting) (ApiGw, error) {

	c, err := client.NewClient(nil)
	if err != nil {
		return nil, err
	}
	return &apiGw{
		client: c,
		opt:    opt,
	}, nil
}

type apiGw struct {
	client *http.Client
	opt    cc.ApiServerSetting
}

// SyncApi 同步网关，如果网关不存在，创建网关，如果网关已存在，更新网关
func (a *apiGw) SyncApi(gwName string, req *SyncApiReq) (*SyncApiResp, error) {
	url := fmt.Sprintf(syncApi, endpoint, gwName)

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("sync api serialization JSON failed: %s", err.Error())
	}

	request, err := a.newRequest("POST", url, jsonData)
	if err != nil {
		return nil, err
	}

	resp, err := a.client.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	result := new(SyncApiResp)

	// 将响应体解析为结构体
	if err = json.NewDecoder(resp.Body).Decode(result); err != nil {
		return nil, err
	}
	return result, nil
}

// SyncStage 同步网关环境，如果环境不存在，创建环境，如果已存在，则更新
func (a *apiGw) SyncStage(gwName string, req *SyncStageReq) (*SyncStageResp, error) {
	url := fmt.Sprintf(syncStage, endpoint, gwName)

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("sync stage serialization JSON failed: %s", err.Error())
	}

	request, err := a.newRequest("POST", url, jsonData)
	if err != nil {
		return nil, err
	}

	resp, err := a.client.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	result := new(SyncStageResp)

	// 将响应体解析为结构体
	if err = json.NewDecoder(resp.Body).Decode(result); err != nil {
		return nil, err
	}

	return result, nil
}

// SyncResources 同步资源
func (a *apiGw) SyncResources(gwName string, req *SyncResourcesReq) (*SyncResourcesResp, error) {

	url := fmt.Sprintf(syncResources, endpoint, gwName)

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("sync resources serialization JSON failed: %s", err.Error())
	}

	request, err := a.newRequest("POST", url, jsonData)
	if err != nil {
		return nil, err
	}

	resp, err := a.client.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	result := new(SyncResourcesResp)

	// 将响应体解析为结构体
	if err = json.NewDecoder(resp.Body).Decode(result); err != nil {
		return nil, err
	}

	return result, nil
}

// ImportResourceDocsBySwagger 根据 swagger 描述文件，导入资源文档
func (a *apiGw) ImportResourceDocsBySwagger(gwName string, req *ImportResourceDocsBySwaggerReq) (
	*ImportResourceDocsBySwaggerResp, error) {
	url := fmt.Sprintf(importResourceDocsBySwagger, endpoint, gwName)

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("import resource docs by swagger serialization JSON failed: %s", err.Error())
	}

	request, err := a.newRequest("POST", url, jsonData)
	if err != nil {
		return nil, err
	}

	resp, err := a.client.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	result := new(ImportResourceDocsBySwaggerResp)

	// 将响应体解析为结构体
	if err = json.NewDecoder(resp.Body).Decode(result); err != nil {
		return nil, err
	}

	return result, nil
}

// CreateResourceVersion implements ApiGw.
func (a *apiGw) CreateResourceVersion(gwName string, req *CreateResourceVersionReq) (
	*CreateResourceVersionResp, error) {

	url := fmt.Sprintf(createResourceVersion, endpoint, gwName)

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("create resource version serialization JSON failed: %s", err.Error())
	}

	request, err := a.newRequest("POST", url, jsonData)
	if err != nil {
		return nil, err
	}

	resp, err := a.client.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	result := new(CreateResourceVersionResp)

	// 将响应体解析为结构体
	if err = json.NewDecoder(resp.Body).Decode(result); err != nil {
		return nil, err
	}

	return result, nil
}

// GetLatestResourceVersion implements ApiGw.
func (a *apiGw) GetLatestResourceVersion(gwName string) (*GetLatestResourceVersionResp, error) {

	url := fmt.Sprintf(getLatestResourceVersion, endpoint, gwName)

	request, err := a.newRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := a.client.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	result := new(GetLatestResourceVersionResp)

	// 将响应体解析为结构体
	if err = json.NewDecoder(resp.Body).Decode(result); err != nil {
		return nil, err
	}

	return result, nil
}

// Release 发布版本
func (a *apiGw) Release(gwName string, req *ReleaseReq) (*ReleaseResp, error) {

	url := fmt.Sprintf(release, endpoint, gwName)

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("release serialization JSON failed: %s", err.Error())
	}

	request, err := a.newRequest("POST", url, jsonData)
	if err != nil {
		return nil, err
	}

	resp, err := a.client.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	result := new(ReleaseResp)

	// 将响应体解析为结构体
	if err = json.NewDecoder(resp.Body).Decode(result); err != nil {
		return nil, err
	}

	return result, nil
}

// ApplyPermissions implements ApiGw.
func (a *apiGw) ApplyPermissions(gwName string, req *ApplyPermissionsReq) (*ApplyPermissionsResp, error) {
	url := fmt.Sprintf(applyPermissions, endpoint, gwName)

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("apply permissions serialization JSON failed: %s", err.Error())
	}

	request, err := a.newRequest("POST", url, jsonData)
	if err != nil {
		return nil, err
	}

	resp, err := a.client.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	result := new(ApplyPermissionsResp)

	// 将响应体解析为结构体
	if err = json.NewDecoder(resp.Body).Decode(result); err != nil {
		return nil, err
	}

	return result, nil
}

// GetApigwPublicKey implements ApiGw.
func (a *apiGw) GetApigwPublicKey(gwName string) (*GetApigwPublicKeyResp, error) {
	url := fmt.Sprintf(getApigwPublicKey, endpoint, gwName)

	request, err := a.newRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := a.client.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	result := new(GetApigwPublicKeyResp)

	// 将响应体解析为结构体
	if err = json.NewDecoder(resp.Body).Decode(result); err != nil {
		return nil, err
	}

	return result, nil
}

func (a *apiGw) newRequest(method, url string, body []byte) (*http.Request, error) {

	var req *http.Request
	var err error

	if len(body) > 0 {
		req, err = http.NewRequest(method, url, bytes.NewBuffer(body))
	} else {
		req, err = http.NewRequest(method, url, nil)
	}

	if err != nil {
		return nil, err
	}

	// 设置请求头
	req.Header.Set("X-Bkapi-Authorization", fmt.Sprintf(`{"bk_app_code": "%s", "bk_app_secret": "%s"}`,
		a.opt.Esb.AppCode, a.opt.Esb.AppSecret))
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// ImportResourceDocsBySwaggerReq 输入参数
type ImportResourceDocsBySwaggerReq struct {
	// Language 文档语言，可选值：zh 表示中文，en 表示英文
	Language string `json:"language"`
	// Swagger 描述文件的内容
	Swagger string `json:"swagger"`
}

// ImportResourceDocsBySwaggerResp 输出参数
type ImportResourceDocsBySwaggerResp struct {
	Code    int         `json:"code"`
	Data    interface{} `json:"data"`
	Result  bool        `json:"result"`
	Message string      `json:"message"`
}

// SyncApiReq 输入参数
type SyncApiReq struct {
	// Description 网关描述
	Description string `json:"description"`
	// Maintainers 网关管理员
	Maintainers []string `json:"maintainers"`
	// IsPublic 网关是否公开
	IsPublic bool `json:"is_public"`
}

// SyncApiResp 输出参数
type SyncApiResp struct {
	Data    SyncData `json:"data"`
	Code    int      `json:"code"`
	Result  bool     `json:"result"`
	Message string   `json:"message"`
}

// SyncData xxx
type SyncData struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// SyncStageReq 输入参数
type SyncStageReq struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Vars        map[string]string `json:"vars"`
	ProxyHttp   ProxyHttp         `json:"proxy_http"`
}

// ProxyHttp xxx
type ProxyHttp struct {
	Timeout          int       `json:"timeout"`
	Upstreams        Upstreams `json:"upstreams"`
	TransformHeaders `json:"transform_headers"`
}

// Upstreams xxx
type Upstreams struct {
	Loadbalance string `json:"loadbalance"`
	Hosts       []Host `json:"hosts"`
}

// Host xx
type Host struct {
	Host   string `json:"host"`
	Weight int    `json:"weight"`
}

// TransformHeaders xxx
type TransformHeaders struct {
	Set    map[string]string `json:"set"`
	Delete []string          `json:"delete"`
}

// SyncResourcesReq 输入参数
type SyncResourcesReq struct {
	Content string `json:"content"`
	Delete  bool   `json:"delete"`
}

// SyncStageResp 输出参数
type SyncStageResp struct {
	Data    SyncData `json:"data"`
	Code    int      `json:"code"`
	Result  bool     `json:"result"`
	Message string   `json:"message"`
}

// SyncResourcesResp 输出参数
type SyncResourcesResp struct {
	Code    int               `json:"code"`
	Message string            `json:"message"`
	Data    SyncResourcesData `json:"data"`
}

// SyncResourcesData xxx
type SyncResourcesData struct {
	Added   []Added   `json:"added"`
	Updated []Updated `json:"updated"`
	Deleted []Deleted `json:"deleted"`
}

// Added xxx
type Added struct {
	ID int `json:"id"`
}

// Updated xxx
type Updated struct {
	ID int `json:"id"`
}

// Deleted xxx
type Deleted struct {
	ID int `json:"id"`
}

// ReleaseReq 输入参数
type ReleaseReq struct {
	Version    string   `json:"version"`
	StageNames []string `json:"stage_names"`
	Comment    string   `json:"comment"`
}

// ReleaseResp 输出参数
type ReleaseResp struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    ReleaseData `json:"data"`
}

// ReleaseData xxx
type ReleaseData struct {
	Version    string   `json:"version"`
	StageNames []string `json:"stage_names"`
}

// CreateResourceVersionReq 输入参数
type CreateResourceVersionReq struct {
	Version string `json:"version"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
}

// CreateResourceVersionResp 输出参数
type CreateResourceVersionResp struct {
	Code    int                       `json:"code"`
	Message string                    `json:"message"`
	Data    CreateResourceVersionData `json:"data"`
}

// CreateResourceVersionData xxx
type CreateResourceVersionData struct {
	ID      int    `json:"id"`
	Version string `json:"version"`
	Title   string `json:"title"`
}

// ApplyPermissionsReq 输入参数
type ApplyPermissionsReq struct {
	TargetAppCode  string `json:"target_app_code"`
	ExpireDays     int    `json:"expire_days"`
	GrantDimension string `json:"grant_dimension"`
	Reason         string `json:"reason"`
}

// ApplyPermissionsResp 输出参数
type ApplyPermissionsResp struct {
	Code    int                 `json:"code"`
	Message string              `json:"message"`
	Data    ApplyPermissionData `json:"data"`
}

// ApplyPermissionData xxx
type ApplyPermissionData struct {
	RecordID int `json:"record_id"`
}

// GetApigwPublicKeyResp 输出参数
type GetApigwPublicKeyResp struct {
	Code    int                   `json:"code"`
	Message string                `json:"message"`
	Data    GetApigwPublicKeyData `json:"data"`
}

// GetApigwPublicKeyData xxx
type GetApigwPublicKeyData struct {
	Issuer    string `json:"issuer"`
	PublicKey string `json:"public_key"`
}

// GetLatestResourceVersionResp 输出参数
type GetLatestResourceVersionResp struct {
	Code    int                          `json:"code"`
	Message string                       `json:"message"`
	Data    GetLatestResourceVersionData `json:"data"`
}

// GetLatestResourceVersionData xxx
type GetLatestResourceVersionData struct {
	Version string `json:"version"`
}
