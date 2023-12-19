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

// Package pbtemplate provides template core protocol struct and convert functions.
package pbtemplate

import (
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	pbbase "github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/base"
)

// Template convert pb Template to table Template
func (m *Template) Template() (*table.Template, error) {
	if m == nil {
		return nil, nil
	}

	return &table.Template{
		ID:         m.Id,
		Spec:       m.Spec.TemplateSpec(),
		Attachment: m.Attachment.TemplateAttachment(),
	}, nil
}

// TemplateSpec convert pb TemplateSpec to table TemplateSpec
func (m *TemplateSpec) TemplateSpec() *table.TemplateSpec {
	if m == nil {
		return nil
	}

	return &table.TemplateSpec{
		Name: m.Name,
		Path: m.Path,
		Memo: m.Memo,
	}
}

// PbTemplateSpec convert table TemplateSpec to pb TemplateSpec
func PbTemplateSpec(spec *table.TemplateSpec) *TemplateSpec { //nolint:revive
	if spec == nil {
		return nil
	}

	return &TemplateSpec{
		Name: spec.Name,
		Path: spec.Path,
		Memo: spec.Memo,
	}
}

// TemplateAttachment convert pb TemplateAttachment to table TemplateAttachment
func (m *TemplateAttachment) TemplateAttachment() *table.TemplateAttachment {
	if m == nil {
		return nil
	}

	return &table.TemplateAttachment{
		BizID:           m.BizId,
		TemplateSpaceID: m.TemplateSpaceId,
	}
}

// PbTemplateAttachment convert table TemplateAttachment to pb TemplateAttachment
func PbTemplateAttachment(at *table.TemplateAttachment) *TemplateAttachment { //nolint:revive
	if at == nil {
		return nil
	}

	return &TemplateAttachment{
		BizId:           at.BizID,
		TemplateSpaceId: at.TemplateSpaceID,
	}
}

// PbTemplates convert table Template to pb Template
func PbTemplates(s []*table.Template) []*Template {
	if s == nil {
		return make([]*Template, 0)
	}

	result := make([]*Template, 0)
	for _, one := range s {
		result = append(result, PbTemplate(one))
	}

	return result
}

// PbTemplate convert table Template to pb Template
func PbTemplate(s *table.Template) *Template {
	if s == nil {
		return nil
	}

	return &Template{
		Id:         s.ID,
		Spec:       PbTemplateSpec(s.Spec),
		Attachment: PbTemplateAttachment(s.Attachment),
		Revision:   pbbase.PbRevision(s.Revision),
	}
}
