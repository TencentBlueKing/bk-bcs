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

package route

import (
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/components"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/components/bcs"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/i18n"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/sessions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/types"
)

var (
	// UnauthorizedError xxx
	UnauthorizedError = errors.New("用户未登入") // nolint
)

// RequestIdGenerator xxx
func RequestIdGenerator(r *http.Request) string {
	if r.Header.Get("X-Request-ID") != "" {
		return r.Header.Get("X-Request-ID")
	}

	uid := uuid.New().String()
	requestId := strings.ReplaceAll(uid, "-", "")
	return requestId
}

// AuthContext :
type AuthContext struct {
	RequestId   string            `json:"request_id"`
	StartTime   time.Time         `json:"start_time"`
	Operator    string            `json:"operator"`
	Username    string            `json:"username"`
	ProjectId   string            `json:"project_id"`
	ProjectCode string            `json:"project_code"`
	ClusterId   string            `json:"cluster_id"`
	BindEnv     *EnvToken         `json:"bind_env"`
	BindBCS     *UserClaimsInfo   `json:"bind_bcs"`
	BindCluster *bcs.Cluster      `json:"bind_cluster"`
	BindProject *bcs.Project      `json:"bind_project"`
	BindSession *types.PodContext `json:"bind_session"`
}

// BKAppCode 返回验证的 AppCode, 兼容bcs网关和蓝鲸网关
func (c *AuthContext) BKAppCode() string {
	// BCS网关
	if c.BindBCS != nil && c.BindBCS.BKAppCode != "" {
		return c.BindBCS.BKAppCode
	}

	return ""
}

// WebAuthRequired Web类型, 不需要鉴权
func WebAuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		authCtx := &AuthContext{
			RequestId: RequestIdGenerator(c.Request),
			StartTime: time.Now(),
		}
		c.Set("auth_context", authCtx)

		// 新增泳道特性
		ctx := types.WithLaneIdCtx(c.Request.Context(), c.Request.Header)
		c.Request = c.Request.WithContext(components.WithRequestIDValue(ctx, authCtx.RequestId))

		c.Next()
	}
}

// APIAuthRequired API类型, 兼容多种鉴权模式
func APIAuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		authCtx := &AuthContext{
			RequestId: RequestIdGenerator(c.Request),
			StartTime: time.Now(),
		}
		c.Set("auth_context", authCtx)

		// 新增泳道特性
		ctx := types.WithLaneIdCtx(c.Request.Context(), c.Request.Header)
		c.Request = c.Request.WithContext(components.WithRequestIDValue(ctx, authCtx.RequestId))

		if c.Request.Method == http.MethodOptions {
			c.Next()
			return
		}

		// websocket 协议单独鉴权
		if c.IsWebsocket() {
			c.Next()
			return
		}

		switch {
		case initContextWithPortalSession(c, authCtx):
		case initContextWithBCSJwt(c, authCtx):
		case initContextWithDevEnv(c, authCtx):
		default:
			// 统一翻译格式解决冲突问题
			err := UnauthorizedError
			c.AbortWithStatusJSON(http.StatusUnauthorized, types.APIResponse{
				Code:      types.ApiErrorCode,
				Message:   i18n.T(c, "%s", err),
				RequestID: authCtx.RequestId,
			})
			return
		}

		c.Next()
	}
}

// EnvToken xxx
type EnvToken struct {
	Username string
}

// initContextWithDevEnv Dev环境, 可以设置环境变量
func initContextWithDevEnv(c *gin.Context, authCtx *AuthContext) bool {
	if config.G.Base.RunEnv != config.DevEnv {
		return false
	}

	// 本地用户认证
	username := os.Getenv("WEBCONSOLE_USERNAME")
	if username != "" {
		authCtx.BindEnv = &EnvToken{Username: username}
		authCtx.Username = username
	}

	// AppCode 认证
	appCode := c.GetHeader("X-BKAPI-JWT-APPCODE")
	if appCode != "" {
		authCtx.BindBCS = &UserClaimsInfo{
			BKAppCode: appCode,
		}
	}

	if username != "" || appCode != "" {
		return true
	}

	return false
}

// BCSJWTDecode BCS 网关 JWT 解码
func BCSJWTDecode(jwtToken string) (*UserClaimsInfo, error) {
	if config.G.BCS.JWTPubKeyObj == nil {
		return nil, errors.New("jwt public key not set")
	}

	token, err := jwt.ParseWithClaims(jwtToken, &UserClaimsInfo{}, func(token *jwt.Token) (interface{}, error) {
		return config.G.BCS.JWTPubKeyObj, nil
	})
	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("jwt token not valid")
	}

	claims, ok := token.Claims.(*UserClaimsInfo)
	if !ok {
		return nil, errors.New("jwt token not bcs issuer")

	}
	return claims, nil
}

// UserClaimsInfo custom jwt claims
type UserClaimsInfo struct {
	SubType      string `json:"sub_type"`
	UserName     string `json:"username"`
	BKAppCode    string `json:"bk_app_code"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	// https://tools.ietf.org/html/rfc7519#section-4.1
	// aud: 接收jwt一方; exp: jwt过期时间; jti: jwt唯一身份认证; IssuedAt: 签发时间; Issuer: jwt签发者
	*jwt.StandardClaims
}

// initContextWithBCSJwt BCS APISix JWT 鉴权
func initContextWithBCSJwt(c *gin.Context, authCtx *AuthContext) bool {
	tokenString := c.GetHeader("Authorization")
	if len(tokenString) == 0 || !strings.HasPrefix(tokenString, "Bearer ") {
		return false
	}
	tokenString = tokenString[7:]

	claims, err := BCSJWTDecode(tokenString)
	if err != nil {
		return false
	}

	authCtx.BindBCS = claims
	authCtx.Username = claims.UserName
	return true
}

func initContextWithPortalSession(c *gin.Context, authCtx *AuthContext) bool {
	// get jwt info from headers
	sessionId := GetSessionId(c)
	if sessionId == "" {
		return false
	}

	podCtx, err := sessions.NewStore().OpenAPIScope().Get(c.Request.Context(), sessionId)
	if err != nil {
		return false
	}

	authCtx.BindSession = podCtx

	return true
}

// MustGetAuthContext 查询鉴权信息
func MustGetAuthContext(c *gin.Context) *AuthContext {
	authCtxObj := c.MustGet("auth_context")

	authCtx, ok := authCtxObj.(*AuthContext)
	if !ok {
		panic("not valid auth_context")
	}

	return authCtx
}

// GetProjectIdOrCode xxx
func GetProjectIdOrCode(c *gin.Context) string {
	if c.Param("projectId") != "" {
		return c.Param("projectId")
	}
	return ""
}

// GetClusterId xxx
func GetClusterId(c *gin.Context) string {
	if c.Param("clusterId") != "" {
		return c.Param("clusterId")
	}
	return ""
}

// GetNamespace xxx
func GetNamespace(c *gin.Context) string {
	if c.Param("namespace") != "" {
		return c.Param("namespace")
	}
	if c.Query("namespace") != "" {
		return c.Query("namespace")
	}
	return ""
}

// GetSessionId xxx
func GetSessionId(c *gin.Context) string {
	if c.Param("sessionId") != "" {
		return c.Param("sessionId")
	}
	if c.Query("session_id") != "" {
		return c.Query("session_id")
	}
	return ""
}
