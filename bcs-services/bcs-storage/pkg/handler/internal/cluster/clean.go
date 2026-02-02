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

// Package cluster xxx
package cluster

import (
	"context"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	storage "github.com/Tencent/bk-bcs/bcs-services/bcs-storage/pkg/proto"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/clean"
)

// CleanHandler 集群清理处理器
type CleanHandler struct {
	cleaner *clean.ClusterCleaner
}

// NewCleanHandler 创建集群清理处理器
func NewCleanHandler() *CleanHandler {
	return &CleanHandler{
		cleaner: clean.NewClusterCleaner(),
	}
}

// CleanClusterData gRPC接口：清理集群数据
func (h *CleanHandler) CleanClusterData(ctx context.Context, req *storage.CleanClusterDataRequest) (
	*storage.CleanClusterDataResponse, error) {
	blog.Infof("receive CleanClusterData request: clusterId=%s", req.ClusterId)

	// 参数验证
	if req.ClusterId == "" {
		blog.Errorf("validate CleanClusterData request failed: clusterId is empty")
		return &storage.CleanClusterDataResponse{
			Code:    400,
			Message: "invalid request: clusterId is empty",
		}, nil
	}

	// 执行清理
	deletedCounts, err := h.cleaner.CleanClusterData(ctx, req.ClusterId)

	// 构造响应
	if err != nil {
		blog.Errorf("clean cluster %s data failed: %v", req.ClusterId, err)
		return &storage.CleanClusterDataResponse{
			Code:    500,
			Message: fmt.Sprintf("clean failed: %v", err),
		}, nil
	}

	blog.Infof("clean cluster %s data success: deleted=%v", req.ClusterId, deletedCounts)

	return &storage.CleanClusterDataResponse{
		Code:          0,
		Message:       "success",
		DeletedCounts: deletedCounts,
	}, nil
}
