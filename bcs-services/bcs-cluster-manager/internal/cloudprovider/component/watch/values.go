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

package watch

import (
	"github.com/Tencent/bk-bcs/bcs-common/common/static"
	"gopkg.in/yaml.v2"

	cloudproviderUtils "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/utils"
	cmoptions "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

var (
	defaultTemplateName = "bcs-k8s-watch" // nolint
	defaultReplicas     = 1

	defaultCpuMem = CpuMem{
		Cpu:    "1",
		Memory: "2Gi",
	}
)

// ValuesTemplate watch values
type ValuesTemplate struct {
	ReplicaCount int `yaml:"replicaCount"`
	Image        struct {
		Registry string `yaml:"registry,omitempty"`
	}
	Env struct {
		ClusterID     string `yaml:"BK_BCS_clusterId"`
		StorageServer string `yaml:"BK_BCS_customStorage"`
		StorageToken  string `yaml:"BK_BCS_customStorageToken"`
		ClientPwd     string `yaml:"BK_BCS_clientKeyPassword"`
	}
	Secret struct {
		CertsOverride bool   `yaml:"bcsCertsOverride"`
		ClientCaCrt   string `yaml:"ca_crt"`
		ClientTlsCrt  string `yaml:"tls_crt"`
		ClientTlsKey  string `yaml:"tls_key"`
	}
	Resources Resource `yaml:"resources"`
}

// Resource xxx
type Resource struct {
	Limits   CpuMem `yaml:"limits"`
	Requests CpuMem `yaml:"requests"`
}

// CpuMem resource
type CpuMem struct {
	Cpu    string `yaml:"cpu"`
	Memory string `yaml:"memory"`
}

// BcsWatch component paras
type BcsWatch struct {
	ClusterID     string
	CustomStorage string
	Replicas      int
}

// GetValues get BcsWatch values
func (bw *BcsWatch) GetValues() (string, error) {
	if bw.Replicas <= 0 {
		bw.Replicas = defaultReplicas
	}

	token, err := cloudproviderUtils.BuildBcsAgentToken(bw.ClusterID, false)
	if err != nil {
		return "", err
	}

	// get config info
	op := cmoptions.GetGlobalCMOptions()
	var (
		clientKey, _  = utils.GetFileContent(op.ClientKey)
		clientCert, _ = utils.GetFileContent(op.ClientCert)
		clientCa, _   = utils.GetFileContent(op.ClientCa)
	)

	values := ValuesTemplate{
		ReplicaCount: bw.Replicas,
		Image: struct {
			Registry string `yaml:"registry,omitempty"`
		}{
			Registry: op.ComponentDeploy.Registry,
		},
		Env: struct {
			ClusterID     string `yaml:"BK_BCS_clusterId"`
			StorageServer string `yaml:"BK_BCS_customStorage"`
			StorageToken  string `yaml:"BK_BCS_customStorageToken"`
			ClientPwd     string `yaml:"BK_BCS_clientKeyPassword"`
		}{
			ClusterID:     bw.ClusterID,
			StorageServer: op.ComponentDeploy.Watch.StorageServer,
			StorageToken:  token,
			ClientPwd:     static.ClientCertPwd,
		},
		Secret: struct {
			CertsOverride bool   `yaml:"bcsCertsOverride"`
			ClientCaCrt   string `yaml:"ca_crt"`
			ClientTlsCrt  string `yaml:"tls_crt"`
			ClientTlsKey  string `yaml:"tls_key"`
		}{
			CertsOverride: true,
			ClientCaCrt:   clientCa,
			ClientTlsCrt:  clientCert,
			ClientTlsKey:  clientKey,
		},
		Resources: Resource{
			Requests: defaultCpuMem,
			Limits:   defaultCpuMem,
		},
	}
	renderValues, _ := yaml.Marshal(values)

	return string(renderValues), nil
}
