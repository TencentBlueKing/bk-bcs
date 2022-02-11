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
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

func AuthRequired() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Next()
	}
}

func AuthWithJWTRequired() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if ctx.Request.Method == "OPTIONS" {
			ctx.Next()
			return
		}
		tokenString := ctx.GetHeader("Authorization")
		if len(tokenString) == 0 || !strings.HasPrefix(tokenString, "Bearer ") {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, "")
			return
		}
		tokenString = tokenString[7:]
		claims := jwt.StandardClaims{}
		TokenSecret := ""
		token, err := jwt.ParseWithClaims(tokenString, &claims, func(t *jwt.Token) (interface{}, error) {
			return []byte(TokenSecret), nil
		})
		if err != nil {
			ctx.AbortWithError(http.StatusUnauthorized, err)
		}
		if !token.Valid {
			ctx.AbortWithStatus(http.StatusUnauthorized)
		}
		ctx.Set("username", claims.Subject)
		ctx.Next()
	}
}

func CorsHandler(allowOrigin string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Header("Access-Control-Allow-Origin", allowOrigin)
		ctx.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, token")
		ctx.Header("Access-Control-Allow-Methods", "POST, GET, DELETE, PUT, OPTIONS")
		ctx.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")
		ctx.Header("Access-Control-Allow-Credentials", "true")
		if ctx.Request.Method == "OPTIONS" {
			ctx.AbortWithStatus(http.StatusNoContent)
		}
		ctx.Next()
	}
}
