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

// Package actions 操作包
package actions

import (
	"context"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/helmmanager"
	"k8s.io/utils/pointer"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/clients/helm"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/clients/k8s"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/operation"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/store/entity"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/utils"
	meshmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/proto/bcs-mesh-manager"
)

// IstioInstallOption istio安装操作选项
type IstioInstallOption struct {
	Model store.MeshManagerModel

	ChartValuesPath string
	ChartRepo       string

	ProjectID             string
	ProjectCode           string
	Name                  string
	Description           string
	Version               string
	ControlPlaneMode      string
	ClusterMode           string
	PrimaryClusters       []string
	RemoteClusters        []string
	SidecarResourceConfig *meshmanager.ResourceConfig
	HighAvailability      *meshmanager.HighAvailability
	LogCollectorConfig    *meshmanager.LogCollectorConfig
	TracingConfig         *meshmanager.TracingConfig
	FeatureConfigs        map[string]*meshmanager.FeatureConfig

	MeshID       string
	NetworkID    string
	ChartVersion string
}

// IstioInstallAction istio安装操作
type IstioInstallAction struct {
	*IstioInstallOption
}

var _ operation.Operation = &IstioInstallAction{}

// NewIstioInstallAction 创建istio安装操作
func NewIstioInstallAction(opt *IstioInstallOption) *IstioInstallAction {
	return &IstioInstallAction{
		IstioInstallOption: opt,
	}
}

// Action 操作名称
func (i *IstioInstallAction) Action() string {
	return "istio-install"
}

// Name 操作实例名称
func (i *IstioInstallAction) Name() string {
	return fmt.Sprintf("istio-install-%s", i.MeshID)
}

// Validate 验证参数
func (i *IstioInstallAction) Validate() error {
	// 必填字段
	if i.ProjectCode == "" && i.ProjectID == "" {
		return fmt.Errorf("project is required")
	}
	if len(i.PrimaryClusters) == 0 {
		return fmt.Errorf("clusters is required")
	}
	if i.Version == "" {
		return fmt.Errorf("chart version is required")
	}
	if i.ChartVersion == "" {
		return fmt.Errorf("chart version is required")
	}
	if i.FeatureConfigs == nil {
		return fmt.Errorf("feature configs is required")
	}
	return nil
}

// Prepare 准备阶段
func (i *IstioInstallAction) Prepare(ctx context.Context) error {
	blog.Infof("prepare istio install for mesh %s", i.MeshID)
	// 这里可以做一些准备工作
	return nil
}

// Execute 执行安装
func (i *IstioInstallAction) Execute(ctx context.Context) error {
	blog.Infof("execute istio install for mesh %s", i.MeshID)

	// 安装主集群中的istio
	for _, cluster := range i.PrimaryClusters {
		if err := i.installIstioForPrimary(ctx, i.ChartVersion, cluster); err != nil {
			blog.Errorf("install istio for primary cluster %s failed, err: %s", cluster, err)
			return fmt.Errorf("install istio for primary cluster %s failed: %s", cluster, err)
		}
	}

	// TODO: 安装远程集群中的istio
	// 1、主集群中先安装egress gateway，获取到clb
	// 2、远程集群中安装istio，使用主集群的clb

	blog.Infof("istio install completed for mesh %s", i.MeshID)
	return nil
}

// Done 完成回调
func (i *IstioInstallAction) Done(err error) {
	m := make(entity.M)
	if err != nil {
		blog.Errorf("istio install failed for mesh %s, err: %s", i.MeshID, err)
		m[entity.FieldKeyStatus] = common.IstioStatusFailed
	} else {
		blog.Infof("istio install success for mesh %s", i.MeshID)
		m[entity.FieldKeyStatus] = common.IstioStatusRunning
	}
	updateErr := i.Model.Update(context.TODO(), i.MeshID, m)
	if updateErr != nil {
		blog.Errorf("update mesh status failed for mesh %s, err: %s", i.MeshID, updateErr)
	}
}

// installIstioForPrimary 为主集群安装istio
func (i *IstioInstallAction) installIstioForPrimary(ctx context.Context, chartVersion, clusterID string) error {
	// 创建 istio-system 命名空间,如果已经存在则忽略
	exist, err := k8s.CheckNamespaceExist(ctx, clusterID, common.IstioNamespace)
	if err != nil {
		blog.Errorf("check namespace %s exist failed, err: %s", common.IstioNamespace, err)
		return fmt.Errorf("check namespace exist failed: %s", err)
	}
	// 不存在则创建
	if !exist {
		if createErr := k8s.CreateNamespace(ctx, clusterID, common.IstioNamespace); createErr != nil {
			blog.Errorf("create namespace %s failed, err: %s", common.IstioNamespace, createErr)
			return fmt.Errorf("create namespace failed: %s", createErr)
		}
	}

	// 安装istio base
	if err := i.installIstioBase(ctx, chartVersion, clusterID); err != nil {
		return fmt.Errorf("install istio base failed: %s", err)
	}

	// 安装istiod
	if err := i.installIstiod(ctx, chartVersion, clusterID); err != nil {
		return fmt.Errorf("install istiod failed: %s", err)
	}

	return nil
}

