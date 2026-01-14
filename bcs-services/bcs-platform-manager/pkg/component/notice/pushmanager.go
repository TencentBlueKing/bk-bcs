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
	"fmt"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/pushmanager"
	pushmgr "github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/component/bcs/pushmanager"
)

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

// PushManager push manager
type PushManager struct {
	Options PushManagerOptions
}

// NewPushManager new push manager
func NewPushManager(o PushManagerOptions) Notice {
	return &PushManager{
		Options: o,
	}
}

// SendSMS send sms
func (p *PushManager) SendSMS(ctx context.Context, templateID string, params map[string]string) error {
	return nil
}

// SendEmail send email
func (p *PushManager) SendEmail(ctx context.Context, templateID string, params map[string]string) error {
	return nil
}

// SendRTX send rtx
func (p *PushManager) SendRTX(ctx context.Context, templateID string, params map[string]string) error {
	return nil
}

// Send send notice
func (p *PushManager) Send(ctx context.Context, templateID string, params map[string]string) error {
	if templateID == "" {
		return fmt.Errorf("templateID is empty")
	}

	req := &pushmanager.PushEvent{
		Domain: p.Options.Domain,
		EventDetail: &pushmanager.EventDetail{
			Fields: map[string]string{
				"types":          strings.Join(p.Options.Types, ","),
				"mail_receivers": strings.Join(p.Options.MailReceivers, ","),
				"msg_receivers":  strings.Join(p.Options.MSGReceivers, ","),
				"rtx_receivers":  strings.Join(p.Options.RTXReceivers, ","),
			},
		},
		PushLevel: p.Options.PushLevel,
		BkBizName: p.Options.BkBizName,
		Dimension: &pushmanager.Dimension{
			Fields: p.Options.Dimension,
		},
	}

	if p.Options.MultiTemplate {
		templates, err := pushmgr.ListPushTemplateByID(ctx, p.Options.Domain, templateID)
		if err != nil {
			return err
		}

		for _, template := range templates {
			for _, t := range p.Options.Types {
				if template.TemplateId == fmt.Sprintf("%s_%s", templateID, t) {
					if t == "msg" {
						req.EventDetail.Fields[fmt.Sprintf("%s_content", t)] = template.Content.GetBody()
					} else {
						req.EventDetail.Fields[fmt.Sprintf("%s_title", t)] = template.Content.GetTitle()
						req.EventDetail.Fields[fmt.Sprintf("%s_content", t)] = template.Content.GetBody()
					}
				}
			}
		}
	} else {
		templateContent, err := pushmgr.GetPushTemplate(ctx, p.Options.Domain, templateID)
		if err != nil {
			return err
		}

		for _, t := range p.Options.Types {
			if t == "msg" {
				req.EventDetail.Fields[fmt.Sprintf("%s_content", t)] = templateContent.GetBody()
			} else {
				req.EventDetail.Fields[fmt.Sprintf("%s_title", t)] = templateContent.GetTitle()
				req.EventDetail.Fields[fmt.Sprintf("%s_content", t)] = templateContent.GetBody()
			}
		}
	}

	eventID, err := pushmgr.PushEvents(ctx, req)
	if err != nil {
		return err
	}

	fmt.Printf("push event success, eventID: %s\n", eventID)
	return nil
}
