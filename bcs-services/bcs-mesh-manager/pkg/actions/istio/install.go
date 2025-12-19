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

package istio

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/cmd/mesh-manager/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/auth"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/operation"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/operation/actions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/store/entity"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/utils"
	meshmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/proto/bcs-mesh-manager"
)

// InstallIstioAction action for installing istio
type InstallIstioAction struct {
	istioConfig *options.IstioConfig
	model       store.MeshManagerModel
	req         *meshmanager.IstioInstallRequest
	resp        *meshmanager.InstallIstioResponse
}

// NewInstallIstioAction create install istio action
func NewInstallIstioAction(istioConfig *options.IstioConfig, model store.MeshManagerModel) *InstallIstioAction {
	return &InstallIstioAction{
		istioConfig: istioConfig,
		model:       model,
	}
}

// Handle handles the install istio request
func (i *InstallIstioAction) Handle(
	ctx context.Context,
	req *meshmanager.IstioInstallRequest,
	resp *meshmanager.InstallIstioResponse,
) error {

	i.req = req
	i.resp = resp

	// 校验请求参数
	if err := i.Validate(); err != nil {
		i.setResp(common.InvalidRequestErrorCode, err.Error())
		return nil
	}

	// 执行安装
	if err := i.install(ctx); err != nil {
		if customErr, ok := err.(*common.CodeMessageError); ok {
			i.setResp(customErr.GetCode(), customErr.GetMessageWithErr())
		} else {
			i.setResp(common.DBErrorCode, err.Error())
		}
		return nil
	}

	i.setResp(common.SuccessCode, "")
	return nil
}

// Validate 验证请求参数
func (i *InstallIstioAction) Validate() error {
	// 校验项目信息
	if i.req.ProjectCode == "" {
		return fmt.Errorf("项目编码或项目 ID 不能为空")
	}

	// 校验主集群
	if len(i.req.PrimaryClusters) == 0 {
		return fmt.Errorf("主集群不能为空")
	}

	// 校验版本
	if i.req.Version.GetValue() == "" {
		return fmt.Errorf("chart version 不能为空")
	}

	// 校验特性配置
	if i.req.FeatureConfigs == nil {
		return fmt.Errorf("特性配置不能为空")
	}

	// 网格名称不能为空
	if i.req.Name.GetValue() == "" {
		return fmt.Errorf("网格名称不能为空")
	}

	// 网格名称不能仅为空格
	if strings.TrimSpace(i.req.Name.GetValue()) == "" {
		return fmt.Errorf("网格名称不能仅为空格")
	}
	// 检查resource参数
	if err := utils.ValidateResource(i.req.SidecarResourceConfig); err != nil {
		blog.Errorf("validate resource failed, err: %s", err)
		return fmt.Errorf("sidecar资源配置验证失败: %w", err)
	}
	if i.req.HighAvailability != nil {
		if err := utils.ValidateResource(i.req.HighAvailability.ResourceConfig); err != nil {
			blog.Errorf("validate resource failed, err: %s", err)
			return fmt.Errorf("高可用资源配置验证失败: %w", err)
		}
	}

	// 检查主从集群版本兼容性
	remoteClusters := make([]string, 0, len(i.req.RemoteClusters))
	for _, cluster := range i.req.RemoteClusters {
		remoteClusters = append(remoteClusters, cluster.ClusterID)
	}
	allClusters := utils.MergeSlices(i.req.PrimaryClusters, remoteClusters)
	if err := utils.ValidateClusterVersion(
		context.TODO(),
		i.istioConfig,
		allClusters,
		i.req.Version.GetValue(),
	); err != nil {
		return err
	}

	// 检查集群中是否已经安装了istio
	if err := utils.ValidateIstioInstalled(context.TODO(), allClusters); err != nil {
		return err
	}
	// 检查可观测性配置是否配置正确
	if err := utils.ValidateObservabilityConfig(i.req.ObservabilityConfig); err != nil {
		return err
	}
	// 检查高可用配置是否正确
	if err := utils.ValidateHighAvailabilityConfig(i.req.HighAvailability); err != nil {
		return err
	}
	return nil
}

// setResp sets the response with code and message
func (i *InstallIstioAction) setResp(code uint32, message string) {
	i.resp.Code = code
	i.resp.Message = message
}

