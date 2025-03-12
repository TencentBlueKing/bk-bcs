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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/clients/cluster"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/clients/helm"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/store"
	fedtasks "github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/task/tasks"
	federationmgr "github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/proto/bcs-federation-manager"
)

// InstallFederation install federation
func (f *FederationManager) InstallFederation(ctx context.Context,
	req *federationmgr.InstallFederationRequest, resp *federationmgr.InstallFederationResponse) error {

	blog.Infof("Received BcsFederationManager.InstallFederation request, clusterId: %v, creator: %v, lbId: %v",
		req.GetClusterId(), req.GetCreator(), req.GetLoadBalancerId())

	// validate
	if err := req.Validate(); err != nil {
		return ErrReturn(resp, fmt.Sprintf("validate install federation request failed, err: %s", err.Error()))
	}

	// check if cluster is installing
	installingCluster, installingTask, err := f.findInstallingFederationCluster(ctx, req.GetClusterId())
	if err != nil {
		return ErrReturn(resp, fmt.Sprintf("get installing cluster failed, err: %s", err.Error()))
	}
	if installingCluster != nil {
		if installingTask != nil {
			resp.Data = &federationmgr.TaskDistributeResponseData{
				TaskId: installingTask.TaskId,
			}
		}
		return ErrReturn(resp, fmt.Sprintf("cluster %s installing federation task already existed", req.GetClusterId()))
	}

	// get basic info for cluster
	cluster, err := f.clusterCli.GetCluster(ctx, req.GetClusterId())
	if err != nil {
		return ErrReturn(resp, fmt.Sprintf("GetCluster error when install federation, err: %s", err.Error()))
	}

	var userToken string
	if req.GetUserToken() != "" {
		userToken = req.GetUserToken()
	} else {
		userToken, err = f.userCli.GetUserToken(req.GetCreator())
		if err != nil {
			return ErrReturn(resp, fmt.Sprintf("GetUserToken error when install federation for user[%s], err: %s", req.GetCreator(), err.Error()))
		}
	}

	// LoadBalancer Id support lb-xxxxx and subnet-xxxxx
	var lbId string
	if req.GetLoadBalancerId() != "" {
		lbId = req.GetLoadBalancerId()
	} else {
		lbId, err = f.clusterCli.GetSubnetId(ctx, req.GetClusterId())
		if err != nil {
			return ErrReturn(resp, fmt.Sprintf("GetLbId error when install federation for cluster[%s], err: %s", req.GetClusterId(), err.Error()))
		}
	}

	// build task for install federation for cluster
	t, err := fedtasks.NewInstallFederationTask(&fedtasks.InstallFederationOptions{
		ProjectId:                    cluster.ProjectID,
		FederationBusinessId:         req.GetFederationBusinessId(),
		FederationProjectId:          req.GetFederationProjectId(),
		FederationProjectCode:        req.GetFederationProjectCode(),
		FederationClusterName:        req.GetFederationClusterName(),
		FederationClusterEnv:         req.GetFederationClusterEnv(),
		FederationClusterDescription: req.GetFederationClusterDescription(),
		FederationClusterLabels:      req.GetFederationClusterLabels(),
		ClusterId:                    req.GetClusterId(),
		UserToken:                    userToken,
		LbId:                         lbId,
	}).BuildTask(req.GetCreator())
	if err != nil {
		return ErrReturn(resp, fmt.Sprintf("BuildTask error when install federation, err: %s", err.Error()))
	}

	// dispatch task
	if err = f.taskmanager.Dispatch(t); err != nil {
		return ErrReturn(resp, fmt.Sprintf("Dispatch federation install task failed, err: %s", err.Error()))
	}

	// success
	resp.Code = IntToUint32Ptr(common.BcsSuccess)
	resp.Message = common.BcsSuccessStr
	resp.Data = &federationmgr.TaskDistributeResponseData{
		TaskId: t.GetTaskID(),
	}
	return nil
}

