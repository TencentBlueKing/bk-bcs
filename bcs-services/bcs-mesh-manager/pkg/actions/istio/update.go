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
	"fmt"
	"slices"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/cmd/mesh-manager/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/operation"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/operation/actions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/store/entity"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/utils"
	meshmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/proto/bcs-mesh-manager"
)

// UpdateIstioAction handles istio update request
type UpdateIstioAction struct {
	istioConfig *options.IstioConfig
	model       store.MeshManagerModel
	req         *meshmanager.IstioRequest
	resp        *meshmanager.UpdateIstioResponse
}

// NewUpdateIstioAction create update istio action
func NewUpdateIstioAction(istioConfig *options.IstioConfig, model store.MeshManagerModel) *UpdateIstioAction {
	return &UpdateIstioAction{
		istioConfig: istioConfig,
		model:       model,
	}
}

// Handle processes the istio update request
func (u *UpdateIstioAction) Handle(
	ctx context.Context,
	req *meshmanager.IstioRequest,
	resp *meshmanager.UpdateIstioResponse,
) error {
	u.req = req
	u.resp = resp

	if err := u.validate(req); err != nil {
		blog.Errorf("update mesh failed, invalid request, %s, param: %v", err.Error(), u.req)
		u.setResp(common.ParamErrorCode, err.Error())
		return nil
	}
	if err := u.update(ctx); err != nil {
		blog.Errorf("update mesh failed, %s, meshID: %s", err.Error(), u.req.MeshID)
		u.setResp(common.DBErrorCode, err.Error())
		return nil
	}

	u.setResp(common.SuccessCode, "")
	blog.Infof("update mesh successfully, meshID: %s", u.req.MeshID)
	return nil
}

// setResp sets the response with code and message
func (u *UpdateIstioAction) setResp(code uint32, message string) {
	u.resp.Code = code
	u.resp.Message = message
}

func (u *UpdateIstioAction) validate(req *meshmanager.IstioRequest) error {
	// 必填字段验证
	if req.MeshID == nil {
		return fmt.Errorf("meshID is required")
	}
	if req.ProjectID == nil {
		return fmt.Errorf("projectID is required")
	}
	if req.Version.GetValue() == "" {
		return fmt.Errorf("chart version is required")
	}
	if req.FeatureConfigs == nil {
		return fmt.Errorf("feature configs is required")
	}
	// 检查可观测性配置是否配置正确
	if err := utils.ValidateObservabilityConfig(req.ObservabilityConfig); err != nil {
		return err
	}

	return nil
}

// update implements the business logic for updating mesh
func (u *UpdateIstioAction) update(ctx context.Context) error {
	// 获取istio信息
	istio, err := u.model.Get(ctx, operator.NewLeafCondition(operator.Eq, operator.M{
		entity.FieldKeyMeshID:    u.req.MeshID.GetValue(),
		entity.FieldKeyProjectID: u.req.ProjectID.GetValue(),
	}))
	if err != nil {
		blog.Errorf("get mesh istio failed, meshID: %s, err: %s", u.req.MeshID.GetValue(), err)
		return err
	}
	// 主从集群信息使用db中的，不可更新，单独接口处理集群更新的情况
	u.req.PrimaryClusters = istio.PrimaryClusters
	u.req.RemoteClusters = istio.RemoteClusters

	// 更新 istio 的状态为更新中
	err = u.model.Update(ctx, u.req.MeshID.GetValue(), entity.M{
		entity.FieldKeyStatus: common.IstioStatusUpdating,
	})
	if err != nil {
		blog.Errorf("update mesh status failed, meshID: %s, err: %s", u.req.MeshID.GetValue(), err)
		return err
	}
	// 获取更新的字段
	updateFields := u.buildUpdateFields(u.req)
	blog.Infof("update fields: %+v for meshID: %s", updateFields, u.req.MeshID.GetValue())

	// 构建更新配置
	updateValues, err := utils.ConvertRequestToValues(istio.Version, u.req)
	if err != nil {
		blog.Errorf("convert request to values failed, meshID: %s, err: %s", u.req.MeshID.GetValue(), err)
		return err
	}
	blog.Infof("update values: %+v for meshID: %s", updateValues, u.req.MeshID.GetValue())

	// 异步更新istio
	action := actions.NewIstioUpdateAction(
		&actions.IstioUpdateOption{
			Model:               u.model,
			ProjectCode:         &istio.ProjectCode,
			MeshID:              &istio.MeshID,
			ChartName:           common.ComponentIstiod,
			ChartVersion:        &istio.ChartVersion,
			ChartRepo:           &u.istioConfig.ChartRepo,
			PrimaryClusters:     istio.PrimaryClusters,
			RemoteClusters:      istio.RemoteClusters,
			UpdateFields:        updateFields,
			UpdateValues:        updateValues,
			ObservabilityConfig: u.req.ObservabilityConfig,
			Version:             istio.Version,
		},
	)
	_, err = operation.GlobalOperator.Dispatch(action, 10*time.Minute)
	if err != nil {
		blog.Errorf("dispatch istio update action failed, err: %s", err)
		return common.NewCodeMessageError(common.InnerErrorCode, "dispatch istio update action failed", err)
	}

	return nil
}

