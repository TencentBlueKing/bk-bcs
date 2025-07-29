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

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-push-manager/internal/constant"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-push-manager/internal/store/mongo"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-push-manager/internal/store/types"
	pb "github.com/Tencent/bk-bcs/bcs-services/bcs-push-manager/proto"
)

// PushTemplateAction defines the business logic for handling push template operations.
type PushTemplateAction struct {
	store *mongo.ModelPushTemplate
}

// NewPushTemplateAction creates a new PushTemplateAction instance.
func NewPushTemplateAction(store *mongo.ModelPushTemplate) *PushTemplateAction {
	return &PushTemplateAction{
		store: store,
	}
}

// CreatePushTemplate handles the logic for creating a new push template.
func (a *PushTemplateAction) CreatePushTemplate(ctx context.Context, req *pb.CreatePushTemplateRequest, rsp *pb.CreatePushTemplateResponse) error {
	// validate request
	if req.Domain == "" {
		rsp.Code = uint32(constant.ResponseCodeBadRequest)
		rsp.Message = constant.ResponseMsgDomainRequired
		return nil
	}
	if req.Template == nil {
		rsp.Code = uint32(constant.ResponseCodeBadRequest)
		rsp.Message = constant.ResponseMsgTemplateRequired
		return nil
	}
	if req.Template.TemplateId == "" {
		rsp.Code = uint32(constant.ResponseCodeBadRequest)
		rsp.Message = constant.ResponseMsgTemplateIDRequired
		return nil
	}

	// convert to internal type
	template := &types.PushTemplate{
		TemplateID:   req.Template.TemplateId,
		Domain:       req.Domain,
		TemplateType: req.Template.TemplateType,
		Creator:      req.Template.Creator,
	}

	// convert TemplateContent
	if req.Template.Content != nil {
		template.Content = types.TemplateContent{
			Title:     req.Template.Content.Title,
			Body:      req.Template.Content.Body,
			Variables: req.Template.Content.Variables,
		}
	}

	// call store layer
	err := a.store.CreatePushTemplate(ctx, template)
	if err != nil {
		rsp.Code = uint32(constant.ResponseCodeInternalError)
		rsp.Message = fmt.Sprintf("failed to create push template: %v", err)
		return nil
	}

	// fill response
	rsp.Code = uint32(constant.ResponseCodeSuccess)
	rsp.Message = constant.ResponseMsgSuccess

	return nil
}

// DeletePushTemplate handles the logic for deleting a push template.
func (a *PushTemplateAction) DeletePushTemplate(ctx context.Context, req *pb.DeletePushTemplateRequest, rsp *pb.DeletePushTemplateResponse) error {
	// validate request
	if req.Domain == "" {
		rsp.Code = uint32(constant.ResponseCodeBadRequest)
		rsp.Message = constant.ResponseMsgDomainRequired
		return nil
	}
	if req.TemplateId == "" {
		rsp.Code = uint32(constant.ResponseCodeBadRequest)
		rsp.Message = constant.ResponseMsgTemplateIDRequired
		return nil
	}

	// call store layer
	err := a.store.DeletePushTemplate(ctx, req.TemplateId)
	if err != nil {
		rsp.Code = uint32(constant.ResponseCodeInternalError)
		rsp.Message = fmt.Sprintf("failed to delete push template: %v", err)
		return nil
	}

	// fill response
	rsp.Code = uint32(constant.ResponseCodeSuccess)
	rsp.Message = constant.ResponseMsgSuccess

	return nil
}

// GetPushTemplate handles the logic for retrieving a single push template.
func (a *PushTemplateAction) GetPushTemplate(ctx context.Context, req *pb.GetPushTemplateRequest, rsp *pb.GetPushTemplateResponse) error {
	// validate request
	if req.Domain == "" {
		rsp.Code = uint32(constant.ResponseCodeBadRequest)
		rsp.Message = constant.ResponseMsgDomainRequired
		return nil
	}
	if req.TemplateId == "" {
		rsp.Code = uint32(constant.ResponseCodeBadRequest)
		rsp.Message = constant.ResponseMsgTemplateIDRequired
		return nil
	}

	// call store layer
	template, err := a.store.GetPushTemplate(ctx, req.TemplateId)
	if err != nil {
		rsp.Code = uint32(constant.ResponseCodeInternalError)
		rsp.Message = fmt.Sprintf("failed to get push template: %v", err)
		return nil
	}
	if template == nil {
		rsp.Code = uint32(constant.ResponseCodeNotFound)
		rsp.Message = constant.ResponseMsgPushTemplateNotFound
		return nil
	}

	// fill response
	rsp.Code = uint32(constant.ResponseCodeSuccess)
	rsp.Message = constant.ResponseMsgSuccess
	rsp.Template = convertToPbPushTemplate(template)

	return nil
}

