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
	"fmt"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"google.golang.org/api/container/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/google/api"
	providerutils "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/clusterops"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/encrypt"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/loop"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

// CreateGKEClusterTask call google interface to create cluster
func CreateGKEClusterTask(taskID string, stepName string) error {
	start := time.Now()

	// get task and task current step
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	// previous step successful when retry task
	if step == nil {
		blog.Infof("CreateGKEClusterTask[%s]: current step[%s] successful and skip", taskID, stepName)
		return nil
	}
	blog.Infof("CreateGKEClusterTask[%s]: task %s run step %s, system: %s, old state: %s, params %v",
		taskID, taskID, stepName, step.System, step.Status, step.Params)

	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	cloudID := step.Params[cloudprovider.CloudIDKey.String()]
	nodeGroupIDs := step.Params[cloudprovider.NodeGroupIDKey.String()]

	// get dependent basic info
	dependInfo, err := cloudprovider.GetClusterDependBasicInfo(cloudprovider.GetBasicInfoReq{
		ClusterID: clusterID,
		CloudID:   cloudID,
	})
	if err != nil {
		blog.Errorf("CreateGKEClusterTask[%s]: GetClusterDependBasicInfo for cluster %s in task %s "+
			"step %s failed, %s", taskID, clusterID, taskID, stepName, err) // nolint
		retErr := fmt.Errorf("get cloud/project information failed, %s", err)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	nodeGroups := make([]*proto.NodeGroup, 0)
	for _, ngID := range strings.Split(nodeGroupIDs, ",") {
		if ngID == "" {
			continue
		}

		nodeGroup, errGet := actions.GetNodeGroupByGroupID(cloudprovider.GetStorageModel(), ngID)
		if errGet != nil {
			blog.Errorf("CreateGKEClusterTask[%s]: GetNodeGroupByGroupID for cluster %s in task %s "+
				"step %s failed, %s", taskID, clusterID, taskID, stepName, errGet)
			retErr := fmt.Errorf("get nodegroup information failed, %s", errGet)
			_ = state.UpdateStepFailure(start, stepName, retErr)
			return retErr
		}
		nodeGroups = append(nodeGroups, nodeGroup)
	}

	// inject taskID
	ctx := cloudprovider.WithTaskIDForContext(context.Background(), taskID)

	// create cluster task
	clsId, err := createGKECluster(ctx, dependInfo, nodeGroups)
	if err != nil {
		blog.Errorf("CreateGKEClusterTask[%s] createGKECluster for cluster[%s] failed, %s",
			taskID, clusterID, err)
		retErr := fmt.Errorf("createGKECluster err, %s", err)
		_ = state.UpdateStepFailure(start, stepName, retErr)

		_ = cloudprovider.UpdateClusterErrMessage(clusterID, fmt.Sprintf("submit createCluster[%s] failed: %v",
			dependInfo.Cluster.GetClusterID(), err))
		return retErr
	}

	dependInfo.Cluster.SystemID = clsId

	err = cloudprovider.UpdateCluster(dependInfo.Cluster)
	if err != nil {
		blog.Errorf("createGKECluster[%s] update cluster systemID[%s] failed %s",
			taskID, dependInfo.Cluster.ClusterID, err)
		retErr := fmt.Errorf("call createGKECluster updateClusterSystemID[%s] api err: %s",
			dependInfo.Cluster.ClusterID, err)
		return retErr
	}

	// update response information to task common params
	if state.Task.CommonParams == nil {
		state.Task.CommonParams = make(map[string]string)
	}
	state.Task.CommonParams[cloudprovider.CloudSystemID.String()] = clsId

	// update step
	if err = state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("CreateGKEClusterTask[%s] task %s %s update to storage fatal", taskID, taskID, stepName)
		return err
	}

	return nil
}

func createGKECluster(ctx context.Context, info *cloudprovider.CloudDependBasicInfo, groups []*proto.NodeGroup) (
	string, error) {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	client, err := api.NewContainerServiceClient(info.CmOption)
	if err != nil {
		return "", fmt.Errorf("create GkeService failed")
	}

	clusterName := strings.ToLower(info.Cluster.ClusterID)
	cluster, err := client.GetCluster(context.Background(), clusterName)
	if err != nil {
		if !strings.Contains(err.Error(), "Not found") && !strings.Contains(err.Error(), "notFound") {
			return "", fmt.Errorf("createGKECluster[%s] get cluster failed, %v", taskID, err)
		}
	}

	if cluster != nil && cluster.Name != "" {
		return clusterName, nil
	}

	_, err = client.CreateCluster(context.Background(), generateCreateClusterRequest(info, groups))
	if err != nil {
		return "", fmt.Errorf("createGKECluster[%s] create cluster failed, %v", taskID, err)
	}

	// gke集群 region级别 zone级别
	clusterType := common.Regions
	if len(strings.Split(info.Cluster.Region, "-")) == 3 {
		clusterType = common.Zones
	}

	info.Cluster.ExtraInfo = make(map[string]string, 0)
	info.Cluster.ExtraInfo[api.GKEClusterLocationType] = clusterType

	blog.Infof("createGKECluster[%s] call createGKECluster UpdateClusterSystemID successful", taskID)

	return clusterName, nil
}

