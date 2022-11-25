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

package actions

import (
	"context"
	"fmt"
	"strconv"

	helmrelease "helm.sh/helm/v3/pkg/release"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/operation"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/release"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/repo"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/store/entity"
)

// ReleaseInstallAction release install action
type ReleaseInstallAction struct {
	model          store.HelmManagerModel
	platform       repo.Platform
	releaseHandler release.Handler

	projectCode string
	projectID   string
	clusterID   string
	name        string
	namespace   string
	repoName    string
	chartName   string
	version     string
	values      []string
	args        []string
	username    string

	contents []byte
	result   *release.HelmInstallResult
}

// ReleaseInstallActionOption options
type ReleaseInstallActionOption struct {
	Model          store.HelmManagerModel
	Platform       repo.Platform
	ReleaseHandler release.Handler

	ProjectCode string
	ProjectID   string
	ClusterID   string
	Name        string
	Namespace   string
	RepoName    string
	ChartName   string
	Version     string
	Values      []string
	Args        []string
	Username    string
}

// NewReleaseInstallAction new release install action
func NewReleaseInstallAction(o *ReleaseInstallActionOption) *ReleaseInstallAction {
	return &ReleaseInstallAction{
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
		username:       o.Username,
	}
}

var _ operation.Operation = &ReleaseInstallAction{}

// Action xxx
func (r *ReleaseInstallAction) Action() string {
	return "Install"
}

// Name xxx
func (r *ReleaseInstallAction) Name() string {
	return fmt.Sprintf("install-%s", r.name)
}

// Validate xxx
func (r *ReleaseInstallAction) Validate() error {
	return nil
}

// Prepare xxx
func (r *ReleaseInstallAction) Prepare(ctx context.Context) error {
	repository, err := r.model.GetRepository(ctx, r.projectCode, r.repoName)
	if err != nil {
		return fmt.Errorf("get %s/%s repo info in cluster %s error, %s",
			r.namespace, r.name, r.clusterID, err.Error())
	}

	// 下载到具体的chart version信息
	contents, err := r.platform.
		User(repo.User{
			Name:     repository.Username,
			Password: repository.Password,
		}).
		Project(repository.GetRepoProjectID()).
		Repository(
			repo.GetRepositoryType(repository.Type),
			repository.GetRepoName(),
		).
		Chart(r.chartName).
		Download(ctx, r.version)
	if err != nil {
		return fmt.Errorf("download chart %s/%s in cluster %s error, %s",
			r.namespace, r.name, r.clusterID, err.Error())
	}

	r.contents = contents
	return nil
}

// Execute xxx
func (r *ReleaseInstallAction) Execute(ctx context.Context) error {
	vls := make([]*release.File, 0, len(r.values))
	for index, v := range r.values {
		vls = append(vls, &release.File{
			Name:    "values-" + strconv.Itoa(index) + ".yaml",
			Content: []byte(v),
		})
	}
	result, err := r.releaseHandler.Cluster(r.clusterID).Install(
		ctx, release.HelmInstallConfig{
			ProjectCode: r.projectCode,
			Name:        r.name,
			Namespace:   r.namespace,
			Chart: &release.File{
				Name:    r.chartName + "-" + r.version + ".tgz",
				Content: r.contents,
			},
			Args:   r.args,
			Values: vls,
			PatchTemplateValues: map[string]string{
				common.PTKProjectID: r.projectID,
				common.PTKClusterID: r.clusterID,
				common.PTKNamespace: r.namespace,
				common.PTKCreator:   r.username,
				common.PTKUpdator:   r.username,
				common.PTKVersion:   r.version,
				common.PTKName:      r.name,
			},
		})
	if err != nil {
		return fmt.Errorf("install %s/%s in cluster %s error, %s",
			r.namespace, r.name, r.clusterID, err.Error())
	}

	r.result = result
	return nil
}

// Done xxx
func (r *ReleaseInstallAction) Done(err error) {
	status := helmrelease.StatusDeployed
	message := ""
	if err != nil {
		status = common.ReleaseStatusInstallFailed
		message = err.Error()
	}
	rl := entity.M{
		entity.FieldKeyChartName:    r.chartName,
		entity.FieldKeyChartVersion: r.version,
		entity.FieldKeyValues:       r.values,
		entity.FieldKeyArgs:         r.args,
		entity.FieldKeyUpdateBy:     r.username,
		entity.FieldKeyStatus:       status.String(),
		entity.FieldKeyMessage:      message,
	}
	if r.result != nil {
		rl.Update(entity.FieldKeyRevision, r.result.Revision)
	}
	_ = r.model.UpdateRelease(context.Background(), r.clusterID, r.namespace, r.name, rl)
}
