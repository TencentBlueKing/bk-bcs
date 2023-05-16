/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * 	http://opensource.org/licenses/MIT
 *
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package handler

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/proto/bcsproject"
)

// HealthzHandler xxx
type HealthzHandler struct{}

// NewHealthz create a healthz hander
func NewHealthz() *HealthzHandler {
	return &HealthzHandler{}
}

// Ping 用于liveness
func (h *HealthzHandler) Ping(ctx context.Context, req *proto.PingRequest, resp *proto.PingResponse) error {
	resp.Data = "pong"
	return nil
}

// Healthz 用于readiness
func (h *HealthzHandler) Healthz(ctx context.Context, req *proto.HealthzRequest, resp *proto.HealthzResponse) error {
	mongoDB := store.GetMongo()
	// 默认状态为正常
	health := "service is ok!"
	if err := mongoDB.Ping(); err != nil {
		health = "service is unhealthy, mongo ping error"
	}

	// 现阶段仅依赖mongo，因此，返回一样
	retData := &proto.HealthzData{
		Status:      health,
		MongoStatus: health,
	}
	resp.Data = retData
	return nil
}
