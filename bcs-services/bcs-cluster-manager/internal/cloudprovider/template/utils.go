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

package template

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/clusterops"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/user"
	iutils "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

const (
	// defaultPolicy default cpu_manager policy
	defaultPolicy = "none"
	staticPolicy  = "static"

	// bk-sops template vars prefix
	prefix   = "CM"
	template = "template"

	// disable xxx
	disable = "disable"
	// singleStack xxx
	singleStack = "singlestack"
	// dualStack xxx
	dualStack = "dualstack"
)

var (
	// cluster render info
	clusterID        = "CM.cluster.ClusterID"
	clusterMasterIPs = "CM.cluster.ClusterMasterIPs"

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
	nodePasswd           = "CM.node.NodePasswd"           // nolint
	nodeCPUManagerPolicy = "CM.node.NodeCPUManagerPolicy" // nolint
	nodeIPList           = "CM.node.NodeIPList"           // nolint

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

func getClusterType(cls *proto.Cluster) string {
	if len(cls.GetExtraClusterID()) > 0 {
		return "1"
	}

	return "0"
}

func getEnv(k, v string) string {
	return fmt.Sprintf("%s=%s", k, v)
}

func getFileContent(file string) (string, error) {
	body, err := os.ReadFile(file)
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

// self builder cluster envs

// EnvKey key
type EnvKey string

// String xxx
func (ek EnvKey) String() string {
	return string(ek)
}

var (
	// K8S_VER cluster version
	k8sVersion EnvKey = "K8S_VER"

	// CRI_TYPE runtime type
	criType EnvKey = "CRI_TYPE"
	// DOCKER_VER docker version
	dockerVersion EnvKey = "DOCKER_VER"
	// CONTAINERD_VER containerd version
	containerdVersion EnvKey = "CONTAINERD_VER"

	// K8S_POD_CIDR pod cidr
	podCidrIpv4 EnvKey = "K8S_POD_CIDR"
	// K8S_SVC_CIDR service cidr
	serviceCidrIpv4 EnvKey = "K8S_SVC_CIDR"
	// K8S_MASK ipv4 mask
	mask EnvKey = "K8S_MASK"

	// K8S_POD_CIDRv6 pod ipv6 cidr
	podCidripv6 EnvKey = "K8S_POD_CIDRv6"
	// K8S_SVC_CIDRv6 service ipv6 cidr
	serviceCidrIpv6 EnvKey = "K8S_SVC_CIDRv6"
	// K8S_IPv6_MASK ipv6 mask
	maskIpv6 EnvKey = "K8S_IPv6_MASK"

	// K8S_IPv6_STATUS = disable - ipv4单栈 | singlestack - ipv6 单栈 | dualstack - ipv4/ipv6 双栈;  集群单双栈
	ipv6Status EnvKey = "K8S_IPv6_STATUS"

	// K8S_CNI cni plugin
	cniPlugin EnvKey = "K8S_CNI"
	// ENABLE_APISERVER_HA enable apiserver ha
	apiserverHa EnvKey = "ENABLE_APISERVER_HA"
)

func getSelfBuilderClusterEnvs(cls *proto.Cluster) (string, error) {
	envs := make([]string, 0)

	// cluster version & runtime & runtimeVersion
	version := getEnv(k8sVersion.String(), cls.GetClusterBasicSettings().Version)
	runtime := getEnv(criType.String(), cls.GetClusterAdvanceSettings().GetContainerRuntime())
	dockerVer := getEnv(dockerVersion.String(), cls.GetClusterAdvanceSettings().GetRuntimeVersion())
	containerdVer := getEnv(containerdVersion.String(), cls.GetClusterAdvanceSettings().GetRuntimeVersion())

	envs = append(envs, version, runtime, dockerVer, containerdVer)

	// check ipv4 or ipv6
	ipType := iutils.IPV4
	if cls.GetNetworkSettings().GetClusterIpType() != "" {
		ipType = cls.GetNetworkSettings().GetClusterIpType()
	}

	switch ipType {
	case iutils.IPV4:
		podCidr := getEnv(podCidrIpv4.String(), cls.GetNetworkSettings().GetClusterIPv4CIDR())
		serviceCidr := getEnv(serviceCidrIpv4.String(), cls.GetNetworkSettings().GetServiceIPv4CIDR())
		size, _ := iutils.GetMaskLenByNum(iutils.IPV4, float64(cls.GetNetworkSettings().GetMaxNodePodNum()))
		ipv4Mask := getEnv(mask.String(), fmt.Sprintf("%v", size))
		stack := getEnv(ipv6Status.String(), disable)

		envs = append(envs, podCidr, serviceCidr, ipv4Mask, stack)
	case iutils.IPV6:
		podCidr := getEnv(podCidripv6.String(), cls.GetNetworkSettings().GetClusterIPv6CIDR())
		serviceCidr := getEnv(serviceCidrIpv6.String(), cls.GetNetworkSettings().GetServiceIPv6CIDR())
		size, _ := iutils.GetMaskLenByNum(iutils.IPV6, float64(cls.GetNetworkSettings().GetMaxNodePodNum()))
		ipv6Mask := getEnv(maskIpv6.String(), fmt.Sprintf("%v", size))
		stack := getEnv(ipv6Status.String(), singleStack)

		envs = append(envs, podCidr, serviceCidr, ipv6Mask, stack)
	case iutils.DualStack:
		ipv4PodCidr := getEnv(podCidrIpv4.String(), cls.GetNetworkSettings().GetClusterIPv4CIDR())
		ipv4ServiceCidr := getEnv(serviceCidrIpv4.String(), cls.GetNetworkSettings().GetServiceIPv4CIDR())
		size, _ := iutils.GetMaskLenByNum(iutils.IPV4, float64(cls.GetNetworkSettings().GetMaxNodePodNum()))
		ipv4Mask := getEnv(mask.String(), fmt.Sprintf("%v", size))

		ipv6PodCidr := getEnv(podCidripv6.String(), cls.GetNetworkSettings().GetClusterIPv6CIDR())
		ipv6ServiceCidr := getEnv(serviceCidrIpv6.String(), cls.GetNetworkSettings().GetServiceIPv6CIDR())
		size, _ = iutils.GetMaskLenByNum(iutils.IPV6, float64(cls.GetNetworkSettings().GetMaxNodePodNum()))
		ipv6Mask := getEnv(maskIpv6.String(), fmt.Sprintf("%v", size))

		stack := getEnv(ipv6Status.String(), dualStack)
		envs = append(envs, ipv4PodCidr, ipv4ServiceCidr, ipv4Mask, ipv6PodCidr, ipv6ServiceCidr, ipv6Mask, stack)
	default:
		return "", fmt.Errorf("not supported ipType[%s]", ipType)
	}

	cni := getEnv(cniPlugin.String(), cls.GetClusterAdvanceSettings().GetNetworkType())
	apiServerHa := getEnv(apiserverHa.String(), strconv.FormatBool(cls.GetClusterAdvanceSettings().GetEnableHa()))

	envs = append(envs, cni, apiServerHa)

	return strings.Join(envs, ";"), nil
}
