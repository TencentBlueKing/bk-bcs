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

package pbstrategy

import (
	"errors"

	"bscp.io/pkg/dal/table"
	pbbase "bscp.io/pkg/protocol/core/base"
	"bscp.io/pkg/runtime/selector"

	pbstruct "github.com/golang/protobuf/ptypes/struct"
)

// StrategySpec convert pb StrategySpec to table StrategySpec
func (m *StrategySpec) StrategySpec() (*table.StrategySpec, error) {
	if m == nil {
		return nil, nil
	}

	scope := new(table.ScopeSelector)
	var err error
	if m.Scope != nil {
		scope, err = m.Scope.ScopeSelector()
		if err != nil {
			return nil, err
		}
	}

	return &table.StrategySpec{
		Name:      m.Name,
		ReleaseID: m.ReleaseId,
		AsDefault: m.AsDefault,
		Scope:     scope,
		Mode:      table.AppMode(m.Mode),
		Namespace: m.Namespace,
		Memo:      m.Memo,
	}, nil
}

// PbStrategySpec convert table StrategySpec to pb StrategySpec
func PbStrategySpec(spec *table.StrategySpec) (*StrategySpec, error) {
	if spec == nil {
		return nil, nil
	}

	scope, err := PbScopeSelector(spec.Scope)
	if err != nil {
		return nil, err
	}

	return &StrategySpec{
		Name:      spec.Name,
		ReleaseId: spec.ReleaseID,
		AsDefault: spec.AsDefault,
		Scope:     scope,
		Mode:      string(spec.Mode),
		Namespace: spec.Namespace,
		Memo:      spec.Memo,
	}, nil
}

// StrategyState convert pb StrategyState to table StrategyState
func (m *StrategyState) StrategyState() *table.StrategyState {
	if m == nil {
		return nil
	}

	return &table.StrategyState{
		PubState: table.PublishState(m.PubState),
	}
}

// PbStrategyState convert table StrategyState to pb StrategyState
func PbStrategyState(ss *table.StrategyState) *StrategyState {
	if ss == nil {
		return nil
	}

	return &StrategyState{
		PubState: string(ss.PubState),
	}
}

// StrategyAttachment convert pb StrategyAttachment to table StrategyAttachment
func (m *StrategyAttachment) StrategyAttachment() *table.StrategyAttachment {
	if m == nil {
		return nil
	}

	return &table.StrategyAttachment{
		BizID:         m.BizId,
		AppID:         m.AppId,
		StrategySetID: m.StrategySetId,
	}
}

// PbStrategyAttachment convert table StrategyAttachment to pb StrategyAttachment
func PbStrategyAttachment(at *table.StrategyAttachment) *StrategyAttachment {
	if at == nil {
		return nil
	}

	return &StrategyAttachment{
		BizId:         at.BizID,
		AppId:         at.AppID,
		StrategySetId: at.StrategySetID,
	}
}

// ScopeSelector convert pb ScopeSelector to table ScopeSelector
func (m *ScopeSelector) ScopeSelector() (*table.ScopeSelector, error) {
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

	subStrategy, err := m.SubStrategy.SubStrategy()
	if err != nil {
		return nil, err
	}

	return &table.ScopeSelector{
		Selector:    selector,
		SubStrategy: subStrategy,
	}, nil
}

// PbScopeSelector convert table ScopeSelector to pb ScopeSelector
func PbScopeSelector(s *table.ScopeSelector) (*ScopeSelector, error) {
	if s == nil {
		return nil, nil
	}

	pbSelector, err := s.Selector.MarshalPB()
	if err != nil {
		return nil, err
	}

	subStrategy, err := PbSubStrategy(s.SubStrategy)
	if err != nil {
		return nil, err
	}

	return &ScopeSelector{
		Selector:    pbSelector,
		SubStrategy: subStrategy,
	}, nil
}

// SubStrategy convert pb SubStrategy to table SubStrategy
func (m *SubStrategy) SubStrategy() (*table.SubStrategy, error) {
	if m == nil {
		return nil, nil
	}

	spec, err := m.Spec.SubStrategySpec()
	if err != nil {
		return nil, err
	}

	return &table.SubStrategy{
		Spec: spec,
	}, nil
}

// PbSubStrategy convert table SubStrategy to pb SubStrategy
func PbSubStrategy(s *table.SubStrategy) (*SubStrategy, error) {
	if s == nil {
		return nil, nil
	}

	spec, err := PbSubStrategySpec(s.Spec)
	if err != nil {
		return nil, err
	}

	return &SubStrategy{
		Spec: spec,
	}, nil
}

// SubStrategySpec convert pb SubStrategySpec to table SubStrategySpec
func (m *SubStrategySpec) SubStrategySpec() (*table.SubStrategySpec, error) {
	if m == nil {
		return nil, nil
	}

	scope, err := m.Scope.SubScopeSelector()
	if err != nil {
		return nil, err
	}

	return &table.SubStrategySpec{
		Name:      m.Name,
		ReleaseID: m.ReleaseId,
		Scope:     scope,
		Memo:      m.Memo,
	}, nil
}

