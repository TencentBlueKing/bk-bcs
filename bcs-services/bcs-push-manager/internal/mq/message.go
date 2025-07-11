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
	// Type specifies the notification channels (e.g., rtx, mail).
	Type []string `json:"type"` // rtx/mail
	// Title is the title of the notification.
	Title string `json:"title"`
	// Content is the body of the notification.
	Content string `json:"content"`
	// Receivers is the list of recipients.
	Receivers []string `json:"receivers"`
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
		Receiver: m.Receivers,
		Title:    m.Title,
		Message:  m.Content,
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
		Receiver:   m.Receivers,
		Title:      m.Title,
		Content:    m.Content,
		Sender:     constant.NotificationDefaultSender,
		BodyFormat: constant.NotificationMailFormat,
	}
	return mailReq
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
	if len(m.Receivers) == 0 {
		blog.Warnf("PushEventMessage has no receivers")
		return fmt.Errorf("PushEventMessage has no receivers")
	}
	return nil
}
