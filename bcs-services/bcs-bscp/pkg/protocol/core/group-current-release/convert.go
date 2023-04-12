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

package pbgcr

import (
	"bscp.io/pkg/dal/table"
)

// GroupCurrentRelease convert pb GroupCurrentRelease to table GroupCurrentRelease
func (m *GroupCurrentRelease) GroupCurrentRelease() (*table.GroupCurrentRelease, error) {
	if m == nil {
		return nil, nil
	}

	return &table.GroupCurrentRelease{
		ID:          m.Id,
		GroupID:     m.GroupId,
		AppID:       m.AppId,
		ReleaseID:   m.ReleaseId,
		ReleaseName: m.ReleaseName,
		StrategyID:  m.StrategyId,
		Edited:      m.Edited,
		BizID:       m.BizId,
	}, nil
}

// PbGroupCurrentReleases convert table GroupCurrentRelease to pb GroupCurrentRelease
func PbGroupCurrentReleases(s []*table.GroupCurrentRelease) ([]*GroupCurrentRelease, error) {
	if s == nil {
		return make([]*GroupCurrentRelease, 0), nil
	}

	result := make([]*GroupCurrentRelease, 0)
	for _, one := range s {
		gcr, err := PbGroupCurrentRelease(one)
		if err != nil {
			return nil, err
		}
		result = append(result, gcr)
	}

	return result, nil
}

// PbGroupCurrentRelease convert table GroupCurrentRelease to pb GroupCurrentRelease
func PbGroupCurrentRelease(s *table.GroupCurrentRelease) (*GroupCurrentRelease, error) {
	if s == nil {
		return nil, nil
	}

	return &GroupCurrentRelease{
		Id:          s.ID,
		GroupId:     s.GroupID,
		AppId:       s.AppID,
		ReleaseId:   s.ReleaseID,
		ReleaseName: s.ReleaseName,
		StrategyId:  s.StrategyID,
		Edited:      s.Edited,
		BizId:       s.BizID,
	}, nil
}
