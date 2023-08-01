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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/utils/contextx"
	helmmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/proto/bcs-helm-manager"
)

// NewGetReleaseManifestAction return a new GetReleaseManifestAction instance
func NewGetReleaseManifestAction(releaseHandler release.Handler) *GetReleaseManifestAction {
	return &GetReleaseManifestAction{
		releaseHandler: releaseHandler,
	}
}

// GetReleaseManifestAction provides the action to do get release manifest
type GetReleaseManifestAction struct {
	ctx context.Context

	releaseHandler release.Handler

	req  *helmmanager.GetReleaseManifestReq
	resp *helmmanager.GetReleaseManifestResp
}

// Handle the release manifest getting process
func (g *GetReleaseManifestAction) Handle(ctx context.Context,
	req *helmmanager.GetReleaseManifestReq, resp *helmmanager.GetReleaseManifestResp) error {
	g.ctx = ctx
	g.req = req
	g.resp = resp

	projectCode := contextx.GetProjectCodeFromCtx(g.ctx)
	clusterID := g.req.GetClusterID()
	namespace := g.req.GetNamespace()
	name := g.req.GetName()

	if err := g.req.Validate(); err != nil {
		blog.Errorf("get release manifest failed, invalid request, %s, param: %v", err.Error(), g.req)
		g.setResp(common.ErrHelmManagerRequestParamInvalid, err.Error(), nil)
		return nil
	}

	result, err := g.getManifest()
	if err != nil {
		blog.Errorf("get release manifest failed, %s, clusterID: %s namespace: %s, name: %s",
			err.Error(), clusterID, namespace, name)
		g.setResp(common.ErrHelmManagerGetActionFailed, err.Error(), nil)
		return nil
	}
	g.setResp(common.ErrHelmManagerSuccess, "ok", result)
	blog.Infof("get release manifest successfully, projectCode: %s, clusterID: %s, namespace: %s, name: %s",
		projectCode, clusterID, namespace, name)
	return nil
}

func (g *GetReleaseManifestAction) getManifest() (map[string]*helmmanager.FileContent, error) {
	clusterID := g.req.GetClusterID()
	namespace := g.req.GetNamespace()
	name := g.req.GetName()
	revision := g.req.GetRevision()

	rel, err := g.releaseHandler.Cluster(clusterID).Get(g.ctx, release.GetOption{
		Namespace: namespace,
		Name:      name,
		Revision:  int(revision),
	})
	if err != nil {
		return nil, err
	}

	result, err := generateFileContents(rel.Manifest)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (g *GetReleaseManifestAction) setResp(err common.HelmManagerError, message string,
	r map[string]*helmmanager.FileContent) {
	code := err.Int32()
	msg := err.ErrorMessage(message)
	g.resp.Code = &code
	g.resp.Message = &msg
	g.resp.Result = err.OK()
	g.resp.Data = r
}
