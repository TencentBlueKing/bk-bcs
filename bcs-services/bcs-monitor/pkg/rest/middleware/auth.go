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

// Package middleware Authorization
package middleware

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component/bcs"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component/iam"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/rest"
)

// AuthenticationRequired API类型, 兼容多种认证模式
func AuthenticationRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		restContext := rest.InitRestContext(c)

		if c.Request.Method == http.MethodOptions {
			c.Next()
			return
		}

		switch {
		case initContextWithAPIGW(restContext):
		case initContextWithBCSJwt(restContext):
		case initContextWithDevEnv(restContext):
		default:
			rest.AbortWithUnauthorizedError(restContext, rest.ErrorUnauthorized)
			return
		}

		c.Next()
	}
}

// ProjectAuthorization project 鉴权
func ProjectAuthorization() gin.HandlerFunc {
	return func(c *gin.Context) {
		restContext, err := rest.GetRestContext(c)
		if err != nil {
			rest.AbortWithWithForbiddenError(restContext, err)
			return
		}

		projectID := restContext.ProjectId
		clusterID := restContext.ClusterId
		username := restContext.Username

		// check cluster
		cls, err := bcs.GetCluster(clusterID)
		if err != nil {
			rest.AbortWithWithForbiddenError(restContext, err)
			return
		}
		if !cls.IsShared && cls.ProjectID != projectID {
			rest.AbortWithWithForbiddenError(restContext, fmt.Errorf("cluster is invalid"))
			return
		}

		// call iam
		client, err := iam.GetProjectPermClient()
		if err != nil {
			rest.AbortWithWithForbiddenError(restContext, err)
			return
		}
		allow, url, _, err := client.CanViewProject(username, projectID)
		if err != nil {
			rest.AbortWithWithForbiddenError(restContext, err)
			return
		}
		if !allow {
			errMsg := fmt.Errorf("permission denied, please apply permission with %s", url)
			rest.AbortWithWithForbiddenError(restContext, errMsg)
			return
		}

		c.Next()
	}
}

// NsScopeAuthorization 命名空间域资源鉴权
func NsScopeAuthorization() gin.HandlerFunc {
	return func(c *gin.Context) {
		restContext, err := rest.GetRestContext(c)
		if err != nil {
			rest.AbortWithWithForbiddenError(restContext, err)
			return
		}

		projectID := restContext.ProjectId
		clusterID := restContext.ClusterId
		username := restContext.Username

		// check cluster
		cls, err := bcs.GetCluster(clusterID)
		if err != nil {
			rest.AbortWithWithForbiddenError(restContext, err)
			return
		}
		if !cls.IsShared && cls.ProjectID != projectID {
			rest.AbortWithWithForbiddenError(restContext, fmt.Errorf("cluster is invalid"))
			return
		}

		// call iam
		client, err := iam.GetClusterPermClient()
		if err != nil {
			rest.AbortWithWithForbiddenError(restContext, err)
			return
		}
		allow, url, _, err := client.CanViewCluster(username, projectID, clusterID)
		if err != nil {
			rest.AbortWithWithForbiddenError(restContext, err)
			return
		}
		if !allow {
			errMsg := fmt.Errorf("permission denied, please apply permission with %s", url)
			rest.AbortWithWithForbiddenError(restContext, errMsg)
			return
		}

		c.Next()
	}
}

// initContextWithDevEnv Dev环境, 可以设置环境变量
func initContextWithDevEnv(c *rest.Context) bool {
	if config.G.Base.RunEnv != config.DevEnv {
		return false
	}

	// 本地用户认证
	username := os.Getenv("BCS_MONITOR_USERNAME")
	if username != "" {
		c.BindEnv = &rest.EnvToken{Username: username}
		c.Username = username
	}

	// AppCode 认证
	appCode := c.GetHeader("X-BKAPI-JWT-APPCODE")
	if appCode != "" {
		c.BindAPIGW = &rest.APIGWToken{
			App: &rest.APIGWApp{AppCode: appCode, Verified: true},
		}
	}

	if username != "" || appCode != "" {
		return true
	}

	return false
}

// BCSJWTDecode BCS 网关 JWT 解码
func BCSJWTDecode(jwtToken string) (*rest.UserClaimsInfo, error) {
	if config.G.BCS.JWTPubKeyObj == nil {
		return nil, errors.New("jwt public key not set")
	}

	token, err := jwt.ParseWithClaims(jwtToken, &rest.UserClaimsInfo{}, func(token *jwt.Token) (interface{}, error) {
		return config.G.BCS.JWTPubKeyObj, nil
	})
	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("jwt token not valid")
	}

	claims, ok := token.Claims.(*rest.UserClaimsInfo)
	if !ok {
		return nil, errors.New("jwt token not bcs issuer")

	}
	return claims, nil
}

// BKAPIGWJWTDecode 蓝鲸APIGW JWT 解析
func BKAPIGWJWTDecode(jwtToken string) (*rest.APIGWToken, error) {
	if config.G.BKAPIGW.JWTPubKeyObj == nil {
		return nil, errors.New("jwt public key not set")
	}

	token, err := jwt.ParseWithClaims(jwtToken, &rest.APIGWToken{}, func(token *jwt.Token) (interface{}, error) {
		return config.G.BKAPIGW.JWTPubKeyObj, nil
	})
	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("jwt token not valid")
	}

	claims, ok := token.Claims.(*rest.APIGWToken)
	if !ok {
		return nil, errors.New("jwt token not BKAPIGW issuer")

	}
	return claims, nil
}

// initContextWithBCSJwt BCS APISix JWT 鉴权
func initContextWithBCSJwt(c *rest.Context) bool {
	tokenString := c.GetHeader("Authorization")
	if len(tokenString) == 0 || !strings.HasPrefix(tokenString, "Bearer ") {
		return false
	}
	tokenString = tokenString[7:]

	claims, err := BCSJWTDecode(tokenString)
	if err != nil {
		return false
	}

	c.BindBCS = claims
	c.Username = claims.UserName
	return true
}

func initContextWithAPIGW(c *rest.Context) bool {
	// get jwt info from headers
	tokenString := c.GetHeader("X-Bkapi-Jwt")
	if tokenString == "" {
		return false
	}

	token, err := BKAPIGWJWTDecode(tokenString)
	if err != nil {
		return false
	}

	c.BindAPIGW = token

	return true
}

// GetProjectIdOrCode :
func GetProjectIdOrCode(c *gin.Context) string {
	if c.Param("projectId") != "" {
		return c.Param("projectId")
	}
	return ""
}

// GetProjectCode 获取 projectCode 参数
func GetProjectCode(c *gin.Context) string {
	if c.Param("projectCode") != "" {
		return c.Param("projectCode")
	}
	return ""
}

// GetClusterId :
func GetClusterId(c *gin.Context) string {
	if c.Param("clusterId") != "" {
		return c.Param("clusterId")
	}
	return ""
}

// GetSessionId :
func GetSessionId(c *gin.Context) string {
	if c.Param("sessionId") != "" {
		return c.Param("sessionId")
	}
	return ""
}
