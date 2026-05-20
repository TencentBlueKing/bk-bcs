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

package config

// NoticeConf notice config
type NoticeConf struct {
	Type        string             `json:"type"`        // 通知类型 pushmanager, other
	PushManager PushManagerOptions `json:"pushmanager"` // push manager 配置
}

// PushManagerOptions push manager options
type PushManagerOptions struct {
	Domain        string            // 业务域
	Dimension     map[string]string // 维度信息
	BkBizName     string            // 业务名称 非必填
	Types         []string          // 发送类型列表 rtx, mail, msg
	RTXReceivers  []string          // rtx接收人列表 types包含rtx时必填
	MailReceivers []string          // 邮件接收人列表 types包含mail时必填
	MSGReceivers  []string          // 短信接收人列表 types包含msg时必填
	PushLevel     string            // 推送等级 fatal/warning/reminder 默认warning
	MultiTemplate bool              // 是否多模版模式 默认false
}
