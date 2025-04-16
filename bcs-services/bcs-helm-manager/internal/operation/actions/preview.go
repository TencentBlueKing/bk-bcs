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

package actions

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	pkgrelease "helm.sh/helm/v3/pkg/release"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/release"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/repo"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/store"
)

// ReleasePreviewAction release preview action
type ReleasePreviewAction struct {
	model          store.HelmManagerModel
	platform       repo.Platform
	releaseHandler release.Handler

	projectCode    string
	projectID      string
	clusterID      string
	name           string
	namespace      string
	repoName       string
	chartName      string
	version        string
	values         []string
	args           []string
	createBy       string
	updateBy       string
	AuthUser       string
	IsShardCluster bool

	contents []byte
}

// ReleasePreviewActionOption options
type ReleasePreviewActionOption struct {
	Model          store.HelmManagerModel
	Platform       repo.Platform
	ReleaseHandler release.Handler

	ProjectCode    string
	ProjectID      string
	ClusterID      string
	Name           string
	Namespace      string
	RepoName       string
	ChartName      string
	Version        string
	Values         []string
	Args           []string
	CreateBy       string
	UpdateBy       string
	AuthUser       string
	IsShardCluster bool
	Content        []byte
}

// NewReleasePreviewAction new release preview action
func NewReleasePreviewAction(o *ReleasePreviewActionOption) *ReleasePreviewAction {
	return &ReleasePreviewAction{
		model:          o.Model,
		platform:       o.Platform,
		releaseHandler: o.ReleaseHandler,
		projectCode:    o.ProjectCode,
		projectID:      o.ProjectID,
		clusterID:      o.ClusterID,
		name:           o.Name,
		namespace:      o.Namespace,
		repoName:       o.RepoName,
		chartName:      o.ChartName,
		version:        o.Version,
		values:         o.Values,
		args:           o.Args,
		createBy:       o.CreateBy,
		updateBy:       o.UpdateBy,
		AuthUser:       o.AuthUser,
		IsShardCluster: o.IsShardCluster,
		contents:       o.Content,
	}
}

// UpgradeRelease xxx
func (r *ReleasePreviewAction) UpgradeRelease(ctx context.Context) (*pkgrelease.Release, error) {
	blog.V(5).Infof("start to upgrade release %s/%s preview", r.namespace, r.name)

	defer func() {
		// 防止部署过程 panic 导致整个程序都挂掉，同时 panic 后返回空的数据
		if r := recover(); r != nil {
			blog.Errorf("upgrade release dry run failed")
		}
	}()

	// get release from helm dry run
	result, err := release.UpgradeRelease(r.releaseHandler, r.projectID, r.projectCode, r.clusterID, r.name,
		r.namespace, r.chartName, r.version, r.createBy, r.updateBy, r.args, nil, r.contents, r.values, true)
	if err != nil {
		return nil, err
	}
	if result == nil || result.Release == nil || (result.Release.Manifest == "" && result.Release.Hooks == nil) {
		blog.Infof("release %s/%s is nil in cluster %s", r.namespace, r.name, r.clusterID)
		return nil, nil
	}

	return result.Release, nil
}
