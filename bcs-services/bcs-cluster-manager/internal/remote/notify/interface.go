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

// Package notify xxx
package notify

import (
	"context"
)

// MessageNotify message notify interface
type MessageNotify interface {
	Notify(ctx context.Context, message MessageBody) error
}

// NotifyType xxx
type NotifyType string // nolint

var (
	// BkMonitorMetrics bk_monitor_metrics
	BkMonitorMetrics NotifyType = "bk_monitor_metrics"
	// BkMonitorEvent bk_monitor_event
	BkMonitorEvent NotifyType = "bk_monitor_event"
	// Rtx rtx
	Rtx NotifyType = "rtx"
	// Email email
	Email NotifyType = "email"
	// Voice voice
	Voice NotifyType = "voice"
)

// String xxx
func (nt NotifyType) String() string {
	return string(nt)
}

// MessageServer notify server config
type MessageServer struct {
	Server      string
	DataId      int64
	AccessToken string
}

// MessageBody message data
type MessageBody struct {
	// 当通知类型为 rxt; email; voice时, 可通过下面字段通知

	// Users 通知用户，通过逗号隔开
	Users string
	// Content 通知内容
	Content string

	// 当通知类型为 bk_monitor_event 时, 可通过下面字段通知

	// EventName 事件名称
	EventName string
	// EventBody 事件内容
	EventBody string

	// 当通知类型为 bk_monitor_metrics 时, 可通过下面字段通知

	// Metrics metrics指标
	Metrics map[string]int
	// Dimension 维度, 通过不同的场景可固定维度
	Dimension map[string]string
}
