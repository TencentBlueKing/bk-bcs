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
	"fmt"
	"strconv"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/huawei/api"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
)

// CreateClusterTask call qcloud interface to create cluster
func CreateClusterTask(taskID string, stepName string) error {
	start := time.Now()

	// get task and task current step
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	// previous step successful when retry task
	if step == nil {
		blog.Infof("CreateClusterTask[%s]: current step[%s] successful and skip", taskID, stepName)
		return nil
	}
	blog.Infof("CreateClusterTask[%s]: task %s run step %s, system: %s, old state: %s, params %v",
		taskID, taskID, stepName, step.System, step.Status, step.Params)

	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	cloudID := step.Params[cloudprovider.CloudIDKey.String()]

	nodeTemplateID := step.Params[cloudprovider.NodeTemplateIDKey.String()]
	operator := state.Task.CommonParams[cloudprovider.OperatorKey.String()]

	// get dependent basic info
	dependInfo, err := cloudprovider.GetClusterDependBasicInfo(cloudprovider.GetBasicInfoReq{
		ClusterID:      clusterID,
		CloudID:        cloudID,
		NodeTemplateID: nodeTemplateID,
	})
	if err != nil {
		blog.Errorf("CreateClusterTask[%s]: GetClusterDependBasicInfo for cluster %s in task %s "+
			"step %s failed, %s", taskID, clusterID, taskID, stepName, err.Error()) // nolint
		retErr := fmt.Errorf("get cloud/project information failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// inject taskID
	ctx := cloudprovider.WithTaskIDForContext(context.Background(), taskID)

	req, err := generateCreateClusterRequest(ctx, dependInfo, operator)
	if err != nil {
		blog.Errorf("createCluster[%s] generateCreateClusterRequest failed: %v", taskID, err)
		return err
	}

	// create cluster task
	clsId, err := createCluster(ctx, dependInfo, req, dependInfo.Cluster.SystemID)
	if err != nil {
		blog.Errorf("CreateClusterTask[%s] createCluster for cluster[%s] failed, %s",
			taskID, clusterID, err.Error())
		retErr := fmt.Errorf("createCluster err, %s", err.Error())
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
		blog.Errorf("CreateClusterTask[%s] task %s %s update to storage fatal", taskID, taskID, stepName)
		return err
	}
	return nil
}

func generateCreateClusterRequest(ctx context.Context, info *cloudprovider.CloudDependBasicInfo,
	operator string) (*api.CreateClusterRequest, error) {

	flavor, err := trans2CCEFlavor(info.Cluster.ClusterBasicSettings.ClusterLevel)
	if err != nil {
		return nil, err
	}

	containerMode := "overlay_l2"
	if info.Cluster.ClusterAdvanceSettings.NetworkType == common.VpcCni {
		containerMode = "vpc-router"
	}

	return &api.CreateClusterRequest{
		Name: info.Cluster.ClusterID,
		Spec: api.CreateClusterSpec{
			Flavor:          flavor,
			Version:         info.Cluster.ClusterBasicSettings.Version,
			Description:     info.Cluster.GetDescription(),
			VpcID:           info.Cluster.VpcID,
			SubnetID:        info.Cluster.ClusterBasicSettings.SubnetID,
			SecurityGroupID: info.Cluster.ClusterAdvanceSettings.ClusterConnectSetting.SecurityGroup,
			ContainerMode:   containerMode,
			ContainerCidr:   info.Cluster.NetworkSettings.ClusterIPv4CIDR,
			ServiceCidr:     info.Cluster.NetworkSettings.ServiceIPv4CIDR,
			Charge: api.ChargePrepaid{
				ChargeType: "POSTPAID_BY_HOUR",
				Period:     0,
				RenewFlag:  "",
			},
			Ipv6Enable: false,
		},
	}, nil
}

func trans2CCEFlavor(s string) (string, error) {
	if len(s) < 2 || string(s[0]) != "L" {
		return "", fmt.Errorf("invalid format, expected prefix 'L'")
	}

	numStr := s[1:]                       // 提取首字母后的部分
	levelNum, err := strconv.Atoi(numStr) // 尝试转换为整数
	if err != nil {
		return "", fmt.Errorf("failed to parse number: %w", err)
	}

	if levelNum <= 0 {
		return "", fmt.Errorf("cluster level must be greater than 0")
	} else if levelNum <= 50 {
		return "cce.s1.small", nil
	} else if levelNum <= 200 {
		return "cce.s1.medium", nil
	} else if levelNum <= 1000 {
		return "cce.s2.large", nil
	}

	// levelNum > 1000
	return "cce.s2.xlarge", nil
}

func createCluster(ctx context.Context, info *cloudprovider.CloudDependBasicInfo,
	request *api.CreateClusterRequest, clsId string) (string, error) {
	client, err := api.NewCceClient(info.CmOption)
	if err != nil {
		return "", err
	}

	rsp, err := client.CreateCluster(request)
	if err != nil {
		return "", err
	}

	return *rsp.Metadata.Uid, nil
}
