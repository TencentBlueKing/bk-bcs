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

package v1http

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/metrics"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/models"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/storages/sqlstore"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/utils"
	"github.com/emicklei/go-restful"
)

// PermissionForm registe form
type PermissionForm struct {
	UserName     string `json:"user_name" validate:"required"`
	ResourceType string `json:"resource_type" validate:"required"`
	Resource     string `json:"resource"`
	Role         string `json:"role" validate:"required"`
}

// PermissionsResp query response for
type PermissionsResp struct {
	ResourceType string `json:"resource_type"`
	Resource     string `json:"resource"`
	Role         string `json:"role"`
}

// GetPermissionForm request form
type GetPermissionForm struct {
	UserName     string `json:"user_name" validate:"required"`
	ResourceType string `json:"resource_type" validate:"required"`
}

// VerifyPermissionForm request form for permission
type VerifyPermissionForm struct {
	UserToken    string `json:"user_token" validate:"required"`
	ResourceType string `json:"resource_type" validate:"required"`
	Resource     string `json:"resource"`
	Action       string `json:"action" validate:"required"`
}

// VerifyPermissionResponse http verify response
type VerifyPermissionResponse struct {
	Allowed bool   `json:"allowed"`
	Message string `json:"message"`
}

// OwnedPermissions action
type OwnedPermissions struct {
	Actions string `json:"actions"`
}

//UserResourceAction resource operation action
type UserResourceAction struct {
	UserId       uint
	ResourceType string
	Resource     string
	Actions      string
}

// UserPermissions user permission definition
type UserPermissions struct {
	ResourceType string
	Resource     string
	Actions      string
}

// PermissionsCache local cache for speed up
var PermissionsCache map[uint][]UserPermissions
var mutex *sync.RWMutex

// initCache sync data from db to cache periodically
func initCache() {
	initRoles := []models.BcsRole{
		{
			Name:    "manager",
			Actions: "GET,POST,PUT,PATCH,DELETE",
		},
		{
			Name:    "viewer",
			Actions: "GET",
		},
	}
	for _, role := range initRoles {
		m := sqlstore.GetRole(role.Name)
		if m == nil {
			err := sqlstore.CreateRole(&role)
			if err != nil {
				blog.Errorf("Failed to init role [%s]: %s", role.Name, err.Error())
			}
		}
	}

	mutex = new(sync.RWMutex)
	var ura []UserResourceAction
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()
	for {
		sqlstore.GCoreDB.Table("bcs_user_resource_roles").Select("bcs_user_resource_roles.user_id, bcs_user_resource_roles.resource_type, bcs_user_resource_roles.resource, bcs_roles.actions").
			Joins("left join bcs_roles on bcs_user_resource_roles.role_id = bcs_roles.id").Scan(&ura)

		mutex.Lock()
		PermissionsCache = make(map[uint][]UserPermissions)
		for _, v := range ura {
			up := UserPermissions{
				ResourceType: v.ResourceType,
				Resource:     v.Resource,
				Actions:      v.Actions,
			}
			PermissionsCache[v.UserId] = append(PermissionsCache[v.UserId], up)
		}
		mutex.Unlock()

		select {
		case <-ticker.C:
		}
	}
}

