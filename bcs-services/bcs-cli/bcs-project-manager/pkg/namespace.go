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

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/proto/bcsproject"
	"github.com/pkg/errors"
)

const (
	listNamespacesUrl           = "/bcsproject/v1/projects/%s/clusters/%s/namespaces"
	namespaceUrl                = "/bcsproject/v1/projects/%s/clusters/%s/namespaces/%s"
	createNamespaceCallbackUrl  = "/bcsproject/v1/projects/%s/clusters/%s/namespaces/%s/callback/create"
	deleteNamespaceCallbackUrl  = "/bcsproject/v1/projects/%s/clusters/%s/namespaces/%s/callback/delete"
	updateNamespaceCallbackUrl  = "/bcsproject/v1/projects/%s/clusters/%s/namespaces/%s/callback/update"
	updateNamespaceVariablesUrl = "/bcsproject/v1/projects/%s/clusters/%s/namespaces/%s/variables"
)

type (

	// ListNamespacesRequest 命名列表参数
	ListNamespacesRequest struct {
		ProjectCode string `json:"projectCode"`
		ClusterID   string `json:"clusterID"`
	}

	// CreateNamespaceRequest 创建命名参数
	CreateNamespaceRequest struct {
		ProjectCode string `json:"projectCode"`
		ClusterID   string `json:"clusterID"`
		Name        string `json:"name"`
		Quota       struct {
			CPURequests    string `json:"cpuRequests"`
			MemoryRequests string `json:"memoryRequests"`
			CPULimits      string `json:"cpuLimits"`
			MemoryLimits   string `json:"memoryLimits"`
		} `json:"quota,omitempty"`
		Labels []struct {
			Key   string `json:"key,omitempty"`
			Value string `json:"value,omitempty"`
		} `json:"labels,omitempty"`
		Annotations []struct {
			Key   string `json:"key,omitempty"`
			Value string `json:"value,omitempty"`
		} `json:"annotations,omitempty"`
		Variables []struct {
			ID          string `json:"id,omitempty"`
			Key         string `json:"key,omitempty"`
			Name        string `json:"name,omitempty"`
			ClusterID   string `json:"clusterID,omitempty"`
			ClusterName string `json:"clusterName,omitempty"`
			Namespace   string `json:"namespace,omitempty"`
			Value       string `json:"value,omitempty"`
			Scope       string `json:"scope,omitempty"`
		} `json:"variables,omitempty"`
	}

	// UpdateNamespaceTemplate 更新命名空间时编辑器输出的内容
	UpdateNamespaceTemplate struct {
		UpdateNamespaceRequest
		Variable []Data `json:"variable"`
	}

	// UpdateNamespaceRequest 更新命名参数
	UpdateNamespaceRequest struct {
		ProjectCode string          `json:"projectCode"`
		ClusterID   string          `json:"clusterID"`
		Name        string          `json:"name"`
		Quota       Quota           `json:"quota"`
		Labels      []Labels        `json:"labels,omitempty"`
		Variables   []VariableValue `json:"variables"`
		Annotations []Annotations   `json:"annotations,omitempty"`
	}

	// VariableValue 变量值
	VariableValue struct {
		Id          string `json:"id,omitempty"`
		Key         string `json:"key,omitempty"`
		Name        string `json:"name,omitempty"`
		ClusterID   string `json:"clusterID,omitempty"`
		ClusterName string `json:"clusterName,omitempty"`
		Namespace   string `json:"namespace,omitempty"`
		Value       string `json:"value,omitempty"`
		Scope       string `json:"scope,omitempty"`
	}

	// Labels 标签
	Labels struct {
		Key   string `json:"key,omitempty"`
		Value string `json:"value,omitempty"`
	}
	// Annotations 注释
	Annotations struct {
		Key   string `json:"key,omitempty"`
		Value string `json:"value,omitempty"`
	}

	// Quota 配额
	Quota struct {
		CPURequests    string `json:"cpuRequests,omitempty"`
		MemoryRequests string `json:"memoryRequests,omitempty"`
		CPULimits      string `json:"cpuLimits,omitempty"`
		MemoryLimits   string `json:"memoryLimits,omitempty"`
	}

	// DeleteNamespaceRequest 删除命名参数
	DeleteNamespaceRequest struct {
		ProjectCode string `json:"projectCode"`
		ClusterID   string `json:"clusterID"`
		Name        string `json:"name"`
	}

	// NamespaceCallbackRequest 命名回调参数
	NamespaceCallbackRequest struct {
		ProjectCode    string `json:"projectCode"`
		ClusterID      string `json:"clusterID"`
		Name           string `json:"name"`
		Title          string `json:"title"`
		CurrentStatus  string `json:"currentStatus"`
		Sn             string `json:"sn"`
		TicketURL      string `json:"ticketUrl"`
		ApproveResult  bool   `json:"approveResult"`
		ApplyInCluster bool   `json:"applyInCluster"`
	}

	// UpdateNamespaceVariablesReq 更新命名变量参数
	UpdateNamespaceVariablesReq struct {
		ProjectCode string `json:"projectCode"`
		ClusterID   string `json:"clusterID"`
		Namespace   string `json:"namespace"`
		Data        []Data `json:"data"`
	}

	// Data 命名变量
	Data struct {
		ID          string `json:"id"`
		Key         string `json:"key"`
		Name        string `json:"name"`
		ClusterID   string `json:"clusterID"`
		ClusterName string `json:"clusterName"`
		Namespace   string `json:"namespace"`
		Value       string `json:"value"`
		Scope       string `json:"scope"`
	}

	// Variable 更新时做对比用
	Variable struct {
		ID          string `json:"id"`
		Key         string `json:"key"`
		Name        string `json:"name"`
		ClusterID   string `json:"clusterID"`
		ClusterName string `json:"clusterName"`
		Namespace   string `json:"namespace"`
		Scope       string `json:"scope"`
	}
)

