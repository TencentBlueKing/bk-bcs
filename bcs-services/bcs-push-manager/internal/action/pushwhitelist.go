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

	"go.mongodb.org/mongo-driver/bson"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-push-manager/internal/constant"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-push-manager/internal/store/mongo"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-push-manager/internal/store/types"
	pb "github.com/Tencent/bk-bcs/bcs-services/bcs-push-manager/proto"
)

// PushWhitelistAction defines the business logic for handling push whitelist operations.
type PushWhitelistAction struct {
	store mongo.PushWhitelistStore
}

// NewPushWhitelistAction creates a new PushWhitelistAction instance.
func NewPushWhitelistAction(store mongo.PushWhitelistStore) *PushWhitelistAction {
	return &PushWhitelistAction{
		store: store,
	}
}

// CreatePushWhitelist handles the logic for creating a new push whitelist.
func (a *PushWhitelistAction) CreatePushWhitelist(ctx context.Context, req *pb.CreatePushWhitelistRequest, rsp *pb.CreatePushWhitelistResponse) error {
	// validate request
	if req.Domain == "" {
		rsp.Code = uint32(constant.ResponseCodeBadRequest)
		rsp.Message = constant.ResponseMsgDomainRequired
		return nil
	}
	if req.Whitelist == nil {
		rsp.Code = uint32(constant.ResponseCodeBadRequest)
		rsp.Message = constant.ResponseMsgWhitelistRequired
		return nil
	}
	if req.Whitelist.WhitelistId == "" {
		rsp.Code = uint32(constant.ResponseCodeBadRequest)
		rsp.Message = constant.ResponseMsgWhitelistIDRequired
		return nil
	}

	// convert to internal type
	whitelist := &types.PushWhitelist{
		WhitelistID:     req.Whitelist.WhitelistId,
		Domain:          req.Domain,
		Reason:          req.Whitelist.Reason,
		Applicant:       req.Whitelist.Applicant,
		Approver:        "",
		WhitelistStatus: func() *int { v := int(constant.WhitelistStatusNone); return &v }(),
		ApprovalStatus:  func() *int { v := int(constant.ApprovalStatusPending); return &v }(),
	}

	// convert Dimension
	if req.Whitelist.Dimension != nil {
		whitelist.Dimension = types.Dimension{
			Fields: req.Whitelist.Dimension.Fields,
		}
		if len(whitelist.Dimension.Fields) == 0 {
			rsp.Code = uint32(constant.ResponseCodeBadRequest)
			rsp.Message = constant.ResponseMsgDimensionFieldsRequired
			return nil
		}
	} else {
		rsp.Code = uint32(constant.ResponseCodeBadRequest)
		rsp.Message = constant.ResponseMsgDimensionRequired
		return nil
	}

	// convert time fields
	if req.Whitelist.StartTime == nil {
		rsp.Code = uint32(constant.ResponseCodeBadRequest)
		rsp.Message = constant.ResponseMsgStartTimeRequired
		return nil
	}
	whitelist.StartTime = req.Whitelist.StartTime.AsTime()

	if req.Whitelist.EndTime == nil {
		rsp.Code = uint32(constant.ResponseCodeBadRequest)
		rsp.Message = constant.ResponseMsgEndTimeRequired
		return nil
	}
	endTime := req.Whitelist.EndTime.AsTime()
	whitelist.EndTime = &endTime

	err := a.store.CreatePushWhitelist(ctx, whitelist)
	if err != nil {
		rsp.Code = uint32(constant.ResponseCodeInternalError)
		rsp.Message = fmt.Sprintf("failed to create push whitelist: %v", err)
		return nil
	}

	rsp.Code = uint32(constant.ResponseCodeSuccess)
	rsp.Message = constant.ResponseMsgSuccess

	return nil
}

// DeletePushWhitelist handles the logic for deleting a push whitelist.
func (a *PushWhitelistAction) DeletePushWhitelist(ctx context.Context, req *pb.DeletePushWhitelistRequest, rsp *pb.DeletePushWhitelistResponse) error {
	// validate request
	if req.Domain == "" {
		rsp.Code = uint32(constant.ResponseCodeBadRequest)
		rsp.Message = constant.ResponseMsgDomainRequired
		return nil
	}
	if req.WhitelistId == "" {
		rsp.Code = uint32(constant.ResponseCodeBadRequest)
		rsp.Message = constant.ResponseMsgWhitelistIDRequired
		return nil
	}

	// call store layer
	err := a.store.DeletePushWhitelist(ctx, req.WhitelistId)
	if err != nil {
		rsp.Code = uint32(constant.ResponseCodeInternalError)
		rsp.Message = fmt.Sprintf("failed to delete push whitelist: %v", err)
		return nil
	}

	// fill response
	rsp.Code = uint32(constant.ResponseCodeSuccess)
	rsp.Message = constant.ResponseMsgSuccess

	return nil
}

