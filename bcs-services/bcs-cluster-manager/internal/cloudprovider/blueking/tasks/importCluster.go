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

package tasks

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/blueking/api"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/clusterops"
	icommon "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/encrypt"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

// ImportClusterNodesTask call tkeInterface or kubeConfig import cluster nodes
func ImportClusterNodesTask(taskID string, stepName string) error {
	start := time.Now()

	// get task and task current step
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	// previous step successful when retry task
	if step == nil {
		blog.Infof("UpdateCreateClusterDBInfoTask[%s]: current step[%s] successful and skip", taskID, stepName)
		return nil
	}
	blog.Infof("ImportClusterNodesTask[%s]: task %s run step %s, system: %s, old state: %s, params %v",
		taskID, taskID, stepName, step.System, step.Status, step.Params)

	// step login started here
	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	cloudID := step.Params[cloudprovider.CloudIDKey.String()]

	basicInfo, err := cloudprovider.GetClusterDependBasicInfo(cloudprovider.GetBasicInfoReq{
		ClusterID:   clusterID,
		CloudID:     cloudID,
		NodeGroupID: "",
	})
	if err != nil {
		blog.Errorf("ImportClusterNodesTask[%s]: getClusterDependBasicInfo failed: %v", taskID, err)
		retErr := fmt.Errorf("getClusterDependBasicInfo failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// import cluster instances
	err = importClusterInstances(basicInfo)
	if err != nil {
		blog.Errorf("ImportClusterNodesTask[%s]: importClusterInstances failed: %v", taskID, err)
		retErr := fmt.Errorf("importClusterInstances failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// update cluster info
	err = cloudprovider.GetStorageModel().UpdateCluster(context.Background(), basicInfo.Cluster)
	if err != nil {
		return err
	}
	// import cluster clusterCredential
	err = importClusterCredential(basicInfo)
	if err != nil {
		blog.Errorf("ImportClusterNodesTask[%s]: importClusterCredential failed: %v", taskID, err)
	}

	// update step
	if err := state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("CreateClusterShieldAlarmTask[%s] task %s %s update to storage fatal",
			taskID, taskID, stepName)
		return err
	}
	return nil
}

func importClusterCredential(data *cloudprovider.CloudDependBasicInfo) error {
	kubeRet, err := encrypt.Decrypt(nil, data.Cluster.KubeConfig)
	if err != nil {
		return err
	}

	config, err := types.GetKubeConfigFromYAMLBody(false, types.YamlInput{
		YamlContent: kubeRet,
	})
	if err != nil {
		return err
	}

	err = cloudprovider.UpdateClusterCredentialByConfig(data.Cluster.ClusterID, config)
	if err != nil {
		return err
	}

	return nil
}

func importClusterInstances(data *cloudprovider.CloudDependBasicInfo) error {
	masterIPs, nodeIPs, err := getClusterInstancesByKubeConfig(data)
	if err != nil {
		return err
	}

	// import cluster and update cluster status
	masterNodes := make(map[string]*proto.Node)
	nodes, err := transInstanceIPToNodes(masterIPs, &cloudprovider.ListNodesOption{
		Common: data.CmOption,
	})
	if err != nil {
		return err
	}
	for _, node := range nodes {
		node.Status = icommon.StatusRunning
		masterNodes[node.InnerIP] = node
	}
	data.Cluster.Master = masterNodes
	// data.Cluster.Status = icommon.StatusRunning

	err = importClusterNodesToCM(context.Background(), nodeIPs, &cloudprovider.ListNodesOption{
		Common:       data.CmOption,
		ClusterVPCID: data.Cluster.VpcID,
		ClusterID:    data.Cluster.ClusterID,
	})
	if err != nil {
		return err
	}

	return nil
}

func importClusterNodesToCM(ctx context.Context, ipList []types.NodeAddress, opt *cloudprovider.ListNodesOption) error {
	nodes, err := transInstanceIPToNodes(ipList, &cloudprovider.ListNodesOption{
		Common: opt.Common,
	})
	if err != nil {
		return err
	}

	for _, n := range nodes {
		node, err := cloudprovider.GetStorageModel().GetNodeByIP(ctx, n.InnerIP)
		if err != nil && !errors.Is(err, drivers.ErrTableRecordNotFound) {
			blog.Errorf("importClusterNodes GetNodeByIP[%s] failed: %v", n.InnerIP, err)
			// no import node when found err
			continue
		}

		if node == nil {
			n.ClusterID = opt.ClusterID
			n.Status = icommon.StatusRunning
			err = cloudprovider.GetStorageModel().CreateNode(ctx, n)
			if err != nil {
				blog.Errorf("importClusterNodes CreateNode[%s] failed: %v", n.InnerIP, err)
			}
			continue
		}
		err = cloudprovider.GetStorageModel().UpdateNode(ctx, n)
		if err != nil {
			blog.Errorf("importClusterNodes UpdateNode[%s] failed: %v", n.InnerIP, err)
		}
	}

	return nil
}

func getNodeIP(node v1.Node) types.NodeAddress {
	var nodeAddress types.NodeAddress
	nodeAddress.NodeName = node.Name

	for _, address := range node.Status.Addresses {
		if address.Type == v1.NodeInternalIP {
			switch {
			case strings.Contains(address.Address, ":"):
				nodeAddress.IPv6Address = address.Address
			default:
				nodeAddress.IPv4Address = address.Address
			}
		}
	}

	return nodeAddress
}

func getClusterInstancesByKubeConfig(data *cloudprovider.CloudDependBasicInfo) ([]types.NodeAddress,
	[]types.NodeAddress, error) {

	kubeConfig, err := encrypt.Decrypt(nil, data.Cluster.KubeConfig)
	if err != nil {
		return nil, nil, err
	}

	kubeCli, err := clusterops.NewKubeClient(base64.StdEncoding.EncodeToString([]byte(kubeConfig)))
	if err != nil {
		return nil, nil, err
	}
	nodeList, err := kubeCli.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, nil, err
	}

	masterIPs, nodeIPs := make([]types.NodeAddress, 0), make([]types.NodeAddress, 0)
	for i := range nodeList.Items {
		ip := getNodeIP(nodeList.Items[i])
		ok := utils.IsMasterNode(nodeList.Items[i].Labels)
		if ok {
			masterIPs = append(masterIPs, ip)
			continue
		}
		nodeIPs = append(nodeIPs, ip)
	}
	blog.Infof("get cluster[%s] masterIPs[%v] nodeIPs[%v]", data.Cluster.ClusterID, masterIPs, nodeIPs)
	return masterIPs, nodeIPs, nil
}

func transInstanceIPToNodes(ipList []types.NodeAddress, opt *cloudprovider.ListNodesOption) ([]*proto.Node, error) {
	var (
		ipAddressList = make([]string, 0)
		ipAddressMap  = make(map[string]types.NodeAddress, 0)
	)
	for _, ip := range ipList {
		ipAddressList = append(ipAddressList, ip.IPv4Address)
		ipAddressMap[ip.IPv4Address] = ip
	}

	nodeMgr := api.NodeManager{}
	nodes, err := nodeMgr.ListNodesByIP(ipAddressList, &cloudprovider.ListNodesOption{
		Common: opt.Common,
	})
	if err != nil {
		return nil, err
	}
	for i := range nodes {
		if address, ok := ipAddressMap[nodes[i].InnerIP]; ok {
			nodes[i].NodeName = address.NodeName
			nodes[i].InnerIPv6 = address.IPv6Address
		}
	}

	return nodes, nil
}
