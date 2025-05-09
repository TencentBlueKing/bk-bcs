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
	"strconv"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"gopkg.in/yaml.v2"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/release"
	helmrelease "helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/releaseutil"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/resource"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/utils/stringx"
	helmmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/proto/bcs-helm-manager"
)

const (
	defaultHelmUser = "admin"
)

// Handler 定义了 helm release 的client集合
type Handler interface {
	Cluster(clusterID string) Cluster
}

// Cluster 定义了每个 helm release client 的操作能力, 用于直接与集群产生helm命令交互
type Cluster interface {
	Get(ctx context.Context, option GetOption) (*Release, error)
	List(ctx context.Context, option ListOption) (int, []*Release, error)
	ListV2(ctx context.Context, option ListOption) (int, []*Release, error)
	Install(ctx context.Context, conf HelmInstallConfig) (*HelmInstallResult, error)
	Uninstall(ctx context.Context, conf HelmUninstallConfig) (*HelmUninstallResult, error)
	Upgrade(ctx context.Context, conf HelmUpgradeConfig) (*HelmUpgradeResult, error)
	Rollback(ctx context.Context, conf HelmRollbackConfig) (*HelmRollbackResult, error)
	History(ctx context.Context, option HelmHistoryOption) ([]*Release, error)
}

// Release 定义了集群中的helm release信息, 一般在命令行通过 helm list 获取
type Release struct {
	Name         string
	Namespace    string
	Revision     int
	Status       string
	Chart        string
	ChartVersion string
	AppVersion   string
	UpdateTime   string
	Description  string
	Values       string
	Manifest     string
	Hooks        []*helmrelease.Hook
	Infos        []*resource.Info
	Objects      []runtime.Object
	Notes        string
}

// Transfer2Release transfer the data into helm release struct
func (r *Release) Transfer2Release() *helmrelease.Release {
	if r == nil {
		return nil
	}
	return &helmrelease.Release{
		Name: r.Name,
		Info: &helmrelease.Info{
			Status: helmrelease.Status(r.Status),
			Notes:  r.Notes,
		},
		Chart:     &chart.Chart{Metadata: &chart.Metadata{Name: r.Chart}},
		Manifest:  r.Manifest,
		Hooks:     r.Hooks,
		Version:   r.Revision,
		Namespace: r.Namespace,
	}
}

// Transfer2Proto transfer the data into protobuf struct
func (r *Release) Transfer2Proto(projectCode, clusterID string) *helmmanager.Release {
	if r == nil {
		return nil
	}
	return &helmmanager.Release{
		Name:         common.GetStringP(r.Name),
		Namespace:    common.GetStringP(r.Namespace),
		Revision:     common.GetUint32P(uint32(r.Revision)),
		Status:       common.GetStringP(r.Status),
		Chart:        common.GetStringP(r.Chart),
		ChartVersion: common.GetStringP(r.ChartVersion),
		AppVersion:   common.GetStringP(r.AppVersion),
		UpdateTime:   common.GetStringP(r.UpdateTime),
		CreateBy:     common.GetStringP(defaultHelmUser),
		UpdateBy:     common.GetStringP(defaultHelmUser),
		Message:      common.GetStringP(r.Description),
		Repo:         common.GetStringP(""),
		ProjectCode:  common.GetStringP(projectCode),
		ClusterID:    common.GetStringP(clusterID),
	}
}

// Transfer2DetailProto transfer the data into detail protobuf struct
func (r *Release) Transfer2DetailProto() *helmmanager.ReleaseDetail {
	if r == nil {
		return nil
	}
	return &helmmanager.ReleaseDetail{
		Name:         common.GetStringP(r.Name),
		Namespace:    common.GetStringP(r.Namespace),
		Revision:     common.GetUint32P(uint32(r.Revision)),
		Status:       common.GetStringP(r.Status),
		Chart:        common.GetStringP(r.Chart),
		ChartVersion: common.GetStringP(r.ChartVersion),
		AppVersion:   common.GetStringP(r.AppVersion),
		UpdateTime:   common.GetStringP(r.UpdateTime),
		Values:       []string{r.Values},
		Description:  common.GetStringP(r.Description),
		Notes:        common.GetStringP(r.Notes),
		CreateBy:     common.GetStringP(defaultHelmUser),
		UpdateBy:     common.GetStringP(defaultHelmUser),
		Message:      common.GetStringP(r.Description),
		Repo:         common.GetStringP(""),
	}
}