// ListNamespaces Get all namespaces under the cluster
func (p *ProjectManagerClient) ListNamespaces(in *ListNamespacesRequest) (*bcsproject.ListNamespacesResponse, error) {
	bs, err := p.do(fmt.Sprintf(listNamespacesUrl, in.ProjectCode, in.ClusterID), http.MethodGet, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("list namespaces failed: %v", err)
	}
	resp := new(bcsproject.ListNamespacesResponse)
	if err := json.Unmarshal(bs, resp); err != nil {
		return nil, errors.Wrapf(err, "list namespaces unmarshal failed with response '%s'", string(bs))
	}
	if resp != nil && resp.Code != 0 {
		return nil, fmt.Errorf("list namespaces response code not 0 but %d: %s", resp.Code, resp.Message)
	}
	return resp, nil
}

// CreateNamespace Create namespace
func (p *ProjectManagerClient) CreateNamespace(in *CreateNamespaceRequest, projectCode, clusterID string) (*bcsproject.CreateNamespaceResponse, error) {
	bs, err := p.do(fmt.Sprintf(listNamespacesUrl, projectCode, clusterID), http.MethodPost, nil, in)
	if err != nil {
		return nil, fmt.Errorf("create namespaces failed: %v", err)
	}
	resp := new(bcsproject.CreateNamespaceResponse)
	if err := json.Unmarshal(bs, resp); err != nil {
		return nil, errors.Wrapf(err, "create namespaces unmarshal failed with response '%s'", string(bs))
	}
	if resp != nil && resp.Code != 0 {
		return nil, fmt.Errorf("create namespaces response code not 0 but %d: %s", resp.Code, resp.Message)
	}
	return resp, nil
}

// UpdateNamespace Update namespace
func (p *ProjectManagerClient) UpdateNamespace(in *UpdateNamespaceRequest, projectCode, clusterID, name string) (*bcsproject.UpdateNamespaceResponse, error) {
	bs, err := p.do(fmt.Sprintf(namespaceUrl, projectCode, clusterID, name), http.MethodPut, nil, in)
	if err != nil {
		return nil, fmt.Errorf("update namespaces failed: %v", err)
	}
	resp := new(bcsproject.UpdateNamespaceResponse)
	if err := json.Unmarshal(bs, resp); err != nil {
		return nil, errors.Wrapf(err, "update namespaces unmarshal failed with response '%s'", string(bs))
	}
	if resp != nil && resp.Code != 0 {
		return nil, fmt.Errorf("update namespaces response code not 0 but %d: %s", resp.Code, resp.Message)
	}
	return resp, nil
}

