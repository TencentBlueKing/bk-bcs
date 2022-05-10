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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/pkg/esb/cmdb"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/pkg/jwt"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/v1http/auth"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/v1http/cluster"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/v1http/credential"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/v1http/permission"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/v1http/tke"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/v1http/token"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/v1http/user"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/storages/cache"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/storages/sqlstore"
	"github.com/emicklei/go-restful"
)

// InitV1Routers init v1 version route,
// it's compatible with bcs-api
func InitV1Routers(ws *restful.WebService, service *permission.PermVerifyClient) {
	initUsersRouters(ws)
	initClustersRouters(ws)
	initTkeRouters(ws)
	initPermissionRouters(ws, service)
	initTokenRouters(ws)
	initExtraTokenRouters(ws, service)
}

// initUsersRouters init users api routers
func initUsersRouters(ws *restful.WebService) {
	ws.Route(auth.AdminAuthFunc(ws.POST("/v1/users/admin/{user_name}")).To(user.CreateAdminUser))
	ws.Route(auth.AdminAuthFunc(ws.GET("/v1/users/admin/{user_name}")).To(user.GetAdminUser))

	ws.Route(auth.AdminAuthFunc(ws.POST("/v1/users/saas/{user_name}")).To(user.CreateSaasUser))
	ws.Route(auth.AdminAuthFunc(ws.GET("/v1/users/saas/{user_name}")).To(user.GetSaasUser))
	ws.Route(auth.AdminAuthFunc(ws.PUT("/v1/users/saas/{user_name}/refresh")).To(user.RefreshSaasToken))

	ws.Route(auth.AuthFunc(ws.POST("/v1/users/plain/{user_name}")).To(user.CreatePlainUser))
	ws.Route(auth.AuthFunc(ws.GET("/v1/users/plain/{user_name}")).To(user.GetPlainUser))
	ws.Route(auth.AuthFunc(ws.PUT("/v1/users/plain/{user_name}/refresh/{expire_time}")).To(user.RefreshPlainToken))
}

// initClustersRouters init cluster api routers
func initClustersRouters(ws *restful.WebService) {
	ws.Route(auth.AdminAuthFunc(ws.POST("/v1/clusters")).To(cluster.CreateCluster))

	ws.Route(auth.AdminAuthFunc(ws.POST("/v1/clusters/{cluster_id}/register_tokens")).To(token.CreateRegisterToken))
	ws.Route(auth.AdminAuthFunc(ws.GET("/v1/clusters/{cluster_id}/register_tokens")).To(token.GetRegisterToken))

	ws.Route(ws.PUT("/v1/clusters/{cluster_id}/credentials").To(credential.UpdateCredentials))
	ws.Route(auth.AdminAuthFunc(ws.GET("/v1/clusters/{cluster_id}/credentials")).To(credential.GetCredentials))

	ws.Route(auth.AdminAuthFunc(ws.GET("/v1/clusters/credentials")).To(credential.ListCredentials))
}

// initPermissionRouters init permission api routers
func initPermissionRouters(ws *restful.WebService, service *permission.PermVerifyClient) {
	ws.Route(auth.AdminAuthFunc(ws.POST("/v1/permissions")).To(permission.GrantPermission))
	ws.Route(auth.AdminAuthFunc(ws.GET("/v1/permissions")).To(permission.GetPermission))
	ws.Route(auth.AdminAuthFunc(ws.DELETE("/v1/permissions")).To(permission.RevokePermission))

	ws.Route(auth.AdminAuthFunc(ws.GET("/v1/permissions/verify")).To(permission.VerifyPermission))
	ws.Route(auth.AdminAuthFunc(ws.GET("/v2/permissions/verify")).To(service.VerifyPermissionV2))
}

// initTokenRouters init bcs token
func initTokenRouters(ws *restful.WebService) {
	tokenHandler := token.NewTokenHandler(sqlstore.NewTokenStore(sqlstore.GCoreDB),
		sqlstore.NewTokenNotifyStore(sqlstore.GCoreDB), cache.RDB, jwt.JWTClient)
	ws.Route(auth.TokenAuthFunc(ws.POST("/v1/tokens").To(tokenHandler.CreateToken)))
	ws.Route(auth.TokenAuthFunc(ws.GET("/v1/users/{username}/tokens").To(tokenHandler.GetToken)))
	ws.Route(auth.TokenAuthFunc(ws.DELETE("/v1/tokens/{token}").To(tokenHandler.DeleteToken)))
	ws.Route(auth.TokenAuthFunc(ws.PUT("/v1/tokens/{token}").To(tokenHandler.UpdateToken)))
	// for Temporary Token
	ws.Route(auth.TokenAuthFunc(ws.POST("/v1/tokens/temp").To(tokenHandler.CreateTempToken)))
	ws.Route(auth.TokenAuthFunc(ws.POST("/v1/tokens/client").To(tokenHandler.CreateClientToken)))
}

// initExtraTokenRouters init bcs extra token for third-party system
func initExtraTokenRouters(ws *restful.WebService, service *permission.PermVerifyClient) {
	tokenHandler := token.NewExtraTokenHandler(sqlstore.NewTokenStore(sqlstore.GCoreDB),
		sqlstore.NewTokenNotifyStore(sqlstore.GCoreDB), cache.RDB, jwt.JWTClient, service.ClusterClient, cmdb.CMDBClient)
	ws.Route(ws.GET("/v1/tokens/extra/getClusterUserToken").To(tokenHandler.GetTokenByUserAndClusterID))
}

// initTkeRouters init tke api routers
func initTkeRouters(ws *restful.WebService) {
	ws.Route(auth.AdminAuthFunc(ws.POST("/v1/tke/cidr/add_cidr")).To(tke.AddTkeCidr))
	ws.Route(auth.AdminAuthFunc(ws.POST("/v1/tke/cidr/apply_cidr")).To(tke.ApplyTkeCidr))
	ws.Route(auth.AdminAuthFunc(ws.POST("/v1/tke/cidr/release_cidr")).To(tke.ReleaseTkeCidr))
	ws.Route(auth.AdminAuthFunc(ws.POST("/v1/tke/cidr/list_count")).To(tke.ListTkeCidr))

	ws.Route(auth.AdminAuthFunc(ws.POST("/v1/tke/{cluster_id}/sync_credentials")).To(tke.SyncTkeClusterCredentials))
}
