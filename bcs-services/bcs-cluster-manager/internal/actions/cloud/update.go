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

package cloud

import (
	"context"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

// UpdateAction update action for online cluster credential
type UpdateAction struct {
	ctx   context.Context
	model store.ClusterManagerModel
	req   *cmproto.UpdateCloudRequest
	resp  *cmproto.UpdateCloudResponse
}

// NewUpdateAction create update action for online cluster credential
func NewUpdateAction(model store.ClusterManagerModel) *UpdateAction {
	return &UpdateAction{
		model: model,
	}
}

func (ua *UpdateAction) updateCloud(destCloud *cmproto.Cloud) error {
	timeStr := time.Now().Format(time.RFC3339)
	destCloud.UpdateTime = timeStr
	destCloud.Updater = ua.req.Updater

	if len(ua.req.Name) != 0 {
		destCloud.Name = ua.req.Name
	}
	if ua.req.OpsPlugins != nil {
		destCloud.OpsPlugins = ua.req.OpsPlugins
	}
	if ua.req.ExtraPlugins != nil {
		destCloud.ExtraPlugins = ua.req.ExtraPlugins
	}
	if ua.req.CloudCredential != nil {
		destCloud.CloudCredential = ua.req.CloudCredential
	}
	if ua.req.OsManagement != nil {
		destCloud.OsManagement = ua.req.OsManagement
	}
	if ua.req.ClusterManagement != nil {
		destCloud.ClusterManagement = ua.req.ClusterManagement
	}
	if ua.req.NodeGroupManagement != nil {
		destCloud.NodeGroupManagement = ua.req.NodeGroupManagement
	}
	if len(ua.req.CloudProvider) > 0 {
		destCloud.CloudProvider = ua.req.CloudProvider
	}
	if len(ua.req.Config) > 0 {
		destCloud.Config = ua.req.Config
	}
	if len(ua.req.Description) > 0 {
		destCloud.Description = ua.req.Description
	}
	if len(ua.req.EngineType) > 0 {
		destCloud.EngineType = ua.req.EngineType
	}
	if len(ua.req.Enable) > 0 {
		destCloud.Enable = ua.req.Enable
	}
	if ua.req.NetworkInfo != nil {
		destCloud.NetworkInfo = ua.req.NetworkInfo
	}
	if ua.req.ConfInfo != nil {
		destCloud.ConfInfo = ua.req.ConfInfo
	}
	if ua.req.PlatformInfo != nil {
		destCloud.PlatformInfo = ua.req.PlatformInfo
	}

	return ua.model.UpdateCloud(ua.ctx, destCloud)
}

func (ua *UpdateAction) setResp(code uint32, msg string) {
	ua.resp.Code = code
	ua.resp.Message = msg
	ua.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

func (ua *UpdateAction) validate() error {
	if err := ua.req.Validate(); err != nil {
		return err
	}

	if len(ua.req.Enable) > 0 {
		if !utils.StringInSlice(ua.req.Enable, cloudEnable) {
			return fmt.Errorf("cloud enable parameter invalid, must be true or false")
		}
	}

	return nil
}

// Handle handle update cluster credential
func (ua *UpdateAction) Handle(
	ctx context.Context, req *cmproto.UpdateCloudRequest, resp *cmproto.UpdateCloudResponse) {

	if req == nil || resp == nil {
		blog.Errorf("update cloud failed, req or resp is empty")
		return
	}
	ua.ctx = ctx
	ua.req = req
	ua.resp = resp

	if err := ua.validate(); err != nil {
		ua.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}
	destCloud, err := ua.model.GetCloud(ua.ctx, req.CloudID)
	if err != nil {
		ua.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		blog.Errorf("find cloud %s failed when pre-update checking, err %s", req.CloudID, err.Error())
		return
	}
	if err = ua.updateCloud(destCloud); err != nil {
		ua.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	err = ua.model.CreateOperationLog(ua.ctx, &cmproto.OperationLog{
		ResourceType: common.Cloud.String(),
		ResourceID:   req.CloudID,
		TaskID:       "",
		Message:      fmt.Sprintf("更新云[%s]模板信息", req.CloudID),
		OpUser:       req.Updater,
		CreateTime:   time.Now().String(),
	})
	if err != nil {
		blog.Errorf("UpdateCloud[%s] CreateOperationLog failed: %v", req.CloudID, err)
	}

	ua.resp.Data = destCloud
	ua.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
	return
}
