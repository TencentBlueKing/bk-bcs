/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 *  Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 *  Licensed under the MIT License (the "License"); you may not use this file except
 *  in compliance with the License. You may obtain a copy of the License at
 *  http://opensource.org/licenses/MIT
 *  Unless required by applicable law or agreed to in writing, software distributed under
 *  the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 *  either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package pkg

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/Tencent/bk-bcs/bcs-common/common/http/httpclient"
	datamanager "github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/proto/bcs-data-manager"
	"github.com/spf13/viper"
)

// Config describe the options Client need
type Config struct {
	// APIServer for bcs-api-gateway address
	APIServer string
	// AuthToken for bcs permission token
	AuthToken string
	// Operator for the bk-repo operations
	Operator string
}

// NewClientWithConfiguration new client with config
func NewClientWithConfiguration() DataManagerClient {
	return NewDataManagerCli(&Config{
		APIServer: viper.GetString("config.apiserver"),
		AuthToken: viper.GetString("config.bcs_token"),
		Operator:  viper.GetString("config.operator"),
	})
}

// DataManagerClient dataManagerClient interface
type DataManagerClient interface {
	GetProjectInfo(req *datamanager.GetProjectInfoRequest) (*datamanager.GetProjectInfoResponse, error)
	GetClusterInfoList(req *datamanager.GetClusterInfoListRequest) (*datamanager.GetClusterInfoListResponse, error)
	GetClusterInfo(req *datamanager.GetClusterInfoRequest) (*datamanager.GetClusterInfoResponse, error)
	GetNamespaceInfoList(req *datamanager.GetNamespaceInfoListRequest) (*datamanager.GetNamespaceInfoListResponse, error)
	GetNamespaceInfo(req *datamanager.GetNamespaceInfoRequest) (*datamanager.GetNamespaceInfoResponse, error)
	GetWorkloadInfoList(req *datamanager.GetWorkloadInfoListRequest) (*datamanager.GetWorkloadInfoListResponse, error)
	GetWorkloadInfo(req *datamanager.GetWorkloadInfoRequest) (*datamanager.GetWorkloadInfoResponse, error)
}

type dataManager struct {
	clientOption  *Config
	requestClient Requester
	defaultHeader http.Header
}

// NewDataManagerCli create client for bcs-mesh-manager
func NewDataManagerCli(config *Config) DataManagerClient {
	m := &dataManager{
		clientOption:  config,
		requestClient: newRequester(),
		defaultHeader: http.Header{},
	}
	m.defaultHeader.Set("Content-Type", "application/json")
	m.defaultHeader.Set("Authorization", "Bearer "+config.AuthToken)
	return m
}

// Requester interface
type Requester interface {
	DoRequest(url, method string, header http.Header, data []byte) ([]byte, error)
}

type requester struct {
	httpCli *httpclient.HttpClient
}

func newRequester() Requester {
	return &requester{httpCli: httpclient.NewHttpClient()}
}

// DoRequest do request
func (r *requester) DoRequest(url, method string, header http.Header, data []byte) ([]byte, error) {
	rsp, err := r.httpCli.Request(url, method, header, data)
	if err != nil {
		return nil, fmt.Errorf("do request error, url: %s, error:%v", url, err)
	}
	return rsp, nil
}

// GetProjectInfo get project
func (m *dataManager) GetProjectInfo(req *datamanager.GetProjectInfoRequest) (
	*datamanager.GetProjectInfoResponse, error) {
	url := m.clientOption.APIServer + PrefixUrl + fmt.Sprintf(GetProjectUrl, req.ProjectID, req.Dimension)
	rsp, err := m.requestClient.DoRequest(url, "GET", m.defaultHeader, nil)
	if err != nil {
		return nil, fmt.Errorf("get project info error, url: %s, error: %v", url, err)
	}
	var result datamanager.GetProjectInfoResponse
	if err = json.Unmarshal(rsp, &result); err != nil {
		return nil, fmt.Errorf("result decode err: %v", err)
	}
	return &result, nil
}

// GetClusterInfoList list clusters
func (m *dataManager) GetClusterInfoList(req *datamanager.GetClusterInfoListRequest) (
	*datamanager.GetClusterInfoListResponse, error) {
	url := m.clientOption.APIServer + PrefixUrl + fmt.Sprintf(ListClusterUrl, req.ProjectID, req.Dimension,
		strconv.Itoa(int(req.Page)), strconv.Itoa(int(req.Size)))
	rsp, err := m.requestClient.DoRequest(url, "GET", m.defaultHeader, nil)
	if err != nil {
		return nil, fmt.Errorf("list cluster info error, url: %s, error: %v", url, err)
	}
	var result datamanager.GetClusterInfoListResponse
	if err = json.Unmarshal(rsp, &result); err != nil {
		return nil, fmt.Errorf("result decode err: %v", err)
	}
	return &result, nil
}

