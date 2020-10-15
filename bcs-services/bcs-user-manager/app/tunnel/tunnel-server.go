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

package tunnel

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/websocketDialer"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/storages/sqlstore"
)

const (
	// Module http tunnel header
	Module = "BCS-API-Tunnel-Module"
	// RegisterToken http tunnel header
	RegisterToken = "BCS-API-Tunnel-Token"
	// Params http tunnel header
	Params = "BCS-API-Tunnel-Params"
	// Cluster http tunnel header
	Cluster = "BCS-API-Tunnel-ClusterId"
	// KubeAgentModule http tunnel header
	KubeAgentModule = "kube-agent"
	// K8sDriverModule http tunnel header
	K8sDriverModule = "k8s-driver"
	// MesosDriverModule http tunnel header
	MesosDriverModule = "mesos-driver"
)

var (
	// DefaultTunnelServer default server implementation
	DefaultTunnelServer *websocketDialer.Server
	errFailedAuth       = errors.New("failed authentication")
)

// RegisterCluster definition of tunnel cluster info
type RegisterCluster struct {
	Address   string `json:"address"`
	UserToken string `json:"userToken"`
	CACert    string `json:"caCert"`
}

// authorizeTunnel authorize an client
func authorizeTunnel(req *http.Request) (string, bool, error) {
	// module name, register_token, cluster is necessary
	moduleName := req.Header.Get(Module)
	if moduleName == "" {
		return "", false, errors.New("module empty")
	}

	registerToken := req.Header.Get(RegisterToken)
	if registerToken == "" {
		return "", false, errors.New("registerToken empty")
	}

	clusterID := req.Header.Get(Cluster)
	if clusterID == "" {
		return "", false, errors.New("clusterID empty")
	}

	var registerCluster RegisterCluster
	params := req.Header.Get(Params)
	bytes, err := base64.StdEncoding.DecodeString(params)
	if err != nil {
		blog.Errorf("error when decode cluster params registered by websocket: %s", err.Error())
		return "", false, err
	}
	if err := json.Unmarshal(bytes, &registerCluster); err != nil {
		blog.Errorf("error when unmarshal cluster params registered by websocket: %s", err.Error())
		return "", false, err
	}

	if registerCluster.Address == "" {
		return "", false, errors.New("client dialer address is empty")
	}

	// bcs-kube-agent must report  ca and usertoken
	if moduleName == KubeAgentModule {
		if registerCluster.CACert == "" || registerCluster.UserToken == "" {
			return "", false, errors.New("address or cacert or token empty")
		}
	}

	var caCert string
	if registerCluster.CACert != "" {
		certBytes, err := base64.StdEncoding.DecodeString(registerCluster.CACert)
		if err != nil {
			blog.Errorf("error when decode cluster [%s] cacert registered by websocket: %s", clusterID, err.Error())
			return "", false, err
		}
		caCert = string(certBytes)
	}

	// validate if the registerToken is correct
	token := sqlstore.GetRegisterToken(clusterID)
	if token == nil {
		blog.Info("haha")
		return "", false, nil
	}
	if token.Token != registerToken {
		return "", false, nil
	}

	if moduleName == KubeAgentModule {
		// for k8s, the registerCluster.Address is kubernetes service url, just save to db
		err = sqlstore.SaveWsCredentials(clusterID, moduleName, registerCluster.Address, caCert, registerCluster.UserToken)
		if err != nil {
			blog.Errorf("error when save websocket credentials: %s", err.Error())
			return "", false, err
		}
		return clusterID, true, nil
	} else if moduleName == MesosDriverModule || moduleName == K8sDriverModule {
		// for mesos, the registerCluster.Address is mesos-driver url. one mesos cluster may have 3 or more mesos-driver,
		// so we should distinguish them, so use {clusterID}-{ip:port} as serverKey
		url, err := url.Parse(registerCluster.Address)
		if err != nil {
			return "", false, nil
		}
		serverKey := clusterID + "-" + url.Host
		err = sqlstore.SaveWsCredentials(serverKey, moduleName, registerCluster.Address, caCert, registerCluster.UserToken)
		if err != nil {
			blog.Errorf("error when save websocket credentials: %s", err.Error())
			return "", false, err
		}
		return serverKey, true, nil
	}

	return "", false, errors.New("unknown client module")
}

// NewTunnelServer create websocket tunnel server
func NewTunnelServer() *websocketDialer.Server {
	DefaultTunnelServer = websocketDialer.New(authorizeTunnel, websocketDialer.DefaultErrorWriter, cleanCredentials)
	return DefaultTunnelServer
}

// cleanCredentials clean client credentials in db
func cleanCredentials(serverKey string) {
	sqlstore.DelWsCredentials(serverKey)
}