// PbSubStrategySpec convert table SubStrategySpec to pb SubStrategySpec
func PbSubStrategySpec(spec *table.SubStrategySpec) (*SubStrategySpec, error) {
	if spec == nil {
		return nil, nil
	}

	scope, err := PbSubScopeSelector(spec.Scope)
	if err != nil {
		return nil, err
	}
	return &SubStrategySpec{
		Name:      spec.Name,
		ReleaseId: spec.ReleaseID,
		Scope:     scope,
		Memo:      spec.Memo,
	}, nil
}

// SubScopeSelector convert pb SubScopeSelector to table SubScopeSelector
func (m *SubScopeSelector) SubScopeSelector() (*table.SubScopeSelector, error) {
	if m == nil {
		return nil, nil
	}

	strategicSelector := new(selector.Selector)
	if m.Selector != nil {
		s, err := UnmarshalSelector(m.Selector)
		if err != nil {
			return nil, err
		}
		strategicSelector = s
	}

	return &table.SubScopeSelector{
		Selector: strategicSelector,
	}, nil
}

// PbSubScopeSelector convert table SubScopeSelector to pb SubScopeSelector
func PbSubScopeSelector(s *table.SubScopeSelector) (*SubScopeSelector, error) {
	if s == nil {
		return nil, nil
	}

	pbSelector, err := s.Selector.MarshalPB()
	if err != nil {
		return nil, err
	}

	return &SubScopeSelector{
		Selector: pbSelector,
	}, nil
}

// PbStrategies convert table Strategy to pb Strategy
func PbStrategies(ss []*table.Strategy) ([]*Strategy, error) {
	if ss == nil {
		return make([]*Strategy, 0), nil
	}

	result := make([]*Strategy, 0)
	for _, s := range ss {
		strategy, err := PbStrategy(s)
		if err != nil {
			return nil, err
		}
		result = append(result, strategy)
	}

	return result, nil
}

// PbStrategy convert table Strategy to pb Strategy
func PbStrategy(s *table.Strategy) (*Strategy, error) {
	if s == nil {
		return nil, nil
	}

	spec, err := PbStrategySpec(s.Spec)
	if err != nil {
		return nil, err
	}

	return &Strategy{
		Id:         s.ID,
		Spec:       spec,
		State:      PbStrategyState(s.State),
		Attachment: PbStrategyAttachment(s.Attachment),
		Revision:   pbbase.PbRevision(s.Revision),
	}, nil
}

// PbCPStrategies convert table CurrentPublishedStrategy to pb CurrentPublishedStrategy
func PbCPStrategies(cps []*table.CurrentPublishedStrategy) ([]*CurrentPublishedStrategy, error) {
	if cps == nil {
		return make([]*CurrentPublishedStrategy, 0), nil
	}

	result := make([]*CurrentPublishedStrategy, 0)
	for _, s := range cps {
		if s == nil {
			return nil, errors.New("CPS strategy is nil, can not be converted")
		}

		strategy, err := PbCPStrategy(s)
		if err != nil {
			return nil, err
		}
		result = append(result, strategy)
	}

	return result, nil
}

// PbCPStrategy convert table CurrentPublishedStrategy to pb CurrentPublishedStrategy
func PbCPStrategy(cps *table.CurrentPublishedStrategy) (*CurrentPublishedStrategy, error) {
	if cps == nil {
		return nil, nil
	}

	spec, err := PbStrategySpec(cps.Spec)
	if err != nil {
		return nil, err
	}

	return &CurrentPublishedStrategy{
		Id:         cps.ID,
		StrategyId: cps.StrategyID,
		Spec:       spec,
		State:      PbStrategyState(cps.State),
		Attachment: PbStrategyAttachment(cps.Attachment),
		Revision:   pbbase.PbCreatedRevision(cps.Revision),
	}, nil
}

// PbPubStrategyHistories convert table PublishedStrategyHistory to pb PublishedStrategyHistory
func PbPubStrategyHistories(psh []*table.PublishedStrategyHistory) ([]*PublishedStrategyHistory, error) {
	if psh == nil {
		return make([]*PublishedStrategyHistory, 0), nil
	}

	result := make([]*PublishedStrategyHistory, 0)
	for _, s := range psh {
		if s == nil {
			return nil, errors.New("PSH is nil, can not be converted")
		}

		history, err := PubStrategyHistory(s)
		if err != nil {
			return nil, err
		}
		result = append(result, history)
	}

	return result, nil
}

// PubStrategyHistory convert table PublishedStrategyHistory to pb PublishedStrategyHistory
func PubStrategyHistory(psh *table.PublishedStrategyHistory) (*PublishedStrategyHistory, error) {
	if psh == nil {
		return nil, nil
	}

	spec, err := PbStrategySpec(psh.Spec)
	if err != nil {
		return nil, err
	}

	return &PublishedStrategyHistory{
		Id:         psh.ID,
		StrategyId: psh.StrategyID,
		Spec:       spec,
		State:      PbStrategyState(psh.State),
		Attachment: PbStrategyAttachment(psh.Attachment),
		Revision:   pbbase.PbCreatedRevision(psh.Revision),
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
