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

package pkg

import (
	"fmt"

	clsapi "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
)

var (
	listProjectClusterURL   = "/bcsapi/v4/clustermanager/v1/projects/%s/clusters?operator=%s"
	createVirtualClusterURL = "/bcsapi/v4/clustermanager/v1/vcluster"
	deleteVirtualClusterURL = "/bcsapi/v4/clustermanager/v1/vcluster/%s?operator=%s"
	updateVirtualQuotaURL   = "/bcsapi/v4/clustermanager/v1/vcluster/%s/quota"
)

// ListProjectCluster list cluster under specified project
func (c *ClusterMgrClient) ListProjectCluster(req *clsapi.ListProjectClusterReq) (
	*clsapi.ListProjectClusterResp, error) {
	if len(req.ProjectID) == 0 || len(req.Operator) == 0 {
		return nil, fmt.Errorf("lost projectID or operator")
	}
	totalURL := fmt.Sprintf(listProjectClusterURL, req.ProjectID, req.Operator)
	resp := &clsapi.ListProjectClusterResp{}
	if err := c.Get(totalURL, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// CreateVirtualCluster create a virtual cluster with default values
func (c *ClusterMgrClient) CreateVirtualCluster(req *clsapi.CreateVirtualClusterReq) (
	*clsapi.CreateVirtualClusterResp, error) {
	if len(req.ProjectID) == 0 || len(req.Creator) == 0 {
		return nil, fmt.Errorf("lost projectID or operator")
	}
	if req.ClusterType != "virtual" {
		return nil, fmt.Errorf("Bad Cluster Type")
	}

	resp := &clsapi.CreateVirtualClusterResp{}
	if err := c.Post(createVirtualClusterURL, req, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// DeleteVirtualCluster create a virtual cluster with specified clusterID
func (c *ClusterMgrClient) DeleteVirtualCluster(req *clsapi.DeleteVirtualClusterReq) (
	*clsapi.DeleteVirtualClusterResp, error) {
	if len(req.ClusterID) == 0 || len(req.Operator) == 0 {
		return nil, fmt.Errorf("lost virtual clusterID or operator")
	}
	req.OnlyDeleteInfo = false
	totalURL := fmt.Sprintf(deleteVirtualClusterURL, req.ClusterID, req.Operator)
	resp := &clsapi.DeleteVirtualClusterResp{}
	if err := c.Delete(totalURL, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// UpdateVirtualClusterQuota update virtual cluster resource if needed
func (c *ClusterMgrClient) UpdateVirtualClusterQuota(req *clsapi.UpdateVirtualClusterQuotaReq) (
	*clsapi.UpdateVirtualClusterQuotaResp, error) {
	if len(req.ClusterID) == 0 || len(req.Updater) == 0 {
		return nil, fmt.Errorf("lost virtual clusterID or operator")
	}

	totalURL := fmt.Sprintf(updateVirtualQuotaURL, req.ClusterID)
	resp := &clsapi.UpdateVirtualClusterQuotaResp{}
	if err := c.Put(totalURL, req, resp); err != nil {
		return nil, err
	}
	return resp, nil
}
