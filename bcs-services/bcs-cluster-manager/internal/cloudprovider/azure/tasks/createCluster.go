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
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/avast/retry-go"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/azure/api"
	providerutils "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/encrypt"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/loop"
)

// CreateAKSClusterTask call azure interface to create cluster
func CreateAKSClusterTask(taskID string, stepName string) error {
	start := time.Now()

	// get task and task current step
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	// previous step successful when retry task
	if step == nil {
		blog.Infof("CreateAKSClusterTask[%s]: current step[%s] successful and skip", taskID, stepName)
		return nil
	}
	blog.Infof("CreateAKSClusterTask[%s]: task %s run step %s, system: %s, old state: %s, params %v",
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
		blog.Errorf("CreateAKSClusterTask[%s]: GetClusterDependBasicInfo for cluster %s in task %s "+
			"step %s failed, %s", taskID, clusterID, taskID, stepName, err.Error()) // nolint
		retErr := fmt.Errorf("get cloud/project information failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	nodeGroups := make([]*proto.NodeGroup, 0)
	for _, ngID := range strings.Split(nodeGroupIDs, ",") {
		nodeGroup, errGet := actions.GetNodeGroupByGroupID(cloudprovider.GetStorageModel(), ngID)
		if errGet != nil {
			blog.Errorf("CreateAKSClusterTask[%s]: GetNodeGroupByGroupID for cluster %s in task %s "+
				"step %s failed, %s", taskID, clusterID, taskID, stepName, errGet.Error())
			retErr := fmt.Errorf("get nodegroup information failed, %s", errGet.Error())
			_ = state.UpdateStepFailure(start, stepName, retErr)
			return retErr
		}
		nodeGroups = append(nodeGroups, nodeGroup)
	}

	// inject taskID
	ctx := cloudprovider.WithTaskIDForContext(context.Background(), taskID)

	// create cluster task
	clsId, err := createAKSCluster(ctx, dependInfo, nodeGroups)
	if err != nil {
		blog.Errorf("CreateAKSClusterTask[%s] createAKSCluster for cluster[%s] failed, %s",
			taskID, clusterID, err.Error())
		retErr := fmt.Errorf("createAKSCluster err, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)

		_ = cloudprovider.UpdateClusterErrMessage(clusterID, fmt.Sprintf("submit createCluster[%s] failed: %v",
			dependInfo.Cluster.GetClusterID(), err))
		return retErr
	}

	// update response information to task common params
	if state.Task.CommonParams == nil {
		state.Task.CommonParams = make(map[string]string)
	}
	state.Task.CommonParams[cloudprovider.CloudSystemID.String()] = clsId

	// update step
	if err = state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("CreateAKSClusterTask[%s] task %s %s update to storage fatal", taskID, taskID, stepName)
		return err
	}
	return nil
}

func createAKSCluster(ctx context.Context, info *cloudprovider.CloudDependBasicInfo, groups []*proto.NodeGroup) (
	string, error) {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	client, err := api.NewAksServiceImplWithCommonOption(info.CmOption)
	if err != nil {
		return "", fmt.Errorf("create AksService failed")
	}

	rgName, ok := info.Cluster.ExtraInfo[common.ClusterResourceGroup]
	if !ok {
		return "", fmt.Errorf("createAKSCluster[%s] %s failed, empty clusterResourceGroup",
			taskID, info.Cluster.ClusterID)
	}

	req, err := generateCreateClusterRequest(info, groups)
	if err != nil {
		return "", fmt.Errorf("createAKSCluster[%s] generateCreateClusterRequest failed, %v", taskID, err)
	}

	aksCluster, err := client.CreateCluster(ctx, rgName, info.Cluster.ClusterName, *req)
	if err != nil {
		return "", fmt.Errorf("createAKSCluster[%s] CreateCluster failed, %v", taskID, err)
	}

	info.Cluster.SystemID = *aksCluster.Name
	info.Cluster.ExtraInfo[common.NodeResourceGroup] = *aksCluster.Properties.NodeResourceGroup
	info.Cluster.ExtraInfo[common.NetworkResourceGroup] = rgName

	err = cloudprovider.UpdateCluster(info.Cluster)
	if err != nil {
		blog.Errorf("createAKSCluster[%s] updateClusterSystemID[%s] failed %s",
			taskID, info.Cluster.ClusterID, err.Error())
		retErr := fmt.Errorf("call createAKSCluster updateClusterSystemID[%s] api err: %s",
			info.Cluster.ClusterID, err.Error())
		return "", retErr
	}
	blog.Infof("createAKSCluster[%s] call createAKSCluster UpdateClusterSystemID successful", taskID)

	return *aksCluster.Name, nil
}

func generateCreateClusterRequest(info *cloudprovider.CloudDependBasicInfo, groups []*proto.NodeGroup) (
	*armcontainerservice.ManagedCluster, error) {
	cluster := info.Cluster
	if cluster.NetworkSettings == nil {
		return nil, fmt.Errorf("generateCreateClusterRequest empty NetworkSettings for cluster %s", cluster.ClusterID)
	}

	var adminUserName, publicKey string
	agentPools := make([]*armcontainerservice.ManagedClusterAgentPoolProfile, 0)
	for _, ng := range groups {
		agentPool, err := genAgentPoolReq(ng, info.CmOption.Account.SubscriptionID,
			cluster.ExtraInfo[common.ClusterResourceGroup], cluster.NetworkSettings.MaxNodePodNum)
		if err != nil {
			return nil, fmt.Errorf("generateCreateClusterRequest genAgentPoolReq failed, %v", err)
		}
		agentPools = append(agentPools, agentPool)

		adminUserName = ng.LaunchTemplate.InitLoginUsername
		if ng.LaunchTemplate.KeyPair != nil {
			publicKey, _ = encrypt.Decrypt(nil, ng.LaunchTemplate.KeyPair.KeyPublic)
		}

		info.Cluster.VpcID = ng.AutoScaling.VpcID
		if len(ng.AutoScaling.SubnetIDs) == 0 {
			return nil, fmt.Errorf("generateCreateClusterRequest nodegroup[%s] subnetIDs is empty", ng.NodeGroupID)
		}
		info.Cluster.ClusterBasicSettings.SubnetID = ng.AutoScaling.SubnetIDs[0]
	}
	keys := make([]*armcontainerservice.SSHPublicKey, 0)
	keys = append(keys, &armcontainerservice.SSHPublicKey{KeyData: to.Ptr(publicKey)})

	req := &armcontainerservice.ManagedCluster{
		Location: to.Ptr(cluster.Region),
		Name:     to.Ptr(cluster.ClusterName),
		Tags: func() map[string]*string {
			tags := make(map[string]*string)
			for k, v := range cluster.ClusterBasicSettings.ClusterTags {
				tags[k] = to.Ptr(v)
			}
			return tags
		}(),
		Properties: &armcontainerservice.ManagedClusterProperties{
			AgentPoolProfiles: agentPools,
			KubernetesVersion: to.Ptr(cluster.ClusterBasicSettings.Version),
			LinuxProfile: &armcontainerservice.LinuxProfile{
				AdminUsername: to.Ptr(adminUserName),
				SSH: &armcontainerservice.SSHConfiguration{
					PublicKeys: keys,
				},
			},
			DNSPrefix: to.Ptr("111-dns"),
			NetworkProfile: &armcontainerservice.NetworkProfile{
				ServiceCidr:  to.Ptr(cluster.NetworkSettings.ServiceIPv4CIDR),
				DNSServiceIP: to.Ptr(genDNSServiceIP(cluster.NetworkSettings.ServiceIPv4CIDR)),
				ServiceCidrs: []*string{to.Ptr(cluster.NetworkSettings.ServiceIPv4CIDR)},
			},
			ServicePrincipalProfile: &armcontainerservice.ManagedClusterServicePrincipalProfile{
				ClientID: to.Ptr(info.CmOption.Account.ClientID),
				Secret:   to.Ptr(info.CmOption.Account.ClientSecret),
			},
		},
	}

	return req, nil
}

func genAgentPoolReq(ng *proto.NodeGroup, subscriptionID, rgName string, podNum uint32) (
	*armcontainerservice.ManagedClusterAgentPoolProfile, error) {
	if ng.LaunchTemplate == nil {
		return nil, fmt.Errorf("generateCreateClusterRequest empty LaunchTemplate for nodegroup %s", ng.Name)
	}

	subnets := ng.AutoScaling.SubnetIDs
	if len(ng.AutoScaling.VpcID) == 0 || len(subnets) == 0 {
		return nil, fmt.Errorf("generateCreateClusterRequest nodegroup[%s] vpcID or subnetID"+
			" can not be empty", ng.Name)
	}
	sysDiskSize, _ := strconv.Atoi(ng.LaunchTemplate.SystemDisk.DiskSize)
	agentPool := &armcontainerservice.ManagedClusterAgentPoolProfile{
		AvailabilityZones: func(zones []string) []*string {
			az := make([]*string, 0)
			for _, v := range zones {
				az = append(az, to.Ptr(v))
			}
			return az
		}(ng.AutoScaling.Zones),
		// 临时设置节点数量,方便后续更新VMSS
		Count: func() *int32 {
			if ng.NodeGroupType == common.CloudClusterNodeGroupTypeSystem {
				return to.Ptr(int32(1))
			}
			return to.Ptr(int32(0))
		}(),
		EnableNodePublicIP: to.Ptr(func(group *proto.NodeGroup) bool {
			if group.LaunchTemplate.InternetAccess != nil {
				return group.LaunchTemplate.InternetAccess.PublicIPAssigned
			}
			return false
		}(ng)),
		Mode:          to.Ptr(armcontainerservice.AgentPoolMode(ng.NodeGroupType)),
		MaxPods:       to.Ptr(int32(podNum)),
		Name:          to.Ptr(getCloudNodeGroupID(ng)),
		OSDiskSizeGB:  to.Ptr(int32(sysDiskSize)),
		OSType:        to.Ptr(armcontainerservice.OSTypeLinux),
		ScaleDownMode: to.Ptr(armcontainerservice.ScaleDownModeDelete),
		Type:          to.Ptr(armcontainerservice.AgentPoolTypeVirtualMachineScaleSets),
		VMSize:        to.Ptr(ng.LaunchTemplate.InstanceType),
		VnetSubnetID: to.Ptr(fmt.Sprintf(
			"/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Network/virtualNetworks/%s/subnets/%s",
			subscriptionID, rgName, ng.AutoScaling.VpcID, subnets[0])),
	}

	return agentPool, nil
}

// 使用cidr的第11个IP作为DNS IP
func genDNSServiceIP(cidr string) string {
	ip, _, _ := net.ParseCIDR(cidr)
	ip = incrementIP(ip, 10)
	return ip.String()
}

func incrementIP(ip net.IP, index int) net.IP {
	for i := 0; i < index; i++ {
		incremented := false
		for j := len(ip) - 1; j >= 0; j-- {
			ip[j]++
			if ip[j] > 0 {
				incremented = true
				break
			}
		}
		if !incremented {
			break
		}
	}

	return ip
}

// CheckAKSClusterStatusTask check cluster create status
func CheckAKSClusterStatusTask(taskID string, stepName string) error {
	start := time.Now()
	// get task and task current step
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	// previous step successful when retry task
	if step == nil {
		blog.Infof("CheckAKSClusterStatusTask[%s]: current step[%s] successful and skip", taskID, stepName)
		return nil
	}
	blog.Infof("CheckAKSClusterStatusTask[%s]: task %s run step %s, system: %s, old state: %s, params %v",
		taskID, taskID, stepName, step.System, step.Status, step.Params)

	// step login started here
	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	cloudID := step.Params[cloudprovider.CloudIDKey.String()]

	dependInfo, err := cloudprovider.GetClusterDependBasicInfo(cloudprovider.GetBasicInfoReq{
		ClusterID: clusterID,
		CloudID:   cloudID,
	})
	if err != nil {
		blog.Errorf("CheckAKSClusterStatusTask[%s]: GetClusterDependBasicInfo for cluster %s in task %s "+
			"step %s failed, %s", taskID, clusterID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("get cloud/project information failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// check cluster status
	ctx := cloudprovider.WithTaskIDForContext(context.Background(), taskID)
	err = checkClusterStatus(ctx, dependInfo)
	if err != nil {
		blog.Errorf("CheckAKSClusterStatusTask[%s] checkClusterStatus[%s] failed: %v",
			taskID, clusterID, err)
		retErr := fmt.Errorf("checkClusterStatus[%s] timeout|abnormal", clusterID)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// update step
	if err = state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("CheckAKSClusterStatusTask[%s] task %s %s update to storage fatal",
			taskID, taskID, stepName)
		return err
	}

	return nil
}

// checkClusterStatus check cluster status
func checkClusterStatus(ctx context.Context, info *cloudprovider.CloudDependBasicInfo) error {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	// get azureCloud client
	cli, err := api.NewAksServiceImplWithCommonOption(info.CmOption)
	if err != nil {
		blog.Errorf("checkClusterStatus[%s] get aks client failed: %s", taskID, err.Error())
		retErr := fmt.Errorf("get cloud aks client err, %s", err.Error())
		return retErr
	}

	var (
		failed = false
	)

	ctx, cancel := context.WithTimeout(ctx, 30*time.Minute)
	defer cancel()

	// loop cluster status
	err = loop.LoopDoFunc(ctx, func() error {
		cluster, errGet := cli.GetCluster(ctx, info, info.Cluster.ExtraInfo[common.ClusterResourceGroup])
		if errGet != nil {
			blog.Errorf("checkClusterStatus[%s] failed: %v", taskID, errGet)
			return nil
		}

		blog.Infof("checkClusterStatus[%s] cluster[%s] current status[%s]", taskID,
			info.Cluster.ClusterID, *cluster.Properties.ProvisioningState)

		switch *cluster.Properties.ProvisioningState {
		case api.ManagedClusterPodIdentityProvisioningStateSucceeded:
			return loop.EndLoop
		case api.ManagedClusterPodIdentityProvisioningStateFailed:
			failed = true
			return loop.EndLoop
		}

		return nil
	}, loop.LoopInterval(10*time.Second))
	if err != nil {
		blog.Errorf("checkClusterStatus[%s] cluster[%s] failed: %v", taskID, info.Cluster.ClusterID, err)
		return err
	}

	if failed {
		blog.Errorf("checkClusterStatus[%s] GetCluster[%s] failed: abnormal", taskID, info.Cluster.ClusterID)
		retErr := fmt.Errorf("cluster[%s] status abnormal", info.Cluster.ClusterID)
		return retErr
	}

	return nil
}

// CheckAKSNodeGroupsStatusTask check cluster nodes status
func CheckAKSNodeGroupsStatusTask(taskID string, stepName string) error {
	start := time.Now()
	// get task and task current step
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	// previous step successful when retry task
	if step == nil {
		blog.Infof("CheckAKSNodeGroupsStatusTask[%s]: current step[%s] successful and skip", taskID, stepName)
		return nil
	}
	blog.Infof("CheckAKSNodeGroupsStatusTask[%s]: task %s run step %s, system: %s, old state: %s, params %v",
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
		blog.Errorf("CheckAKSNodeGroupsStatusTask[%s]: GetClusterDependBasicInfo for cluster %s in task %s "+
			"step %s failed, %s", taskID, clusterID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("get cloud/project information failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// check cluster status
	ctx := cloudprovider.WithTaskIDForContext(context.Background(), taskID)
	addSuccessNodeGroups, addFailureNodeGroups, err := checkNodesGroupStatus(ctx, dependInfo, systemID, nodeGroupIDs)
	if err != nil {
		blog.Errorf("CheckAKSNodeGroupsStatusTask[%s] checkNodesGroupStatus[%s] failed: %v",
			taskID, clusterID, err)
		retErr := fmt.Errorf("CheckAKSNodeGroupsStatusTask[%s] timeout|abnormal", clusterID)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// update response information to task common params
	if state.Task.CommonParams == nil {
		state.Task.CommonParams = make(map[string]string)
	}
	if len(addFailureNodeGroups) > 0 {
		state.Task.CommonParams[cloudprovider.FailedNodeGroupIDsKey.String()] =
			strings.Join(addFailureNodeGroups, ",")
	}
	if len(addSuccessNodeGroups) == 0 {
		blog.Errorf("CheckAKSNodeGroupsStatusTask[%s] nodegroups init failed", taskID)
		retErr := fmt.Errorf("节点池初始化失败, 请联系管理员")
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	state.Task.CommonParams[cloudprovider.SuccessNodeGroupIDsKey.String()] =
		strings.Join(addSuccessNodeGroups, ",")
	// update step
	if err = state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("CheckAKSNodeGroupsStatusTask[%s] task %s %s update to storage fatal",
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

	// get azureCloud client
	cli, err := api.NewAksServiceImplWithCommonOption(info.CmOption)
	if err != nil {
		blog.Errorf("checkNodesGroupStatus[%s] get aks client failed: %s", taskID, err.Error())
		retErr := fmt.Errorf("get cloud aks client err, %s", err.Error())
		return nil, nil, retErr
	}
	nodeGroups := make([]*proto.NodeGroup, 0)
	for _, ngID := range nodeGroupIDs {
		nodeGroup, errGet := actions.GetNodeGroupByGroupID(cloudprovider.GetStorageModel(), ngID)
		if errGet != nil {
			return nil, nil, fmt.Errorf("checkNodesGroupStatus GetNodeGroupByGroupID failed, %s", errGet.Error())
		}
		nodeGroups = append(nodeGroups, nodeGroup)

	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	// loop cluster status
	err = loop.LoopDoFunc(ctx, func() error {
		index := 0
		running, failure := make([]string, 0), make([]string, 0)
		for _, ng := range nodeGroups {
			aksAgentPool, errQuery := cli.GetPoolAndReturn(ctx, info.Cluster.ExtraInfo[common.ClusterResourceGroup],
				systemID, getCloudNodeGroupID(ng))
			if errQuery != nil {
				blog.Errorf("checkNodesGroupStatus[%s] failed: %v", taskID, errQuery)
				return nil
			}
			if aksAgentPool == nil {
				blog.Errorf("checkNodesGroupStatus[%s] GetPoolAndReturn[%s] not found", taskID, ng.NodeGroupID)
				return nil
			}

			blog.Infof("checkNodesGroupStatus[%s] nodeGroup[%s] status %s",
				taskID, ng.NodeGroupID, *aksAgentPool.Properties.ProvisioningState)

			switch *aksAgentPool.Properties.ProvisioningState {
			case api.AgentPoolPodIdentityProvisioningStateSucceeded:
				running = append(running, ng.NodeGroupID)
				index++
			case api.AgentPoolPodIdentityProvisioningStateFailed:
				failure = append(failure, ng.NodeGroupID)
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

// UpdateAKSNodesGroupToDBTask update AKS nodepool
func UpdateAKSNodesGroupToDBTask(taskID string, stepName string) error {
	start := time.Now()
	// get task and task current step
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	// previous step successful when retry task
	if step == nil {
		blog.Infof("UpdateAKSNodesGroupToDBTask[%s]: current step[%s] successful and skip", taskID, stepName)
		return nil
	}
	blog.Infof("UpdateAKSNodesGroupToDBTask[%s]: task %s run step %s, system: %s, old state: %s, params %v",
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
		blog.Errorf("UpdateAKSNodesGroupToDBTask[%s]: GetClusterDependBasicInfo for cluster %s in task %s "+
			"step %s failed, %s", taskID, clusterID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("get cloud/project information failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// check cluster status
	ctx := cloudprovider.WithTaskIDForContext(context.Background(), taskID)
	err = updateNodeGroups(ctx, dependInfo, addFailedNodeGroupIDs, addSuccessNodeGroupIDs)
	if err != nil {
		blog.Errorf("UpdateAKSNodesGroupToDBTask[%s] updateNodeGroups[%s] failed: %v",
			taskID, clusterID, err)
		retErr := fmt.Errorf("UpdateAKSNodesGroupToDBTask[%s] timeout|abnormal", clusterID)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// update step
	if err = state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("UpdateAKSNodesGroupToDBTask[%s] task %s %s update to storage fatal",
			taskID, taskID, stepName)
		return err
	}

	return nil
}

func updateNodeGroups(ctx context.Context, info *cloudprovider.CloudDependBasicInfo,
	addFailedNodeGroupIDs, addSuccessNodeGroupIDs []string) error {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)
	nodeResourceGroup := info.Cluster.ExtraInfo[common.NodeResourceGroup]
	clusterResourceGroup := info.Cluster.ExtraInfo[common.ClusterResourceGroup]

	if len(addFailedNodeGroupIDs) > 0 {
		for _, ngID := range addFailedNodeGroupIDs {
			err := cloudprovider.UpdateNodeGroupStatus(ngID, common.StatusCreateNodeGroupFailed)
			if err != nil {
				return fmt.Errorf("updateNodeGroups updateNodeGroupStatus[%s] failed, %v", ngID, err)
			}
		}
	}

	for _, ngID := range addSuccessNodeGroupIDs {
		nodeGroup, err := actions.GetNodeGroupByGroupID(cloudprovider.GetStorageModel(), ngID)
		if err != nil {
			return fmt.Errorf("updateNodeGroups GetNodeGroupByGroupID failed, %s", err.Error())
		}

		// get azureCloud client
		cli, err := api.NewAksServiceImplWithCommonOption(info.CmOption)
		if err != nil {
			blog.Errorf("updateNodeGroups[%s] get aks client failed: %s", taskID, err.Error())
			return fmt.Errorf("get cloud aks client err, %s", err.Error())
		}

		aksAgentPool, err := cli.GetPoolAndReturn(ctx, cloudprovider.GetClusterResourceGroup(info.Cluster),
			info.Cluster.SystemID, getCloudNodeGroupID(nodeGroup))
		if err != nil {
			blog.Errorf("updateNodeGroups[%s] GetPoolAndReturn failed: %v", taskID, err)
			return fmt.Errorf("updateNodeGroups GetPoolAndReturn[%s] failed, %s",
				nodeGroup.NodeGroupID, err.Error())
		}

		err = processVmss(ctx, cli, aksAgentPool, nodeGroup, nodeResourceGroup, clusterResourceGroup)
		if err != nil {
			return fmt.Errorf("updateNodeGroups processVmss[%s] failed, %s", nodeGroup.NodeGroupID, err.Error())
		}

		nodeGroup.CloudNodeGroupID = *aksAgentPool.Name
		nodeGroup.Status = common.StatusRunning

		err = cloudprovider.GetStorageModel().UpdateNodeGroup(context.Background(), nodeGroup)
		if err != nil {
			return fmt.Errorf("updateNodeGroups UpdateNodeGroup[%s] failed, %s",
				nodeGroup.NodeGroupID, err.Error())
		}
	}

	return nil
}

func processVmss(ctx context.Context, cli api.AksService, pool *armcontainerservice.AgentPool,
	nodeGroup *proto.NodeGroup, rg, crg string) error {
	set, err := cli.MatchNodeGroup(ctx, rg, *pool.Name)
	if err != nil {
		return fmt.Errorf("processVmss call MatchNodeGroup[%s] falied, %v", nodeGroup.NodeGroupID, err)
	}

	vmSet, err := scaleSystemVmss(ctx, cli, set, nodeGroup, rg)
	if err != nil {
		return fmt.Errorf("processVmss scaleSystemVmss[%s] failed, %s", nodeGroup.NodeGroupID, err.Error())
	}

	vmSet.SKU.Capacity = to.Ptr(int64(nodeGroup.AutoScaling.DesiredSize))

	// 字段对齐
	_ = cli.AgentPoolToNodeGroup(pool, nodeGroup)

	finalVmss, err := updateVmss(ctx, cli, nodeGroup, vmSet, rg, crg)
	if err != nil {
		return fmt.Errorf("processVmss updateVmss[%s] failed, %s", nodeGroup.NodeGroupID, err.Error())
	}

	_ = cli.SetToNodeGroup(finalVmss, nodeGroup)

	nodeGroup.AutoScaling.DesiredSize = uint32(*finalVmss.SKU.Capacity)

	return nil
}

// 由于创建系统池时,节点数不能为0, 此处先删除创建的初始节点, 待更新vmss后重新创建
func scaleSystemVmss(rootCtx context.Context, cli api.AksService, set *armcompute.VirtualMachineScaleSet,
	group *proto.NodeGroup, rg string) (*armcompute.VirtualMachineScaleSet, error) {
	if group.NodeGroupType == common.CloudClusterNodeGroupTypeSystem {
		set.SKU.Capacity = to.Ptr(int64(0))
		ctx, cancel := context.WithTimeout(rootCtx, 5*time.Minute)
		defer cancel()

		api.SetImageReferenceNull(set)
		vmSet, err := cli.UpdateSetWithName(ctx, set, rg, *set.Name)
		if err != nil {
			return nil, fmt.Errorf("scaleSystemVmss call UpdateSetWithName[%s][%s] failed, %v", rg, *set.Name, err)
		}

		return vmSet, nil
	}

	return set, nil
}

// CheckAKSClusterNodesStatusTask check cluster nodes status
func CheckAKSClusterNodesStatusTask(taskID string, stepName string) error {
	start := time.Now()
	// get task and task current step
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	// previous step successful when retry task
	if step == nil {
		blog.Infof("CheckAKSClusterNodesStatusTask[%s]: current step[%s] successful and skip", taskID, stepName)
		return nil
	}
	blog.Infof("CheckAKSClusterNodesStatusTask[%s]: task %s run step %s, system: %s, old state: %s, params %v",
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
		blog.Errorf("CheckAKSClusterNodesStatusTask[%s]: GetClusterDependBasicInfo for cluster %s in task %s "+
			"step %s failed, %s", taskID, clusterID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("get cloud/project information failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// check cluster status
	ctx := cloudprovider.WithTaskIDForContext(context.Background(), taskID)
	addSuccessNodes, addFailureNodes, err := checkClusterNodesStatus(ctx, dependInfo, nodeGroupIDs)
	if err != nil {
		blog.Errorf("CheckAKSClusterStatusTask[%s] checkClusterStatus[%s] failed: %v",
			taskID, clusterID, err)
		retErr := fmt.Errorf("checkClusterStatus[%s] timeout|abnormal", clusterID)
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
		blog.Errorf("CheckCreateClusterNodeStatusTask[%s] nodes init failed", taskID)
		retErr := fmt.Errorf("节点初始化失败, 请联系管理员")
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	state.Task.CommonParams[cloudprovider.SuccessClusterNodeIDsKey.String()] = strings.Join(addSuccessNodes, ",")

	// update step
	if err = state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("CheckCreateClusterNodeStatusTask[%s] task %s %s update to storage fatal", taskID, taskID, stepName)
		return err
	}

	return nil
}

func checkClusterNodesStatus(ctx context.Context, info *cloudprovider.CloudDependBasicInfo, // nolint
	nodeGroupIDs []string) ([]string, []string, error) {
	var totalNodesNum uint32
	var addSuccessNodes, addFailureNodes = make([]string, 0), make([]string, 0)
	nodePoolList := make([]string, 0)
	poolVmss := make(map[string]string)
	taskID := cloudprovider.GetTaskIDFromContext(ctx)
	for _, ngID := range nodeGroupIDs {
		nodeGroup, err := actions.GetNodeGroupByGroupID(cloudprovider.GetStorageModel(), ngID)
		if err != nil {
			return nil, nil, fmt.Errorf("get nodegroup information failed, %s", err.Error())
		}
		totalNodesNum += nodeGroup.AutoScaling.DesiredSize
		nodePoolList = append(nodePoolList, nodeGroup.CloudNodeGroupID)
		poolVmss[nodeGroup.CloudNodeGroupID] = nodeGroup.AutoScaling.AutoScalingID
	}

	// get azureCloud client
	cli, err := api.NewAksServiceImplWithCommonOption(info.CmOption)
	if err != nil {
		blog.Errorf("checkClusterNodesStatus[%s] get aks client failed: %s", taskID, err.Error())
		return nil, nil, fmt.Errorf("get cloud aks client err, %s", err.Error())
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	// loop cluster status
	err = loop.LoopDoFunc(ctx, func() error {
		index := 0
		running, failure := make([]string, 0), make([]string, 0)
		for _, pool := range nodePoolList {
			instances, errGet := cli.ListInstanceAndReturn(ctx, info.Cluster.ExtraInfo[common.NodeResourceGroup],
				poolVmss[pool])
			if errGet != nil {
				blog.Errorf("checkClusterNodesStatus[%s] failed: %v", taskID, errGet)
				return nil
			}

			blog.Infof("checkClusterNodesStatus[%s] nodeGroup[%s], current instances %d ",
				taskID, pool, len(instances))

			for _, instance := range instances {
				blog.Infof("checkClusterNodesStatus[%s] node[%s] state %s", taskID, *instance.Name,
					*instance.Properties.ProvisioningState)
				switch *instance.Properties.ProvisioningState {
				case api.VMProvisioningStateSucceeded:
					index++
					running = append(running, *instance.Name)
				case api.VMProvisioningStateFailed:
					failure = append(failure, *instance.Name)
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
	}, loop.LoopInterval(10*time.Second))
	// other error
	if err != nil && !errors.Is(err, context.DeadlineExceeded) {
		blog.Errorf("checkClusterNodesStatus[%s] ListNodes failed: %v", taskID, err)
		return nil, nil, err
	}
	// timeout error
	if errors.Is(err, context.DeadlineExceeded) {
		blog.Errorf("checkClusterNodesStatus[%s] ListNodes failed: %v", taskID, err)

		running, failure := make([]string, 0), make([]string, 0)
		for _, pool := range nodePoolList {
			instances, errGet := cli.ListInstanceAndReturn(ctx, info.Cluster.ExtraInfo[common.NodeResourceGroup],
				fmt.Sprintf("%s-%s", pool, "vmss"))
			if errGet != nil {
				blog.Errorf("checkClusterNodesStatus[%s] failed: %v", taskID, errGet)
				return nil, nil, errGet
			}

			for _, instance := range instances {
				blog.Infof("checkClusterNodesStatus[%s] node[%s] state %s", taskID, *instance.Name,
					*instance.Properties.ProvisioningState)
				switch *instance.Properties.ProvisioningState {
				case api.VMProvisioningStateSucceeded:
					running = append(running, *instance.Name)
				default:
					failure = append(failure, *instance.Name)
				}
			}
		}

		addSuccessNodes = running
		addFailureNodes = failure
	}
	blog.Infof("checkClusterNodesStatus[%s] success[%v] failure[%v]", taskID, addSuccessNodes, addFailureNodes)

	return addSuccessNodes, addFailureNodes, nil
}

// UpdateAKSNodesToDBTask update AKS nodes
func UpdateAKSNodesToDBTask(taskID string, stepName string) error {
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
			"step %s failed, %s", taskID, clusterID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("get cloud/project information failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// check cluster status
	ctx := cloudprovider.WithTaskIDForContext(context.Background(), taskID)
	err = updateNodeToDB(ctx, state, dependInfo, nodeGroupIDs)
	if err != nil {
		blog.Errorf("UpdateNodesToDBTask[%s] checkNodesGroupStatus[%s] failed: %v",
			taskID, clusterID, err)
		retErr := fmt.Errorf("UpdateNodesToDBTask[%s] timeout|abnormal", clusterID)
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
	nodeResourceGroup := info.Cluster.ExtraInfo[common.NodeResourceGroup]
	// get azureCloud client
	cli, err := api.NewAksServiceImplWithCommonOption(info.CmOption)
	if err != nil {
		blog.Errorf("updateNodeToDB[%s] get aks client failed: %s", taskID, err.Error())
		return fmt.Errorf("updateNodeToDB get aks client err, %s", err.Error())
	}

	addSuccessNodes := make([]string, 0)
	for _, ngID := range nodeGroupIDs {
		nodeGroup, err := actions.GetNodeGroupByGroupID(cloudprovider.GetStorageModel(), ngID)
		if err != nil {
			return fmt.Errorf("updateNodeToDB GetNodeGroupByGroupID information failed, %s", err.Error())
		}

		vmssList, err := cli.ListInstanceAndReturn(ctx, nodeResourceGroup, nodeGroup.AutoScaling.AutoScalingID)
		if err != nil {
			return fmt.Errorf("updateNodeToDB ListInstanceAndReturn failed, %s", err.Error())
		}
		interfaceList := make([]*armnetwork.Interface, 0)
		// 获取 interface list
		err = retry.Do(func() error {
			interfaceList, err = cli.ListSetInterfaceAndReturn(ctx, nodeResourceGroup, nodeGroup.AutoScaling.AutoScalingID)
			if err != nil {
				return fmt.Errorf("updateNodeToDB ListSetInterfaceAndReturn failed, %v", err)
			}
			return nil
		}, retry.Context(ctx), retry.Attempts(3))
		if err != nil {
			return fmt.Errorf("updateNodeToDB get vm network interface failed, %v", err)
		}

		info.NodeGroup = nodeGroup
		nodes, err := vmToNode(cli, info, vmssList, interfaceList)
		if err != nil {
			return fmt.Errorf("updateNodeToDB vmToNode failed, %v", err)
		}
		for _, n := range nodes {
			if n.Status == "running" {
				n.Status = common.StatusRunning
				addSuccessNodes = append(addSuccessNodes, n.InnerIP)
			} else {
				n.Status = common.StatusAddNodesFailed
			}
			n.NodeGroupID = nodeGroup.NodeGroupID
			err = cloudprovider.GetStorageModel().CreateNode(context.Background(), n)
			if err != nil {
				return fmt.Errorf("updateNodeToDB CreateNode[%s] failed, %v", n.NodeName, err)
			}
		}
	}
	state.Task.CommonParams[cloudprovider.NodeIPsKey.String()] = strings.Join(addSuccessNodes, ",")

	return nil
}

// RegisterAKSClusterKubeConfigTask register cluster kubeconfig
func RegisterAKSClusterKubeConfigTask(taskID string, stepName string) error {
	start := time.Now()

	// get task and task current step
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	// previous step successful when retry task
	if step == nil {
		blog.Infof("RegisterAKSClusterKubeConfigTask[%s]: current step[%s] successful and skip", taskID, stepName)
		return nil
	}
	blog.Infof("RegisterAKSClusterKubeConfigTask[%s] task %s run current step %s, system: %s, old state: %s, params %v",
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
		blog.Errorf("RegisterAKSClusterKubeConfigTask[%s] GetClusterDependBasicInfo in task %s step %s failed, %s",
			taskID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("get cloud/project information failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	err = importClusterCredential(ctx, dependInfo)
	if err != nil {
		blog.Errorf("RegisterAKSClusterKubeConfigTask[%s] importClusterCredential failed: %s", taskID, err.Error())
		retErr := fmt.Errorf("importClusterCredential failed %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	blog.Infof("RegisterAKSClusterKubeConfigTask[%s] importClusterCredential success", taskID)

	// update step
	if err = state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("RegisterAKSClusterKubeConfigTask[%s:%s] update to storage fatal", taskID, stepName)
		return err
	}

	return nil
}
