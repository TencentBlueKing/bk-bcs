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

// Package release xxx
package release

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"helm.sh/helm/v3/pkg/storage/driver"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/release"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/store/entity"
	helmmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/proto/bcs-helm-manager"
)

// NewGetReleaseDetailV1Action return a new GetReleaseDetailAction instance
func NewGetReleaseDetailV1Action(model store.HelmManagerModel,
	releaseHandler release.Handler) *GetReleaseDetailV1Action {
	return &GetReleaseDetailV1Action{
		model:          model,
		releaseHandler: releaseHandler,
	}
}

// GetReleaseDetailV1Action provides the action to do get chart detail info
type GetReleaseDetailV1Action struct {
	ctx context.Context

	model          store.HelmManagerModel
	releaseHandler release.Handler

	req  *helmmanager.GetReleaseDetailV1Req
	resp *helmmanager.GetReleaseDetailV1Resp
}

// Handle the chart detail getting process
func (g *GetReleaseDetailV1Action) Handle(ctx context.Context,
	req *helmmanager.GetReleaseDetailV1Req, resp *helmmanager.GetReleaseDetailV1Resp) error {
	g.ctx = ctx
	g.req = req
	g.resp = resp

	if err := g.req.Validate(); err != nil {
		blog.Errorf("get release detail failed, invalid request, %s, param: %v", err.Error(), g.req)
		g.setResp(common.ErrHelmManagerRequestParamInvalid, err.Error(), nil)
		return nil
	}

	detail, err := g.getDetail()
	if err != nil {
		blog.Errorf("get release detail failed, %s, clusterID: %s namespace: %s, name: %s",
			err.Error(), g.req.GetClusterID(), g.req.GetNamespace(), g.req.GetName())
		g.setResp(common.ErrHelmManagerGetActionFailed, err.Error(), nil)
		return nil
	}

	g.setResp(common.ErrHelmManagerSuccess, "ok", detail)
	blog.Infof("get release detail successfully, clusterID: %s namespace: %s, name: %s",
		g.req.GetClusterID(), g.req.GetNamespace(), g.req.GetName())
	return nil
}

func (g *GetReleaseDetailV1Action) getDetail() (*helmmanager.ReleaseDetail, error) {
	clusterID := g.req.GetClusterID()
	namespace := g.req.GetNamespace()
	name := g.req.GetName()

	rl, err := g.releaseHandler.Cluster(clusterID).Get(g.ctx, release.GetOption{
		Namespace: namespace,
		Name:      name,
	})
	if err != nil && !errors.Is(err, driver.ErrReleaseNotFound) {
		blog.Errorf("releaseHandler get release detail failed, %s, clusterID: %s namespace: %s, name: %s",
			err.Error(), g.req.GetClusterID(), g.req.GetNamespace(), g.req.GetName())
		return nil, err
	}
	detail := rl.Transfer2DetailProto()

	storedRelease, err := g.model.GetRelease(g.ctx, clusterID, namespace, name)
	if err != nil {
		blog.Warnf("get release from db failed, %s, clusterID: %s namespace: %s, name: %s",
			err.Error(), g.req.GetClusterID(), g.req.GetNamespace(), g.req.GetName())
	}

	result := g.mergeRelease(detail, storedRelease)
	if result == nil {
		blog.Errorf("merge release is empty, detail: %+v, storedRelease: +%v", detail, storedRelease)
		return nil, driver.ErrReleaseNotFound
	}
	return result, nil
}

func (g *GetReleaseDetailV1Action) mergeRelease(detail *helmmanager.ReleaseDetail,
	rl *entity.Release) *helmmanager.ReleaseDetail {
	if rl == nil {
		return detail
	}
	rl.Args = filterArgs(rl.Args)
	if detail == nil {
		return rl.Transfer2DetailProto()
	}

	t := time.Unix(rl.UpdateTime, 0).UTC().Format(time.RFC3339)
	if t >= detail.GetUpdateTime() {
		detail.Values = rl.Values
		detail.ChartVersion = &rl.ChartVersion
		detail.UpdateTime = common.GetStringP(t)
		detail.Status = &rl.Status
		detail.Message = &rl.Message
	}
	detail.Args = rl.Args
	detail.ValueFile = &rl.ValueFile
	detail.CreateBy = &rl.CreateBy
	detail.UpdateBy = &rl.UpdateBy
	detail.Repo = &rl.Repo
	return detail
}

func filterArgs(args []string) []string {
	result := make([]string, 0)
	for _, v := range args {
		if strings.HasPrefix(v, "--") {
			result = append(result, v)
		}
	}
	return result
}

func (g *GetReleaseDetailV1Action) setResp(err common.HelmManagerError, message string, r *helmmanager.ReleaseDetail) {
	code := err.Int32()
	msg := err.ErrorMessage(message)
	g.resp.Code = &code
	g.resp.Message = &msg
	g.resp.Result = err.OK()
	g.resp.Data = r
}
