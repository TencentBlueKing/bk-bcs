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

// Package notifytemplate xxx
package notifytemplate

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	storeopt "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

// ListAction action for list notifyTemplate
type ListAction struct {
	ctx   context.Context
	model store.ClusterManagerModel

	req                *cmproto.ListNotifyTemplateRequest
	resp               *cmproto.ListNotifyTemplateResponse
	notifyTemplateList []*cmproto.NotifyTemplate
}

// NewListAction create list action for notify template
func NewListAction(model store.ClusterManagerModel) *ListAction {
	return &ListAction{
		model: model,
	}
}

func (la *ListAction) listNotifyTemplates() error {
	condM := make(operator.M)
	//! we don't setting bson tag in proto file
	//! all fields are in lowcase
	if len(la.req.ProjectID) != 0 {
		condM["projectid"] = la.req.ProjectID
	}
	if len(la.req.NotifyTemplateID) != 0 {
		condM["notifytemplateid"] = la.req.NotifyTemplateID
	}

	cond := operator.NewLeafCondition(operator.Eq, condM)
	templates, err := la.model.ListNotifyTemplate(la.ctx, cond, &storeopt.ListOption{
		Sort: map[string]int{
			"updatetime": -1,
		},
	})
	if err != nil {
		return err
	}

	for i := range templates {
		templates[i].UpdateTime = utils.TransTimeFormat(templates[i].UpdateTime)
		templates[i].CreateTime = utils.TransTimeFormat(templates[i].CreateTime)

		la.notifyTemplateList = append(la.notifyTemplateList, templates[i])
	}
	return nil
}

func (la *ListAction) setResp(code uint32, msg string) {
	la.resp.Code = code
	la.resp.Message = msg
	la.resp.Result = (code == common.BcsErrClusterManagerSuccess)
	la.resp.Data = la.notifyTemplateList
}

// Handle handle list cluster notifyTemplate list
func (la *ListAction) Handle(
	ctx context.Context, req *cmproto.ListNotifyTemplateRequest, resp *cmproto.ListNotifyTemplateResponse) {
	if req == nil || resp == nil {
		blog.Errorf("list notifyTemplate failed, req or resp is empty")
		return
	}
	la.ctx = ctx
	la.req = req
	la.resp = resp

	if err := req.Validate(); err != nil {
		la.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}
	if err := la.listNotifyTemplates(); err != nil {
		la.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}
	la.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}