// Transfer2HistoryProto transfer the data into history protobuf struct
func (r *Release) Transfer2HistoryProto() *helmmanager.ReleaseHistory {
	if r == nil {
		return nil
	}
	return &helmmanager.ReleaseHistory{
		Revision:     common.GetUint32P(uint32(r.Revision)),
		Name:         common.GetStringP(r.Name),
		Namespace:    common.GetStringP(r.Namespace),
		UpdateTime:   common.GetStringP(r.UpdateTime),
		Description:  common.GetStringP(r.Description),
		Status:       common.GetStringP(r.Status),
		Chart:        common.GetStringP(r.Chart),
		ChartVersion: common.GetStringP(r.ChartVersion),
		AppVersion:   common.GetStringP(r.AppVersion),
		Values:       common.GetStringP(r.Values),
	}
}

// ReleasesSlice define a slice of Release
type ReleasesSlice []*Release

// Len return the length of the slice
func (r ReleasesSlice) Len() int {
	return len(r)
}

// Less return true if the i-th element is less than the j-th element
func (r ReleasesSlice) Less(i, j int) bool {
	return r[i].Revision > r[j].Revision
}

// Swap swap the i-th element and the j-th element
func (r ReleasesSlice) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

// Config 定义了 Handler 的配置参数
type Config struct {
	PatchTemplates []*File
}

// GetOption 定义了 Cluster.Get 的查询参数
type GetOption struct {
	Namespace string
	Name      string
	// 需要获取的版本，如果为 0，则获取最新的版本
	Revision int
	// GetObject，是否从集群中获取资源
	GetObject bool
}

// ListOption 定义了 Cluster.List 的查询参数
type ListOption struct {
	Page      int64
	Size      int64
	Namespace string
	Name      string
}

// HelmInstallConfig 定义了helm执行install时的控制参数
type HelmInstallConfig struct {
	// simulate a install action
	DryRun     bool
	Replace    bool
	ClientOnly bool

	ProjectCode string
	Name        string
	Namespace   string

	Args []string

	Chart               *File
	Values              []*File
	PatchTemplateValues map[string]string
}

// HelmInstallResult 定义了helm执行install的返回结果
type HelmInstallResult struct {
	Release    *release.Release
	Revision   int
	Status     string
	AppVersion string
	UpdateTime string
}

// ToUpgradeResult transfer to upgrade result
func (h *HelmInstallResult) ToUpgradeResult() *HelmUpgradeResult {
	if h == nil {
		return nil
	}
	return &HelmUpgradeResult{
		Release:    h.Release,
		Revision:   h.Revision,
		Status:     h.Status,
		AppVersion: h.AppVersion,
		UpdateTime: h.UpdateTime,
	}
}

// HelmUninstallConfig 定义了helm执行uninstall时的控制参数
type HelmUninstallConfig struct {
	// simulate a uninstall action
	DryRun bool

	Name      string
	Namespace string
}

// HelmUninstallResult 定义了helm执行uninstall时的返回结果
type HelmUninstallResult struct {
}

// HelmUpgradeConfig 定义了helm执行upgrade时的控制参数
type HelmUpgradeConfig struct {
	// simulate a upgrade action
	DryRun bool

	ProjectCode string
	Name        string
	Namespace   string

	Args []string

	Chart               *File
	Values              []*File
	PatchTemplateValues map[string]string
}

// ToInstallConfig transfer to install config
func (h *HelmUpgradeConfig) ToInstallConfig() HelmInstallConfig {
	return HelmInstallConfig{
		DryRun:              h.DryRun,
		ProjectCode:         h.ProjectCode,
		Name:                h.Name,
		Namespace:           h.Namespace,
		Args:                h.Args,
		Chart:               h.Chart,
		Values:              h.Values,
		PatchTemplateValues: h.PatchTemplateValues,
	}
}

// HelmUpgradeResult 定义了helm执行upgrade时的返回结果
type HelmUpgradeResult struct {
	Release    *release.Release
	Revision   int
	Status     string
	AppVersion string
	UpdateTime string
}

// HelmRollbackConfig 定义了helm执行rollback时的控制参数
type HelmRollbackConfig struct {
	// simulate a rollback action
	DryRun bool

	Name      string
	Namespace string
	Revision  int
}

// HelmRollbackResult 定义了helm执行rollback时的返回结果
type HelmRollbackResult struct {
}

