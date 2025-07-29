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

// Package action defines the business logic for handling various operations.
package action

import (
	"context"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-push-manager/internal/constant"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-push-manager/internal/mq"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-push-manager/internal/store/mongo"
	third "github.com/Tencent/bk-bcs/bcs-services/bcs-push-manager/internal/thirdparty"
)

// NotificationAction defines the action for handling notification messages.
type NotificationAction struct {
	ThirdpartyClient third.Client
	WhitelistStore   *mongo.ModelPushWhitelist
	EventStore       *mongo.ModelPushEvent
	MaxRetry         int
	RetryInterval    time.Duration
	Chn              *amqp.Channel
}

// HandleMsg handles the consumption of notification messages from RabbitMQ.
func (n *NotificationAction) HandleMsg(messages <-chan amqp.Delivery, done <-chan bool) {
	for {
		select {
		case msg, ok := <-messages:
			if !ok {
				blog.Infof("message channel closed")
				return
			}
			pushMsg, err := mq.UnmarshalPushEventMessage(msg.Body)
			if err != nil {
				blog.Infof("failed to unmarshal message: %v", err)
				continue
			}
			ctx := context.Background()
			event, err := n.EventStore.GetPushEvent(ctx, pushMsg.EventID)
			if err != nil {
				blog.Errorf("failed to query event, eventID: %s, err: %v", pushMsg.EventID, err)
				continue
			}
			if event == nil {
				blog.Warnf("event not found, eventID: %s", pushMsg.EventID)
				continue
			}
			blog.Infof("start handle event: %s, domain: %s, dimension: %+v", event.EventID, event.Domain, event.Dimension)
			allowed, err := n.WhitelistStore.IsDimensionWhitelisted(ctx, event.Domain, event.Dimension)
			if err != nil {
				blog.Infof("whitelist validation exception: %v", err)
				continue
			}
			if allowed {
				err := n.EventStore.UpdatePushEventStatus(ctx, event.EventID, constant.EventStatusWhitelisted)
				if err != nil {
					blog.Errorf("failed to update event status to WHITELISTED, eventID: %s, err: %v", event.EventID, err)
					continue
				}
				blog.Infof("eventID %s is whitelisted, update status to WHITELISTED, skip notification", event.EventID)
				continue
			} else {
				blog.Infof("eventID %s is NOT whitelisted, will send notification", event.EventID)
			}

			allSuccess := true
			for _, pushType := range pushMsg.Type {
				var sendErr error
				for i := 0; i < n.MaxRetry; i++ {
					blog.Infof("try send notification, eventID: %s, type: %s, try: %d", event.EventID, pushType, i+1)
					sendErr = n.sendNotification(pushType, pushMsg)
					if sendErr == nil {
						err := n.EventStore.AppendNotificationResult(ctx, event.EventID, pushType, constant.NotificationResultSuccess)
						if err != nil {
							blog.Errorf("failed to append notification result, eventID: %s, type: %s, err: %v", event.EventID, pushType, err)
						}
						blog.Infof("send notification success, eventID: %s, type: %s", event.EventID, pushType)
						break
					}
					blog.Infof("failed to send %s: %v, retrying %d time(s)", pushType, sendErr, i+1)
					time.Sleep(n.RetryInterval)
				}
				if sendErr != nil {
					allSuccess = false
					err := n.EventStore.AppendNotificationResult(ctx, event.EventID, pushType, constant.NotificationResultFailed)
					if err != nil {
						blog.Errorf("failed to append notification result, eventID: %s, type: %s, err: %v", event.EventID, pushType, err)
					}
					blog.Infof("send notification failed, eventID: %s, type: %s, err: %v", event.EventID, pushType, sendErr)
				}
			}
			if allSuccess {
				err := n.EventStore.UpdatePushEventStatus(ctx, event.EventID, constant.EventStatusSuccess)
				if err != nil {
					blog.Errorf("failed to update event status to SUCCESS, eventID: %s, err: %v", event.EventID, err)
				}
				blog.Infof("all notifications sent successfully, eventID: %s, update status to SUCCESS", event.EventID)
			} else {
				err := n.EventStore.UpdatePushEventStatus(ctx, event.EventID, constant.EventStatusFailed)
				if err != nil {
					blog.Errorf("failed to update event status to FAILED, eventID: %s, err: %v", event.EventID, err)
				}
				blog.Infof("some notifications failed, eventID: %s, update status to FAILED", event.EventID)
			}
		case <-done:
			blog.Infof("received done signal, exiting consumption")
			return
		}
	}
}

// sendNotification sends a notification based on the push type.
func (n *NotificationAction) sendNotification(pushType string, pushMsg *mq.PushEventMessage) error {
	switch pushType {
	case constant.PushTypeRtx:
		if n.ThirdpartyClient != nil {
			req := pushMsg.ToRtxRequest()
			return n.ThirdpartyClient.SendRtx(req)
		}
		return fmt.Errorf("cannot send %s notification: ThirdpartyClient is nil", pushType)
	case constant.PushTypeMail:
		if n.ThirdpartyClient != nil {
			req := pushMsg.ToMailRequest()
			return n.ThirdpartyClient.SendMail(req)
		}
		return fmt.Errorf("cannot send %s notification: ThirdpartyClient is nil", pushType)
	case constant.PushTypeMsg:
		if n.ThirdpartyClient != nil {
			req := pushMsg.ToMsgRequest()
			return n.ThirdpartyClient.SendMsg(req)
		}
		return fmt.Errorf("cannot send %s notification: ThirdpartyClient is nil", pushType)
	default:
		blog.Infof("unknown notification type: %s", pushType)
		return fmt.Errorf("unknown notification type: %s", pushType)
	}
	return nil
}