// UnInstallFederation un install federation
func (f *FederationManager) UnInstallFederation(ctx context.Context,
	req *federationmgr.UnInstallFederationRequest, resp *federationmgr.UnInstallFederationResponse) error {

	blog.Infof("Received BcsFederationManager.UnInstallFederation request, clusterId: %v, operator: %v", req.GetClusterId(), req.GetOperator())

	// validate
	if err := req.Validate(); err != nil {
		return ErrReturn(resp, fmt.Sprintf("validate UnInstallFederation request failed, err: %s", err.Error()))
	}

	federationClusterId := req.GetClusterId()

	// check exist
	fedCluster, err := f.store.GetFederationCluster(ctx, federationClusterId)
	if err != nil {
		return ErrReturn(resp, fmt.Sprintf("GetFederationCluster failed, clusterId: %s, err: %s", req.GetClusterId(), err.Error()))
	}

	// 1.check sub cluster exist
	subClusters, err := f.store.ListSubClusters(ctx, &store.SubClusterListOptions{
		FederationClusterID: fedCluster.FederationClusterID,
	})
	if err != nil {
		return ErrReturn(resp, fmt.Sprintf("ListSubClusters failed, clusterId: %s, err: %s", req.GetClusterId(), err.Error()))
	}
	if len(subClusters) > 0 {
		return ErrReturn(resp, fmt.Sprintf("cluster %s still has %d sub cluster, can not uninstall federation", req.GetClusterId(), len(subClusters)))
	}

	// 2.get host cluster and uninstall federation modules
	hostCluster, err := f.clusterCli.GetCluster(ctx, fedCluster.HostClusterID)
	if err != nil {
		return ErrReturn(resp, fmt.Sprintf("Get federation cluster[%s]'s host cluster[%s] failed, err: %s", federationClusterId, fedCluster.HostClusterID, err.Error()))
	}

	// uninstall bcs-unified-apiserver
	// NOCC:vetshadow/shadow(设计如此:这里err可以被覆盖)
	if err := f.helmCli.UninstallUnifiedApiserver(&helm.BcsUnifiedApiserverOptions{
		ReleaseBaseOptions: helm.ReleaseBaseOptions{
			ProjectID: hostCluster.GetProjectID(),
			ClusterID: hostCluster.GetClusterID(),
		},
	}); err != nil {
		return ErrReturn(resp, fmt.Sprintf("Uninstall bcs-unified-apiserver failed, clusterId: %s, err: %s", req.GetClusterId(), err.Error()))
	}

	// uninstall bcs-clusternet-controller
	// NOCC:vetshadow/shadow(设计如此:这里err可以被覆盖)
	if err := f.helmCli.UninstallClusternetController(&helm.ReleaseBaseOptions{
		ProjectID: hostCluster.GetProjectID(),
		ClusterID: hostCluster.GetClusterID(),
	}); err != nil {
		return ErrReturn(resp, fmt.Sprintf("Uninstall bcs-clusternet-controller failed, clusterId: %s, err: %s", req.GetClusterId(), err.Error()))
	}

	// uninstall bcs-clusternet-scheduler
	// NOCC:vetshadow/shadow(设计如此:这里err可以被覆盖)
	if err := f.helmCli.UninstallClusternetScheduler(&helm.ReleaseBaseOptions{
		ProjectID: hostCluster.GetProjectID(),
		ClusterID: hostCluster.GetClusterID(),
	}); err != nil {
		return ErrReturn(resp, fmt.Sprintf("Uninstall bcs-clusternet-scheduler failed, clusterId: %s, err: %s", req.GetClusterId(), err.Error()))
	}

	// uninstall bcs-clusternet-hub
	// NOCC:vetshadow/shadow(设计如此:这里err可以被覆盖)
	if err := f.helmCli.UninstallClusternetHub(&helm.ReleaseBaseOptions{
		ProjectID: hostCluster.GetProjectID(),
		ClusterID: hostCluster.GetClusterID(),
	}); err != nil {
		return ErrReturn(resp, fmt.Sprintf("Uninstall bcs-clusternet-hub failed, clusterId: %s, err: %s", req.GetClusterId(), err.Error()))
	}

	// 3.delete federation cluster from cluster manager
	// NOCC:vetshadow/shadow(设计如此:这里err可以被覆盖)
	if err := f.clusterCli.DeleteFederationCluster(ctx, fedCluster.FederationClusterID, req.Operator); err != nil {
		return ErrReturn(resp, fmt.Sprintf("Delete federation cluster failed, clusterId: %s, err: %s", req.GetClusterId(), err.Error()))
	}

	// 4.delete federation cluster from store
	// NOCC:vetshadow/shadow(设计如此:这里err可以被覆盖)
	if err := f.store.DeleteFederationCluster(ctx, &store.FederationClusterDeleteOptions{
		FederationClusterID: fedCluster.FederationClusterID,
		Updater:             req.GetOperator(),
	}); err != nil {
		return ErrReturn(resp, fmt.Sprintf("Delete federation cluster from store failed, clusterId: %s, err: %s", req.GetClusterId(), err.Error()))
	}

	// 5.delete host cluster federation label
	if err := f.clusterCli.DeleteHostClusterLabel(context.Background(), fedCluster.HostClusterID); err != nil {
		return ErrReturn(resp, fmt.Sprintf("Delete host cluster label failed, clusterId: %s, err: %s", req.GetClusterId(), err.Error()))
	}

	// success
	resp.Code = IntToUint32Ptr(common.BcsSuccess)
	resp.Message = common.BcsSuccessStr
	resp.Data = nil
	return nil
}

