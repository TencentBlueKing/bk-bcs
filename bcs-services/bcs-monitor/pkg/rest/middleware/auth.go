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

package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/rest"

	bcsJwt "github.com/Tencent/bk-bcs/bcs-common/pkg/auth/jwt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

// APIAuthRequired API类型, 兼容多种鉴权模式
func AuthRequired() gin.HandlerFunc {
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
			rest.AbortWithUnauthorizedError(restContext, rest.UnauthorizedError)
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

// BCSJWTDecode BCS JWT 解析
func BCSJWTDecode(jwtToken string) (*bcsJwt.UserClaimsInfo, error) {
	if config.G.BCS.JWTPubKeyObj == nil {
		return nil, errors.New("jwt public key not set")
	}

	token, err := jwt.ParseWithClaims(jwtToken, &bcsJwt.UserClaimsInfo{}, func(token *jwt.Token) (interface{}, error) {
		return config.G.BCS.JWTPubKeyObj, nil
	})
	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("jwt token not valid")
	}

	claims, ok := token.Claims.(*bcsJwt.UserClaimsInfo)
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

// GetProjectIdOrCode
func GetProjectIdOrCode(c *gin.Context) string {
	if c.Param("projectId") != "" {
		return c.Param("projectId")
	}
	return ""
}

// GetClusterId
func GetClusterId(c *gin.Context) string {
	if c.Param("clusterId") != "" {
		return c.Param("clusterId")
	}
	return ""
}

// GetSessionId
func GetSessionId(c *gin.Context) string {
	if c.Param("sessionId") != "" {
		return c.Param("sessionId")
	}
	return ""
}
