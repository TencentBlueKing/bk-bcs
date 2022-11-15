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

package pbss

import (
	"bscp.io/pkg/dal/table"
	pbbase "bscp.io/pkg/protocol/core/base"
)

// StrategySetSpec convert pb StrategySetSpec to table StrategySetSpec
func (m *StrategySetSpec) StrategySetSpec() *table.StrategySetSpec {
	if m == nil {
		return nil
	}

	return &table.StrategySetSpec{
		Name: m.Name,
		Memo: m.Memo,
		Mode: table.AppMode(m.Mode),
	}
}

// PbStrategySetSpec convert table StrategySetSpec to pb StrategySetSpec
func PbStrategySetSpec(spec *table.StrategySetSpec) *StrategySetSpec {
	if spec == nil {
		return nil
	}

	return &StrategySetSpec{
		Name: spec.Name,
		Memo: spec.Memo,
		Mode: string(spec.Mode),
	}
}

// StrategySetState convert pb StrategySetState to table StrategySetState
func (m *StrategySetState) StrategySetState() *table.StrategySetState {
	if m == nil {
		return nil
	}

	return &table.StrategySetState{
		Status: table.StrategySetStatusType(m.Status),
	}
}

// PbStrategySetState convert table StrategySetState to pb StrategySetState
func PbStrategySetState(s *table.StrategySetState) *StrategySetState {
	if s == nil {
		return nil
	}

	return &StrategySetState{
		Status: string(s.Status),
	}
}

// StrategySetAttachment convert pb StrategySetAttachment to table StrategySetAttachment
func (m *StrategySetAttachment) StrategySetAttachment() *table.StrategySetAttachment {
	if m == nil {
		return nil
	}

	return &table.StrategySetAttachment{
		BizID: m.BizId,
		AppID: m.AppId,
	}
}

// PbStrategySetAttachment convert table StrategySetAttachment to pb StrategySetAttachment
func PbStrategySetAttachment(at *table.StrategySetAttachment) *StrategySetAttachment {
	if at == nil {
		return nil
	}

	return &StrategySetAttachment{
		BizId: at.BizID,
		AppId: at.AppID,
	}
}

// PbStrategySets convert table StrategySet to pb StrategySet
func PbStrategySets(s []*table.StrategySet) []*StrategySet {
	if s == nil {
		return make([]*StrategySet, 0)
	}

	result := make([]*StrategySet, 0)
	for _, one := range s {
		result = append(result, PbStrategySet(one))
	}

	return result
}

// PbStrategySet convert table StrategySet to pb StrategySet
func PbStrategySet(s *table.StrategySet) *StrategySet {
	if s == nil {
		return nil
	}

	return &StrategySet{
		Id:         s.ID,
		Spec:       PbStrategySetSpec(s.Spec),
		State:      PbStrategySetState(s.State),
		Attachment: PbStrategySetAttachment(s.Attachment),
		Revision:   pbbase.PbRevision(s.Revision),
	}
}
