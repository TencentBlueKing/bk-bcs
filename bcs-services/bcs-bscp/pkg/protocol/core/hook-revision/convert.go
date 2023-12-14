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

// Package pbhr provides hook revision core protocol struct and convert functions.
package pbhr

import (
	pbstruct "github.com/golang/protobuf/ptypes/struct"

	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	pbbase "github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/base"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/runtime/selector"
)

// HookRevisionSpace convert pb HookRevisionSpace to table HookRevisionSpace
func (m *HookRevision) HookRevisionSpace() (*table.HookRevision, error) {
	if m == nil {
		return nil, nil
	}

	spec, err := m.Spec.HookRevisionSpec()
	if err != nil {
		return nil, err
	}

	return &table.HookRevision{
		ID:         m.Id,
		Spec:       spec,
		Attachment: m.Attachment.HookRevisionAttachment(),
	}, nil
}

// HookRevisionSpec convert pb HookRevisionSpace to table HookRevisionSpace
func (m *HookRevisionSpec) HookRevisionSpec() (*table.HookRevisionSpec, error) {
	if m == nil {
		return nil, nil
	}

	return &table.HookRevisionSpec{
		Name:    m.Name,
		Content: m.Content,
		Memo:    m.Memo,
	}, nil
}

// PbHookRevisionSpec convert table HookRevisionSpec to pb HookRevisionSpec
func PbHookRevisionSpec(spec *table.HookRevisionSpec) *HookRevisionSpec {
	if spec == nil {
		return nil
	}

	return &HookRevisionSpec{
		Name:    spec.Name,
		Content: spec.Content,
		State:   spec.State.String(),
		Memo:    spec.Memo,
	}
}

// HookRevisionAttachment convert pb HookRevisionAttachment to table HookRevisionAttachment
func (m *HookRevisionAttachment) HookRevisionAttachment() *table.HookRevisionAttachment {
	if m == nil {
		return nil
	}

	return &table.HookRevisionAttachment{
		BizID:  m.BizId,
		HookID: m.HookId,
	}
}

// HookRevisionSpaceAttachment convert table HookRevisionAttachment to pb HookRevisionAttachment
func HookRevisionSpaceAttachment(at *table.HookRevisionAttachment) *HookRevisionAttachment {
	if at == nil {
		return nil
	}

	return &HookRevisionAttachment{
		BizId:  at.BizID,
		HookId: at.HookID,
	}
}

// PbHookRevisionSpaces convert table HookRevision to pb HookRevision
func PbHookRevisionSpaces(s []*table.HookRevision) []*HookRevision {
	if s == nil {
		return make([]*HookRevision, 0)
	}

	result := make([]*HookRevision, 0)
	for _, one := range s {
		result = append(result, PbHookRevision(one))
	}

	return result
}

// PbHookRevision convert table HookRevision to pb HookRevision
func PbHookRevision(s *table.HookRevision) *HookRevision {
	if s == nil {
		return nil
	}

	return &HookRevision{
		Id:         s.ID,
		Spec:       PbHookRevisionSpec(s.Spec),
		Attachment: HookRevisionSpaceAttachment(s.Attachment),
		Revision:   pbbase.PbRevision(s.Revision),
	}
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
