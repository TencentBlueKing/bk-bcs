/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"math"

	types "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/clusterops"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	k8scorev1 "k8s.io/api/core/v1"
	k8sresource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CalculateResourceAllocRate calculate resource allocate rate,
// (existed resource quota) / (cluster total resource)
func CalculateResourceAllocRate(
	quotaList []types.ResourceQuota, nodeList *k8scorev1.NodeList) (float32, error) {
	if nodeList == nil || len(nodeList.Items) == 0 {
		return math.MaxFloat32, nil
	}
	totalAllocatedCPU := k8sresource.NewMilliQuantity(0, k8sresource.BinarySI)
	for _, quota := range quotaList {
		if len(quota.ResourceQuota) == 0 {
			continue
		}
		tmpQuota := &k8scorev1.ResourceQuota{}
		if err := json.Unmarshal([]byte(quota.ResourceQuota), tmpQuota); err != nil {
			blog.Warnf("decode quota %s to k8s ResourceQuota failed, err %s", quota, err.Error())
			continue
		}
		if tmpQuota.Spec.Hard == nil {
			continue
		}
		cpuValue := tmpQuota.Spec.Hard.Cpu()
		tmpValue := cpuValue.MilliValue()
		blog.Infof("%d", tmpValue)
		totalAllocatedCPU.Add(*cpuValue)
	}
	totalNodeCPU := k8sresource.NewMilliQuantity(0, k8sresource.BinarySI)
	for _, node := range nodeList.Items {
		if node.Status.Allocatable == nil {
			continue
		}
		cpuValue := node.Status.Allocatable.Cpu()
		tmpValue := cpuValue.MilliValue()
		blog.Infof("%d", tmpValue)
		totalNodeCPU.Add(*cpuValue)
	}
	if totalNodeCPU.CmpInt64(0) == -1 {
		return math.MaxFloat32, nil
	}
	return float32(totalAllocatedCPU.MilliValue()) * 1.0 / float32(totalNodeCPU.MilliValue()) * 1.0, nil
}

// CreateQuotaToCluster create namespace and quota to cluster
func CreateQuotaToCluster(
	ctx context.Context, kubeop *clusterops.K8SOperator, clusterID string,
	ns *k8scorev1.Namespace, quota *k8scorev1.ResourceQuota) error {
	if quota == nil || ns == nil {
		return fmt.Errorf("quota cannot be empty")
	}
	kubeClient, err := kubeop.GetClusterClient(clusterID)
	if err != nil {
		return err
	}
	// create namespace
	_, err = kubeClient.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create to cluster %s, err %s", clusterID, err.Error())
	}
	// create quota
	_, err = kubeClient.CoreV1().ResourceQuotas(ns.GetName()).Create(ctx, quota, metav1.CreateOptions{})
	if err != nil {
		// rollback namespace when create quota failed
		if inErr := kubeClient.CoreV1().Namespaces().Delete(ctx, ns.GetName(), metav1.DeleteOptions{}); inErr != nil {
			blog.Warnf("rollback namespace from cluster %s failed, err %s", clusterID, inErr.Error())
		}
		return fmt.Errorf("failed to create quota to cluster %s, err %s", clusterID, err.Error())
	}
	return nil
}
