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

package release

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/auth"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/release"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/store/entity"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/utils/contextx"
	helmmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/proto/bcs-helm-manager"
)

// NewImportClusterReleaseAction return a new NewImportClusterReleaseAction instance
func NewImportClusterReleaseAction(model store.HelmManagerModel,
	releaseHandler release.Handler) *ImportClusterReleaseAction {
	return &ImportClusterReleaseAction{
		model:          model,
		releaseHandler: releaseHandler,
	}
}

// ImportClusterReleaseAction provides the action to do list chart releases
type ImportClusterReleaseAction struct {
	ctx context.Context

	model          store.HelmManagerModel
	releaseHandler release.Handler

	req  *helmmanager.ImportClusterReleaseReq
	resp *helmmanager.ImportClusterReleaseResp
}

// Handle the listing process
func (l *ImportClusterReleaseAction) Handle(ctx context.Context,
	req *helmmanager.ImportClusterReleaseReq, resp *helmmanager.ImportClusterReleaseResp) error {
	l.ctx = ctx
	l.req = req
	l.resp = resp

	if err := l.req.Validate(); err != nil {
		blog.Errorf("import cluster release failed, invalid request, %s, param: %v", err.Error(), l.req)
		l.setResp(common.ErrHelmManagerRequestParamInvalid, err.Error())
		return nil
	}
	if err := l.importClusterRelease(); err != nil {
		blog.Errorf("import cluster release %s failed, clusterID: %s, namespace: %s, error: %s",
			l.req.GetName(), l.req.GetClusterID(), l.req.GetNamespace(), err.Error())
		l.setResp(common.ErrHelmManagerImportFailed, err.Error())
		return nil
	}

	l.setResp(common.ErrHelmManagerSuccess, "ok")
	blog.Infof("import cluster release successfully, projectCode: %s, clusterID: %s, namespace: %s, name: %s",
		l.req.GetProjectCode(), l.req.GetClusterID(), l.req.GetNamespace(), l.req.GetName())
	return nil
}
func (l *ImportClusterReleaseAction) importClusterRelease() error {
	projectCode := contextx.GetProjectCodeFromCtx(l.ctx)
	clusterID := l.req.GetClusterID()
	namespace := l.req.GetNamespace()
	name := l.req.GetName()
	repoName := l.req.GetRepoName()
	chartName := l.req.GetChartName()
	createBy := auth.GetUserFromCtx(l.ctx)
	// 查询DB是否有该Release记录
	getRelease, err := l.model.GetRelease(l.ctx, clusterID, namespace, name)
	if err != nil {
		if !errors.Is(err, drivers.ErrTableRecordNotFound) {
			return err
		}
	}
	// 记录已存在
	if getRelease != nil {
		return fmt.Errorf("release %s already exists", name)
	}
	// 获取Release信息
	detail, err := l.getDetail()
	if err != nil {
		return err
	}
	updateTime, _ := time.Parse(time.RFC3339, detail.UpdateTime)
	// 在DB中记录该Release
	return l.model.CreateRelease(l.ctx, &entity.Release{
		Name:         l.req.GetName(),
		ProjectCode:  projectCode,
		Namespace:    l.req.GetNamespace(),
		ClusterID:    l.req.GetClusterID(),
		Repo:         repoName,
		ChartName:    chartName,
		ChartVersion: detail.ChartVersion,
		Values:       []string{detail.Values},
		CreateBy:     createBy,
		CreateTime:   updateTime.Unix(),
		UpdateTime:   updateTime.Unix(),
		Status:       detail.Status,
	})
}

func (l *ImportClusterReleaseAction) getDetail() (*release.Release, error) {
	clusterID := l.req.GetClusterID()
	namespace := l.req.GetNamespace()
	name := l.req.GetName()
	rl, err := l.releaseHandler.Cluster(clusterID).Get(l.ctx, release.GetOption{
		Namespace: namespace,
		Name:      name,
	})
	if err != nil {
		blog.Errorf("releaseHandler get release detail failed, %s, clusterID: %s namespace: %s, name: %s",
			err.Error(), l.req.GetClusterID(), l.req.GetNamespace(), l.req.GetName())
		return nil, err
	}
	return rl, nil
}

func (l *ImportClusterReleaseAction) setResp(err common.HelmManagerError, message string) {
	code := err.Int32()
	msg := err.ErrorMessage(message)
	l.resp.Code = &code
	l.resp.Message = &msg
	l.resp.Result = err.OK()
}
