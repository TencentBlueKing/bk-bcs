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

// Package permission xxx
package permission

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common"
	"github.com/Tencent/bk-bcs/bcs-common/common/types"
	restful "github.com/emicklei/go-restful/v3"

	blog "github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/pkg/log"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/pkg/metrics"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/models"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/storages/sqlstore"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/config"
)

// PermissionForm registe form
// nolint
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
	UserToken    string       `json:"user_token" validate:"required"`
	ResourceType ResourceType `json:"resource_type" validate:"required"`
	Resource     string       `json:"resource"`
	Action       string       `json:"action" validate:"required"`
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

// UserResourceAction resource operation action
type UserResourceAction struct {
	UserId       uint
	ResourceType ResourceType
	Resource     string
	Actions      string
}

// UserPermissions user permission definition
type UserPermissions struct {
	ResourceType ResourceType
	Resource     string
	Actions      string
}

// PermissionsCache local cache for speed up
var PermissionsCache map[uint][]UserPermissions

// Mutex rwLock
var Mutex *sync.RWMutex

// InitCache sync data from db to cache periodically
func InitCache() {
	// init bcs roles
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
	// init roles
	// create roles
	for _, role := range initRoles {
		m := sqlstore.GetRole(role.Name)
		if m == nil {
			err := sqlstore.CreateRole(&role)
			if err != nil {
				blog.Log(context.Background()).Errorf("Failed to init role [%s]: %s", role.Name, err.Error())
			}
		}
	}

	// user resource
	Mutex = new(sync.RWMutex)
	var ura []UserResourceAction
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()
	for {
		// get role from db
		sqlstore.GCoreDB.Table("bcs_user_resource_roles").Select(
			"bcs_user_resource_roles.user_id, bcs_user_resource_roles.resource_type, bcs_user_resource_roles." +
				"resource, bcs_roles.actions").
			Joins("left join bcs_roles on bcs_user_resource_roles.role_id = bcs_roles.id").Scan(&ura)

		// set cache mutex lock
		Mutex.Lock()
		// cache permission
		PermissionsCache = make(map[uint][]UserPermissions)
		for _, v := range ura {
			up := UserPermissions{
				ResourceType: v.ResourceType,
				Resource:     v.Resource,
				Actions:      v.Actions,
			}
			PermissionsCache[v.UserId] = append(PermissionsCache[v.UserId], up)
		}
		Mutex.Unlock()

		// wait to get roles
		// nolint
		select {
		case <-ticker.C:
		}
	}
}

