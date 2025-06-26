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
		updateFields["description"] = req.Description.GetValue()
	}
	if req.Name != nil {
		updateFields["name"] = req.Name.GetValue()
	}
	if req.ControlPlaneMode != nil {
		updateFields["controlPlaneMode"] = req.ControlPlaneMode.GetValue()
	}
	if req.ClusterMode != nil {
		updateFields["clusterMode"] = req.ClusterMode.GetValue()
	}
	if req.PrimaryClusters != nil {
		updateFields["primaryClusters"] = req.PrimaryClusters
	}
	if req.RemoteClusters != nil {
		updateFields["remoteClusters"] = req.RemoteClusters
	}
	if req.DifferentNetwork != nil {
		updateFields["differentNetwork"] = req.DifferentNetwork.GetValue()
	}
	return updateFields
}

// buildResourceConfigs builds resource related configurations
func buildResourceConfigs(req *meshmanager.IstioRequest, updateFields entity.M) entity.M {
	// Update Sidecar resource config
	if req.SidecarResourceConfig != nil {
		sidecarResourceConfig := &entity.ResourceConfig{}
		if req.SidecarResourceConfig.CpuRequest != nil {
			sidecarResourceConfig.CpuRequest = req.SidecarResourceConfig.CpuRequest.GetValue()
		}
		if req.SidecarResourceConfig.CpuLimit != nil {
			sidecarResourceConfig.CpuLimit = req.SidecarResourceConfig.CpuLimit.GetValue()
		}
		if req.SidecarResourceConfig.MemoryRequest != nil {
			sidecarResourceConfig.MemoryRequest = req.SidecarResourceConfig.MemoryRequest.GetValue()
		}
		if req.SidecarResourceConfig.MemoryLimit != nil {
			sidecarResourceConfig.MemoryLimit = req.SidecarResourceConfig.MemoryLimit.GetValue()
		}
		updateFields["sidecarResourceConfig"] = sidecarResourceConfig
	}
	return updateFields
}

// buildHighAvailability builds high availability related configurations
func buildHighAvailability(req *meshmanager.IstioRequest, updateFields entity.M) entity.M {
	if req.HighAvailability != nil {
		highAvailability := &entity.HighAvailability{}

		if req.HighAvailability.AutoscaleEnabled != nil {
			highAvailability.AutoscaleEnabled = req.HighAvailability.AutoscaleEnabled.GetValue()
		}
		if req.HighAvailability.AutoscaleMin != nil {
			highAvailability.AutoscaleMin = req.HighAvailability.AutoscaleMin.GetValue()
		}
		if req.HighAvailability.AutoscaleMax != nil {
			highAvailability.AutoscaleMax = req.HighAvailability.AutoscaleMax.GetValue()
		}
		if req.HighAvailability.ReplicaCount != nil {
			highAvailability.ReplicaCount = req.HighAvailability.ReplicaCount.GetValue()
		}
		if req.HighAvailability.TargetCPUAverageUtilizationPercent != nil {
			highAvailability.TargetCPUAverageUtilizationPercent =
				req.HighAvailability.TargetCPUAverageUtilizationPercent.GetValue()
		}

		// 构建资源配置
		if req.HighAvailability.ResourceConfig != nil {
			highAvailability.ResourceConfig = &entity.ResourceConfig{}
			if req.HighAvailability.ResourceConfig.CpuRequest != nil {
				highAvailability.ResourceConfig.CpuRequest =
					req.HighAvailability.ResourceConfig.CpuRequest.GetValue()
			}
			if req.HighAvailability.ResourceConfig.CpuLimit != nil {
				highAvailability.ResourceConfig.CpuLimit =
					req.HighAvailability.ResourceConfig.CpuLimit.GetValue()
			}
			if req.HighAvailability.ResourceConfig.MemoryRequest != nil {
				highAvailability.ResourceConfig.MemoryRequest =
					req.HighAvailability.ResourceConfig.MemoryRequest.GetValue()
			}
			if req.HighAvailability.ResourceConfig.MemoryLimit != nil {
				highAvailability.ResourceConfig.MemoryLimit =
					req.HighAvailability.ResourceConfig.MemoryLimit.GetValue()
			}
		}

		// 构建专用节点配置
		if req.HighAvailability.DedicatedNode != nil {
			highAvailability.DedicatedNode = &entity.DedicatedNode{}
			if req.HighAvailability.DedicatedNode.Enabled != nil {
				highAvailability.DedicatedNode.Enabled =
					req.HighAvailability.DedicatedNode.Enabled.GetValue()
			}
			if req.HighAvailability.DedicatedNode.NodeLabels != nil {
				highAvailability.DedicatedNode.NodeLabels =
					req.HighAvailability.DedicatedNode.NodeLabels
			}
		}

		updateFields["highAvailability"] = highAvailability
	}
	return updateFields
}

