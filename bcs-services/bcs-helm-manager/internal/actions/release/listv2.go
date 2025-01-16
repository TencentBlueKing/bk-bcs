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
	"fmt"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"golang.org/x/sync/errgroup"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/component/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/component/project"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/release"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/store/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/utils/contextx"
	helmmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/proto/bcs-helm-manager"
)

// NewListReleaseV2Action return a new ListReleaseAction instance
func NewListReleaseV2Action(model store.HelmManagerModel, releaseHandler release.Handler) *ListReleaseV2Action {
	return &ListReleaseV2Action{
		ListReleaseV1Action: ListReleaseV1Action{
			model:          model,
			releaseHandler: releaseHandler,
		},
	}
}

// ListReleaseV2Action provides the action to do list chart release
type ListReleaseV2Action struct {
	ListReleaseV1Action
}

// Handle the listing process
func (l *ListReleaseV2Action) Handle(ctx context.Context,
	req *helmmanager.ListReleaseV1Req, resp *helmmanager.ListReleaseV1Resp) error {
	l.ctx = ctx
	l.req = req
	l.resp = resp

	if err := l.req.Validate(); err != nil {
		blog.Errorf("list release failed, invalid request, %s, param: %v", err.Error(), l.req)
		l.setResp(common.ErrHelmManagerRequestParamInvalid, err.Error(), nil)
		return nil
	}

	result, err := l.list()
	if err != nil {
		blog.Errorf("get release list failed, %s, clusterID: %s namespace: %s",
			err.Error(), l.req.GetClusterID(), l.req.GetNamespace())
		l.setResp(common.ErrHelmManagerListActionFailed, err.Error(), nil)
		return nil
	}
	l.setResp(common.ErrHelmManagerSuccess, "ok", result)

	// append web_annotations
	l.resp.WebAnnotations = l.getWebAnnotations()

	blog.Infof("get release list successfully, clusterID: %s namespace: %s",
		l.req.GetClusterID(), l.req.GetNamespace())
	return nil
}

func (l *ListReleaseV2Action) list() (*helmmanager.ReleaseListData, error) {
	clusterID := l.req.GetClusterID()
	option := l.getOption()

	// if cluster is shared, return release form database only
	cluster, err := clustermanager.GetCluster(clusterID)
	if err != nil {
		return nil, fmt.Errorf("get cluster info error, %s", err.Error())
	}

	// get release from store
	_, rls, err := l.model.ListRelease(l.ctx, l.getCondition(cluster.IsShared), &utils.ListOption{})
	if err != nil {
		return nil, fmt.Errorf("list release from db error, %s", err.Error())
	}
	dbReleases := make([]*helmmanager.Release, 0, len(rls))
	for _, item := range rls {
		dbReleases = append(dbReleases, item.Transfer2Proto())
	}

	if cluster.IsShared && len(option.Namespace) == 0 {
		clusterReleases, errr := l.listReleaseByNamespaces()
		if errr != nil {
			return nil, errr
		}
		return l.mergeReleases(clusterReleases, dbReleases), nil
	}

	// get release from cluster
	_, origin, err := l.releaseHandler.Cluster(clusterID).ListV2(l.ctx, option)
	if err != nil {
		return nil, fmt.Errorf("list release from cluster error, %s", err.Error())
	}
	clusterReleases := make([]*helmmanager.Release, 0, len(origin))
	for _, item := range origin {
		clusterReleases = append(clusterReleases, item.Transfer2Proto(contextx.GetProjectCodeFromCtx(l.ctx), clusterID))
	}

	// merge release
	return l.mergeReleases(clusterReleases, dbReleases), nil
}

// 共享集群支持查询集群下所有命名空间的release
func (l *ListReleaseV2Action) listReleaseByNamespaces() ([]*helmmanager.Release, error) {
	namespaces, err := project.ListNamespaces(l.ctx, l.req.GetProjectCode(), l.req.GetClusterID())
	if err != nil {
		return nil, err
	}
	clusterReleases := make([]*helmmanager.Release, 0)
	eg := errgroup.Group{}
	mux := sync.Mutex{}
	eg.SetLimit(10)
	for _, data := range namespaces {
		nsData := data
		eg.Go(func() error {

			// get release from cluster
			_, originReleases, errr := l.releaseHandler.Cluster(l.req.GetClusterID()).ListV2(l.ctx, release.ListOption{
				Namespace: nsData.Name,
				Name:      "",
			})
			if errr != nil {
				return fmt.Errorf("list release from cluster error, %s", errr.Error())
			}
			mux.Lock()
			for _, item := range originReleases {
				clusterReleases = append(clusterReleases, item.Transfer2Proto(contextx.GetProjectCodeFromCtx(l.ctx),
					l.req.GetClusterID()))
			}
			mux.Unlock()
			return nil
		})
	}
	err = eg.Wait()
	if err != nil {
		return nil, err
	}
	return clusterReleases, nil
}
