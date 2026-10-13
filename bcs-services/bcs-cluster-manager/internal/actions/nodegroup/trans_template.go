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

package nodegroup

import (
	"context"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/auth"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
)

// TransNgToNtAction trans nodeGroup to nodeTemplate
type TransNgToNtAction struct {
	ctx   context.Context
	model store.ClusterManagerModel
	req   *cmproto.TransNodeGroupToNodeTemplateRequest
	resp  *cmproto.TransNodeGroupToNodeTemplateResponse

	nt *cmproto.NodeTemplate
}

// NewTransNgToNtAction create update action for group
func NewTransNgToNtAction(model store.ClusterManagerModel) *TransNgToNtAction {
	return &TransNgToNtAction{
		model: model,
	}
}

func (ta *TransNgToNtAction) setResp(code uint32, msg string) {
	ta.resp.Code = code
	ta.resp.Message = msg
	ta.resp.Result = (code == common.BcsErrClusterManagerSuccess)
	ta.resp.Template = ta.nt
}

// validate check
func (ta *TransNgToNtAction) validate() error {
	if err := ta.req.Validate(); err != nil {
		return err
	}

	return nil
}

func (ta *TransNgToNtAction) transNgToNt() error {
	ng, err := ta.model.GetNodeGroup(ta.ctx, ta.req.GetNodeGroupID())
	if err != nil {
		return err
	}

	timeStr := time.Now().UTC().Format(time.RFC3339)
	nt := ng.NodeTemplate
	nt.Name = ng.Name
	nt.ProjectID = ng.ProjectID
	nt.NodeTemplateID = utils.GenerateTemplateID(utils.NodeTemplate)
	nt.Creator = ng.Creator
	nt.Updater = ng.Updater
	nt.CreateTime = timeStr
	nt.UpdateTime = timeStr
	nt.Desc = fmt.Sprintf("copy from %s", ng.NodeGroupID)

	ta.nt = nt
	err = ta.model.CreateNodeTemplate(ta.ctx, nt)
	if err != nil {
		return err
	}

	err = ta.model.CreateOperationLog(ta.ctx, &cmproto.OperationLog{
		ResourceType: common.NodeGroup.String(),
		ResourceID:   ta.req.GetNodeGroupID(),
		TaskID:       "",
		Message:      fmt.Sprintf("节点池[%s]转换为节点模板", ta.req.GetNodeGroupID()),
		OpUser:       auth.GetUserFromCtx(ta.ctx),
		CreateTime:   time.Now().UTC().Format(time.RFC3339),
		ResourceName: ng.GetName(),
	})
	if err != nil {
		blog.Errorf("TransNodeGroupToNodeTemplate[%s] CreateOperationLog failed: %v", ta.req.GetNodeGroupID(), err)
	}
	return nil
}

// Handle handle update cluster credential
func (ta *TransNgToNtAction) Handle(ctx context.Context,
	req *cmproto.TransNodeGroupToNodeTemplateRequest, resp *cmproto.TransNodeGroupToNodeTemplateResponse) {

	if req == nil || resp == nil {
		blog.Errorf("trans nodeGroup to nodeTemplate failed, req or resp is empty")
		return
	}
	ta.ctx = ctx
	ta.req = req
	ta.resp = resp

	err := ta.validate()
	if err != nil {
		ta.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	err = ta.transNgToNt()
	if err != nil {
		blog.Errorf("nodegroup %s trans nodetemplate failed: %s",
			ta.req.NodeGroupID, err.Error())
		ta.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	blog.Infof("nodegroup %s trans to nodetemplate successfully", ta.req.NodeGroupID)
	ta.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}
