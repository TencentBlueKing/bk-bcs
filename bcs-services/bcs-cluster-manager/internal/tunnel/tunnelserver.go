/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package tunnel

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/modules"
	"github.com/Tencent/bk-bcs/bcs-common/common/websocketDialer"
	types "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
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
	KubeAgentModule = "kubeagent"
	// MesosDriverModule http tunnel header
	MesosDriverModule = "mesosdriver"
)

// RegisterCluster definition of tunnel cluster info
type RegisterCluster struct {
	Address   string `json:"address"`
	UserToken string `json:"userToken"`
	CACert    string `json:"caCert"`
}

// WsTunnelServerCallback tunnel server wrapper
type WsTunnelServerCallback struct {
	tunnelServer *websocketDialer.Server
	model        store.ClusterManagerModel
}

// NewWsTunnelServerCallback create websocket tunnel
func NewWsTunnelServerCallback(model store.ClusterManagerModel) *WsTunnelServerCallback {
	wts := &WsTunnelServerCallback{
		model: model,
	}
	wts.tunnelServer = websocketDialer.New(
		wts.authorizeTunnel,
		websocketDialer.DefaultErrorWriter,
		wts.cleanCredential)
	return wts
}

// GetTunnelServer get websocket tunnel server
func (wts *WsTunnelServerCallback) GetTunnelServer() *websocketDialer.Server {
	return wts.tunnelServer
}

// authorizeTunnel authorize an client
// 1. check connection module name, clusterID, cluster credential
// 2. store cluster credentials
func (wts *WsTunnelServerCallback) authorizeTunnel(req *http.Request) (string, bool, error) {
	// module name, cluster is necessary
	moduleName := req.Header.Get(Module)
	if moduleName == "" {
		blog.Errorf("module empty")
		return "", false, fmt.Errorf("module empty")
	}
	clusterID := req.Header.Get(Cluster)
	if clusterID == "" {
		blog.Errorf("clusterID empty")
		return "", false, fmt.Errorf("clusterID empty")
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
		blog.Errorf("client dialer address is empty")
		return "", false, fmt.Errorf("client dialer address is empty")
	}

	// bcs-kube-agent must report ca and usertoken
	if moduleName == KubeAgentModule {
		if registerCluster.CACert == "" || registerCluster.UserToken == "" {
			blog.Errorf("address or cacert or token empty")
			return "", false, fmt.Errorf("address or cacert or token empty")
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

	if moduleName == KubeAgentModule {
		newCredential := &types.ClusterCredential{
			ServerKey:     clusterID,
			ClusterID:     clusterID,
			ClientModule:  moduleName,
			ServerAddress: registerCluster.Address,
			CaCertData:    caCert,
			UserToken:     registerCluster.UserToken,
			ConnectMode:   modules.BCSConnectModeTunnel,
		}
		if err := wts.model.PutClusterCredential(context.TODO(), newCredential); err != nil {
			blog.Errorf("error when put cluster credential, err %s", err.Error())
			return "", false, err
		}
		return clusterID, true, nil
	} else if moduleName == MesosDriverModule {
		// for mesos, the registerCluster.Address is mesos-driver url.
		// one mesos cluster may have 3 or more mesos-driver,
		// so we should distinguish them, so use {clusterID}-{ip:port} as serverKey
		url, err := url.Parse(registerCluster.Address)
		if err != nil {
			return "", false, nil
		}
		serverKey := clusterID + "-" + url.Host
		newCredential := &types.ClusterCredential{
			ServerKey:     serverKey,
			ClusterID:     clusterID,
			ClientModule:  moduleName,
			ServerAddress: registerCluster.Address,
			CaCertData:    caCert,
			UserToken:     registerCluster.UserToken,
			ConnectMode:   modules.BCSConnectModeTunnel,
		}
		if err = wts.model.PutClusterCredential(context.TODO(), newCredential); err != nil {
			blog.Errorf("error when put cluster credential, err %s", err.Error())
			return "", false, err
		}
		return serverKey, true, nil
	}
	return "", false, fmt.Errorf("unknown client module")
}

// clean credential
func (wts *WsTunnelServerCallback) cleanCredential(serverKey string) {
	// when multiple kube-agent connect to cluster-manager with same clientKey,
	// delete credential will make connection unusable
	// wts.model.DeleteClusterCredential(context.TODO(), serverKey)
}
