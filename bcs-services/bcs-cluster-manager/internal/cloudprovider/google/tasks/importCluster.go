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
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	k8scorev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/google/api"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/clusterops"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/encrypt"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

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
	_ = cloudprovider.GetStorageModel().UpdateCluster(context.Background(), basicInfo.Cluster)

	// update step
	if err = state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("ImportClusterNodesTask[%s] task %s %s update to storage fatal", taskID, taskID, stepName)
		return err
	}
	return nil
}

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

	ctx := cloudprovider.WithTaskIDForContext(context.Background(), taskID)

	err = importClusterCredential(ctx, basicInfo)
	if err != nil {
		blog.Errorf("RegisterClusterKubeConfigTask[%s]: importClusterCredential failed: %v", taskID, err)
		retErr := fmt.Errorf("importClusterCredential failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// update step
	if err = state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("RegisterClusterKubeConfigTask[%s] task %s %s update to storage fatal", taskID, taskID, stepName)
		return err
	}
	return nil
}

func importClusterCredential(ctx context.Context, data *cloudprovider.CloudDependBasicInfo) error { // nolint
	if data.Cluster.KubeConfig == "" {
		// gke集群 region级别 zone级别
		clusterType := common.Regions
		if len(strings.Split(data.Cluster.Region, "-")) == 3 {
			clusterType = common.Zones
		}
		kubeConfig, err := api.GetClusterKubeConfig(context.Background(), data.CmOption.Account.ServiceAccountSecret,
			data.CmOption.Account.GkeProjectID, data.Cluster.Region, clusterType, data.Cluster.SystemID)
		if err != nil {
			return fmt.Errorf("SyncClusterCloudInfo GetClusterKubeConfig failed: %v", err)
		}

		data.Cluster.KubeConfig = kubeConfig
		err = cloudprovider.UpdateCluster(data.Cluster)
		if err != nil {
			return err
		}
	}
	configByte, err := encrypt.Decrypt(nil, data.Cluster.KubeConfig)
	if err != nil {
		return fmt.Errorf("failed to decode kubeconfig, %v", err)
	}
	typesConfig := &types.Config{}
	err = json.Unmarshal([]byte(configByte), typesConfig)
	if err != nil {
		return fmt.Errorf("failed to unmarshal kubeconfig, %v", err)
	}
	err = cloudprovider.UpdateClusterCredentialByConfig(data.Cluster.ClusterID, typesConfig)
	if err != nil {
		return err
	}

	return nil
}

func importClusterInstances(data *cloudprovider.CloudDependBasicInfo) error {
	config, _ := encrypt.Decrypt(nil, data.Cluster.KubeConfig)
	kubeRet := base64.StdEncoding.EncodeToString([]byte(config))

	kubeCli, err := clusterops.NewKubeClient(kubeRet)
	if err != nil {
		return fmt.Errorf("importClusterInstances NewKubeClient failed: %v", err)
	}

	nodes, err := kubeCli.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("list nodes failed, %s", err.Error())
	}

	// get container runtime info here due to GKE API is not support
	if len(nodes.Items) > 0 {
		crv := strings.Split(nodes.Items[0].Status.NodeInfo.ContainerRuntimeVersion, "://")
		if len(crv) == 2 {
			data.Cluster.ClusterAdvanceSettings = &proto.ClusterAdvanceSetting{
				ContainerRuntime: crv[0],
				RuntimeVersion:   crv[1],
			}
			err = cloudprovider.GetStorageModel().UpdateCluster(context.Background(), data.Cluster)
			if err != nil {
				blog.Errorf("importClusterInstances update cluster[%s] failed: %v", data.Cluster.ClusterName, err)
			}
		}
	}

	/*
		gceCli, err := api.NewComputeServiceClient(data.CmOption)
		if err != nil {
			return fmt.Errorf("get gce client failed, %s", err.Error())
		}

		err = importClusterNodesToCM(context.Background(), gceCli, nodes.Items, data.Cluster.ClusterID)
		if err != nil {
			return err
		}
	*/

	return nil
}

// ImportClusterNodesToCM writes cluster nodes to DB
func importClusterNodesToCM(
	ctx context.Context, gceCli *api.ComputeServiceClient, nodes []k8scorev1.Node, clusterID string) error {

	for _, v := range nodes {
		nodeZone := ""
		zone, ok := v.Labels[utils.ZoneKubernetesFlag]
		if ok {
			nodeZone = zone
		}
		zone, ok = v.Labels[utils.ZoneTopologyFlag]
		if ok && nodeZone == "" {
			nodeZone = zone
		}

		var (
			node = &proto.Node{}
		)

		instance, err := gceCli.GetInstance(ctx, nodeZone, v.Name)
		if err == nil {
			node = api.InstanceToNode(gceCli, instance)
		} else {
			blog.Errorf("ImportClusterNodesToCM failed: %v", err)
			node.Region = v.Labels[utils.RegionTopologyFlag]
			node.InstanceType = v.Labels[utils.NodeInstanceTypeFlag]
			node.NodeName = v.Labels[utils.NodeNameFlag]
		}

		ipv4, ipv6 := utils.GetNodeIPAddress(&v)
		node.ZoneName = nodeZone
		node.InnerIP = utils.SliceToString(ipv4)
		node.InnerIPv6 = utils.SliceToString(ipv6)
		node.ClusterID = clusterID
		node.Status = common.StatusRunning

		err = cloudprovider.GetStorageModel().CreateNode(ctx, node)
		if err != nil {
			blog.Errorf("ImportClusterNodesToCM CreateNode[%s] failed: %v", v.Name, err)
			continue
		}
	}

	return nil
}
