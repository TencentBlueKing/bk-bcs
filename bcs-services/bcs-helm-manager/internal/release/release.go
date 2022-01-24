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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/common"
	helmmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/proto/bcs-helm-manager"
)

// Handler 定义了 helm release 的client集合
type Handler interface {
	Cluster(clusterID string) Cluster
}

// Cluster 定义了每个 helm release client 的操作能力, 用于直接与集群产生helm命令交互
type Cluster interface {
	List(ctx context.Context, option ListOption) (int, []*Release, error)
	Install(ctx context.Context, conf HelmInstallConfig) (*HelmInstallResult, error)
	Uninstall(ctx context.Context, conf HelmUninstallConfig) (*HelmUninstallResult, error)
	Upgrade(ctx context.Context, conf HelmUpgradeConfig) (*HelmUpgradeResult, error)
	Rollback(ctx context.Context, conf HelmRollbackConfig) (*HelmRollbackResult, error)
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
}

// Transfer2Proto transfer the data into protobuf struct
func (r *Release) Transfer2Proto() *helmmanager.Release {
	return &helmmanager.Release{
		Name:         common.GetStringP(r.Name),
		Namespace:    common.GetStringP(r.Namespace),
		Revision:     common.GetUint32P(uint32(r.Revision)),
		Status:       common.GetStringP(r.Status),
		Chart:        common.GetStringP(r.Chart),
		ChartVersion: common.GetStringP(r.ChartVersion),
		AppVersion:   common.GetStringP(r.AppVersion),
		UpdateTime:   common.GetStringP(r.UpdateTime),
	}
}

// Transfer2DetailProto transfer the data into detail protobuf struct
func (r *Release) Transfer2DetailProto() *helmmanager.ReleaseDetail {
	return &helmmanager.ReleaseDetail{
		Name:         common.GetStringP(r.Name),
		Namespace:    common.GetStringP(r.Namespace),
		Revision:     common.GetUint32P(uint32(r.Revision)),
		Status:       common.GetStringP(r.Status),
		Chart:        common.GetStringP(r.Chart),
		ChartVersion: common.GetStringP(r.ChartVersion),
		AppVersion:   common.GetStringP(r.AppVersion),
		UpdateTime:   common.GetStringP(r.UpdateTime),
	}
}

// Config 定义了 Handler 的配置参数
type Config struct {
	APIServer string
	Token     string

	KubeConfigTemplate string
	HelmBinary         string

	PatchTemplates []*File
	VarTemplates   []*File
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
	DryRun bool

	Name      string
	Namespace string

	Args []string

	Chart               *File
	Values              []*File
	PatchTemplateValues map[string]string
	VarTemplateValues   map[string]string
}

// HelmInstallResult 定义了helm执行install的返回结果
type HelmInstallResult struct {
	Revision   int
	Status     string
	AppVersion string
	UpdateTime string
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

	Name      string
	Namespace string

	Args []string

	Chart               *File
	Values              []*File
	PatchTemplateValues map[string]string
	VarTemplateValues   map[string]string
}

// HelmUpgradeResult 定义了helm执行upgrade时的返回结果
type HelmUpgradeResult struct {
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