// buildObservability builds observability related configurations
func buildObservability(req *meshmanager.IstioRequest, updateFields entity.M) entity.M {
	if req.ObservabilityConfig != nil {
		observabilityConfig := &entity.ObservabilityConfig{}

		// 构建指标配置
		if req.ObservabilityConfig.MetricsConfig != nil {
			observabilityConfig.MetricsConfig = &entity.MetricsConfig{}
			if req.ObservabilityConfig.MetricsConfig.ControlPlaneMetricsEnabled != nil {
				observabilityConfig.MetricsConfig.ControlPlaneMetricsEnabled =
					req.ObservabilityConfig.MetricsConfig.ControlPlaneMetricsEnabled.GetValue()
			}
			if req.ObservabilityConfig.MetricsConfig.DataPlaneMetricsEnabled != nil {
				observabilityConfig.MetricsConfig.DataPlaneMetricsEnabled =
					req.ObservabilityConfig.MetricsConfig.DataPlaneMetricsEnabled.GetValue()
			}
		}

		// 构建日志收集配置
		if req.ObservabilityConfig.LogCollectorConfig != nil {
			observabilityConfig.LogCollectorConfig = &entity.LogCollectorConfig{}
			if req.ObservabilityConfig.LogCollectorConfig.Enabled != nil {
				observabilityConfig.LogCollectorConfig.Enabled =
					req.ObservabilityConfig.LogCollectorConfig.Enabled.GetValue()
			}
			if req.ObservabilityConfig.LogCollectorConfig.AccessLogEncoding != nil {
				observabilityConfig.LogCollectorConfig.AccessLogEncoding =
					req.ObservabilityConfig.LogCollectorConfig.AccessLogEncoding.GetValue()
			}
			if req.ObservabilityConfig.LogCollectorConfig.AccessLogFormat != nil {
				observabilityConfig.LogCollectorConfig.AccessLogFormat =
					req.ObservabilityConfig.LogCollectorConfig.AccessLogFormat.GetValue()
			}
		}

		// 构建链路追踪配置
		if req.ObservabilityConfig.TracingConfig != nil {
			observabilityConfig.TracingConfig = &entity.TracingConfig{}
			if req.ObservabilityConfig.TracingConfig.Enabled != nil {
				observabilityConfig.TracingConfig.Enabled =
					req.ObservabilityConfig.TracingConfig.Enabled.GetValue()
			}
			if req.ObservabilityConfig.TracingConfig.Endpoint != nil {
				observabilityConfig.TracingConfig.Endpoint =
					req.ObservabilityConfig.TracingConfig.Endpoint.GetValue()
			}
			if req.ObservabilityConfig.TracingConfig.BkToken != nil {
				observabilityConfig.TracingConfig.BkToken =
					req.ObservabilityConfig.TracingConfig.BkToken.GetValue()
			}
			if req.ObservabilityConfig.TracingConfig.TraceSamplingPercent != nil {
				observabilityConfig.TracingConfig.TraceSamplingPercent =
					req.ObservabilityConfig.TracingConfig.TraceSamplingPercent.GetValue()
			}
		}

		updateFields["observabilityConfig"] = observabilityConfig
	}
	return updateFields
}

// buildFeatureConfigs builds feature configurations
func buildFeatureConfigs(req *meshmanager.IstioRequest, updateFields entity.M) entity.M {
	if len(req.FeatureConfigs) == 0 {
		return updateFields
	}
	blog.Infof("更新特征配置: %s", req.MeshID.GetValue())
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
	updateFields["featureConfigs"] = featureConfigs
	return updateFields
}
