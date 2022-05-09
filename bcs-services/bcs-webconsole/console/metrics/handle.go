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

package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// RequestCollect 统计请求耗时
func RequestCollect(handler string) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		code := strconv.FormatInt(int64(c.Writer.Status()), 10)
		requestDuration := getRequestDuration(c, start)
		ReportAPIRequestMetric(handler, c.Request.Method, respStatusTransform(c.Writer.Status()), code, requestDuration)
	}
}

// getRequestDuration 获取请求耗时, 可以是统计整个函数时间，或者在函数内计算好(长链接场景)
func getRequestDuration(c *gin.Context, start time.Time) time.Duration {
	requestDuration := c.Value(HttpRequestDurationKey)
	duration, ok := requestDuration.(time.Duration)
	if ok {
		return duration
	}

	return time.Since(start)
}

func respStatusTransform(status int) string {
	if status == http.StatusOK {
		return SucStatus
	}
	return ErrStatus
}
