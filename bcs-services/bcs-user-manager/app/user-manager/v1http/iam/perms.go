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
 *
 */

package iam

import (
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/cloudaccount"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/cluster"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/namespace"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/project"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/templateset"
	authutil "github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/utils"
	iamsdk "github.com/TencentBlueKing/iam-go-sdk"
	"github.com/emicklei/go-restful"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/pkg/constant"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/models"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/config"
)

// PermRequest perm request
type PermRequest struct {
	ActionIDs []string `json:"action_ids"`
	PermCtx   *PermCtx `json:"perm_ctx"`
}

// PermCtx perm context
type PermCtx struct {
	ResourceType string `json:"resource_type"`
	ProjectID    string `json:"project_id"`
	ClusterID    string `json:"cluster_id"`
	Namespace    string `json:"name"`
	TemplateID   string `json:"template_id"`
	AccountID    string `json:"account_id"`
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

	// get perm
	permReq := iam.PermissionRequest{
		SystemID: config.GetGlobalConfig().IAMConfig.SystemID,
		UserName: user.Name,
	}
	var result map[string]bool
	if form.PermCtx == nil {
		result, err = config.GloablIAMClient.MultiActionsAllowedWithoutResource(form.ActionIDs, permReq)
	} else {
		nodes := make([]iam.ResourceNode, 0)
		nodes = append(nodes, getResourceNodeFromPermCtx(form.PermCtx))
		result, err = config.GloablIAMClient.ResourceMultiActionsAllowed(form.ActionIDs, permReq, nodes)
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
		form.PermCtx.ResourceType = getResourceTypeFromAction(actionID)
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

	// get perm
	permReq := iam.PermissionRequest{
		SystemID: config.GetGlobalConfig().IAMConfig.SystemID,
		UserName: user.Name,
	}
	var allow bool
	var applyURL string
	if form.PermCtx == nil {
		allow, err = config.GloablIAMClient.IsAllowedWithoutResource(actionID, permReq, true)
	} else {
		node := getResourceNodeFromPermCtx(form.PermCtx)
		allow, err = config.GloablIAMClient.IsAllowedWithResource(actionID, permReq, []iam.ResourceNode{node}, true)
	}
	if err != nil {
		msg := fmt.Sprintf("get perm failed, err %s", err.Error())
		utils.WriteServerError(response, common.BcsErrApiBadRequest, msg)
		return
	}

	if !allow {
		applyURL, err = getApplyURL(getApplicationsFromPermCtx(form.PermCtx, actionID))
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

// 根据 resource type 拼装 iam.ResourceNode
func getResourceNodeFromPermCtx(permCtx *PermCtx) iam.ResourceNode {
	node := iam.ResourceNode{System: config.GetGlobalConfig().IAMConfig.SystemID, RType: permCtx.ResourceType}
	switch permCtx.ResourceType {
	case Project:
		node.RInstance = permCtx.ProjectID
		node.Rp = project.ProjectResourcePath{}
	case Cluster:
		node.RInstance = permCtx.ClusterID
		node.Rp = cloudaccount.AccountResourcePath{ProjectID: permCtx.ProjectID}
	case Namespace:
		node.RInstance = authutil.CalcIAMNsID(permCtx.ClusterID, permCtx.Namespace)
		node.Rp = namespace.NamespaceResourcePath{ProjectID: permCtx.ProjectID, ClusterID: permCtx.ClusterID}
	case TemplateSet:
		node.RInstance = permCtx.TemplateID
		node.Rp = cloudaccount.AccountResourcePath{ProjectID: permCtx.ProjectID}
	case CloudAccount:
		node.RInstance = permCtx.AccountID
		node.Rp = cloudaccount.AccountResourcePath{ProjectID: permCtx.ProjectID}
	}
	return node
}

// 根据 resource type 拼装 iam.ApplicationAction
func getApplicationsFromPermCtx(permCtx *PermCtx, actionsID string) []iam.ApplicationAction {
	apps := make([]iam.ApplicationAction, 0)
	switch permCtx.ResourceType {
	case Project:
		apps = project.BuildProjectSameInstanceApplication(false, []string{actionsID}, []string{permCtx.ProjectID})
	case Cluster:
		apps = cluster.BuildClusterSameInstanceApplication(false, []string{actionsID}, []cluster.ProjectClusterData{
			{
				Project: permCtx.ProjectID,
				Cluster: permCtx.ClusterID,
			},
		})
	case Namespace:
		apps = append(apps, namespace.BuildNamespaceApplicationInstance(namespace.NamespaceApplicationAction{
			ActionID: actionsID,
			Data: []namespace.ProjectNamespaceData{{
				Project:   permCtx.ProjectID,
				Cluster:   permCtx.ClusterID,
				Namespace: authutil.CalcIAMNsID(permCtx.ClusterID, permCtx.Namespace),
			}},
		}))
	case TemplateSet:
		instances := make([][]iam.Instance, 0)
		instances = append(instances, []iam.Instance{
			{
				ResourceType: Project,
				ResourceID:   permCtx.ProjectID,
			},
			{
				ResourceType: TemplateSet,
				ResourceID:   permCtx.TemplateID,
			},
		})
		rr := make([]iamsdk.ApplicationRelatedResourceType, 0)
		rr = append(rr, authutil.BuildRelatedSystemResource(iam.SystemIDBKBCS, permCtx.ResourceType, instances))
		apps = []iam.ApplicationAction{
			{
				ActionID:         actionsID,
				RelatedResources: rr,
			},
		}
	case CloudAccount:
		apps = cloudaccount.BuildAccountSameInstanceApplication(false, []string{actionsID},
			[]cloudaccount.ProjectAccountData{
				{
					Project: permCtx.ProjectID,
					Account: permCtx.AccountID,
				},
			})
	}
	return apps
}

// getApplyURL get apply url
func getApplyURL(applications []iam.ApplicationAction) (string, error) {
	url, err := config.GloablIAMClient.GetApplyURL(iam.ApplicationRequest{
		SystemID: config.GetGlobalConfig().IAMConfig.SystemID}, applications, iam.BkUser{
		BkUserName: iam.SystemUser,
	})
	if err != nil {
		return iam.IamAppURL, err
	}
	return url, nil
}

// NOCC:CCN_thresholde(设计如此),golint/fnsize(设计如此)
func getResourceTypeFromAction(action string) string {
	switch action {
	case project.ProjectCreate.String():
		return ""
	case project.ProjectView.String():
		return string(project.SysProject)
	case project.ProjectEdit.String():
		return string(project.SysProject)
	case project.ProjectDelete.String():
		return string(project.SysProject)
	case cluster.ClusterCreate.String():
		return string(project.SysProject)
	case cluster.ClusterView.String():
		return string(cluster.SysCluster)
	case cluster.ClusterManage.String():
		return string(cluster.SysCluster)
	case cluster.ClusterDelete.String():
		return string(cluster.SysCluster)
	case cluster.ClusterUse.String():
		return string(cluster.SysCluster)
	case namespace.NameSpaceCreate.String():
		return string(cluster.SysCluster)
	case namespace.NameSpaceView.String():
		return string(namespace.SysNamespace)
	case namespace.NameSpaceUpdate.String():
		return string(namespace.SysNamespace)
	case namespace.NameSpaceDelete.String():
		return string(namespace.SysNamespace)
	case namespace.NameSpaceList.String():
		return string(cluster.SysCluster)
	case cluster.ClusterScopedCreate.String():
		return string(cluster.SysCluster)
	case cluster.ClusterScopedView.String():
		return string(cluster.SysCluster)
	case cluster.ClusterScopedUpdate.String():
		return string(cluster.SysCluster)
	case cluster.ClusterScopedDelete.String():
		return string(cluster.SysCluster)
	case namespace.NameSpaceScopedCreate.String():
		return string(namespace.SysNamespace)
	case namespace.NameSpaceScopedView.String():
		return string(namespace.SysNamespace)
	case namespace.NameSpaceScopedUpdate.String():
		return string(namespace.SysNamespace)
	case namespace.NameSpaceScopedDelete.String():
		return string(namespace.SysNamespace)
	case templateset.TemplateSetCreate.String():
		return string(project.SysProject)
	case templateset.TemplateSetView.String():
		return string(templateset.SysTemplateSet)
	case templateset.TemplateSetCopy.String():
		return string(templateset.SysTemplateSet)
	case templateset.TemplateSetUpdate.String():
		return string(templateset.SysTemplateSet)
	case templateset.TemplateSetDelete.String():
		return string(templateset.SysTemplateSet)
	case templateset.TemplateSetInstantiate.String():
		return string(templateset.SysTemplateSet)
	case cloudaccount.AccountCreate.String():
		return string(project.SysProject)
	case cloudaccount.AccountManage.String():
		return string(cloudaccount.SysCloudAccount)
	case cloudaccount.AccountUse.String():
		return string(cloudaccount.SysCloudAccount)
	default:
		return ""
	}
}
