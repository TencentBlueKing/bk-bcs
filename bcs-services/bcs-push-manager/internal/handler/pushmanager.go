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

// Package handler provides the gRPC handlers for the PushManager service.
package handler

import (
	"context"
	"fmt"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-push-manager/internal/action"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-push-manager/internal/constant"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-push-manager/internal/mq"
	pb "github.com/Tencent/bk-bcs/bcs-services/bcs-push-manager/proto"
)

// PushManagerService implements the handler methods for the PushManager service.
type PushManagerService struct {
	pushEventAction     *action.PushEventAction
	pushWhitelistAction *action.PushWhitelistAction
	pushTemplateAction  *action.PushTemplateAction
	mq                  mq.MQ
}

// NewPushManagerService creates a new PushManagerService instance.
func NewPushManagerService(
	pushEventAction *action.PushEventAction,
	pushWhitelistAction *action.PushWhitelistAction,
	pushTemplateAction *action.PushTemplateAction,
	mq mq.MQ,
) *PushManagerService {
	return &PushManagerService{
		pushEventAction:     pushEventAction,
		pushWhitelistAction: pushWhitelistAction,
		pushTemplateAction:  pushTemplateAction,
		mq:                  mq,
	}
}

// ===== event Related =====

// CreatePushEvent handles the creation of a event.
func (p *PushManagerService) CreatePushEvent(ctx context.Context, req *pb.CreatePushEventRequest, rsp *pb.CreatePushEventResponse) error {
	blog.Infof("create event received, req %+v", req)
	err := p.pushEventAction.CreatePushEvent(ctx, req, rsp)
	if err != nil {
		blog.Errorf("create event failed to create in db, err %v", err)
		return err
	}
	if rsp.Code != uint32(constant.ResponseCodeSuccess) {
		blog.Errorf("create event returned non-success code %d", rsp.Code)
		return nil
	}
	blog.Infof("create event successfully created in db, eventID %s, pushing to mq", rsp.EventId)
	if err := p.sendEventToMQ(req, rsp); err != nil {
		blog.Errorf("create event failed to push to mq, err %v", err)
		return err
	}
	blog.Infof("create event successfully pushed to mq, eventID %s", rsp.EventId)
	return nil
}

// DeletePushEvent handles the deletion of a event.
func (p *PushManagerService) DeletePushEvent(ctx context.Context, req *pb.DeletePushEventRequest, rsp *pb.DeletePushEventResponse) error {
	blog.Infof("delete event received, req %+v", req)
	err := p.pushEventAction.DeletePushEvent(ctx, req, rsp)
	if err != nil {
		blog.Errorf("delete event failed, err %v", err)
	} else {
		blog.Infof("delete event success, rsp %+v", rsp)
	}
	return err
}

// GetPushEvent handles the retrieval of a single event.
func (p *PushManagerService) GetPushEvent(ctx context.Context, req *pb.GetPushEventRequest, rsp *pb.GetPushEventResponse) error {
	blog.Infof("get event received, req %+v", req)
	err := p.pushEventAction.GetPushEvent(ctx, req, rsp)
	if err != nil {
		blog.Errorf("get event failed, err %v", err)
	} else {
		blog.Infof("get event success, eventID %s", req.EventId)
	}
	return err
}

// ListPushEvents handles the listing of events.
func (p *PushManagerService) ListPushEvents(ctx context.Context, req *pb.ListPushEventsRequest, rsp *pb.ListPushEventsResponse) error {
	blog.Infof("list events received, req %+v", req)
	err := p.pushEventAction.ListPushEvents(ctx, req, rsp)
	if err != nil {
		blog.Errorf("list events failed, err %v", err)
	} else {
		blog.Infof("list events success, total %d", rsp.Total)
	}
	return err
}

