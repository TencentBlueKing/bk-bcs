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

// NewGetReleaseDetailAction return a new GetReleaseDetailAction instance
func NewGetReleaseDetailAction(model store.HelmManagerModel, releaseHandler release.Handler) *GetReleaseDetailAction {
	return &GetReleaseDetailAction{
		model:          model,
		releaseHandler: releaseHandler,
	}
}

// GetReleaseDetailAction provides the action to do get chart detail info
type GetReleaseDetailAction struct {
	ctx context.Context

	model          store.HelmManagerModel
	releaseHandler release.Handler

	req  *helmmanager.GetReleaseDetailReq
	resp *helmmanager.GetReleaseDetailResp
}

// Handle the chart detail getting process
func (g *GetReleaseDetailAction) Handle(ctx context.Context,
	req *helmmanager.GetReleaseDetailReq, resp *helmmanager.GetReleaseDetailResp) error {

	if req == nil || resp == nil {
		blog.Errorf("get release detail failed, req or resp is empty")
		return common.ErrHelmManagerReqOrRespEmpty.GenError()
	}
	g.ctx = ctx
	g.req = req
	g.resp = resp

	if err := g.req.Validate(); err != nil {
		blog.Errorf("get release detail failed, invalid request, %s, param: %v", err.Error(), g.req)
		g.setResp(common.ErrHelmManagerRequestParamInvalid, err.Error(), nil)
		return nil
	}

	return g.getDetail()
}

func (g *GetReleaseDetailAction) getDetail() error {
	clusterID := g.req.GetClusterID()
	namespace := g.req.GetNamespace()
	name := g.req.GetName()

	_, origin, err := g.releaseHandler.Cluster(clusterID).List(g.ctx, release.ListOption{
		Namespace: namespace,
		Name:      name,
		Size:      1,
	})
	if err != nil {
		blog.Errorf("get release detail failed, %s, clusterID: %s namespace: %s, name: %s",
			err.Error(), clusterID, namespace, name)
		g.setResp(common.ErrHelmManagerGetActionFailed, err.Error(), nil)
		return nil
	}

	if len(origin) == 0 {
		blog.Errorf("get release detail failed, target not found, clusterID: %s namespace: %s, name: %s",
			clusterID, namespace, name)
		g.setResp(common.ErrHelmManagerGetActionFailed, "release not found", nil)
		return nil
	}

	rls := origin[0].Transfer2DetailProto()
	storedRelease, err := g.model.GetRelease(g.ctx, clusterID, namespace, name, int(rls.GetRevision()))
	if err != nil {
		blog.Errorf("get release detail from store failed, %s, "+
			"clusterID: %s namespace: %s, name: %s, revision: %d",
			err.Error(), clusterID, namespace, name, rls.GetRevision())
		g.setResp(common.ErrHelmManagerGetActionFailed, err.Error(), nil)
		return nil
	}

	rls.Values = storedRelease.Values
	g.setResp(common.ErrHelmManagerSuccess, "ok", rls)
	blog.Infof("get release detail successfully, "+
		"clusterID: %s namespace: %s, name: %s, revision: %d",
		clusterID, namespace, name, rls.GetRevision())
	return nil
}

func (g *GetReleaseDetailAction) setResp(err common.HelmManagerError, message string, r *helmmanager.ReleaseDetail) {
	code := err.Int32()
	msg := err.ErrorMessage(message)
	g.resp.Code = &code
	g.resp.Message = &msg
	g.resp.Result = err.OK()
	g.resp.Data = r
}
