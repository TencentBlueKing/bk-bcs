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
	"strings"

	v1beta1 "github.com/clusternet/clusternet/pkg/apis/clusters/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// ClusterRegistrationRequestClusterLabelKey cluster registration request cluster label key
	ClusterRegistrationRequestClusterLabelKey = "clusters.clusternet.io/cluster-name"
)

// GetManagedCluster get managed cluster by clusterID
func (h *clusterClient) GetManagedCluster(hostClusterId, subClusterID string) (*v1beta1.ManagedCluster, error) {
	client, err := h.getClusternetClientByClusterId(hostClusterId)
	if err != nil {
		return nil, err
	}

	mclsList, err := client.ClustersV1beta1().ManagedClusters("").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	mclsName := strings.ToLower(subClusterID)
	for _, mcls := range mclsList.Items {
		if mcls.Name == mclsName {
			return &mcls, nil
		}
	}

	// not found, return nil error and nil obj means not found
	return nil, nil
}

// DeleteManagedCluster delete managed cluster by clusterID
func (h *clusterClient) DeleteManagedCluster(hostClusterId, subClusterID string) error {
	client, err := h.getClusternetClientByClusterId(hostClusterId)
	if err != nil {
		return err
	}

	// get before delete
	mcls, err := h.GetManagedCluster(hostClusterId, subClusterID)
	if err != nil {
		return err
	}

	// delete when found
	if mcls != nil {
		return client.ClustersV1beta1().ManagedClusters(mcls.Namespace).Delete(context.Background(), mcls.Name, metav1.DeleteOptions{})
	}
	// already deleted
	return nil
}

// GetClusterRegistrationRequest get cluster registration request by clusterID
func (h *clusterClient) GetClusterRegistrationRequest(hostClusterId, subClusterID string) (*v1beta1.ClusterRegistrationRequest, error) {
	client, err := h.getClusternetClientByClusterId(hostClusterId)
	if err != nil {
		return nil, err
	}

	crReqName := strings.ToLower(subClusterID)
	crReqList, err := client.ClustersV1beta1().ClusterRegistrationRequests().List(context.Background(), metav1.ListOptions{
		LabelSelector: ClusterRegistrationRequestClusterLabelKey + "=" + crReqName,
	})
	if err != nil {
		return nil, err
	}

	if len(crReqList.Items) != 0 {
		return &crReqList.Items[0], nil
	}

	// not found, return nil error and nil obj means not found
	return nil, nil
}

// DeleteClusterRegistrationRequest delete cluster registration request by clusterID
func (h *clusterClient) DeleteClusterRegistrationRequest(hostClusterId, subClusterID string) error {
	client, err := h.getClusternetClientByClusterId(hostClusterId)
	if err != nil {
		return err
	}

	// get before delete
	crReq, err := h.GetClusterRegistrationRequest(hostClusterId, subClusterID)
	if err != nil {
		return err
	}

	// delete when found
	if crReq != nil {
		return client.ClustersV1beta1().ClusterRegistrationRequests().Delete(context.Background(), crReq.Name, metav1.DeleteOptions{})
	}
	// already deleted
	return nil
}
