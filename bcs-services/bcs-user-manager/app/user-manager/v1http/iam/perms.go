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

package iam

import (
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
	"github.com/emicklei/go-restful"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/pkg/constant"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/models"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/v1http/auth"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/config"
)

// PermRequest perm request
type PermRequest struct {
	ActionIDs []string      `json:"action_ids"`
	PermCtx   *auth.PermCtx `json:"perm_ctx"`
}

// GetPerms get perm
func GetPerms(request *restful.Request, response *restful.Response) {
	form := PermRequest{}
	_ = request.ReadEntity(&form)
	err := utils.Validate.Struct(&form)
	if err != nil {
		_ = response.WriteHeaderAndEntity(400, utils.FormatValidationError(err))
		return
	}

	// get current user
	currentUser := request.Attribute(constant.CurrentUserAttr)
	var user *models.BcsUser
	if v, ok := currentUser.(*models.BcsUser); ok {
		user = v
	}
	if user == nil {
		utils.WriteUnauthorizedError(response, common.BcsErrApiUnauthorized, "user is not valid")
		return
	}

	// permission switch for special case
	if config.GetGlobalConfig().PermissionSwitch {
		result := map[string]bool{}
		for _, actionID := range form.ActionIDs {
			result[actionID] = true
		}
		data := utils.CreateResponseData(nil, "success", map[string]interface{}{"perms": result})
		_, _ = response.Write([]byte(data))
		return
	}

	// get perm
	permReq := iam.PermissionRequest{
		SystemID: config.GetGlobalConfig().IAMConfig.SystemID,
		UserName: user.Name,
	}
	var result map[string]bool
	if form.PermCtx == nil || form.PermCtx.ResourceType == "" {
		result, err = config.GloablIAMClient(utils.GetTenantIDFromContext(request.Request.Context())).
			MultiActionsAllowedWithoutResource(form.ActionIDs, permReq)
	} else {
		nodes := make([]iam.ResourceNode, 0)
		nodes = append(nodes, auth.GetResourceNodeFromPermCtx(form.PermCtx))
		result, err = config.GloablIAMClient(utils.GetTenantIDFromContext(request.Request.Context())).
			ResourceMultiActionsAllowed(form.ActionIDs, permReq, nodes)
	}
	if err != nil {
		msg := fmt.Sprintf("get perm failed, err %s", err.Error())
		utils.WriteServerError(response, common.BcsErrApiBadRequest, msg)
		return
	}

	data := utils.CreateResponseData(nil, "success", map[string]interface{}{"perms": result})
	_, _ = response.Write([]byte(data))
}

// GetPermByActionID get perm by action id
func GetPermByActionID(request *restful.Request, response *restful.Response) {
	actionID := request.PathParameter("action_id")
	form := PermRequest{}
	_ = request.ReadEntity(&form)
	err := utils.Validate.Struct(&form)
	if err != nil {
		_ = response.WriteHeaderAndEntity(400, utils.FormatValidationError(err))
		return
	}
	if form.PermCtx != nil && form.PermCtx.ResourceType == "" {
		form.PermCtx.ResourceType = auth.GetResourceTypeFromAction(actionID)
	}

	// get current user
	currentUser := request.Attribute(constant.CurrentUserAttr)
	var user *models.BcsUser
	if v, ok := currentUser.(*models.BcsUser); ok {
		user = v
	}
	if user == nil {
		utils.WriteUnauthorizedError(response, common.BcsErrApiUnauthorized, "user is not valid")
		return
	}

	// permission switch for special case
	if config.GetGlobalConfig().PermissionSwitch {
		result := map[string]bool{
			actionID: true,
		}
		data := utils.CreateResponseData(nil, "success", map[string]interface{}{"perms": result})
		_, _ = response.Write([]byte(data))
		return
	}

	// get perm
	permReq := iam.PermissionRequest{
		SystemID: config.GetGlobalConfig().IAMConfig.SystemID,
		UserName: user.Name,
	}
	var allow bool
	var applyURL string
	if form.PermCtx == nil || form.PermCtx.ResourceType == "" {
		allow, err = config.GloablIAMClient(utils.GetTenantIDFromContext(request.Request.Context())).
			IsAllowedWithoutResource(actionID, permReq, true)
	} else {
		node := auth.GetResourceNodeFromPermCtx(form.PermCtx)
		allow, err = config.GloablIAMClient(utils.GetTenantIDFromContext(request.Request.Context())).
			IsAllowedWithResource(actionID, permReq, []iam.ResourceNode{node}, true)
	}
	if err != nil {
		msg := fmt.Sprintf("get perm failed, err %s", err.Error())
		utils.WriteServerError(response, common.BcsErrApiBadRequest, msg)
		return
	}

	if !allow {
		applyURL, err = auth.GetApplyURL(auth.GetApplicationsFromPermCtx(form.PermCtx, actionID),
			utils.GetTenantIDFromContext(request.Request.Context()))
	}
	if err != nil {
		msg := fmt.Sprintf("get apply url failed, err %s", err.Error())
		utils.WriteServerError(response, common.BcsErrApiBadRequest, msg)
		return
	}

	data := utils.CreateResponseData(nil, "success", map[string]interface{}{
		"perms": map[string]interface{}{actionID: allow, "apply_url": applyURL}})
	_, _ = response.Write([]byte(data))
}