// GrantPermission grant permissions
func GrantPermission(request *restful.Request, response *restful.Response) {
	start := time.Now()

	//var form []PermissionForm
	var bp types.BcsPermission
	_ = request.ReadEntity(&bp)
	if bp.Kind != types.BcsDataType_PERMISSION {
		blog.Warnf("BcsPermission kind must be permission")
		message := fmt.Sprintf("errcode: %d, BcsPermission kind must be permission", common.BcsErrApiBadRequest)
		utils.WriteClientError(response, common.BcsErrApiBadRequest, message)
		return
	}
	if bp.APIVersion != "v1" {
		blog.Warnf("BcsPermission apiVersion must be v1")
		message := fmt.Sprintf("errcode: %d, BcsPermission apiVersion must be v1", common.BcsErrApiBadRequest)
		utils.WriteClientError(response, common.BcsErrApiBadRequest, message)
		return
	}

	for _, v := range bp.Spec.Permissions {
		if v.ResourceType == "" {
			metrics.ReportRequestAPIMetrics("GrantPermission", request.Request.Method, metrics.ErrStatus, start)
			blog.Warnf("resource_type must not be empty")
			message := fmt.Sprintf("errcode: %d, resource_type is empty", common.BcsErrApiBadRequest)
			utils.WriteClientError(response, common.BcsErrApiBadRequest, message)
			return
		}
		user := &models.BcsUser{
			Name: v.UserName,
		}
		userInDb := sqlstore.GetUserByCondition(user)
		if userInDb == nil {
			metrics.ReportRequestAPIMetrics("GrantPermission", request.Request.Method, metrics.ErrStatus, start)
			blog.Warnf("failed to grant permission to user [%s], user not exist", v.UserName)
			message := fmt.Sprintf("errcode: %d, failed to grant permission to user [%s], user not exist", common.BcsErrApiBadRequest, v.UserName)
			utils.WriteClientError(response, common.BcsErrApiBadRequest, message)
			return
		}
		roleInDb := sqlstore.GetRole(v.Role)
		if roleInDb == nil {
			metrics.ReportRequestAPIMetrics("GrantPermission", request.Request.Method, metrics.ErrStatus, start)
			blog.Warnf("failed to grant role [%s] permission to user [%s], role not exist", v.Role, v.UserName)
			message := fmt.Sprintf("errcode: %d, failed to grant role [%s] permission to user [%s], role not exist", common.BcsErrApiBadRequest, v.Role, v.UserName)
			utils.WriteClientError(response, common.BcsErrApiBadRequest, message)
			return
		}

		userResourceRole := &models.BcsUserResourceRole{
			UserId:       userInDb.ID,
			ResourceType: v.ResourceType,
			Resource:     v.Resource,
			RoleId:       roleInDb.ID,
		}
		urrInDb := sqlstore.GetUrrByCondition(userResourceRole)
		if urrInDb != nil {
			blog.Warnf("role [%s] of resourcetype [%s] and resource [%s] for user [%s] already exist", v.Role, v.ResourceType, v.Resource, v.UserName) //nolint
			continue
		}
		err := sqlstore.CreateUserResourceRole(userResourceRole)
		if err != nil {
			metrics.ReportRequestAPIMetrics("GrantPermission", request.Request.Method, metrics.ErrStatus, start)
			blog.Errorf("failed to grant role [%s] of resourcetype [%s] and resource [%s] to user [%s]: %s", v.Role, v.ResourceType, v.Resource, v.UserName, err.Error())                                                          //nolint
			message := fmt.Sprintf("errcode: %d, failed to grant role [%s] of resourcetype [%s] and resource [%s] to user [%s]: %s", common.BcsErrApiInternalDbError, v.Role, v.ResourceType, v.Resource, v.UserName, err.Error()) //nolint
			utils.WriteServerError(response, common.BcsErrApiInternalDbError, message)
			return
		}
	}
	data := utils.CreateResponeData(nil, "success", nil)
	response.Write([]byte(data))

	metrics.ReportRequestAPIMetrics("GrantPermission", request.Request.Method, metrics.SucStatus, start)
}

// GetPermission get permissions of a user for a resourceType
func GetPermission(request *restful.Request, response *restful.Response) {
	start := time.Now()

	var form GetPermissionForm
	_ = request.ReadEntity(&form)
	err := utils.Validate.Struct(&form)
	if err != nil {
		metrics.ReportRequestAPIMetrics("GetPermission", request.Request.Method, metrics.ErrStatus, start)
		_ = response.WriteHeaderAndEntity(400, utils.FormatValidationError(err))
		return
	}

	user := &models.BcsUser{
		Name: form.UserName,
	}
	userInDb := sqlstore.GetUserByCondition(user)
	if userInDb == nil {
		metrics.ReportRequestAPIMetrics("GetPermission", request.Request.Method, metrics.ErrStatus, start)
		blog.Warnf("user [%s] not found", form.UserName)
		message := fmt.Sprintf("errcode: %d, user [%s] not found", common.BcsErrApiBadRequest, form.UserName)
		utils.WriteClientError(response, common.BcsErrApiBadRequest, message)
		return
	}

	var permissions []PermissionsResp
	sqlstore.GCoreDB.Table("bcs_user_resource_roles").Select("bcs_user_resource_roles.resource_type, bcs_user_resource_roles.resource, bcs_roles.name as role").
		Joins("left join bcs_roles on bcs_user_resource_roles.role_id = bcs_roles.id where bcs_user_resource_roles.user_id = ? and bcs_user_resource_roles.resource_type = ?", userInDb.ID, form.ResourceType).Scan(&permissions)

	data := utils.CreateResponeData(nil, "success", permissions)
	response.Write([]byte(data))

	metrics.ReportRequestAPIMetrics("GetPermission", request.Request.Method, metrics.SucStatus, start)
}

