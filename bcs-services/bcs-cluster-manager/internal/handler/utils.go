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

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	grpcmeta "google.golang.org/grpc/metadata"
)

func requestIDFromContext(ctx context.Context) (string, error) {
	meta, ok := grpcmeta.FromIncomingContext(ctx)
	if !ok {
		blog.Warnf("get grpc metadata from context failed")
		return "", fmt.Errorf("get grpc metadata from context failed")
	}
	requestIDStrs := meta.Get("X-Request-Id")
	if len(requestIDStrs) == 0 {
		return "", nil
	}
	return requestIDStrs[0], nil
}
