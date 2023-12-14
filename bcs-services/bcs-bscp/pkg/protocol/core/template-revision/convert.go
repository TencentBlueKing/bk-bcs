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

// Package pbtr provides template revision core protocol struct and convert functions.
package pbtr

import (
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	pbbase "github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/base"
	pbci "github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/config-item"
	pbcontent "github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/content"
)

// TemplateRevision convert pb TemplateRevision to table TemplateRevision
func (m *TemplateRevision) TemplateRevision() *table.TemplateRevision {
	if m == nil {
		return nil
	}

	return &table.TemplateRevision{
		ID:         m.Id,
		Spec:       m.Spec.TemplateRevisionSpec(),
		Attachment: m.Attachment.TemplateRevisionAttachment(),
	}
}

// TemplateRevisionSpec convert pb TemplateRevisionSpec to table TemplateRevisionSpec
func (m *TemplateRevisionSpec) TemplateRevisionSpec() *table.TemplateRevisionSpec {
	if m == nil {
		return nil
	}

	return &table.TemplateRevisionSpec{
		RevisionName: m.RevisionName,
		RevisionMemo: m.RevisionMemo,
		Name:         m.Name,
		Path:         m.Path,
		FileType:     table.FileFormat(m.FileType),
		FileMode:     table.FileMode(m.FileMode),
		Permission:   m.Permission.FilePermission(),
		ContentSpec:  m.ContentSpec.ContentSpec(),
	}
}

// PbTemplateRevisionSpec convert table TemplateRevisionSpec to pb TemplateRevisionSpec
func PbTemplateRevisionSpec(spec *table.TemplateRevisionSpec) *TemplateRevisionSpec {
	if spec == nil {
		return nil
	}

	return &TemplateRevisionSpec{
		RevisionName: spec.RevisionName,
		RevisionMemo: spec.RevisionMemo,
		Name:         spec.Name,
		Path:         spec.Path,
		FileType:     string(spec.FileType),
		FileMode:     string(spec.FileMode),
		Permission:   pbci.PbFilePermission(spec.Permission),
		ContentSpec:  pbcontent.PbContentSpec(spec.ContentSpec),
	}
}

// TemplateRevisionAttachment convert pb TemplateRevisionAttachment to table TemplateRevisionAttachment
func (m *TemplateRevisionAttachment) TemplateRevisionAttachment() *table.TemplateRevisionAttachment {
	if m == nil {
		return nil
	}

	return &table.TemplateRevisionAttachment{
		BizID:           m.BizId,
		TemplateSpaceID: m.TemplateSpaceId,
		TemplateID:      m.TemplateId,
	}
}

// PbTemplateRevisionAttachment convert table TemplateRevisionAttachment to pb TemplateRevisionAttachment
func PbTemplateRevisionAttachment(at *table.TemplateRevisionAttachment) *TemplateRevisionAttachment {
	if at == nil {
		return nil
	}

	return &TemplateRevisionAttachment{
		BizId:           at.BizID,
		TemplateSpaceId: at.TemplateSpaceID,
		TemplateId:      at.TemplateID,
	}
}

// PbTemplateRevisions convert table TemplateRevision to pb TemplateRevision
func PbTemplateRevisions(s []*table.TemplateRevision) []*TemplateRevision {
	if s == nil {
		return make([]*TemplateRevision, 0)
	}

	result := make([]*TemplateRevision, 0)
	for _, one := range s {
		result = append(result, PbTemplateRevision(one))
	}

	return result
}

// PbTemplateRevision convert table TemplateRevision to pb TemplateRevision
func PbTemplateRevision(s *table.TemplateRevision) *TemplateRevision {
	if s == nil {
		return nil
	}

	return &TemplateRevision{
		Id:         s.ID,
		Spec:       PbTemplateRevisionSpec(s.Spec),
		Attachment: PbTemplateRevisionAttachment(s.Attachment),
		Revision:   pbbase.PbCreatedRevision(s.Revision),
	}
}
