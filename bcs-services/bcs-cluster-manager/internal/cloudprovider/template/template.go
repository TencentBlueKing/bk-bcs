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
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
)

const (
	apiServer  = "apiServer"
	etcdServer = "etcdServer"

	createCluster = "create_cluster"
	addNodes = "add_nodes"

	// defaultPolicy default cpu_manager policy
	defaultPolicy = "none"
	staticPolicy  = "static"

	// bk-sops template vars prefix
	prefix = "CM"
	template = "template"
)

var (
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
	addNodesExtraEnv      = "CM.cluster.AddNodesExtraEnv"

	nodePasswd           = "CM.node.NodePasswd"
	nodeCPUManagerPolicy = "CM.node.NodeCPUManagerPolicy"
	nodeIPList           = "CM.node.NodeIPList"
	nodeOperator         = "CM.node.NodeOperator"

	templateBusinessID = "CM.template.BusinessID"
	templateOperator   = "CM.template.Operator"
)

// using task commonName inject dynamic parameters when processing
var (
	// DynamicParameterInject inject parameter for bk-sops
	DynamicParameterInject = map[string]string{
		nodeIPList: "NodeIPList",
	}
)

// ExtraInfo extra template values
type ExtraInfo struct {
	InstancePasswd string
	NodeIPList     string
	NodeOperator   string
	BusinessID     string
	Operator       string
}

// GenerateBKopsStep generate common bk-sops step
func GenerateBKopsStep(taskName, stepName string, cls *proto.Cluster, plugin *proto.BKOpsPlugin, info ExtraInfo) (*proto.Step, error) {
	now := time.Now().Format(time.RFC3339)

	step := &proto.Step{
		Name:   stepName,
		System: plugin.System,
		Params: make(map[string]string),
		Retry:  0,
		Start:  now,
		Status: cloudprovider.TaskStatusNotStarted,
		// method name is registered name to taskServer
		TaskMethod: cloudprovider.BKSOPTask,
		TaskName:   "标准运维任务",
	}
	step.Params["url"] = plugin.Link

	constants := make(map[string]string)
	for k, v := range plugin.Params {
		if strings.HasPrefix(v, prefix) {
			tValue, err := getTemplateParameterByName(v, cls, info)
			if err != nil {
				blog.Errorf("%s GenerateBKopsStep failed: %v", taskName, err)
				return nil, err
			}
			if strings.Contains(v, template) {
				step.Params[k] = tValue
				continue
			}
			constants[fmt.Sprintf("${%s}", k)] = tValue
			continue
		}
		step.Params[k] = v
	}
	constantsbyte, err := json.Marshal(&constants)
	if err != nil {
		blog.Errorf("%s GenerateBKopsStep failed: %v", taskName, err)
		return nil, err
	}
	step.Params["constants"] = string(constantsbyte)

	return step, nil
}

func getTemplateParameterByName(name string, cluster *proto.Cluster, extra ExtraInfo) (string, error) {
	if cluster == nil {
		errMsg := fmt.Errorf("cluster is empty when getTemplateParameterByName")
		blog.Errorf(errMsg.Error())
		return "", errMsg
	}

	switch name {
	case clusterID:
		return cluster.GetClusterID(), nil
	case clusterMasterIPs:
		return getClusterMasterIPs(cluster), nil
	case clusterMasterDomain:
		masterDomain := getMasterDomain(cluster)
		if len(masterDomain) == 0 {
			return "", fmt.Errorf("cluster %s masterDomain empty", cluster.GetExtraClusterID())
		}
		return masterDomain, nil
	case clusterEtcdDomain:
		etcdDomain := getEtcdDomain(cluster)
		if len(etcdDomain) == 0 {
			return "", fmt.Errorf("cluster %s etcdDomain empty", cluster.GetExtraClusterID())
		}
		return etcdDomain, nil
	case clusterRegion:
		return cluster.GetRegion(), nil
	case clusterVPC:
		return cluster.GetVpcID(), nil
	case clusterNetworkType:
		return cluster.GetNetworkType(), nil
	case clusterBizID:
		return cluster.GetBusinessID(), nil
	case clusterModuleID:
		return cluster.GetModuleID(), nil
	case clusterExtraID:
		return getClusterType(cluster), nil
	case clusterExtraClusterID:
		return cluster.GetExtraClusterID(), nil
	case clusterProjectID:
		return cluster.GetProjectID(), nil
	case nodePasswd:
		return extra.InstancePasswd, nil
	case nodeCPUManagerPolicy:
		return defaultPolicy, nil
	// dynamic parameter
	case nodeIPList:
		if len(extra.NodeIPList) == 0 {
			return nodeIPList, nil
		}
		return extra.NodeIPList, nil
	case nodeOperator:
		return extra.NodeOperator, nil
	case templateBusinessID:
		return extra.BusinessID, nil
	case templateOperator:
		return extra.Operator, nil
	case clusterExtraEnv:
		return getClusterCreateExtraEnv(cluster), nil
	case addNodesExtraEnv:
		return getAddNodesExtraEnv(cluster), nil
	default:
	}

	return "", fmt.Errorf("getTemplateParameterByName unSupportType %s", name)
}
