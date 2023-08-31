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

	"bscp.io/pkg/iam/meta"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	pbcs "bscp.io/pkg/protocol/config-server"
	pbds "bscp.io/pkg/protocol/data-service"
)

// UpdateConfigHook update a ConfigHook
func (s *Service) UpdateConfigHook(ctx context.Context, req *pbcs.UpdateConfigHookReq) (*pbcs.UpdateConfigHookResp, error) {

	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.UpdateConfigHookResp)

	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Release, Action: meta.Update}, BizID: grpcKit.BizID}
	if err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res); err != nil {
		return nil, err
	}

	r := &pbds.UpdateConfigHookReq{
		BizId:      req.BizId,
		AppId:      req.AppId,
		PreHookId:  req.PreHookId,
		PostHookId: req.PostHookId,
	}
	if _, e := s.client.DS.UpdateConfigHook(grpcKit.RpcCtx(), r); e != nil {
		logs.Errorf("update ConfigHook failed, err: %v, rid: %s", e, grpcKit.Rid)
		return nil, e
	}

	return resp, nil
}
