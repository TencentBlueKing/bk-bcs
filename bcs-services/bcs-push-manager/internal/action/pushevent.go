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
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-push-manager/internal/constant"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-push-manager/internal/store/mongo"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-push-manager/internal/store/types"
	pb "github.com/Tencent/bk-bcs/bcs-services/bcs-push-manager/proto"
)

// PushEventAction defines the business logic for handling push event operations.
type PushEventAction struct {
	store *mongo.ModelPushEvent
}

// NewPushEventAction creates a new PushEventAction instance.
func NewPushEventAction(store *mongo.ModelPushEvent) *PushEventAction {
	return &PushEventAction{
		store: store,
	}
}

// CreatePushEvent handles the logic for creating a new push event.
func (a *PushEventAction) CreatePushEvent(ctx context.Context, req *pb.CreatePushEventRequest, rsp *pb.CreatePushEventResponse) error {
	// validate request
	if req.Domain == "" {
		rsp.Code = uint32(constant.ResponseCodeBadRequest)
		rsp.Message = constant.ResponseMsgDomainRequired
		return nil
	}
	if req.Event == nil {
		rsp.Code = uint32(constant.ResponseCodeBadRequest)
		rsp.Message = constant.ResponseMsgEventRequired
		return nil
	}
	if req.Event.EventId == "" {
		req.Event.EventId = a.generateEventID(req.Domain)
	}
	if !a.isValidPushLevel(req.Event.PushLevel) {
		req.Event.PushLevel = constant.AlertLevelWarning
	}

	// convert to internal type
	event := &types.PushEvent{
		EventID:             req.Event.EventId,
		Domain:              req.Domain,
		RuleID:              req.Event.RuleId,
		PushLevel:           req.Event.PushLevel,
		Status:              int(constant.EventStatusPending),
		BkBizName:           req.Event.BkBizName,
		NotificationResults: types.NotificationResults{Fields: map[string]string{}},
	}

	// convert EventDetail
	if req.Event.EventDetail != nil {
		event.EventDetail = types.EventDetail{
			Fields: req.Event.EventDetail.Fields,
		}
		if err := a.validateEventDetail(&event.EventDetail, rsp); err != nil {
			return nil
		}
	} else {
		rsp.Code = uint32(constant.ResponseCodeBadRequest)
		rsp.Message = constant.ResponseMsgEventDetailRequired
		return nil
	}

	// convert Dimension
	if req.Event.Dimension != nil {
		event.Dimension = types.Dimension{
			Fields: req.Event.Dimension.Fields,
		}
		if len(event.Dimension.Fields) == 0 {
			rsp.Code = uint32(constant.ResponseCodeBadRequest)
			rsp.Message = constant.ResponseMsgDimensionFieldsRequired
			return nil
		}
	} else {
		rsp.Code = uint32(constant.ResponseCodeBadRequest)
		rsp.Message = constant.ResponseMsgDimensionRequired
		return nil
	}

	if req.Event.MetricData != nil {
		event.MetricData = types.MetricData{
			MetricValue: req.Event.MetricData.MetricValue,
		}
		if req.Event.MetricData.Timestamp != nil {
			event.MetricData.Timestamp = req.Event.MetricData.Timestamp.AsTime()
		}
	} else {
		rsp.Code = uint32(constant.ResponseCodeBadRequest)
		rsp.Message = constant.ResponseMsgMetricDataRequired
		return nil
	}

	// call store layer
	err := a.store.CreatePushEvent(ctx, event)
	if err != nil {
		rsp.Code = uint32(constant.ResponseCodeInternalError)
		rsp.Message = fmt.Sprintf("failed to create push event: %v", err)
		return nil
	}

	// fill response
	rsp.Code = uint32(constant.ResponseCodeSuccess)
	rsp.Message = constant.ResponseMsgSuccess
	rsp.EventId = event.EventID

	return nil
}

