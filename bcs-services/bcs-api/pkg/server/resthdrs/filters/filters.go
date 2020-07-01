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

package filters

import (
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-api/pkg/auth"
	m "github.com/Tencent/bk-bcs/bcs-services/bcs-api/pkg/models"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-api/pkg/server/resthdrs/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-api/pkg/server/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-api/pkg/storages/sqlstore"
	"github.com/emicklei/go-restful"
)

const (
	CurrentUserAttr      = "bke_current_user"
	CurrentCluster       = "bke_current_cluster"
	CurrentUserTokenType = "bke_current_usertoken_type"
)

// ====================== //
// Authentication filters //
// ====================== //

// AuthenticatedRequired aborts current request if it isn't authenticated, which means there is no "CurrentUserAttr"
// attribute can be found in request.
func AuthenticatedRequired(request *restful.Request, response *restful.Response, chain *restful.FilterChain) {
	// If there is already "currentUser" attribute in request object, skip this authenticater
	if request.Attribute(CurrentUserAttr) == nil {
		message := fmt.Sprintf("errcode：%d,  anonymous requests is forbidden, please provide a valid token", common.BcsErrApiUnauthorized)
		utils.WriteUnauthorizedError(response, "UNAUTHORIZED", message)
		return
	}

	chain.ProcessFilter(request, response)
}

// tokenAuthenticate authenticates current user by bearer token
func TokenAuthenticate(request *restful.Request, response *restful.Response, chain *restful.FilterChain) {
	// If there is already "currentUser" attribute in request object, skip this authenticater
	if request.Attribute(CurrentUserAttr) != nil {
		chain.ProcessFilter(request, response)
		return
	}

	authenticater := auth.NewTokenAuthenticater(request.Request, &auth.TokenAuthConfig{
		SourceBearerEnabled: true,
	})
	user, hasExpired := authenticater.GetUser()
	if user != nil && !hasExpired {
		user.BackendType = types.UserBackendTypeDefault
		request.SetAttribute(CurrentUserAttr, user)
		userTokenType := authenticater.GetUserTokenType()
		request.SetAttribute(CurrentUserTokenType, int(userTokenType))
	}
	// Set current user to request for later procedure
	chain.ProcessFilter(request, response)
}

// SuperTokenAuthenticate authenticates current user whether is super user by bearer token
func SuperTokenAuthenticate(request *restful.Request, response *restful.Response, chain *restful.FilterChain) {

	authenticater := auth.NewTokenAuthenticater(request.Request, auth.DefaultTokenAuthConfig)
	user, hasExpired := authenticater.GetUser()
	if user != nil && !hasExpired && user.IsSuperUser {
		chain.ProcessFilter(request, response)
		return
	}

	message := fmt.Sprintf("errcode：%d,  anonymous requests is forbidden, please provide a valid token", common.BcsErrApiUnauthorized)
	utils.WriteUnauthorizedError(response, "UNAUTHORIZED", message)
	return

}

// accessTokenAuthenticate authenticates the user using access_token parameter in query string.
func AccessTokenAuthenticate(request *restful.Request, response *restful.Response, chain *restful.FilterChain) {
	// If there is already "currentUser" attribute in request object, skip this authenticater
	if request.Attribute(CurrentUserAttr) != nil {
		chain.ProcessFilter(request, response)
		return
	}

	accessToken := request.QueryParameter("access_token")
	blog.Info("access_token: %s", accessToken)
	if accessToken != "" {
		user, err := auth.VerifyAccessTokenAndCreateUser(accessToken)
		if err != nil {
			blog.Warnf("Failed to get user from access_token(%s): %s", accessToken, err.Error())
		}
		if user != nil {
			user.BackendType = types.UserBackendTypeBCSAuth
			user.BackendCredentials = m.BackendCredentials{
				"access_token": accessToken,
			}
			// Set current user to request for later procedure
			request.SetAttribute(CurrentUserAttr, user)
			request.SetAttribute(CurrentUserTokenType, m.UserTokenTypeKubeConfigForPaas)
		}
	}

	chain.ProcessFilter(request, response)
}

// =============== //
// Cluster Filters //
// =============== //

// ClusterVerify verifies if given cluster_id is valid.
func ClusterIdVerify(request *restful.Request, response *restful.Response, chain *restful.FilterChain) {
	clusterId := request.PathParameter("cluster_id")
	cluster := sqlstore.GetCluster(clusterId)
	if cluster == nil {
		message := fmt.Sprintf("errcode: %d, cluster %s not found", common.BcsErrApiK8sClusterNotFound, clusterId)
		utils.WriteNotFoundError(response, "CLUSTER_NOT_FOUND", message)
		return
	}

	request.SetAttribute(CurrentCluster, cluster)
	chain.ProcessFilter(request, response)

}

// Get CurrentUser from request object
func GetUser(req *restful.Request) *m.User {
	user := req.Attribute(CurrentUserAttr)
	ret, ok := user.(*m.User)
	if ok {
		return ret
	}

	return nil
}

func GetCluster(req *restful.Request) *m.Cluster {
	cluster := req.Attribute(CurrentCluster)
	ret, ok := cluster.(*m.Cluster)
	if ok {
		return ret
	}

	return nil

}

func GetUserTokenType(req *restful.Request) int {
	userTokenType := req.Attribute(CurrentUserTokenType)
	return userTokenType.(int)

}
