/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/jwt"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/cluster"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/project"
	jwtGo "github.com/dgrijalva/jwt-go"
	"github.com/micro/go-micro/v2/server"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/common"
	projectClient "github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/component/project"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/utils/stringx"
	middleauth "github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/pkg/middleware/auth"
)

// JWTClientConfig jwt client config
type JWTClientConfig struct {
	Enable         bool
	PublicKey      string
	PublicKeyFile  string
	PrivateKey     string
	PrivateKeyFile string
	ExemptClients  string
}

var (
	jwtClient *jwt.JWTClient
	jwtConfig *JWTClientConfig
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
func SkipHandler(req server.Request) bool {
	// disable auth
	if !jwtConfig.Enable {
		return true
	}
	return stringx.StringInSlice(req.Method(), NoAuthMethod)
}

// SkipClient skip client
func SkipClient(req server.Request, client string) bool {
	clientIDs := stringx.SplitString(jwtConfig.ExemptClients)
	return stringx.StringInSlice(client, clientIDs)
}

// resourceID resource id
type resourceID struct {
	ProjectCode string `json:"ProjectCode,omitempty"`
	ProjectID   string `json:"ProjectID,omitempty"`
	ClusterID   string `json:"clusterID,omitempty"`
}

// CheckUserPerm check user perm
func CheckUserPerm(req server.Request, username string) (bool, error) {
	blog.Infof("CheckUserPerm: method/%s, username: %s", req.Method(), username)

	body := req.Body()
	b, err := json.Marshal(body)
	if err != nil {
		return false, err
	}

	resourceID := &resourceID{}
	err = json.Unmarshal(b, resourceID)
	if err != nil {
		return false, err
	}

	action, ok := ActionPermissions[req.Method()]
	if !ok {
		return false, errors.New("operation has not authorized")
	}

	if len(resourceID.ProjectCode) > 0 && len(resourceID.ProjectID) == 0 {
		projectCode := resourceID.ProjectCode
		projectID, err := projectClient.GetProjectIDByCode(username, projectCode)
		if err != nil {
			err := fmt.Errorf("CheckUserPerm get project id error, projectCode: %s, err: %s", projectCode, err.Error())
			blog.Errorf("%s", err)
			return false, err
		}
		resourceID.ProjectID = projectID
	}

	allow, _, err := callIAM(username, action, *resourceID)
	if err != nil {
		return false, err
	}
	return allow, nil
}

func callIAM(username, action string, resourceID resourceID) (bool, string, error) {
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
	default:
		return false, "", errors.New("permission denied")
	}
}
