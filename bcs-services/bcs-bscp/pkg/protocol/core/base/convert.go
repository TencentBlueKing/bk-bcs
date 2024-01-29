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

// Package pbbase provides base core protocol struct and convert functions.
package pbbase

import (
	"errors"
	"time"

	pbstruct "github.com/golang/protobuf/ptypes/struct"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/runtime/filter"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/types"
)

// BasePage convert pb BasePage to types BasePage
func (m *BasePage) BasePage() *types.BasePage {
	if m == nil {
		return nil
	}

	return &types.BasePage{
		Start: m.Start,
		Limit: uint(m.Limit),
		Sort:  m.Sort,
		Order: types.Order(m.Order),
	}
}

// PbRevision convert table Revision to pb Revision
func PbRevision(r *table.Revision) *Revision {
	if r == nil {
		return nil
	}

	return &Revision{
		Creator:  r.Creator,
		Reviser:  r.Reviser,
		CreateAt: r.CreatedAt.Format(time.RFC3339),
		UpdateAt: r.UpdatedAt.Format(time.RFC3339),
	}
}

// PbCreatedRevision convert table PbCreatedRevision to pb PbCreatedRevision
func PbCreatedRevision(r *table.CreatedRevision) *CreatedRevision {
	if r == nil {
		return nil
	}

	return &CreatedRevision{
		Creator:  r.Creator,
		CreateAt: r.CreatedAt.Format(time.RFC3339),
	}
}

// UnmarshalFromPbStructToExpr parsed a expression from pb struct message.
func UnmarshalFromPbStructToExpr(st *pbstruct.Struct) (*filter.Expression, error) {
	if st == nil {
		return nil, errors.New("pb struct is nil")
	}

	bytes, err := st.MarshalJSON()
	if err != nil {
		return nil, err
	}

	ft := new(filter.Expression)
	if err := ft.UnmarshalJSON(bytes); err != nil {
		return nil, err
	}

	return ft, nil
}
