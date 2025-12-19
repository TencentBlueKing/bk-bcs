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

// Package mq defines message queue related interfaces, message structures, and implementations.
package mq

import (
	"encoding/json"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-push-manager/internal/constant"
	thirdpb "github.com/Tencent/bk-bcs/bcs-services/bcs-push-manager/pkg/bcsapi/thirdparty-service"
)

// PushEventMessage defines the structure of a message for a push event.
type PushEventMessage struct {
	// EventID is the unique identifier for the event.
	EventID string `json:"event_id"`
	// Type specifies the notification channels (e.g., rtx, mail, msg).
	Type []string `json:"type"`
	// RTX-specific fields
	RTXReceivers []string `json:"rtx_receivers"`
	RTXContent   string   `json:"rtx_content"`
	RTXTitle     string   `json:"rtx_title"`

	// Mail-specific fields
	MailReceivers []string `json:"mail_receivers"`
	MailContent   string   `json:"mail_content"`
	MailTitle     string   `json:"mail_title"`

	// Msg-specific fields
	MsgReceivers []string `json:"msg_receivers"`
	MsgContent   string   `json:"msg_content"`

	// Extra contains additional data for the notification.
	Extra map[string]string `json:"extra"`
}

// ToRtxRequest converts a PushEventMessage to a thirdparty SendRtxRequest.
func (m *PushEventMessage) ToRtxRequest() *thirdpb.SendRtxRequest {
	if err := validatePushEventMessage(m); err != nil {
		blog.Errorf("cannot convert to SendRtxRequest: %v", err)
		return nil
	}
	rtxReq := &thirdpb.SendRtxRequest{
		Receiver: m.RTXReceivers,
		Title:    m.RTXTitle,
		Message:  m.RTXContent,
		Sender:   constant.NotificationDefaultSender,
	}
	return rtxReq
}

// ToMailRequest converts a PushEventMessage to a thirdparty SendMailRequest.
func (m *PushEventMessage) ToMailRequest() *thirdpb.SendMailRequest {
	if err := validatePushEventMessage(m); err != nil {
		blog.Errorf("cannot convert to SendMailRequest: %v", err)
		return nil
	}
	mailReq := &thirdpb.SendMailRequest{
		Receiver:   m.MailReceivers,
		Title:      m.MailTitle,
		Content:    m.MailContent,
		Sender:     constant.NotificationDefaultSender,
		BodyFormat: constant.NotificationMailFormat,
	}
	return mailReq
}

// ToMsgRequest converts a PushEventMessage to a thirdparty SendMsgRequest.
func (m *PushEventMessage) ToMsgRequest() *thirdpb.SendMsgRequest {
	if err := validatePushEventMessage(m); err != nil {
		blog.Errorf("cannot convert to SendMsgRequest: %v", err)
		return nil
	}
	msgParam := &thirdpb.MsgParam{
		Content: m.MsgContent,
	}
	receiver := &thirdpb.Receiver{
		ReceiverType: constant.MsgReceiverTypeSingle,
		ReceiverIds:  m.MsgReceivers,
	}
	msgReq := &thirdpb.SendMsgRequest{
		Im:       constant.MsgDefaultImWework,
		MsgType:  constant.MsgDefaultTypeMarkdown,
		MsgParam: msgParam,
		Receiver: receiver,
	}
	return msgReq
}

// Marshal serializes the PushEventMessage to JSON.
func (m *PushEventMessage) Marshal() ([]byte, error) {
	return json.Marshal(m)
}

// UnmarshalPushEventMessage deserializes a JSON byte slice into a PushEventMessage.
func UnmarshalPushEventMessage(data []byte) (*PushEventMessage, error) {
	var msg PushEventMessage
	err := json.Unmarshal(data, &msg)
	return &msg, err
}

// validatePushEventMessage validates the PushEventMessage for nil and receivers length.
func validatePushEventMessage(m *PushEventMessage) error {
	if m == nil {
		blog.Errorf("PushEventMessage is nil")
		return fmt.Errorf("PushEventMessage is nil")
	}
	for _, channel := range m.Type {
		switch channel {
		case constant.PushTypeRtx:
			if len(m.RTXReceivers) == 0 {
				blog.Warnf("PushEventMessage has no RTX receivers")
				return fmt.Errorf("PushEventMessage has no RTX receivers")
			}
		case constant.PushTypeMail:
			if len(m.MailReceivers) == 0 {
				blog.Warnf("PushEventMessage has no Mail receivers")
				return fmt.Errorf("PushEventMessage has no Mail receivers")
			}
		case constant.PushTypeMsg:
			if len(m.MsgReceivers) == 0 {
				blog.Warnf("PushEventMessage has no Msg receivers")
				return fmt.Errorf("PushEventMessage has no Msg receivers")
			}
		}
	}
	return nil
}
