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

// Package server xxx
package server

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/notify"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/notify/monitor"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

// SendMessageToServer send message to server
func SendMessageToServer(ctx context.Context, nt notify.NotifyType,
	server notify.MessageServer, data notify.MessageBody) error {
	taskId := utils.GetTaskIDFromContext(ctx)

	notifyServer := buildNotifyServer(nt, server)
	if notifyServer != nil {
		err := notifyServer.Notify(ctx, data)
		if err != nil {
			blog.Errorf("SendMessageToServer[%s] [%s] failed: %v", taskId, nt, err)
			return err
		}
	}

	blog.Infof("notify SendMessageToServer[%s] successful[%+v]", taskId, data)
	return nil
}

func buildNotifyServer(nType notify.NotifyType, server notify.MessageServer) notify.MessageNotify {
	switch nType {
	case notify.BkMonitorMetrics:
		return monitor.NewMetricsNotify(server)
	case notify.BkMonitorEvent:
		return monitor.NewEventsNotify(server)
	default:
	}

	return nil
}
