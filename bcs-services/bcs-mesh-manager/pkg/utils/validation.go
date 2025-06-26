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
		return fmt.Errorf("get cluster version failed, err: %s, clusterID: %s", err, clusterID)
	}

	istioVersionConfig := istioConfig.IstioVersions[istioVersion]
	if istioVersionConfig == nil {
		blog.Errorf("istio version %s not found", istioVersion)
		return fmt.Errorf("istio version %s not found", istioVersion)
	}

	if istioVersionConfig.KubeVersion == "" {
		blog.Warnf("istio version %s kube version is empty, compatible with all versions, clusterID: %s",
			istioVersion, clusterID)
		return nil
	}
	if !IsVersionSupported(version, istioVersionConfig.KubeVersion) {
		return fmt.Errorf("cluster %s version is not compatible with istio version %s", clusterID, istioVersion)
	}
	return nil
}

// ValidateIstioInstalled 检查集群中是否已经安装了istio
func ValidateIstioInstalled(ctx context.Context, clusterIDs []string) error {
	for _, clusterID := range clusterIDs {
		installed, err := k8s.CheckIstioInstalled(ctx, clusterID)
		if err != nil {
			blog.Errorf("check cluster installed istio failed, err: %s, clusterID: %s", err, clusterID)
			return err
		}
		if installed {
			return fmt.Errorf("cluster %s already installed istio", clusterID)
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
				return fmt.Errorf("log collector access log endcoding is invalid, must be TEXT or JSON")
			}
		}
	}
	// 检查otel tracing配置
	if observabilityConfig.TracingConfig != nil && observabilityConfig.TracingConfig.Enabled.GetValue() {
		// 检查endpoint
		if observabilityConfig.TracingConfig.Endpoint.GetValue() == "" {
			return fmt.Errorf("otel tracing endpoint is required")
		}
		// 采样率 ,0 - 100 之间
		if observabilityConfig.TracingConfig.TraceSamplingPercent.GetValue() < 0 ||
			observabilityConfig.TracingConfig.TraceSamplingPercent.GetValue() > 100 {
			return fmt.Errorf("otel tracing trace sampling percent is invalid")
		}

		// 检查上报地址是否配置正确, 只检查service和port, path非必须（<1.21不需要）
		service, port, _, err := ParseOpenTelemetryEndpoint(observabilityConfig.TracingConfig.Endpoint.GetValue())
		if err != nil {
			return fmt.Errorf("otel tracing endpoint is invalid, err: %s", err)
		}
		if service == "" {
			return fmt.Errorf("otel tracing endpoint is invalid")
		}
		if port == 0 {
			return fmt.Errorf("otel tracing port is invalid")
		}

	}
	return nil
}
