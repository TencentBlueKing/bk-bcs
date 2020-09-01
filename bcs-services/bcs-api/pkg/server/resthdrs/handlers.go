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

package resthdrs

import (
	"github.com/Tencent/bk-bcs/bcs-services/bcs-api/pkg/server/resthdrs/filters"
	"github.com/emicklei/go-restful"
)

// AddAuthF appends filters required for user authentication
func AddAuthF(rb *restful.RouteBuilder) *restful.RouteBuilder {
	rb.Filter(filters.AccessTokenAuthenticate).
		Filter(filters.TokenAuthenticate).
		Filter(filters.AuthenticatedRequired)
	return rb
}

// AddClusterF appends filters required for cluster validation to given RouteBuilder object
func AddAuthClusterF(rb *restful.RouteBuilder) *restful.RouteBuilder {
	AddAuthF(rb)
	rb.Filter(filters.ClusterIdVerify)
	return rb
}

func AddSuperUserAuthF(rb *restful.RouteBuilder) *restful.RouteBuilder {
	rb.Filter(filters.SuperTokenAuthenticate)
	return rb
}

func CreateRestContainer(pathPrefix string) *restful.Container {
	// ws
	ws := new(restful.WebService)
	ws.Path(pathPrefix).Consumes(restful.MIME_JSON).Produces(restful.MIME_JSON)

	// NoAuthRequired: agent would only need a registerToken to register the credentials to bke-server
	ws.Route(ws.PUT("/clusters/{cluster_id}/credentials").To(UpdateCredentials))

	// Handlers for BCS
	ws.Route(AddAuthF(ws.POST("/clusters/bcs")).To(CreateBCSCluster))
	ws.Route(AddAuthF(ws.GET("/clusters/bcs/query_by_id/")).To(QueryBCSClusterByID))

	// Handlers for tke cluster
	ws.Route(AddAuthClusterF(ws.POST("/clusters/{cluster_id}/bind_lb")).To(BindLb))
	ws.Route(AddAuthClusterF(ws.GET("/clusters/{cluster_id}/get_lb_status")).To(GetLbStatus))
	ws.Route(AddAuthClusterF(ws.POST("/clusters/{cluster_id}/sync_credentials")).To(SyncTkeClusterCredentials))
	ws.Route(AddSuperUserAuthF(ws.POST("/tke/lb/subnet")).To(UpdateTkeLbSubnet))

	// Handlers for tke cidr adreess management
	ws.Route(AddSuperUserAuthF(ws.POST("/clusters/cidr/add_cidr")).To(AddTkeCidr))
	ws.Route(AddAuthF(ws.POST("/clusters/cidr/apply_cidr")).To(ApplyTkeCidr))
	ws.Route(AddAuthF(ws.POST("/clusters/cidr/release_cidr")).To(ReleaseTkeCidr))
	ws.Route(AddAuthF(ws.GET("/clusters/cidr/list_count")).To(ListTkeCidr))

	// Basic handlers
	ws.Route(AddAuthF(ws.POST("/clusters/")).To(CreatePlainCluster))
	ws.Route(AddAuthClusterF(ws.GET("/clusters/{cluster_id}/client_credentials")).To(GetClientCredentials))
	ws.Route(AddAuthClusterF(ws.GET("/clusters/{cluster_id}/credentials")).To(GetCredentials))
	ws.Route(AddAuthClusterF(ws.GET("/clusters/{cluster_id}/register_tokens")).To(ListRegisterTokens))
	ws.Route(AddAuthClusterF(ws.POST("/clusters/{cluster_id}/register_tokens")).To(CreateRegisterToken))

	// Handlers for bcs-services
	ws.Route(ws.GET("/clusters/bcs/query_by_cluster_id/").To(QueryBCSClusterByClusterID))

	// TODO: Add user management endpoints for admin user
	ws.Route(AddSuperUserAuthF(ws.POST("/users/")).To(CreateUser))
	ws.Route(AddSuperUserAuthF(ws.GET("/users/{user_name}")).To(QueryBCSUserByName))
	ws.Route(AddSuperUserAuthF(ws.POST("/users/{user_id}/tokens")).To(CreateUserToken))
	// ws.Route(ws.POST("/account/tokens").To(ListAccountTokens))

	container := restful.NewContainer()
	container.Add(ws)
	return container
}
