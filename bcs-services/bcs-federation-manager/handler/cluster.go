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

// Package handler xxx
package handler

import (
	"context"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	taskstore "github.com/Tencent/bk-bcs/bcs-common/common/task/store"
	tasktypes "github.com/Tencent/bk-bcs/bcs-common/common/task/types"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/store"
	fedsteps "github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/task/steps"
	installsteps "github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/task/steps/install_federation_steps"
	fedtasks "github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/task/tasks"
	federationmgr "github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/proto/bcs-federation-manager"
)

// GetFederationCluster get federation cluster
func (f *FederationManager) GetFederationCluster(ctx context.Context,
	req *federationmgr.GetFederationClusterRequest, resp *federationmgr.GetFederationClusterResponse) error {

	blog.Infof("Receive GetFederationCluster request, clusterId: %s", req.GetClusterId())

	if err := req.Validate(); err != nil {
		return ErrReturn(resp, fmt.Sprintf("validate GetFederationCluster request failed, err: %s", err.Error()))
	}

	fedCluster, err := f.store.GetFederationCluster(ctx, req.GetClusterId())
	if err != nil {
		return ErrReturn(resp, err.Error())
	}

	subClusters, err := f.store.ListSubClusters(ctx, &store.SubClusterListOptions{
		FederationClusterID: fedCluster.FederationClusterID,
	})
	if err != nil {
		return ErrReturn(resp, err.Error())
	}

	fedNamespaces, err := f.clusterCli.ListFederationNamespaces(fedCluster.FederationClusterID)
	if err != nil {
		return ErrReturn(resp, err.Error())
	}

	resp.Code = IntToUint32Ptr(common.BcsSuccess)
	resp.Message = common.BcsSuccessStr
	resp.Data = TransferFedCluster(fedCluster, subClusters, fedNamespaces)
	return nil
}

// ListProjectFederation list project federation
func (f *FederationManager) ListProjectFederation(ctx context.Context,
	req *federationmgr.ListProjectFederationRequest, resp *federationmgr.ListProjectFederationResponse) error {

	blog.Infof("Receive ListProjectFederation request, projectID: %s", req.GetProjectId())

	if err := req.Validate(); err != nil {
		return ErrReturn(resp, fmt.Sprintf("validate ListProjectFederation request failed, err: %s", err.Error()))
	}

	// list federation clusters from store
	fedClusters, err := f.store.ListFederationClusters(ctx, &store.FederationListOptions{
		Conditions: map[string]string{
			"project_id": req.GetProjectId(),
		}})
	if err != nil {
		return ErrReturn(resp, err.Error())
	}

	// transfer
	clusters := make([]*federationmgr.FederationCluster, 0)
	for _, fedCluster := range fedClusters {
		clusters = append(clusters, TransferFedCluster(fedCluster, nil, nil))
	}

	resp.Code = IntToUint32Ptr(common.BcsSuccess)
	resp.Message = common.BcsSuccessStr
	resp.Data = clusters
	return nil
}

// ListProjectInstallingFederation list project federation which is installing
func (f *FederationManager) ListProjectInstallingFederation(ctx context.Context,
	req *federationmgr.ListProjectInstallingFederationRequest, resp *federationmgr.ListProjectInstallingFederationResponse) error {

	blog.Infof("Receive ListProjectFederationInstalling request, projectID: %s", req.GetProjectId())

	if err := req.Validate(); err != nil {
		return ErrReturn(resp, fmt.Sprintf("validate ListProjectFederationInstalling request failed, err: %s", err.Error()))
	}

	// get all clusters from cluster manager
	clusters, err := f.clusterCli.ListProjectCluster(ctx, req.GetProjectId())
	if err != nil {
		return ErrReturn(resp, err.Error())
	}

	// check all clusters wheather is installing
	clsWithTasks := make([]*federationmgr.FederationClusterWithTask, 0)
	for _, cls := range clusters {
		// get cluster from store
		fedCluster, t, err := f.findInstallingFederationCluster(ctx, cls.ClusterID)
		if err != nil {
			return ErrReturn(resp, err.Error())
		}
		if fedCluster != nil {
			clsWithTask := &federationmgr.FederationClusterWithTask{
				Cluster: fedCluster,
				Task:    t,
			}
			clsWithTasks = append(clsWithTasks, clsWithTask)
		}
	}

	resp.Code = IntToUint32Ptr(common.BcsSuccess)
	resp.Data = clsWithTasks
	resp.Message = common.BcsSuccessStr
	return nil
}

