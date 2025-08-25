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
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/clustermanager"

	clustermanagerclient "github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/clients/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/clients/k8s"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/utils"
	meshmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/proto/bcs-mesh-manager"
)

// GetClusterInfoAction action for get cluster info
type GetClusterInfoAction struct {
	model store.MeshManagerModel
	req   *meshmanager.GetClusterInfoRequest
	resp  *meshmanager.GetClusterInfoResponse
}

// NewGetClusterInfoAction create get cluster info action
func NewGetClusterInfoAction(model store.MeshManagerModel) *GetClusterInfoAction {
	return &GetClusterInfoAction{
		model: model,
	}
}

// Handle processes the get cluster info request
func (c *GetClusterInfoAction) Handle(
	ctx context.Context,
	req *meshmanager.GetClusterInfoRequest,
	resp *meshmanager.GetClusterInfoResponse,
) error {
	c.req = req
	c.resp = resp

	if err := c.req.Validate(); err != nil {
		blog.Errorf("check cluster istio installation failed, invalid request, %s, param: %v", err.Error(), c.req)
		c.setResp(common.ParamErrorCode, err.Error(), nil)
		return nil
	}

	result, err := c.getClusterInfo(ctx)
	if err != nil {
		blog.Errorf("check cluster istio installation failed, %s, projectCode: %s", err.Error(), c.req.ProjectCode)
		c.setResp(common.InnerErrorCode, err.Error(), nil)
		return nil
	}
	c.setResp(common.SuccessCode, "check installation success", result)

	return nil
}

// setResp sets the response with code, message and data
func (c *GetClusterInfoAction) setResp(
	code uint32,
	message string,
	data *meshmanager.ClusterInfoData) {
	c.resp.Code = code
	c.resp.Message = message
	c.resp.Data = data
}

// 集群类型
const (
	// ClusterTypeFederation 联邦集群
	ClusterTypeFederation = "federation"
	// ClusterTypeVCluster vcluster集群
	ClusterTypeVCluster = "vcluster"
)

// getClusterInfo 获取集群信息
func (c *GetClusterInfoAction) getClusterInfo(ctx context.Context) (*meshmanager.ClusterInfoData, error) {
	clusters, err := clustermanagerclient.ListProjectClusters(ctx, utils.GetProjectIDFromCtx(ctx))
	if err != nil {
		blog.Errorf("list project clusters failed: %s, projectCode: %s", err.Error(), c.req.ProjectCode)
		return nil, err
	}

	var wg sync.WaitGroup
	resultsCh := make(chan *meshmanager.ClusterInfo, len(clusters))

	for _, cluster := range clusters {
		if cluster == nil {
			continue
		}

		clusterInfo := &meshmanager.ClusterInfo{
			ClusterID:   cluster.ClusterID,
			ClusterName: cluster.ClusterName,
			ClusterType: cluster.ClusterType,
			IsShared:    cluster.IsShared,
			IsInstalled: false,
			Status:      cluster.Status,
			Version:     getClusterVersion(cluster),
			Region:      cluster.Region,
		}

		// 共享集群、联邦集群、vcluster集群不检查istio安装状态
		if cluster.IsShared || cluster.ClusterType == ClusterTypeFederation || cluster.ClusterType == ClusterTypeVCluster {
			clusterInfo.IsInstalled = false
			resultsCh <- clusterInfo
			continue
		}

		wg.Add(1)
		go func(clusterID string, clusterInfo *meshmanager.ClusterInfo) {
			defer wg.Done()
			isInstalled, err := k8s.CheckIstioInstalled(ctx, clusterID)
			if err != nil {
				blog.Warnf("check cluster %s istio installation failed: %s", clusterID, err.Error())
				// 检查失败则不下发该集群信息
				return
			}
			clusterInfo.IsInstalled = isInstalled
			resultsCh <- clusterInfo
		}(cluster.ClusterID, clusterInfo)
	}

	go func() {
		wg.Wait()
		close(resultsCh)
	}()

	clusterInfos := make([]*meshmanager.ClusterInfo, 0)
	for clusterInfo := range resultsCh {
		if clusterInfo != nil {
			clusterInfos = append(clusterInfos, clusterInfo)
		}
	}

	return &meshmanager.ClusterInfoData{
		Clusters: clusterInfos,
	}, nil
}

// getClusterVersion 获取集群版本信息
func getClusterVersion(cluster *clustermanager.Cluster) string {
	if cluster.ClusterBasicSettings != nil {
		return cluster.ClusterBasicSettings.Version
	}
	return ""
}
