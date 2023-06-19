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
	_struct "github.com/golang/protobuf/ptypes/struct"
	"google.golang.org/protobuf/types/known/structpb"
	"helm.sh/helm/v3/pkg/storage/driver"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/release"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/utils/contextx"
	helmmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/proto/bcs-helm-manager"
)

// NewGetReleaseStatusAction return a new GetGetReleaseStatusAction instance
func NewGetReleaseStatusAction(releaseHandler release.Handler) *GetReleaseStatusAction {
	return &GetReleaseStatusAction{
		releaseHandler: releaseHandler,
	}
}

// GetReleaseStatusAction provides the action to do get release status
type GetReleaseStatusAction struct {
	ctx context.Context

	releaseHandler release.Handler

	req  *helmmanager.GetReleaseStatusReq
	resp *helmmanager.CommonListResp
}

// Handle the release status getting process
func (g *GetReleaseStatusAction) Handle(ctx context.Context,
	req *helmmanager.GetReleaseStatusReq, resp *helmmanager.CommonListResp) error {
	g.ctx = ctx
	g.req = req
	g.resp = resp

	if err := g.req.Validate(); err != nil {
		blog.Errorf("get release status failed, invalid request, %s, param: %v", err.Error(), g.req)
		g.setResp(common.ErrHelmManagerRequestParamInvalid, err.Error(), nil)
		return nil
	}

	return g.getResourceMainfest()
}

func (g *GetReleaseStatusAction) getResourceMainfest() error {
	projectCode := contextx.GetProjectCodeFromCtx(g.ctx)
	clusterID := g.req.GetClusterID()
	namespace := g.req.GetNamespace()
	name := g.req.GetName()

	rl, err := g.releaseHandler.Cluster(clusterID).Get(g.ctx, release.GetOption{
		Namespace: namespace,
		Name:      name,
		GetObject: true,
	})
	if err != nil && !errors.Is(err, driver.ErrReleaseNotFound) {
		blog.Errorf("get release failed, %s, clusterID: %s namespace: %s, name: %s", err.Error(), clusterID,
			namespace, name)
		g.setResp(common.ErrHelmManagerGetActionFailed, err.Error(), nil)
		return nil
	}
	if rl == nil {
		g.setResp(common.ErrHelmManagerSuccess, "ok", &structpb.ListValue{})
		return nil
	}
	blog.V(6).Infof("release %s, namespace: %s, manifest: %s", rl.Name, rl.Namespace, rl.Manifest)
	if rl.Objects == nil {
		g.setResp(common.ErrHelmManagerSuccess, "ok", &structpb.ListValue{})
		blog.Infof("get release status successfully, projectCode: %s, clusterID: %s, namespace: %s, name: %s",
			projectCode, clusterID, namespace, name)
		return nil
	}

	result, err := common.MarshalInterfacesToListValue(rl.Objects)
	if err != nil {
		blog.Errorf("marshal objects err, %s", err.Error())
		g.setResp(common.ErrHelmManagerGetActionFailed, err.Error(), nil)
		return nil
	}

	g.setResp(common.ErrHelmManagerSuccess, "ok", result)
	blog.Infof("get release status successfully, projectCode: %s, clusterID: %s, namespace: %s, name: %s",
		projectCode, clusterID, namespace, name)
	return nil
}

func (g *GetReleaseStatusAction) setResp(err common.HelmManagerError, message string,
	r *_struct.ListValue) {
	code := err.Int32()
	msg := err.ErrorMessage(message)
	g.resp.Code = &code
	g.resp.Message = &msg
	g.resp.Result = err.OK()
	g.resp.Data = r
}
