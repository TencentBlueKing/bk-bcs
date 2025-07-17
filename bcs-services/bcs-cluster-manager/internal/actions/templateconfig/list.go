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

// Package templateconfig xxx
package templateconfig

import (
	"context"
	"errors"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	storeopt "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/options"
)

// ListAction action for list templateConfig
type ListAction struct {
	ctx   context.Context
	model store.ClusterManagerModel
	req   *cmproto.ListTemplateConfigRequest
	resp  *cmproto.ListTemplateConfigResponse

	configInfos []*cmproto.TemplateConfigInfo
}

// NewListAction create list action for notify template
func NewListAction(model store.ClusterManagerModel) *ListAction {
	return &ListAction{
		model: model,
	}
}

func (la *ListAction) listTemplateConfigs() error {
	opt := &storeopt.ListOption{Sort: map[string]int{
		"updatetime": -1,
	}}
	configs, err := getTemplateConfigInfos(la.ctx, la.model, la.req.BusinessID, la.req.ProjectID,
		la.req.ClusterID, la.req.Provider, la.req.ConfigType, opt)
	if err != nil && !errors.Is(err, drivers.ErrTableRecordNotFound) {
		blog.Errorf("ListTemplateConfig businessID[%s] projectID[%s] clusterID[%s] provider[%s] "+
			"configType[%s] failed, %s", la.req.BusinessID, la.req.ProjectID, la.req.ClusterID, la.req.Provider,
			la.req.ConfigType, err.Error())
		return err
	}

	la.configInfos = configs

	return nil
}

// setResp set response
func (la *ListAction) setResp(code uint32, msg string) {
	la.resp.Code = code
	la.resp.Message = msg
	la.resp.Result = (code == common.BcsErrClusterManagerSuccess)
	la.resp.Data = la.configInfos
}

// Handle handle list cluster templateConfig list
func (la *ListAction) Handle(
	ctx context.Context, req *cmproto.ListTemplateConfigRequest, resp *cmproto.ListTemplateConfigResponse) {
	if req == nil || resp == nil {
		blog.Errorf("list templateConfig failed, req or resp is empty")
		return
	}
	la.ctx = ctx
	la.req = req
	la.resp = resp

	if err := req.Validate(); err != nil {
		la.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}
	if err := la.listTemplateConfigs(); err != nil {
		la.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}
	la.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}
