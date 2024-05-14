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

// Package pbgroup provides group core protocol struct and convert functions.
package pbgroup

import (
	pbstruct "github.com/golang/protobuf/ptypes/struct"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	pbbase "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/base"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/runtime/selector"
)

// Group convert pb Group to table Group
func (m *Group) Group() (*table.Group, error) {
	if m == nil {
		return nil, nil
	}

	spec, err := m.Spec.GroupSpec()
	if err != nil {
		return nil, err
	}

	return &table.Group{
		ID:         m.Id,
		Spec:       spec,
		Attachment: m.Attachment.GroupAttachment(),
	}, nil
}

// GroupSpec convert pb GroupSpec to table GroupSpec
func (m *GroupSpec) GroupSpec() (*table.GroupSpec, error) {
	if m == nil {
		return nil, nil
	}

	selector := new(selector.Selector)
	if m.Selector != nil {
		s, err := UnmarshalSelector(m.Selector)
		if err != nil {
			return nil, err
		}
		selector = s
	}

	return &table.GroupSpec{
		Name:     m.Name,
		Public:   m.Public,
		Mode:     table.GroupMode(m.Mode),
		Selector: selector,
		UID:      m.Uid,
	}, nil
}

// PbGroupSpec convert table GroupSpec to pb GroupSpec
func PbGroupSpec(spec *table.GroupSpec) (*GroupSpec, error) { //nolint:revive
	if spec == nil {
		return nil, nil
	}

	sel := new(pbstruct.Struct)
	if spec.Selector != nil {
		s, err := spec.Selector.MarshalPB()
		if err != nil {
			return nil, err
		}
		sel = s
	}

	return &GroupSpec{
		Name:     spec.Name,
		Public:   spec.Public,
		Mode:     string(spec.Mode),
		Selector: sel,
		Uid:      spec.UID,
	}, nil
}

// GroupAttachment convert pb GroupAttachment to table GroupAttachment
func (m *GroupAttachment) GroupAttachment() *table.GroupAttachment {
	if m == nil {
		return nil
	}

	return &table.GroupAttachment{
		BizID: m.BizId,
	}
}

// PbGroupAttachment convert table GroupAttachment to pb GroupAttachment
func PbGroupAttachment(at *table.GroupAttachment) *GroupAttachment { //nolint:revive
	if at == nil {
		return nil
	}

	return &GroupAttachment{
		BizId: at.BizID,
	}
}

// PbGroups convert table Group to pb Group
func PbGroups(s []*table.Group) ([]*Group, error) {
	if s == nil {
		return make([]*Group, 0), nil
	}

	result := make([]*Group, 0)
	for _, one := range s {
		group, err := PbGroup(one)
		if err != nil {
			return nil, err
		}
		result = append(result, group)
	}

	return result, nil
}

// PbGroup convert table Group to pb Group
func PbGroup(s *table.Group) (*Group, error) {
	if s == nil {
		return nil, nil
	}

	spec, err := PbGroupSpec(s.Spec)
	if err != nil {
		return nil, err
	}

	return &Group{
		Id:         s.ID,
		Spec:       spec,
		Attachment: PbGroupAttachment(s.Attachment),
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

// UnmarshalElement unmarshal pb struct to element.
func UnmarshalElement(pb *pbstruct.Struct) (*selector.Element, error) {
	json, err := pb.MarshalJSON()
	if err != nil {
		return nil, err
	}

	s := new(selector.Element)
	if err = s.UnmarshalJSON(json); err != nil {
		return nil, err
	}

	return s, nil
}
