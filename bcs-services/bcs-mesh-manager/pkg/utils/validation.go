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

package utils

import (
	"context"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/cmd/mesh-manager/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/clients/k8s"
	meshmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/proto/bcs-mesh-manager"
)

// ValidateClusterVersion 检查集群版本是否支持istio版本
func ValidateClusterVersion(
	ctx context.Context,
	istioConfig *options.IstioConfig,
	clusterIDs []string,
	istioVersion string,
) error {
	for _, clusterID := range clusterIDs {
		if err := ValidateSingleClusterVersion(ctx, istioConfig, clusterID, istioVersion); err != nil {
			return err
		}
	}
	return nil
}

// ValidateSingleClusterVersion 检查单个集群版本是否支持istio版本
func ValidateSingleClusterVersion(
	ctx context.Context,
	istioConfig *options.IstioConfig,
	clusterID string,
	istioVersion string,
) error {
	version, err := k8s.GetClusterVersion(ctx, clusterID)
	if err != nil {
		return fmt.Errorf("获取集群版本失败")
	}

	istioVersionConfig := istioConfig.IstioVersions[istioVersion]
	if istioVersionConfig == nil {
		return fmt.Errorf("未找到指定的 istio 版本")
	}

	if istioVersionConfig.KubeVersion == "" {
		blog.Warnf("istio version %s kube version is empty, compatible with all versions, clusterID: %s",
			istioVersion, clusterID)
		return nil
	}
	if !IsVersionSupported(version, istioVersionConfig.KubeVersion) {
		return fmt.Errorf("集群版本与 istio 版本不兼容")
	}
	return nil
}

// ValidateIstioInstalled 检查集群中是否已经安装了istio
func ValidateIstioInstalled(ctx context.Context, clusterIDs []string) error {
	for _, clusterID := range clusterIDs {
		installed, err := k8s.CheckIstioInstalled(ctx, clusterID)
		if err != nil {
			return fmt.Errorf("检查集群 istio 安装状态失败")
		}
		if installed {
			return fmt.Errorf("集群已安装 istio")
		}
	}
	return nil
}

// ValidateObservabilityConfig 检查可观性配置是否配置正确
func ValidateObservabilityConfig(observabilityConfig *meshmanager.ObservabilityConfig) error {
	if observabilityConfig == nil {
		return nil
	}
	// 日志采集参数是否正确
	if observabilityConfig.LogCollectorConfig != nil {
		if observabilityConfig.LogCollectorConfig.Enabled.GetValue() {
			// TEXT or JSON
			if observabilityConfig.LogCollectorConfig.AccessLogEncoding.GetValue() != "TEXT" &&
				observabilityConfig.LogCollectorConfig.AccessLogEncoding.GetValue() != "JSON" {
				return fmt.Errorf("日志编码格式需为 TEXT 或 JSON")
			}
		}
	}
	// 检查otel tracing配置
	if observabilityConfig.TracingConfig != nil && observabilityConfig.TracingConfig.Enabled.GetValue() {
		// 检查endpoint
		if observabilityConfig.TracingConfig.Endpoint.GetValue() == "" {
			return fmt.Errorf("endpoint 不能为空")
		}
		// 采样率 ,0 - 100 之间
		if observabilityConfig.TracingConfig.TraceSamplingPercent.GetValue() < 0 ||
			observabilityConfig.TracingConfig.TraceSamplingPercent.GetValue() > 100 {
			return fmt.Errorf("链路追踪采样率无效，必须在 0-100 之间")
		}

		// 检查上报地址是否配置正确, 只检查service和port, path非必须（<1.21不需要）
		service, port, _, err := ParseOpenTelemetryEndpoint(observabilityConfig.TracingConfig.Endpoint.GetValue())
		if err != nil {
			return fmt.Errorf("endpoint 格式无效")
		}
		if service == "" {
			return fmt.Errorf("endpoint 无效")
		}
		if port == 0 {
			return fmt.Errorf("endpoint 无效")
		}

	}
	return nil
}

