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

package handler

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/auth"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/metrics"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/utils/contextx"
)

// Response interface for protoc response
type Response interface {
	GetCode() uint32
	GetMessage() string
	GetResult() bool
}

func recorder(ctx context.Context, name string, req interface{}, resp Response) func() {
	enterTime := time.Now()

	reqID := contextx.GetRequestIDFromCtx(ctx)
	if reqID == "" {
		blog.Warnf("recorder get grpc metadata from context %s has empty requestID", name)
	}

	blog.Infof("record: receive %s, requestID: %s, username: %s", name, reqID, auth.GetUserFromCtx(ctx))

	return func() {
		var code int
		var message string
		if resp != nil {
			code = int(resp.GetCode())
			message = resp.GetMessage()
		}
		metrics.ReportAPIRequestMetric(name, strconv.Itoa(code), enterTime)
		blog.Infof("record: leave %s, requestID: %s, latency: %s, req: %v, resp: %v", name, reqID,
			time.Since(enterTime), req, fmt.Sprintf("code: %d message: %s", code, message))
	}
}