// DeletePushEvent handles the logic for deleting a push event.
func (a *PushEventAction) DeletePushEvent(ctx context.Context, req *pb.DeletePushEventRequest, rsp *pb.DeletePushEventResponse) error {
	// validate request
	if req.Domain == "" {
		rsp.Code = uint32(constant.ResponseCodeBadRequest)
		rsp.Message = constant.ResponseMsgDomainRequired
		return nil
	}
	if req.EventId == "" {
		rsp.Code = uint32(constant.ResponseCodeBadRequest)
		rsp.Message = constant.ResponseMsgEventIDRequired
		return nil
	}

	// call store layer
	err := a.store.DeletePushEvent(ctx, req.EventId)
	if err != nil {
		rsp.Code = uint32(constant.ResponseCodeInternalError)
		rsp.Message = fmt.Sprintf("failed to delete push event: %v", err)
		return nil
	}

	// fill response
	rsp.Code = uint32(constant.ResponseCodeSuccess)
	rsp.Message = constant.ResponseMsgSuccess

	return nil
}

// GetPushEvent handles the logic for retrieving a single push event.
func (a *PushEventAction) GetPushEvent(ctx context.Context, req *pb.GetPushEventRequest, rsp *pb.GetPushEventResponse) error {
	// validate request
	if req.Domain == "" {
		rsp.Code = uint32(constant.ResponseCodeBadRequest)
		rsp.Message = constant.ResponseMsgDomainRequired
		return nil
	}
	if req.EventId == "" {
		rsp.Code = uint32(constant.ResponseCodeBadRequest)
		rsp.Message = constant.ResponseMsgEventIDRequired
		return nil
	}

	// call store layer
	event, err := a.store.GetPushEvent(ctx, req.EventId)
	if err != nil {
		rsp.Code = uint32(constant.ResponseCodeInternalError)
		rsp.Message = fmt.Sprintf("failed to get push event: %v", err)
		return nil
	}
	if event == nil {
		rsp.Code = uint32(constant.ResponseCodeNotFound)
		rsp.Message = constant.ResponseMsgPushEventNotFound
		return nil
	}

	// convert to pb type
	pbEvent := &pb.PushEvent{
		EventId:   event.EventID,
		Domain:    event.Domain,
		RuleId:    event.RuleID,
		PushLevel: event.PushLevel,
		Status:    int32(event.Status),
		BkBizName: event.BkBizName,
		EventDetail: &pb.EventDetail{
			Fields: event.EventDetail.Fields,
		},
		NotificationResults: &pb.NotificationResults{
			Fields: event.NotificationResults.Fields,
		},
		Dimension: &pb.Dimension{
			Fields: event.Dimension.Fields,
		},
		MetricData: &pb.MetricData{
			MetricValue: event.MetricData.MetricValue,
			Timestamp:   timestamppb.New(event.MetricData.Timestamp),
		},
		CreatedAt: timestamppb.New(event.CreatedAt),
		UpdatedAt: timestamppb.New(event.UpdatedAt),
	}

	// fill response
	rsp.Code = uint32(constant.ResponseCodeSuccess)
	rsp.Message = constant.ResponseMsgSuccess
	rsp.Event = pbEvent

	return nil
}