// GrantPermission grant permissions
// NOCC:CCN_thresholde(设计如此),golint/fnsize(设计如此)
func GrantPermission(request *restful.Request, response *restful.Response) {
	start := time.Now()
	ctx := request.Request.Context()

	// var form []PermissionForm
	var bp types.BcsPermission
	_ = request.ReadEntity(&bp)
	if bp.Kind != types.BcsDataType_PERMISSION {
		blog.Log(ctx).Warnf("BcsPermission kind must be permission")
		message := fmt.Sprintf("errcode: %d, BcsPermission kind must be permission", common.BcsErrApiBadRequest)
		utils.WriteClientError(response, common.BcsErrApiBadRequest, message)
		metrics.ReportRequestAPIMetrics("GrantPermission", request.Request.Method, metrics.ErrStatus, start)
		return
	}
	// check apiVersion
	if bp.APIVersion != "v1" {
		blog.Log(ctx).Warnf("BcsPermission apiVersion must be v1")
		message := fmt.Sprintf("errcode: %d, BcsPermission apiVersion must be v1", common.BcsErrApiBadRequest)
		utils.WriteClientError(response, common.BcsErrApiBadRequest, message)
		metrics.ReportRequestAPIMetrics("GrantPermission", request.Request.Method, metrics.ErrStatus, start)
		return
	}

	// grant permission
	for _, v := range bp.Spec.Permissions {
		if v.ResourceType == "" {
			metrics.ReportRequestAPIMetrics("GrantPermission", request.Request.Method, metrics.ErrStatus, start)
			blog.Log(ctx).Warnf("resource_type must not be empty")
			message := fmt.Sprintf("errcode: %d, resource_type is empty", common.BcsErrApiBadRequest)
			utils.WriteClientError(response, common.BcsErrApiBadRequest, message)
			return
		}
		user := &models.BcsUser{
			Name: v.UserName,
		}
		// get user
		userInDb := sqlstore.GetUserByCondition(user)
		if userInDb == nil {
			metrics.ReportRequestAPIMetrics("GrantPermission", request.Request.Method, metrics.ErrStatus, start)
			blog.Log(ctx).Warnf("failed to grant permission to user [%s], user not exist", v.UserName)
			message := fmt.Sprintf("errcode: %d, failed to grant permission to user [%s], user not exist",
				common.BcsErrApiBadRequest, v.UserName)
			utils.WriteClientError(response, common.BcsErrApiBadRequest, message)
			return
		}
		// get role
		roleInDb := sqlstore.GetRole(v.Role)
		if roleInDb == nil {
			metrics.ReportRequestAPIMetrics("GrantPermission", request.Request.Method, metrics.ErrStatus, start)
			blog.Log(ctx).Warnf("failed to grant role [%s] permission to user [%s], role not exist", v.Role, v.UserName)
			message := fmt.Sprintf("errcode: %d, failed to grant role [%s] permission to user [%s], role not exist",
				common.BcsErrApiBadRequest, v.Role, v.UserName)
			utils.WriteClientError(response, common.BcsErrApiBadRequest, message)
			return
		}

		// get user resource role
		userResourceRole := &models.BcsUserResourceRole{
			UserId:       userInDb.ID,
			ResourceType: v.ResourceType,
			Resource:     v.Resource,
			RoleId:       roleInDb.ID,
		}
		urrInDb := sqlstore.GetUrrByCondition(userResourceRole)
		if urrInDb != nil {
			blog.Log(ctx).Warnf("role [%s] of resourcetype [%s] and resource [%s] for user [%s] already exist",
				v.Role, v.ResourceType, v.Resource, v.UserName)
			continue
		}
		// create user resource role
		err := sqlstore.CreateUserResourceRole(userResourceRole)
		if err != nil {
			metrics.ReportRequestAPIMetrics("GrantPermission", request.Request.Method, metrics.ErrStatus, start)
			blog.Log(ctx).Errorf("failed to grant role [%s] of resourcetype [%s] and resource [%s] to user [%s]: %s",
				v.Role,
				v.ResourceType, v.Resource, v.UserName, err.Error()) //nolint
			message := fmt.Sprintf(
				"errcode: %d, failed to grant role [%s] of resourcetype [%s] and resource [%s] to user [%s]: %s",
				common.BcsErrApiInternalDbError, v.Role, v.ResourceType, v.Resource, v.UserName, err.Error()) //nolint
			utils.WriteServerError(response, common.BcsErrApiInternalDbError, message)
			return
		}
	}
	// response
	data := utils.CreateResponseData(nil, "success", nil)
	_, _ = response.Write([]byte(data))

	metrics.ReportRequestAPIMetrics("GrantPermission", request.Request.Method, metrics.SucStatus, start)
}

// GetPermission get permissions of a user for a resourceType
func GetPermission(request *restful.Request, response *restful.Response) {
	start := time.Now()
	ctx := request.Request.Context()

	// parse permission form
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
	// get user from db
	userInDb := sqlstore.GetUserByCondition(user)
	if userInDb == nil {
		metrics.ReportRequestAPIMetrics("GetPermission", request.Request.Method, metrics.ErrStatus, start)
		blog.Log(ctx).Warnf("user [%s] not found", form.UserName)
		message := fmt.Sprintf("errcode: %d, user [%s] not found", common.BcsErrApiBadRequest, form.UserName)
		utils.WriteClientError(response, common.BcsErrApiBadRequest, message)
		return
	}

	// get user permissions
	var permissions []PermissionsResp
	sqlstore.GCoreDB.Table("bcs_user_resource_roles").Select(
		"bcs_user_resource_roles.resource_type, bcs_user_resource_roles.resource, bcs_roles.name as role").
		Joins(
			"left join bcs_roles on bcs_user_resource_roles.role_id = bcs_roles.id where bcs_user_resource_roles."+
				"user_id = ? and bcs_user_resource_roles.resource_type = ?", userInDb.ID, form.ResourceType).
		Scan(&permissions)

	data := utils.CreateResponseData(nil, "success", permissions)
	_, _ = response.Write([]byte(data))

	metrics.ReportRequestAPIMetrics("GetPermission", request.Request.Method, metrics.SucStatus, start)
}

