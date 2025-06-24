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
