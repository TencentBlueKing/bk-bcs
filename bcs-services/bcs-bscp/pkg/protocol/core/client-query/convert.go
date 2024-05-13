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

// Package pbcq xxx
package pbcq

import (
	"encoding/json"

	structpb "google.golang.org/protobuf/types/known/structpb"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
)

// ClientQuerySpec convert pb ClientQuerySpec to table ClientQuerySpec
func (c *ClientQuerySpec) ClientQuerySpec() *table.ClientQuerySpec {
	if c == nil {
		return nil
	}

	return &table.ClientQuerySpec{
		Creator:         c.Creator,
		SearchName:      c.SearchName,
		SearchType:      c.ClientQuerySpec().SearchType,
		SearchCondition: c.ClientQuerySpec().SearchCondition,
		CreatedAt:       c.CreatedAt.AsTime(),
		UpdatedAt:       c.UpdatedAt.AsTime(),
	}
}

// PbClientQuerySpec convert table ClientQuerySpec to pb ClientQuerySpec
func PbClientQuerySpec(spec *table.ClientQuerySpec) *ClientQuerySpec { //nolint:revive
	if spec == nil {
		return nil
	}

	searchCondition := new(structpb.Struct)
	err := json.Unmarshal([]byte(spec.SearchCondition), &searchCondition)
	if err != nil {
		return nil
	}

	return &ClientQuerySpec{
		Creator:         spec.Creator,
		SearchName:      spec.SearchName,
		SearchType:      string(spec.SearchType),
		SearchCondition: searchCondition,
		CreatedAt:       timestamppb.New(spec.CreatedAt),
		UpdatedAt:       timestamppb.New(spec.UpdatedAt),
	}
}

// ClientQueryAttachment convert pb ClientQueryAttachment to table ClientQueryAttachment
func (c *ClientQueryAttachment) ClientQueryAttachment() *table.ClientQueryAttachment {
	if c == nil {
		return nil
	}

	return &table.ClientQueryAttachment{
		BizID: c.BizId,
		AppID: c.AppId,
	}
}

// PbClientQueryAttachment convert table PbClientQueryAttachment to pb PbClientQueryAttachment
func PbClientQueryAttachment(attachment *table.ClientQueryAttachment) *ClientQueryAttachment { // nolint
	if attachment == nil {
		return nil
	}
	return &ClientQueryAttachment{
		BizId: attachment.BizID,
		AppId: attachment.AppID,
	}
}

// PbClientQuery convert table ClientQuery to pb ClientQuery
func PbClientQuery(c *table.ClientQuery) *ClientQuery {
	if c == nil {
		return nil
	}

	return &ClientQuery{
		Id:         c.ID,
		Spec:       PbClientQuerySpec(c.Spec),
		Attachment: PbClientQueryAttachment(c.Attachment),
	}
}

// PbClientQuerys convert table ClientQuery to pb ClientQuery
func PbClientQuerys(c []*table.ClientQuery) []*ClientQuery {
	if c == nil {
		return make([]*ClientQuery, 0)
	}
	result := make([]*ClientQuery, 0)
	for _, v := range c {
		result = append(result, PbClientQuery(v))
	}
	return result
}