// RevokePermission revoke permissions
func RevokePermission(request *restful.Request, response *restful.Response) {
	start := time.Now()
	ctx := request.Request.Context()

	// var form []PermissionForm
	var bp types.BcsPermission
	_ = request.ReadEntity(&bp)
	if bp.Kind != types.BcsDataType_PERMISSION {
		blog.Log(ctx).Warnf("BcsPermission kind must be permission")
		message := fmt.Sprintf("errcode: %d, BcsPermission kind must be permission", common.BcsErrApiBadRequest)
		utils.WriteClientError(response, common.BcsErrApiBadRequest, message)
		metrics.ReportRequestAPIMetrics("RevokePermission", request.Request.Method, metrics.ErrStatus, start)
		return
	}
	// check apiVersion
	if bp.APIVersion != "v1" {
		blog.Log(ctx).Warnf("BcsPermission apiVersion must be v1")
		message := fmt.Sprintf("errcode: %d, BcsPermission apiVersion must be v1", common.BcsErrApiBadRequest)
		utils.WriteClientError(response, common.BcsErrApiBadRequest, message)
		metrics.ReportRequestAPIMetrics("RevokePermission", request.Request.Method, metrics.ErrStatus, start)
		return
	}

	// get user permission
	for _, v := range bp.Spec.Permissions {
		user := &models.BcsUser{
			Name: v.UserName,
		}
		userInDb := sqlstore.GetUserByCondition(user)
		if userInDb == nil {
			metrics.ReportRequestAPIMetrics("RevokePermission", request.Request.Method, metrics.ErrStatus, start)
			blog.Log(ctx).Warnf("failed to revoke permission of user [%s], user not exist", v.UserName)
			message := fmt.Sprintf("errcode: %d, failed to revoke permission of user [%s], user not exist",
				common.BcsErrApiBadRequest, v.UserName)
			utils.WriteClientError(response, common.BcsErrApiBadRequest, message)
			return
		}
		// get role
		roleInDb := sqlstore.GetRole(v.Role)
		if roleInDb == nil {
			metrics.ReportRequestAPIMetrics("RevokePermission", request.Request.Method, metrics.ErrStatus, start)
			blog.Log(ctx).Warnf("failed to revoke permission of role [%s] from user [%s], role not exist",
				v.Role, v.UserName)
			message := fmt.Sprintf("errcode: %d, failed to revoke permission of role [%s] from user [%s], "+
				"role not exist",
				common.BcsErrApiBadRequest, v.Role, v.UserName)
			utils.WriteClientError(response, common.BcsErrApiBadRequest, message)
			return
		}

		// get bcs user resource role
		userResourceRole := &models.BcsUserResourceRole{
			UserId:       userInDb.ID,
			ResourceType: v.ResourceType,
			Resource:     v.Resource,
			RoleId:       roleInDb.ID,
		}
		urrInDb := sqlstore.GetUrrByCondition(userResourceRole)
		if urrInDb == nil {
			blog.Log(ctx).Warnf("userResourceRole not exist, user [%s], resource_type [%s], resource [%s], role [%s]",
				v.UserName, v.ResourceType, v.Resource, v.Role)
			continue
		}

		// delete user resource role
		err := sqlstore.DeleteUserResourceRole(urrInDb)
		if err != nil {
			metrics.ReportRequestAPIMetrics("RevokePermission", request.Request.Method, metrics.ErrStatus, start)
			blog.Log(ctx).Errorf("failed to delete role [%s] of resourcetype [%s] and resource [%s] from user [%s]: %s",
				v.Role,
				v.ResourceType, v.Resource, v.UserName, err.Error()) //nolint
			message := fmt.Sprintf(
				"errcode: %d, failed to delete role [%s] of resourcetype [%s] and resource [%s] from user [%s]: %s",
				common.BcsErrApiInternalDbError, v.Role, v.ResourceType, v.Resource, v.UserName, err.Error()) //nolint
			utils.WriteServerError(response, common.BcsErrApiInternalDbError, message)
			return
		}
	}

	data := utils.CreateResponseData(nil, "success", nil)
	_, _ = response.Write([]byte(data))

	metrics.ReportRequestAPIMetrics("RevokePermission", request.Request.Method, metrics.SucStatus, start)
}