// GetPushWhitelist handles the logic for retrieving a single push whitelist.
func (a *PushWhitelistAction) GetPushWhitelist(ctx context.Context, req *pb.GetPushWhitelistRequest, rsp *pb.GetPushWhitelistResponse) error {
	// validate request
	if req.Domain == "" {
		rsp.Code = uint32(constant.ResponseCodeBadRequest)
		rsp.Message = constant.ResponseMsgDomainRequired
		return nil
	}
	if req.WhitelistId == "" {
		rsp.Code = uint32(constant.ResponseCodeBadRequest)
		rsp.Message = constant.ResponseMsgWhitelistIDRequired
		return nil
	}

	// call store layer
	whitelist, err := a.store.GetPushWhitelist(ctx, req.WhitelistId)
	if err != nil {
		rsp.Code = uint32(constant.ResponseCodeInternalError)
		rsp.Message = fmt.Sprintf("failed to get push whitelist: %v", err)
		return nil
	}
	if whitelist == nil {
		rsp.Code = uint32(constant.ResponseCodeNotFound)
		rsp.Message = constant.ResponseMsgPushWhitelistNotFound
		return nil
	}

	// convert to pb type
	pbWhitelist := &pb.PushWhitelist{
		WhitelistId:     whitelist.WhitelistID,
		Domain:          whitelist.Domain,
		Reason:          whitelist.Reason,
		Applicant:       whitelist.Applicant,
		Approver:        whitelist.Approver,
		WhitelistStatus: convertWhitelistStatus(whitelist.WhitelistStatus),
		ApprovalStatus:  convertApprovalStatus(whitelist.ApprovalStatus),
		Dimension: &pb.Dimension{
			Fields: whitelist.Dimension.Fields,
		},
		StartTime: timestamppb.New(whitelist.StartTime),
		CreatedAt: timestamppb.New(whitelist.CreatedAt),
		UpdatedAt: timestamppb.New(whitelist.UpdatedAt),
	}

	if whitelist.EndTime != nil {
		pbWhitelist.EndTime = timestamppb.New(*whitelist.EndTime)
	}
	if whitelist.ApprovedAt != nil {
		pbWhitelist.ApprovedAt = timestamppb.New(*whitelist.ApprovedAt)
	}

	// fill response
	rsp.Code = uint32(constant.ResponseCodeSuccess)
	rsp.Message = constant.ResponseMsgSuccess
	rsp.Whitelist = pbWhitelist

	return nil
}

// UpdatePushWhitelist handles the logic for updating a push whitelist.
func (a *PushWhitelistAction) UpdatePushWhitelist(ctx context.Context, req *pb.UpdatePushWhitelistRequest, rsp *pb.UpdatePushWhitelistResponse) error {
	// validate request
	if req.Domain == "" {
		rsp.Code = uint32(constant.ResponseCodeBadRequest)
		rsp.Message = constant.ResponseMsgDomainRequired
		return nil
	}
	if req.WhitelistId == "" {
		rsp.Code = uint32(constant.ResponseCodeBadRequest)
		rsp.Message = constant.ResponseMsgWhitelistIDRequired
		return nil
	}
	if req.Whitelist == nil {
		rsp.Code = uint32(constant.ResponseCodeBadRequest)
		rsp.Message = constant.ResponseMsgWhitelistRequired
		return nil
	}

	existingWhitelist, err := a.store.GetPushWhitelist(ctx, req.WhitelistId)
	if err != nil {
		rsp.Code = uint32(constant.ResponseCodeInternalError)
		rsp.Message = fmt.Sprintf("failed to retrieve whitelist: %v", err)
		return nil
	}
	if existingWhitelist == nil {
		rsp.Code = uint32(constant.ResponseCodeBadRequest)
		rsp.Message = constant.ResponseMsgPushWhitelistNotFound
		return nil
	}

	if err := validateDomainMatch(existingWhitelist.Domain, req.Domain); err != nil {
		rsp.Code = uint32(constant.ResponseCodeBadRequest)
		rsp.Message = err.Error()
		return nil
	}

	// build update fields
	updateFields := bson.M{}
	if req.Whitelist.Reason != "" {
		updateFields["reason"] = req.Whitelist.Reason
	}
	if req.Whitelist.Applicant != "" {
		updateFields["applicant"] = req.Whitelist.Applicant
	}
	if req.Whitelist.Approver != "" {
		updateFields["approver"] = req.Whitelist.Approver
	}
	if req.Whitelist.WhitelistStatus != nil {
		updateFields["whitelist_status"] = *req.Whitelist.WhitelistStatus
	}
	if req.Whitelist.ApprovalStatus != nil {
		updateFields["approval_status"] = *req.Whitelist.ApprovalStatus
	}
	if req.Whitelist.Dimension != nil {
		if len(req.Whitelist.Dimension.Fields) == 0 {
			rsp.Code = uint32(constant.ResponseCodeBadRequest)
			rsp.Message = constant.ResponseMsgDimensionFieldsRequired
			return nil
		}
		updateFields["dimension"] = types.Dimension{
			Fields: req.Whitelist.Dimension.Fields,
		}
	}
	if req.Whitelist.StartTime != nil {
		updateFields["start_time"] = req.Whitelist.StartTime.AsTime()
	}
	if req.Whitelist.EndTime != nil {
		endTime := req.Whitelist.EndTime.AsTime()
		updateFields["end_time"] = &endTime
	}
	if req.Whitelist.ApprovedAt != nil {
		approvedAt := req.Whitelist.ApprovedAt.AsTime()
		updateFields["approved_at"] = &approvedAt
	}

	if len(updateFields) == 0 {
		rsp.Code = uint32(constant.ResponseCodeBadRequest)
		rsp.Message = constant.ResponseMsgNoFieldsToUpdate
		return nil
	}

	update := bson.M{"$set": updateFields}

	// call store layer
	err = a.store.UpdatePushWhitelist(ctx, req.WhitelistId, update)
	if err != nil {
		rsp.Code = uint32(constant.ResponseCodeInternalError)
		rsp.Message = fmt.Sprintf("failed to update push whitelist: %v", err)
		return nil
	}

	// fill response
	rsp.Code = uint32(constant.ResponseCodeSuccess)
	rsp.Message = constant.ResponseMsgSuccess

	return nil
}