// UpdatePushEvent handles the update of a event.
func (p *PushManagerService) UpdatePushEvent(ctx context.Context, req *pb.UpdatePushEventRequest, rsp *pb.UpdatePushEventResponse) error {
	blog.Infof("update event received, req %+v", req)
	err := p.pushEventAction.UpdatePushEvent(ctx, req, rsp)
	if err != nil {
		blog.Errorf("update event failed, err %v", err)
	} else {
		blog.Infof("update event success, eventID %s", req.EventId)
	}
	return err
}

// ===== whitelist Related =====

// CreatePushWhitelist handles the creation of a whitelist.
func (p *PushManagerService) CreatePushWhitelist(ctx context.Context, req *pb.CreatePushWhitelistRequest, rsp *pb.CreatePushWhitelistResponse) error {
	blog.Infof("create whitelist received, req %+v", req)
	err := p.pushWhitelistAction.CreatePushWhitelist(ctx, req, rsp)
	if err != nil {
		blog.Errorf("create whitelist failed, err %v", err)
	} else {
		blog.Infof("create whitelist success, rsp %+v", rsp)
	}
	return err
}

// DeletePushWhitelist handles the deletion of a whitelist.
func (p *PushManagerService) DeletePushWhitelist(ctx context.Context, req *pb.DeletePushWhitelistRequest, rsp *pb.DeletePushWhitelistResponse) error {
	blog.Infof("delete whitelist received, req %+v", req)
	err := p.pushWhitelistAction.DeletePushWhitelist(ctx, req, rsp)
	if err != nil {
		blog.Errorf("delete whitelist failed, err %v", err)
	} else {
		blog.Infof("delete whitelist success, rsp %+v", rsp)
	}
	return err
}

// UpdatePushWhitelist handles the update of a whitelist.
func (p *PushManagerService) UpdatePushWhitelist(ctx context.Context, req *pb.UpdatePushWhitelistRequest, rsp *pb.UpdatePushWhitelistResponse) error {
	blog.Infof("update whitelist received, req %+v", req)
	err := p.pushWhitelistAction.UpdatePushWhitelist(ctx, req, rsp)
	if err != nil {
		blog.Errorf("update whitelist failed, err %v", err)
	} else {
		blog.Infof("update whitelist success, whitelistID %s", req.WhitelistId)
	}
	return err
}

// ListPushWhitelists handles the listing of whitelists.
func (p *PushManagerService) ListPushWhitelists(ctx context.Context, req *pb.ListPushWhitelistsRequest, rsp *pb.ListPushWhitelistsResponse) error {
	blog.Infof("list whitelists received, req %+v", req)
	err := p.pushWhitelistAction.ListPushWhitelists(ctx, req, rsp)
	if err != nil {
		blog.Errorf("list whitelists failed, err %v", err)
	} else {
		blog.Infof("list whitelists success, total %d", rsp.Total)
	}
	return err
}

// GetPushWhitelist handles the retrieval of a single whitelist.
func (p *PushManagerService) GetPushWhitelist(ctx context.Context, req *pb.GetPushWhitelistRequest, rsp *pb.GetPushWhitelistResponse) error {
	blog.Infof("get whitelist received, req %+v", req)
	err := p.pushWhitelistAction.GetPushWhitelist(ctx, req, rsp)
	if err != nil {
		blog.Errorf("get whitelist failed, err %v", err)
	} else {
		blog.Infof("get whitelist success, whitelistID %s", req.WhitelistId)
	}
	return err
}

// ===== Push Template Related =====

// CreatePushTemplate handles the creation of a template.
func (p *PushManagerService) CreatePushTemplate(ctx context.Context, req *pb.CreatePushTemplateRequest, rsp *pb.CreatePushTemplateResponse) error {
	blog.Infof("create template received, req %+v", req)
	err := p.pushTemplateAction.CreatePushTemplate(ctx, req, rsp)
	if err != nil {
		blog.Errorf("create template failed, err %v", err)
	} else {
		blog.Infof("create template success, rsp %+v", rsp)
	}
	return err
}

