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

package route

import (
	"net/http"
	"os"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/types"

	bcsJwt "github.com/Tencent/bk-bcs/bcs-common/pkg/auth/jwt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

var (
	UnauthorizedError = errors.New("用户未登入")
)

// AuthContext :
type AuthContext struct {
	RequestId string `json:"request_id"`
	Operator  string `json:"operator"`
	ProjectId string `json:"project_id"`
	ClusterId string `json:"cluster_id"`
	Username  string `json:"username"`
	// BindAPIGWToken *utils.APIGWToken `json:"bind_jwt"`
}

// WebAuthRequired Web类型, 不需要鉴权
func WebAuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		authCtx := &AuthContext{
			RequestId: uuid.New().String(),
		}
		c.Set("auth", authCtx)

		c.Next()
	}
}

// APIAuthRequired API类型, 兼容多种鉴权模式
func APIAuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodOptions {
			c.Next()
			return
		}
		authCtx := &AuthContext{
			RequestId: uuid.New().String(),
		}

		switch {
		case initContextWithBCSJwt(c, authCtx):
		case initContextWithDevEnv(c, authCtx):
		default:
			c.AbortWithStatusJSON(http.StatusUnauthorized, types.APIResponse{
				Code:      types.ApiErrorCode,
				Message:   UnauthorizedError.Error(),
				RequestID: authCtx.RequestId,
			})
			return
		}

		authCtx.ProjectId = c.Param("projectId")
		authCtx.ClusterId = c.Param("clusterId")

		// 设置鉴权
		c.Set("auth_context", authCtx)

		c.Next()
	}
}

// initContextWithDevEnv Dev环境, 可以设置环境变量
func initContextWithDevEnv(c *gin.Context, authCtx *AuthContext) bool {
	// DEV环境
	if config.G.Base.RunEnv == config.DevEnv {
		username := os.Getenv("WEBCONSOLE_USERNAME")
		if username != "" {
			authCtx.Username = username
			return true
		}
	}
	return false
}

func JWTDecode(jwtToken string) (*bcsJwt.UserClaimsInfo, error) {
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

// initContextWithBCSJwt BCS APISix JWT 鉴权
func initContextWithBCSJwt(c *gin.Context, authCtx *AuthContext) bool {
	tokenString := c.GetHeader("Authorization")
	if len(tokenString) == 0 || !strings.HasPrefix(tokenString, "Bearer ") {
		return false
	}
	tokenString = tokenString[7:]

	claims, err := JWTDecode(tokenString)
	if err != nil {
		return false
	}

	authCtx.Username = claims.UserName
	return true
}

// GetAuthContext 查询鉴权信息
func GetAuthContext(c *gin.Context) (*AuthContext, error) {
	authCtxObj, ok := c.Get("auth_context")
	if !ok {
		return nil, UnauthorizedError
	}

	authCtx, ok := authCtxObj.(*AuthContext)
	if !ok {
		return nil, UnauthorizedError
	}

	return authCtx, nil
}
