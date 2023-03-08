/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "as IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package service

import (
	"context"
	"strings"

	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	pbcs "bscp.io/pkg/protocol/config-server"
	"bscp.io/pkg/thirdparty/esb/cmdb"
)

// ListBiz list business
func (s *Service) ListBiz(ctx context.Context, req *pbcs.ListBizReq) (*pbcs.ListBizResp, error) {
	kt := kit.FromGrpcContext(ctx)
	resp := new(pbcs.ListBizResp)

	params := &cmdb.SearchBizParams{
		Fields: []string{},
		Condition: map[string]string{
			"bk_biz_maintainer": kt.User,
		},
	}
	sbresp, err := s.client.Esb.Cmdb().SearchBusiness(ctx, params)
	if err != nil {
		logs.Errorf("search business failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	bizList := []*pbcs.ListBizResp_BizData{}
	for _, biz := range sbresp.Info {
		bizList = append(bizList, &pbcs.ListBizResp_BizData{
			BizId:         uint32(biz.BizID),
			BizName:       biz.BizName,
			BizMaintainer: strings.Split(biz.BizMaintainer, ","),
		})
	}

	resp = &pbcs.ListBizResp{
		BizList: bizList,
	}
	return resp, nil
}