// 构建更新字段
func (u *UpdateIstioAction) buildUpdateFields(req *meshmanager.IstioRequest) entity.M {
	updateFields := entity.M{}
	updateFields = buildBasicFields(req, updateFields)
	updateFields = buildResourceConfigs(req, updateFields)
	updateFields = buildHighAvailability(req, updateFields)
	updateFields = buildFeatureConfigs(req, updateFields)
	updateFields = buildObservability(req, updateFields)
	return updateFields
}

// buildBasicFields builds basic fields from request
func buildBasicFields(req *meshmanager.IstioRequest, updateFields entity.M) entity.M {
	if req.Description != nil {
		updateFields[entity.FieldKeyDescription] = req.Description.GetValue()
	}
	if req.Name != nil {
		updateFields[entity.FieldKeyName] = req.Name.GetValue()
	}
	if req.ControlPlaneMode != nil {
		updateFields[entity.FieldKeyControlPlaneMode] = req.ControlPlaneMode.GetValue()
	}
	if req.ClusterMode != nil {
		updateFields[entity.FieldKeyClusterMode] = req.ClusterMode.GetValue()
	}
	if req.DifferentNetwork != nil {
		updateFields[entity.FieldKeyDifferentNetwork] = req.DifferentNetwork.GetValue()
	}

	return updateFields
}

// buildResourceConfigs builds resource related configurations
func buildResourceConfigs(req *meshmanager.IstioRequest, updateFields entity.M) entity.M {
	// Update Sidecar resource config
	if req.SidecarResourceConfig != nil {
		if req.SidecarResourceConfig.CpuRequest != nil {
			updateFields[entity.DotKeySidecarCPURequest] = req.SidecarResourceConfig.CpuRequest.GetValue()
		}
		if req.SidecarResourceConfig.CpuLimit != nil {
			updateFields[entity.DotKeySidecarCPULimit] = req.SidecarResourceConfig.CpuLimit.GetValue()
		}
		if req.SidecarResourceConfig.MemoryRequest != nil {
			updateFields[entity.DotKeySidecarMemoryRequest] = req.SidecarResourceConfig.MemoryRequest.GetValue()
		}
		if req.SidecarResourceConfig.MemoryLimit != nil {
			updateFields[entity.DotKeySidecarMemoryLimit] = req.SidecarResourceConfig.MemoryLimit.GetValue()
		}
	}
	return updateFields
}

// buildHighAvailability builds high availability related configurations
func buildHighAvailability(req *meshmanager.IstioRequest, updateFields entity.M) entity.M {
	if req.HighAvailability == nil {
		return updateFields
	}

	// 更新基本字段
	if req.HighAvailability.AutoscaleEnabled != nil {
		updateFields[entity.DotKeyHAAutoscaleEnabled] = req.HighAvailability.AutoscaleEnabled.GetValue()
	}
	if req.HighAvailability.AutoscaleMin != nil {
		updateFields[entity.DotKeyHAAutoscaleMin] = req.HighAvailability.AutoscaleMin.GetValue()
	}
	if req.HighAvailability.AutoscaleMax != nil {
		updateFields[entity.DotKeyHAAutoscaleMax] = req.HighAvailability.AutoscaleMax.GetValue()
	}
	if req.HighAvailability.ReplicaCount != nil {
		updateFields[entity.DotKeyHAReplicaCount] = req.HighAvailability.ReplicaCount.GetValue()
	}
	if req.HighAvailability.TargetCPUAverageUtilizationPercent != nil {
		updateFields[entity.DotKeyHATargetCPUAverageUtilizationPercent] =
			req.HighAvailability.TargetCPUAverageUtilizationPercent.GetValue()
	}

	// 构建资源配置
	if req.HighAvailability.ResourceConfig != nil {
		if req.HighAvailability.ResourceConfig.CpuRequest != nil {
			updateFields[entity.DotKeyHAResourceCPURequest] =
				req.HighAvailability.ResourceConfig.CpuRequest.GetValue()
		}
		if req.HighAvailability.ResourceConfig.CpuLimit != nil {
			updateFields[entity.DotKeyHAResourceCPULimit] =
				req.HighAvailability.ResourceConfig.CpuLimit.GetValue()
		}
		if req.HighAvailability.ResourceConfig.MemoryRequest != nil {
			updateFields[entity.DotKeyHAResourceMemoryRequest] =
				req.HighAvailability.ResourceConfig.MemoryRequest.GetValue()
		}
		if req.HighAvailability.ResourceConfig.MemoryLimit != nil {
			updateFields[entity.DotKeyHAResourceMemoryLimit] =
				req.HighAvailability.ResourceConfig.MemoryLimit.GetValue()
		}
	}

	// 构建专用节点配置
	if req.HighAvailability.DedicatedNode != nil {
		if req.HighAvailability.DedicatedNode.Enabled != nil {
			updateFields[entity.DotKeyHADedicatedNodeEnabled] = req.HighAvailability.DedicatedNode.Enabled.GetValue()
		}
		if req.HighAvailability.DedicatedNode.NodeLabels != nil {
			updateFields[entity.DotKeyHADedicatedNodeNodeLabels] = req.HighAvailability.DedicatedNode.NodeLabels
		}
	}

	return updateFields
}