// File 定义了release中需要的文件信息
type File struct {
	Name    string
	Content []byte
}

// HelmHistoryOption 定义了helm执行history时的查询参数
type HelmHistoryOption struct {
	Name      string
	Namespace string
	Max       int
}

// InstallRelease install release
func InstallRelease(releaseHandler Handler, projectID, projectCode, clusterID, releaseName,
	releaseNamespace, chartName, version, creator, updator string, args []string, bcsSysVar map[string]string,
	contents []byte, values []string, dryRun, replace, clientOnly bool) (*HelmInstallResult, error) {
	vls := make([]*File, 0, len(values))
	for index, v := range values {
		vls = append(vls, &File{
			Name:    "values-" + strconv.Itoa(index) + ".yaml",
			Content: []byte(v),
		})
	}
	return releaseHandler.Cluster(clusterID).Install(
		context.Background(),
		HelmInstallConfig{
			DryRun:      dryRun,
			Replace:     replace,
			ClientOnly:  clientOnly,
			ProjectCode: projectCode,
			Name:        releaseName,
			Namespace:   releaseNamespace,
			Chart: &File{
				Name:    chartName + "-" + version + ".tgz",
				Content: contents,
			},
			Args:   args,
			Values: vls,
			PatchTemplateValues: map[string]string{
				common.PTKProjectID: projectID,
				common.PTKClusterID: clusterID,
				common.PTKNamespace: releaseNamespace,
				common.PTKCreator:   stringx.ReplaceIllegalChars(creator),
				common.PTKUpdator:   stringx.ReplaceIllegalChars(updator),
				common.PTKVersion:   version,
				common.PTKName:      releaseName,
			},
		})
}

// UpgradeRelease upgrade release
func UpgradeRelease(releaseHandler Handler, projectID, projectCode, clusterID, releaseName,
	releaseNamespace, chartName, version, creator, updator string, args []string, bcsSysVar map[string]string,
	contents []byte, values []string, dryRun bool) (*HelmUpgradeResult, error) {
	vls := make([]*File, 0, len(values))
	for index, v := range values {
		vls = append(vls, &File{
			Name:    "values-" + strconv.Itoa(index) + ".yaml",
			Content: []byte(v),
		})
	}
	return releaseHandler.Cluster(clusterID).Upgrade(
		context.Background(),
		HelmUpgradeConfig{
			DryRun:      dryRun,
			ProjectCode: projectCode,
			Name:        releaseName,
			Namespace:   releaseNamespace,
			Chart: &File{
				Name:    chartName + "-" + version + ".tgz",
				Content: contents,
			},
			Args:   args,
			Values: vls,
			PatchTemplateValues: map[string]string{
				common.PTKProjectID: projectID,
				common.PTKClusterID: clusterID,
				common.PTKNamespace: releaseNamespace,
				common.PTKCreator:   stringx.ReplaceIllegalChars(creator),
				common.PTKUpdator:   stringx.ReplaceIllegalChars(updator),
				common.PTKVersion:   version,
				common.PTKName:      releaseName,
			},
		})
}

// SimpleHead defines what the structure of the head of a manifest file
type SimpleHead struct {
	Version  string `json:"apiVersion" yaml:"apiVersion"`
	Kind     string `json:"kind,omitempty" yaml:"kind"`
	Metadata *struct {
		Namespace string `json:"namespace" yaml:"namespace"`
		Name      string `json:"name" yaml:"name"`
	} `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// GetManifestSimpleHeadFromRelease get SimpleHead from release manifest
/**
apiVersion: apps/v1
kind: Deployment
metadata:
	namespace: default
	name: deploy
**/
func GetManifestSimpleHeadFromRelease(rl *release.Release, releaseNamespace string) ([]SimpleHead, error) {
	manifest := rl.Manifest
	for _, v := range rl.Hooks {
		manifest += "\n---\n" + v.Manifest
	}
	manifests := releaseutil.SplitManifests(manifest)
	heads := make([]SimpleHead, 0)
	for i := range manifests {
		content := manifests[i]
		var entry SimpleHead
		if err := yaml.Unmarshal([]byte(content), &entry); err != nil {
			blog.Errorf("YAML parse error, %s", err)
			return nil, err
		}
		if entry.Metadata == nil {
			continue
		}
		if entry.Metadata.Namespace == "" {
			entry.Metadata.Namespace = releaseNamespace
		}
		heads = append(heads, entry)
	}
	return heads, nil
}
