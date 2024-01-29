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

// Package pbhook provides hook core protocol struct and convert functions.
package pbhook

import (
	pbstruct "github.com/golang/protobuf/ptypes/struct"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	pbbase "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/base"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/runtime/selector"
)

// Hook convert pb Hook to table Hook
func (m *Hook) Hook() (*table.Hook, error) {
	if m == nil {
		return nil, nil
	}

	spec, err := m.Spec.HookSpec()
	if err != nil {
		return nil, err
	}

	return &table.Hook{
		ID:         m.Id,
		Spec:       spec,
		Attachment: m.Attachment.HookAttachment(),
	}, nil
}

// HookSpec convert pb HookSpec to table HookSpec
func (m *HookSpec) HookSpec() (*table.HookSpec, error) {
	if m == nil {
		return nil, nil
	}

	return &table.HookSpec{
		Name: m.Name,
		Type: table.ScriptType(m.Type),
		Tag:  m.Tag,
		Memo: m.Memo,
	}, nil
}

// PbHookSpec convert table HookSpec to pb HookSpec
func PbHookSpec(spec *table.HookSpec) *HookSpec { //nolint:revive
	if spec == nil {
		return nil
	}

	return &HookSpec{
		Name: spec.Name,
		Type: string(spec.Type),
		Tag:  spec.Tag,
		Memo: spec.Memo,
	}
}

// HookAttachment convert pb HookAttachment to table HookAttachment
func (m *HookAttachment) HookAttachment() *table.HookAttachment {
	if m == nil {
		return nil
	}

	return &table.HookAttachment{
		BizID: m.BizId,
	}
}

// PbHookAttachment convert table HookAttachment to pb HookAttachment
func PbHookAttachment(at *table.HookAttachment) *HookAttachment { //nolint:revive
	if at == nil {
		return nil
	}

	return &HookAttachment{
		BizId: at.BizID,
	}
}

// PbHooks convert table Hook to pb Hook
func PbHooks(s []*table.Hook) []*Hook {
	if s == nil {
		return make([]*Hook, 0)
	}

	result := make([]*Hook, 0)
	for _, one := range s {
		result = append(result, PbHook(one))
	}

	return result
}

// PbHook convert table Hook to pb Hook
func PbHook(s *table.Hook) *Hook {
	if s == nil {
		return nil
	}

	return &Hook{
		Id:         s.ID,
		Spec:       PbHookSpec(s.Spec),
		Attachment: PbHookAttachment(s.Attachment),
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