// VerifyPermission [GET] path /usermanager/v1/permissions/verify
func VerifyPermission(request *restful.Request, response *restful.Response) {
	start := time.Now()
	ctx := request.Request.Context()

	// parse permission form
	var form VerifyPermissionForm
	_ = request.ReadEntity(&form)
	err := utils.Validate.Struct(&form)
	if err != nil {
		blog.Log(ctx).Errorf("formation of perssiom request from %s is invalid, %s", request.Request.RemoteAddr,
			err.Error())
		metrics.ReportRequestAPIMetrics("VerifyPermission", request.Request.Method, metrics.ErrStatus, start)
		_ = response.WriteHeaderAndEntity(400, utils.FormatValidationError(err))
		return
	}

	// get user from token
	user, hasExpired := getUserFromToken(form.UserToken)
	if user == nil {
		blog.Log(ctx).Warnf("usertoken [%s] is invalid from %s, type: %s, resource: %s",
			form.UserToken, request.Request.RemoteAddr, form.ResourceType, form.Resource)
		data := utils.CreateResponseData(nil, "success", &VerifyPermissionResponse{
			Allowed: false,
			Message: fmt.Sprintf("usertoken [%s] is invalid", form.UserToken),
		})
		_, _ = response.Write([]byte(data))
		metrics.ReportRequestAPIMetrics("VerifyPermission", request.Request.Method, metrics.ErrStatus, start)
		return
	}
	// check token expired
	if hasExpired {
		blog.Log(ctx).Warnf("usertoken [%s] is expired from %s, type: %s, resource: %s",
			form.UserToken, request.Request.RemoteAddr, form.ResourceType, form.Resource)

		data := utils.CreateResponseData(nil, "success", &VerifyPermissionResponse{
			Allowed: false,
			Message: fmt.Sprintf("usertoken [%s] is expired", form.UserToken),
		})
		_, _ = response.Write([]byte(data))
		metrics.ReportRequestAPIMetrics("VerifyPermission", request.Request.Method, metrics.ErrStatus, start)
		return
	}

	// check resource type
	switch form.ResourceType {
	case Cluster, Storage:
		// verify old permission
		allowed, message := verifyResourceReplica(user.ID, form.ResourceType, form.Resource, form.Action)

		data := utils.CreateResponseData(nil, "success", &VerifyPermissionResponse{
			Allowed: allowed,
			Message: message,
		})
		blog.Log(ctx).Infof("user %s access to type: %s, resource: %s, action: %s, permission: %t",
			user.Name, form.ResourceType, form.Resource, form.Action, allowed)
		_, _ = response.Write([]byte(data))
	default:
		// verify default permission
		allowed, message := verifyResourceReplica(user.ID, form.ResourceType, "", form.Action)

		data := utils.CreateResponseData(nil, "success", &VerifyPermissionResponse{
			Allowed: allowed,
			Message: message,
		})
		blog.Log(ctx).Infof("user %s access to type: %s, action: %s, permission: %t",
			user.Name, form.ResourceType, form.Action, allowed)
		_, _ = response.Write([]byte(data))
	}

	metrics.ReportRequestAPIMetrics("VerifyPermission", request.Request.Method, metrics.SucStatus, start)
}

