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

package tasks

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/Tencent/bk-bcs/bcs-common/common/modules"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/types"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/blueking/api"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/clusterops"
	icommon "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ImportClusterNodesTask call tkeInterface or kubeConfig import cluster nodes
func ImportClusterNodesTask(taskID string, stepName string) error {
	start := time.Now()
	// get task information and validate
	task, err := cloudprovider.GetStorageModel().GetTask(context.Background(), taskID)
	if err != nil {
		blog.Errorf("ImportClusterNodesTask[%s]: task %s get detail task information from storage failed, %s. "+
			"task retry", taskID, taskID, err.Error())
		return err
	}

	state := &cloudprovider.TaskState{Task: task, JobResult: cloudprovider.NewJobSyncResult(task)}
	if state.IsTerminated() {
		blog.Errorf("ImportClusterNodesTask[%s]: task %s is terminated, step %s skip", taskID, taskID, stepName)
		return fmt.Errorf("task %s terminated", taskID)
	}
	step, err := state.IsReadyToStep(stepName)
	if err != nil {
		blog.Errorf("ImportClusterNodesTask[%s]: task %s not turn to run step %s, err %s", taskID, taskID, stepName, err.Error())
		return err
	}
	// previous step successful when retry task
	if step == nil {
		blog.Infof("ImportClusterNodesTask[%s]: current step[%s] successful and skip", taskID, stepName)
		return nil
	}
	blog.Infof("ImportClusterNodesTask[%s]: task %s run step %s, system: %s, old state: %s, params %v",
		taskID, taskID, stepName, step.System, step.Status, step.Params)

	// step login started here
	clusterID := step.Params["ClusterID"]
	cloudID := step.Params["CloudID"]

	basicInfo, err := cloudprovider.GetClusterDependBasicInfo(clusterID, cloudID)
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

	// update cluster status
	cloudprovider.UpdateClusterStatus(clusterID, icommon.StatusRunning)
	// import cluster clustercreential
	err = importClusterCredential(basicInfo)
	if err != nil {
		blog.Errorf("ImportClusterNodesTask[%s]: importClusterCredential failed: %v", taskID, err)
	}

	// update step
	if err := state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("CreateClusterShieldAlarmTask[%s] task %s %s update to storage fatal", taskID, taskID, stepName)
		return err
	}
	return nil
}

func importClusterCredential(data *cloudprovider.CloudDependBasicInfo) error {
	kubeRet, err := base64.StdEncoding.DecodeString(data.Cluster.KubeConfig)
	if err != nil {
		return err
	}

	config, err := types.GetKubeConfigFromYAMLBody(false, types.YamlInput{
		YamlContent: string(kubeRet),
	})
	if err != nil {
		return err
	}

	// first import cluster need to auto generate clusterCredential info, subsequently kube-agent report to update
	// currently, bcs only support token auth, kubeConfigList length greater 0, get zeroth kubeConfig
	var (
		server = ""
		caCertData = ""
		token = ""
	)
	if len(config.Clusters) > 0 {
		server = config.Clusters[0].Cluster.Server
		caCertData = string(config.Clusters[0].Cluster.CertificateAuthorityData)
	}
	if len(config.AuthInfos) > 0 {
		token = config.AuthInfos[0].AuthInfo.Token
	}

	if server == "" || caCertData == "" || token == "" {
		return fmt.Errorf("importClusterCredential parse kubeConfig failed: %v", "[server|caCertData|token] null")
	}

	now := time.Now().Format(time.RFC3339)
	err = cloudprovider.GetStorageModel().PutClusterCredential(context.Background(), &proto.ClusterCredential{
		ServerKey:            data.Cluster.ClusterID,
		ClusterID:            data.Cluster.ClusterID,
		ClientModule:         modules.BCSModuleKubeagent,
		ServerAddress:        server,
		CaCertData:           caCertData,
		UserToken:            token,
		ConnectMode:          modules.BCSConnectModeDirect,
		CreateTime:           now,
		UpdateTime:           now,
	})
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

	// import cluster
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

func importClusterNodesToCM(ctx context.Context, ipList []string, opt *cloudprovider.ListNodesOption) error {
	nodeMgr := api.NodeManager{}
	nodes, err := nodeMgr.ListNodesByIP(ipList, &cloudprovider.ListNodesOption{
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
			err := cloudprovider.GetStorageModel().CreateNode(ctx, n)
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

func getNodeIP(node v1.Node) string {
	nodeIP := ""

	for _, address := range node.Status.Addresses {
		if address.Type == v1.NodeInternalIP {
			nodeIP = address.Address
		}
	}

	return nodeIP
}

func getClusterInstancesByKubeConfig(data *cloudprovider.CloudDependBasicInfo) ([]string, []string, error) {
	kubeCli, err := clusterops.NewKubeClient(data.Cluster.KubeConfig)
	if err != nil {
		return nil, nil, err
	}

	nodeList, err := kubeCli.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, nil, err
	}

	masterIPs, nodeIPs := make([]string, 0), make([]string, 0)
	for i := range nodeList.Items {
		ip := getNodeIP(nodeList.Items[i])
		_, ok := nodeList.Items[i].Labels[icommon.MasterRole]
		if ok {
			masterIPs = append(masterIPs, ip)
		} else {
			nodeIPs = append(nodeIPs, ip)
		}
	}

	blog.Infof("get cluster[%s] masterIPs[%v] nodeIPs[%v]", data.Cluster.ClusterID, masterIPs, nodeIPs)
	return masterIPs, nodeIPs, nil
}

func transInstanceIPToNodes(ipList []string, opt *cloudprovider.ListNodesOption) ([]*proto.Node, error) {
	nodeMgr := api.NodeManager{}
	nodes, err := nodeMgr.ListNodesByIP(ipList, &cloudprovider.ListNodesOption{
		Common: opt.Common,
	})
	if err != nil {
		return nil, err
	}

	return nodes, nil
}
