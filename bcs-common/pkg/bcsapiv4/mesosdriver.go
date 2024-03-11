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

package bcsapiv4

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Tencent/bk-bcs/bcs-common/common/http/httpclient"
	commtypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	moduleDiscovery "github.com/Tencent/bk-bcs/bcs-common/pkg/module-discovery"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/networkdetection/types"
	schedtypes "github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/types"
	blog "k8s.io/klog/v2"
)

// MesosDriver http API SDK definition
type MesosDriver interface {
	UpdateAgentExtendedResources(clusterID string, er *commtypes.ExtendedResource) error
	GetNodes(clusterid string) ([]*types.NodeInfo, error)
	CeateDeployment(clusterid string, deploy []byte) error
	FetchDeployment(deploy *types.DeployDetection) (interface{}, error)
	FetchPods(clusterid, ns, name string) ([]byte, error)
}

// MesosDriverClientConfig mesos driver client configuration
type MesosDriverClientConfig struct {
	ZkAddr string
	// http client cert config
	ClientCert *commtypes.CertConfig
}

// MesosDriverClient client for mesosdriver
type MesosDriverClient struct {
	conf *MesosDriverClientConfig

	// to discovery mesos driver address
	moduleDiscovery moduleDiscovery.ModuleDiscovery
	// http client
	cli *httpclient.HttpClient
}

// NewMesosDriverClient new MesosDriverClient object
func NewMesosDriverClient(conf *MesosDriverClientConfig) (*MesosDriverClient, error) {
	m := &MesosDriverClient{
		conf: conf,
	}

	var err error
	// start module discovery to discovery mesos driver
	// moduleDiscovery.GetModuleServers(module) can return server information
	m.moduleDiscovery, err = moduleDiscovery.NewDiscoveryV2(
		m.conf.ZkAddr,
		[]string{commtypes.BCS_MODULE_MESOSAPISERVER})
	if err != nil {
		return nil, err
	}
	blog.Infof("NewDiscoveryV2 done")

	// init http client
	m.cli = httpclient.NewHttpClient()
	// if https
	if m.conf.ClientCert != nil && m.conf.ClientCert.IsSSL {
		blog.Infof("NetworkDetection http client cert ssl")
		_ = m.cli.SetTlsVerity(m.conf.ClientCert.CAFile, m.conf.ClientCert.CertFile, m.conf.ClientCert.KeyFile,
			m.conf.ClientCert.CertPasswd)
	}
	m.cli.SetHeader("Content-Type", "application/json")
	m.cli.SetHeader("Accept", "application/json")
	return m, nil
}

// getModuleAddr xxx
// get module address
// clusterid, example BCS-MESOS-10001
// return first parameter: module address, example 127.0.0.1:8090
func (m *MesosDriverClient) getModuleAddr() (string, error) {
	serv, err := m.moduleDiscovery.GetRandModuleServer(commtypes.BCS_MODULE_MESOSAPISERVER)
	if err != nil {
		blog.Errorf("discovery zk %s module %s error %s",
			m.conf.ZkAddr, commtypes.BCS_MODULE_MESOSAPISERVER, err.Error())
		return "", err
	}
	// serv is string object
	data, _ := serv.(string)
	var servInfo *commtypes.BcsMesosApiserverInfo
	err = json.Unmarshal([]byte(data), &servInfo)
	if err != nil {
		blog.Errorf("getModuleAddr Unmarshal data(%s) to commtypes.BcsMesosApiserverInfo failed: %s", data, err.Error())
		return "", err
	}

	return fmt.Sprintf("%s://%s:%d", servInfo.Scheme, servInfo.IP, servInfo.Port), nil
}

// UpdateAgentExtendedResources update agent external resources
func (m *MesosDriverClient) UpdateAgentExtendedResources(clusterID string, er *commtypes.ExtendedResource) error {
	by, _ := json.Marshal(er)
	_, err := m.requestMesosApiserver(clusterID, http.MethodPut, "agentsettings/extendedresource", by)
	if err != nil {
		blog.Errorf("update agent %s external resources error %s", er.InnerIP, err.Error())
		return err
	}
	blog.Infof("update agent %s external resources %s success", er.InnerIP, er.Name)
	return nil
}