// ValidateHighAvailabilityConfig 检查高可用配置是否正确
func ValidateHighAvailabilityConfig(highAvailability *meshmanager.HighAvailability) error {
	if highAvailability == nil {
		return nil
	}

	// 副本数不能小于1
	if highAvailability.ReplicaCount != nil {
		replicaCount := highAvailability.ReplicaCount.GetValue()
		if replicaCount <= 0 {
			return fmt.Errorf("副本数必须大于 0")
		}
	}

	// 如果开启了自动扩缩容，需要检查相关配置
	if highAvailability.AutoscaleEnabled != nil && highAvailability.AutoscaleEnabled.GetValue() {
		// 检查最小副本数
		if highAvailability.AutoscaleMin == nil {
			return fmt.Errorf("最小副本数不能为空")
		}
		autoscaleMin := highAvailability.AutoscaleMin.GetValue()
		if autoscaleMin <= 0 {
			return fmt.Errorf("最小副本数必须大于 0")
		}

		// 检查最大副本数
		if highAvailability.AutoscaleMax == nil {
			return fmt.Errorf("最大副本数不能为空")
		}
		autoscaleMax := highAvailability.AutoscaleMax.GetValue()
		if autoscaleMax <= 0 {
			return fmt.Errorf("最大副本数必须大于 0")
		}

		// 检查最小副本数不能大于最大副本数
		if autoscaleMin > autoscaleMax {
			return fmt.Errorf("最小副本数不能大于最大副本数")
		}

		// 检查目标CPU使用率
		if highAvailability.TargetCPUAverageUtilizationPercent != nil {
			targetCPU := highAvailability.TargetCPUAverageUtilizationPercent.GetValue()
			if targetCPU <= 0 || targetCPU > 100 {
				return fmt.Errorf("目标 CPU 使用率必须在 1-100 之间")
			}
		}
	}

	return nil
}

// ValidateResource 验证 Istio 请求中的资源配置
// 检查 Sidecar 资源配置和 HighAvailability 资源配置的合法性
// 确保 limit >= request（当 limit 不为空且不为零时）
func ValidateResource(resourceConfig *meshmanager.ResourceConfig) error {
	// 检查sidecar resource参数
	if resourceConfig != nil {
		if err := validateResourceLimit(
			resourceConfig.CpuRequest.GetValue(),
			resourceConfig.CpuLimit.GetValue(),
		); err != nil {
			return err
		}
		if err := validateResourceLimit(
			resourceConfig.MemoryRequest.GetValue(),
			resourceConfig.MemoryLimit.GetValue(),
		); err != nil {
			return err
		}
	}
	return nil
}

// validateResourceLimit 检查limit和request是否合法，并且limit >= request
// 如果limit为0或nil，则认为不进行资源限制，不需要大于request
func validateResourceLimit(request string, limit string) error {
	var (
		requestQuantity resource.Quantity
		limitQuantity   resource.Quantity
		err             error
	)

	// request 不能为空
	if request == "" {
		return fmt.Errorf("request cannot be empty")
	}

	// 解析 request
	requestQuantity, err = resource.ParseQuantity(request)
	if err != nil {
		return fmt.Errorf("request %s is invalid, err: %s", request, err)
	}
	// request 必须大于0
	if requestQuantity.IsZero() {
		return fmt.Errorf("request %s must be greater than 0", request)
	}

	// 解析 limit
	if limit != "" {
		limitQuantity, err = resource.ParseQuantity(limit)
		if err != nil {
			return fmt.Errorf("limit %s is invalid, err: %s", limit, err)
		}
	}

	// 只有当 limit 不为空且不为0时，才进行大小比较
	if limit != "" && !limitQuantity.IsZero() && limitQuantity.Cmp(requestQuantity) < 0 {
		return fmt.Errorf("limit %s must be greater than or equal to request %s", limit, request)
	}

	return nil
}
