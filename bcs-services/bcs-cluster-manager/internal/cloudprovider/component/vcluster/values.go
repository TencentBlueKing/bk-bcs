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

package vcluster

import (
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	cutils "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	cmoptions "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/user"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"

	"gopkg.in/yaml.v2"
)

// ValuesTemplates for vcluster values
type ValuesTemplates struct {
	Service     Service    `yaml:"service"`
	Etcd        OriginEtcd `yaml:"etcd"`
	TkeEtcd     TkeEtcd    `yaml:"tkeEtcd"`
	ServiceCIDR string     `yaml:"serviceCIDR"`
	KubeAgent   KubeAgent  `yaml:"kubeAgent"`
}

// Service devnet:false ; idc:true
type Service struct {
	Disable bool `yaml:"disabled"`
}

// OriginEtcd use origin etcd
type OriginEtcd struct {
	Disable bool `yaml:"disabled"`
}

// TkeEtcd use cloud tke etcd
type TkeEtcd struct {
	Enabled    bool   `yaml:"enabled"`
	Servers    string `yaml:"servers"`
	Ca         string `yaml:"ca"`
	ClientCert string `yaml:"clientCert"`
	ClientKey  string `yaml:"clientKey"`
}

// KubeAgent for kubeAgent conf
type KubeAgent struct {
	Secret    Secret `yaml:"secret"`
	ClusterID string `yaml:"clusterID"`
	Args      Args   `yaml:"args"`
}

// Secret for kube-agent cert
type Secret struct {
	BcsCa         string `yaml:"bcsCa"`
	BcsClientCert string `yaml:"bcsClientCert"`
	BcsClientKey  string `yaml:"bcsClientKey"`
}

// Args for kube-agent vales
type Args struct {
	BkBcsApi        string `yaml:"BK_BCS_API"`
	WebSocketTunnel string `yaml:"BK_BCS_kubeAgentWSTunnel"`
	KubeAgentProxy  string `yaml:"BK_BCS_kubeAgentProxy,omitempty"`
	Token           string `yaml:"BK_BCS_APIToken"`
}

// 原生etcd: false开启 true关闭
// tke etcd: true开启 false关闭

// Vcluster component paras
type Vcluster struct {
	Env utils.EnvType

	EtcdServers    string
	EtcdCA         string
	EtcdClientCert string
	EtcdClientKey  string

	ServiceCIDR string
	ClusterID   string
	ClusterEnv  string

	// idc环境需要连接k8s代理地址
	AgentProxyAddress string
}

// GetValues get vcluster values
func (vc *Vcluster) GetValues() (string, error) {
	// get config info
	op := cmoptions.GetGlobalCMOptions()
	var (
		clientKey, _  = utils.GetFileContent(op.ClientKey)
		clientCert, _ = utils.GetFileContent(op.ClientCert)
		clientCa, _   = utils.GetFileContent(op.ClientCa)
	)

	values := ValuesTemplates{
		Etcd:        OriginEtcd{},
		TkeEtcd:     TkeEtcd{},
		ServiceCIDR: vc.ServiceCIDR,
		KubeAgent: KubeAgent{
			ClusterID: vc.ClusterID,
			Secret: Secret{
				BcsCa:         clientCa,
				BcsClientCert: clientCert,
				BcsClientKey:  clientKey,
			},
		},
	}

	switch vc.Env {
	case utils.DEVNET:
		values.Service.Disable = false
		values.Etcd.Disable = false
		values.TkeEtcd.Enabled = false
		values.KubeAgent.Args.WebSocketTunnel = "true"
		values.KubeAgent.Args.KubeAgentProxy = ""
		switch vc.ClusterEnv {
		case common.Prod:
			values.KubeAgent.Args.BkBcsApi = op.ComponentDeploy.Vcluster.WsServer
		case common.Debug:
			values.KubeAgent.Args.BkBcsApi = op.ComponentDeploy.Vcluster.DebugWsServer
		default:
			values.KubeAgent.Args.BkBcsApi = op.ComponentDeploy.Vcluster.WsServer
		}
	case utils.IDC:
		values.Service.Disable = true
		values.Etcd.Disable = true
		values.TkeEtcd.Enabled = true
		values.TkeEtcd.Servers = vc.EtcdServers
		values.TkeEtcd.Ca = vc.EtcdCA
		values.TkeEtcd.ClientKey = vc.EtcdClientKey
		values.TkeEtcd.ClientCert = vc.EtcdClientCert
		values.KubeAgent.Args.KubeAgentProxy = vc.AgentProxyAddress
		values.KubeAgent.Args.BkBcsApi = op.ComponentDeploy.Vcluster.HttpServer
	default:
		return "", fmt.Errorf("vcluster not support env[%s]", vc.Env.String())
	}

	// generate cluster token for kube-agent
	if user.GetUserManagerClient() == nil {
		return "", fmt.Errorf("generate token failed: user module empty")
	}
	token, err := cutils.BuildBcsAgentToken(vc.ClusterID, false)
	if err != nil {
		return "", err
	}
	values.KubeAgent.Args.Token = token

	blog.Infof("Vcluster GetValues %+v", values)

	renderValues, _ := yaml.Marshal(values)

	return string(renderValues), nil
}
