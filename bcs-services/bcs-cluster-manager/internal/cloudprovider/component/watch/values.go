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

package watch

import (
	cmoptions "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"

	"gopkg.in/yaml.v2"
)

var (
	defaultTemplateName = "bcs-k8s-watch"
	defaultReplicas     = 1
)

// ValuesTemplate watch values
type ValuesTemplate struct {
	ReplicaCount int `yaml:"replicaCount"`
	Env          struct {
		ClusterID     string `yaml:"BK_BCS_clusterId"`
		StorageServer string `yaml:"BK_BCS_customStorage"`
	}
	Secret struct {
		ClientCaCrt  string `yaml:"ca_crt"`
		ClientTlsCrt string `yaml:"tls_crt"`
		ClientTlsKey string `yaml:"tls_key"`
	}
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

	// get config info
	op := cmoptions.GetGlobalCMOptions()
	var (
		clientKey, _  = utils.GetFileContent(op.ClientKey)
		clientCert, _ = utils.GetFileContent(op.ClientCert)
		clientCa, _   = utils.GetFileContent(op.ClientCa)
	)

	values := ValuesTemplate{
		ReplicaCount: bw.Replicas,
		Env: struct {
			ClusterID     string `yaml:"BK_BCS_clusterId"`
			StorageServer string `yaml:"BK_BCS_customStorage"`
		}{
			ClusterID:     bw.ClusterID,
			StorageServer: op.ComponentDeploy.Watch.StorageServer,
		},
		Secret: struct {
			ClientCaCrt  string `yaml:"ca_crt"`
			ClientTlsCrt string `yaml:"tls_crt"`
			ClientTlsKey string `yaml:"tls_key"`
		}{
			ClientCaCrt:  clientCa,
			ClientTlsCrt: clientCert,
			ClientTlsKey: clientKey,
		},
	}
	renderValues, _ := yaml.Marshal(values)

	return string(renderValues), nil
}
