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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-api/pkg/storages/sqlstore"
)

const (
	Module        = "BCS-API-Tunnel-Module"
	RegisterToken = "BCS-API-Tunnel-Token"
	Params        = "BCS-API-Tunnel-Params"
	Cluster       = "BCS-API-Tunnel-ClusterId"

	KubeAgentModule   = "kube-agent"
	K8sDriverModule   = "k8s-driver"
	MesosDriverModule = "mesos-driver"
)

var (
	DefaultTunnelServer *websocketDialer.Server
	errFailedAuth       = errors.New("failed authentication")
)

type RegisterCluster struct {
	Address   string `json:"address"`
	UserToken string `json:"userToken"`
	CACert    string `json:"caCert"`
}

// authorizeTunnel authorize an client
func authorizeTunnel(req *http.Request) (string, bool, error) {
	moduleName := req.Header.Get(Module)
	if moduleName == "" {
		return "", false, errors.New("module empty")
	}

	registerToken := req.Header.Get(RegisterToken)
	if registerToken == "" {
		return "", false, errors.New("registerToken empty")
	}

	clusterId := req.Header.Get(Cluster)
	if clusterId == "" {
		return "", false, errors.New("clusterId empty")
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

	if moduleName == KubeAgentModule {
		if registerCluster.CACert == "" || registerCluster.UserToken == "" {
			return "", false, errors.New("address or cacert or token empty")
		}
	}

	var caCert string
	if registerCluster.CACert != "" {
		certBytes, err := base64.StdEncoding.DecodeString(registerCluster.CACert)
		if err != nil {
			blog.Errorf("error when decode cluster [%s] cacert registered by websocket: %s", clusterId, err.Error())
			return "", false, err
		}
		caCert = string(certBytes)
	}

	// validate if the registerToken is correct
	if moduleName == KubeAgentModule {
		token := sqlstore.GetRegisterToken(clusterId)
		if token == nil {
			return "", false, nil
		}
		if token.Token != registerToken {
			return "", false, nil
		}

		err = sqlstore.SaveWsCredentials(clusterId, moduleName, registerCluster.Address, caCert, registerCluster.UserToken)
		if err != nil {
			blog.Errorf("error when save websocket credentials: %s", err.Error())
			return "", false, err
		}
		return clusterId, true, nil
	} else if moduleName == MesosDriverModule || moduleName == K8sDriverModule {
		cluster := sqlstore.GetClusterByBCSInfo("", clusterId)
		if cluster == nil {
			return "", false, nil
		}
		token := sqlstore.GetRegisterToken(cluster.ID)
		if token == nil {
			return "", false, nil
		}
		if token.Token != registerToken {
			return "", false, nil
		}

		url, err := url.Parse(registerCluster.Address)
		if err != nil {
			return "", false, nil
		}
		serverKey := cluster.ID + "-" + url.Host
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
