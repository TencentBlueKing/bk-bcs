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

// Package pbtv provides template variable core protocol struct and convert functions.
package pbtv

import (
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	pbbase "github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/base"
)

// TemplateVariable convert pb TemplateVariable to table TemplateVariable
func (m *TemplateVariable) TemplateVariable() (*table.TemplateVariable, error) {
	if m == nil {
		return nil, nil
	}

	return &table.TemplateVariable{
		ID:         m.Id,
		Spec:       m.Spec.TemplateVariableSpec(),
		Attachment: m.Attachment.TemplateVariableAttachment(),
	}, nil
}

// TemplateVariableSpec convert pb TemplateVariableSpec to table TemplateVariableSpec
func (m *TemplateVariableSpec) TemplateVariableSpec() *table.TemplateVariableSpec {
	if m == nil {
		return nil
	}

	return &table.TemplateVariableSpec{
		Name:       m.Name,
		Type:       table.VariableType(m.Type),
		DefaultVal: m.DefaultVal,
		Memo:       m.Memo,
	}
}

// PbTemplateVariableSpecs convert table TemplateVariableSpec to pb TemplateVariableSpec
func PbTemplateVariableSpecs(s []*table.TemplateVariableSpec) []*TemplateVariableSpec {
	if len(s) == 0 {
		return make([]*TemplateVariableSpec, 0)
	}

	result := make([]*TemplateVariableSpec, 0)
	for _, one := range s {
		result = append(result, PbTemplateVariableSpec(one))
	}

	return result
}

// PbTemplateVariableSpec convert table TemplateVariableSpec to pb TemplateVariableSpec
func PbTemplateVariableSpec(spec *table.TemplateVariableSpec) *TemplateVariableSpec {
	if spec == nil {
		return nil
	}

	return &TemplateVariableSpec{
		Name:       spec.Name,
		Type:       string(spec.Type),
		DefaultVal: spec.DefaultVal,
		Memo:       spec.Memo,
	}
}

// TemplateVariableAttachment convert pb TemplateVariableAttachment to table TemplateVariableAttachment
func (m *TemplateVariableAttachment) TemplateVariableAttachment() *table.TemplateVariableAttachment {
	if m == nil {
		return nil
	}

	return &table.TemplateVariableAttachment{
		BizID: m.BizId,
	}
}

// PbTemplateVariableAttachment convert table TemplateVariableAttachment to pb TemplateVariableAttachment
func PbTemplateVariableAttachment(at *table.TemplateVariableAttachment) *TemplateVariableAttachment {
	if at == nil {
		return nil
	}

	return &TemplateVariableAttachment{
		BizId: at.BizID,
	}
}

// PbTemplateVariables convert table TemplateVariable to pb TemplateVariable
func PbTemplateVariables(s []*table.TemplateVariable) []*TemplateVariable {
	if s == nil {
		return make([]*TemplateVariable, 0)
	}

	result := make([]*TemplateVariable, 0)
	for _, one := range s {
		result = append(result, PbTemplateVariable(one))
	}

	return result
}

// PbTemplateVariable convert table TemplateVariable to pb TemplateVariable
func PbTemplateVariable(s *table.TemplateVariable) *TemplateVariable {
	if s == nil {
		return nil
	}

	return &TemplateVariable{
		Id:         s.ID,
		Spec:       PbTemplateVariableSpec(s.Spec),
		Attachment: PbTemplateVariableAttachment(s.Attachment),
		Revision:   pbbase.PbRevision(s.Revision),
	}
}