// ListPushWhitelists handles the logic for listing push whitelists.
func (a *PushWhitelistAction) ListPushWhitelists(ctx context.Context, req *pb.ListPushWhitelistsRequest, rsp *pb.ListPushWhitelistsResponse) error {
	// validate request
	if req.Domain == "" {
		rsp.Code = uint32(constant.ResponseCodeBadRequest)
		rsp.Message = constant.ResponseMsgDomainRequired
		return nil
	}

	// set default pagination
	page := int64(req.Page)
	if page <= 0 {
		page = constant.DefaultPage
	}
	pageSize := int64(req.PageSize)
	if pageSize <= 0 {
		pageSize = constant.DefaultPageSize
	}

	// build filter
	filter := bson.M{"domain": req.Domain}
	if req.Applicant != "" {
		filter["applicant"] = req.Applicant
	}
	if req.ApprovalStatus != nil {
		filter["approval_status"] = *req.ApprovalStatus
	}
	if req.WhitelistStatus != nil {
		filter["whitelist_status"] = *req.WhitelistStatus
	}

	// call store layer
	whitelists, total, err := a.store.ListPushWhitelists(ctx, filter, page, pageSize)
	if err != nil {
		rsp.Code = uint32(constant.ResponseCodeInternalError)
		rsp.Message = fmt.Sprintf("failed to list push whitelists: %v", err)
		return nil
	}

	// convert to pb type
	var pbWhitelists []*pb.PushWhitelist
	for _, whitelist := range whitelists {
		pbWhitelist := &pb.PushWhitelist{
			WhitelistId:     whitelist.WhitelistID,
			Domain:          whitelist.Domain,
			Reason:          whitelist.Reason,
			Applicant:       whitelist.Applicant,
			Approver:        whitelist.Approver,
			WhitelistStatus: convertWhitelistStatus(whitelist.WhitelistStatus),
			ApprovalStatus:  convertApprovalStatus(whitelist.ApprovalStatus),
			Dimension: &pb.Dimension{
				Fields: whitelist.Dimension.Fields,
			},
			StartTime: timestamppb.New(whitelist.StartTime),
			CreatedAt: timestamppb.New(whitelist.CreatedAt),
			UpdatedAt: timestamppb.New(whitelist.UpdatedAt),
		}

		if whitelist.EndTime != nil {
			pbWhitelist.EndTime = timestamppb.New(*whitelist.EndTime)
		}
		if whitelist.ApprovedAt != nil {
			pbWhitelist.ApprovedAt = timestamppb.New(*whitelist.ApprovedAt)
		}
		pbWhitelists = append(pbWhitelists, pbWhitelist)
	}

	// fill response
	rsp.Code = uint32(constant.ResponseCodeSuccess)
	rsp.Message = constant.ResponseMsgSuccess
	rsp.Whitelists = pbWhitelists
	rsp.Total = total

	return nil
}

// convertWhitelistStatus converts a whitelist status from *int to *int32.
func convertWhitelistStatus(status *int) *int32 {
	if status != nil {
		v := int32(*status)
		return &v
	}
	return nil
}

// convertApprovalStatus converts an approval status from *int to *int32.
func convertApprovalStatus(status *int) *int32 {
	if status != nil {
		v := int32(*status)
		return &v
	}
	return nil
}