// GetNodes get cluster all nodes
func (m *MesosDriverClient) GetNodes(clusterid string) ([]*types.NodeInfo, error) {
	by, err := m.requestMesosApiserver(clusterid, http.MethodGet, "cluster/resources", nil)
	if err != nil {
		blog.Errorf("Get cluster %s nodes error %s", clusterid, err.Error())
		return nil, err
	}

	// Unmarshal BcsClusterResource
	var resource *commtypes.BcsClusterResource
	err = json.Unmarshal(by, &resource)
	if err != nil {
		blog.Errorf("Unmarshal body(%s) to commtypes.BcsClusterResource failed: %s", string(by), err.Error())
		return nil, err
	}

	// create NodeInfo Object
	nodes := make([]*types.NodeInfo, 0)
	for _, agent := range resource.Agents {
		node := &types.NodeInfo{
			Ip:        agent.IP,
			Clusterid: clusterid,
		}
		nodes = append(nodes, node)
	}
	return nodes, nil
}

// CeateDeployment deploy application
func (m *MesosDriverClient) CeateDeployment(clusterid string, deploy []byte) error {
	_, err := m.requestMesosApiserver(clusterid, http.MethodPost, "namespaces/bcs-system/deployments", deploy)

	return err
}

// FetchDeployment fetch application
func (m *MesosDriverClient) FetchDeployment(deploy *types.DeployDetection) (interface{}, error) {
	by, err := m.requestMesosApiserver(deploy.Clusterid, http.MethodGet,
		"/namespaces/bcs-system/applications", nil)
	if err != nil {
		return nil, err
	}

	var apps []*schedtypes.Application
	err = json.Unmarshal(by, &apps)
	if err != nil {
		return nil, fmt.Errorf("Unmarshal body(%s) to schedtypes.Application failed: %s", string(by), err.Error())
	}
	for _, app := range apps {
		if app.ObjectMeta.Annotations == nil {
			continue
		}
		if app.ObjectMeta.Annotations["idc"] == deploy.Idc {
			return app, nil
		}
	}

	return nil, fmt.Errorf("Not found")
}

// FetchPods fetch application't pods
func (m *MesosDriverClient) FetchPods(clusterid, ns, name string) ([]byte, error) {
	by, err := m.requestMesosApiserver(clusterid, http.MethodGet,
		fmt.Sprintf("/namespaces/%s/applications/%s/taskgroups", ns, name), nil)
	return by, err
}

// requestMesosApiserver xxx
// method=http.method: POST、GET、PUT、DELETE
// request url = address/url
// payload is request body
// if error!=nil, then request mesos failed, errom.Error() is failed message
// if error==nil, []byte is response body information
func (m *MesosDriverClient) requestMesosApiserver(clusterid, method, url string, payload []byte) ([]byte, error) {
	// get mesos api address
	addr, err := m.getModuleAddr()
	if err != nil {
		return nil, fmt.Errorf("get cluster %s mesosapi failed: %s", clusterid, err.Error())
	}
	uri := fmt.Sprintf("%s/mesosdriver/v4/%s", addr, url)
	m.cli.SetHeader("BCS-ClusterID", clusterid)
	blog.V(3).Infof("request %s body(%s)", uri, string(payload))

	var by []byte
	switch method {
	case "GET":
		by, err = m.cli.GET(uri, nil, payload)
	case "POST":
		by, err = m.cli.POST(uri, nil, payload)
	case "DELETE":
		by, err = m.cli.DELETE(uri, nil, payload)
	case "PUT":
		by, err = m.cli.PUT(uri, nil, payload)
	default:
		err = fmt.Errorf("uri %s method %s is invalid", uri, method)
	}
	if err != nil {
		return nil, err
	}

	// unmarshal response.body
	var result *commtypes.APIResponse
	err = json.Unmarshal(by, &result)
	if err != nil {
		return nil, fmt.Errorf("Unmarshal body(%s) failed: %s", string(by), err.Error())
	}
	// if result.Result==false, then request failed
	if !result.Result {
		return nil, fmt.Errorf("request %s failed: %s", uri, result.Message)
	}
	by, _ = json.Marshal(result.Data)
	return by, nil
}
