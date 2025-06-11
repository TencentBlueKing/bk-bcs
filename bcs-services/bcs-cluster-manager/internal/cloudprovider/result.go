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

package cloudprovider

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/project"
)

var (
	// ErrJobType err
	ErrJobType = "unSupport jobType %s"
)

// JobType for task type
type JobType string

var (
	// CreateClusterJob for createCluster job
	CreateClusterJob JobType = "create-cluster"
	// CreateVirtualClusterJob create virtual job
	CreateVirtualClusterJob JobType = "create-virtual-cluster"
	// ImportClusterJob for importCluster job
	ImportClusterJob JobType = "import-cluster"
	// DeleteClusterJob for deleteCluster job
	DeleteClusterJob JobType = "delete-cluster"
	// DeleteVirtualClusterJob for deleteVirtualCluster job
	DeleteVirtualClusterJob JobType = "delete-virtual-cluster"
	// AddNodeJob for addNodes job
	AddNodeJob JobType = "add-node"
	// DeleteNodeJob for deleteNodes job
	DeleteNodeJob JobType = "delete-node"
	// AddExternalNodeJob for addNodes job
	AddExternalNodeJob JobType = "add-external-node"
	// DeleteExternalNodeJob for deleteNodes job
	DeleteExternalNodeJob JobType = "delete-external-node"
	// SwitchClusterNetworkJob for switchClusterNetwork job
	SwitchClusterNetworkJob JobType = "switch-cluster-network"

	// CreateNodeGroupJob for createNodeGroup job
	CreateNodeGroupJob JobType = "create-nodegroup"
	// UpdateNodeGroupJob for updateNodeGroup job
	UpdateNodeGroupJob JobType = "update-nodegroup"
	// DeleteNodeGroupJob for deleteNodeGroup job
	DeleteNodeGroupJob JobType = "delete-nodegroup"
	// UpdateNodeGroupDesiredNodeJob for UpdateNodeGroupDesiredNodeJob job
	UpdateNodeGroupDesiredNodeJob JobType = "update-nodegroup-desired-node"
	// CleanNodeGroupNodesJob for cleanNodeGroupNodes job
	CleanNodeGroupNodesJob JobType = "clean-nodegroup-nodes"

	// MoveNodesToNodeGroupJob for moveNodesToNodeGroup job
	MoveNodesToNodeGroupJob JobType = "move-nodes-to-nodegroup"
	// SwitchNodeGroupAutoScalingJob for switchNodeGroupAutoScaling job
	SwitchNodeGroupAutoScalingJob JobType = "switch-nodegroup-autoscaling"
	// UpdateAutoScalingOptionJob for updateAutoScalingOption job
	UpdateAutoScalingOptionJob JobType = "update-autoscalingoption"
	// SwitchAutoScalingOptionStatusJob for switchAutoScalingOptionStatus job
	SwitchAutoScalingOptionStatusJob JobType = "switch-autoscalingoption-status"
)

// String xxx
func (jt JobType) String() string {
	return string(jt)
}

// StatusResult for job result status
type StatusResult struct {
	Success string
	Failure string
}

// SyncJobResult for sync job result
type SyncJobResult struct {
	JobType     JobType
	TaskID      string
	ClusterID   string
	NodeGroupID string
	NodeIPs     []string
	NodeIDs     []string
	Status      StatusResult
}

func printOperationLog(taskId, clusterId, groupId string, operation string, isSuccess bool) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	cls, err := GetStorageModel().GetCluster(ctx, clusterId)
	if err != nil {
		blog.Errorf("task[%s] printClusterOperation failed: %v", taskId, err)
		return
	}
	proInfo, errLocal := project.GetProjectManagerClient().GetProjectInfo(ctx, cls.ProjectID, true)
	if errLocal != nil {
		blog.Errorf("task[%s] printClusterOperation GetProjectInfo[%s] failed: %v",
			taskId, cls.ProjectID, errLocal)
		return
	}
	task, err := GetStorageModel().GetTask(ctx, taskId)
	if err != nil {
		blog.Errorf("task[%s] printClusterOperation GetTask failed: %v", taskId, err)
		return
	}

	var bizId = "0"
	stepInfo, ok := task.GetSteps()[task.GetCurrentStep()]
	if ok && len(stepInfo.GetParams()[BkSopsBizIDKey.String()]) > 0 {
		bizId = stepInfo.GetParams()[BkSopsBizIDKey.String()]
	}

	result := common.TaskStatusSuccess
	if !isSuccess {
		result = common.TaskStatusFailure
	}

	blog.Infof("task %s operation %s for cluster %s nodegroup %s in project %s result: %v, Url: %v, "+
		"currentStep: %s:%s", taskId, operation, clusterId, groupId, cls.ProjectID, result,
		fmt.Sprintf(options.GetGlobalCMOptions().ComponentDeploy.BcsClusterUrl, proInfo.GetProjectCode(), clusterId),
		bizId, task.GetCurrentStep())
}

