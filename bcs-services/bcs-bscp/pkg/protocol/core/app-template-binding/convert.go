/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package pbatb

import (
	"bscp.io/pkg/dal/table"
	pbbase "bscp.io/pkg/protocol/core/base"
)

// AppTemplateBinding convert pb AppTemplateBinding to table AppTemplateBinding
func (m *AppTemplateBinding) AppTemplateBinding() (*table.AppTemplateBinding, error) {
	if m == nil {
		return nil, nil
	}

	return &table.AppTemplateBinding{
		ID:         m.Id,
		Spec:       m.Spec.AppTemplateBindingSpec(),
		Attachment: m.Attachment.AppTemplateBindingAttachment(),
	}, nil
}

// AppTemplateBindingSpec convert pb AppTemplateBindingSpec to table AppTemplateBindingSpec
func (m *AppTemplateBindingSpec) AppTemplateBindingSpec() *table.AppTemplateBindingSpec {
	if m == nil {
		return nil
	}

	result := make([]*table.TemplateBinding, 0)
	for _, one := range m.Bindings {
		result = append(result, one.TemplateBinding())
	}

	return &table.AppTemplateBindingSpec{
		TemplateSpaceIDs:    m.TemplateSpaceIds,
		TemplateSetIDs:      m.TemplateSetIds,
		TemplateIDs:         m.TemplateIds,
		TemplateRevisionIDs: m.TemplateRevisionIds,
		Bindings:            result,
	}
}

// PbAppTemplateBindingSpec convert table AppTemplateBindingSpec to pb AppTemplateBindingSpec
func PbAppTemplateBindingSpec(spec *table.AppTemplateBindingSpec) *AppTemplateBindingSpec {
	if spec == nil {
		return nil
	}

	result := make([]*TemplateBinding, 0)
	for _, one := range spec.Bindings {
		result = append(result, PbTemplateBinding(one))
	}

	return &AppTemplateBindingSpec{
		TemplateSpaceIds:    spec.TemplateSpaceIDs,
		TemplateSetIds:      spec.TemplateSetIDs,
		TemplateIds:         spec.TemplateIDs,
		TemplateRevisionIds: spec.TemplateRevisionIDs,
		Bindings:            result,
	}
}

// TemplateBinding convert pb TemplateBinding to table TemplateBinding
func (m *TemplateBinding) TemplateBinding() *table.TemplateBinding {
	if m == nil {
		return nil
	}

	return &table.TemplateBinding{
		TemplateSetID:       m.TemplateSetId,
		TemplateRevisionIDs: m.TemplateRevisionIds,
	}
}

// PbTemplateBinding convert table TemplateBinding to pb TemplateBinding
func PbTemplateBinding(spec *table.TemplateBinding) *TemplateBinding {
	if spec == nil {
		return nil
	}

	return &TemplateBinding{
		TemplateSetId:       spec.TemplateSetID,
		TemplateRevisionIds: spec.TemplateRevisionIDs,
	}
}

// AppTemplateBindingAttachment convert pb AppTemplateBindingAttachment to table AppTemplateBindingAttachment
func (m *AppTemplateBindingAttachment) AppTemplateBindingAttachment() *table.AppTemplateBindingAttachment {
	if m == nil {
		return nil
	}

	return &table.AppTemplateBindingAttachment{
		BizID: m.BizId,
		AppID: m.AppId,
	}
}

// PbAppTemplateBindingAttachment convert table AppTemplateBindingAttachment to pb AppTemplateBindingAttachment
func PbAppTemplateBindingAttachment(at *table.AppTemplateBindingAttachment) *AppTemplateBindingAttachment {
	if at == nil {
		return nil
	}

	return &AppTemplateBindingAttachment{
		BizId: at.BizID,
		AppId: at.AppID,
	}
}

// PbAppTemplateBindings convert table AppTemplateBinding to pb AppTemplateBinding
func PbAppTemplateBindings(s []*table.AppTemplateBinding) []*AppTemplateBinding {
	if s == nil {
		return make([]*AppTemplateBinding, 0)
	}

	result := make([]*AppTemplateBinding, 0)
	for _, one := range s {
		result = append(result, PbAppTemplateBinding(one))
	}

	return result
}

// PbAppTemplateBinding convert table AppTemplateBinding to pb AppTemplateBinding
func PbAppTemplateBinding(s *table.AppTemplateBinding) *AppTemplateBinding {
	if s == nil {
		return nil
	}

	return &AppTemplateBinding{
		Id:         s.ID,
		Spec:       PbAppTemplateBindingSpec(s.Spec),
		Attachment: PbAppTemplateBindingAttachment(s.Attachment),
		Revision:   pbbase.PbRevision(s.Revision),
	}
}
