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

// Package pod pod operate
package pod

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/storage"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/storage/entity"
)

// SampleRequset 示例请求
type SampleRequset struct {
	ProjectId string `json:"projectId" in:"path=projectId" validate:"required"`
	ClusterId string `json:"clusterId" in:"path=clusterId" validate:"required"`
}

// SampleResponse 示例响应
type SampleResponse struct {
	Id string `json:"id"`
}

// GetPodContainers 获取 Pod 容器列表
// @Summary 获取 Pod 容器列表
// @Tags    Logs
// @Produce json
// @Success 200 {array} k8sclient.Container
// @Router  /namespaces/:namespace/pods/:pod/containers [get]
func GetPodContainers(c context.Context, req *SampleRequset) (*entity.Audit, error) {
	audit, err := storage.GlobalStorage.GetAudit(c, req.ProjectId, req.ClusterId)
	if err != nil {
		return nil, err
	}
	return audit, nil
}
