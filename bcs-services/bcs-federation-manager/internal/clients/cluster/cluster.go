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

// Package cluster xxx
package cluster

import (
	"context"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/clustermanager"
)

// GetCluster get cluster info
func (h *clusterClient) GetCluster(ctx context.Context, clusterId string) (*clustermanager.Cluster, error) {
	resp, err := h.clusterSvc.GetCluster(h.getMetadataCtx(ctx), &clustermanager.GetClusterReq{
		ClusterID: clusterId,
	})
	if err != nil {
		return nil, err
	}

	if !resp.Result {
		return nil, fmt.Errorf("get cluster %s error: %s", clusterId, resp.Message)
	}

	return resp.Data, nil
}

// ListProjectCluster list project cluster
func (h *clusterClient) ListProjectCluster(ctx context.Context, projectId string) ([]*clustermanager.Cluster, error) {
	resp, err := h.clusterSvc.ListProjectCluster(h.getMetadataCtx(ctx), &clustermanager.ListProjectClusterReq{
		ProjectID: projectId,
	})
	if err != nil {
		return nil, err
	}
	if !resp.Result {
		return nil, fmt.Errorf("list cluster for project %s error: %s", projectId, resp.Message)
	}

	return resp.Data, nil
}

// GetSubnetId get cluster subnetId
func (h *clusterClient) GetSubnetId(ctx context.Context, clusterId string) (string, error) {
	// get vpc and region for host cluster
	cluster, err := h.GetCluster(ctx, clusterId)
	if err != nil {
		return "", fmt.Errorf("get cluster %s failed, err: %v", clusterId, err)
	}
	vpcId, region := cluster.GetVpcID(), cluster.GetRegion()
	if vpcId == "" || region == "" {
		return "", fmt.Errorf("cluster %s vpcId[%s] or region[%s] is empty", clusterId, vpcId, region)
	}

	//  get subnetId from clustermanager for cluster
	resp, err := h.clusterSvc.ListCloudSubnets(h.getMetadataCtx(ctx), &clustermanager.ListCloudSubnetsRequest{
		CloudID: DefaultProviderTencent,
		Region:  region,
		VpcID:   vpcId,
	})
	if err != nil {
		return "", fmt.Errorf("list cloud subnet for cluster %s failed, err: %v", clusterId, err)
	}
	if !resp.Result {
		return "", fmt.Errorf("list cloud subnet for cluster %s failed, err: %s", clusterId, resp.Message)
	}

	// find target subnetId which can not contains "BCS-K8S-" and maxAvailableIpCount is max
	subnetId := ""
	maxAvailableIpCount := uint64(0)
	for _, subnet := range resp.Data {
		// can not user subnet which contains "BCS-K8S-""
		if h.isMatchPattern(ClusterSubnetsRegexPattern, subnet.GetSubnetName()) {
			continue
		}

		// find max available ip count subnet
		if subnet.GetAvailableIPAddressCount() > 0 && subnet.GetAvailableIPAddressCount() > maxAvailableIpCount {
			subnetId = subnet.GetSubnetID()
			maxAvailableIpCount = subnet.GetAvailableIPAddressCount()
		}
	}
	if subnetId == "" || maxAvailableIpCount == 0 {
		return "", fmt.Errorf("there is not enough subnet for cluster: %s", clusterId)
	}

	return subnetId, nil
}

// CreateFederationCluster register federation cluster
func (h *clusterClient) CreateFederationCluster(ctx context.Context, req *FederationClusterCreateReq) (string, error) {
	// AddFederationClusterLabel
	req.Labels[FederationClusterTypeLabelKeyFedCluster] = FederationClusterTypeLabelValueTrue
	cmReq := &clustermanager.CreateClusterReq{
		BusinessID:  req.BusinessId,
		ProjectID:   req.ProjectId,
		ClusterName: req.ClusterName,
		Creator:     req.Creator,
		Environment: req.Environment,
		Labels:      req.Labels,
		Description: req.Description,

		Provider:             DefaultProviderBlueking,
		Region:               DefaultRegionDefault,
		EngineType:           DefaultEnginTypeK8s,
		ClusterType:          DefaultClusterTypeFederation,
		ManageType:           DefaultManageTypeIndependent,
		NetworkType:          DefaultNetworkTypeOverlay,
		IsExclusive:          true,
		NetworkSettings:      &clustermanager.NetworkSetting{},
		ClusterBasicSettings: &clustermanager.ClusterBasicSetting{},
		// create record only, when OnlyCreateInfo=true, Status is valid
		Status:         ClusterStatusInitialization,
		OnlyCreateInfo: true,
	}

	resp, err := h.clusterSvc.CreateCluster(h.getMetadataCtx(ctx), cmReq)
	if err != nil {
		return "", fmt.Errorf("request import cluster interface failed, err: %v", err)
	}

	if resp.Code != 0 {
		return "", fmt.Errorf("import cluster failed, err: %s", resp.Message)
	}

	return resp.Data.GetClusterID(), nil
}

// UpdateFederationClusterCredentials update cluster credentials
func (h *clusterClient) UpdateFederationClusterCredentials(ctx context.Context, clusterId, kubeconfig string) error {
	// update cluster kubeconfig
	resp, err := h.clusterSvc.UpdateClusterKubeConfig(h.getMetadataCtx(ctx), &clustermanager.UpdateClusterKubeConfigReq{
		ClusterID:  clusterId,
		KubeConfig: kubeconfig,
	})
	if err != nil {
		return fmt.Errorf("update cluster %s kubeconfig failed, err: %v", clusterId, err)
	}
	if !resp.Result {
		return fmt.Errorf("update cluster %s kubeconfig failed: %s", clusterId, resp.Message)
	}

	return nil
}

