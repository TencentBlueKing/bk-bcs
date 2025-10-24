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

// Package handle xxx
package handle

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-bk-apisix-gateway/syncing/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-bk-apisix-gateway/syncing/gateway/gateway"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-bk-apisix-gateway/syncing/gateway/resource"
)

// SyncResources xxx
type SyncResources struct {
	syncConfig *config.SyncConfig
}

// NewSyncResources xxx
func NewSyncResources(syncConfig *config.SyncConfig) *SyncResources {
	return &SyncResources{
		syncConfig: syncConfig,
	}
}

// SyncGatewayResources xxx
func (s *SyncResources) SyncGatewayResources(ctx context.Context) error {
	// 同步网关
	gateway := gateway.NewGateway(s.syncConfig)
	err := gateway.GetGateway(ctx)
	if err != nil {
		// 报错，网关不存在
		// 尝试创建网关
		err = gateway.CreateGateway(ctx)
		if err != nil {
			return err
		}
	} else {
		// 网关存在，更新网关
		err = gateway.UpdateGateway(ctx)
		if err != nil {
			return err
		}
	}
	blog.Info("create or update gateway success!")
	// 导入资源
	resource := resource.NewResource(s.syncConfig)
	err = resource.ImportResource(ctx)
	if err != nil {
		return err
	}
	blog.Info("create or update resource success!")

	// 一键发布
	err = gateway.PublishGateway(ctx)
	if err != nil {
		return err
	}
	blog.Info("publish gateway success!")

	return nil
}
