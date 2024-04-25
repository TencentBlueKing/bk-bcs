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

// Package auth xxx
package auth

import (
	"context"
	"encoding/json"
	"errors"
	"regexp"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/jwt"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/cluster"
	middleauth "github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/middleware"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/namespace"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/project"
	authutils "github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/utils"
	jwtGo "github.com/golang-jwt/jwt/v4"
	"go-micro.dev/v4/server"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/utils/contextx"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/utils/stringx"
)

// JWTClientConfig jwt client config
type JWTClientConfig struct {
	Enable         bool
	PublicKey      string
	PublicKeyFile  string
	PrivateKey     string
	PrivateKeyFile string
}

var (
	jwtClient *jwt.JWTClient
	jwtConfig *JWTClientConfig // nolint
)

// NewJWTClient new a jwt client
func NewJWTClient(c JWTClientConfig) (*jwt.JWTClient, error) {
	jwtOpt, err := getJWTOpt(c)
	if err != nil {
		return nil, common.ErrHelmManagerAuthFailed.GenError()
	}
	jwtConfig = &c
	jwtClient, err = jwt.NewJWTClient(*jwtOpt)
	if err != nil {
		return nil, err
	}

	return jwtClient, nil
}

// GetJWTClient get jwt client
func GetJWTClient() *jwt.JWTClient {
	return jwtClient
}

// GetUserFromCtx 通过 ctx 获取当前用户
func GetUserFromCtx(ctx context.Context) string {
	authUser, _ := middleauth.GetUserFromContext(ctx)
	return authUser.GetUsername()
}

// GetRealUserFromCtx 通过 ctx 判断当前用户是否是真实用户
func GetRealUserFromCtx(ctx context.Context) string {
	authUser, _ := middleauth.GetUserFromContext(ctx)
	return authUser.Username
}

func getJWTOpt(c JWTClientConfig) (*jwt.JWTOptions, error) {
	jwtOpt := &jwt.JWTOptions{
		VerifyKeyFile: c.PublicKeyFile,
		SignKeyFile:   c.PrivateKeyFile,
	}
	publicKey := c.PublicKey
	privateKey := c.PrivateKey

	if publicKey != "" {
		key, err := jwtGo.ParseRSAPublicKeyFromPEM([]byte(publicKey))
		if err != nil {
			return nil, err
		}
		jwtOpt.VerifyKey = key
	}
	if privateKey != "" {
		key, err := jwtGo.ParseRSAPrivateKeyFromPEM([]byte(privateKey))
		if err != nil {
			return nil, err
		}
		jwtOpt.SignKey = key
	}
	return jwtOpt, nil
}

// NoAuthMethod 不需要用户身份认证的方法
var NoAuthMethod = []string{
	"HelmManager.Available",
}

// SkipHandler skip handler
func SkipHandler(ctx context.Context, req server.Request) bool {
	// disable auth
	if !options.GlobalOptions.JWT.Enable {
		return true
	}
	return stringx.StringInSlice(req.Method(), NoAuthMethod)
}

// SkipClient skip client
func SkipClient(ctx context.Context, req server.Request, client string) bool {
	resourceID, err := getResourceID(req)
	if err != nil {
		return false
	}

	creds := options.GlobalOptions.Credentials
	for _, v := range creds {
		if !v.Enable {
			continue
		}
		if v.Name != client {
			continue
		}
		if match, _ := regexp.MatchString(v.Scopes.ProjectCode, resourceID.ProjectCode); match &&
			len(v.Scopes.ProjectCode) != 0 {
			return true
		}
		if match, _ := regexp.MatchString(v.Scopes.ClusterID, resourceID.ClusterID); match &&
			len(v.Scopes.ClusterID) != 0 {
			return true
		}
	}
	return false
}

func getResourceID(req server.Request) (*options.CredentialScope, error) {
	body := req.Body()
	b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	resourceID := &options.CredentialScope{}
	err = json.Unmarshal(b, resourceID)
	if err != nil {
		return nil, err
	}
	return resourceID, nil
}

// CheckUserPerm check user perm
func CheckUserPerm(ctx context.Context, req server.Request, username string) (bool, error) {
	blog.Infof("CheckUserPerm: method/%s, username: %s", req.Method(), username)

	action, ok := ActionPermissions[req.Method()]
	if !ok {
		return false, errors.New("operation has not authorized")
	}

	resourceID, err := getResourceID(req)
	if err != nil {
		return false, err
	}

	resourceID.ProjectID = contextx.GetProjectIDFromCtx(ctx)

	allow, url, resources, err := CallIAM(username, action, *resourceID)
	if err != nil {
		return false, err
	}
	if !allow && url != "" {
		return false, &authutils.PermDeniedError{
			Perms: authutils.PermData{
				ApplyURL:   url,
				ActionList: resources,
			},
		}
	}
	return allow, nil
}

// CallIAM call iam
func CallIAM(username, action string, resourceID options.CredentialScope) (bool, string,
	[]authutils.ResourceAction, error) {
	// related actions
	switch action {
	case cluster.CanManageClusterOperation:
		return ClusterIamClient.CanManageCluster(username, resourceID.ProjectID, resourceID.ClusterID)
	case cluster.CanViewClusterOperation:
		return ClusterIamClient.CanViewCluster(username, resourceID.ProjectID, resourceID.ClusterID)
	case project.CanEditProjectOperation:
		return ProjectIamClient.CanEditProject(username, resourceID.ProjectID)
	case project.CanViewProjectOperation:
		return ProjectIamClient.CanViewProject(username, resourceID.ProjectID)
	case namespace.CanCreateNamespaceScopedResourceOperation:
		return NamespaceIamClient.CanCreateNamespaceScopedResource(username, resourceID.ProjectID,
			resourceID.ClusterID, resourceID.Namespace)
	case namespace.CanViewNamespaceScopedResourceOperation:
		return NamespaceIamClient.CanViewNamespaceScopedResource(username, resourceID.ProjectID,
			resourceID.ClusterID, resourceID.Namespace)
	case namespace.CanUpdateNamespaceScopedResourceOperation:
		return NamespaceIamClient.CanUpdateNamespaceScopedResource(username, resourceID.ProjectID,
			resourceID.ClusterID, resourceID.Namespace)
	case namespace.CanDeleteNamespaceScopedResourceOperation:
		return NamespaceIamClient.CanDeleteNamespaceScopedResource(username, resourceID.ProjectID,
			resourceID.ClusterID, resourceID.Namespace)
	default:
		return false, "", nil, errors.New("permission denied")
	}
}