// UpdatePushTemplate handles the logic for updating a push template.
func (a *PushTemplateAction) UpdatePushTemplate(ctx context.Context, req *pb.UpdatePushTemplateRequest, rsp *pb.UpdatePushTemplateResponse) error {
	// validate request
	if req.Domain == "" {
		rsp.Code = uint32(constant.ResponseCodeBadRequest)
		rsp.Message = constant.ResponseMsgDomainRequired
		return nil
	}
	if req.TemplateId == "" {
		rsp.Code = uint32(constant.ResponseCodeBadRequest)
		rsp.Message = constant.ResponseMsgTemplateIDRequired
		return nil
	}
	if req.Template == nil {
		rsp.Code = uint32(constant.ResponseCodeBadRequest)
		rsp.Message = constant.ResponseMsgTemplateRequired
		return nil
	}

	existingTemplate, err := a.store.GetPushTemplate(ctx, req.TemplateId)
	if err != nil {
		rsp.Code = uint32(constant.ResponseCodeInternalError)
		rsp.Message = fmt.Sprintf("failed to retrieve template, templateID: %s, err: %v", req.TemplateId, err)
		return nil
	}
	if existingTemplate == nil {
		rsp.Code = uint32(constant.ResponseCodeNotFound)
		rsp.Message = constant.ResponseMsgPushTemplateNotFound
		return nil
	}

	if err := validateDomainMatch(existingTemplate.Domain, req.Domain); err != nil {
		rsp.Code = uint32(constant.ResponseCodeBadRequest)
		rsp.Message = err.Error()
		return nil
	}

	// build update fields
	updateFields := operator.M{}
	if req.Template.TemplateType != "" {
		updateFields["template_type"] = req.Template.TemplateType
	}
	if req.Template.Creator != "" {
		updateFields["creator"] = req.Template.Creator
	}
	if req.Template.Content != nil {
		if req.Template.Content.Title == "" {
			rsp.Code = uint32(constant.ResponseCodeBadRequest)
			rsp.Message = constant.ResponseMsgTitleRequired
			return nil
		}
		if req.Template.Content.Body == "" {
			rsp.Code = uint32(constant.ResponseCodeBadRequest)
			rsp.Message = constant.ResponseMsgBodyRequired
			return nil
		}

		updateFields["content"] = types.TemplateContent{
			Title:     req.Template.Content.Title,
			Body:      req.Template.Content.Body,
			Variables: req.Template.Content.Variables,
		}
	}

	if len(updateFields) == 0 {
		rsp.Code = uint32(constant.ResponseCodeBadRequest)
		rsp.Message = constant.ResponseMsgNoFieldsToUpdate
		return nil
	}

	update := operator.M{"$set": updateFields}

	// call store layer
	err = a.store.UpdatePushTemplate(ctx, req.TemplateId, update)
	if err != nil {
		rsp.Code = uint32(constant.ResponseCodeInternalError)
		rsp.Message = fmt.Sprintf("failed to update push template: %v", err)
		return nil
	}

	// fill response
	rsp.Code = uint32(constant.ResponseCodeSuccess)
	rsp.Message = constant.ResponseMsgSuccess

	return nil
}

// ListPushTemplates handles the logic for listing push templates.
func (a *PushTemplateAction) ListPushTemplates(ctx context.Context, req *pb.ListPushTemplatesRequest, rsp *pb.ListPushTemplatesResponse) error {
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
	if req.TemplateType != "" {
		filter["template_type"] = req.TemplateType
	}
	if req.Creator != "" {
		filter["creator"] = req.Creator
	}

	// call store layer
	templates, total, err := a.store.ListPushTemplates(ctx, filter, page, pageSize)
	if err != nil {
		rsp.Code = uint32(constant.ResponseCodeInternalError)
		rsp.Message = fmt.Sprintf("failed to list push templates: %v", err)
		return nil
	}

	// convert to pb type
	var pbTemplates []*pb.PushTemplate
	for _, template := range templates {
		pbTemplates = append(pbTemplates, convertToPbPushTemplate(template))
	}

	// fill response
	rsp.Code = uint32(constant.ResponseCodeSuccess)
	rsp.Message = constant.ResponseMsgSuccess
	rsp.Templates = pbTemplates
	rsp.Total = total

	return nil
}

// convertToPbPushTemplate converts a PushTemplate from internal type to protobuf type.
func convertToPbPushTemplate(template *types.PushTemplate) *pb.PushTemplate {
	if template == nil {
		return nil
	}
	return &pb.PushTemplate{
		TemplateId:   template.TemplateID,
		Domain:       template.Domain,
		TemplateType: template.TemplateType,
		Creator:      template.Creator,
		Content: &pb.TemplateContent{
			Title:     template.Content.Title,
			Body:      template.Content.Body,
			Variables: template.Content.Variables,
		},
		CreatedAt: timestamppb.New(template.CreatedAt),
	}
}

// validateDomainMatch checks if the domain of the retrieved data matches the requested domain.
func validateDomainMatch(retrievedDomain, requestedDomain string) error {
	if retrievedDomain != requestedDomain {
		return fmt.Errorf(constant.ResponseMsgDomainMismatch)
	}
	return nil
}
