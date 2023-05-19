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

package pbts

import (
	pbstruct "github.com/golang/protobuf/ptypes/struct"

	"bscp.io/pkg/dal/table"
	pbbase "bscp.io/pkg/protocol/core/base"
	"bscp.io/pkg/runtime/selector"
)

// TemplateSpace convert pb TemplateSpace to table TemplateSpace
func (m *TemplateSpace) TemplateSpace() (*table.TemplateSpace, error) {
	if m == nil {
		return nil, nil
	}

	spec, err := m.Spec.TemplateSpaceSpec()
	if err != nil {
		return nil, err
	}

	return &table.TemplateSpace{
		ID:         m.Id,
		Spec:       spec,
		Attachment: m.Attachment.TemplateSpaceAttachment(),
	}, nil
}

// TemplateSpaceSpec convert pb TemplateSpaceSpec to table TemplateSpaceSpec
func (m *TemplateSpaceSpec) TemplateSpaceSpec() (*table.TemplateSpaceSpec, error) {
	if m == nil {
		return nil, nil
	}

	return &table.TemplateSpaceSpec{
		Name: m.Name,
		Memo: m.Memo,
	}, nil
}

// PbTemplateSpaceSpec convert table TemplateSpaceSpec to pb TemplateSpaceSpec
func PbTemplateSpaceSpec(spec *table.TemplateSpaceSpec) (*TemplateSpaceSpec, error) {
	if spec == nil {
		return nil, nil
	}

	return &TemplateSpaceSpec{
		Name: spec.Name,
		Memo: spec.Memo,
	}, nil
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
func PbTemplateSpaces(s []*table.TemplateSpace) ([]*TemplateSpace, error) {
	if s == nil {
		return make([]*TemplateSpace, 0), nil
	}

	result := make([]*TemplateSpace, 0)
	for _, one := range s {
		hook, err := PbTemplateSpace(one)
		if err != nil {
			return nil, err
		}
		result = append(result, hook)
	}

	return result, nil
}

// PbTemplateSpace convert table TemplateSpace to pb TemplateSpace
func PbTemplateSpace(s *table.TemplateSpace) (*TemplateSpace, error) {
	if s == nil {
		return nil, nil
	}

	spec, err := PbTemplateSpaceSpec(s.Spec)
	if err != nil {
		return nil, err
	}

	return &TemplateSpace{
		Id:         s.ID,
		Spec:       spec,
		Attachment: PbTemplateSpaceAttachment(s.Attachment),
		Revision:   pbbase.PbRevision(s.Revision),
	}, nil
}

// UnmarshalSelector unmarshal pb struct to selector.
func UnmarshalSelector(pb *pbstruct.Struct) (*selector.Selector, error) {
	json, err := pb.MarshalJSON()
	if err != nil {
		return nil, err
	}

	s := new(selector.Selector)
	if err = s.Unmarshal(json); err != nil {
		return nil, err
	}

	return s, nil
}
