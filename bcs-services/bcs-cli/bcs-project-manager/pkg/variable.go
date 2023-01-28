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

package pkg

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/go-querystring/query"
	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/proto/bcsproject"
)

const (
	listVariableDefinitionsUrl = "/bcsproject/v1/projects/%s/variables"
	variableUrl                = "/bcsproject/v1/projects/%s/variables"
	updateVariableUrl          = "/bcsproject/v1/projects/%s/variables/%s"
	renderVariablesUrl         = "/bcsproject/v1/projects/%s/clusters/%s/namespaces/%s/variables/render"
)

type (

	// ListVariableDefinitionsRequest 列表变量定义请求
	ListVariableDefinitionsRequest struct {
		Scope     string `url:"scope,omitempty"`
		SearchKey string `url:"searchKey,omitempty"`
		Offset    int64  `url:"offset,omitempty"`
		Limit     int64  `url:"limit,omitempty"`
		All       bool   `url:"all,omitempty"`
	}

	// CreateVariableRequest 创建变量请求
	CreateVariableRequest struct {
		ProjectCode string `json:"projectCode"`
		Name        string `json:"name"`
		Key         string `json:"key"`
		Scope       string `json:"scope"`
		Default     string `json:"default"`
		Desc        string `json:"desc"`
	}

	// UpdateVariableRequest 更新变量请求
	UpdateVariableRequest struct {
		ProjectCode string `json:"projectCode"`
		VariableID  string `json:"variableID"`
		Name        string `json:"name"`
		Key         string `json:"key"`
		Scope       string `json:"scope"`
		Default     string `json:"default"`
		Desc        string `json:"desc"`
	}

	// DeleteVariableDefinitionsRequest 删除变量定义请求
	DeleteVariableDefinitionsRequest struct {
		IdList string `url:"idList,omitempty"`
	}

	// RenderVariablesRequest 渲染变量请求
	RenderVariablesRequest struct {
		KeyList string `url:"keyList,omitempty"`
	}
)

// ListVariableDefinitions Get a list of project variable definitions based on a condition
func (p *ProjectManagerClient) ListVariableDefinitions(in *ListVariableDefinitionsRequest, projectCode string) (*bcsproject.ListVariableDefinitionsResponse, error) {
	v, err := query.Values(in)
	if err != nil {
		return nil, fmt.Errorf("slice and Array values default to encoding as multiple URL values failed: %v", err)
	}
	bs, err := p.do(fmt.Sprintf(listVariableDefinitionsUrl, projectCode), http.MethodGet, v, nil)
	if err != nil {
		return nil, err
	}
	resp := new(bcsproject.ListVariableDefinitionsResponse)
	if err := json.Unmarshal(bs, resp); err != nil {
		return nil, errors.Wrapf(err, "list variable unmarshal failed with response '%s'", string(bs))
	}
	if resp != nil && resp.Code != 0 {
		return nil, fmt.Errorf("list variable response code not 0 but %d: %s", resp.Code, resp.Message)
	}
	return resp, nil
}

// CreateVariable Create project variables
func (p *ProjectManagerClient) CreateVariable(in *CreateVariableRequest) (*bcsproject.CreateVariableResponse, error) {
	bs, err := p.do(fmt.Sprintf(variableUrl, in.ProjectCode), http.MethodPost, nil, in)
	if err != nil {
		return nil, fmt.Errorf("create variable failed: %v", err)
	}
	resp := new(bcsproject.CreateVariableResponse)
	if err := json.Unmarshal(bs, resp); err != nil {
		return nil, errors.Wrapf(err, "create variable unmarshal failed with response '%s'", string(bs))
	}
	if resp != nil && resp.Code != 0 {
		return nil, fmt.Errorf("create variable response code not 0 but %d: %s", resp.Code, resp.Message)
	}
	return resp, nil
}

// UpdateVariable Edit project variables
func (p *ProjectManagerClient) UpdateVariable(in *UpdateVariableRequest) (*bcsproject.CreateVariableResponse, error) {
	bs, err := p.do(fmt.Sprintf(updateVariableUrl, in.ProjectCode, in.VariableID), http.MethodPut, nil, in)
	if err != nil {
		return nil, fmt.Errorf("edit variable failed: %v", err)
	}
	resp := new(bcsproject.CreateVariableResponse)
	if err := json.Unmarshal(bs, resp); err != nil {
		return nil, errors.Wrapf(err, "edit variable unmarshal failed with response '%s'", string(bs))
	}
	if resp != nil && resp.Code != 0 {
		return nil, fmt.Errorf("edit variable response code not 0 but %d: %s", resp.Code, resp.Message)
	}
	return resp, nil
}

// DeleteVariableDefinitions Delete project variable
func (p *ProjectManagerClient) DeleteVariableDefinitions(in *DeleteVariableDefinitionsRequest, projectCode string) (*bcsproject.DeleteVariableDefinitionsResponse, error) {
	v, err := query.Values(in)
	if err != nil {
		return nil, fmt.Errorf("slice and Array values default to encoding as multiple URL values failed: %v", err)
	}
	bs, err := p.do(fmt.Sprintf(variableUrl, projectCode), http.MethodDelete, v, nil)
	if err != nil {
		return nil, fmt.Errorf("delete variable failed: %v", err)
	}
	resp := new(bcsproject.DeleteVariableDefinitionsResponse)
	if err := json.Unmarshal(bs, resp); err != nil {
		return nil, errors.Wrapf(err, "delete variable unmarshal failed with response '%s'", string(bs))
	}
	if resp != nil && resp.Code != 0 {
		return nil, fmt.Errorf("delete variable response code not 0 but %d: %s", resp.Code, resp.Message)
	}
	return resp, nil
}

// RenderVariables render variable's value under a specific cluster, namespace
func (p *ProjectManagerClient) RenderVariables(in *RenderVariablesRequest, projectCode, clusterID, namespace string) (*bcsproject.RenderVariablesResponse, error) {
	v, err := query.Values(in)
	if err != nil {
		return nil, fmt.Errorf("slice and Array values default to encoding as multiple URL values failed: %v", err)
	}
	bs, err := p.do(fmt.Sprintf(renderVariablesUrl, projectCode, clusterID, namespace), http.MethodGet, v, nil)
	if err != nil {
		return nil, fmt.Errorf("render variable failed: %v", err)
	}
	resp := new(bcsproject.RenderVariablesResponse)
	if err := json.Unmarshal(bs, resp); err != nil {
		return nil, errors.Wrapf(err, "render variable unmarshal failed with response '%s'", string(bs))
	}
	if resp != nil && resp.Code != 0 {
		return nil, fmt.Errorf("render variable response code not 0 but %d: %s", resp.Code, resp.Message)
	}
	return resp, nil
}