// UpdateJobResultStatus update job status by result
func (sjr *SyncJobResult) UpdateJobResultStatus(isSuccess bool) error {
	if sjr == nil {
		return ErrServerIsNil
	}

	blog.Infof("task[%s] JobType[%s] isSuccess[%v] ClusterID[%s] nodeIPs[%v]",
		sjr.TaskID, sjr.JobType, isSuccess, sjr.ClusterID, sjr.NodeIPs)
	printOperationLog(sjr.TaskID, sjr.ClusterID, sjr.NodeGroupID, sjr.JobType.String(), isSuccess)

	defer func() {
		_ = SendUserNotifyByTemplates(sjr.ClusterID, sjr.NodeGroupID, sjr.TaskID, isSuccess)
	}()

	switch sjr.JobType {
	case CreateClusterJob, CreateVirtualClusterJob:
		sjr.Status = generateStatusResult(common.StatusRunning, common.StatusCreateClusterFailed)
		return sjr.updateClusterResultStatus(isSuccess)
	case DeleteClusterJob, DeleteVirtualClusterJob:
		sjr.Status = generateStatusResult(common.StatusDeleted, common.StatusDeleteClusterFailed)
		return sjr.updateClusterResultStatus(isSuccess)
	// 上架节点
	case AddNodeJob, AddExternalNodeJob:
		sjr.Status = generateStatusResult(common.StatusRunning, common.StatusAddNodesFailed)
		return sjr.updateNodesResultStatus(isSuccess)
	// 下架节点
	case DeleteNodeJob, DeleteExternalNodeJob:
		sjr.Status = generateStatusResult("", common.StatusRemoveNodesFailed)
		return sjr.deleteNodesResultStatus(isSuccess)
	case ImportClusterJob:
		sjr.Status = generateStatusResult(common.StatusRunning, common.StatusImportClusterFailed)
		return sjr.updateClusterResultStatus(isSuccess)
	// CA 扩容节点
	case UpdateNodeGroupDesiredNodeJob:
		sjr.Status = generateStatusResult(common.StatusRunning, common.StatusAddNodesFailed)
		return sjr.updateCANodesResultStatus(isSuccess)
	// CA 缩容节点
	case CleanNodeGroupNodesJob:
		sjr.Status = generateStatusResult("", common.StatusRemoveCANodesFailed)
		return sjr.deleteCANodesResultStatus(isSuccess)
	case DeleteNodeGroupJob:
		sjr.Status = generateStatusResult("", common.StatusDeleteNodeGroupFailed)
		return sjr.updateDeleteNodeGroupStatus(isSuccess)
	case CreateNodeGroupJob:
		sjr.Status = generateStatusResult(common.StatusRunning, common.StatusCreateNodeGroupFailed)
		return sjr.updateNodeGroupStatus(isSuccess)
	case UpdateNodeGroupJob:
		sjr.Status = generateStatusResult(common.StatusRunning, common.StatusNodeGroupUpdateFailed)
		return sjr.updateNodeGroupStatus(isSuccess)
	case SwitchNodeGroupAutoScalingJob:
		sjr.Status = generateStatusResult(common.StatusRunning, common.StatusNodeGroupUpdateFailed)
		return sjr.updateNodeGroupStatus(isSuccess)
	case UpdateAutoScalingOptionJob:
		sjr.Status = generateStatusResult(common.StatusAutoScalingOptionNormal,
			common.StatusAutoScalingOptionUpdateFailed)
		return sjr.updateAutoScalingStatus(isSuccess)
	case SwitchAutoScalingOptionStatusJob:
		sjr.Status = generateStatusResult(common.StatusAutoScalingOptionNormal,
			common.StatusAutoScalingOptionUpdateFailed)
		return sjr.updateAutoScalingStatus(isSuccess)
	case SwitchClusterNetworkJob:
		sjr.Status = generateStatusResult("", common.TaskStatusFailure)
		return sjr.updateClusterNetworkStatus(isSuccess)
	}

	return fmt.Errorf(ErrJobType, sjr.JobType)
}

