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

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

// GetAction action for getting cluster credential
type GetAction struct {
	ctx context.Context

	model store.ClusterManagerModel
	req   *cmproto.GetNodeGroupRequest
	resp  *cmproto.GetNodeGroupResponse
}

// NewGetAction create get action for online cluster credential
func NewGetAction(model store.ClusterManagerModel) *GetAction {
	return &GetAction{
		model: model,
	}
}

func (ga *GetAction) setResp(code uint32, msg string) {
	ga.resp.Code = code
	ga.resp.Message = msg
	ga.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

// Handle handle get cluster credential
func (ga *GetAction) Handle(
	ctx context.Context, req *cmproto.GetNodeGroupRequest, resp *cmproto.GetNodeGroupResponse) {
	if req == nil || resp == nil {
		blog.Errorf("get NodeGroup failed, req or resp is empty")
		return
	}
	ga.ctx = ctx
	ga.req = req
	ga.resp = resp

	if err := req.Validate(); err != nil {
		ga.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	group, err := ga.model.GetNodeGroup(ctx, req.NodeGroupID)
	if err != nil {
		ga.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	group = ga.pruneNodeGroup(group)
	resp.Data = group
	ga.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}

func removeSensitiveInfo(group *cmproto.NodeGroup) *cmproto.NodeGroup {
	if group != nil && group.LaunchTemplate != nil {
		if group.LaunchTemplate.InitLoginPassword != "" {
			group.LaunchTemplate.InitLoginPassword = "<masked>"
		}
		if group.LaunchTemplate.KeyPair != nil {
			if group.LaunchTemplate.KeyPair.KeyPublic != "" {
				group.LaunchTemplate.KeyPair.KeyPublic = "<masked>"
			}
			group.LaunchTemplate.KeyPair.KeySecret = ""
		}
	}
	return group
}

func (ga *GetAction) pruneNodeGroup(group *cmproto.NodeGroup) *cmproto.NodeGroup {
	// remove sensitive password in response
	group = removeSensitiveInfo(group)

	// decode userscript
	if group.NodeTemplate != nil {
		scaleOutPreScript, _ := utils.Base64Decode(group.NodeTemplate.PreStartUserScript)
		group.NodeTemplate.PreStartUserScript = scaleOutPreScript
		userScript, _ := utils.Base64Decode(group.NodeTemplate.UserScript)
		group.NodeTemplate.UserScript = userScript
		scaleInPre, _ := utils.Base64Decode(group.NodeTemplate.ScaleInPreScript)
		group.NodeTemplate.ScaleInPreScript = scaleInPre
		scaleInPost, _ := utils.Base64Decode(group.NodeTemplate.ScaleInPostScript)
		group.NodeTemplate.ScaleInPostScript = scaleInPost
	}

	return group
}

// GetExternalNodeScriptAction for getting external node script action
type GetExternalNodeScriptAction struct {
	ctx context.Context

	model     store.ClusterManagerModel
	req       *cmproto.GetExternalNodeScriptRequest
	resp      *cmproto.GetExternalNodeScriptResponse
	nodeGroup *cmproto.NodeGroup
	cloud     *cmproto.Cloud
}

// NewGetExternalNodesScriptAction create get action for group script
func NewGetExternalNodesScriptAction(model store.ClusterManagerModel) *GetExternalNodeScriptAction {
	return &GetExternalNodeScriptAction{
		model: model,
	}
}

func (ga *GetExternalNodeScriptAction) setResp(code uint32, msg string) {
	ga.resp.Code = code
	ga.resp.Message = msg
	ga.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

func (ga *GetExternalNodeScriptAction) validate() error {
	if err := ga.req.Validate(); err != nil {
		return err
	}

	if ga.nodeGroup.Provider != utils.TencentCloud {
		return fmt.Errorf("GetExternalNodeScriptAction not supported cloudType[%s]", ga.nodeGroup.Provider)
	}
	if ga.nodeGroup.NodeGroupType != common.External.String() {
		return fmt.Errorf("GetExternalNodeScriptAction nodeGroupType[%s]", ga.nodeGroup.NodeGroupType)
	}
	if ga.nodeGroup.CloudNodeGroupID == "" {
		return fmt.Errorf("GetExternalNodeScriptAction cloudNodeGroup empty")
	}

	return nil
}

func (ga *GetExternalNodeScriptAction) getExternalNodeScript() error {
	mgr, err := cloudprovider.GetNodeGroupMgr(ga.cloud.CloudProvider)
	if err != nil {
		blog.Errorf("get NodeGroup Manager cloudprovider %s/%s failed, %s",
			ga.cloud.CloudID, ga.cloud.CloudProvider, err.Error())
		return err
	}

	script, err := mgr.GetExternalNodeScript(ga.nodeGroup, ga.req.GetInternal())
	if err != nil {
		blog.Errorf("GetExternalNodeScriptAction GetExternalNodeScript failed: %v", err)
		return err
	}
	ga.resp.Data = script

	return nil
}

func (ga *GetExternalNodeScriptAction) getRelativeData() error {
	group, err := ga.model.GetNodeGroup(ga.ctx, ga.req.NodeGroupID)
	if err != nil {
		return err
	}
	ga.nodeGroup = group

	cloud, err := ga.model.GetCloud(ga.ctx, group.Provider)
	if err != nil {
		return err
	}
	ga.cloud = cloud

	return nil
}

// Handle handle get cluster credential
func (ga *GetExternalNodeScriptAction) Handle(
	ctx context.Context, req *cmproto.GetExternalNodeScriptRequest, resp *cmproto.GetExternalNodeScriptResponse) {
	if req == nil || resp == nil {
		blog.Errorf("get externalNodeScript failed, req or resp is empty")
		return
	}
	ga.ctx = ctx
	ga.req = req
	ga.resp = resp

	err := ga.getRelativeData()
	if err != nil {
		ga.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	if err = ga.validate(); err != nil {
		ga.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	if err = ga.getExternalNodeScript(); err != nil {
		ga.setResp(common.BcsErrClusterManagerExternalNodeScriptErr, err.Error())
		return
	}

	ga.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}
