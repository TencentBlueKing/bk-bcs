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

package project

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	storeopt "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/options"
)

// ListAction action for list online cluster credential
type ListAction struct {
	ctx         context.Context
	model       store.ClusterManagerModel
	req         *cmproto.ListProjectRequest
	resp        *cmproto.ListProjectResponse
	projectList []*cmproto.Project
}

// NewListAction create list action for cluster credential
func NewListAction(model store.ClusterManagerModel) *ListAction {
	return &ListAction{
		model: model,
	}
}

func (la *ListAction) listProject() error {
	condM := make(operator.M)

	if len(la.req.Name) != 0 {
		condM["name"] = la.req.Name
	}
	if len(la.req.EnglishName) != 0 {
		condM["englishname"] = la.req.EnglishName
	}
	if la.req.ProjectType > 0 {
		condM["projecttype"] = la.req.ProjectType
	}
	if la.req.UseBKRes {
		condM["usebkres"] = la.req.UseBKRes
	}
	if la.req.IsOffline {
		condM["isoffline"] = la.req.IsOffline
	}
	if la.req.UseBKRes {
		condM["usebkres"] = la.req.UseBKRes
	}
	if len(la.req.Kind) != 0 {
		condM["kind"] = la.req.Kind
	}
	if len(la.req.BusinessID) != 0 {
		condM["businessid"] = la.req.BusinessID
	}
	if len(la.req.DeployType) != 0 {
		condM["deploytype"] = la.req.DeployType
	}
	if len(la.req.BgID) != 0 {
		condM["bgid"] = la.req.BgID
	}
	if len(la.req.BgName) != 0 {
		condM["bgname"] = la.req.BgName
	}
	if len(la.req.DeptID) != 0 {
		condM["deptid"] = la.req.DeptID
	}
	if len(la.req.CenterID) != 0 {
		condM["centerid"] = la.req.CenterID
	}
	if len(la.req.CenterName) != 0 {
		condM["centername"] = la.req.CenterName
	}
	if la.req.IsSecret {
		condM["issecret"] = la.req.IsSecret
	}

	cond := operator.NewLeafCondition(operator.Eq, condM)
	projects, err := la.model.ListProject(la.ctx, cond, &storeopt.ListOption{})
	if err != nil {
		return err
	}
	for i := range projects {
		la.projectList = append(la.projectList, &projects[i])
	}
	return nil
}

func (la *ListAction) setResp(code uint32, msg string) {
	la.resp.Code = code
	la.resp.Message = msg
	la.resp.Result = (code == common.BcsErrClusterManagerSuccess)
	la.resp.Data = la.projectList
}

// Handle handle list cluster credential
func (la *ListAction) Handle(
	ctx context.Context, req *cmproto.ListProjectRequest, resp *cmproto.ListProjectResponse) {
	if req == nil || resp == nil {
		blog.Errorf("list project failed, req or resp is empty")
		return
	}
	la.ctx = ctx
	la.req = req
	la.resp = resp

	if err := req.Validate(); err != nil {
		la.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}
	if err := la.listProject(); err != nil {
		la.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}
	la.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
	return
}