// UpdateFederationClusterStatus update cluster status
func (h *clusterClient) UpdateFederationClusterStatus(ctx context.Context, clusterId, status string) error {
	if _, ok := ClusterStatusList[status]; !ok {
		return fmt.Errorf("cluster status can not be %s", status)
	}

	resp, err := h.clusterSvc.UpdateCluster(h.getMetadataCtx(ctx), &clustermanager.UpdateClusterReq{
		ClusterID: clusterId,
		Status:    status,
	})
	if err != nil {
		return fmt.Errorf("update cluster %s status error, err: %v", clusterId, err)
	}

	if !resp.Result {
		return fmt.Errorf("update cluster %s status failed: %s", clusterId, resp.Message)
	}
	return nil
}

// DeleteFederationCluster delete cluster from clustermanager
func (h *clusterClient) DeleteFederationCluster(ctx context.Context, clusterId string, operator string) error {
	resp, err := h.clusterSvc.DeleteCluster(h.getMetadataCtx(ctx), &clustermanager.DeleteClusterReq{
		ClusterID:      clusterId,
		OnlyDeleteInfo: true,
		Operator:       operator,
	})
	if err != nil {
		return fmt.Errorf("delete cluster %s error, err: %v", clusterId, err)
	}
	if !resp.Result {
		return fmt.Errorf("delete cluster %s failed: %s", clusterId, resp.Message)
	}

	return nil
}

// UpdateHostClusterLabel add label to cluster
func (h *clusterClient) UpdateHostClusterLabel(ctx context.Context, clusterId string) error {
	return h.AddClusterLabels(ctx, clusterId,
		map[string]string{FederationClusterTypeLabelKeyHostCluster: FederationClusterTypeLabelValueTrue},
	)
}

// DeleteClusterLabels delete label from cluster
func (h *clusterClient) DeleteHostClusterLabel(ctx context.Context, clusterId string) error {
	return h.DeleteClusterLabels(ctx, clusterId,
		[]string{FederationClusterTypeLabelKeyHostCluster},
	)
}

// UpdateSubClusterLabel add label to cluster
func (h *clusterClient) UpdateSubClusterLabel(ctx context.Context, clusterId string) error {
	return h.AddClusterLabels(ctx, clusterId,
		map[string]string{FederationClusterTypeLabelKeySubCluster: FederationClusterTypeLabelValueTrue},
	)
}

// DeleteSubClusterLabel delete label from cluster
func (h *clusterClient) DeleteSubClusterLabel(ctx context.Context, clusterId string) error {
	return h.DeleteClusterLabels(ctx, clusterId,
		[]string{FederationClusterTypeLabelKeySubCluster},
	)
}

// AddClusterLabels add label to cluster
func (h *clusterClient) AddClusterLabels(ctx context.Context, clusterId string, labels map[string]string) error {
	// get old labels
	clusterLabels, err := h.GetClusterLabels(ctx, clusterId)
	if err != nil {
		return fmt.Errorf("get cluster %s labels error, err: %v", clusterId, err)
	}

	// merge old labels and new labels
	updateLabels := make(map[string]string)
	if clusterLabels != nil {
		updateLabels = clusterLabels
	}

	// add labels
	for k, v := range labels {
		// cover or add old label key
		updateLabels[k] = v
	}

	return h.UpdateClusterLabels(ctx, clusterId, updateLabels)
}

// DeleteClusterLabels delete label from cluster manager
func (h *clusterClient) DeleteClusterLabels(ctx context.Context, clusterId string, labelKeys []string) error {
	// get old labels
	clusterLabels, err := h.GetClusterLabels(ctx, clusterId)
	if err != nil {
		return fmt.Errorf("get cluster %s labels error, err: %v", clusterId, err)
	}

	// merge old labels and new labels
	if clusterLabels == nil {
		// nothing to delete
		return nil
	}
	updateLabels := clusterLabels

	for _, labelKey := range labelKeys {
		// if not exist
		if _, ok := updateLabels[labelKey]; !ok {
			continue
		}

		// delete old label key
		delete(updateLabels, labelKey)
	}

	return h.UpdateClusterLabels(ctx, clusterId, updateLabels)
}

// GetClusterLabels get cluster labels
func (h *clusterClient) GetClusterLabels(ctx context.Context, clusterId string) (map[string]string, error) {
	getResp, err := h.clusterSvc.GetCluster(h.getMetadataCtx(ctx), &clustermanager.GetClusterReq{
		ClusterID: clusterId,
	})
	if err != nil {
		return nil, fmt.Errorf("get cluster %s error, err: %v", clusterId, err)
	}

	if !getResp.Result {
		return nil, fmt.Errorf("get cluster %s failed: %s", clusterId, getResp.Message)
	}

	cluster := getResp.Data
	return cluster.GetLabels(), nil
}

// UpdateClusterLabels update cluster labels
func (h *clusterClient) UpdateClusterLabels(ctx context.Context, clusterId string, updateLabels map[string]string) error {
	cluster, err := h.GetCluster(ctx, clusterId)
	if err != nil {
		return fmt.Errorf("get cluster %s error when update cluster labels, err: %v", clusterId, err)
	}

	updateResp, err := h.clusterSvc.UpdateCluster(h.getMetadataCtx(ctx), &clustermanager.UpdateClusterReq{
		ClusterID: clusterId,
		Labels2:   &clustermanager.MapStruct{Values: updateLabels},
		Status:    cluster.GetStatus(),
	})
	if err != nil {
		return fmt.Errorf("update cluster %s labels error, err: %v", clusterId, err)
	}

	if !updateResp.Result {
		return fmt.Errorf("update cluster %s labels failed: %s", clusterId, updateResp.Message)
	}
	return nil
}