func generateCreateClusterRequest(info *cloudprovider.CloudDependBasicInfo,
	groups []*proto.NodeGroup) *container.CreateClusterRequest {
	req := &container.CreateClusterRequest{
		Cluster: &container.Cluster{
			Name:                  strings.ToLower(info.Cluster.ClusterID),
			Description:           info.Cluster.GetDescription(),
			ResourceLabels:        info.Cluster.Labels,
			InitialClusterVersion: info.Cluster.ClusterBasicSettings.Version,
			EnableKubernetesAlpha: false,
			ClusterIpv4Cidr:       info.Cluster.NetworkSettings.ClusterIPv4CIDR,
			LoggingConfig: &container.LoggingConfig{
				ComponentConfig: &container.LoggingComponentConfig{
					EnableComponents: []string{
						"SYSTEM_COMPONENTS",
						"WORKLOADS",
						"APISERVER",
						"SCHEDULER",
						"CONTROLLER_MANAGER",
					},
				},
			},
			IpAllocationPolicy: &container.IPAllocationPolicy{
				ClusterIpv4CidrBlock:  info.Cluster.NetworkSettings.ClusterIPv4CIDR,
				CreateSubnetwork:      false,
				NodeIpv4CidrBlock:     "",
				ServicesIpv4CidrBlock: info.Cluster.NetworkSettings.ServiceIPv4CIDR,
				UseIpAliases:          true,
			},
			NetworkConfig: &container.NetworkConfig{
				DnsConfig: &container.DNSConfig{
					ClusterDns: "KUBE_DNS",
				},
			},
			AddonsConfig: &container.AddonsConfig{
				HttpLoadBalancing: &container.HttpLoadBalancing{
					Disabled: false, // 启用 HTTP 负载均衡
				},
				DnsCacheConfig: &container.DnsCacheConfig{
					Enabled: true,
				},
			},
			MasterAuthorizedNetworksConfig: &container.MasterAuthorizedNetworksConfig{
				Enabled: false,
			},
			NodePools:         []*container.NodePool{},
			Locations:         []string{},
			MaintenancePolicy: &container.MaintenancePolicy{},
		},
	}

	internet := info.Cluster.ClusterAdvanceSettings.ClusterConnectSetting.Internet
	if internet != nil && len(internet.PublicAccessCidrs) > 0 {
		// 开启白名单
		req.Cluster.MasterAuthorizedNetworksConfig.Enabled = true
		cidrBlock := make([]*container.CidrBlock, 0)
		for _, cidr := range internet.PublicAccessCidrs {
			cidrBlock = append(cidrBlock, &container.CidrBlock{
				CidrBlock: cidr,
			})
		}

		req.Cluster.MasterAuthorizedNetworksConfig.CidrBlocks = cidrBlock

		if info.Cluster.ClusterAdvanceSettings.ClusterConnectSetting.IsExtranet {
			req.Cluster.MasterAuthorizedNetworksConfig.GcpPublicCidrsAccessEnabled = true
		}
	}

	for _, template := range info.Cluster.Template {
		if template.GetNodeRole() == common.NodeRoleMaster {
			if len(template.Zone) != 0 {
				req.Cluster.Locations = append(req.Cluster.Locations, template.Zone)
			}
		}
	}

	req.Cluster.PrivateClusterConfig = &container.PrivateClusterConfig{
		// 是否将主服务器的内部 IP 地址用作集群终端节点 就是在控制台中选择了专用集群
		EnablePrivateEndpoint: !info.Cluster.ClusterAdvanceSettings.ClusterConnectSetting.IsExtranet,
		EnablePrivateNodes:    true,
		MasterGlobalAccessConfig: &container.PrivateClusterMasterGlobalAccessConfig{
			Enabled: true,
		},
	}

	if info.Cluster.ManageType == common.ClusterManageTypeManaged {
		req.Cluster.Autopilot = &container.Autopilot{
			Enabled: true,
		}
	} else {
		req.Cluster.NetworkPolicy = &container.NetworkPolicy{
			Enabled:  true,
			Provider: "CALICO",
		}

		req.Cluster.MonitoringConfig = &container.MonitoringConfig{
			AdvancedDatapathObservabilityConfig: &container.AdvancedDatapathObservabilityConfig{
				EnableMetrics: false,
				RelayMode:     "disabled",
			},
			ComponentConfig: &container.MonitoringComponentConfig{
				EnableComponents: []string{},
			},
			ManagedPrometheusConfig: &container.ManagedPrometheusConfig{
				Enabled: false,
			},
		}

		for _, ng := range groups {
			ng.CloudNodeGroupID = strings.ToLower(ng.NodeGroupID)
			nodePool := api.GenerateNodePool(generateCreateClusterNodePoolInput(ng, info.Cluster))
			req.Cluster.NodePools = append(req.Cluster.NodePools, nodePool)
		}
	}

	if info.Cluster.NetworkSettings != nil && info.Cluster.NetworkSettings.MaxNodePodNum > 0 {
		req.Cluster.DefaultMaxPodsConstraint = &container.MaxPodsConstraint{
			MaxPodsPerNode: int64(info.Cluster.NetworkSettings.MaxNodePodNum),
		}
	}

	if info.Cluster.VpcID != "" {
		req.Cluster.Network = info.Cluster.VpcID
	}
	if info.Cluster.ClusterAdvanceSettings.ClusterConnectSetting.SubnetId != "" {
		req.Cluster.Subnetwork = info.Cluster.ClusterAdvanceSettings.ClusterConnectSetting.SubnetId
	}

	return req
}