// DeleteNamespace Delete namespace
func (p *ProjectManagerClient) DeleteNamespace(in *DeleteNamespaceRequest) (*bcsproject.DeleteNamespaceResponse, error) {
	bs, err := p.do(fmt.Sprintf(namespaceUrl, in.ProjectCode, in.ClusterID, in.Name), http.MethodDelete, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("delete namespaces failed: %v", err)
	}
	resp := new(bcsproject.DeleteNamespaceResponse)
	if err := json.Unmarshal(bs, resp); err != nil {
		return nil, errors.Wrapf(err, "delete namespaces unmarshal failed with response '%s'", string(bs))
	}
	if resp != nil && resp.Code != 0 {
		return nil, fmt.Errorf("delete namespaces response code not 0 but %d: %s", resp.Code, resp.Message)
	}
	return resp, nil
}

// CreateNamespaceCallback create namespace ITSM callback
func (p *ProjectManagerClient) CreateNamespaceCallback(in *NamespaceCallbackRequest) (*bcsproject.NamespaceCallbackResponse, error) {
	bs, err := p.do(fmt.Sprintf(createNamespaceCallbackUrl, in.ProjectCode, in.ClusterID, in.Name), http.MethodPost, nil, in)
	if err != nil {
		return nil, fmt.Errorf("create namespaces itsm callback failed: %v", err)
	}
	resp := new(bcsproject.NamespaceCallbackResponse)
	if err := json.Unmarshal(bs, resp); err != nil {
		return nil, errors.Wrapf(err, "create namespaces itsm callback unmarshal failed with response '%s'", string(bs))
	}
	if resp != nil && resp.Code != 0 {
		return nil, fmt.Errorf("create namespaces itsm callback response code not 0 but %d: %s", resp.Code, resp.Message)
	}
	return resp, nil
}

// DeleteNamespaceCallback Delete namespace ITSM callback
func (p *ProjectManagerClient) DeleteNamespaceCallback(in *NamespaceCallbackRequest) (*bcsproject.NamespaceCallbackResponse, error) {
	bs, err := p.do(fmt.Sprintf(deleteNamespaceCallbackUrl, in.ProjectCode, in.ClusterID, in.Name), http.MethodPost, nil, in)
	if err != nil {
		return nil, fmt.Errorf("delete namespaces itsm callback failed: %v", err)
	}
	resp := new(bcsproject.NamespaceCallbackResponse)
	if err := json.Unmarshal(bs, resp); err != nil {
		return nil, errors.Wrapf(err, "delete namespaces itsm callback unmarshal failed with response '%s'", string(bs))
	}
	if resp != nil && resp.Code != 0 {
		return nil, fmt.Errorf("delete namespaces itsm callback response code not 0 but %d: %s", resp.Code, resp.Message)
	}
	return resp, nil
}

// UpdateNamespaceCallback Update namespace ITSM callback
func (p *ProjectManagerClient) UpdateNamespaceCallback(in *NamespaceCallbackRequest) (*bcsproject.NamespaceCallbackResponse, error) {
	bs, err := p.do(fmt.Sprintf(updateNamespaceCallbackUrl, in.ProjectCode, in.ClusterID, in.Name), http.MethodPost, nil, in)
	if err != nil {
		return nil, fmt.Errorf("update namespaces itsm callback failed: %v", err)
	}
	resp := new(bcsproject.NamespaceCallbackResponse)
	if err := json.Unmarshal(bs, resp); err != nil {
		return nil, errors.Wrapf(err, "update namespaces itsm callback unmarshal failed with response '%s'", string(bs))
	}
	if resp != nil && resp.Code != 0 {
		return nil, fmt.Errorf("update namespaces itsm callback response code not 0 but %d: %s", resp.Code, resp.Message)
	}
	return resp, nil
}

// UpdateNamespaceVariables 更新命名变量
func (p *ProjectManagerClient) UpdateNamespaceVariables(in *UpdateNamespaceVariablesReq) (*bcsproject.UpdateNamespacesVariablesResponse, error) {
	bs, err := p.do(fmt.Sprintf(updateNamespaceVariablesUrl, in.ProjectCode, in.ClusterID, in.Namespace), http.MethodPut, nil, in)
	if err != nil {
		return nil, fmt.Errorf("update namespaces variables failed: %v", err)
	}
	resp := new(bcsproject.UpdateNamespacesVariablesResponse)
	if err := json.Unmarshal(bs, resp); err != nil {
		return nil, errors.Wrapf(err, "update namespaces variables unmarshal failed with response '%s'", string(bs))
	}
	if resp != nil && resp.Code != 0 {
		return nil, fmt.Errorf("update namespaces variables response code not 0 but %d: %s", resp.Code, resp.Message)
	}
	return resp, nil
}
