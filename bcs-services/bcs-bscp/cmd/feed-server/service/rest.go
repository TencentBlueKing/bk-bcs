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
	"net/http"

	"github.com/go-chi/render"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/cmd/feed-server/bll/types"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/errf"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/iam/meta"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
)

// ListFileAppLatestReleaseMetaRest list an app's latest release metadata only when the app's configures is file type.
func (s *Service) ListFileAppLatestReleaseMetaRest(r *http.Request) (interface{}, error) {
	kt := kit.MustGetKit(r.Context())
	opt := new(types.ListFileAppLatestReleaseMetaReq)
	if err := render.Bind(r, opt); err != nil {
		return nil, err
	}

	res := &meta.ResourceAttribute{Basic: meta.Basic{Type: meta.Release, Action: meta.Find}, BizID: opt.BizId}
	authorized, err := s.bll.Auth().Authorize(kt, res)
	if err != nil {
		return nil, err
	}

	if !authorized {
		return nil, errf.ErrPermissionDenied
	}

	meta := &types.AppInstanceMeta{
		BizID:     opt.BizId,
		AppID:     opt.AppId,
		Namespace: opt.Namespace,
		Uid:       opt.Uid,
		Labels:    opt.Labels,
	}

	cancel := kt.CtxWithTimeoutMS(1500)
	defer cancel()

	metas, err := s.bll.Release().ListAppLatestReleaseMeta(kt, meta)
	if err != nil {
		return nil, err
	}

	return metas, nil
}