// verifyResourceReplica verify whether a user have permission for s resource, return true or false
func verifyResourceReplica(userID uint, resourceType ResourceType, resource, action string) (bool, string) {
	var op []OwnedPermissions
	if resource == "" {

		Mutex.RLock()
		for _, v := range PermissionsCache[userID] {
			if v.ResourceType == resourceType {
				op = append(op, OwnedPermissions{Actions: v.Actions})
			}
		}
		Mutex.RUnlock()
	} else {

		// cache permission
		Mutex.RLock()
		for _, v := range PermissionsCache[userID] {
			if v.ResourceType == resourceType && (v.Resource == resource || v.Resource == "*") {
				op = append(op, OwnedPermissions{Actions: v.Actions})
			}
		}
		Mutex.RUnlock()
	}
	// get resource
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

// getUserInfoByToken get user info by token
func getUserInfoByToken(ctx context.Context, s string) (*models.BcsUser, bool, bool) {
	user, hasExpired := getUserFromToken(s)
	if user != nil {
		blog.Log(ctx).Infof("getUserInfoByToken getUserFromToken for %s success", user.Name)
		return user, false, hasExpired
	}

	// get user from temp token
	tempToken, hasExpired := getUserFromTempToken(s)
	if tempToken != nil {
		blog.Log(ctx).Infof("getUserInfoByToken getUserFromTempToken for %s success", tempToken.Username)
		return &models.BcsUser{
			ID:        tempToken.ID,
			Name:      tempToken.Username,
			UserType:  tempToken.UserType,
			UserToken: tempToken.Token,
			CreatedBy: tempToken.CreatedBy,
			CreatedAt: tempToken.CreatedAt,
			UpdatedAt: tempToken.UpdatedAt,
			ExpiresAt: tempToken.ExpiresAt,
		}, true, hasExpired
	}

	blog.Log(ctx).Errorf("getUserInfoByToken failed: invalid token[%s]", s)
	return nil, false, false
}

// get user form token
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

// getUserFromTempToken
func getUserFromTempToken(s string) (*models.BcsTempToken, bool) {
	token := &models.BcsTempToken{
		Token: s,
	}
	tokenStore := sqlstore.NewTokenStore(sqlstore.GCoreDB, config.GlobalCryptor)
	tempUser := tokenStore.GetTempTokenByCondition(token)
	if tempUser == nil {
		return nil, false
	}

	if tempUser.HasExpired() {
		return tempUser, true
	}

	return tempUser, false
}

// verifyPermissionV1
func verifyPermissionV1(ctx context.Context, user *models.BcsUser, req VerifyPermissionReq) (bool, string) {
	switch req.ResourceType {
	case Cluster, Storage:
		/// verifyResourceReplica
		allowed, message := verifyResourceReplica(user.ID, req.ResourceType, req.Resource, req.Action)
		blog.Log(ctx).Infof("user %s access to type: %s, resource: %s, action: %s, permission: %t",
			user.Name, req.ResourceType, req.Resource, req.Action, allowed)
		return allowed, message
	default:
		// verifyResourceReplica
		allowed, message := verifyResourceReplica(user.ID, req.ResourceType, "", req.Action)
		blog.Log(ctx).Infof("user %s access to type: %s, action: %s, permission: %t",
			user.Name, req.ResourceType, req.Action, allowed)

		return allowed, message
	}
}

// VerifyPermissionV2 [GET] path /usermanager/v2/permissions/verify
// NOCC:golint/fnsize(设计如此)
func (cli *PermVerifyClient) VerifyPermissionV2(request *restful.Request, response *restful.Response) {
	start := time.Now()
	ctx := request.Request.Context()

	// parse permission req
	var req VerifyPermissionReq
	_ = request.ReadEntity(&req)
	err := utils.Validate.Struct(&req)
	if err != nil {
		blog.Log(request.Request.Context()).Errorf("formation of permission request from %s is invalid, %s",
			request.Request.RemoteAddr, err.Error())
		metrics.ReportRequestAPIMetrics("VerifyPermissionV2", request.Request.Method, metrics.ErrStatus, start)
		_ = response.WriteHeaderAndEntity(400, utils.FormatValidationError(err))
		return
	}

	// check request
	err = req.validate()
	if err != nil {
		blog.Log(ctx).Errorf("VerifyPermissionV2 permission request from %s is invalid, %s", request.Request.RemoteAddr,
			err.Error())
		metrics.ReportRequestAPIMetrics("VerifyPermissionV2", request.Request.Method, metrics.ErrStatus, start)
		utils.WriteClientError(response, common.BcsErrApiBadRequest, err.Error())
		return
	}

	// permission switch for special case
	if cli.PermSwitch {
		switchPermission(ctx, request, response, start)
		return
	}

	// userInfo by token
	user, temp, hasExpired := getUserInfoByToken(ctx, req.UserToken)
	if user == nil {
		blog.Log(ctx).Warnf("AuthToken [%s] is invalid from %s, type: %s, resource: %s",
			req.UserToken, request.Request.RemoteAddr, req.ResourceType, req.Resource)
		data := utils.CreateResponseData(nil, "success", &VerifyPermissionResponse{
			Allowed: false,
			Message: fmt.Sprintf("AuthToken [%s] is invalid", req.UserToken),
		})
		_, _ = response.Write([]byte(data))
		metrics.ReportRequestAPIMetrics("VerifyPermissionV2", request.Request.Method, metrics.ErrStatus, start)
		return
	}
	// check token expired
	if hasExpired {
		blog.Log(ctx).Warnf("AuthToken [%s] is expired from %s, type: %s, resource: %s",
			req.UserToken, request.Request.RemoteAddr, req.ResourceType, req.Resource)

		data := utils.CreateResponseData(nil, "success", &VerifyPermissionResponse{
			Allowed: false,
			Message: fmt.Sprintf("usertoken [%s] is expired", req.UserToken),
		})
		_, _ = response.Write([]byte(data))
		metrics.ReportRequestAPIMetrics("VerifyPermissionV2", request.Request.Method, metrics.ErrStatus, start)
		return
	}

	// skip permission if user is admin
	if user.IsAdmin() {
		data := utils.CreateResponseData(nil, "success", &VerifyPermissionResponse{
			Allowed: true,
			Message: "admin user skip cluster permission check",
		})
		blog.Log(ctx).Infof("admin user %s access to type: %s, permission: %t", user.Name, req.ResourceType, true)
		_, _ = response.Write([]byte(data))

		metrics.ReportRequestAPIMetrics("VerifyPermissionV2", request.Request.Method, metrics.SucStatus, start)
		return
	}

	// v2 permission will be compatible with v1 permission
	if !temp {
		allowed, message := verifyPermissionV1(ctx, user, req)
		if allowed {
			data := utils.CreateResponseData(nil, "success", &VerifyPermissionResponse{
				Allowed: allowed,
				Message: message,
			})
			_, _ = response.Write([]byte(data))
			metrics.ReportRequestAPIMetrics("VerifyPermissionV2", request.Request.Method, metrics.SucStatus, start)
			return
		}
	}

	// VerifyPermissionV2
	cli.verifyV2Permission(ctx, req, user, response)
	metrics.ReportRequestAPIMetrics("VerifyPermissionV2", request.Request.Method, metrics.SucStatus, start)
}

// switch permission
func switchPermission(ctx context.Context, request *restful.Request, response *restful.Response, start time.Time) {
	blog.Log(ctx).Infof("VerifyPermissionV2 permission from %s, switch is true", request.Request.RemoteAddr)
	metrics.ReportRequestAPIMetrics("VerifyPermissionV2", request.Request.Method, metrics.SucStatus, start)
	data := utils.CreateResponseData(nil, "success", &VerifyPermissionResponse{
		Allowed: true,
		Message: "",
	})
	_, _ = response.Write([]byte(data))
}

// verify v2 permission
func (cli *PermVerifyClient) verifyV2Permission(ctx context.Context, req VerifyPermissionReq, user *models.BcsUser,
	response *restful.Response) {
	switch req.ResourceType {
	case Cluster:
		resource := ClusterResource{
			ClusterType: req.ClusterType,
			ProjectID:   req.ProjectID,
			ClusterID:   req.ClusterID,
			URL:         req.RequestURL,
		}

		// cluster permission
		blog.Log(ctx).Infof("user %s access to type: %s, resource: [%s]:[%s], action: %s, url: %s, project: %s",
			user.Name, "cluster", resource.ClusterType, resource.ClusterID, req.Action, req.RequestURL,
			req.ProjectID)

		allowed, message := cli.VerifyClusterPermission(ctx, user, req.Action, resource)
		data := utils.CreateResponseData(nil, "success", &VerifyPermissionResponse{
			Allowed: allowed,
			Message: message,
		})
		blog.Log(ctx).Infof("user %s access to type: %s, resource: %s, action: %s, permission: %t",
			user.Name, "cluster", resource.ClusterType, req.Action, allowed)
		_, _ = response.Write([]byte(data))
	// verify storage permission
	case Storage:
		allowed, message := verifyResourceReplica(user.ID, req.ResourceType, req.Resource, req.Action)

		data := utils.CreateResponseData(nil, "success", &VerifyPermissionResponse{
			Allowed: allowed,
			Message: message,
		})
		blog.Log(ctx).Infof("user %s access to type: %s, resource: %s, action: %s, permission: %t",
			user.Name, req.ResourceType, req.Resource, req.Action, allowed)
		_, _ = response.Write([]byte(data))
	default:
		// verify default permission
		allowed, message := verifyResourceReplica(user.ID, req.ResourceType, "", req.Action)

		data := utils.CreateResponseData(nil, "success", &VerifyPermissionResponse{
			Allowed: allowed,
			Message: message,
		})
		blog.Log(ctx).Infof("user %s access to type: %s, action: %s, permission: %t",
			user.Name, req.ResourceType, req.Action, allowed)
		_, _ = response.Write([]byte(data))
	}
}
