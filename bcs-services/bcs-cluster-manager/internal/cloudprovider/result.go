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

package cloudprovider

import (
	"context"
	"fmt"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
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
	// DeleteClusterJob for deleteCluster job
	DeleteClusterJob JobType = "delete-cluster"
	// AddNodeJob for addNodes job
	AddNodeJob JobType = "add-node"
	// DeleteNodeJob for deleteNodes job
	DeleteNodeJob JobType = "delete-node"

	// CAAddNodeJob for CAAddNodes job
	CAAddNodeJob JobType = "ca-add-node"
	// CADeleteNodeJob for CADeleteNodes job
	CADeleteNodeJob JobType = "ca-delete-node"
	// CADeleteNodeGroup for CADeleteNodeGroup job
	CADeleteNodeGroup JobType = "ca-delete-nodegroup"
)

// String to string
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

// UpdateJobResultStatus update job status by result
func (sjr *SyncJobResult) UpdateJobResultStatus(isSuccess bool) error {
	if sjr == nil {
		return ErrServerIsNil
	}

	blog.Infof("task[%s] JobType[%s] isSuccess[%v] ClusterID[%s] nodeIPs[%v]",
		sjr.TaskID, sjr.JobType, isSuccess, sjr.ClusterID, sjr.NodeIPs)

	switch sjr.JobType {
	case CreateClusterJob:
		sjr.Status = generateStatusResult(common.StatusRunning, common.StatusCreateClusterFailed)
		return sjr.updateClusterResultStatus(isSuccess)
	case DeleteClusterJob:
		sjr.Status = generateStatusResult(common.StatusDeleted, common.StatusDeleteClusterFailed)
		return sjr.updateClusterResultStatus(isSuccess)
	case AddNodeJob:
		sjr.Status = generateStatusResult(common.StatusRunning, common.StatusAddNodesFailed)
		return sjr.updateNodesResultStatus(isSuccess)
	case DeleteNodeJob:
		sjr.Status = generateStatusResult("", common.StatusRemoveNodesFailed)
		return sjr.deleteNodesResultStatus(isSuccess)
	case CAAddNodeJob:
		sjr.Status = generateStatusResult(common.StatusRunning, common.StatusAddNodesFailed)
		return sjr.updateCANodesResultStatus(isSuccess)
	case CADeleteNodeJob:
		sjr.Status = generateStatusResult("", common.StatusRemoveNodesFailed)
		return sjr.deleteCANodesResultStatus(isSuccess)
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
		return fmt.Errorf("SyncJobResult deleteCANodesResultStatus failed: %v", "NodeIPs is empty")
	}

	if isSuccess {
		blog.Infof("task[%s] deleteCANodesResultStatus isSuccess[%v] InnerIPs[%v]", sjr.TaskID, isSuccess, sjr.NodeIPs)
		err := sjr.updateNodeGroupDesiredNum()
		if err != nil {
			blog.Errorf("task[%s] deleteCANodesResultStatus failed: %v", sjr.TaskID, err)
		}
		return deleteNodesByNodeIPs(sjr.NodeIPs)
	}

	return sjr.updateNodeStatusByIP(sjr.NodeIPs, sjr.Status.Failure)
}

func (sjr *SyncJobResult) updateNodeGroupDesiredNum() error {
	nodeGroupID := sjr.NodeGroupID
	if len(nodeGroupID) == 0 {
		return fmt.Errorf("task[%s] updateNodeGroupDesiredNum nodeGroupID is empty", sjr.TaskID)
	}

	group, err := GetStorageModel().GetNodeGroup(context.Background(), nodeGroupID)
	if err != nil {
		return fmt.Errorf("task[%s] updateNodeGroupDesiredNum get NodeGroup[%s] failed %s", sjr.TaskID, nodeGroupID, err.Error())
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
		return deleteNodesByNodeIPs(sjr.NodeIPs)
	}

	return sjr.updateNodeStatusByIP(sjr.NodeIPs, sjr.Status.Failure)
}

func deleteNodesByNodeIPs(nodeIPs []string) error {
	return GetStorageModel().DeleteNodesByIPs(context.Background(), nodeIPs)
}

func deleteNodesByNodeIDs(nodeIDs []string) error {
	return GetStorageModel().DeleteNodesByNodeIDs(context.Background(), nodeIDs)
}

func (sjr *SyncJobResult) updateClusterResultStatus(isSuccess bool) error {
	cluster, err := GetStorageModel().GetCluster(context.Background(), sjr.ClusterID)
	if err != nil {
		return err
	}

	cluster.Status = sjr.Status.Failure
	if isSuccess {
		cluster.Status = sjr.Status.Success
	}

	err = GetStorageModel().UpdateCluster(context.Background(), cluster)
	if err != nil {
		return err
	}

	return nil
}

func (sjr *SyncJobResult) updateCANodesResultStatus(isSuccess bool) error {
	if len(sjr.NodeIPs) == 0 {
		return fmt.Errorf("SyncJobResult updateCANodesResultStatus failed: %v", "NodeIPs is empty")
	}

	if !isSuccess {
		return sjr.updateNodeStatusByIP(sjr.NodeIPs, sjr.Status.Failure)
	}

	return sjr.updateNodeStatusByIP(sjr.NodeIPs, sjr.Status.Success)
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
		return sjr.updateNodeStatusByIP(sjr.NodeIPs, getStatus())
	}

	return sjr.updateNodeStatusByNodeID(sjr.NodeIDs, getStatus())
}

// updateNodeStatus set node status
func (sjr *SyncJobResult) updateNodeStatusByIP(ipList []string, status string) error {
	if len(ipList) == 0 {
		return nil
	}

	for _, ip := range ipList {
		node, err := GetStorageModel().GetNodeByIP(context.Background(), ip)
		if err != nil {
			continue
		}
		blog.Infof("task[%s] nodeIP[%s] status[%s]", sjr.TaskID, ip, node.Status)
		if node.Status == status || utils.StringInSlice(node.Status, []string{common.StatusAddNodesFailed, common.StatusRunning}) {
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

// updateNodeStatusByNodeID set node status
func (sjr *SyncJobResult) updateNodeStatusByNodeID(idList []string, status string) error {
	if len(idList) == 0 {
		return nil
	}

	for _, id := range idList {
		node, err := GetStorageModel().GetNode(context.Background(), id)
		if err != nil {
			continue
		}
		blog.Infof("task[%s] nodeIP[%s] status[%s]", sjr.TaskID, id, node.Status)
		if node.Status == status || utils.StringInSlice(node.Status, []string{common.StatusAddNodesFailed, common.StatusRunning}) {
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
		JobType:     JobType(task.CommonParams["JobType"]),
		ClusterID:   task.ClusterID,
		NodeGroupID: task.NodeGroupID,
		NodeIPs:     strings.Split(task.CommonParams["NodeIPs"], ","),
		NodeIDs:     strings.Split(task.CommonParams["NodeIDs"], ","),
	}
}
