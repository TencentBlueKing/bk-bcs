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

package cloudvpc

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
)

// ListSubnetsAction action for list subnets
type ListSubnetsAction struct {
	ctx     context.Context
	cloud   *cmproto.Cloud
	cluster *cmproto.Cluster
	model   store.ClusterManagerModel
	req     *cmproto.ListSubnetsRequest
	resp    *cmproto.ListSubnetsResponse
	subnets []*cmproto.Subnet
}

// NewListSubnetsAction create list action for subnets
func NewListSubnetsAction(model store.ClusterManagerModel) *ListSubnetsAction {
	return &ListSubnetsAction{
		model: model,
	}
}

func (la *ListSubnetsAction) validate() error {
	if err := la.req.Validate(); err != nil {
		return err
	}
	// get cluster basic info(project/cluster/cloud)
	err := la.getClusterBasicInfo()
	if err != nil {
		return err
	}
	return nil
}

// getCloudProjectInfo get cluster/cloud/project info
func (la *ListSubnetsAction) getClusterBasicInfo() error {
	cluster, err := la.model.GetCluster(la.ctx, la.req.ClusterID)
	if err != nil {
		blog.Errorf("get Cluster %s failed when ListNodeType, %s", la.req.ClusterID, err.Error())
		return err
	}
	la.cluster = cluster

	cloud, err := actions.GetCloudByCloudID(la.model, la.cluster.Provider)
	if err != nil {
		blog.Errorf("get Cluster %s provider %s failed, %s",
			la.cluster.ClusterID, la.cluster.Provider, err.Error(),
		)
		return err
	}
	la.cloud = cloud

	return nil
}

func (la *ListSubnetsAction) setResp(code uint32, msg string) {
	la.resp.Code = code
	la.resp.Message = msg
	la.resp.Result = (code == common.BcsErrClusterManagerSuccess)
	la.resp.Data = la.subnets
}

// Handle handle list vpc subnets
func (la *ListSubnetsAction) Handle(
	ctx context.Context, req *cmproto.ListSubnetsRequest, resp *cmproto.ListSubnetsResponse) {
	if req == nil || resp == nil {
		blog.Errorf("list subnets failed, req or resp is empty")
		return
	}
	la.ctx = ctx
	la.req = req
	la.resp = resp

	if err := req.Validate(); err != nil {
		la.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	// create vpc client with cloudProvider
	vpcMgr, err := cloudprovider.GetVPCMgr(la.cloud.CloudProvider)
	if err != nil {
		blog.Errorf("get cloudprovider %s VPCManager for list subnets in cluster %s failed, %s",
			la.cloud.CloudProvider, la.req.ClusterID, err.Error(),
		)
		la.setResp(common.BcsErrClusterManagerCloudProviderErr, err.Error())
		return
	}
	cmOption, err := cloudprovider.GetCredential(&cloudprovider.CredentialData{
		Cloud:     la.cloud,
		AccountID: la.cluster.CloudAccountID,
	})
	if err != nil {
		blog.Errorf("get credential for cloudprovider %s/%s list subnets in cluster %s failed, %s",
			la.cloud.CloudID, la.cloud.CloudProvider, la.req.ClusterID, err.Error(),
		)
		la.setResp(common.BcsErrClusterManagerCloudProviderErr, err.Error())
		return
	}
	cmOption.Region = la.cluster.Region

	// get subnet list
	subnets, err := vpcMgr.ListSubnets(la.req.VpcID, cmOption)
	if err != nil {
		la.setResp(common.BcsErrClusterManagerCloudProviderErr, err.Error())
		return
	}
	la.subnets = subnets
	la.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
	return
}