func generateStatusResult(successStatus string, failStatus string) StatusResult {
	return StatusResult{
		Success: successStatus,
		Failure: failStatus,
	}
}

func (sjr *SyncJobResult) deleteCANodesResultStatus(isSuccess bool) error {
	if len(sjr.NodeIPs) == 0 {
		return fmt.Errorf("SyncJobResult[%s] deleteCANodesResultStatus failed: %v", sjr.TaskID, "NodeIPs is empty")
	}

	blog.Infof("SyncJobResult[%s] deleteCANodesResultStatus[%v] %s:%s", sjr.TaskID,
		isSuccess, sjr.Status.Success, sjr.Status.Failure)

	if isSuccess {
		blog.Infof("task[%s] deleteCANodesResultStatus isSuccess[%v] InnerIPs[%v]", sjr.TaskID, isSuccess, sjr.NodeIPs)
		err := sjr.updateNodeGroupDesiredNum()
		if err != nil {
			blog.Errorf("task[%s] deleteCANodesResultStatus failed: %v", sjr.TaskID, err)
		}

		return sjr.deleteClusterNodesByIP(sjr.ClusterID, sjr.NodeIPs)
	}

	return sjr.updateNodeStatusByIP(sjr.ClusterID, sjr.NodeIPs, sjr.Status.Failure)
}

func (sjr *SyncJobResult) updateNodeGroupDesiredNum() error {
	nodeGroupID := sjr.NodeGroupID
	if len(nodeGroupID) == 0 {
		return fmt.Errorf("task[%s] updateNodeGroupDesiredNum nodeGroupID is empty", sjr.TaskID)
	}

	group, err := GetStorageModel().GetNodeGroup(context.Background(), nodeGroupID)
	if err != nil {
		return fmt.Errorf("task[%s] updateNodeGroupDesiredNum get NodeGroup[%s] failed %s", sjr.TaskID,
			nodeGroupID, err.Error())
	}

	blog.Infof("task[%s] update nodeGroup current[%d] clean[%d]", sjr.TaskID,
		group.AutoScaling.DesiredSize, len(sjr.NodeIPs))

	// update desired size
	currentSize := group.AutoScaling.DesiredSize
	if int(currentSize) >= len(sjr.NodeIPs) {
		group.AutoScaling.DesiredSize = uint32(int(currentSize) - len(sjr.NodeIPs))
	} else {
		group.AutoScaling.DesiredSize = 0
	}

	// update nodeGroup desired nodes num
	err = GetStorageModel().UpdateNodeGroup(context.Background(), group)
	if err != nil {
		return fmt.Errorf("task[%s] updateNodeGroupDesiredNum[%s] update NodeGroup failed %s", sjr.TaskID,
			nodeGroupID, err.Error())
	}

	return nil
}

func (sjr *SyncJobResult) deleteNodesResultStatus(isSuccess bool) error {
	if len(sjr.NodeIPs) == 0 {
		return fmt.Errorf("SyncJobResult deleteNodesResultStatus failed: %v", "NodeIPs&NodeIDs is empty")
	}

	if isSuccess {
		return sjr.deleteClusterNodesByIP(sjr.ClusterID, sjr.NodeIPs)
	}

	return sjr.updateNodeStatusByIP(sjr.ClusterID, sjr.NodeIPs, sjr.Status.Failure)
}

func (sjr *SyncJobResult) deleteClusterNodesByIP(clusterID string, nodeIPs []string) error {
	for _, ip := range nodeIPs {
		err := GetStorageModel().DeleteClusterNodeByIP(context.Background(), clusterID, ip)
		if err != nil {
			blog.Errorf("deleteClusterNodesByIP[%s] cluster&IP[%s:%s] failed: %v", sjr.TaskID, clusterID, ip, err)
			continue
		}
		blog.Infof("deleteClusterNodesByIP[%s] cluster&IP[%s:%s]", sjr.TaskID, clusterID, ip)
	}

	return nil
}

