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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/aws/business"
	providerutils "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/clusterops"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	icommon "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/loop"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

var defaultAddons = []string{"coredns", "kube-proxy", "vpc-cni", "eks-pod-identity-agent"}

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

// createEKSCluster 创建一个Amazon EKS集群，并更新集群信息。
func createEKSCluster(ctx context.Context, info *cloudprovider.CloudDependBasicInfo) (
	string, error) {
	// 从上下文中获取任务ID
	taskID := cloudprovider.GetTaskIDFromContext(ctx)
	cluster := info.Cluster

	// 创建一个新的AWS客户端集
	client, err := api.NewAWSClientSet(info.CmOption)
	if err != nil {
		return "", fmt.Errorf("create eksService failed")
	}

	// 获取IAM角色信息
	role, err := client.GetRole(&iam.GetRoleInput{RoleName: aws.String(cluster.GetClusterIamRole())})
	if err != nil {
		// 如果获取角色失败，记录错误日志并返回错误
		blog.Errorf("GetRole[%s] failed, %v", taskID, err)
		return "", err
	}

	// generateCreateClusterInput 生成创建集群的输入参数
	input, err := generateCreateClusterInput(info, role.Arn)
	if err != nil {
		// 如果生成输入参数失败，返回错误
		return "", fmt.Errorf("generateCreateClusterInput failed, %v", err)
	}

	// CreateEksCluster 调用API创建EKS集群
	eksCluster, err := client.CreateEksCluster(input)
	if err != nil {
		// 如果创建集群失败，返回错误
		return "", fmt.Errorf("call CreateEksCluster failed, %v", err)
	}

	info.Cluster.SystemID = *eksCluster.Name
	info.Cluster.VpcID = *eksCluster.ResourcesVpcConfig.VpcId

	// 更新集群信息
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

// generateCreateClusterInput 根据提供的云提供商基本信息和角色ARN生成创建EKS集群的输入参数
func generateCreateClusterInput(info *cloudprovider.CloudDependBasicInfo, roleArn *string) (
	*eks.CreateClusterInput, error) {

	var (
		cluster   = info.Cluster      // 获取集群信息
		subnetIds = make([]string, 0) // 初始化子网ID切片
		err       error
	)

	if cluster.GetClusterAdvanceSettings().GetNetworkType() == icommon.VpcCni {
		subnetIds, err = business.AllocateClusterVpcCniSubnets(context.Background(), cluster.VpcID,
			cluster.GetNetworkSettings().GetSubnetSource().GetNew(), info.CmOption)
		if err != nil {
			return nil, err
		}
		if len(info.Cluster.GetNetworkSettings().GetSubnetSource().GetExisted().GetIds()) > 0 {
			subnetIds = append(subnetIds, info.Cluster.GetNetworkSettings().GetSubnetSource().GetExisted().GetIds()...)
		}
	}
	// 如果subnetIds为空，则返回错误
	if len(subnetIds) == 0 {
		return nil, errors.New("generateCreateClusterInput subnetIds is empty")
	}
	info.Cluster.NetworkSettings.EniSubnetIDs = subnetIds

	sgs := strings.Split(cluster.ClusterAdvanceSettings.ClusterConnectSetting.SecurityGroup, ",")
	input := &eks.CreateClusterInput{
		AccessConfig: &eks.CreateAccessConfigRequest{
			AuthenticationMode:                      aws.String(api.ClusterAuthenticationModeAM), // 设置认证模式
			BootstrapClusterCreatorAdminPermissions: aws.Bool(true),                              // 默认启用集群创建者管理员权限
		},
		Name:    aws.String(strings.ToLower(cluster.ClusterID)), // 设置集群名称
		RoleArn: roleArn,                                        // 设置角色ARN
		ResourcesVpcConfig: &eks.VpcConfigRequest{
			SubnetIds:             aws.StringSlice(subnetIds),                                                 // 设置子网ID
			SecurityGroupIds:      aws.StringSlice(sgs),                                                       // 设置安全组ID
			EndpointPrivateAccess: aws.Bool(!cluster.ClusterAdvanceSettings.ClusterConnectSetting.IsExtranet), // 设置私有访问端点
			EndpointPublicAccess:  aws.Bool(cluster.ClusterAdvanceSettings.ClusterConnectSetting.IsExtranet),  // 设置公共访问端点
			PublicAccessCidrs: func() []*string {
				// 如果集群的网络设置中的公共访问CIDR存在，则返回对应的切片，否则返回nil
				if cluster.ClusterAdvanceSettings.ClusterConnectSetting.Internet == nil ||
					cluster.ClusterAdvanceSettings.ClusterConnectSetting.Internet.PublicAccessCidrs == nil {
					return nil
				}
				return aws.StringSlice(cluster.ClusterAdvanceSettings.ClusterConnectSetting.Internet.PublicAccessCidrs)
			}(),
		},
		UpgradePolicy: func(setting *proto.ClusterBasicSetting) *eks.UpgradePolicyRequest {
			// 如果升级策略未设置，则默认使用EXTENDED策略
			if setting.UpgradePolicy == nil {
				return &eks.UpgradePolicyRequest{SupportType: aws.String(api.ClusterUpdatePolicyExtended)}
			}
			// 如果升级策略不是EXTENDED或STANDARD，则使用EXTENDED策略
			if setting.UpgradePolicy.SupportType != api.ClusterUpdatePolicyExtended &&
				setting.UpgradePolicy.SupportType != api.ClusterUpdatePolicyStandard {
				return &eks.UpgradePolicyRequest{SupportType: aws.String(api.ClusterUpdatePolicyExtended)}
			}
			// 否则使用设置的升级策略
			return &eks.UpgradePolicyRequest{SupportType: aws.String(setting.UpgradePolicy.SupportType)}
		}(cluster.ClusterBasicSettings),
		Version: aws.String(cluster.ClusterBasicSettings.Version), // 设置Kubernetes版本
	}
	input.KubernetesNetworkConfig = generateKubernetesNetworkConfig(cluster)
	if len(cluster.ClusterBasicSettings.ClusterTags) > 0 {
		input.Tags = aws.StringMap(cluster.ClusterBasicSettings.ClusterTags)
	}

	return input, nil
}

// generateKubernetesNetworkConfig network config
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

	err = createAddon(ctx, dependInfo)
	if err != nil {
		blog.Errorf("CheckEKSClusterStatusTask[%s] createAddon[%s] failed: %v",
			taskID, clusterID, err)
		retErr := fmt.Errorf("createAddon[%s] failed, %s", clusterID, err.Error())
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

// createAddon 函数用于在AWS EKS集群上创建默认的addons。
func createAddon(ctx context.Context, info *cloudprovider.CloudDependBasicInfo) error {
	// 从上下文中获取任务ID，用于日志记录
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	// 初始化AWS EKS客户端
	cli, err := api.NewEksClient(info.CmOption)
	if err != nil {
		blog.Errorf("checkClusterStatus[%s] get aws client failed: %s", taskID, err.Error())
		retErr := fmt.Errorf("get cloud aws client err, %s", err.Error())
		return retErr
	}

	// 遍历默认的addons列表
	for _, addon := range defaultAddons {
		// 调用CreateAddon方法创建addon
		_, err = cli.CreateAddon(&eks.CreateAddonInput{
			ClusterName: aws.String(info.Cluster.ClusterName), // 设置集群名称
			AddonName:   aws.String(addon),                    // 设置要创建的addon名称
		})
		// 如果创建addon失败，则直接返回错误
		if err != nil {
			return err
		}
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

	var totalNodesNum uint32
	for _, ngID := range strings.Split(nodeGroupIDs, ",") {
		nodeGroup, _ := actions.GetNodeGroupByGroupID(cloudprovider.GetStorageModel(), ngID)
		// 运行至此, 认为节点池已创建成功
		err = cloudprovider.UpdateNodeGroupStatus(ngID, common.StatusRunning)
		if err != nil {
			blog.Errorf("CheckEKSClusterNodesStatusTask[%s] UpdateNodeGroupStatus failed, %s", taskID, err.Error())
			return fmt.Errorf("UpdateNodeGroupStatus failed, %s", err.Error())
		}
		totalNodesNum += nodeGroup.AutoScaling.DesiredSize
	}
	// check cluster status
	ctx := cloudprovider.WithTaskIDForContext(context.Background(), taskID)
	addSuccessNodes, addFailureNodes, err := checkClusterNodesStatus(ctx, dependInfo, totalNodesNum)
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
	if totalNodesNum != 0 && len(addSuccessNodes) == 0 {
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
	totalNodesNum uint32) ([]string, []string, error) {
	var addSuccessNodes, addFailureNodes = make([]string, 0), make([]string, 0)
	taskID := cloudprovider.GetTaskIDFromContext(ctx)
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

		nodes, errLocal := k8sOperator.ListClusterNodes(context.Background(), info.Cluster.ClusterID) // nolint
		if errLocal != nil {
			blog.Errorf("checkClusterNodesStatus[%s] cluster[%s] failed: %v", taskID, info.Cluster.ClusterID, err)
			return nil, nil, errLocal
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

	nodeIPs, instanceIDs, nodeNames := make([]string, 0), make([]string, 0), make([]string, 0)
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
				nodeNames = append(nodeNames, n.GetNodeName())
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
	state.Task.CommonParams[cloudprovider.NodeNamesKey.String()] = strings.Join(nodeNames, ",")
	// dynamic inject paras
	state.Task.CommonParams[cloudprovider.DynamicNodeIPListKey.String()] = strings.Join(nodeIPs, ",")

	return nil
}