// buildObservability builds observability related configurations
func buildObservability(req *meshmanager.IstioRequest, updateFields entity.M) entity.M {
	if req.ObservabilityConfig == nil {
		return updateFields
	}

	// 构建指标配置
	if req.ObservabilityConfig.MetricsConfig != nil {
		if req.ObservabilityConfig.MetricsConfig.ControlPlaneMetricsEnabled != nil {
			updateFields[entity.DotKeyObsMetricsControlPlaneEnabled] =
				req.ObservabilityConfig.MetricsConfig.ControlPlaneMetricsEnabled.GetValue()
		}
		if req.ObservabilityConfig.MetricsConfig.DataPlaneMetricsEnabled != nil {
			updateFields[entity.DotKeyObsMetricsDataPlaneEnabled] =
				req.ObservabilityConfig.MetricsConfig.DataPlaneMetricsEnabled.GetValue()
		}
	}

	// 构建日志收集配置
	if req.ObservabilityConfig.LogCollectorConfig != nil {
		if req.ObservabilityConfig.LogCollectorConfig.Enabled != nil {
			updateFields[entity.DotKeyObsLogEnabled] =
				req.ObservabilityConfig.LogCollectorConfig.Enabled.GetValue()
		}
		if req.ObservabilityConfig.LogCollectorConfig.AccessLogEncoding != nil {
			updateFields[entity.DotKeyObsLogEncoding] =
				req.ObservabilityConfig.LogCollectorConfig.AccessLogEncoding.GetValue()
		}
		if req.ObservabilityConfig.LogCollectorConfig.AccessLogFormat != nil {
			updateFields[entity.DotKeyObsLogFormat] =
				req.ObservabilityConfig.LogCollectorConfig.AccessLogFormat.GetValue()
		}
	}

	// 构建链路追踪配置
	if req.ObservabilityConfig.TracingConfig != nil {
		if req.ObservabilityConfig.TracingConfig.Enabled != nil {
			updateFields[entity.DotKeyObsTracingEnabled] =
				req.ObservabilityConfig.TracingConfig.Enabled.GetValue()
		}
		if req.ObservabilityConfig.TracingConfig.Endpoint != nil {
			updateFields[entity.DotKeyObsTracingEndpoint] =
				req.ObservabilityConfig.TracingConfig.Endpoint.GetValue()
		}
		if req.ObservabilityConfig.TracingConfig.BkToken != nil {
			updateFields[entity.DotKeyObsTracingBkToken] =
				req.ObservabilityConfig.TracingConfig.BkToken.GetValue()
		}
		if req.ObservabilityConfig.TracingConfig.TraceSamplingPercent != nil {
			updateFields[entity.DotKeyObsTracingTraceSamplingPercent] =
				req.ObservabilityConfig.TracingConfig.TraceSamplingPercent.GetValue()
		}
	}

	return updateFields
}

// buildFeatureConfigs builds feature configurations
func buildFeatureConfigs(req *meshmanager.IstioRequest, updateFields entity.M) entity.M {
	if len(req.FeatureConfigs) == 0 {
		return updateFields
	}
	featureConfigs := make(map[string]*entity.FeatureConfig)
	for name, config := range req.FeatureConfigs {
		// Only save supported features
		if !slices.Contains(common.SupportedFeatures, name) {
			continue
		}
		featureConfigs[name] = &entity.FeatureConfig{
			Name:            config.Name,
			Description:     config.Description,
			Value:           config.Value,
			DefaultValue:    config.DefaultValue,
			AvailableValues: config.AvailableValues,
			SupportVersions: config.SupportVersions,
		}
	}
	updateFields[entity.FieldKeyFeatureConfigs] = featureConfigs
	return updateFields
}
