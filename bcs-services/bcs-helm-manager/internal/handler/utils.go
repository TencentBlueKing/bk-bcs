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
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/metrics"

	grpcmeta "google.golang.org/grpc/metadata"
)

// Response interface for protoc response
type Response interface {
	GetCode() uint32
	GetMessage() string
	GetResult() bool
}

func recorder(ctx context.Context, name string, req interface{}, resp Response) func() {
	enterTime := time.Now()

	reqID := requestID(ctx)
	if reqID == "" {
		blog.Warnf("recorder get grpc metadata from context %s has empty requestID", name)
	}

	blog.Infof("record: receive %s, reqID: %s", name, reqID)

	return func() {
		metrics.ReportAPIRequestMetric(name, strconv.Itoa(int(resp.GetCode())), enterTime)
		blog.Infof("record: leave %s, reqID: %s, req: %v, resp: %v", name, reqID, req, func() string {

			// ignore resp data when list actions
			if strings.HasPrefix(name, "List") {
				return "[ignore resp logs for List action]"
			}

			return fmt.Sprintf("%v", resp)
		}())
	}
}

func requestID(ctx context.Context) string {
	meta, ok := grpcmeta.FromIncomingContext(ctx)
	if !ok {
		return ""
	}

	if sl := meta.Get("X-Request-Id"); len(sl) > 0 {
		return sl[0]
	}

	return ""
}
