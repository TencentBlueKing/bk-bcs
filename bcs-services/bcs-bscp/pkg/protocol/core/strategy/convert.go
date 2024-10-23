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

// Package pbstrategy provides pbstrategy core protocol struct and convert functions.
package pbstrategy

import (
	"time"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	group "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/group"
)

// StrategySpec convert pb StrategySpec to table StrategySpec
func (m *StrategySpec) StrategySpec() *table.StrategySpec {
	if m == nil {
		return nil
	}

	return &table.StrategySpec{
		Name:             m.Name,
		ReleaseID:        m.ReleaseId,
		AsDefault:        m.AsDefault,
		Scope:            m.Scope.Scope(),
		Namespace:        m.Namespace,
		Memo:             m.Memo,
		PublishType:      table.PublishType(m.PublishType),
		PublishTime:      m.PublishTime,
		PublishStatus:    table.PublishStatus(m.PublishStatus),
		RejectReason:     m.RejectReason,
		Approver:         m.Approver,
		ApproverProgress: m.ApproverProgress,
	}
}

// PbStrategySpec convert table StrategySpec to pb StrategySpec
// nolint revive
func PbStrategySpec(s *table.StrategySpec) *StrategySpec {
	if s == nil {
		return nil
	}

	return &StrategySpec{
		Name:             s.Name,
		ReleaseId:        s.ReleaseID,
		AsDefault:        s.AsDefault,
		Scope:            PbScope(s.Scope),
		Namespace:        s.Namespace,
		PublishType:      string(s.PublishType),
		PublishTime:      s.PublishTime,
		PublishStatus:    string(s.PublishStatus),
		RejectReason:     s.RejectReason,
		Approver:         s.Approver,
		ApproverProgress: s.ApproverProgress,
		Memo:             s.Memo,
	}
}

// Scope convert pb Scope to table Scope
func (s *Scope) Scope() *table.Scope {
	if s == nil {
		return nil
	}

	group := []*table.Group{}

	for _, v := range s.Groups {
		if vv, err := v.Group(); err == nil {
			group = append(group, vv)
		}
	}

	return &table.Scope{
		Groups: group,
	}
}

// PbScope convert table Scope to pb Scope
func PbScope(s *table.Scope) *Scope {
	if s == nil {
		return nil
	}

	scode := &Scope{}
	groups, err := group.PbGroups(s.Groups)
	if err != nil {
		return scode
	}
	scode.Groups = groups
	return scode
}

// PbStrategyState convert table StrategyState to pb StrategyState
// nolint revive
func PbStrategyState(s *table.StrategyState) *StrategyState {
	if s == nil {
		return nil
	}

	return &StrategyState{
		PubState: string(s.PubState),
	}
}

// PbStrategyAttachment convert table StrategyAttachment to pb StrategyAttachment
// nolint revive
func PbStrategyAttachment(s *table.StrategyAttachment) *StrategyAttachment {
	if s == nil {
		return nil
	}

	return &StrategyAttachment{
		BizId:         s.BizID,
		AppId:         s.AppID,
		StrategySetId: s.StrategySetID,
	}
}

// PbRevision convert table Revision to pb Revision
// nolint revive
func PbRevision(s *table.Revision) *Revision {
	if s == nil {
		return nil
	}

	return &Revision{
		Creator:   s.Creator,
		Reviser:   s.Reviser,
		CreatedAt: s.CreatedAt.Format(time.DateTime),
		UpdatedAt: s.CreatedAt.Format(time.DateTime),
	}
}
