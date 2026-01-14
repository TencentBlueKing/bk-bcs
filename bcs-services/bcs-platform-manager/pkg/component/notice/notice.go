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

// Package notice xxx
package notice

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/config"
)

// TypePushManager pushmanager type
const TypePushManager = "pushmanager"

var Obj Notice

// demo
// err := notice.Obj.Send(ctx, "TEMPLATE_ID", nil)
// if err != nil {
// 	return err
// }

// Notice notice interface
type Notice interface {
	SendSMS(ctx context.Context, templateID string, params map[string]string) error
	SendEmail(ctx context.Context, templateID string, params map[string]string) error
	SendRTX(ctx context.Context, templateID string, params map[string]string) error
	Send(ctx context.Context, templateID string, params map[string]string) error
}

// InitNotice init notice
func InitNotice() {
	noticeCnf := config.G.Notice
	if noticeCnf.Type == "" || noticeCnf.Type == TypePushManager {
		Obj = NewPushManager(PushManagerOptions{
			Domain:    noticeCnf.PushManager.Domain,
			Dimension: noticeCnf.PushManager.Dimension,
			BkBizName: noticeCnf.PushManager.BkBizName,
			Types:     noticeCnf.PushManager.Types,
		})
	}
}
