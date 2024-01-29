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

// Package pbts provides template space core protocol struct and convert functions.
package pbts

import (
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	pbbase "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/base"
)

// TemplateSpace convert pb TemplateSpace to table TemplateSpace
func (m *TemplateSpace) TemplateSpace() (*table.TemplateSpace, error) {
	if m == nil {
		return nil, nil
	}

	return &table.TemplateSpace{
		ID:         m.Id,
		Spec:       m.Spec.TemplateSpaceSpec(),
		Attachment: m.Attachment.TemplateSpaceAttachment(),
	}, nil
}

// TemplateSpaceSpec convert pb TemplateSpaceSpec to table TemplateSpaceSpec
func (m *TemplateSpaceSpec) TemplateSpaceSpec() *table.TemplateSpaceSpec {
	if m == nil {
		return nil
	}

	return &table.TemplateSpaceSpec{
		Name: m.Name,
		Memo: m.Memo,
	}
}

// PbTemplateSpaceSpec convert table TemplateSpaceSpec to pb TemplateSpaceSpec
func PbTemplateSpaceSpec(spec *table.TemplateSpaceSpec) *TemplateSpaceSpec {
	if spec == nil {
		return nil
	}

	return &TemplateSpaceSpec{
		Name: spec.Name,
		Memo: spec.Memo,
	}
}

// TemplateSpaceAttachment convert pb TemplateSpaceAttachment to table TemplateSpaceAttachment
func (m *TemplateSpaceAttachment) TemplateSpaceAttachment() *table.TemplateSpaceAttachment {
	if m == nil {
		return nil
	}

	return &table.TemplateSpaceAttachment{
		BizID: m.BizId,
	}
}

// PbTemplateSpaceAttachment convert table TemplateSpaceAttachment to pb TemplateSpaceAttachment
func PbTemplateSpaceAttachment(at *table.TemplateSpaceAttachment) *TemplateSpaceAttachment {
	if at == nil {
		return nil
	}

	return &TemplateSpaceAttachment{
		BizId: at.BizID,
	}
}

// PbTemplateSpaces convert table TemplateSpace to pb TemplateSpace
func PbTemplateSpaces(s []*table.TemplateSpace) []*TemplateSpace {
	if s == nil {
		return make([]*TemplateSpace, 0)
	}

	result := make([]*TemplateSpace, 0)
	for _, one := range s {
		result = append(result, PbTemplateSpace(one))
	}

	return result
}

// PbTemplateSpace convert table TemplateSpace to pb TemplateSpace
func PbTemplateSpace(s *table.TemplateSpace) *TemplateSpace {
	if s == nil {
		return nil
	}

	return &TemplateSpace{
		Id:         s.ID,
		Spec:       PbTemplateSpaceSpec(s.Spec),
		Attachment: PbTemplateSpaceAttachment(s.Attachment),
		Revision:   pbbase.PbRevision(s.Revision),
	}
}