func (sjr *SyncJobResult) updateClusterResultStatus(isSuccess bool) error {
	cluster, err := GetStorageModel().GetCluster(context.Background(), sjr.ClusterID)
	if err != nil {
		return err
	}

	cluster.Status = sjr.Status.Failure
	if isSuccess {
		cluster.Status = sjr.Status.Success
		cluster.Message = ""
	}

	err = GetStorageModel().UpdateCluster(context.Background(), cluster)
	if err != nil {
		return err
	}

	return nil
}

func (sjr *SyncJobResult) updateClusterNetworkStatus(isSuccess bool) error {
	cluster, err := GetStorageModel().GetCluster(context.Background(), sjr.ClusterID)
	if err != nil {
		return err
	}

	cluster.NetworkSettings.Status = sjr.Status.Failure
	if isSuccess {
		cluster.NetworkSettings.Status = sjr.Status.Success
		cluster.Message = ""
	}

	err = GetStorageModel().UpdateCluster(context.Background(), cluster)
	if err != nil {
		return err
	}

	return nil
}

func (sjr *SyncJobResult) updateCANodesResultStatus(isSuccess bool) error {
	if len(sjr.NodeIPs) == 0 {
		return fmt.Errorf("SyncJobResult[%s] updateCANodesResultStatus failed: %v", sjr.TaskID, "NodeIPs is empty")
	}

	blog.Infof("SyncJobResult[%s] updateCANodesResultStatus %s %s", sjr.TaskID, sjr.Status.Failure, sjr.Status.Success)

	if !isSuccess {
		return sjr.updateNodeStatusByIP(sjr.ClusterID, sjr.NodeIPs, sjr.Status.Failure)
	}

	return sjr.updateNodeStatusByIP(sjr.ClusterID, sjr.NodeIPs, sjr.Status.Success)
}

func (sjr *SyncJobResult) updateAutoScalingStatus(isSuccess bool) error {
	if len(sjr.ClusterID) == 0 {
		return fmt.Errorf("SyncJobResult updateAutoScalingStatus failed: %v", "ClusterID is empty")
	}
	option, err := GetStorageModel().GetAutoScalingOption(context.Background(), sjr.ClusterID)
	if err != nil {
		return err
	}
	blog.Infof("task[%s] cluster[%s] status[%s]", sjr.TaskID, sjr.ClusterID, option.Status)
	if !option.EnableAutoscale {
		sjr.Status.Success = common.StatusAutoScalingOptionStopped
	}

	task, err := GetStorageModel().GetTask(context.Background(), sjr.TaskID)
	if err != nil {
		return err
	}

	if isSuccess {
		option.Status = sjr.Status.Success
		option.ErrorMessage = ""
	} else {
		option.Status = sjr.Status.Failure
		option.ErrorMessage = task.Message
		option.EnableAutoscale = !option.EnableAutoscale
	}
	err = GetStorageModel().UpdateAutoScalingOption(context.Background(), option)
	if err != nil {
		return err
	}

	return nil
}

func (sjr *SyncJobResult) updateNodeGroupStatus(isSuccess bool) error {
	if len(sjr.NodeGroupID) == 0 {
		return fmt.Errorf("SyncJobResult updateNodeGroupStatus failed: %v", "NodeGroupID is empty")
	}

	getStatus := func() string {
		if isSuccess {
			return sjr.Status.Success
		}

		return sjr.Status.Failure
	}

	return sjr.updateNodeGroupStatusByID(sjr.NodeGroupID, getStatus())
}

func (sjr *SyncJobResult) updateDeleteNodeGroupStatus(isSuccess bool) error {
	if len(sjr.NodeGroupID) == 0 {
		return fmt.Errorf("SyncJobResult updateDeleteNodeGroupStatus failed: %v", "NodeGroupID is empty")
	}

	getStatus := func() string {
		if isSuccess {
			return sjr.Status.Success
		}

		return sjr.Status.Failure
	}

	if !isSuccess {
		return sjr.updateNodeGroupStatusByID(sjr.NodeGroupID, getStatus())
	}

	return sjr.deleteNodeGroupByID(sjr.NodeGroupID)
}