// generateCreateClusterNodePoolInput generate create node pool input
func generateCreateClusterNodePoolInput(group *proto.NodeGroup, cluster *proto.Cluster) *api.CreateNodePoolRequest {
	if group.NodeTemplate.MaxPodsPerNode == 0 {
		group.NodeTemplate.MaxPodsPerNode = 110
	}
	return &api.CreateNodePoolRequest{
		NodePool: &api.NodePool{
			// gke nodePool名称中不允许有大写字母
			Name:             group.CloudNodeGroupID,
			Config:           generateNodeConfig(group),
			InitialNodeCount: 0,
			Locations:        group.AutoScaling.Zones,
			MaxPodsConstraint: &api.MaxPodsConstraint{
				MaxPodsPerNode: int64(group.NodeTemplate.MaxPodsPerNode),
			},
			Autoscaling: &api.NodePoolAutoscaling{
				// 不开启谷歌云 CA 组件，因为需要部署 BCS 自己的 CA 组件
				Enabled: false,
			},
			Management: generateNodeManagement(group, cluster),
		},
	}
}

// CheckGKEClusterStatusTask check cluster create status
func CheckGKEClusterStatusTask(taskID string, stepName string) error {
	start := time.Now()
	// get task and task current step
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	// previous step successful when retry task
	if step == nil {
		blog.Infof("CheckGKEClusterStatusTask[%s]: current step[%s] successful and skip", taskID, stepName)
		return nil
	}
	blog.Infof("CheckGKEClusterStatusTask[%s]: task %s run step %s, system: %s, old state: %s, params %v",
		taskID, taskID, stepName, step.System, step.Status, step.Params)

	// step login started here
	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	cloudID := step.Params[cloudprovider.CloudIDKey.String()]

	dependInfo, err := cloudprovider.GetClusterDependBasicInfo(cloudprovider.GetBasicInfoReq{
		ClusterID: clusterID,
		CloudID:   cloudID,
	})
	if err != nil {
		blog.Errorf("CheckGKEClusterStatusTask[%s]: GetClusterDependBasicInfo for cluster %s in task %s "+
			"step %s failed, %s", taskID, clusterID, taskID, stepName, err)
		retErr := fmt.Errorf("get cloud/project information failed, %s", err)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// check cluster status
	ctx := cloudprovider.WithTaskIDForContext(context.Background(), taskID)
	err = checkClusterStatus(ctx, dependInfo)
	if err != nil {
		blog.Errorf("CheckGKEClusterStatusTask[%s] checkClusterStatus[%s] failed: %v",
			taskID, clusterID, err)
		retErr := fmt.Errorf("checkClusterStatus[%s] check status failed: %s", clusterID, err)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// update step
	if err = state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("CheckGKEClusterStatusTask[%s] task %s %s update to storage fatal",
			taskID, taskID, stepName)
		return err
	}

	return nil
}

// checkClusterStatus check cluster status
func checkClusterStatus(ctx context.Context, info *cloudprovider.CloudDependBasicInfo) error {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	client, err := api.NewContainerServiceClient(info.CmOption)
	if err != nil {
		blog.Errorf("checkClusterStatus[%s] get gke client failed: %s", taskID, err)
		retErr := fmt.Errorf("get cloud gke client err, %s", err)
		return retErr
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Minute)
	defer cancel()

	err = loop.LoopDoFunc(ctx, func() error {
		cluster, errGet := client.GetCluster(ctx, info.Cluster.SystemID)
		if errGet != nil {
			blog.Errorf("checkClusterStatus[%s] failed: %v", taskID, errGet)
			return nil
		}

		blog.Infof("checkClusterStatus[%s] cluster[%s] current status[%s]",
			taskID, info.Cluster.ClusterID, cluster.Status)

		switch cluster.Status {
		case api.ClusterStatusRunning:
			return loop.EndLoop
		case api.ClusterStatusError:
			return fmt.Errorf("cluster status is error: %s", cluster.StatusMessage)
		}

		return nil
	}, loop.LoopInterval(10*time.Second))
	if err != nil {
		blog.Errorf("checkClusterStatus[%s] cluster[%s] failed: %v", taskID, info.Cluster.ClusterID, err)
		return err
	}

	return nil
}

// CheckGKENodeGroupsStatusTask check cluster nodes status
func CheckGKENodeGroupsStatusTask(taskID string, stepName string) error {
	start := time.Now()
	// get task and task current step
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	// previous step successful when retry task
	if step == nil {
		blog.Infof("CheckGKENodeGroupsStatusTask[%s]: current step[%s] successful and skip", taskID, stepName)
		return nil
	}
	blog.Infof("CheckGKENodeGroupsStatusTask[%s]: task %s run step %s, system: %s, old state: %s, params %v",
		taskID, taskID, stepName, step.System, step.Status, step.Params)

	// step login started here
	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	cloudID := step.Params[cloudprovider.CloudIDKey.String()]
	nodeGroupIDs := cloudprovider.ParseNodeIpOrIdFromCommonMap(step.Params,
		cloudprovider.NodeGroupIDKey.String(), ",")
	systemID := state.Task.CommonParams[cloudprovider.CloudSystemID.String()]
	dependInfo, err := cloudprovider.GetClusterDependBasicInfo(cloudprovider.GetBasicInfoReq{
		ClusterID: clusterID,
		CloudID:   cloudID,
	})
	if err != nil {
		blog.Errorf("CheckGKENodeGroupsStatusTask[%s]: GetClusterDependBasicInfo for cluster %s in task %s "+
			"step %s failed, %s", taskID, clusterID, taskID, stepName, err)
		retErr := fmt.Errorf("get cloud/project information failed, %s", err)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// check cluster status
	ctx := cloudprovider.WithTaskIDForContext(context.Background(), taskID)
	addSuccessNodeGroups, addFailureNodeGroups, err := checkNodesGroupStatus(ctx, dependInfo, systemID, nodeGroupIDs)
	if err != nil {
		blog.Errorf("CheckGKENodeGroupsStatusTask[%s] checkNodesGroupStatus[%s] failed: %v",
			taskID, clusterID, err)
		retErr := fmt.Errorf("CheckGKENodeGroupsStatusTask[%s] check nodegroup status failed: %s", clusterID, err)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// update response information to task common params
	if state.Task.CommonParams == nil {
		state.Task.CommonParams = make(map[string]string)
	}
	if len(addFailureNodeGroups) > 0 {
		state.Task.CommonParams[cloudprovider.FailedNodeGroupIDsKey.String()] = strings.Join(addFailureNodeGroups, ",")
	}
	if len(addSuccessNodeGroups) == 0 {
		blog.Errorf("CheckGKENodeGroupsStatusTask[%s] nodegroups init failed", taskID)
		retErr := fmt.Errorf("节点池初始化失败, 请联系管理员")
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	state.Task.CommonParams[cloudprovider.SuccessNodeGroupIDsKey.String()] = strings.Join(addSuccessNodeGroups, ",")
	// update step
	if err = state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("CheckGKENodeGroupsStatusTask[%s] task %s %s update to storage fatal",
			taskID, taskID, stepName)
		return err
	}

	return nil
}

func checkNodesGroupStatus(ctx context.Context, info *cloudprovider.CloudDependBasicInfo,
	systemID string, nodeGroupIDs []string) ([]string, []string, error) {
	var (
		addSuccessNodeGroups = make([]string, 0)
		addFailureNodeGroups = make([]string, 0)
	)
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	client, err := api.NewContainerServiceClient(info.CmOption)
	if err != nil {
		blog.Errorf("checkClusterStatus[%s] get gke client failed: %s", taskID, err)
		retErr := fmt.Errorf("get cloud gke client err, %s", err)
		return nil, nil, retErr
	}

	nodeGroups := make([]*proto.NodeGroup, 0)
	for _, ngID := range nodeGroupIDs {
		nodeGroup, errGet := actions.GetNodeGroupByGroupID(cloudprovider.GetStorageModel(), ngID)
		if errGet != nil {
			return nil, nil, fmt.Errorf("checkNodesGroupStatus GetNodeGroupByGroupID failed, %s", errGet)
		}
		nodeGroups = append(nodeGroups, nodeGroup)

	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	// loop cluster status
	err = loop.LoopDoFunc(ctx, func() error {
		index := 0
		running, failure := make([]string, 0), make([]string, 0)
		for _, group := range nodeGroups {
			cloudNodeGroupID := strings.ToLower(group.NodeGroupID)
			np, errPool := client.GetClusterNodePool(context.Background(), systemID, cloudNodeGroupID)
			if errPool != nil {
				blog.Errorf("taskID[%s] GetClusterNodePool[%s/%s] failed: %v", taskID, systemID,
					cloudNodeGroupID, errPool)
				return nil
			}
			if np == nil {
				return nil
			}

			switch {
			case np.Status == api.NodeGroupStatusProvisioning:
				blog.Infof("taskID[%s] GetClusterNodePool[%s] still creating, status[%s]",
					taskID, cloudNodeGroupID, np.Status)
				return nil
			case np.Status == api.NodeGroupStatusRunning:
				if !utils.StringInSlice(group.NodeGroupID, running) {
					running = append(running, group.NodeGroupID)
				}
				index++
			case np.Status == api.NodeGroupStatusRunningWithError:
				if !utils.StringInSlice(group.NodeGroupID, failure) {
					failure = append(failure, group.NodeGroupID)
				}
				index++
			case np.Status == api.NodeGroupStatusError:
				if !utils.StringInSlice(group.NodeGroupID, failure) {
					failure = append(failure, group.NodeGroupID)
				}
				index++
			}
		}
		if index == len(nodeGroups) {
			addSuccessNodeGroups = running
			addFailureNodeGroups = failure
			return loop.EndLoop
		}

		return nil
	}, loop.LoopInterval(10*time.Second))
	if err != nil {
		blog.Errorf("checkNodesGroupStatus[%s] cluster[%s] failed: %v", taskID, info.Cluster.ClusterID, err)
		return nil, nil, err
	}

	blog.Infof("checkNodesGroupStatus[%s] success[%v] failure[%v]",
		taskID, addSuccessNodeGroups, addFailureNodeGroups)

	return addSuccessNodeGroups, addFailureNodeGroups, nil
}

// UpdateGKENodesGroupToDBTask update GKE nodepool
func UpdateGKENodesGroupToDBTask(taskID string, stepName string) error {
	start := time.Now()
	// get task and task current step
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	// previous step successful when retry task
	if step == nil {
		blog.Infof("UpdateGKENodesGroupToDBTask[%s]: current step[%s] successful and skip", taskID, stepName)
		return nil
	}
	blog.Infof("UpdateGKENodesGroupToDBTask[%s]: task %s run step %s, system: %s, old state: %s, params %v",
		taskID, taskID, stepName, step.System, step.Status, step.Params)

	// step login started here
	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	cloudID := step.Params[cloudprovider.CloudIDKey.String()]
	addSuccessNodeGroupIDs := cloudprovider.ParseNodeIpOrIdFromCommonMap(state.Task.CommonParams,
		cloudprovider.SuccessNodeGroupIDsKey.String(), ",")
	addFailedNodeGroupIDs := cloudprovider.ParseNodeIpOrIdFromCommonMap(state.Task.CommonParams,
		cloudprovider.FailedNodeGroupIDsKey.String(), ",")

	dependInfo, err := cloudprovider.GetClusterDependBasicInfo(cloudprovider.GetBasicInfoReq{
		ClusterID: clusterID,
		CloudID:   cloudID,
	})
	if err != nil {
		blog.Errorf("UpdateGKENodesGroupToDBTask[%s]: GetClusterDependBasicInfo for cluster %s in task %s "+
			"step %s failed, %s", taskID, clusterID, taskID, stepName, err)
		retErr := fmt.Errorf("get cloud/project information failed, %s", err)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// check cluster status
	ctx := cloudprovider.WithTaskIDForContext(context.Background(), taskID)
	err = updateNodeGroups(ctx, dependInfo, addFailedNodeGroupIDs, addSuccessNodeGroupIDs)
	if err != nil {
		blog.Errorf("UpdateGKENodesGroupToDBTask[%s] updateNodeGroups[%s] failed: %v",
			taskID, clusterID, err)
		retErr := fmt.Errorf("UpdateGKENodesGroupToDBTask[%s] update nodegroups failed, %s", clusterID, err)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// update step
	if err = state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("UpdateGKENodesGroupToDBTask[%s] task %s %s update to storage fatal",
			taskID, taskID, stepName)
		return err
	}

	return nil
}

func updateNodeGroups(ctx context.Context, info *cloudprovider.CloudDependBasicInfo,
	addFailedNodeGroupIDs, addSuccessNodeGroupIDs []string) error {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	if len(addFailedNodeGroupIDs) > 0 {
		for _, ngID := range addFailedNodeGroupIDs {
			err := cloudprovider.UpdateNodeGroupStatus(ngID, common.StatusCreateNodeGroupFailed)
			if err != nil {
				return fmt.Errorf("updateNodeGroups updateNodeGroupStatus[%s] failed, %v", ngID, err)
			}
		}
	}

	// get google cloud client
	client, err := api.NewGCPClientSet(info.CmOption)
	if err != nil {
		blog.Errorf("updateNodeGroups[%s]: get gcp client failed, %s", taskID, err)
		retErr := fmt.Errorf("get cloud as client err, %s", err)
		return retErr
	}
	containerCli := client.ContainerServiceClient
	computeCli := client.ComputeServiceClient

	for _, ngID := range addSuccessNodeGroupIDs {
		group, err := actions.GetNodeGroupByGroupID(cloudprovider.GetStorageModel(), ngID)
		if err != nil {
			return fmt.Errorf("updateNodeGroups GetNodeGroupByGroupID failed, %s", err)
		}

		desiredSize := group.AutoScaling.DesiredSize
		group.CloudNodeGroupID = strings.ToLower(group.NodeGroupID)
		np, errPool := containerCli.GetClusterNodePool(context.Background(),
			info.Cluster.SystemID, group.CloudNodeGroupID)
		if errPool != nil {
			blog.Errorf("taskID[%s] GetClusterNodePool[%s/%s] failed: %v", taskID, info.Cluster.SystemID,
				group.CloudNodeGroupID, errPool)
			return nil
		}

		// get instanceGroupManager
		newIt, igm, err := getIgmAndIt(computeCli, np, group, taskID)
		if err != nil {
			blog.Errorf("updateNodeGroups[%s]: getIgmAndIt failed: %v", taskID, err)
			retErr := fmt.Errorf("getIgmAndIt failed, %s", err)
			return retErr
		}

		group.Status = common.StatusRunning
		group = generateNodeGroupFromIgmAndIt(group, igm, newIt, info.CmOption)
		group.AutoScaling.DesiredSize = desiredSize
		// update node group cloud args id
		err = cloudprovider.GetStorageModel().UpdateNodeGroup(context.Background(), group)
		if err != nil {
			blog.Errorf("updateNodeGroups[%s]: updateNodeGroupCloudArgsID[%s] failed, %s", taskID, ngID, err)
			retErr := fmt.Errorf("call updateNodeGroups updateNodeGroupCloudArgsID[%s] api err, %s", ngID, err)
			return retErr
		}

		if desiredSize > 0 {
			mig, err := api.GetInstanceGroupManager(client.ComputeServiceClient, group.AutoScaling.AutoScalingID)
			if err != nil {
				blog.Errorf("applyInstanceMachines[%s] GetInstanceGroupManager[%s] failed: %v",
					taskID, group.AutoScaling.AutoScalingID, err)
				return err
			}
			instanceNames := generateInstanceName(mig.BaseInstanceName, uint64(desiredSize))
			_, err = api.CreateInstanceForGroupManager(client.ComputeServiceClient,
				group.AutoScaling.AutoScalingID, instanceNames)
			if err != nil {
				blog.Errorf("applyInstanceMachines[%s] CreateInstanceForGroupManager[%s : %+v] failed: %v",
					taskID, group.AutoScaling.AutoScalingID, instanceNames, err)
				return err
			}
		}
	}

	return nil
}

// CheckGKEClusterNodesStatusTask check cluster nodes status
func CheckGKEClusterNodesStatusTask(taskID string, stepName string) error {
	start := time.Now()
	// get task and task current step
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	// previous step successful when retry task
	if step == nil {
		blog.Infof("CheckGKEClusterNodesStatusTask[%s]: current step[%s] successful and skip", taskID, stepName)
		return nil
	}
	blog.Infof("CheckGKEClusterNodesStatusTask[%s]: task %s run step %s, system: %s, old state: %s, params %v",
		taskID, taskID, stepName, step.System, step.Status, step.Params)

	// step login started here
	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	cloudID := step.Params[cloudprovider.CloudIDKey.String()]
	nodeGroupIDs := cloudprovider.ParseNodeIpOrIdFromCommonMap(state.Task.CommonParams,
		cloudprovider.SuccessNodeGroupIDsKey.String(), ",")

	dependInfo, err := cloudprovider.GetClusterDependBasicInfo(cloudprovider.GetBasicInfoReq{
		ClusterID: clusterID,
		CloudID:   cloudID,
	})
	if err != nil {
		blog.Errorf("CheckGKEClusterNodesStatusTask[%s]: GetClusterDependBasicInfo for cluster %s in task %s "+
			"step %s failed, %s", taskID, clusterID, taskID, stepName, err)
		retErr := fmt.Errorf("get cloud/project information failed, %s", err)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// check cluster status
	ctx := cloudprovider.WithTaskIDForContext(context.Background(), taskID)
	addSuccessNodes, addFailureNodes, err := checkClusterNodesStatus(ctx, dependInfo, nodeGroupIDs)
	if err != nil {
		blog.Errorf("CheckGKEClusterNodesStatusTask[%s] checkClusterStatus[%s] failed: %v",
			taskID, clusterID, err)
		retErr := fmt.Errorf("checkClusterStatus[%s] check cluster nodes status failed: %s", clusterID, err)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// update response information to task common params
	if state.Task.CommonParams == nil {
		state.Task.CommonParams = make(map[string]string)
	}
	if len(addFailureNodes) > 0 {
		state.Task.CommonParams[cloudprovider.FailedClusterNodeIDsKey.String()] = strings.Join(addFailureNodes, ",")
	}
	if len(addSuccessNodes) == 0 {
		blog.Errorf("CheckGKEClusterNodesStatusTask[%s] nodes init failed", taskID)
		retErr := fmt.Errorf("节点初始化失败, 请联系管理员")
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	state.Task.CommonParams[cloudprovider.SuccessClusterNodeIDsKey.String()] = strings.Join(addSuccessNodes, ",")

	// update step
	if err = state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("CheckGKEClusterNodesStatusTask[%s] task %s %s update to storage fatal", taskID, taskID, stepName)
		return err
	}

	return nil
}

func checkClusterNodesStatus(ctx context.Context, info *cloudprovider.CloudDependBasicInfo,
	nodeGroupIDs []string) ([]string, []string, error) {
	var totalNodesNum uint32
	var addSuccessNodes, addFailureNodes = make([]string, 0), make([]string, 0)
	asIDs := make([]string, 0)
	taskID := cloudprovider.GetTaskIDFromContext(ctx)
	for _, ngID := range nodeGroupIDs {
		nodeGroup, err := actions.GetNodeGroupByGroupID(cloudprovider.GetStorageModel(), ngID)
		if err != nil {
			return nil, nil, fmt.Errorf("get nodegroup information failed, %s", err)
		}
		totalNodesNum += nodeGroup.AutoScaling.DesiredSize
		asIDs = append(asIDs, nodeGroup.AutoScaling.AutoScalingID)
	}

	cli, err := api.NewComputeServiceClient(info.CmOption)
	if err != nil {
		blog.Errorf("checkClusterNodesStatus[%s] get gke client failed: %s", taskID, err)
		return nil, nil, fmt.Errorf("get cloud gke client err, %s", err)
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	// loop cluster status
	err = loop.LoopDoFunc(ctx, func() error {
		index := 0
		running, failure := make([]string, 0), make([]string, 0)
		for _, id := range asIDs {
			igmInfo, errGet := api.GetGCEResourceInfo(id)
			if errGet != nil {
				return fmt.Errorf("checkClusterNodesStatus[%s] get igm info failed: %v", taskID, errGet)
			}

			instances, errGet := cli.ListInstanceGroupsInstances(ctx, igmInfo[3], igmInfo[(len(igmInfo)-1)])
			if errGet != nil {
				blog.Errorf("checkClusterNodesStatus[%s] failed: %v", taskID, errGet)
				return nil
			}

			blog.Infof("checkClusterNodesStatus[%s] AutoScalingID[%s], current instances %d ",
				taskID, id, len(instances))

			for _, instance := range instances {
				blog.Infof("checkClusterNodesStatus[%s] node[%s] state %s", taskID, instance.Instance, instance.Status)
				switch instance.Status {
				case api.InstanceStatusRunning:
					if !utils.StringInSlice(instance.Instance, running) {
						running = append(running, instance.Instance)
					}
					index++
				case api.InstanceStatusTerminated:
					if !utils.StringInSlice(instance.Instance, failure) {
						failure = append(failure, instance.Instance)
					}
					index++
				}
			}
		}

		if index == int(totalNodesNum) {
			addSuccessNodes = running
			addFailureNodes = failure
			return loop.EndLoop
		}

		return nil
	}, loop.LoopInterval(1*time.Second))
	// other error
	if err != nil {
		blog.Errorf("checkClusterNodesStatus[%s] ListNodes failed: %v", taskID, err)
		return nil, nil, err
	}
	blog.Infof("checkClusterNodesStatus[%s] success[%v] failure[%v]", taskID, addSuccessNodes, addFailureNodes)

	return addSuccessNodes, addFailureNodes, nil
}

// UpdateGKENodesToDBTask update GKE nodes
func UpdateGKENodesToDBTask(taskID string, stepName string) error {
	start := time.Now()
	// get task and task current step
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	// previous step successful when retry task
	if step == nil {
		blog.Infof("UpdateNodesToDBTask[%s]: current step[%s] successful and skip", taskID, stepName)
		return nil
	}
	blog.Infof("UpdateNodesToDBTask[%s]: task %s run step %s, system: %s, old state: %s, params %v",
		taskID, taskID, stepName, step.System, step.Status, step.Params)

	// step login started here
	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	cloudID := step.Params[cloudprovider.CloudIDKey.String()]
	nodeGroupIDs := cloudprovider.ParseNodeIpOrIdFromCommonMap(state.Task.CommonParams,
		cloudprovider.SuccessNodeGroupIDsKey.String(), ",")

	dependInfo, err := cloudprovider.GetClusterDependBasicInfo(cloudprovider.GetBasicInfoReq{
		ClusterID: clusterID,
		CloudID:   cloudID,
	})
	if err != nil {
		blog.Errorf("UpdateNodesToDBTask[%s]: GetClusterDependBasicInfo for cluster %s in task %s "+
			"step %s failed, %s", taskID, clusterID, taskID, stepName, err)
		retErr := fmt.Errorf("get cloud/project information failed, %s", err)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// check cluster status
	ctx := cloudprovider.WithTaskIDForContext(context.Background(), taskID)
	err = updateNodeToDB(ctx, state, dependInfo, nodeGroupIDs)
	if err != nil {
		blog.Errorf("UpdateNodesToDBTask[%s] checkNodesGroupStatus[%s] failed: %v",
			taskID, clusterID, err)
		retErr := fmt.Errorf("UpdateNodesToDBTask[%s] update node to db failed, %s", clusterID, err)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// sync clusterData to pass-cc
	providerutils.SyncClusterInfoToPassCC(taskID, dependInfo.Cluster)

	// sync cluster perms
	providerutils.AuthClusterResourceCreatorPerm(ctx, dependInfo.Cluster.ClusterID,
		dependInfo.Cluster.ClusterName, dependInfo.Cluster.Creator)

	// update step
	if err = state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("UpdateNodesToDBTask[%s] task %s %s update to storage fatal",
			taskID, taskID, stepName)
		return err
	}

	return nil
}

func updateNodeToDB(ctx context.Context, state *cloudprovider.TaskState, info *cloudprovider.CloudDependBasicInfo,
	nodeGroupIDs []string) error {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)
	cli, err := api.NewComputeServiceClient(info.CmOption)
	if err != nil {
		blog.Errorf("updateNodeToDB[%s] get gke client failed: %s", taskID, err)
		return fmt.Errorf("get cloud gke client err, %s", err)
	}

	addSuccessNodes := make([]string, 0)
	for _, ngID := range nodeGroupIDs {
		nodeGroup, errGet := actions.GetNodeGroupByGroupID(cloudprovider.GetStorageModel(), ngID)
		if errGet != nil {
			return fmt.Errorf("updateNodeToDB GetNodeGroupByGroupID information failed, %s", errGet)
		}

		igmInfo, errGet := api.GetGCEResourceInfo(nodeGroup.AutoScaling.AutoScalingID)
		if errGet != nil {
			return fmt.Errorf("updateNodeToDB[%s] get igm info failed: %v", taskID, errGet)
		}

		instances, errGet := cli.ListInstanceGroupsInstances(ctx, igmInfo[3], igmInfo[(len(igmInfo)-1)])
		if errGet != nil {
			blog.Errorf("updateNodeToDB[%s] failed: %v", taskID, errGet)
			return fmt.Errorf("updateNodeToDB ListInstanceGroupsInstances failed, %s", errGet)
		}

		for _, instance := range instances {
			inInfo, errGet := api.GetGCEResourceInfo(instance.Instance)
			if errGet != nil {
				return fmt.Errorf("updateNodeToDB get gce resource info[%s] failed, %v", instance.Instance, errGet)
			}

			ins, errGet := cli.GetInstance(ctx, inInfo[3], inInfo[len(inInfo)-1])
			if errGet != nil {
				return fmt.Errorf("updateNodeToDB get instance[%s] failed, %v", instance.Instance, errGet)
			}

			node := api.InstanceToNode(cli, ins)

			if ins.Status == api.InstanceStatusRunning {
				addSuccessNodes = append(addSuccessNodes, node.InnerIP)
			}

			node.NodeGroupID = nodeGroup.NodeGroupID
			node.ClusterID = info.Cluster.ClusterID
			errGet = cloudprovider.GetStorageModel().CreateNode(context.Background(), node)
			if errGet != nil {
				return fmt.Errorf("updateNodeToDB CreateNode[%s] failed, %v", node.NodeName, errGet)
			}
		}
	}

	state.Task.CommonParams[cloudprovider.DynamicNodeIPListKey.String()] = strings.Join(addSuccessNodes, ",")
	state.Task.CommonParams[cloudprovider.NodeIPsKey.String()] = strings.Join(addSuccessNodes, ",")
	state.Task.CommonParams[cloudprovider.NodeNamesKey.String()] = strings.Join(addSuccessNodes, ",")

	return nil
}

// RegisterGKEClusterKubeConfigTask register cluster kubeconfig
func RegisterGKEClusterKubeConfigTask(taskID string, stepName string) error { // nolint
	start := time.Now()

	// get task and task current step
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	// previous step successful when retry task
	if step == nil {
		blog.Infof("RegisterGKEClusterKubeConfigTask[%s]: current step[%s] successful and skip", taskID, stepName)
		return nil
	}
	blog.Infof("RegisterGKEClusterKubeConfigTask[%s] task %s run current step %s, system: %s, old state: %s, params %v",
		taskID, taskID, stepName, step.System, step.Status, step.Params)

	// inject taskID
	ctx := cloudprovider.WithTaskIDForContext(context.Background(), taskID)

	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	cloudID := step.Params[cloudprovider.CloudIDKey.String()]

	// handler logic
	dependInfo, err := cloudprovider.GetClusterDependBasicInfo(cloudprovider.GetBasicInfoReq{
		ClusterID: clusterID,
		CloudID:   cloudID,
	})
	if err != nil {
		blog.Errorf("RegisterGKEClusterKubeConfigTask[%s] GetClusterDependBasicInfo in task %s step %s failed, %s",
			taskID, taskID, stepName, err)
		retErr := fmt.Errorf("get cloud/project information failed, %s", err)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	err = importClusterCredential(ctx, dependInfo)
	if err != nil {
		blog.Errorf("RegisterGKEClusterKubeConfigTask[%s] importClusterCredential failed: %s", taskID, err)
		retErr := fmt.Errorf("importClusterCredential failed %s", err)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	config, _ := encrypt.Decrypt(nil, dependInfo.Cluster.KubeConfig)
	kubeRet := base64.StdEncoding.EncodeToString([]byte(config))

	kubeCli, err := clusterops.NewKubeClient(kubeRet)
	if err != nil {
		blog.Errorf("RegisterGKEClusterKubeConfigTask[%s] importClusterInstances NewKubeClient failed: %s", taskID, err)
		retErr := fmt.Errorf("importClusterInstances NewKubeClient failed %s", err)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	addSuccessNodes := state.Task.CommonParams[cloudprovider.SuccessClusterNodeIDsKey.String()]
	if len(addSuccessNodes) > 0 {
		var nodes *corev1.NodeList
		toCtx, cancel := context.WithTimeout(ctx, 10*time.Minute)
		defer cancel()
		err = loop.LoopDoFunc(toCtx, func() error {
			nodes, err = kubeCli.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
			if err != nil {
				return nil
			}

			if len(nodes.Items) > 0 {
				return loop.EndLoop
			}

			return nil
		}, loop.LoopInterval(2*time.Second))
		if err != nil {
			blog.Errorf("RegisterGKEClusterKubeConfigTask[%s] importClusterInstances list node failed: %s",
				taskID, err)
			retErr := fmt.Errorf("importClusterInstances list node failed %s", err)
			_ = state.UpdateStepFailure(start, stepName, retErr)
			return retErr
		}
		// get container runtime info here due to GKE API is not support
		if len(nodes.Items) > 0 {
			crv := strings.Split(nodes.Items[0].Status.NodeInfo.ContainerRuntimeVersion, "://")
			if len(crv) == 2 {
				dependInfo.Cluster.ClusterAdvanceSettings = &proto.ClusterAdvanceSetting{
					ContainerRuntime: crv[0],
					RuntimeVersion:   crv[1],
				}
				err = cloudprovider.GetStorageModel().UpdateCluster(context.Background(), dependInfo.Cluster)
				if err != nil {
					blog.Errorf("RegisterGKEClusterKubeConfigTask[%s] importClusterInstances update cluster[%s]"+
						"failed: %s", taskID, err)
					retErr := fmt.Errorf("importClusterInstances update cluster[%s] failed: %s",
						dependInfo.Cluster.ClusterID, err)
					_ = state.UpdateStepFailure(start, stepName, retErr)
				}
			}
		}
	}

	blog.Infof("RegisterGKEClusterKubeConfigTask[%s] importClusterCredential success", taskID)

	// update step
	if err = state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("RegisterGKEClusterKubeConfigTask[%s:%s] update to storage fatal", taskID, stepName)
		return err
	}

	return nil
}