// ListPushEvents handles the logic for listing push events.
func (a *PushEventAction) ListPushEvents(ctx context.Context, req *pb.ListPushEventsRequest, rsp *pb.ListPushEventsResponse) error {
	// validate request
	if req.Domain == "" {
		rsp.Code = uint32(constant.ResponseCodeBadRequest)
		rsp.Message = constant.ResponseMsgDomainRequired
		return nil
	}

	// set default pagination
	page := int64(req.Page)
	if page < 1 {
		page = constant.DefaultPage
	}
	pageSize := int64(req.PageSize)
	if pageSize <= 0 {
		pageSize = constant.DefaultPageSize
	}

	// build filter
	filter := operator.M{"domain": req.Domain}
	if req.RuleId != "" {
		filter["rule_id"] = req.RuleId
	}
	if req.Status != 0 {
		filter["status"] = req.Status
	}
	if req.PushLevel != "" {
		filter["push_level"] = req.PushLevel
	}
	if req.StartTime != nil && req.EndTime != nil {
		filter["created_at"] = operator.M{
			"$gte": req.StartTime.AsTime(),
			"$lte": req.EndTime.AsTime(),
		}
	}

	events, total, err := a.store.ListPushEvents(ctx, filter, page, pageSize)
	if err != nil {
		rsp.Code = uint32(constant.ResponseCodeInternalError)
		rsp.Message = fmt.Sprintf("failed to list push events: %v", err)
		return nil
	}

	// convert to pb type
	var pbEvents []*pb.PushEvent
	for _, event := range events {
		pbEvent := &pb.PushEvent{
			EventId:   event.EventID,
			Domain:    event.Domain,
			RuleId:    event.RuleID,
			PushLevel: event.PushLevel,
			Status:    int32(event.Status),
			BkBizName: event.BkBizName,
			EventDetail: &pb.EventDetail{
				Fields: event.EventDetail.Fields,
			},
			NotificationResults: &pb.NotificationResults{
				Fields: event.NotificationResults.Fields,
			},
			Dimension: &pb.Dimension{
				Fields: event.Dimension.Fields,
			},
			MetricData: &pb.MetricData{
				MetricValue: event.MetricData.MetricValue,
				Timestamp:   timestamppb.New(event.MetricData.Timestamp),
			},
			CreatedAt: timestamppb.New(event.CreatedAt),
			UpdatedAt: timestamppb.New(event.UpdatedAt),
		}
		pbEvents = append(pbEvents, pbEvent)
	}

	// fill response
	rsp.Code = uint32(constant.ResponseCodeSuccess)
	rsp.Message = constant.ResponseMsgSuccess
	rsp.Events = pbEvents
	rsp.Total = total

	return nil
}

// UpdatePushEvent handles the logic for updating a push event.
func (a *PushEventAction) UpdatePushEvent(ctx context.Context, req *pb.UpdatePushEventRequest, rsp *pb.UpdatePushEventResponse) error {
	// validate request
	if req.Domain == "" {
		rsp.Code = uint32(constant.ResponseCodeBadRequest)
		rsp.Message = constant.ResponseMsgDomainRequired
		return nil
	}
	if req.EventId == "" {
		rsp.Code = uint32(constant.ResponseCodeBadRequest)
		rsp.Message = constant.ResponseMsgEventIDRequired
		return nil
	}
	if req.Event == nil {
		rsp.Code = uint32(constant.ResponseCodeBadRequest)
		rsp.Message = constant.ResponseMsgEventRequired
		return nil
	}

	// build update fields
	updateFields := operator.M{}
	if req.Event.RuleId != "" {
		updateFields["rule_id"] = req.Event.RuleId
	}
	if req.Event.PushLevel != "" {
		updateFields["push_level"] = req.Event.PushLevel
	}
	if req.Event.Status != 0 {
		updateFields["status"] = req.Event.Status
	}
	if req.Event.BkBizName != "" {
		updateFields["bk_biz_name"] = req.Event.BkBizName
	}
	if req.Event.EventDetail != nil {
		updateFields["event_detail"] = types.EventDetail{
			Fields: req.Event.EventDetail.Fields,
		}
	}
	if req.Event.NotificationResults != nil {
		updateFields["notification_results"] = types.NotificationResults{
			Fields: req.Event.NotificationResults.Fields,
		}
	}
	if req.Event.Dimension != nil {
		updateFields["dimension"] = types.Dimension{
			Fields: req.Event.Dimension.Fields,
		}
	}
	if req.Event.MetricData != nil {
		metricData := types.MetricData{
			MetricValue: req.Event.MetricData.MetricValue,
		}
		if req.Event.MetricData.Timestamp != nil {
			metricData.Timestamp = req.Event.MetricData.Timestamp.AsTime()
		}
		updateFields["metric_data"] = metricData
	}

	if len(updateFields) == 0 {
		rsp.Code = uint32(constant.ResponseCodeBadRequest)
		rsp.Message = constant.ResponseMsgNoFieldsToUpdate
		return nil
	}

	update := operator.M{"$set": updateFields}

	// call store layer
	err := a.store.UpdatePushEvent(ctx, req.EventId, update)
	if err != nil {
		rsp.Code = uint32(constant.ResponseCodeInternalError)
		rsp.Message = fmt.Sprintf("failed to update push event: %v", err)
		return nil
	}

	// fill response
	rsp.Code = uint32(constant.ResponseCodeSuccess)
	rsp.Message = constant.ResponseMsgSuccess

	return nil
}