// installIstioBase 安装istio base组件
func (i *IstioInstallAction) installIstioBase(ctx context.Context, chartVersion, clusterID string) error {
	baseValues, err := utils.GenBaseValues(i.ChartValuesPath, chartVersion, clusterID, i.MeshID, i.NetworkID)
	if err != nil {
		return fmt.Errorf("gen base values failed: %s", err)
	}
	blog.Infof("install istio base values: %s for cluster: %s, mesh: %s, network: %s",
		baseValues, clusterID, i.MeshID, i.NetworkID)

	resp, err := helm.Install(ctx, &helmmanager.InstallReleaseV1Req{
		ProjectCode: pointer.String(i.ProjectCode),
		ClusterID:   pointer.String(clusterID),
		Name:        pointer.String(common.IstioInstallBaseName),
		Namespace:   pointer.String(common.IstioNamespace),
		Chart:       pointer.String(common.ComponentIstioBase),
		Repository:  pointer.String(i.ChartRepo),
		Version:     pointer.String(chartVersion),
		Values:      []string{baseValues},
		Args:        []string{"--wait"},
	})
	blog.Infof("install istio base resp: %+v", resp)
	if err != nil {
		blog.Errorf("install istio base failed, err: %s", err)
		return fmt.Errorf("install istio base failed: %s", err)
	}
	if resp.Result != nil && !*resp.Result {
		blog.Errorf("install istio base failed, err: %s", *resp.Message)
		return fmt.Errorf("install istio base failed: %s", *resp.Message)
	}
	// 查询是否安装成功 查询详情 每隔10s查询一次 直到安装成功，超时2min
	timeout := time.NewTimer(2 * time.Minute)
	defer timeout.Stop()
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeout.C:
			blog.Errorf("install istio base timeout for cluster %s", clusterID)
			return fmt.Errorf("install istio base timeout for cluster %s", clusterID)
		case <-ticker.C:
			// 查询安装状态
			release, err := helm.GetReleaseDetail(ctx, &helmmanager.GetReleaseDetailV1Req{
				ProjectCode: pointer.String(i.ProjectCode),
				ClusterID:   pointer.String(clusterID),
				Name:        pointer.String(common.IstioInstallBaseName),
				Namespace:   pointer.String(common.IstioNamespace),
			})
			blog.Infof("[loop]get istio base release: %+v, err: %s, cluster: %s", release, err, clusterID)
			if err != nil {
				blog.Errorf("get istio base release failed, err: %s", err)
				return fmt.Errorf("get istio base release failed: %s", err)
			}
			if release.Data != nil && release.Data.Status != nil {
				if *release.Data.Status == helm.ReleaseStatusDeployed {
					blog.Infof("install istio base success for cluster %s", clusterID)
					return nil
				}
			}
		}
	}
}

// installIstiod 安装istiod组件
func (i *IstioInstallAction) installIstiod(ctx context.Context, chartVersion, clusterID string) error {
	istiodValues, err := utils.GenIstiodValues(
		i.ChartValuesPath,
		common.IstioInstallModePrimary,
		chartVersion,
		clusterID,
		"",
		clusterID,
		i.MeshID,
		i.NetworkID,
		i.FeatureConfigs,
	)
	if err != nil {
		return fmt.Errorf("gen istiod values failed: %s", err)
	}
	blog.Infof("install istiod values: %s for cluster: %s, mesh: %s, network: %s",
		istiodValues, clusterID, i.MeshID, i.NetworkID)

	resp, err := helm.Install(ctx, &helmmanager.InstallReleaseV1Req{
		ProjectCode: pointer.String(i.ProjectCode),
		ClusterID:   pointer.String(clusterID),
		Name:        pointer.String(common.IstioInstallIstiodName),
		Namespace:   pointer.String(common.IstioNamespace),
		Chart:       pointer.String(common.ComponentIstiod),
		Repository:  pointer.String(i.ChartRepo),
		Version:     pointer.String(chartVersion),
		Values:      []string{istiodValues},
		Args:        []string{"--wait"},
	})
	if err != nil {
		blog.Errorf("install istiod failed, err: %s", err)
		return fmt.Errorf("install istiod failed: %s", err)
	}
	if resp.Result != nil && !*resp.Result {
		blog.Errorf("install istiod failed, err: %s", *resp.Message)
		return fmt.Errorf("install istiod failed: %s", *resp.Message)
	}

	// 查询是否安装成功 查询详情 每隔10s查询一次 直到安装成功，超时2min
	timeout := time.NewTimer(2 * time.Minute)
	defer timeout.Stop()
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeout.C:
			blog.Errorf("install istiod timeout for cluster %s", clusterID)
			return fmt.Errorf("install istiod timeout for cluster %s", clusterID)
		case <-ticker.C:
			// 查询安装状态
			release, err := helm.GetReleaseDetail(ctx, &helmmanager.GetReleaseDetailV1Req{
				ProjectCode: pointer.String(i.ProjectCode),
				ClusterID:   pointer.String(clusterID),
				Name:        pointer.String(common.IstioInstallIstiodName),
				Namespace:   pointer.String(common.IstioNamespace),
			})
			blog.Infof("[loop]get istiod release: %+v, err: %s, cluster: %s", release, err, clusterID)
			if err != nil {
				blog.Errorf("get istiod release failed, err: %s", err)
				return fmt.Errorf("get istiod release failed: %s", err)
			}
			if release.Data != nil && release.Data.Status != nil {
				if *release.Data.Status == helm.ReleaseStatusDeployed {
					blog.Infof("install istiod success for cluster %s", clusterID)
					return nil
				}
			}
		}
	}
}
