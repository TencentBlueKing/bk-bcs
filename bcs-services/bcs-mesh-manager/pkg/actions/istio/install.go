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
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/cmd/mesh-manager/options"
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
	req         *meshmanager.IstioRequest
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
	req *meshmanager.IstioRequest,
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
	// 必填字段
	if i.req.ProjectCode.GetValue() == "" && i.req.ProjectID.GetValue() == "" {
		return fmt.Errorf("project is required")
	}
	if len(i.req.PrimaryClusters) == 0 {
		return fmt.Errorf("clusters is required")
	}
	if i.req.Version.GetValue() == "" {
		return fmt.Errorf("chart version is required")
	}
	if i.req.FeatureConfigs == nil {
		return fmt.Errorf("feature configs is required")
	}
	// 检查resource参数（limit和request 合法，并且limit >= request）
	if err := i.validateResource(i.req); err != nil {
		blog.Errorf("validate resource failed, err: %s", err)
		return err
	}

	// 检查主从集群版本兼容性
	allClusters := make([]string, 0, len(i.req.PrimaryClusters)+len(i.req.RemoteClusters))
	allClusters = append(allClusters, i.req.PrimaryClusters...)
	allClusters = append(allClusters, i.req.RemoteClusters...)
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
	return nil
}

func (i *InstallIstioAction) validateResource(req *meshmanager.IstioRequest) error {
	// 检查sidecar resource参数（limit和request 合法，并且limit >= request）
	if req.SidecarResourceConfig == nil {
		return nil
	}
	if err := utils.ValidateResourceLimit(
		req.SidecarResourceConfig.CpuRequest.GetValue(),
		req.SidecarResourceConfig.CpuLimit.GetValue(),
	); err != nil {
		return err
	}
	if err := utils.ValidateResourceLimit(
		req.SidecarResourceConfig.MemoryRequest.GetValue(),
		req.SidecarResourceConfig.MemoryLimit.GetValue(),
	); err != nil {
		return err
	}
	// 检查hpa中resource参数（limit和request 合法，并且limit >= request）
	if req.HighAvailability == nil {
		return nil
	}
	if req.HighAvailability.ResourceConfig == nil {
		return nil
	}
	if err := utils.ValidateResourceLimit(
		req.HighAvailability.ResourceConfig.CpuRequest.GetValue(),
		req.HighAvailability.ResourceConfig.CpuLimit.GetValue(),
	); err != nil {
		return err
	}
	if err := utils.ValidateResourceLimit(
		req.HighAvailability.ResourceConfig.MemoryRequest.GetValue(),
		req.HighAvailability.ResourceConfig.MemoryLimit.GetValue(),
	); err != nil {
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

	chartVersion, err := i.getIstioChartVersion(i.req.Version.GetValue())
	if err != nil {
		blog.Errorf("get istio chart version failed, err: %s", err)
		return common.NewCodeMessageError(common.InnerErrorCode, "get istio chart version failed", err)
	}
	meshIstio.ChartVersion = chartVersion
	// 状态设置为安装中
	meshIstio.Status = common.IstioStatusInstalling

	// 写入DB，状态更新为安装中
	err = i.model.Create(ctx, meshIstio)
	if err != nil {
		blog.Errorf("create mesh istio failed, err: %s", err)
		return common.NewCodeMessageError(common.DBErrorCode, "create mesh istio failed", err)
	}

	// 创建并开始任务
	action := actions.NewIstioInstallAction(
		&common.IstioInstallOption{
			ChartValuesPath: i.istioConfig.ChartValuesPath,
			ChartRepo:       i.istioConfig.ChartRepo,
			MeshID:          meshID,
			NetworkID:       networkID,
			ChartVersion:    chartVersion,

			ProjectID:             i.req.ProjectID.GetValue(),
			ProjectCode:           i.req.ProjectCode.GetValue(),
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
