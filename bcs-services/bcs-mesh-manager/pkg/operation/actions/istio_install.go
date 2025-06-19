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
)

// IstioInstallAction istio安装操作
type IstioInstallAction struct {
	model store.MeshManagerModel

	*common.IstioInstallOption
}

var _ operation.Operation = &IstioInstallAction{}

// NewIstioInstallAction 创建istio安装操作
func NewIstioInstallAction(opt *common.IstioInstallOption, model store.MeshManagerModel) *IstioInstallAction {
	return &IstioInstallAction{
		IstioInstallOption: opt,
		model:              model,
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
	blog.Infof("[%s]prepare istio install", i.MeshID)
	// 这里可以做一些准备工作
	return nil
}

// Execute 执行安装
func (i *IstioInstallAction) Execute(ctx context.Context) error {
	blog.Infof("[%s]execute istio install", i.MeshID)

	// 安装主集群中的istio
	for _, cluster := range i.PrimaryClusters {
		if err := i.installIstioForPrimary(ctx, i.ChartVersion, cluster); err != nil {
			blog.Errorf("[%s]install istio for primary cluster %s failed, err: %s", i.MeshID, cluster, err)
			return fmt.Errorf("install istio for primary cluster %s failed: %s", cluster, err)
		}
	}

	// TODO: 安装远程集群中的istio
	// 1、主集群中先安装egress gateway，获取到clb
	// 2、远程集群中安装istio，使用主集群的clb

	blog.Infof("[%s]istio install completed", i.MeshID)
	return nil
}

// Done 完成回调
func (i *IstioInstallAction) Done(err error) {
	m := make(entity.M)
	if err != nil {
		blog.Errorf("[%s]istio install failed, err: %s", i.MeshID, err)
		m[entity.FieldKeyStatus] = common.IstioStatusFailed
		m[entity.FieldKeyStatusMessage] = err.Error()
	} else {
		blog.Infof("[%s]istio install success", i.MeshID)
		m[entity.FieldKeyStatus] = common.IstioStatusRunning
	}
	updateErr := i.model.Update(context.TODO(), i.MeshID, m)
	if updateErr != nil {
		blog.Errorf("[%s]update mesh status failed, err: %s", i.MeshID, updateErr)
	}
}

// installIstioForPrimary 为主集群安装istio
func (i *IstioInstallAction) installIstioForPrimary(ctx context.Context, chartVersion, clusterID string) error {
	// 创建 istio-system 命名空间,如果已经存在则忽略
	exist, err := k8s.CheckNamespaceExist(ctx, clusterID, common.IstioNamespace)
	if err != nil {
		blog.Errorf("[%s]check namespace %s exist failed, err: %s", i.MeshID, common.IstioNamespace, err)
		return fmt.Errorf("check namespace exist failed: %s", err)
	}
	// 不存在则创建
	if !exist {
		if createErr := k8s.CreateNamespace(ctx, clusterID, common.IstioNamespace); createErr != nil {
			blog.Errorf("[%s]create namespace %s failed, err: %s", i.MeshID, common.IstioNamespace, createErr)
			return fmt.Errorf("create namespace failed: %s", createErr)
		}
	}

	// 安装istio base
	if err := i.installComponent(
		ctx,
		chartVersion,
		clusterID,
		common.IstioInstallBaseName,
		common.ComponentIstioBase,
		func() (string, error) {
			return utils.GenBaseValues(i.IstioInstallOption)
		},
	); err != nil {
		return fmt.Errorf("install istio base failed: %s", err)
	}

	// 安装istiod
	if err := i.installComponent(
		ctx,
		chartVersion,
		clusterID,
		common.IstioInstallIstiodName,
		common.ComponentIstiod,
		func() (string, error) {
			return utils.GenIstiodValues(common.IstioInstallModePrimary, "", i.IstioInstallOption)
		},
	); err != nil {
		return fmt.Errorf("install istiod failed: %s", err)
	}

	return nil
}

// installComponent 通用安装istio组件方法
func (i *IstioInstallAction) installComponent(
	ctx context.Context,
	chartVersion, clusterID, componentName, chartName string,
	valuesGenFunc func() (string, error),
) error {
	values, err := valuesGenFunc()
	if err != nil {
		return fmt.Errorf("gen %s values failed: %s", componentName, err)
	}
	blog.Infof("install %s values: %s for cluster: %s, mesh: %s, network: %s",
		componentName, values, clusterID, i.MeshID, i.NetworkID)

	resp, err := helm.Install(ctx, &helmmanager.InstallReleaseV1Req{
		ProjectCode: pointer.String(i.ProjectCode),
		ClusterID:   pointer.String(clusterID),
		Name:        pointer.String(componentName),
		Namespace:   pointer.String(common.IstioNamespace),
		Chart:       pointer.String(chartName),
		Repository:  pointer.String(i.ChartRepo),
		Version:     pointer.String(chartVersion),
		Values:      []string{values},
		Args:        []string{"--wait"},
	})
	if err != nil {
		blog.Errorf("install %s failed, err: %s", componentName, err)
		return fmt.Errorf("install %s failed: %s", componentName, err)
	}
	if resp.Result != nil && !*resp.Result {
		blog.Errorf("install %s failed, err: %s", componentName, *resp.Message)
		return fmt.Errorf("install %s failed: %s", componentName, *resp.Message)
	}
	// 查询是否安装成功
	timeout := time.NewTimer(2 * time.Minute)
	defer timeout.Stop()
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeout.C:
			blog.Errorf("install %s timeout for cluster %s", componentName, clusterID)
			return fmt.Errorf("install %s timeout for cluster %s", componentName, clusterID)
		case <-ticker.C:
			release, err := helm.GetReleaseDetail(ctx, &helmmanager.GetReleaseDetailV1Req{
				ProjectCode: pointer.String(i.ProjectCode),
				ClusterID:   pointer.String(clusterID),
				Name:        pointer.String(componentName),
				Namespace:   pointer.String(common.IstioNamespace),
			})
			blog.Infof("[loop]get %s release: %+v, err: %s, cluster: %s", componentName, release, err, clusterID)
			if err != nil {
				blog.Errorf("get %s release failed, err: %s", componentName, err)
				return fmt.Errorf("get %s release failed: %s", componentName, err)
			}
			if release.Data != nil && release.Data.Status != nil {
				if *release.Data.Status == helm.ReleaseStatusDeployed {
					blog.Infof("install %s success for cluster %s", componentName, clusterID)
					return nil
				}
			}
		}
	}
}