// generateEventID xxx
func (a *PushEventAction) generateEventID(domain string) string {
	timestamp := time.Now().Format("20060102150405")
	source := domain + timestamp
	hash := md5.Sum([]byte(source))
	hashPrefix := hex.EncodeToString(hash[:8])
	eventID := fmt.Sprintf("%s_%s_%s", domain, timestamp, hashPrefix)
	if len(eventID) > 32 {
		return eventID[:32]
	}
	return eventID
}

// isValidPushLevel xxx
func (a *PushEventAction) isValidPushLevel(level string) bool {
	validLevels := map[string]bool{
		constant.AlertLevelFatal:    true,
		constant.AlertLevelWarning:  true,
		constant.AlertLevelReminder: true,
	}
	return validLevels[level]
}

// validateEventDetail xxx
func (a *PushEventAction) validateEventDetail(detail *types.EventDetail, rsp *pb.CreatePushEventResponse) error {
	fields := detail.Fields

	pushTypes := strings.Split(fields[constant.EventDetailKeyTypes], ",")

	for _, pushType := range pushTypes {
		switch pushType {
		case constant.PushTypeRtx:
			requiredRTXFields := []string{
				constant.EventDetailKeyRTXReceivers,
				constant.EventDetailKeyRTXContent,
				constant.EventDetailKeyRTXTitle,
			}
			for _, field := range requiredRTXFields {
				if fields[field] == "" {
					rsp.Code = uint32(constant.ResponseCodeBadRequest)
					rsp.Message = fmt.Sprintf("event detail missing required non-empty field for RTX: %s", field)
					return errors.New("invalid event detail")
				}
			}
		case constant.PushTypeMail:
			requiredMailFields := []string{
				constant.EventDetailKeyMailReceivers,
				constant.EventDetailKeyMailContent,
				constant.EventDetailKeyMailTitle,
			}
			for _, field := range requiredMailFields {
				if fields[field] == "" {
					rsp.Code = uint32(constant.ResponseCodeBadRequest)
					rsp.Message = fmt.Sprintf("event detail missing required non-empty field for Mail: %s", field)
					return errors.New("invalid event detail")
				}
			}
		case constant.PushTypeMsg:
			requiredMsgFields := []string{
				constant.EventDetailKeyMsgReceivers,
				constant.EventDetailKeyMsgContent,
			}
			for _, field := range requiredMsgFields {
				if fields[field] == "" {
					rsp.Code = uint32(constant.ResponseCodeBadRequest)
					rsp.Message = fmt.Sprintf("event detail missing required non-empty field for Msg: %s", field)
					return errors.New("invalid event detail")
				}
			}
		default:
			rsp.Code = uint32(constant.ResponseCodeBadRequest)
			rsp.Message = fmt.Sprintf("event detail contains invalid push type: %s", pushType)
			return errors.New("invalid event detail")
		}
	}

	return nil
}
