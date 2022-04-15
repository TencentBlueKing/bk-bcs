/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * 	http://opensource.org/licenses/MIT
 *
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package wrapper

import (
	"context"
	"strings"

	bcsJwt "github.com/Tencent/bk-bcs/bcs-common/pkg/auth/jwt"
	jwtGo "github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	goAttr "github.com/ssrathi/go-attr"
	"go-micro.dev/v4/metadata"
	"go-micro.dev/v4/server"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/cluster"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/ctxkey"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/envs"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/errcode"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/runmode"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/runtime"
	conf "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/config"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/project"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/errorx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/slice"
)

// NewContextInjectWrapper 创建 "向请求的 Context 注入信息" 装饰器
func NewContextInjectWrapper() server.HandlerWrapper {
	return func(fn server.HandlerFunc) server.HandlerFunc {
		return func(ctx context.Context, req server.Request, rsp interface{}) (err error) {
			// 1. 获取或生成 UUID，并作为 requestID 注入到 context
			ctx = context.WithValue(ctx, ctxkey.RequestIDKey, uuid.New().String())

			var username string
			if canExemptAuth(req) {
				username = envs.AnonymousUsername
			} else {
				// 2. 从 GoMicro Metadata（headers）中获取 jwtToken，转换为 username
				md, ok := metadata.FromContext(ctx)
				if !ok {
					return errorx.New(errcode.Unauth, "failed to get micro's metadata")
				}

				username, err = parseUsername(md)
				if err != nil {
					return err
				}
			}
			ctx = context.WithValue(ctx, ctxkey.UsernameKey, username)

			// 3. 注入 Project，Cluster 信息
			if needInjectProjCluster(req) {
				projInfo, clusterInfo, err := fetchProjCluster(req)
				if err != nil {
					return err
				}
				ctx = context.WithValue(ctx, ctxkey.ProjKey, projInfo)
				ctx = context.WithValue(ctx, ctxkey.ClusterKey, clusterInfo)
			}

			// 实际执行业务逻辑，获取返回结果
			return fn(ctx, req, rsp)
		}
	}
}

// NoAuthEndpoints 不需要用户身份认证的方法
var NoAuthEndpoints = []string{
	"Basic.Version",
	"Basic.Ping",
	"Basic.Healthz",
}

// 检查当前请求是否允许免除用户认证
func canExemptAuth(req server.Request) bool {
	// 禁用身份认证
	if conf.G.Auth.Disabled {
		return true
	}
	// 单元测试 / 开发模式
	if runtime.RunMode == runmode.UnitTest || runtime.RunMode == runmode.Dev {
		return true
	}
	// 特殊指定的，不需要认证的方法
	return slice.StringInSlice(req.Endpoint(), NoAuthEndpoints)
}

// 通过 micro metadata（headers）信息，解析出用户名
func parseUsername(md metadata.Metadata) (string, error) {
	jwtToken, ok := md.Get("Authorization")
	if !ok {
		return "", errorx.New(errcode.Unauth, "failed to get authorization token!")
	}
	if len(jwtToken) == 0 || !strings.HasPrefix(jwtToken, "Bearer ") {
		return "", errorx.New(errcode.Unauth, "authorization token error")
	}

	claims, err := jwtDecode(jwtToken[7:])
	if err != nil {
		return "", err
	}
	return claims.UserName, nil
}

// 解析 jwt
func jwtDecode(jwtToken string) (*bcsJwt.UserClaimsInfo, error) {
	if conf.G.Auth.JWTPubKeyObj == nil {
		return nil, errorx.New(errcode.Unauth, "jwt public key uninitialized")
	}

	token, err := jwtGo.ParseWithClaims(
		jwtToken,
		&bcsJwt.UserClaimsInfo{},
		func(token *jwtGo.Token) (interface{}, error) {
			return conf.G.Auth.JWTPubKeyObj, nil
		},
	)
	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errorx.New(errcode.Unauth, "jwt token invalid")
	}

	claims, ok := token.Claims.(*bcsJwt.UserClaimsInfo)
	if !ok {
		return nil, errorx.New(errcode.Unauth, "jwt token's issuer isn't bcs")
	}
	return claims, nil
}

// NoInjectProjClusterEndpoints 不需要注入项目 & 集群信息的方法
var NoInjectProjClusterEndpoints = []string{
	"Basic.Version",
	"Basic.Ping",
	"Basic.Healthz",
	"Basic.Echo",
	"Resource.GetK8SResTemplate",
	// TODO 获取表单数据的也不需要
}

// 需要注入项目 & 集群信息
func needInjectProjCluster(req server.Request) bool {
	return !slice.StringInSlice(req.Endpoint(), NoInjectProjClusterEndpoints)
}

// 获取项目，集群信息
func fetchProjCluster(req server.Request) (*project.Project, *cluster.Cluster, error) {
	projectID, err := goAttr.GetValue(req.Body(), "ProjectID")
	if err != nil {
		return nil, nil, errorx.New(errcode.General, "Get ProjectID from Request Failed: %v", err)
	}
	projInfo, err := project.GetProjectInfo(projectID.(string))
	if err != nil {
		return nil, nil, errorx.New(errcode.General, "获取项目 %s 信息失败：%v", projectID, err)
	}
	clusterID, err := goAttr.GetValue(req.Body(), "ClusterID")
	if err != nil {
		return nil, nil, errorx.New(errcode.General, "Get ClusterID from Request Failed: %v", err)
	}
	clusterInfo, err := cluster.GetClusterInfo(clusterID.(string))
	if err != nil {
		return nil, nil, errorx.New(errcode.General, "获取集群 %s 信息失败：%v", clusterID, err)
	}
	// 若集群类型非共享集群，则需确认集群的项目 ID 与请求参数中的一致
	if !slice.StringInSlice(clusterInfo.Type, cluster.SharedClusterTypes) && clusterInfo.ProjID != projInfo.ID {
		return nil, nil, errorx.New(errcode.ValidateErr, "集群 %s 不属于指定项目!", clusterID)
	}
	return projInfo, clusterInfo, nil
}
