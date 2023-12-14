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

// Package pbtset provides template set core protocol struct and convert functions.
package pbtset

import (
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	pbbase "github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/base"
)

// TemplateSet convert pb TemplateSet to table TemplateSet
func (m *TemplateSet) TemplateSet() (*table.TemplateSet, error) {
	if m == nil {
		return nil, nil
	}

	return &table.TemplateSet{
		ID:         m.Id,
		Spec:       m.Spec.TemplateSetSpec(),
		Attachment: m.Attachment.TemplateSetAttachment(),
	}, nil
}

// TemplateSetSpec convert pb TemplateSetSpec to table TemplateSetSpec
func (m *TemplateSetSpec) TemplateSetSpec() *table.TemplateSetSpec {
	if m == nil {
		return nil
	}

	return &table.TemplateSetSpec{
		Name:        m.Name,
		Memo:        m.Memo,
		TemplateIDs: m.TemplateIds,
		Public:      m.Public,
		BoundApps:   m.BoundApps,
	}
}

// PbTemplateSetSpec convert table TemplateSetSpec to pb TemplateSetSpec
func PbTemplateSetSpec(spec *table.TemplateSetSpec) *TemplateSetSpec {
	if spec == nil {
		return nil
	}

	return &TemplateSetSpec{
		Name:        spec.Name,
		Memo:        spec.Memo,
		TemplateIds: spec.TemplateIDs,
		Public:      spec.Public,
		BoundApps:   spec.BoundApps,
	}
}

// TemplateSetAttachment convert pb TemplateSetAttachment to table TemplateSetAttachment
func (m *TemplateSetAttachment) TemplateSetAttachment() *table.TemplateSetAttachment {
	if m == nil {
		return nil
	}

	return &table.TemplateSetAttachment{
		BizID:           m.BizId,
		TemplateSpaceID: m.TemplateSpaceId,
	}
}

// PbTemplateSetAttachment convert table TemplateSetAttachment to pb TemplateSetAttachment
func PbTemplateSetAttachment(at *table.TemplateSetAttachment) *TemplateSetAttachment {
	if at == nil {
		return nil
	}

	return &TemplateSetAttachment{
		BizId:           at.BizID,
		TemplateSpaceId: at.TemplateSpaceID,
	}
}

// PbTemplateSets convert table TemplateSet to pb TemplateSet
func PbTemplateSets(s []*table.TemplateSet) []*TemplateSet {
	if s == nil {
		return make([]*TemplateSet, 0)
	}

	result := make([]*TemplateSet, 0)
	for _, one := range s {
		result = append(result, PbTemplateSet(one))
	}

	return result
}

// PbTemplateSet convert table TemplateSet to pb TemplateSet
func PbTemplateSet(s *table.TemplateSet) *TemplateSet {
	if s == nil {
		return nil
	}

	return &TemplateSet{
		Id:         s.ID,
		Spec:       PbTemplateSetSpec(s.Spec),
		Attachment: PbTemplateSetAttachment(s.Attachment),
		Revision:   pbbase.PbRevision(s.Revision),
	}
}
