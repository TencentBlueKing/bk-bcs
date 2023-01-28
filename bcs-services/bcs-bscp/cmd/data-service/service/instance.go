/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "as IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package service

import (
	"context"
	"time"

	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	pbbase "bscp.io/pkg/protocol/core/base"
	pbinstance "bscp.io/pkg/protocol/core/instance"
	pbds "bscp.io/pkg/protocol/data-service"
	"bscp.io/pkg/types"
)

// CreateCRInstance create current released instance.
func (s *Service) CreateCRInstance(ctx context.Context, req *pbds.CreateCRInstanceReq) (*pbds.CreateResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	instance := &table.CurrentReleasedInstance{
		Spec:       req.Spec.ReleasedInstanceSpec(),
		Attachment: req.Attachment.ReleaseAttachment(),
		Revision: &table.CreatedRevision{
			Creator:   grpcKit.User,
			CreatedAt: time.Now(),
		},
	}
	id, err := s.dao.CRInstance().Create(grpcKit, instance)
	if err != nil {
		logs.Errorf("create current released instance failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp := &pbds.CreateResp{Id: id}
	return resp, nil
}

// ListCRInstances list current released instances by query condition.
func (s *Service) ListCRInstances(ctx context.Context, req *pbds.ListCRInstancesReq) (
	*pbds.ListCRInstancesResp, error) {

	grpcKit := kit.FromGrpcContext(ctx)

	// parse pb struct filter to filter.Expression.
	filter, err := pbbase.UnmarshalFromPbStructToExpr(req.Filter)
	if err != nil {
		logs.Errorf("unmarshal pb struct to expression failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	query := &types.ListCRInstancesOption{
		BizID:  req.BizId,
		Filter: filter,
		Page:   req.Page.BasePage(),
	}

	details, err := s.dao.CRInstance().List(grpcKit, query)
	if err != nil {
		logs.Errorf("list current released instance failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp := &pbds.ListCRInstancesResp{
		Count:   details.Count,
		Details: pbinstance.PbCRInstances(details.Details),
	}
	return resp, nil
}

// DeleteCRInstance delete current released instance.
func (s *Service) DeleteCRInstance(ctx context.Context, req *pbds.DeleteCRInstanceReq) (*pbbase.EmptyResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	instance := &table.CurrentReleasedInstance{
		ID:         req.Id,
		Attachment: req.Attachment.ReleaseAttachment(),
	}
	if err := s.dao.CRInstance().Delete(grpcKit, instance); err != nil {
		logs.Errorf("delete current released instance failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	return new(pbbase.EmptyResp), nil
}