// RevokePermission revoke permissions
func RevokePermission(request *restful.Request, response *restful.Response) {
	start := time.Now()

	//var form []PermissionForm
	var bp types.BcsPermission
	_ = request.ReadEntity(&bp)
	if bp.Kind != types.BcsDataType_PERMISSION {
		blog.Warnf("BcsPermission kind must be permission")
		message := fmt.Sprintf("errcode: %d, BcsPermission kind must be permission", common.BcsErrApiBadRequest)
		utils.WriteClientError(response, common.BcsErrApiBadRequest, message)
		return
	}
	if bp.APIVersion != "v1" {
		blog.Warnf("BcsPermission apiVersion must be v1")
		message := fmt.Sprintf("errcode: %d, BcsPermission apiVersion must be v1", common.BcsErrApiBadRequest)
		utils.WriteClientError(response, common.BcsErrApiBadRequest, message)
		return
	}
	for _, v := range bp.Spec.Permissions {
		user := &models.BcsUser{
			Name: v.UserName,
		}
		userInDb := sqlstore.GetUserByCondition(user)
		if userInDb == nil {
			metrics.ReportRequestAPIMetrics("RevokePermission", request.Request.Method, metrics.ErrStatus, start)
			blog.Warnf("failed to revoke permission of user [%s], user not exist", v.UserName)
			message := fmt.Sprintf("errcode: %d, failed to revoke permission of user [%s], user not exist", common.BcsErrApiBadRequest, v.UserName)
			utils.WriteClientError(response, common.BcsErrApiBadRequest, message)
			return
		}
		roleInDb := sqlstore.GetRole(v.Role)
		if roleInDb == nil {
			metrics.ReportRequestAPIMetrics("RevokePermission", request.Request.Method, metrics.ErrStatus, start)
			blog.Warnf("failed to revoke permission of role [%s] from user [%s], role not exist", v.Role, v.UserName)
			message := fmt.Sprintf("errcode: %d, failed to revoke permission of role [%s] from user [%s], role not exist", common.BcsErrApiBadRequest, v.Role, v.UserName)
			utils.WriteClientError(response, common.BcsErrApiBadRequest, message)
			return
		}

		userResourceRole := &models.BcsUserResourceRole{
			UserId:       userInDb.ID,
			ResourceType: v.ResourceType,
			Resource:     v.Resource,
			RoleId:       roleInDb.ID,
		}
		urrInDb := sqlstore.GetUrrByCondition(userResourceRole)
		if urrInDb == nil {
			blog.Warnf("userResourceRole not exist, user [%s], resource_type [%s], resource [%s], role [%s]", v.UserName, v.ResourceType, v.Resource, v.Role)
			continue
		}

		err := sqlstore.DeleteUserResourceRole(urrInDb)
		if err != nil {
			metrics.ReportRequestAPIMetrics("RevokePermission", request.Request.Method, metrics.ErrStatus, start)
			blog.Errorf("failed to delete role [%s] of resourcetype [%s] and resource [%s] from user [%s]: %s", v.Role, v.ResourceType, v.Resource, v.UserName, err.Error())                                                          //nolint
			message := fmt.Sprintf("errcode: %d, failed to delete role [%s] of resourcetype [%s] and resource [%s] from user [%s]: %s", common.BcsErrApiInternalDbError, v.Role, v.ResourceType, v.Resource, v.UserName, err.Error()) //nolint
			utils.WriteServerError(response, common.BcsErrApiInternalDbError, message)
			return
		}
	}

	data := utils.CreateResponeData(nil, "success", nil)
	response.Write([]byte(data))

	metrics.ReportRequestAPIMetrics("RevokePermission", request.Request.Method, metrics.SucStatus, start)
}

