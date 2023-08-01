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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"io/ioutil"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/clusterops"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/user"
	iutils "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

const (
	apiServer  = "apiServer"
	etcdServer = "etcdServer"

	createCluster = "create_cluster"
	addNodes      = "add_nodes"

	// defaultPolicy default cpu_manager policy
	defaultPolicy = "none"
	staticPolicy  = "static"

	// bk-sops template vars prefix
	prefix   = "CM"
	template = "template"
)

var (
	// cluster render info
	clusterID           = "CM.cluster.ClusterID"
	clusterMasterIPs    = "CM.cluster.ClusterMasterIPs"
	clusterMasterDomain = "CM.cluster.ClusterMasterDomain"
	clusterEtcdDomain   = "CM.cluster.ClusterEtcdDomain"

	clusterRegion         = "CM.cluster.ClusterRegion"
	clusterVPC            = "CM.cluster.ClusterVPC"
	clusterNetworkType    = "CM.cluster.ClusterNetworkType"
	clusterBizID          = "CM.cluster.ClusterBizID"
	clusterModuleID       = "CM.cluster.ClusterModuleID"
	clusterExtraID        = "CM.cluster.ClusterExtraID"
	clusterExtraClusterID = "CM.cluster.ClusterExtraClusterID"
	clusterProjectID      = "CM.cluster.ClusterProjectID"
	clusterExtraEnv       = "CM.cluster.CreateClusterExtraEnv"
	clusterManageEnv      = "CM.cluster.ClusterManageType"
	addNodesExtraEnv      = "CM.cluster.AddNodesExtraEnv"
	bcsCommonInfo         = "CM.bcs.CommonInfo"
	clusterKubeConfig     = "CM.cluster.Kubeconfig"

	// node render info
	// NOCC:gas/crypto(误报)
	nodePasswd           = "CM.node.NodePasswd"
	nodeCPUManagerPolicy = "CM.node.NodeCPUManagerPolicy"
	nodeIPList           = "CM.node.NodeIPList"

	// nodeGroup render info
	nodeGroupID = "CM.nodeGroup.NodeGroupID"

	externalNodeScript = "CM.node.Script"
	// 操作人员,追溯记录
	nodeOperator = "CM.node.NodeOperator"

	// 作为step参数动态注入业务ID和操作人员信息
	templateBusinessID = "CM.template.BusinessID"
	templateOperator   = "CM.template.Operator"

	// NodeIPList dynamic inject node ips
	NodeIPList = "CM.node.NodeIPList"
	// ExternalNodeScript external script
	ExternalNodeScript = "CM.node.Script"
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
	// BCSTokenKey xxx
	BCSTokenKey BcsKey = "bcs_token"
	// BCSApiIpsKey xxx
	BCSApiIpsKey BcsKey = "bcs_api_ips"
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

func getFileContent(file string) (string, error) {
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
		bcsEnvs = append(bcsEnvs, getEnv(BCSClientCert.String(), base64.StdEncoding.EncodeToString([]byte(clientCert))))
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
		token, err := utils.BuildBcsAgentToken(cluster.ClusterID, false)
		if err != nil {
			blog.Errorf("getBcsEnvs BuildBcsAgentToken[%s] failed: %v", cluster.ClusterID, err)
			return "", err
		}

		bcsEnvs = append(bcsEnvs, getEnv(BCSTokenKey.String(), token))
	}
	if options.GetEditionInfo().IsCommunicationEdition() {
		ipStr, err := getInitClusterIPs(common.InitClusterID)
		if err != nil {
			blog.Errorf("getBcsEnvs BuildBcsInitClusterIPs[%s] failed: %v", common.InitClusterID, err)
			return "", err
		}

		bcsEnvs = append(bcsEnvs, getEnv(BCSApiIpsKey.String(), ipStr))
	}

	return strings.Join(bcsEnvs, ";"), nil
}

// getInitClusterIPs 获取创始集群IP列表
func getInitClusterIPs(clusterID string) (string, error) {
	k8sOperator := clusterops.NewK8SOperator(options.GetGlobalCMOptions(), cloudprovider.GetStorageModel())
	nodes, err := k8sOperator.ListClusterNodes(context.Background(), clusterID)
	if err != nil {
		blog.Errorf("getInitClusterIPs[%s] failed: %v", clusterID, err)
		return "", err
	}

	var ips = make([]string, 0)
	for i := range nodes {
		ipv4s, ipv6s := iutils.GetNodeIPAddress(nodes[i])
		if len(ipv4s) == 0 && len(ipv6s) == 0 {
			continue
		}
		if len(ipv4s) > 0 {
			ips = append(ips, ipv4s...)
		}
		if len(ipv6s) > 0 {
			ips = append(ips, ipv6s...)
		}
	}

	return strings.Join(ips, ","), nil
}
