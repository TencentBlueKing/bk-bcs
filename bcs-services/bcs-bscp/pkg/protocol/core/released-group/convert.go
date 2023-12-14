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

// Package pbrg provides released group core protocol struct and convert functions.
package pbrg

import (
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/types"
)

// ReleasedGroup convert pb ReleasedGroup to table ReleasedGroup
func (m *ReleasedGroup) ReleasedGroup() (*table.ReleasedGroup, error) {
	if m == nil {
		return nil, nil
	}

	return &table.ReleasedGroup{
		ID:         m.Id,
		GroupID:    m.GroupId,
		AppID:      m.AppId,
		ReleaseID:  m.ReleaseId,
		StrategyID: m.StrategyId,
		Edited:     m.Edited,
		BizID:      m.BizId,
	}, nil
}

// PbReleasedGroups convert table ReleasedGroup to pb ReleasedGroup
func PbReleasedGroups(s []*table.ReleasedGroup) ([]*ReleasedGroup, error) {
	if s == nil {
		return make([]*ReleasedGroup, 0), nil
	}

	result := make([]*ReleasedGroup, 0)
	for _, one := range s {
		gcr, err := PbReleasedGroup(one)
		if err != nil {
			return nil, err
		}
		result = append(result, gcr)
	}

	return result, nil
}

// PbReleasedGroup convert table ReleasedGroup to pb ReleasedGroup
func PbReleasedGroup(s *table.ReleasedGroup) (*ReleasedGroup, error) {
	if s == nil {
		return nil, nil
	}

	return &ReleasedGroup{
		Id:         s.ID,
		GroupId:    s.GroupID,
		AppId:      s.AppID,
		ReleaseId:  s.ReleaseID,
		StrategyId: s.StrategyID,
		Edited:     s.Edited,
		BizId:      s.BizID,
	}, nil
}

// CacheReleasedGroups convert table ReleasedGroup to ReleasedGroupCache
func CacheReleasedGroups(s []*table.ReleasedGroup) []*types.ReleasedGroupCache {
	if s == nil {
		return make([]*types.ReleasedGroupCache, 0)
	}

	result := make([]*types.ReleasedGroupCache, 0)
	for _, one := range s {
		result = append(result, CacheReleasedGroup(one))
	}

	return result
}

// CacheReleasedGroup convert table ReleasedGroup to pb ReleasedGroupCache
func CacheReleasedGroup(s *table.ReleasedGroup) *types.ReleasedGroupCache {
	if s == nil {
		return nil
	}

	return &types.ReleasedGroupCache{
		ID:         s.ID,
		GroupID:    s.GroupID,
		AppID:      s.AppID,
		ReleaseID:  s.ReleaseID,
		StrategyID: s.StrategyID,
		Mode:       s.Mode,
		Selector:   s.Selector,
		UID:        s.UID,
		BizID:      s.BizID,
		UpdatedAt:  s.UpdatedAt,
	}
}
