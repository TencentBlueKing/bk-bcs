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

// Package tasks xxx
package tasks

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	k8scorev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/huawei/api"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/huawei/business"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/clusterops"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/encrypt"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

// RegisterClusterKubeConfigTask register cluster kubeConfig connection
func RegisterClusterKubeConfigTask(taskID string, stepName string) error {
	start := time.Now()
	// get task information and validate
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	if step == nil {
		return nil
	}

	// step login started here
	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	cloudID := step.Params[cloudprovider.CloudIDKey.String()]

	basicInfo, err := cloudprovider.GetClusterDependBasicInfo(cloudprovider.GetBasicInfoReq{
		ClusterID: clusterID,
		CloudID:   cloudID,
	})
	if err != nil {
		blog.Errorf("RegisterClusterKubeConfigTask[%s]: getClusterDependBasicInfo failed: %v", taskID, err)
		retErr := fmt.Errorf("getClusterDependBasicInfo failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	err = importClusterCredential(basicInfo)
	if err != nil {
		blog.Errorf("RegisterClusterKubeConfigTask[%s]: importClusterCredential failed: %v", taskID, err)
		retErr := fmt.Errorf("importClusterCredential failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// update step
	if err := state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("RegisterClusterKubeConfigTask[%s] task %s %s update to storage fatal", taskID, taskID, stepName)
		return err
	}

	return nil
}

func importClusterCredential(info *cloudprovider.CloudDependBasicInfo) error {
	if info.Cluster.KubeConfig == "" {
		cli, err := api.NewCceClient(info.CmOption)
		if err != nil {
			return err
		}

		kubeConfig, err := cli.GetClusterKubeConfig(info.Cluster.SystemID,
			info.Cluster.GetClusterAdvanceSettings().GetClusterConnectSetting().GetIsExtranet())
		info.Cluster.KubeConfig, _ = encrypt.Encrypt(nil, kubeConfig)

		err = cloudprovider.UpdateCluster(info.Cluster)
		if err != nil {
			return err
		}
	}

	configByte, _ := encrypt.Decrypt(nil, info.Cluster.KubeConfig)

	typesConfig := &types.Config{}
	err := json.Unmarshal([]byte(configByte), typesConfig)
	if err != nil {
		return fmt.Errorf("failed to unmarshal kubeconfig, %v", err)
	}
	err = cloudprovider.UpdateClusterCredentialByConfig(info.Cluster.ClusterID, typesConfig)
	if err != nil {
		return err
	}

	return nil
}

// ImportClusterNodesTask call gkeInterface or kubeConfig import cluster nodes
func ImportClusterNodesTask(taskID string, stepName string) error {
	start := time.Now()
	// get task information and validate
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	if step == nil {
		return nil
	}

	// step login started here
	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	cloudID := step.Params[cloudprovider.CloudIDKey.String()]

	basicInfo, err := cloudprovider.GetClusterDependBasicInfo(cloudprovider.GetBasicInfoReq{
		ClusterID: clusterID,
		CloudID:   cloudID,
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

	// update cluster masterNodes info
	err = cloudprovider.GetStorageModel().UpdateCluster(context.Background(), basicInfo.Cluster)
	if err != nil {
		blog.Errorf("ImportClusterNodesTask[%s] task %s %s update cluster fatal", taskID, taskID, stepName)
		return err
	}

	// update step
	if err = state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("ImportClusterNodesTask[%s] task %s %s update to storage fatal", taskID, taskID, stepName)
		return err
	}

	return nil
}

func importClusterInstances(info *cloudprovider.CloudDependBasicInfo) error {
	kubeConfigByte, err := encrypt.Decrypt(nil, info.Cluster.KubeConfig)
	if err != nil {
		return fmt.Errorf("decode kube config failed: %v", err)
	}

	kubeRet := base64.StdEncoding.EncodeToString([]byte(kubeConfigByte))
	kubeCli, err := clusterops.NewKubeClient(kubeRet)
	if err != nil {
		return fmt.Errorf("importClusterInstances NewKubeClient failed: %v", err)
	}

	nodes, err := kubeCli.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("list nodes failed, %s", err.Error())
	}

	err = importClusterNodesToCM(context.Background(), nodes.Items, info)
	if err != nil {
		return err
	}

	return nil
}

func importClusterNodesToCM(ctx context.Context, nodes []k8scorev1.Node,
	info *cloudprovider.CloudDependBasicInfo) error {
	// 获取zones

	ecsCli, err := api.NewEcsClient(info.CmOption)
	if err != nil {
		return err
	}

	zones, err := ecsCli.ListAvailabilityZones()
	if err != nil {
		return err
	}

	for _, n := range nodes {
		var (
			nodeZone string
			node     = &proto.Node{}
		)

		zone, ok := n.Labels[utils.ZoneKubernetesFlag]
		if ok {
			nodeZone = zone
		}
		zone, ok = n.Labels[utils.ZoneTopologyFlag]
		if ok && nodeZone == "" {
			nodeZone = zone
		}

		ipv4, ipv6 := utils.GetNodeIPAddress(&n)
		node.ZoneID = nodeZone
		node.InnerIP = utils.SliceToString(ipv4)
		node.InnerIPv6 = utils.SliceToString(ipv6)
		node.ClusterID = info.Cluster.ClusterID
		node.Status = common.StatusRunning
		node.NodeID = n.Spec.ProviderID
		node.NodeName = n.Name
		node.InstanceType = n.Labels[utils.NodeInstanceTypeFlag]

		if nodeZone != "" {
			for _, v := range zones {
				if v.ZoneName == nodeZone {
					node.ZoneName = fmt.Sprintf("可用区%d", func() int {
						return business.GetZoneNameByZoneId(info.Cluster.Region, v.ZoneName)
					}())
				}
			}
		}

		err := cloudprovider.GetStorageModel().CreateNode(ctx, node)
		if err != nil {
			blog.Errorf("ImportClusterNodesToCM CreateNode[%s] failed: %v", n.Name, err)
			continue
		}
	}

	return nil
}
