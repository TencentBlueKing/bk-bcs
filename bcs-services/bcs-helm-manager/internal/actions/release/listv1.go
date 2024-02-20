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
	"sort"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	authUtils "github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
	helmrelease "helm.sh/helm/v3/pkg/release"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/auth"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/component/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/release"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/store/entity"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/store/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/utils/contextx"
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

	// append web_annotations
	l.resp.WebAnnotations = l.getWebAnnotations()

	blog.Infof("get release list successfully, clusterID: %s namespace: %s",
		l.req.GetClusterID(), l.req.GetNamespace())
	return nil
}

func (l *ListReleaseV1Action) list() (*helmmanager.ReleaseListData, error) {
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
		return l.mergeReleases(nil, dbReleases), nil
	}

	// get release from cluster
	_, origin, err := l.releaseHandler.Cluster(clusterID).List(l.ctx, option)
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

func (l *ListReleaseV1Action) mergeReleases(clusterReleases,
	dbReleases []*helmmanager.Release) *helmmanager.ReleaseListData {
	release := make([]*helmmanager.Release, 0)
	newReleaseMap := make(map[string]*helmmanager.Release, 0)
	for i, v := range clusterReleases {
		newReleaseMap[l.getReleaseKey(v.GetNamespace(), v.GetName())] = clusterReleases[i]
	}

	for i, v := range dbReleases {
		// 如果 release 在集群中存在，则只更新 release 字段，否则直接使用数据库的数据
		if _, ok := newReleaseMap[l.getReleaseKey(v.GetNamespace(), v.GetName())]; ok {
			// 使用数据库中的状态，因为创建或者更新失败状态，原生中没有该状态
			if v.GetStatus() != "" {
				newReleaseMap[l.getReleaseKey(v.GetNamespace(), v.GetName())].Status = v.Status
				newReleaseMap[l.getReleaseKey(v.GetNamespace(), v.GetName())].Message = v.Message
			}
			// 集群中版本比数据库中的版本新，则使用集群中的数据
			if *v.Revision >= *newReleaseMap[l.getReleaseKey(v.GetNamespace(), v.GetName())].Revision {
				newReleaseMap[l.getReleaseKey(v.GetNamespace(), v.GetName())].ChartVersion = v.ChartVersion
				newReleaseMap[l.getReleaseKey(v.GetNamespace(), v.GetName())].UpdateTime = v.UpdateTime
				newReleaseMap[l.getReleaseKey(v.GetNamespace(), v.GetName())].Message = v.Message
			}
			newReleaseMap[l.getReleaseKey(v.GetNamespace(), v.GetName())].CreateBy = v.CreateBy
			newReleaseMap[l.getReleaseKey(v.GetNamespace(), v.GetName())].UpdateBy = v.UpdateBy
			newReleaseMap[l.getReleaseKey(v.GetNamespace(), v.GetName())].Repo = v.Repo
			newReleaseMap[l.getReleaseKey(v.GetNamespace(), v.GetName())].Env = v.Env
			continue
		}
		// 数据库中状态正常的数据，但在集群中不存在，则不展示
		if v.GetStatus() == helmrelease.StatusDeployed.String() {
			continue
		}
		newReleaseMap[l.getReleaseKey(v.GetNamespace(), v.GetName())] = dbReleases[i]
	}

	for k := range newReleaseMap {
		nsID := authUtils.CalcIAMNsID(l.req.GetClusterID(), *newReleaseMap[k].Namespace)
		newReleaseMap[k].IamNamespaceID = &nsID
		release = append(release, newReleaseMap[k])
	}

	// sort
	sort.Sort(ReleasesSortByUpdateTime(release))
	total := len(release)
	if l.req.GetPage() > 0 && l.req.GetSize() > 0 {
		release = filterIndex(int((l.req.GetPage()-1)*l.req.GetSize()), int(l.req.GetSize()), release)
	}
	return &helmmanager.ReleaseListData{
		Page:  common.GetUint32P(l.req.GetPage()),
		Size:  common.GetUint32P(l.req.GetSize()),
		Total: common.GetUint32P(uint32(total)),
		Data:  release,
	}
}

func (l *ListReleaseV1Action) getWebAnnotations() *helmmanager.WebAnnotations {
	namespaces := make([]string, 0)
	for _, v := range l.resp.Data.Data {
		namespaces = append(namespaces, v.GetNamespace())
	}
	if len(namespaces) == 0 {
		return nil
	}

	username := auth.GetUserFromCtx(l.ctx)
	projectID := contextx.GetProjectIDFromCtx(l.ctx)
	perms, err := auth.GetUserNamespacePermList(username, projectID, l.req.GetClusterID(), namespaces)
	if err != nil {
		blog.Errorf("get user %s namespace perms failed, err: %s", username, err.Error())
		return nil
	}

	s, err := common.MarshalInterfaceToValue(perms)
	if err != nil {
		blog.Errorf("MarshalInterfaceToValue failed, perms %v, err: %s", perms, err.Error())
		return nil
	}
	webAnnotations := &helmmanager.WebAnnotations{
		Perms: s,
	}
	return webAnnotations
}

func (l *ListReleaseV1Action) getOption() release.ListOption {
	return release.ListOption{
		Namespace: l.req.GetNamespace(),
		Name:      strings.ToLower(l.req.GetName()),
	}
}

func (l *ListReleaseV1Action) getReleaseKey(namespace, name string) string {
	return fmt.Sprintf("%s/%s", namespace, name)
}

func (l *ListReleaseV1Action) getCondition(shared bool) *operator.Condition {
	cond := make(operator.M)
	if shared {
		cond.Update(entity.FieldKeyProjectCode, contextx.GetProjectCodeFromCtx(l.ctx))
	}
	cond.Update(entity.FieldKeyClusterID, l.req.GetClusterID())
	if len(l.req.GetNamespace()) != 0 {
		cond.Update(entity.FieldKeyNamespace, l.req.GetNamespace())
	}
	if len(l.req.GetName()) != 0 {
		cond.Update(entity.FieldKeyName, primitive.Regex{Pattern: strings.ToLower(l.req.GetName()), Options: "i"})
	}
	cond1 := operator.NewLeafCondition(operator.Eq, cond)
	cond2 := operator.NewLeafCondition(operator.Ne, operator.M{entity.FieldKeyChartName: ""})
	return operator.NewBranchCondition(operator.And, cond1, cond2)
}

func (l *ListReleaseV1Action) setResp(err common.HelmManagerError, message string, r *helmmanager.ReleaseListData) {
	code := err.Int32()
	msg := err.ErrorMessage(message)
	l.resp.Code = &code
	l.resp.Message = &msg
	l.resp.Result = err.OK()
	l.resp.Data = r
}
