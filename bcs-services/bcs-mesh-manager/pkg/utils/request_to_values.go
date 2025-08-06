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
	"fmt"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/common"
	meshmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/proto/bcs-mesh-manager"
)

// ConvertRequestToValues 从 IstioRequest 构建 IstiodInstallValues 部署配置
func ConvertRequestToValues(istioVersion string, req *meshmanager.IstioRequest) (*common.IstiodInstallValues, error) {
	if req == nil {
		return nil, fmt.Errorf("request is nil")
	}
	installValues := &common.IstiodInstallValues{}

	// 构建基础配置
	buildBasicConfig(req, installValues)

	// 构建Sidecar资源配置
	if err := GenIstiodValuesBySidecarResource(req.SidecarResourceConfig, installValues); err != nil {
		blog.Warnf("failed to build sidecar resource config: %v", err)
		return nil, err
	}

	// 构建高可用配置
	if err := GenIstiodValuesByHighAvailability(req.HighAvailability, installValues); err != nil {
		blog.Warnf("failed to build high availability config: %v", err)
		return nil, err
	}

	// 构建功能特性配置
	GenIstiodValuesByFeature(req.FeatureConfigs, installValues)

	// 构建可观测性配置
	if err := GenIstiodValuesByObservability(istioVersion, req.ObservabilityConfig, installValues); err != nil {
		blog.Warnf("failed to build observability config: %v", err)
		return nil, err
	}

	return installValues, nil
}

// buildBasicConfig 构建基础配置
func buildBasicConfig(
	req *meshmanager.IstioRequest,
	installValues *common.IstiodInstallValues,
) {
	// 构建MultiCluster配置
	if len(req.PrimaryClusters) > 0 {
		if installValues.Global == nil {
			installValues.Global = &common.IstiodGlobalConfig{}
		}
		if installValues.Global.MultiCluster == nil {
			installValues.Global.MultiCluster = &common.IstiodMultiClusterConfig{}
		}
		// todo: 兼容多集群逻辑
		clusterName := strings.ToLower(req.PrimaryClusters[0])
		installValues.Global.MultiCluster.ClusterName = &clusterName
	}
}