// DeletePushTemplate handles the deletion of a template.
func (p *PushManagerService) DeletePushTemplate(ctx context.Context, req *pb.DeletePushTemplateRequest, rsp *pb.DeletePushTemplateResponse) error {
	blog.Infof("delete template received, req %+v", req)
	err := p.pushTemplateAction.DeletePushTemplate(ctx, req, rsp)
	if err != nil {
		blog.Errorf("delete template failed, err %v", err)
	} else {
		blog.Infof("delete template success, rsp %+v", rsp)
	}
	return err
}

// UpdatePushTemplate handles the update of a template.
func (p *PushManagerService) UpdatePushTemplate(ctx context.Context, req *pb.UpdatePushTemplateRequest, rsp *pb.UpdatePushTemplateResponse) error {
	blog.Infof("update template received, req %+v", req)
	err := p.pushTemplateAction.UpdatePushTemplate(ctx, req, rsp)
	if err != nil {
		blog.Errorf("update template failed, err %v", err)
	} else {
		blog.Infof("update template success, templateID %s", req.TemplateId)
	}
	return err
}

// ListPushTemplates handles the listing of templates.
func (p *PushManagerService) ListPushTemplates(ctx context.Context, req *pb.ListPushTemplatesRequest, rsp *pb.ListPushTemplatesResponse) error {
	blog.Infof("list templates received, req %+v", req)
	err := p.pushTemplateAction.ListPushTemplates(ctx, req, rsp)
	if err != nil {
		blog.Errorf("list templates failed, err %v", err)
	} else {
		blog.Infof("list templates success, total %d", rsp.Total)
	}
	return err
}

// GetPushTemplate handles the retrieval of a template.
func (p *PushManagerService) GetPushTemplate(ctx context.Context, req *pb.GetPushTemplateRequest, rsp *pb.GetPushTemplateResponse) error {
	blog.Infof("get template received, req %+v", req)
	err := p.pushTemplateAction.GetPushTemplate(ctx, req, rsp)
	if err != nil {
		blog.Errorf("get template failed, err %v", err)
	} else {
		blog.Infof("get template success, templateID %s", req.TemplateId)
	}
	return err
}

// sendEventToMQ sends the event to the message queue.
func (p *PushManagerService) sendEventToMQ(req *pb.CreatePushEventRequest, rsp *pb.CreatePushEventResponse) error {
	if req == nil {
		blog.Errorf("sendEventToMQ failed: req is nil")
		return fmt.Errorf("req is nil")
	}
	if req.Event == nil {
		blog.Errorf("sendEventToMQ failed: req.Event is nil")
		return fmt.Errorf("req.Event is nil")
	}
	event := req.Event
	if event.EventDetail == nil {
		blog.Errorf("sendEventToMQ failed: event.EventDetail is nil")
		return fmt.Errorf("event.EventDetail is nil")
	}

	if rsp == nil {
		blog.Errorf("sendEventToMQ failed: rsp is nil")
		return fmt.Errorf("rsp is nil")
	}

	title := event.EventDetail.Fields[constant.EventDetailKeyTitle]
	content := event.EventDetail.Fields[constant.EventDetailKeyContent]
	receiversStr := event.EventDetail.Fields[constant.EventDetailKeyReceivers]
	receivers := strings.Split(receiversStr, ",")
	pushChannelsStr := event.EventDetail.Fields[constant.EventDetailKeyTypes]
	pushChannels := strings.Split(pushChannelsStr, ",")
	message := &mq.PushEventMessage{
		EventID:   rsp.EventId,
		Type:      pushChannels,
		Title:     title,
		Content:   content,
		Receivers: receivers,
		Extra:     event.EventDetail.Fields,
	}
	body, err := message.Marshal()
	if err != nil {
		blog.Errorf("create event failed to marshal message, err %s", err.Error())
		return err
	}
	routingKey := fmt.Sprintf(constant.MQRoutingKeyFormat, req.Domain)
	if err := p.mq.Publish(routingKey, body); err != nil {
		blog.Errorf("create event failed to publish message, err %s", err.Error())
		return err
	}
	blog.Infof("create event successfully pushed to mq, eventID %s", rsp.EventId)
	return nil
}
