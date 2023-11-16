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

package cloudresource

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
)

// ListCloudProjectsAction list action for projects
type ListCloudProjectsAction struct {
	ctx context.Context

	cloud   *cmproto.Cloud
	account *cmproto.CloudAccount

	model       store.ClusterManagerModel
	req         *cmproto.ListCloudProjectsRequest
	resp        *cmproto.ListCloudProjectsResponse
	projectList []*cmproto.CloudProject
}

// NewListCloudProjectsAction create list action for projects
func NewListCloudProjectsAction(model store.ClusterManagerModel) *ListCloudProjectsAction {
	return &ListCloudProjectsAction{
		model: model,
	}
}

func (la *ListCloudProjectsAction) validate() error {
	if err := la.req.Validate(); err != nil {
		return err
	}
	err := la.getRelativeData()
	if err != nil {
		return err
	}

	return nil
}

func (la *ListCloudProjectsAction) getRelativeData() error {
	cloud, err := actions.GetCloudByCloudID(la.model, la.req.CloudID)
	if err != nil {
		return err
	}
	la.cloud = cloud

	if la.req.AccountID != "" {
		account, err := la.model.GetCloudAccount(la.ctx, la.req.CloudID, la.req.AccountID, false)
		if err != nil {
			return err
		}

		la.account = account
	}

	return nil
}

func (la *ListCloudProjectsAction) setResp(code uint32, msg string) {
	la.resp.Code = code
	la.resp.Message = msg
	la.resp.Result = (code == common.BcsErrClusterManagerSuccess)
	la.resp.Data = la.projectList
}

func (la *ListCloudProjectsAction) listCloudProjects() error {
	clsMgr, err := cloudprovider.GetClusterMgr(la.cloud.CloudProvider)
	if err != nil {
		blog.Errorf("get cloudprovider %s clusterManager for "+
			"list projects failed, %s", la.cloud.CloudProvider, err.Error())
		return err
	}
	cmOption, err := cloudprovider.GetCredential(&cloudprovider.CredentialData{
		Cloud:     la.cloud,
		AccountID: la.req.AccountID,
	})
	if err != nil {
		blog.Errorf("get credential for cloudprovider %s/%s list projects failed, %s",
			la.cloud.CloudID, la.cloud.CloudProvider, err.Error())
		return err
	}
	cmOption.Region = la.req.Region

	// get project list
	projectList, err := clsMgr.ListProjects(cmOption)
	if err != nil {
		return err
	}
	la.projectList = projectList
	return nil
}

// Handle handle list image os request
func (la *ListCloudProjectsAction) Handle(ctx context.Context,
	req *cmproto.ListCloudProjectsRequest, resp *cmproto.ListCloudProjectsResponse) {
	if req == nil || resp == nil {
		blog.Errorf("list cloud projects failed, req or resp is empty")
		return
	}
	la.ctx = ctx
	la.req = req
	la.resp = resp

	if err := la.validate(); err != nil {
		la.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	if err := la.listCloudProjects(); err != nil {
		la.setResp(common.BcsErrClusterManagerCloudProviderErr, err.Error())
		return
	}

	la.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}
