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

package handler

import (
	"context"
	"errors"

	meshmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/proto/bcs-mesh-manager"
)

// ListIstioVersion implements meshmanager.MeshManagerHandler
func (m *MeshManager) ListIstioVersion(
	ctx context.Context,
	req *meshmanager.ListIstioVersionRequest,
	resp *meshmanager.ListIstioVersionResponse,
) error {
	// 获取版本配置并输出
	istioConfig := m.opt.IstioConfig
	if istioConfig == nil {
		return errors.New("istio config is nil")
	}
	for _, version := range istioConfig.IstioVersion {
		resp.Data = append(resp.Data, &meshmanager.IstioVersion{
			Name:         version.Name,
			ChartVersion: version.ChartVersion,
			KubeVersion:  version.KubeVersion,
		})
	}

	return nil
}
