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

package pbhook

import (
	"bscp.io/pkg/dal/table"
	pbbase "bscp.io/pkg/protocol/core/base"
	"bscp.io/pkg/runtime/selector"
	pbstruct "github.com/golang/protobuf/ptypes/struct"
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
		Type: table.HookType(m.Type),
		Tag:  m.Tag,
		Memo: m.Memo,
	}, nil
}

// PbHookSpec convert table HookSpec to pb HookSpec
func PbHookSpec(spec *table.HookSpec) (*HookSpec, error) {
	if spec == nil {
		return nil, nil
	}

	return &HookSpec{
		Name:        spec.Name,
		ReleaseName: spec.Name,
		Type:        string(spec.Type),
		Tag:         spec.Tag,
		Memo:        spec.Memo,
		PublishNum:  spec.PublishNum,
	}, nil
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
func PbHookAttachment(at *table.HookAttachment) *HookAttachment {
	if at == nil {
		return nil
	}

	return &HookAttachment{
		BizId: at.BizID,
	}
}

// PbHooks convert table Hook to pb Hook
func PbHooks(s []*table.Hook) ([]*Hook, error) {
	if s == nil {
		return make([]*Hook, 0), nil
	}

	result := make([]*Hook, 0)
	for _, one := range s {
		hook, err := PbHook(one)
		if err != nil {
			return nil, err
		}
		result = append(result, hook)
	}

	return result, nil
}

// PbHook convert table Hook to pb Hook
func PbHook(s *table.Hook) (*Hook, error) {
	if s == nil {
		return nil, nil
	}

	spec, err := PbHookSpec(s.Spec)
	if err != nil {
		return nil, err
	}

	return &Hook{
		Id:         s.ID,
		Spec:       spec,
		Attachment: PbHookAttachment(s.Attachment),
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
