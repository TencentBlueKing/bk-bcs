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
	"fmt"
	"sort"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	helmrelease "helm.sh/helm/v3/pkg/release"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/release"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/store/entity"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/store/utils"
	helmmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/proto/bcs-helm-manager"
)

// NewListReleaseV1Action return a new ListReleaseAction instance
func NewListReleaseV1Action(model store.HelmManagerModel, releaseHandler release.Handler) *ListReleaseV1Action {
	return &ListReleaseV1Action{
		model:          model,
		releaseHandler: releaseHandler,
	}
}

// ListReleaseV1Action provides the action to do list chart release
type ListReleaseV1Action struct {
	ctx            context.Context
	model          store.HelmManagerModel
	releaseHandler release.Handler
	req            *helmmanager.ListReleaseV1Req
	resp           *helmmanager.ListReleaseV1Resp
}

// Handle the listing process
func (l *ListReleaseV1Action) Handle(ctx context.Context,
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
	blog.Infof("get release list successfully, clusterID: %s namespace: %s",
		l.req.GetClusterID(), l.req.GetNamespace())
	return nil
}

func (l *ListReleaseV1Action) list() (*helmmanager.ReleaseListData, error) {
	clusterID := l.req.GetClusterID()
	option := l.getOption()

	// get release from cluster
	_, origin, err := l.releaseHandler.Cluster(clusterID).List(l.ctx, option)
	if err != nil {
		return nil, fmt.Errorf("list release from cluster error, %s", err.Error())
	}
	clusterReleases := make([]*helmmanager.Release, 0, len(origin))
	for _, item := range origin {
		clusterReleases = append(clusterReleases, item.Transfer2Proto())
	}

	// get release from store
	_, rls, err := l.model.ListRelease(l.ctx, l.getCondition(), &utils.ListOption{})
	if err != nil {
		return nil, fmt.Errorf("list release from db error, %s", err.Error())
	}
	dbReleases := make([]*helmmanager.Release, 0, len(rls))
	for _, item := range rls {
		dbReleases = append(dbReleases, item.Transfer2Proto())
	}

	// merge release
	return l.mergeReleases(clusterReleases, dbReleases), nil
}

func (l *ListReleaseV1Action) mergeReleases(clusterReleases,
	dbReleases []*helmmanager.Release) *helmmanager.ReleaseListData {
	release := make([]*helmmanager.Release, 0)
	newReleaseMap := make(map[string]*helmmanager.Release, 0)
	for i, v := range clusterReleases {
		newReleaseMap[l.getReleaseKey(v.GetNamespace(), v.GetName())] = clusterReleases[i]
	}
	for i, v := range dbReleases {
		if n, ok := newReleaseMap[l.getReleaseKey(v.GetNamespace(), v.GetName())]; ok {
			if n.GetStatus() != helmrelease.StatusDeployed.String() && v.GetStatus() != "" {
				newReleaseMap[l.getReleaseKey(v.GetNamespace(), v.GetName())].Status = v.Status
			}
			newReleaseMap[l.getReleaseKey(v.GetNamespace(), v.GetName())].CreateBy = v.CreateBy
			newReleaseMap[l.getReleaseKey(v.GetNamespace(), v.GetName())].UpdateBy = v.UpdateBy
			newReleaseMap[l.getReleaseKey(v.GetNamespace(), v.GetName())].Message = v.Message
			continue
		}
		newReleaseMap[l.getReleaseKey(v.GetNamespace(), v.GetName())] = dbReleases[i]
	}

	for k := range newReleaseMap {
		release = append(release, newReleaseMap[k])
	}

	// sort
	sort.Sort(ReleasesSortByUpdateTime(release))
	total := len(release)
	if l.req.GetPage() > 0 && l.req.GetSize() > 0 {
		release = filterIndex(int((l.req.GetPage()-1)*l.req.GetSize()), int(l.req.GetSize()), release)
	}
	return &helmmanager.ReleaseListData{
		Page:  common.GetUint32P(uint32(l.req.GetPage())),
		Size:  common.GetUint32P(uint32(l.req.GetSize())),
		Total: common.GetUint32P(uint32(total)),
		Data:  release,
	}
}

func (l *ListReleaseV1Action) getOption() release.ListOption {
	return release.ListOption{
		Namespace: l.req.GetNamespace(),
		Name:      l.req.GetName(),
	}
}

func (l *ListReleaseV1Action) getReleaseKey(namespace, name string) string {
	return fmt.Sprintf("%s/%s", namespace, name)
}

func (l *ListReleaseV1Action) getCondition() *operator.Condition {
	cond := make(operator.M)
	cond.Update(entity.FieldKeyClusterID, l.req.GetClusterID())
	if l.req.Namespace != nil {
		cond.Update(entity.FieldKeyNamespace, l.req.GetNamespace())
	}
	if l.req.Name != nil {
		cond.Update(entity.FieldKeyName, l.req.GetName())
	}
	return operator.NewLeafCondition(operator.Eq, cond)
}

func (l *ListReleaseV1Action) setResp(err common.HelmManagerError, message string, r *helmmanager.ReleaseListData) {
	code := err.Int32()
	msg := err.ErrorMessage(message)
	l.resp.Code = &code
	l.resp.Message = &msg
	l.resp.Result = err.OK()
	l.resp.Data = r
}
