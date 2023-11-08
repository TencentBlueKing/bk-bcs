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

package addons

import (
	"context"
	"errors"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"golang.org/x/sync/errgroup"
	helmrelease "helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage/driver"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/release"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/repo"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/store"
	helmmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/proto/bcs-helm-manager"
)

// NewListAddonsAction return a new ListAddonsAction instance
func NewListAddonsAction(model store.HelmManagerModel, addons release.AddonsSlice,
	platform repo.Platform, releaseHandler release.Handler) *ListAddonsAction {
	return &ListAddonsAction{
		model:          model,
		addons:         addons,
		platform:       platform,
		releaseHandler: releaseHandler,
	}
}

// ListAddonsAction provides the action to do list addons
type ListAddonsAction struct {
	model          store.HelmManagerModel
	addons         release.AddonsSlice
	platform       repo.Platform
	releaseHandler release.Handler

	req  *helmmanager.ListAddonsReq
	resp *helmmanager.ListAddonsResp
}

// Handle the addons listing process
// NOCC:golint/fnsize(设计如此：无法拆分代码行数)
func (l *ListAddonsAction) Handle(ctx context.Context,
	req *helmmanager.ListAddonsReq, resp *helmmanager.ListAddonsResp) error {
	l.req = req
	l.resp = resp

	if err := req.Validate(); err != nil {
		blog.Errorf("get addons detail failed, invalid request, %s, param: %v", err.Error(), l.req)
		l.setResp(common.ErrHelmManagerRequestParamInvalid, err.Error(), nil)
		return nil
	}

	// 创建一个Addons数组，容量为配置文件数组的长度
	eg := errgroup.Group{}
	mux := sync.Mutex{}
	addons := make([]*helmmanager.Addons, 0, len(l.addons.Addons))
	for _, value := range l.addons.Addons {
		addon := value.ToAddonsProto()
		addons = append(addons, addon)
	}

	for i := range addons {
		i := i
		eg.Go(func() error {
			// get latest version
			version, err := l.getLatestVersion(ctx, addons[i].GetChartName())
			if err != nil {
				blog.Errorf("get addons latest version failed, %s", err.Error())
			}
			mux.Lock()
			addons[i].Version = &version
			mux.Unlock()

			// get current status
			rl, err := l.model.GetRelease(ctx, l.req.GetClusterID(), addons[i].GetNamespace(),
				addons[i].GetReleaseName())
			if err != nil && !errors.Is(err, drivers.ErrTableRecordNotFound) {
				blog.Errorf("get addons %s status failed, %s", addons[i].GetName(), err.Error())
				return err
			}
			mux.Lock()
			revision := 0
			if rl != nil {
				addons[i].CurrentVersion = &rl.ChartVersion
				addons[i].Status = &rl.Status
				addons[i].Message = &rl.Message
				if len(rl.Values) > 0 {
					addons[i].CurrentValues = &rl.Values[len(rl.Values)-1]
				}
				revision = rl.Revision
			}
			mux.Unlock()

			// 读取集群状态
			clusterRelease, err := l.releaseHandler.Cluster(l.req.GetClusterID()).
				Get(ctx, release.GetOption{Namespace: addons[i].GetNamespace(), Name: addons[i].GetReleaseName()})
			mux.Lock()
			if err != nil {
				if errors.Is(err, driver.ErrReleaseNotFound) &&
					*addons[i].Status == helmrelease.StatusDeployed.String() && *addons[i].ChartName != "" {
					// 如果集群中没有改 release，则置为未安装状态
					if *addons[i].Status == helmrelease.StatusDeployed.String() && *addons[i].ChartName != "" {
						addons[i].Status = common.GetStringP("")
						addons[i].Message = common.GetStringP("")
					}
				} else {
					blog.Warnf("releaseHandler get release detail failed, %s, clusterID: %s namespace: %s, name: %s",
						err.Error(), l.req.GetClusterID(), addons[i].GetNamespace(), addons[i].GetReleaseName())
				}
			}
			if clusterRelease != nil && clusterRelease.Revision > revision {
				addons[i].CurrentVersion = &clusterRelease.ChartVersion
				addons[i].Status = &clusterRelease.Status
				addons[i].Message = &clusterRelease.Description
			}
			mux.Unlock()
			return nil
		})
	}
	_ = eg.Wait()

	l.setResp(common.ErrHelmManagerSuccess, "ok", addons)
	return nil
}

func (l *ListAddonsAction) setResp(err common.HelmManagerError, message string, r []*helmmanager.Addons) {
	code := err.Int32()
	msg := err.ErrorMessage(message)
	l.resp.Code = &code
	l.resp.Message = &msg
	l.resp.Result = err.OK()
	l.resp.Data = r
}

func (l *ListAddonsAction) getLatestVersion(ctx context.Context, chartName string) (string, error) {
	repository, err := l.model.GetProjectRepository(ctx, l.req.GetProjectCode(), common.PublicRepoName)
	if err != nil {
		return "", err
	}

	detail, err := l.platform.
		User(repo.User{
			Name:     repository.Username,
			Password: repository.Password,
		}).
		Project(repository.GetRepoProjectID()).
		Repository(
			repo.GetRepositoryType(repository.Type),
			repository.GetRepoName(),
		).
		GetChartDetail(ctx, chartName)
	if err != nil {
		return "", err
	}
	return detail.Version, nil
}
