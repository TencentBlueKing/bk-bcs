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

package service

import (
	"context"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	pbcontent "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/content"
	pbds "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/data-service"
)

// CreateContent create content.
func (s *Service) CreateContent(ctx context.Context, req *pbds.CreateContentReq) (*pbds.CreateResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	content := &table.Content{
		Spec:       req.Spec.ContentSpec(),
		Attachment: req.Attachment.ContentAttachment(),
		Revision: &table.CreatedRevision{
			Creator: grpcKit.User,
		},
	}
	id, err := s.dao.Content().Create(grpcKit, content)
	if err != nil {
		logs.Errorf("create content failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp := &pbds.CreateResp{Id: id}
	return resp, nil
}

// GetContent get content by id
func (s *Service) GetContent(ctx context.Context, req *pbds.GetContentReq) (*pbcontent.Content, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	content, err := s.dao.Content().Get(grpcKit, req.Id, req.BizId)
	if err != nil {
		logs.Errorf("list content failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp := pbcontent.PbContent(content)
	return resp, nil
}
