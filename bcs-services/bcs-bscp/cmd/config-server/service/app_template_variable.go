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

// ExtractTemplateVariables extract template variables
func (s *Service) ExtractTemplateVariables(ctx context.Context, req *pbcs.ExtractTemplateVariablesReq) (
	*pbcs.ExtractTemplateVariablesResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.ExtractTemplateVariablesResp)

	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.TemplateVariable, Action: meta.Find},
		BizID: req.BizId}
	if err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res); err != nil {
		return nil, err
	}

	r := &pbds.ExtractTemplateVariablesReq{
		BizId: req.BizId,
		AppId: req.AppId,
	}

	rp, err := s.client.DS.ExtractTemplateVariables(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("list template variables failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp = &pbcs.ExtractTemplateVariablesResp{
		Details: rp.Details,
	}
	return resp, nil
}
