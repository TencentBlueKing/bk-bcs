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
	"github.com/emicklei/go-restful"
)

//InitV1Routers init v1 version route,
// it's compatable with bcs-api
func InitV1Routers(ws *restful.WebService) {
	initUsersRouters(ws)
	initClustersRouters(ws)
	initTkeRouters(ws)
	initPermissionRouters(ws)

	go initCache()
}

// initUsersRouters init users api routers
func initUsersRouters(ws *restful.WebService) {
	ws.Route(AdminAuthFunc(ws.POST("/v1/users/admin/{user_name}")).To(CreateAdminUser))
	ws.Route(AdminAuthFunc(ws.GET("/v1/users/admin/{user_name}")).To(GetAdminUser))

	ws.Route(AdminAuthFunc(ws.POST("/v1/users/saas/{user_name}")).To(CreateSaasUser))
	ws.Route(AdminAuthFunc(ws.GET("/v1/users/saas/{user_name}")).To(GetSaasUser))
	ws.Route(AdminAuthFunc(ws.PUT("/v1/users/saas/{user_name}/refresh")).To(RefreshSaasToken))

	ws.Route(AuthFunc(ws.POST("/v1/users/plain/{user_name}")).To(CreatePlainUser))
	ws.Route(AuthFunc(ws.GET("/v1/users/plain/{user_name}")).To(GetPlainUser))
	ws.Route(AuthFunc(ws.PUT("/v1/users/plain/{user_name}/refresh/{expire_time}")).To(RefreshPlainToken))
}

// initClustersRouters init cluster api routers
func initClustersRouters(ws *restful.WebService) {
	ws.Route(AdminAuthFunc(ws.POST("/v1/clusters")).To(CreateCluster))

	ws.Route(AdminAuthFunc(ws.POST("/v1/clusters/{cluster_id}/register_tokens")).To(CreateRegisterToken))
	ws.Route(AdminAuthFunc(ws.GET("/v1/clusters/{cluster_id}/register_tokens")).To(GetRegisterToken))

	ws.Route(ws.PUT("/v1/clusters/{cluster_id}/credentials").To(UpdateCredentials))
	ws.Route(AdminAuthFunc(ws.GET("/v1/clusters/{cluster_id}/credentials")).To(GetCredentials))

	ws.Route(AdminAuthFunc(ws.GET("/v1/clusters/credentials")).To(ListCredentials))
}

// initPermissionRouters init permission api routers
func initPermissionRouters(ws *restful.WebService) {
	ws.Route(AdminAuthFunc(ws.POST("/v1/permissions")).To(GrantPermission))
	ws.Route(AdminAuthFunc(ws.GET("/v1/permissions")).To(GetPermission))
	ws.Route(AdminAuthFunc(ws.DELETE("/v1/permissions")).To(RevokePermission))

	ws.Route(AdminAuthFunc(ws.GET("/v1/permissions/verify")).To(VerifyPermission))
}

// initTkeRouters init tke api routers
func initTkeRouters(ws *restful.WebService) {
	ws.Route(AdminAuthFunc(ws.POST("/v1/tke/cidr/add_cidr")).To(AddTkeCidr))
	ws.Route(AdminAuthFunc(ws.POST("/v1/tke/cidr/apply_cidr")).To(ApplyTkeCidr))
	ws.Route(AdminAuthFunc(ws.POST("/v1/tke/cidr/release_cidr")).To(ReleaseTkeCidr))
	ws.Route(AdminAuthFunc(ws.POST("/v1/tke/cidr/list_count")).To(ListTkeCidr))

	ws.Route(AdminAuthFunc(ws.POST("/v1/tke/{cluster_id}/sync_credentials")).To(SyncTkeClusterCredentials))
}
