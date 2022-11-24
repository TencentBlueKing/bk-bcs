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
	"errors"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"helm.sh/helm/v3/pkg/storage/driver"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/release"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/utils/contextx"
	helmmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/proto/bcs-helm-manager"
)

// NewGetReleaseHistoryAction return a new GetReleaseHistoryAction instance
func NewGetReleaseHistoryAction(releaseHandler release.Handler) *GetReleaseHistoryAction {
	return &GetReleaseHistoryAction{
		releaseHandler: releaseHandler,
	}
}

// GetReleaseHistoryAction provides the action to do get release history
type GetReleaseHistoryAction struct {
	ctx context.Context

	releaseHandler release.Handler

	req  *helmmanager.GetReleaseHistoryReq
	resp *helmmanager.GetReleaseHistoryResp
}

// Handle the release history getting process
func (g *GetReleaseHistoryAction) Handle(ctx context.Context,
	req *helmmanager.GetReleaseHistoryReq, resp *helmmanager.GetReleaseHistoryResp) error {
	g.ctx = ctx
	g.req = req
	g.resp = resp

	if err := g.req.Validate(); err != nil {
		blog.Errorf("get release history failed, invalid request, %s, param: %v", err.Error(), g.req)
		g.setResp(common.ErrHelmManagerRequestParamInvalid, err.Error(), nil)
		return nil
	}

	return g.getHistory()
}

func (g *GetReleaseHistoryAction) getHistory() error {
	projectCode := contextx.GetProjectCodeFromCtx(g.ctx)
	clusterID := g.req.GetClusterID()
	namespace := g.req.GetNamespace()
	name := g.req.GetName()

	history, err := g.releaseHandler.Cluster(clusterID).History(g.ctx, release.HelmHistoryOption{
		Namespace: namespace,
		Name:      name,
	})
	if err != nil && !errors.Is(err, driver.ErrReleaseNotFound) {
		blog.Errorf("get release history failed, %s, clusterID: %s namespace: %s, name: %s",
			err.Error(), clusterID, namespace, name)
		g.setResp(common.ErrHelmManagerGetActionFailed, err.Error(), nil)
		return nil
	}

	rh := make([]*helmmanager.ReleaseHistory, 0)
	for _, v := range history {
		if g.req.GetFilter() != "" && v.Status != g.req.GetFilter() {
			continue
		}
		rh = append(rh, v.Transfer2HistoryProto())
	}
	g.setResp(common.ErrHelmManagerSuccess, "ok", rh)
	blog.Infof("get release history successfully, projectCode: %s, clusterID: %s, namespace: %s, name: %s",
		projectCode, clusterID, namespace, name)
	return nil
}

func (g *GetReleaseHistoryAction) setResp(err common.HelmManagerError, message string,
	r []*helmmanager.ReleaseHistory) {
	code := err.Int32()
	msg := err.ErrorMessage(message)
	g.resp.Code = &code
	g.resp.Message = &msg
	g.resp.Result = err.OK()
	g.resp.Data = r
}
