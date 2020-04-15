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

package mesosdriver

import (
	"encoding/json"
	"fmt"
	"net/http"

	"bk-bcs/bcs-common/common/blog"
	"bk-bcs/bcs-common/common/http/httpclient"
	commtypes "bk-bcs/bcs-common/common/types"
	moduleDiscovery "bk-bcs/bcs-common/pkg/module-discovery"
)

type MesosDriverClient struct {
	conf *Config

	//to discovery mesos driver address
	moduleDiscovery moduleDiscovery.ModuleDiscovery
	//http client
	cli *httpclient.HttpClient
}

// new MesosPlatform object
func NewMesosPlatform(conf *Config) (*MesosDriverClient, error) {
	m := &MesosDriverClient{
		conf: conf,
	}

	var err error
	//start module discovery to discovery mesos driver
	//moduleDiscovery.GetModuleServers(module) can return server information
	m.moduleDiscovery, err = moduleDiscovery.NewDiscoveryV2(m.conf.ZkAddr, []string{commtypes.BCS_MODULE_MESOSAPISERVER})
	if err != nil {
		return nil, err
	}
	blog.Infof("NewDiscoveryV2 done")

	//init http client
	m.cli = httpclient.NewHttpClient()
	//if https
	if m.conf.ClientCert.IsSSL {
		blog.Infof("NetworkDetection http client cert ssl")
		m.cli.SetTlsVerity(m.conf.ClientCert.CAFile, m.conf.ClientCert.CertFile, m.conf.ClientCert.KeyFile,
			m.conf.ClientCert.CertPasswd)
	}
	m.cli.SetHeader("Content-Type", "application/json")
	m.cli.SetHeader("Accept", "application/json")
	return m, nil
}

//get module address
//clusterid, example BCS-MESOS-10001
//return first parameter: module address, example 127.0.0.1:8090
func (m *MesosDriverClient) getModuleAddr(clusterid string) (string, error) {
	serv, err := m.moduleDiscovery.GetRandModuleServer(commtypes.BCS_MODULE_MESOSAPISERVER)
	if err != nil {
		blog.Errorf("discovery zk %s module %s error %s", m.conf.ZkAddr, commtypes.BCS_MODULE_MESOSAPISERVER, err.Error())
		return "", err
	}
	//serv is string object
	data, _ := serv.(string)
	var servInfo *commtypes.BcsMesosApiserverInfo
	err = json.Unmarshal([]byte(data), &servInfo)
	if err != nil {
		blog.Errorf("getModuleAddr Unmarshal data(%s) to commtypes.BcsMesosApiserverInfo failed: %s", data, err.Error())
		return "", err
	}

	return fmt.Sprintf("%s://%s:%d", servInfo.Scheme, servInfo.IP, servInfo.Port), nil
}

//update agent external resources
func (m *MesosDriverClient) UpdateAgentExtendedResources(er *commtypes.ExtendedResource) error {
	_, err := m.requestMesosApiserver(m.conf.ClusterId, http.MethodPut, "agentsettings/extendedresources", nil)
	if err != nil {
		blog.Errorf("update agent %s external resources error %s", er.InnerIP, err.Error())
		return err
	}
	blog.Infof("update agent %s external resources %s success", er.InnerIP, er.Name)
	return nil
}

//method=http.method: POST、GET、PUT、DELETE
//request url = address/url
//payload is request body
//if error!=nil, then request mesos failed, errom.Error() is failed message
//if error==nil, []byte is response body information
func (m *MesosDriverClient) requestMesosApiserver(clusterid, method, url string, payload []byte) ([]byte, error) {
	//get mesos api address
	addr, err := m.getModuleAddr(clusterid)
	if err != nil {
		return nil, fmt.Errorf("get cluster %s mesosapi failed: %s", clusterid, err.Error())
	}
	uri := fmt.Sprintf("%s/mesosdriver/v4/%s", addr, url)
	m.cli.SetHeader("BCS-ClusterID", clusterid)

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

	//unmarshal response.body
	var result *commtypes.APIResponse
	err = json.Unmarshal(by, &result)
	if err != nil {
		return nil, fmt.Errorf("Unmarshal body(%s) failed: %s", string(by), err.Error())
	}
	//if result.Result==false, then request failed
	if !result.Result {
		return nil, fmt.Errorf("request %s failed: %s", uri, result.Message)
	}
	by, _ = json.Marshal(result.Data)
	return by, nil
}
