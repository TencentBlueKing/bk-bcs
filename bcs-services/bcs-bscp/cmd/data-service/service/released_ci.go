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

package service

import (
	"context"

	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	pbrci "bscp.io/pkg/protocol/core/released-ci"
	pbds "bscp.io/pkg/protocol/data-service"
)

// GetReleasedConfigItem get released config item
func (s *Service) GetReleasedConfigItem(ctx context.Context, req *pbds.GetReleasedCIReq) (
	*pbrci.ReleasedConfigItem, error) {

	kt := kit.FromGrpcContext(ctx)

	releasedCI, err := s.dao.ReleasedCI().Get(kt, req.ConfigItemId, req.BizId, req.ReleaseId)
	if err != nil {
		logs.Errorf("get released config item failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return pbrci.PbReleasedConfigItem(releasedCI), nil
}
