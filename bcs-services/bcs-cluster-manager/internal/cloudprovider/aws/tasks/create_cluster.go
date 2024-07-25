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
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/avast/retry-go"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/aws/aws-sdk-go/service/iam"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/aws/api"
	providerutils "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/clusterops"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/loop"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

// CreateEKSClusterTask call aws interface to create cluster
func CreateEKSClusterTask(taskID string, stepName string) error {
	start := time.Now()

	// get task and task current step
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	// previous step successful when retry task
	if step == nil {
		blog.Infof("CreateEKSClusterTask[%s]: current step[%s] successful and skip", taskID, stepName)
		return nil
	}
	blog.Infof("CreateEKSClusterTask[%s]: task %s run step %s, system: %s, old state: %s, params %v",
		taskID, taskID, stepName, step.System, step.Status, step.Params)

	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	cloudID := step.Params[cloudprovider.CloudIDKey.String()]

	// get dependent basic info
	dependInfo, err := cloudprovider.GetClusterDependBasicInfo(cloudprovider.GetBasicInfoReq{
		ClusterID: clusterID,
		CloudID:   cloudID,
	})
	if err != nil {
		blog.Errorf("CreateEKSClusterTask[%s]: GetClusterDependBasicInfo for cluster %s in task %s "+
			"step %s failed, %s", taskID, clusterID, taskID, stepName, err.Error()) // nolint
		retErr := fmt.Errorf("get cloud/project information failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// inject taskID
	ctx := cloudprovider.WithTaskIDForContext(context.Background(), taskID)
	// create cluster task
	clsId, err := createEKSCluster(ctx, dependInfo)
	if err != nil {
		blog.Errorf("CreateEKSClusterTask[%s] createEKSCluster for cluster[%s] failed, %s",
			taskID, clusterID, err.Error())
		retErr := fmt.Errorf("createEKSCluster err, %s", err.Error())
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
		blog.Errorf("CreateEKSClusterTask[%s] task %s %s update to storage fatal", taskID, taskID, stepName)
		return err
	}

	return nil
}

func createEKSCluster(ctx context.Context, info *cloudprovider.CloudDependBasicInfo) (
	string, error) {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)
	cluster := info.Cluster

	client, err := api.NewAWSClientSet(info.CmOption)
	if err != nil {
		return "", fmt.Errorf("create eksService failed")
	}

	role, err := client.GetRole(&iam.GetRoleInput{RoleName: aws.String(cluster.GetClusterIamRole())})
	if err != nil {
		blog.Errorf("GetRole[%s] failed, %v", taskID, err)
		return "", err
	}

	input := generateCreateClusterInput(info.Cluster, role.Arn)

	eksCluster, err := client.CreateEksCluster(input)
	if err != nil {
		return "", fmt.Errorf("call CreateEksCluster failed, %v", err)
	}

	info.Cluster.SystemID = *eksCluster.Name
	info.Cluster.VpcID = *eksCluster.ResourcesVpcConfig.VpcId

	err = cloudprovider.UpdateCluster(info.Cluster)
	if err != nil {
		blog.Errorf("createEKSCluster[%s] UpdateCluster[%s] failed %s",
			taskID, info.Cluster.ClusterID, err.Error())
		retErr := fmt.Errorf("call createEKSCluster UpdateCluster[%s] api err: %s",
			info.Cluster.ClusterID, err.Error())
		return "", retErr
	}
	blog.Infof("createEKSCluster[%s] run successful", taskID)

	return *eksCluster.Name, nil
}

func generateCreateClusterInput(cluster *proto.Cluster, roleArn *string) *eks.CreateClusterInput {
	subnets := strings.Split(cluster.ClusterBasicSettings.SubnetID, ",")
	sgs := strings.Split(cluster.ClusterAdvanceSettings.ClusterConnectSetting.SecurityGroup, ",")
	input := &eks.CreateClusterInput{
		Name:    aws.String(cluster.ClusterName),
		RoleArn: roleArn,
		ResourcesVpcConfig: &eks.VpcConfigRequest{
			SubnetIds:             aws.StringSlice(subnets),
			SecurityGroupIds:      aws.StringSlice(sgs),
			EndpointPrivateAccess: aws.Bool(true),
			EndpointPublicAccess:  aws.Bool(cluster.ClusterAdvanceSettings.ClusterConnectSetting.Internet.PublicIPAssigned),
		},
		Version: aws.String(cluster.ClusterBasicSettings.Version),
	}
	input.KubernetesNetworkConfig = generateKubernetesNetworkConfig(cluster)
	if len(cluster.ClusterBasicSettings.ClusterTags) > 0 {
		input.Tags = aws.StringMap(cluster.ClusterBasicSettings.ClusterTags)
	}

	return input
}

func generateKubernetesNetworkConfig(cluster *proto.Cluster) *eks.KubernetesNetworkConfigRequest {
	req := &eks.KubernetesNetworkConfigRequest{}
	if cluster != nil && cluster.NetworkSettings != nil {
		switch cluster.NetworkSettings.ClusterIpType {
		case "ipv4":
			req.IpFamily = aws.String("ipv4")
			req.ServiceIpv4Cidr = aws.String(cluster.NetworkSettings.ServiceIPv4CIDR)
		case "ipv6":
			req.IpFamily = aws.String("ipv6")
		default:
			req.IpFamily = aws.String("ipv4")
			req.ServiceIpv4Cidr = aws.String(cluster.NetworkSettings.ServiceIPv4CIDR)
		}

	}

	return req
}

// CheckEKSClusterStatusTask check cluster status
func CheckEKSClusterStatusTask(taskID string, stepName string) error {
	start := time.Now()
	// get task and task current step
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	// previous step successful when retry task
	if step == nil {
		blog.Infof("CheckEKSClusterStatusTask[%s]: current step[%s] successful and skip", taskID, stepName)
		return nil
	}
	blog.Infof("CheckEKSClusterStatusTask[%s]: task %s run step %s, system: %s, old state: %s, params %v",
		taskID, taskID, stepName, step.System, step.Status, step.Params)

	// step login started here
	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	cloudID := step.Params[cloudprovider.CloudIDKey.String()]

	dependInfo, err := cloudprovider.GetClusterDependBasicInfo(cloudprovider.GetBasicInfoReq{
		ClusterID: clusterID,
		CloudID:   cloudID,
	})
	if err != nil {
		blog.Errorf("CheckEKSClusterStatusTask[%s]: GetClusterDependBasicInfo for cluster %s in task %s "+
			"step %s failed, %s", taskID, clusterID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("get cloud/project information failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// check cluster status
	ctx := cloudprovider.WithTaskIDForContext(context.Background(), taskID)
	err = checkClusterStatus(ctx, dependInfo)
	if err != nil {
		blog.Errorf("CheckEKSClusterStatusTask[%s] checkClusterStatus[%s] failed: %v",
			taskID, clusterID, err)
		retErr := fmt.Errorf("checkClusterStatus[%s] timeout|abnormal", clusterID)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// update step
	if err = state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("CheckEKSClusterStatusTask[%s] task %s %s update to storage fatal",
			taskID, taskID, stepName)
		return err
	}

	return nil
}

// checkClusterStatus check cluster status
func checkClusterStatus(ctx context.Context, info *cloudprovider.CloudDependBasicInfo) error {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	// get awsCloud client
	cli, err := api.NewEksClient(info.CmOption)
	if err != nil {
		blog.Errorf("checkClusterStatus[%s] get aws client failed: %s", taskID, err.Error())
		retErr := fmt.Errorf("get cloud aws client err, %s", err.Error())
		return retErr
	}

	var (
		failed = false
	)

	ctx, cancel := context.WithTimeout(ctx, 30*time.Minute)
	defer cancel()

	// loop cluster status
	err = loop.LoopDoFunc(ctx, func() error {
		cluster, errGet := cli.GetEksCluster(info.Cluster.SystemID)
		if errGet != nil {
			blog.Errorf("checkClusterStatus[%s] failed: %v", taskID, errGet)
			return nil
		}

		blog.Infof("checkClusterStatus[%s] cluster[%s] current status[%s]", taskID,
			info.Cluster.ClusterID, *cluster.Status)

		switch *cluster.Status {
		case eks.ClusterStatusActive:
			return loop.EndLoop
		case eks.ClusterStatusFailed:
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

// RegisterEKSClusterKubeConfigTask register cluster kubeconfig
func RegisterEKSClusterKubeConfigTask(taskID string, stepName string) error {
	start := time.Now()

	// get task and task current step
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	// previous step successful when retry task
	if step == nil {
		blog.Infof("RegisterAWSClusterKubeConfigTask[%s]: current step[%s] successful and skip", taskID, stepName)
		return nil
	}
	blog.Infof("RegisterAWSClusterKubeConfigTask[%s] task %s run current step %s, system: %s, old state: %s, params %v",
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
		blog.Errorf("RegisterAWSClusterKubeConfigTask[%s] GetClusterDependBasicInfo in task %s step %s failed, %s",
			taskID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("get cloud/project information failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	err = importClusterCredential(ctx, dependInfo)
	if err != nil {
		blog.Errorf("RegisterAWSClusterKubeConfigTask[%s] importClusterCredential failed: %s", taskID, err.Error())
		retErr := fmt.Errorf("importClusterCredential failed %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	blog.Infof("RegisterAWSClusterKubeConfigTask[%s] importClusterCredential success", taskID)

	// update step
	if err = state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("RegisterAWSClusterKubeConfigTask[%s:%s] update to storage fatal", taskID, stepName)
		return err
	}

	return nil
}

// CheckEKSClusterNodesStatusTask check cluster nodes status
func CheckEKSClusterNodesStatusTask(taskID string, stepName string) error {
	start := time.Now()
	// get task and task current step
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	// previous step successful when retry task
	if step == nil {
		blog.Infof("CheckEKSClusterNodesStatusTask[%s]: current step[%s] successful and skip", taskID, stepName)
		return nil
	}
	blog.Infof("CheckEKSClusterNodesStatusTask[%s]: task %s run step %s, system: %s, old state: %s, params %v",
		taskID, taskID, stepName, step.System, step.Status, step.Params)

	// step login started here
	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	cloudID := step.Params[cloudprovider.CloudIDKey.String()]
	nodeGroupIDs := step.Params[cloudprovider.NodeGroupIDKey.String()]
	state.Task.CommonParams[cloudprovider.SuccessNodeGroupIDsKey.String()] = nodeGroupIDs

	dependInfo, err := cloudprovider.GetClusterDependBasicInfo(cloudprovider.GetBasicInfoReq{
		ClusterID: clusterID,
		CloudID:   cloudID,
	})
	if err != nil {
		blog.Errorf("CheckEKSClusterNodesStatusTask[%s]: GetClusterDependBasicInfo for cluster %s in task %s "+
			"step %s failed, %s", taskID, clusterID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("get cloud/project information failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// check cluster status
	ctx := cloudprovider.WithTaskIDForContext(context.Background(), taskID)
	addSuccessNodes, addFailureNodes, err := checkClusterNodesStatus(ctx, dependInfo, strings.Split(nodeGroupIDs, ","))
	if err != nil {
		blog.Errorf("CheckEKSClusterStatusTask[%s] checkClusterStatus[%s] failed: %v",
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
	taskID := cloudprovider.GetTaskIDFromContext(ctx)
	for _, ngID := range nodeGroupIDs {
		nodeGroup, err := actions.GetNodeGroupByGroupID(cloudprovider.GetStorageModel(), ngID)
		if err != nil {
			return nil, nil, fmt.Errorf("get nodegroup information failed, %s", err.Error())
		}
		// 运行至此, 认为节点池已创建成功
		err = cloudprovider.UpdateNodeGroupStatus(ngID, common.StatusRunning)
		if err != nil {
			blog.Errorf("checkClusterNodesStatus[%s] UpdateNodeGroupStatus failed, %s", taskID, err.Error())
			return nil, nil, fmt.Errorf("UpdateNodeGroupStatus failed, %s", err.Error())
		}
		totalNodesNum += nodeGroup.AutoScaling.DesiredSize
	}

	k8sOperator := clusterops.NewK8SOperator(options.GetGlobalCMOptions(), cloudprovider.GetStorageModel())

	// wait node group state to normal
	timeCtx, cancel := context.WithTimeout(context.TODO(), 10*time.Minute)
	defer cancel()

	// wait all nodes to be ready
	err := loop.LoopDoFunc(timeCtx, func() error {
		running := make([]string, 0)

		nodes, err := k8sOperator.ListClusterNodes(context.Background(), info.Cluster.ClusterID)
		if err != nil {
			blog.Errorf("checkClusterNodesStatus[%s] cluster[%s] failed: %v", taskID, info.Cluster.ClusterID, err)
			return nil
		}

		for _, ins := range nodes {
			if utils.CheckNodeIfReady(ins) {
				blog.Infof("checkClusterNodesStatus[%s] node[%s] ready", taskID, ins.Name)
				// get instanceID
				providerID := strings.Split(ins.Spec.ProviderID, "/")
				running = append(running, providerID[len(providerID)-1])
			}
		}

		blog.Infof("checkClusterNodesStatus[%s] ready nodes[%+v]", taskID, running)
		if len(running) == int(totalNodesNum) {
			addSuccessNodes = running
			return loop.EndLoop
		}

		return nil
	}, loop.LoopInterval(30*time.Second))
	// other error
	if err != nil && !errors.Is(err, context.DeadlineExceeded) {
		blog.Errorf("checkClusterNodesStatus[%s] check nodes status failed: %v", taskID, err)
		return nil, nil, err
	}

	// timeout error
	if errors.Is(err, context.DeadlineExceeded) {
		running, failure := make([]string, 0), make([]string, 0)

		nodes, err := k8sOperator.ListClusterNodes(context.Background(), info.Cluster.ClusterID) // nolint
		if err != nil {
			blog.Errorf("checkClusterNodesStatus[%s] cluster[%s] failed: %v", taskID, info.Cluster.ClusterID, err)
			return nil, nil, err
		}

		for _, ins := range nodes {
			if utils.CheckNodeIfReady(ins) {
				providerID := strings.Split(ins.Spec.ProviderID, "/")
				running = append(running, providerID[len(providerID)-1])
			} else {
				providerID := strings.Split(ins.Spec.ProviderID, "/")
				failure = append(failure, providerID[len(providerID)-1])
			}
		}

		addSuccessNodes = running
		addFailureNodes = failure
	}
	blog.Infof("checkClusterNodesStatus[%s] success[%v] failure[%v]", taskID, addSuccessNodes, addFailureNodes)

	return addSuccessNodes, addFailureNodes, nil
}

// UpdateEKSNodesToDBTask update AWS nodes
func UpdateEKSNodesToDBTask(taskID string, stepName string) error {
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
	addSuccessNodes := state.Task.CommonParams[cloudprovider.SuccessClusterNodeIDsKey.String()]
	addFailureNodes := state.Task.CommonParams[cloudprovider.FailedClusterNodeIDsKey.String()]

	nodeIPs, instanceIDs := make([]string, 0), make([]string, 0)
	nmClient := api.NodeManager{}
	nodes := make([]*proto.Node, 0)

	successInstanceID := strings.Split(addSuccessNodes, ",")
	failureInstanceID := strings.Split(addFailureNodes, ",")
	if addSuccessNodes != "" {
		instanceIDs = append(instanceIDs, successInstanceID...)
	}
	if addFailureNodes != "" {
		instanceIDs = append(instanceIDs, failureInstanceID...)
	}

	for _, ngID := range nodeGroupIDs {
		nodeGroup, err := actions.GetNodeGroupByGroupID(cloudprovider.GetStorageModel(), ngID)
		if err != nil {
			return fmt.Errorf("updateNodeToDB GetNodeGroupByGroupID information failed, %s", err.Error())
		}

		info.NodeGroup = nodeGroup

		err = retry.Do(func() error {
			nodes, err = nmClient.ListNodesByInstanceID(instanceIDs, &cloudprovider.ListNodesOption{
				Common:       info.CmOption,
				ClusterVPCID: info.Cluster.VpcID,
			})
			if err != nil {
				return err
			}
			return nil
		}, retry.Attempts(3))
		if err != nil {
			blog.Errorf("updateNodeToDB[%s] failed: %v", taskID, err)
			return err
		}

		for _, n := range nodes {
			n.ClusterID = info.NodeGroup.ClusterID
			n.NodeGroupID = info.NodeGroup.NodeGroupID
			if utils.StringInSlice(n.NodeID, successInstanceID) {
				n.Status = common.StatusRunning
				nodeIPs = append(nodeIPs, n.InnerIP)
			} else {
				n.Status = common.StatusAddNodesFailed
			}

			err = cloudprovider.SaveNodeInfoToDB(ctx, n, true)
			if err != nil {
				blog.Errorf("updateNodeToDB[%s] SaveNodeInfoToDB[%s] failed: %v", taskID, n.InnerIP, err)
			}
		}
	}
	state.Task.CommonParams[cloudprovider.NodeIPsKey.String()] = strings.Join(nodeIPs, ",")

	return nil
}