// GetFederationByHostCluster get federation by host cluster
func (f *FederationManager) GetFederationByHostCluster(ctx context.Context,
	req *federationmgr.GetFederationByHostClusterRequest, resp *federationmgr.GetFederationByHostClusterResponse) error {

	blog.Infof("Received GetFederationByHostCluster request, clusterId: %v",
		req.GetClusterId())

	fedClusters, err := f.store.ListFederationClusters(ctx, &store.FederationListOptions{
		Conditions: map[string]string{
			"host_cluster_id": req.GetClusterId(),
		}})
	if err != nil {
		return ErrReturn(resp, err.Error())
	}

	// existed fed cluster for host cluster
	if len(fedClusters) != 0 {
		resp.Code = IntToUint32Ptr(common.BcsSuccess)
		resp.Message = common.BcsSuccessStr
		resp.Data = &federationmgr.FederationClusterWithTask{
			Cluster: TransferFedCluster(fedClusters[0], nil, nil),
			Task:    nil,
		}
		return nil
	}

	// try to get in processing cluster
	cls, task, err := f.findInstallingFederationCluster(ctx, req.GetClusterId())
	if err != nil {
		return ErrReturn(resp, err.Error())
	}
	// not found in processing cluster
	if cls == nil {
		return ErrReturn(resp, fmt.Sprintf("not found federation cluster for host cluster %s", req.GetClusterId()))
	}

	// get cluster from task
	resp.Code = IntToUint32Ptr(common.BcsSuccess)
	resp.Message = common.BcsSuccessStr
	resp.Data = &federationmgr.FederationClusterWithTask{
		Cluster: cls,
		Task:    task,
	}
	return nil
}

// findInstallingFederationCluster get cluster which is in installing
func (f *FederationManager) findInstallingFederationCluster(ctx context.Context, hostClusterId string) (
	*federationmgr.FederationCluster, *federationmgr.Task, error) {

	// can not find existed fed cluster, check task list for installing
	cond := operator.NewBranchCondition(operator.And,
		operator.NewLeafCondition(
			operator.Eq,
			operator.M{
				"taskType":  fedtasks.InstallFederationTaskName.Type,
				"taskIndex": hostClusterId,
			}),
		operator.NewLeafCondition(
			operator.In,
			operator.M{
				"status": []string{
					// only in processing tasks
					tasktypes.TaskStatusInit,
					tasktypes.TaskStatusRunning,
					tasktypes.TaskStatusNotStarted,
					tasktypes.TaskStatusForceTerminate,
				}},
		),
	)

	opt := &taskstore.ListOption{
		Offset: DefaultTaskListOffset,
		Limit:  1,
		Sort: map[string]int{
			"start": -1,
		},
	}

	// check whether there is a task in the installation process
	taskList, err := f.taskmanager.ListTask(ctx, cond, opt)
	if err != nil {
		return nil, nil, err
	}
	if len(taskList) == 0 {
		// not found installing cluster in tasks
		// which means cluster has no task for installation in process
		return nil, nil, nil
	}

	t := &taskList[0]
	cls, err := constractClusterFromInstallingTask(t)
	if err != nil {
		return nil, nil, fmt.Errorf("constractClusterFromInstallingTask error, err %s", err.Error())
	}

	return cls, transferTask(t), nil
}

func constractClusterFromInstallingTask(t *tasktypes.Task) (*federationmgr.FederationCluster, error) {
	if t.GetTaskType() != fedtasks.InstallFederationTaskName.Type {
		return nil, fmt.Errorf("task type is not %s", fedtasks.InstallFederationTaskName.Type)
	}

	hostClusterId, ok := t.GetCommonParams(fedsteps.ClusterIdKey)
	if !ok {
		return nil, fmt.Errorf("not found %s key from task %s", fedsteps.ClusterIdKey, t.GetTaskID())
	}

	registerStep, ok := t.GetStep(installsteps.RegisterClusterStepName.Name)
	if !ok {
		return nil, fmt.Errorf("not found %s step from task %s", installsteps.RegisterClusterStepName.Name, t.GetTaskID())
	}

	registerParams := registerStep.GetParamsAll()
	cluster := &federationmgr.FederationCluster{
		// federation cluster id has not import to clustermanager, use temporary id
		FederationClusterId:   formatTempFederationClusterId(hostClusterId),
		FederationClusterName: registerParams[fedsteps.FederationClusterNameKey],
		HostClusterId:         hostClusterId,
		ProjectCode:           registerParams[fedsteps.FederationProjectCodeKey],
		ProjectId:             registerParams[fedsteps.FederationProjectIdKey],
		CreatedTime:           "",
		UpdatedTime:           "",
		Status:                mapTaskStatus2ClusterStatus(t.GetStatus()),
		SubClusters:           nil,
		FederationNamespaces:  nil,
	}

	return cluster, nil
}

func formatTempFederationClusterId(hostclusterId string) string {
	return fmt.Sprintf("%s-%s", hostclusterId, "federation")
}

func mapTaskStatus2ClusterStatus(taskStatus string) string {
	switch taskStatus {
	case tasktypes.TaskStatusFailure, tasktypes.TaskStatusTimeout, tasktypes.TaskStatusForceTerminate:
		return store.CreateFailedStatus
	case tasktypes.TaskStatusInit, tasktypes.TaskStatusRunning, tasktypes.TaskStatusNotStarted:
		return store.CreatingStatus
	default:
		return store.UnknownStatus
	}
}
