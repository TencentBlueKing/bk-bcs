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

package moduleflag

import (
	"context"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/auth"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
)

// DeleteAction action for delete cloud moduleFlag
type DeleteAction struct {
	ctx context.Context

	model store.ClusterManagerModel
	req   *cmproto.DeleteCloudModuleFlagRequest
	resp  *cmproto.DeleteCloudModuleFlagResponse
}

// NewDeleteAction create delete action for cloudModuleFlag
func NewDeleteAction(model store.ClusterManagerModel) *DeleteAction {
	return &DeleteAction{
		model: model,
	}
}

func (da *DeleteAction) setResp(code uint32, msg string) {
	da.resp.Code = code
	da.resp.Message = msg
}

func (da *DeleteAction) validate() error {
	err := da.req.Validate()
	if err != nil {
		return err
	}

	return nil
}

func (da *DeleteAction) deleteCloudModuleFlag() error {
	if len(da.req.FlagNameList) == 0 {
		err := da.model.DeleteCloudModuleFlag(da.ctx, da.req.CloudID, da.req.Version, da.req.ModuleID, "")
		if err != nil {
			blog.Errorf("DeleteCloudModuleFlag %s:%s failed: %v", da.req.CloudID, da.req.ModuleID, err)
			return err
		}

		return nil
	}

	for _, name := range da.req.FlagNameList {
		// try to get original cloud module flag for return
		_, err := da.model.GetCloudModuleFlag(da.ctx, da.req.CloudID, da.req.Version, da.req.ModuleID, name)
		if err != nil {
			blog.Errorf("Get CloudModuleFlag %s:%s in pre-delete checking failed, err %s",
				da.req.CloudID, da.req.ModuleID, err.Error())
			return err
		}

		if err = da.model.DeleteCloudModuleFlag(da.ctx, da.req.CloudID, da.req.Version,
			da.req.ModuleID, name); err != nil {
			blog.Errorf("delete CloudModuleFlag %s:%s failed, err %s", da.req.CloudID, da.req.ModuleID,
				err.Error())
			return err
		}
	}

	return nil
}

// Handle handle delete cloud module flag
func (da *DeleteAction) Handle(
	ctx context.Context, req *cmproto.DeleteCloudModuleFlagRequest, resp *cmproto.DeleteCloudModuleFlagResponse) {
	if req == nil || resp == nil {
		blog.Errorf("delete cloudModuleFlag failed, req or resp is empty")
		return
	}
	da.ctx = ctx
	da.req = req
	da.resp = resp

	if err := da.validate(); err != nil {
		da.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	if err := da.deleteCloudModuleFlag(); err != nil {
		da.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	err := da.model.CreateOperationLog(da.ctx, &cmproto.OperationLog{
		ResourceType: common.Cloud.String(),
		ResourceID:   da.req.CloudID,
		TaskID:       "",
		Message:      fmt.Sprintf("删除云组件参数[%s]", da.req.CloudID),
		OpUser:       auth.GetUserFromCtx(da.ctx),
		CreateTime:   time.Now().Format(time.RFC3339),
	})
	if err != nil {
		blog.Errorf("DeleteCloudModuleFlag[%s] CreateOperationLog failed: %v", da.req.CloudID, err)
	}

	da.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}