//VerifyPermission [GET] path /usermanager/v1/permissions/verify
func VerifyPermission(request *restful.Request, response *restful.Response) {
	start := time.Now()

	var form VerifyPermissionForm
	_ = request.ReadEntity(&form)
	err := utils.Validate.Struct(&form)
	if err != nil {
		blog.Errorf("formation of perssiom request from %s is invalid, %s", request.Request.RemoteAddr, err.Error())
		metrics.ReportRequestAPIMetrics("VerifyPermission", request.Request.Method, metrics.ErrStatus, start)
		_ = response.WriteHeaderAndEntity(400, utils.FormatValidationError(err))
		return
	}

	user, hasExpired := getUserFromToken(form.UserToken)
	if user == nil {
		blog.Warnf("usertoken [%s] is invalid from %s, type: %s, resource: %s",
			form.UserToken, request.Request.RemoteAddr, form.ResourceType, form.Resource)
		data := utils.CreateResponeData(nil, "success", &VerifyPermissionResponse{
			Allowed: false,
			Message: fmt.Sprintf("usertoken [%s] is invalid", form.UserToken),
		})
		_, _ = response.Write([]byte(data))
		return
	}
	if hasExpired {
		blog.Warnf("usertoken [%s] is expired from %s, type: %s, resource: %s",
			form.UserToken, request.Request.RemoteAddr, form.ResourceType, form.Resource)

		data := utils.CreateResponeData(nil, "success", &VerifyPermissionResponse{
			Allowed: false,
			Message: fmt.Sprintf("usertoken [%s] is expired", form.UserToken),
		})
		_, _ = response.Write([]byte(data))
		return
	}

	switch form.ResourceType {
	case "cluster", "storage":
		allowed, message := verifyResourceReplica(user.ID, form.ResourceType, form.Resource, form.Action)

		data := utils.CreateResponeData(nil, "success", &VerifyPermissionResponse{
			Allowed: allowed,
			Message: message,
		})
		blog.Infof("user %s access to type: %s, resource: %s, action: %s, permission: %t",
			user.Name, form.ResourceType, form.Resource, form.Action, allowed)
		_, _ = response.Write([]byte(data))
	default:
		allowed, message := verifyResourceReplica(user.ID, form.ResourceType, "", form.Action)

		data := utils.CreateResponeData(nil, "success", &VerifyPermissionResponse{
			Allowed: allowed,
			Message: message,
		})
		blog.Infof("user %s access to type: %s, action: %s, permission: %t",
			user.Name, form.ResourceType, form.Action, allowed)
		_, _ = response.Write([]byte(data))
	}

	metrics.ReportRequestAPIMetrics("VerifyPermission", request.Request.Method, metrics.SucStatus, start)
}

// verifyResourceReplica verify whether a user have permission for s resource, return true or false
func verifyResourceReplica(userID uint, resourceType, resource, action string) (bool, string) {
	var op []OwnedPermissions
	if resource == "" {
		//sqlstore.GCoreDB.Table("bcs_user_resource_roles").Select("bcs_roles.actions").
		//	Joins("left join bcs_roles on bcs_user_resource_roles.role_id = bcs_roles.id where bcs_user_resource_roles.user_id = ? and bcs_user_resource_roles.resource_type = ?", userId, resourceType).Scan(&op) //nolint

		mutex.RLock()
		for _, v := range PermissionsCache[userID] {
			if v.ResourceType == resourceType {
				op = append(op, OwnedPermissions{Actions: v.Actions})
			}
		}
		mutex.RUnlock()
	} else {
		//sqlstore.GCoreDB.Table("bcs_user_resource_roles").Select("bcs_roles.actions").
		//	Joins("left join bcs_roles on bcs_user_resource_roles.role_id = bcs_roles.id where bcs_user_resource_roles.user_id = ? and bcs_user_resource_roles.resource_type = ?
		//	and (bcs_user_resource_roles.resource = ? or bcs_user_resource_roles.resource = ?)", userId, resourceType, resource, "*").Scan(&op) //nolint

		mutex.RLock()
		for _, v := range PermissionsCache[userID] {
			if v.ResourceType == resourceType && (v.Resource == resource || v.Resource == "*") {
				op = append(op, OwnedPermissions{Actions: v.Actions})
			}
		}
		mutex.RUnlock()
	}
	for _, p := range op {
		actions := strings.Split(p.Actions, ",")
		for _, a := range actions {
			if action == a {
				return true, ""
			}
		}
	}
	return false, "no permission"
}

func getUserFromToken(s string) (*models.BcsUser, bool) {
	u := models.BcsUser{
		UserToken: s,
	}
	user := sqlstore.GetUserByCondition(&u)

	if user == nil {
		return nil, false
	}

	if user.HasExpired() {
		return user, true
	}

	return user, false
}