// deleteNodeGroupByID delete nodegroup by ID
func (sjr *SyncJobResult) deleteNodeGroupByID(nodeGroupID string) error {
	group, err := GetStorageModel().GetNodeGroup(context.Background(), nodeGroupID)
	if errors.Is(err, drivers.ErrTableRecordNotFound) {
		return nil
	}
	if err != nil {
		return err
	}

	blog.Infof("task[%s] nodeGroup[%s] status[%s]", sjr.TaskID, nodeGroupID, group.Status)
	err = GetStorageModel().DeleteNodeGroup(context.Background(), nodeGroupID)
	if err != nil {
		return err
	}

	return nil
}

// updateNodeGroupStatusByID set nodeGroup status
func (sjr *SyncJobResult) updateNodeGroupStatusByID(nodeGroupID string, status string) error {
	group, err := GetStorageModel().GetNodeGroup(context.Background(), nodeGroupID)
	if err != nil {
		return err
	}
	blog.Infof("task[%s] nodeGroup[%s] status[%s]", sjr.TaskID, nodeGroupID, group.Status)
	if group.Status == status {
		return nil
	}
	group.Status = status
	err = GetStorageModel().UpdateNodeGroup(context.Background(), group)
	if err != nil {
		return err
	}

	return nil
}

func (sjr *SyncJobResult) updateNodesResultStatus(isSuccess bool) error {
	if len(sjr.NodeIPs) == 0 && len(sjr.NodeIDs) == 0 {
		return fmt.Errorf("SyncJobResult updateNodesStatus failed: %v", "NodeIPs&NodeIDs is empty")
	}

	getStatus := func() string {
		if isSuccess {
			return sjr.Status.Success
		}

		return sjr.Status.Failure
	}

	if len(sjr.NodeIPs) > 0 {
		return sjr.updateNodeStatusByIP(sjr.ClusterID, sjr.NodeIPs, getStatus())
	}

	return sjr.updateNodeStatusByNodeID(sjr.ClusterID, sjr.NodeIDs, getStatus())
}

// updateNodeStatus set node status
func (sjr *SyncJobResult) updateNodeStatusByIP(clusterID string, ipList []string, status string) error {
	if len(ipList) == 0 {
		return nil
	}

	blog.Infof("SyncJobResult[%s] updateNodeStatusByIP status[%s] ipList[%+v]", sjr.TaskID, status, ipList)

	for _, ip := range ipList {
		node, err := GetStorageModel().GetClusterNodeByIP(context.Background(), clusterID, ip)
		if err != nil {
			blog.Errorf("SyncJobResult[%s] updateNodeStatusByIP GetNodeByIP[%s] failed: %v", sjr.TaskID, ip, err)
			continue
		}
		blog.Infof("task[%s] nodeIP[%s] status[%s]", sjr.TaskID, ip, node.Status)

		if node.Status == status {
			continue
		}

		node.Status = status
		err = GetStorageModel().UpdateNode(context.Background(), node)
		if err != nil {
			blog.Errorf("SyncJobResult[%s] updateNodeStatusByIP UpdateNode[%s] failed: %v", sjr.TaskID, ip, err)
			continue
		}
	}

	return nil
}

// updateNodeStatusByNodeID set node status
func (sjr *SyncJobResult) updateNodeStatusByNodeID(clusterID string, idList []string, status string) error {
	if len(idList) == 0 {
		return nil
	}

	for _, id := range idList {
		node, err := GetStorageModel().GetClusterNode(context.Background(), clusterID, id)
		if err != nil {
			continue
		}
		blog.Infof("task[%s] nodeIP[%s] status[%s]", sjr.TaskID, id, node.Status)
		if node.Status == status {
			continue
		}

		node.Status = status
		err = GetStorageModel().UpdateNode(context.Background(), node)
		if err != nil {
			continue
		}
	}

	return nil
}

// NewJobSyncResult init SyncJobResult
func NewJobSyncResult(task *cmproto.Task) *SyncJobResult {
	return &SyncJobResult{
		TaskID:      task.TaskID,
		JobType:     JobType(task.CommonParams[JobTypeKey.String()]),
		ClusterID:   task.ClusterID,
		NodeGroupID: task.NodeGroupID,
		NodeIPs:     ParseNodeIpOrIdFromCommonMap(task.CommonParams, NodeIPsKey.String(), ","),
		NodeIDs:     ParseNodeIpOrIdFromCommonMap(task.CommonParams, NodeIDsKey.String(), ","),
	}
}