// RegisterSubcluster register subcluster
func (f *FederationManager) RegisterSubcluster(ctx context.Context,
	req *federationmgr.RegisterSubclusterRequest, resp *federationmgr.RegisterSubclusterResponse) error {

	blog.Infof("Received BcsFederationManager.RegisterSubcluster request, clusterId: %v, creator: %v, subclusterId: %v",
		req.GetClusterId(), req.GetCreator(), req.GetSubclusterId())

	// validate
	if err := req.Validate(); err != nil {
		return ErrReturn(resp, fmt.Sprintf("validate register subcluster request failed, err: %s", err.Error()))
	}

	// get basic info for cluster
	cluster, err := f.clusterCli.GetCluster(ctx, req.GetClusterId())
	if err != nil {
		return ErrReturn(resp, fmt.Sprintf("GetCluster error when register subcluster, err: %s", err.Error()))
	}

	// check subclusterId is valid
	_, err = f.clusterCli.GetCluster(ctx, req.GetSubclusterId())
	if err != nil {
		return ErrReturn(resp, fmt.Sprintf("GetCluster error when register subcluster, err: %s", err.Error()))
	}

	// get user userToken
	var userToken string
	if req.GetUserToken() != "" {
		userToken = req.GetUserToken()
	} else {
		userToken, err = f.userCli.GetUserToken(req.GetCreator())
		if err != nil {
			return ErrReturn(resp, fmt.Sprintf("GetUserToken error when register subcluster for user[%s], err: %s", req.GetCreator(), err.Error()))
		}
	}

	t, err := fedtasks.NewRegisterSubclusterTask(&fedtasks.RegisterSubclusterOptions{
		ProjectId:      cluster.ProjectID,
		ClusterId:      req.GetClusterId(),
		SubClusterId:   req.GetSubclusterId(),
		UserToken:      userToken,
		GatewayAddress: f.bcsGateWay.Endpoint,
	}).BuildTask(req.GetCreator())
	if err != nil {
		return ErrReturn(resp, fmt.Sprintf("BuildTask error when register subcluster, err: %s", err.Error()))
	}

	if err = f.taskmanager.Dispatch(t); err != nil {
		return ErrReturn(resp, fmt.Sprintf("Dispatch register subcluster task failed, err: %s", err.Error()))
	}

	// success
	resp.Code = IntToUint32Ptr(common.BcsSuccess)
	resp.Message = common.BcsSuccessStr
	resp.Data = &federationmgr.TaskDistributeResponseData{
		TaskId: t.GetTaskID(),
	}
	return nil
}

// RemoveSubcluster remove subcluster from federation cluster
func (f *FederationManager) RemoveSubcluster(ctx context.Context,
	req *federationmgr.RemoveSubclusterRequest, resp *federationmgr.RemoveSubclusterResponse) error {

	blog.Infof("Received BcsFederationManager.RemoveSubcluster request, clusterId: %v, creator: %v, subclusterId: %v",
		req.GetClusterId(), req.GetUser(), req.GetSubclusterId())

	// validate
	if err := req.Validate(); err != nil {
		return ErrReturn(resp, fmt.Sprintf("validate remove subcluster request failed, err: %s", err.Error()))
	}

	// get basic info for fedCluster
	fedCluster, err := f.clusterCli.GetCluster(ctx, req.GetClusterId())
	if err != nil {
		return ErrReturn(resp, fmt.Sprintf("GetCluster error when remove subcluster, err: %s", err.Error()))
	}

	t, err := fedtasks.NewRemoveSubclusterTask(&fedtasks.RemoveSubclusterOptions{
		ProjectId:    fedCluster.ProjectID,
		ClusterId:    req.GetClusterId(),
		SubClusterId: req.GetSubclusterId(),
	}).BuildTask(req.GetUser())
	if err != nil {
		return ErrReturn(resp, fmt.Sprintf("BuildTask error when remove subcluster, err: %s", err.Error()))
	}

	// update sub cluster status and taskId
	subCluster, err := f.store.GetSubCluster(ctx, req.GetClusterId(), req.GetSubclusterId())
	if err != nil {
		return ErrReturn(resp, fmt.Sprintf("GetSubCluster error when remove subcluster, err: %s", err.Error()))
	}
	subCluster.Status = store.DeletingStatus
	subCluster.Labels[cluster.FederationClusterTaskIDLabelKey] = t.GetTaskID()
	// NOCC:vetshadow/shadow(设计如此:这里err可以被覆盖)
	if err := f.store.UpdateSubCluster(ctx, subCluster, req.GetUser()); err != nil {
		return ErrReturn(resp, fmt.Sprintf("UpdateSubCluster error when remove subcluster, err: %s", err.Error()))
	}

	if err = f.taskmanager.Dispatch(t); err != nil {
		return ErrReturn(resp, fmt.Sprintf("Dispatch remove subcluster task failed, err: %s", err.Error()))
	}

	// success
	resp.Code = IntToUint32Ptr(common.BcsSuccess)
	resp.Message = common.BcsSuccessStr
	resp.Data = &federationmgr.TaskDistributeResponseData{
		TaskId: t.GetTaskID(),
	}
	return nil
}
