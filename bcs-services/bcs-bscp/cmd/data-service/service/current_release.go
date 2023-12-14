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

package service

import (
	"context"

	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	pbds "github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/data-service"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/types"
)

// CountGroupsReleasedApps count each group's published apps.
func (s *Service) CountGroupsReleasedApps(ctx context.Context, req *pbds.CountGroupsReleasedAppsReq) (
	*pbds.CountGroupsReleasedAppsResp, error) {
	kt := kit.FromGrpcContext(ctx)

	counts, err := s.dao.ReleasedGroup().CountGroupsReleasedApps(kt, &types.CountGroupsReleasedAppsOption{
		BizID:  req.BizId,
		Groups: req.Groups,
	})
	if err != nil {
		logs.Errorf("count groups published apps failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	data := make([]*pbds.CountGroupsReleasedAppsResp_CountGroupsReleasedAppsData, len(counts))
	for i, count := range counts {
		data[i] = &pbds.CountGroupsReleasedAppsResp_CountGroupsReleasedAppsData{
			GroupId: count.GroupID,
			Count:   count.Counts,
			Edited:  count.Edited,
		}
	}
	resp := &pbds.CountGroupsReleasedAppsResp{
		Data: data,
	}

	return resp, nil
}
