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

package i18n

import "github.com/gin-gonic/gin"

// defaultGetLngHandler ...
func defaultGetLngHandler(c *gin.Context, defaultLng string) string {
	if c == nil || c.Request == nil {
		return defaultLng
	}

	// 优先判断cookie
	lng, err := c.Cookie("blueking_language")
	if err == nil {
		return lng
	}

	lng = c.GetHeader("Accept-Language")
	if lng != "" {
		return lng
	}

	lng = c.Query("lang")
	if lng == "" {
		return defaultLng
	}

	return lng
}

// Localize 国际化
func Localize(opts ...Option) gin.HandlerFunc {
	NewI18n(opts...)
	return func(context *gin.Context) {
		atI18n.SetCurrentContext(context)
	}
}