// GetClusterInfo get cluster
func (m *dataManager) GetClusterInfo(req *datamanager.GetClusterInfoRequest) (
	*datamanager.GetClusterInfoResponse, error) {
	url := m.clientOption.APIServer + PrefixUrl + fmt.Sprintf(GetClusterUrl, req.ClusterID, req.Dimension)
	rsp, err := m.requestClient.DoRequest(url, "GET", m.defaultHeader, nil)
	if err != nil {
		return nil, fmt.Errorf("get cluster info error, url: %s, error: %v", url, err)
	}
	var result datamanager.GetClusterInfoResponse
	if err = json.Unmarshal(rsp, &result); err != nil {
		return nil, fmt.Errorf("result decode err: %v", err)
	}
	return &result, nil
}

// GetNamespaceInfoList list namespaces
func (m *dataManager) GetNamespaceInfoList(req *datamanager.GetNamespaceInfoListRequest) (
	*datamanager.GetNamespaceInfoListResponse, error) {
	url := m.clientOption.APIServer + PrefixUrl + fmt.Sprintf(ListNamespaceUrl, req.ClusterID, req.Dimension,
		strconv.Itoa(int(req.Page)), strconv.Itoa(int(req.Size)))
	rsp, err := m.requestClient.DoRequest(url, "GET", m.defaultHeader, nil)
	if err != nil {
		return nil, fmt.Errorf("list namespace info error, url: %s, error: %v", url, err)
	}
	var result datamanager.GetNamespaceInfoListResponse
	if err = json.Unmarshal(rsp, &result); err != nil {
		return nil, fmt.Errorf("result decode err: %v", err)
	}
	return &result, nil
}

// GetNamespaceInfo get namespace
func (m *dataManager) GetNamespaceInfo(
	req *datamanager.GetNamespaceInfoRequest) (*datamanager.GetNamespaceInfoResponse, error) {
	url := m.clientOption.APIServer + PrefixUrl + fmt.Sprintf(GetNamespaceUrl, req.ClusterID, req.Namespace,
		req.Dimension)
	rsp, err := m.requestClient.DoRequest(url, "GET", m.defaultHeader, nil)
	if err != nil {
		return nil, fmt.Errorf("get namespace info error, url: %s, error: %v", url, err)
	}
	var result datamanager.GetNamespaceInfoResponse
	if err = json.Unmarshal(rsp, &result); err != nil {
		return nil, fmt.Errorf("result decode err: %v", err)
	}
	return &result, nil
}

// GetWorkloadInfoList list workloads
func (m *dataManager) GetWorkloadInfoList(req *datamanager.GetWorkloadInfoListRequest) (
	*datamanager.GetWorkloadInfoListResponse, error) {
	url := m.clientOption.APIServer + PrefixUrl +
		fmt.Sprintf(ListWorkloadUrl, req.ClusterID, req.Namespace, req.WorkloadType, req.Dimension,
			strconv.Itoa(int(req.Page)), strconv.Itoa(int(req.Size)))
	rsp, err := m.requestClient.DoRequest(url, "GET", m.defaultHeader, nil)
	if err != nil {
		return nil, fmt.Errorf("list workload info error, url: %s, error: %v", url, err)
	}
	var result datamanager.GetWorkloadInfoListResponse
	if err = json.Unmarshal(rsp, &result); err != nil {
		return nil, fmt.Errorf("result decode err: %v", err)
	}
	return &result, nil
}

// GetWorkloadInfo get workload
func (m *dataManager) GetWorkloadInfo(req *datamanager.GetWorkloadInfoRequest) (*datamanager.GetWorkloadInfoResponse,
	error) {
	url := m.clientOption.APIServer + PrefixUrl +
		fmt.Sprintf(GetWorkloadUrl, req.ClusterID, req.Namespace, req.WorkloadType, req.WorkloadName, req.Dimension)
	rsp, err := m.requestClient.DoRequest(url, "GET", m.defaultHeader, nil)
	if err != nil {
		return nil, fmt.Errorf("get workload info error, url: %s, error: %v", url, err)
	}
	var result datamanager.GetWorkloadInfoResponse
	if err = json.Unmarshal(rsp, &result); err != nil {
		return nil, fmt.Errorf("result decode err: %v", err)
	}
	return &result, nil
}
