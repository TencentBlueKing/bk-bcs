/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package release

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/release"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/store"
	helmmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/proto/bcs-helm-manager"
)

const (
	defaultSize = 1000
)

// NewListReleaseAction return a new ListReleaseAction instance
func NewListReleaseAction(model store.HelmManagerModel, releaseHandler release.Handler) *ListReleaseAction {
	return &ListReleaseAction{
		model:          model,
		releaseHandler: releaseHandler,
	}
}

// ListReleaseAction provides the action to do list chart release
type ListReleaseAction struct {
	ctx context.Context

	model          store.HelmManagerModel
	releaseHandler release.Handler

	req  *helmmanager.ListReleaseReq
	resp *helmmanager.ListReleaseResp
}

// Handle the listing process
func (l *ListReleaseAction) Handle(ctx context.Context,
	req *helmmanager.ListReleaseReq, resp *helmmanager.ListReleaseResp) error {

	if req == nil || resp == nil {
		blog.Errorf("get release failed, req or resp is empty")
		return common.ErrHelmManagerReqOrRespEmpty.GenError()
	}
	l.ctx = ctx
	l.req = req
	l.resp = resp

	if err := l.req.Validate(); err != nil {
		blog.Errorf("list release failed, invalid request, %s, param: %v", err.Error(), l.req)
		l.setResp(common.ErrHelmManagerRequestParamInvalid, err.Error(), nil)
		return nil
	}

	return l.list()
}

func (l *ListReleaseAction) list() error {
	clusterID := l.req.GetClusterID()
	option := l.getOption()

	total, origin, err := l.releaseHandler.Cluster(clusterID).List(l.ctx, option)
	if err != nil {
		blog.Errorf("list release failed, %s, clusterID: %s", err.Error(), clusterID)
		l.setResp(common.ErrHelmManagerListActionFailed, err.Error(), nil)
		return nil
	}

	r := make([]*helmmanager.Release, 0, len(origin))
	for _, item := range origin {
		r = append(r, item.Transfer2Proto())
	}
	l.setResp(common.ErrHelmManagerSuccess, "ok", &helmmanager.ReleaseListData{
		Page:  common.GetUint32P(uint32(option.Page)),
		Size:  common.GetUint32P(uint32(option.Size)),
		Total: common.GetUint32P(uint32(total)),
		Data:  r,
	})
	blog.Infof("list release successfully")
	return nil
}

func (l *ListReleaseAction) getOption() release.ListOption {
	size := l.req.GetSize()
	if size == 0 {
		size = defaultSize
	}

	return release.ListOption{
		Page:      int64(l.req.GetPage()),
		Size:      int64(size),
		Namespace: l.req.GetNamespace(),
		Name:      l.req.GetName(),
	}
}

func (l *ListReleaseAction) setResp(err common.HelmManagerError, message string, r *helmmanager.ReleaseListData) {
	code := err.Int32()
	msg := err.ErrorMessage(message)
	l.resp.Code = &code
	l.resp.Message = &msg
	l.resp.Result = err.OK()
	l.resp.Data = r
}
