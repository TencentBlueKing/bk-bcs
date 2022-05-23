/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package nodegroup

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
)

// CreateAction action for create nodeGroup
type CreateAction struct {
	ctx   context.Context
	model store.ClusterManagerModel
	req   *cmproto.CreateNodeGroupRequest
	resp  *cmproto.CreateNodeGroupResponse

	cluster *cmproto.Cluster
	cloud   *cmproto.Cloud
}

// NewCreateAction create namespace action
func NewCreateAction(model store.ClusterManagerModel) *CreateAction {
	return &CreateAction{
		model: model,
	}
}

func (ca *CreateAction) getRelativeResource() error {
	//get relative cluster for information injection
	cluster, err := ca.model.GetCluster(ca.ctx, ca.req.ClusterID)
	if err != nil {
		blog.Errorf("can not get relative Cluster %s when create NodeGroup", ca.req.ClusterID)
		return fmt.Errorf("get relative cluster %s info err, %s", ca.req.ClusterID, err.Error())
	}
	ca.cluster = cluster

	// clusterManager and nodeGroup is different, clusterManager tencent_cloud, nodeGroup yunti
	// if nodeGroup provider is null, use cluster provider
	if len(ca.req.Provider) == 0 {
		ca.req.Provider = cluster.Provider
	}

	cloud, err := actions.GetCloudByCloudID(ca.model, ca.req.Provider)
	if err != nil {
		blog.Errorf("can not get relative Cloud %s when create NodeGroup for Cluster %s, %s",
			ca.req.Provider, ca.req.ClusterID, err.Error(),
		)
		return err
	}
	ca.cloud = cloud

	return nil
}

func (ca *CreateAction) constructNodeGroup() *cmproto.NodeGroup {
	timeStr := time.Now().Format(time.RFC3339)
	group := &cmproto.NodeGroup{
		Name:            ca.req.Name,
		ClusterID:       ca.req.ClusterID,
		Region:          ca.req.Region,
		ProjectID:       ca.cluster.ProjectID,
		EnableAutoscale: ca.req.EnableAutoscale,
		AutoScaling:     ca.req.AutoScaling,
		LaunchTemplate:  ca.req.LaunchTemplate,
		Labels:          ca.req.Labels,
		Taints:          ca.req.Taints,
		NodeOS:          ca.req.NodeOS,
		Provider:        ca.req.Provider,
		ConsumerID:      ca.req.ConsumerID,
		Creator:         ca.req.Creator,
		CreateTime:      timeStr,
		UpdateTime:      timeStr,
	}
	return group
}

func (ca *CreateAction) setResp(code uint32, msg string) {
	ca.resp.Code = code
	ca.resp.Message = msg
	ca.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

// Handle create nodeGroup request
func (ca *CreateAction) Handle(ctx context.Context,
	req *cmproto.CreateNodeGroupRequest, resp *cmproto.CreateNodeGroupResponse) {
	if req == nil || resp == nil {
		blog.Errorf("create NodeGroup failed, req or resp is empty")
		return
	}
	ca.ctx = ctx
	ca.req = req
	ca.resp = resp

	if err := req.Validate(); err != nil {
		ca.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	// getRelativeResource get cluster / project/ cloud provider
	if err := ca.getRelativeResource(); err != nil {
		ca.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}
	group := ca.constructNodeGroup()

	// create nodegroup with cloudprovider
	mgr, err := cloudprovider.GetNodeGroupMgr(ca.cloud.CloudProvider)
	if err != nil {
		blog.Errorf("get NodeGroup Manager cloudprovider %s/%s for create nodegroup in Cluster %s failed, %s",
			ca.cloud.CloudID, ca.cloud.CloudProvider, ca.cluster.ClusterID, err.Error(),
		)
		ca.setResp(common.BcsErrClusterManagerCloudProviderErr, err.Error())
		return
	}
	cmOption, err := cloudprovider.GetCredential(&cloudprovider.CredentialData{
		Cloud:     ca.cloud,
		AccountID: ca.cluster.CloudAccountID,
	})
	if err != nil {
		blog.Errorf("get Credential for Cloud %s/%s when create NodeGroup for cluster %s failed, %s",
			ca.cloud.CloudID, ca.cloud.CloudProvider, ca.cluster.ClusterID, err.Error(),
		)
		ca.setResp(common.BcsErrClusterManagerCloudProviderErr, err.Error())
		return
	}
	cmOption.Region = group.Region

	// cloud provider nodeGroup
	if err = mgr.CreateNodeGroup(group, &cloudprovider.CreateNodeGroupOption{
		CommonOption: *cmOption,
	}); err != nil {
		blog.Errorf("create NodeGroup in cloudprovider %s/%s for Cluster %s failed, %s",
			ca.cloud.CloudID, ca.cloud.CloudProvider, ca.cluster.ClusterID, err.Error(),
		)
		ca.setResp(common.BcsErrClusterManagerCloudProviderErr, err.Error())
		return
	}

	// finally store NodeGroup information to DB
	if err = ca.model.CreateNodeGroup(ca.ctx, group); err != nil {
		blog.Errorf("store nodegroup %+v information to DB failed, %s", group, err.Error())
		if errors.Is(err, drivers.ErrTableRecordDuplicateKey) {
			ca.setResp(common.BcsErrClusterManagerDatabaseRecordDuplicateKey, err.Error())
			return
		}
		ca.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}
	blog.Infof("create nodegroup %s information for Cluster %s to DB successfully", group, ca.cluster.ClusterID)

	err = ca.model.CreateOperationLog(ca.ctx, &cmproto.OperationLog{
		ResourceType: common.NodeGroup.String(),
		ResourceID:   group.NodeGroupID,
		TaskID:       "",
		Message:      fmt.Sprintf("集群%s创建节点池%s", ca.cluster.ClusterID, group.NodeGroupID),
		OpUser:       req.Creator,
		CreateTime:   time.Now().String(),
	})
	if err != nil {
		blog.Errorf("CreateNodeGroup[%s] CreateOperationLog failed: %v", ca.cluster.ClusterID, err)
	}
	ca.resp.Data = group
	ca.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
	return
}