// install implements the business logic for installing mesh istio
func (i *InstallIstioAction) install(ctx context.Context) error {
	// 创建 Mesh 实体并转换
	meshIstio := &entity.MeshIstio{}
	meshIstio.TransferFromProto(i.req)
	meshID := utils.GenMeshID()
	networkID := utils.GenNetworkID()
	meshIstio.MeshID = meshID
	meshIstio.NetworkID = networkID
	meshIstio.ReleaseNames = i.buildReleaseNames()
	meshIstio.CreateBy = auth.GetUserFromCtx(ctx)
	meshIstio.CreateTime = time.Now().UnixMilli()
	chartVersion, err := i.getIstioChartVersion(i.req.Version.GetValue())
	if err != nil {
		blog.Errorf("get istio chart version failed, err: %s", err)
		return common.NewCodeMessageError(common.InnerErrorCode, "get istio chart version failed", err)
	}
	meshIstio.ChartVersion = chartVersion
	// 状态设置为安装中
	meshIstio.Status = common.IstioStatusInstalling
	revision := ""
	if i.req.Revision.GetValue() != "" {
		revision = i.req.Revision.GetValue()
	} else {
		revision = common.GetRevision(chartVersion)
	}
	meshIstio.Revision = revision
	// 写入DB，状态更新为安装中
	err = i.model.Create(ctx, meshIstio)
	if err != nil {
		blog.Errorf("create mesh istio failed, err: %s", err)
		return common.NewCodeMessageError(common.DBErrorCode, "create mesh istio failed", err)
	}
	// 创建并开始任务
	action := actions.NewIstioInstallAction(
		&common.IstioInstallOption{
			ChartValuesPath:       i.istioConfig.ChartValuesPath,
			ChartRepo:             i.istioConfig.ChartRepo,
			MeshID:                meshID,
			NetworkID:             networkID,
			ChartVersion:          chartVersion,
			ProjectCode:           i.req.ProjectCode,
			Name:                  i.req.Name.GetValue(),
			Description:           i.req.Description.GetValue(),
			Version:               i.req.Version.GetValue(),
			ControlPlaneMode:      i.req.ControlPlaneMode.GetValue(),
			ClusterMode:           i.req.ClusterMode.GetValue(),
			PrimaryClusters:       i.req.PrimaryClusters,
			RemoteClusters:        i.req.RemoteClusters,
			SidecarResourceConfig: i.req.SidecarResourceConfig,
			HighAvailability:      i.req.HighAvailability,
			ObservabilityConfig:   i.req.ObservabilityConfig,
			FeatureConfigs:        i.req.FeatureConfigs,
			MultiClusterEnabled:   i.req.MultiClusterEnabled.GetValue(),
			CLBID:                 i.req.ClbID.GetValue(),
			Revision:              revision,
		},
		i.model,
	)
	// 异步执行，10分钟超时
	_, err = operation.GlobalOperator.Dispatch(action, 10*time.Minute)
	if err != nil {
		blog.Errorf("dispatch istio install action failed, err: %s", err)
		return common.NewCodeMessageError(common.InstallIstioErrorCode, "dispatch istio install action failed", err)
	}

	// 返回安装结果
	i.resp.Code = common.SuccessCode
	i.resp.Message = "安装进行中"
	i.resp.MeshID = meshID

	// Create mesh in database
	return nil
}

func (i *InstallIstioAction) getIstioChartVersion(version string) (string, error) {
	// 获取版本配置并输出
	if i.istioConfig == nil {
		return "", errors.New("istio config is nil")
	}
	// 根据版本获取最新的一个chartVersion
	if i.istioConfig.IstioVersions[version] == nil {
		return "", errors.New("version not found")
	}
	return i.istioConfig.IstioVersions[version].ChartVersion, nil
}

// buildReleaseNames 构建ReleaseNames map
func (i *InstallIstioAction) buildReleaseNames() map[string]map[string]string {
	releaseNames := make(map[string]map[string]string)

	remoteClusters := make([]string, 0, len(i.req.RemoteClusters))
	for _, cluster := range i.req.RemoteClusters {
		remoteClusters = append(remoteClusters, cluster.ClusterID)
	}
	allClusters := utils.MergeSlices(i.req.PrimaryClusters, remoteClusters)

	for _, clusterID := range allClusters {
		releaseNames[clusterID] = map[string]string{
			common.ComponentIstioBase: common.IstioInstallBaseName,
			common.ComponentIstiod:    common.IstioInstallIstiodName,
		}
		// 开启多集群模式，则存储东西向网关的release name
		if i.req.MultiClusterEnabled.GetValue() {
			releaseNames[clusterID][common.ComponentIstioGateway] = common.IstioInstallIstioGatewayName
		}
	}

	return releaseNames
}
