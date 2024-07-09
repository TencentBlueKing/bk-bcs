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

// Package wrapper xxx
package wrapper

import (
	"context"
	"encoding/json"
	"strings"

	bcsJwt "github.com/Tencent/bk-bcs/bcs-common/pkg/auth/jwt"
	jwtGo "github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"go-micro.dev/v4/metadata"
	"go-micro.dev/v4/server"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/cluster"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/ctxkey"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/envs"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/errcode"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/runmode"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/runtime"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/component/project"
	conf "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/config"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/i18n"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/errorx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/slice"
)

// NewContextInjectWrapper 创建 "向请求的 Context 注入信息" 装饰器
func NewContextInjectWrapper() server.HandlerWrapper {
	return func(fn server.HandlerFunc) server.HandlerFunc {
		return func(ctx context.Context, req server.Request, rsp interface{}) (err error) {
			md, ok := metadata.FromContext(ctx)
			if !ok {
				return errorx.New(errcode.General, "failed to get micro's metadata")
			}
			username := envs.AnonymousUsername

			// 内部用户，不鉴权
			clientName, ok := md.Get(ctxkey.InnerClientHeaderKey)
			if ok {
				username = clientName
			}
			// 1. 从 Metadata（headers）中获取 jwtToken，转换为 username
			if !canExemptAuth(req) && clientName == "" {
				if username, err = parseUsername(md); err != nil {
					return err
				}
			}
			ctx = context.WithValue(ctx, ctxkey.UsernameKey, username)

			// 2. 注入 Project，Cluster 信息
			if needInjectProjCluster(req) {
				projInfo, clusterInfo, err := fetchProjCluster(ctx, req)
				if err != nil {
					return err
				}
				ctx = context.WithValue(ctx, ctxkey.ProjKey, projInfo)
				ctx = context.WithValue(ctx, ctxkey.ClusterKey, clusterInfo)
			}

			// 3. 解析语言版本信息
			ctx = context.WithValue(ctx, ctxkey.LangKey, i18n.GetLangFromCookies(md))

			// 实际执行业务逻辑，获取返回结果
			return fn(ctx, req, rsp)
		}
	}
}

// getOrCreateReqID 尝试读取 X-Request-Id，若不存在则随机生成
func getOrCreateReqID(md metadata.Metadata) string {
	if reqID, ok := md.Get("x-request-id"); ok {
		return reqID
	}
	return uuid.New().String()
}

// NoAuthEndpoints 不需要用户身份认证的方法
var NoAuthEndpoints = []string{
	"Basic.Version",
	"Basic.Ping",
	"Basic.Healthz",
	// 清理缓存走单独的 Token 认证
	"Resource.InvalidateDiscoveryCache",
}

// canExemptAuth 检查当前请求是否允许免除用户认证
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

// parseUsername 通过 micro metadata（headers）信息，解析出用户名
func parseUsername(md metadata.Metadata) (string, error) {
	authorization, ok := md.Get("Authorization")
	if !ok {
		return "", errorx.New(errcode.Unauth, "failed to get authorization token!")
	}
	if len(authorization) == 0 || !strings.HasPrefix(authorization, "Bearer ") {
		return "", errorx.New(errcode.Unauth, "authorization token error")
	}

	u, err := jwtDecode(authorization[7:])
	if err != nil {
		return "", err
	}
	if u.SubType == bcsJwt.User.String() {
		if u.UserName == "" {
			return u.ClientID, nil
		}
		return u.UserName, nil
	}
	if u.SubType == bcsJwt.Client.String() {
		if username, ok := md.Get(ctxkey.CustomUsernameHeaderKey); ok {
			return username, nil
		}
		return "", errorx.New(errcode.Unauth, "username is empty")
	}
	return u.UserName, nil
}

// jwtDecode 解析 jwt
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
	// 订阅 API 比较特殊，单独走 Info 注入逻辑
	"Resource.Subscribe",
	// Example & Tmpl API 不需要 Info 注入
	"Resource.GetK8SResTemplate",
	"Resource.GetFormSupportedAPIVersions",
	// 清理缓存无需获取 Info 信息
	"Resource.InvalidateDiscoveryCache",
}

// needInjectProjCluster 需要注入项目 & 集群信息
func needInjectProjCluster(req server.Request) bool {
	return !slice.StringInSlice(req.Endpoint(), NoInjectProjClusterEndpoints)
}

// fetchProjCluster 获取项目，集群信息
func fetchProjCluster(ctx context.Context, req server.Request) (*project.Project, *cluster.Cluster, error) {
	resourceID, err := getResourceID(req)
	if err != nil {
		return nil, nil, errorx.New(errcode.General, "Parse params failed: %v", err)
	}
	projInfo, err := project.GetProjectInfo(ctx, resourceID.ProjectID)
	if err != nil {
		return nil, nil, errorx.New(errcode.General, i18n.GetMsg(ctx, "获取项目 %s 信息失败：%v"),
			resourceID.ProjectID, err)
	}
	// 有些接口没有集群 ID 参数
	if resourceID.ClusterID == "" {
		return projInfo, nil, nil
	}
	clusterInfo, err := cluster.GetClusterInfo(ctx, resourceID.ClusterID)
	if err != nil {
		return nil, nil, errorx.New(errcode.General, i18n.GetMsg(ctx, "获取集群 %s 信息失败：%v"),
			resourceID.ClusterID, err)
	}
	// 若集群类型非共享集群，则需确认集群的项目 ID 与请求参数中的一致
	if !slice.StringInSlice(clusterInfo.Type, cluster.SharedClusterTypes) && clusterInfo.ProjID != projInfo.ID {
		return nil, nil, errorx.New(errcode.ValidateErr, i18n.GetMsg(ctx, "集群 %s 不属于指定项目!"),
			resourceID.ClusterID)
	}
	return projInfo, clusterInfo, nil
}

type resource struct {
	ProjectCode string `json:"projectCode" yaml:"projectCode"`
	ClusterID   string `json:"clusterID" yaml:"clusterID"`
	ProjectID   string `json:"projectID" yaml:"projectID"`
}

func getResourceID(req server.Request) (*resource, error) {
	body := req.Body()
	b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	resourceID := &resource{}
	err = json.Unmarshal(b, resourceID)
	if err != nil {
		return nil, err
	}
	if resourceID.ProjectID == "" {
		resourceID.ProjectID = resourceID.ProjectCode
	}
	return resourceID, nil
}

// GetUserAgentFromCtx 通过 ctx 获取 userAgent
func GetUserAgentFromCtx(ctx context.Context) string {
	md, ok := metadata.FromContext(ctx)
	if !ok {
		return ""
	}
	userAgent, _ := md.Get(ctxkey.UserAgentHeaderKey)
	return userAgent
}

// GetSourceIPFromCtx 通过 ctx 获取 sourceIP
func GetSourceIPFromCtx(ctx context.Context) string {
	md, ok := metadata.FromContext(ctx)
	if !ok {
		return ""
	}
	forwarded, _ := md.Get(ctxkey.ForwardedForHeaderKey)
	return forwarded
}

// GetProjectCodeFromCtx 通过 ctx 获取 projectCode
func GetProjectCodeFromCtx(ctx context.Context) string {
	p := ctx.Value(ctxkey.ProjKey)
	if p == nil {
		return ""
	}
	return p.(*project.Project).Code
}

// GetUserFromCtx 通过 ctx 获取当前用户
func GetUserFromCtx(ctx context.Context) string {
	username, _ := ctx.Value(ctxkey.UsernameKey).(string)
	return username
}
