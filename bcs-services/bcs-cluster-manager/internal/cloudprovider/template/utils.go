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

package template

import (
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"strings"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/user"
)

// BcsKey bcsEnvs key
type BcsKey string

// String xxx
func (bk BcsKey) String() string {
	return string(bk)
}

var (
	// BCSCA xxx
	BCSCA BcsKey = "bcs_ca"
	// BCSClientCert xxx
	BCSClientCert BcsKey = "bcs_client_cert"
	// BCSClientKey xxx
	BCSClientKey BcsKey = "bcs_client_key"
	// BCSToken xxx
	BCSTokenKey BcsKey = "bcs_token"
)

func getClusterMasterIPs(cluster *proto.Cluster) string {
	masterIPs := make([]string, 0)
	for ip := range cluster.Master {
		masterIPs = append(masterIPs, ip)
	}

	return strings.Join(masterIPs, ",")
}

func getMasterDomain(cls *proto.Cluster) string {
	server, ok := cls.ExtraInfo[apiServer]
	if ok {
		return server
	}

	return ""
}

func getEtcdDomain(cls *proto.Cluster) string {
	etcd, ok := cls.ExtraInfo[etcdServer]
	if ok {
		return etcd
	}

	return ""
}

func getClusterType(cls *proto.Cluster) string {
	if len(cls.GetExtraClusterID()) > 0 {
		return "1"
	}

	return "0"
}

func getClusterCreateExtraEnv(cls *proto.Cluster) string {
	value, ok := cls.ExtraInfo[createCluster]
	if ok {
		return value
	}

	return ""
}

func getAddNodesExtraEnv(cls *proto.Cluster) string {
	value, ok := cls.ExtraInfo[addNodes]
	if ok {
		return value
	}

	return ""
}

func getEnv(k, v string) string {
	return fmt.Sprintf("%s=%s", k, v)
}

func getFileContent(file string) (string, error){
	body, err := ioutil.ReadFile(file)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// getBcsEnvs get bcs platform common parameters
func getBcsEnvs(cluster *proto.Cluster) (string, error) {
	cloud, err := cloudprovider.GetStorageModel().GetCloud(context.Background(), cluster.Provider)
	if err != nil {
		return "", err
	}
	// credential info
	bcsEnvs := make([]string, 0)
	opts := options.GetGlobalCMOptions()
	if opts.ClientCa != "" {
		clientCa, _ := getFileContent(opts.ClientCa)
		bcsEnvs = append(bcsEnvs, getEnv(BCSCA.String(), base64.StdEncoding.EncodeToString([]byte(clientCa))))
	}
	if opts.ClientCert != "" {
		clientCert, _ := getFileContent(opts.ClientCert)
		bcsEnvs = append(bcsEnvs, getEnv(BCSCA.String(), base64.StdEncoding.EncodeToString([]byte(clientCert))))
	}
	if opts.ClientKey != "" {
		clientKey, _ := getFileContent(opts.ClientKey)
		bcsEnvs = append(bcsEnvs, getEnv(BCSClientKey.String(), base64.StdEncoding.EncodeToString([]byte(clientKey))))
	}

	// get cloud platform common config
	for k, v := range cloud.PlatformInfo {
		bcsEnvs = append(bcsEnvs, getEnv(k, v))
	}

	if user.GetUserManagerClient() != nil {
		token, err := utils.BuildBcsAgentToken(cluster)
		if err != nil {
			return "", err
		}

		bcsEnvs = append(bcsEnvs, getEnv(BCSTokenKey.String(), token))
	}

	return strings.Join(bcsEnvs, ","), nil
}
