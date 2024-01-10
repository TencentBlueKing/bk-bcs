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

// Package pbatv provides app template variable core protocol struct and convert functions.
package pbatv

import (
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	pbbase "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/base"
	pbtv "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/template-variable"
)

// AppTemplateVariable convert pb AppTemplateVariable to table AppTemplateVariable
func (m *AppTemplateVariable) AppTemplateVariable() (*table.AppTemplateVariable, error) {
	if m == nil {
		return nil, nil
	}

	return &table.AppTemplateVariable{
		ID:         m.Id,
		Spec:       m.Spec.AppTemplateVariableSpec(),
		Attachment: m.Attachment.AppTemplateVariableAttachment(),
	}, nil
}

// AppTemplateVariableSpec convert pb AppTemplateVariableSpec to table AppTemplateVariableSpec
func (m *AppTemplateVariableSpec) AppTemplateVariableSpec() *table.AppTemplateVariableSpec {
	if m == nil {
		return nil
	}

	result := make([]*table.TemplateVariableSpec, 0)
	for _, one := range m.Variables {
		result = append(result, one.TemplateVariableSpec())
	}

	return &table.AppTemplateVariableSpec{
		Variables: result,
	}
}

// PbAppTemplateVariableSpec convert table AppTemplateVariableSpec to pb AppTemplateVariableSpec
func PbAppTemplateVariableSpec(spec *table.AppTemplateVariableSpec) *AppTemplateVariableSpec {
	if spec == nil {
		return nil
	}

	result := make([]*pbtv.TemplateVariableSpec, 0)
	for _, one := range spec.Variables {
		result = append(result, pbtv.PbTemplateVariableSpec(one))
	}

	return &AppTemplateVariableSpec{
		Variables: result,
	}
}

// AppTemplateVariableAttachment convert pb AppTemplateVariableAttachment to table AppTemplateVariableAttachment
func (m *AppTemplateVariableAttachment) AppTemplateVariableAttachment() *table.AppTemplateVariableAttachment {
	if m == nil {
		return nil
	}

	return &table.AppTemplateVariableAttachment{
		BizID: m.BizId,
		AppID: m.AppId,
	}
}

// PbAppTemplateVariableAttachment convert table AppTemplateVariableAttachment to pb AppTemplateVariableAttachment
func PbAppTemplateVariableAttachment(at *table.AppTemplateVariableAttachment) *AppTemplateVariableAttachment {
	if at == nil {
		return nil
	}

	return &AppTemplateVariableAttachment{
		BizId: at.BizID,
		AppId: at.AppID,
	}
}

// PbAppTemplateVariables convert table AppTemplateVariable to pb AppTemplateVariable
func PbAppTemplateVariables(s []*table.AppTemplateVariable) []*AppTemplateVariable {
	if s == nil {
		return make([]*AppTemplateVariable, 0)
	}

	result := make([]*AppTemplateVariable, 0)
	for _, one := range s {
		result = append(result, PbAppTemplateVariable(one))
	}

	return result
}

// PbAppTemplateVariable convert table AppTemplateVariable to pb AppTemplateVariable
func PbAppTemplateVariable(s *table.AppTemplateVariable) *AppTemplateVariable {
	if s == nil {
		return nil
	}

	return &AppTemplateVariable{
		Id:         s.ID,
		Spec:       PbAppTemplateVariableSpec(s.Spec),
		Attachment: PbAppTemplateVariableAttachment(s.Attachment),
		Revision:   pbbase.PbRevision(s.Revision),
	}
}
