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

package autoscalingoption

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
	ctx        context.Context
	model      store.ClusterManagerModel
	req        *cmproto.ListAutoScalingOptionRequest
	resp       *cmproto.ListAutoScalingOptionResponse
	optionList []*cmproto.ClusterAutoScalingOption
}

// NewListAction create list action for cluster credential
func NewListAction(model store.ClusterManagerModel) *ListAction {
	return &ListAction{
		model: model,
	}
}

func (la *ListAction) listAutoScalingOption() error {
	condM := make(operator.M)

	if len(la.req.ClusterID) != 0 {
		condM["clusterid"] = la.req.ClusterID
	}
	if len(la.req.ProjectID) != 0 {
		condM["projectid"] = la.req.ProjectID
	}
	if len(la.req.Creator) != 0 {
		condM["creator"] = la.req.Creator
	}
	if len(la.req.Updater) != 0 {
		condM["updater"] = la.req.Updater
	}

	cond := operator.NewLeafCondition(operator.Eq, condM)
	options, err := la.model.ListAutoScalingOption(la.ctx, cond, &storeopt.ListOption{})
	if err != nil {
		return err
	}
	for i := range options {
		la.optionList = append(la.optionList, &options[i])
	}
	return nil
}

func (la *ListAction) setResp(code uint32, msg string) {
	la.resp.Code = code
	la.resp.Message = msg
	la.resp.Result = (code == common.BcsErrClusterManagerSuccess)
	la.resp.Data = la.optionList
}

// Handle handle list cluster credential
func (la *ListAction) Handle(
	ctx context.Context, req *cmproto.ListAutoScalingOptionRequest, resp *cmproto.ListAutoScalingOptionResponse) {
	if req == nil || resp == nil {
		blog.Errorf("list ClusterAutoScalingOption failed, req or resp is empty")
		return
	}
	la.ctx = ctx
	la.req = req
	la.resp = resp

	if err := req.Validate(); err != nil {
		la.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}
	if err := la.listAutoScalingOption(); err != nil {
		la.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}
	la.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
	return
}
