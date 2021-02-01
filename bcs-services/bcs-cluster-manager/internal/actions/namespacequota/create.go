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

package namespacequota

import (
	"context"
	"encoding/json"
	"fmt"
	"math"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/clusterops"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	storeopt "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/types"

	k8scorev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	k8sresource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CreateAction action for creating quota
type CreateAction struct {
	ctx              context.Context
	model            store.ClusterManagerModel
	k8sop            *clusterops.K8SOperator
	ns               *types.Namespace
	req              *cmproto.CreateNamespaceQuotaReq
	resp             *cmproto.CreateNamespaceQuotaResp
	quota            *k8scorev1.ResourceQuota
	allocatedCluster string
}

// NewCreateAction create action for creating quota
func NewCreateAction(model store.ClusterManagerModel, k8sop *clusterops.K8SOperator) *CreateAction {
	return &CreateAction{
		model: model,
		k8sop: k8sop,
	}
}

func (ca *CreateAction) validate() error {
	if err := ca.req.Validate(); err != nil {
		return err
	}
	quota := &k8scorev1.ResourceQuota{}
	if err := json.Unmarshal([]byte(ca.req.ResourceQuota), quota); err != nil {
		return err
	}
	if quota.Name != ca.req.Namespace || quota.Namespace != ca.req.Namespace {
		return fmt.Errorf("resource quota name and namespace should be the name of namespace %s", ca.req.Namespace)
	}
	ca.quota = quota
	return nil
}

func (ca *CreateAction) getNamespaceFromStore() error {
	ns, err := ca.model.GetNamespace(ca.ctx, ca.req.Namespace, ca.req.FederationClusterID)
	if err != nil {
		return err
	}
	ca.ns = ns
	return nil
}

func (ca *CreateAction) listNodesFromCluster(cluster string) (*k8scorev1.NodeList, error) {
	kubeClient, err := ca.k8sop.GetClusterClient(cluster)
	if err != nil {
		return nil, err
	}
	nodeList, err := kubeClient.CoreV1().Nodes().List(ca.ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return nodeList, err
}

func (ca *CreateAction) listQuotasByCluster(cluster string) ([]types.NamespaceQuota, error) {
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		"federationClusterID": ca.req.FederationClusterID,
		"clusterID":           cluster,
	})
	quotaList, err := ca.model.ListQuota(ca.ctx, cond, &storeopt.ListOption{})
	if err != nil {
		return nil, err
	}
	return quotaList, nil
}

// calculate resource allocate rate, (existed resource quota) / (cluster total resource)
func (ca *CreateAction) calculateResourceAllocRate(
	quotaList []types.NamespaceQuota, nodeList *k8scorev1.NodeList) (float32, error) {
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

// allocate one cluster for the request resource quota
func (ca *CreateAction) allocateOneCluster() error {
	// 计算已经分配的quota与集群总资源的比值，最后算出比值最小的集群
	condM := operator.M{
		"region":              ca.req.Region,
		"federationClusterID": ca.req.FederationClusterID,
		"engineType":          common.ClusterEngineTypeK8s,
	}
	cond := operator.NewLeafCondition(operator.Eq, condM)
	clusterList, err := ca.model.ListCluster(ca.ctx, cond, &storeopt.ListOption{})
	if err != nil {
		return err
	}
	minResRate := float32(math.MaxFloat32)
	targetCluster := ""
	for _, cluster := range clusterList {
		nodes, err := ca.listNodesFromCluster(cluster.ClusterID)
		if err != nil {
			return err
		}
		quotas, err := ca.listQuotasByCluster(cluster.ClusterID)
		if err != nil {
			return err
		}
		tmpRate, err := ca.calculateResourceAllocRate(quotas, nodes)
		if err != nil {
			return err
		}
		if tmpRate <= minResRate {
			targetCluster = cluster.ClusterID
		}
	}
	if len(targetCluster) == 0 {
		return fmt.Errorf("no found target cluster")
	}
	ca.allocatedCluster = targetCluster
	return nil
}

// ensure namespace and namespace quota to cluster
func (ca *CreateAction) createQuotaToCluster() error {
	if ca.quota == nil {
		return fmt.Errorf("request quota is empty")
	}
	kubeClient, err := ca.k8sop.GetClusterClient(ca.allocatedCluster)
	if err != nil {
		return err
	}
	// ensure namespace
	_, err = kubeClient.CoreV1().Namespaces().Get(ca.ctx, ca.ns.Name, metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			_, err = kubeClient.CoreV1().Namespaces().Create(ca.ctx, &k8scorev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: ca.ns.Name,
				},
			}, metav1.CreateOptions{})
			if err != nil {
				return err
			}
		}
	}

	_, err = kubeClient.CoreV1().ResourceQuotas(ca.req.Namespace).Create(ca.ctx, ca.quota, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	return nil
}

// create namespace resoucequota to store
func (ca *CreateAction) createQuotaToStore() error {
	ca.quota.ClusterName = ca.allocatedCluster
	newQuota := &types.NamespaceQuota{
		Namespace:           ca.req.Namespace,
		FederationClusterID: ca.req.FederationClusterID,
		ClusterID:           ca.allocatedCluster,
		ResourceQuota:       ca.req.ResourceQuota,
	}
	if err := ca.model.CreateQuota(ca.ctx, newQuota); err != nil {
		return err
	}
	return nil
}

func (ca *CreateAction) setResp(code uint64, msg string) {
	ca.resp.Seq = ca.req.Seq
	ca.resp.ErrCode = code
	ca.resp.ErrMsg = msg
	ca.resp.ClusterID = ca.allocatedCluster
}

// Handle handle namespace quota request
func (ca *CreateAction) Handle(
	ctx context.Context, req *cmproto.CreateNamespaceQuotaReq, resp *cmproto.CreateNamespaceQuotaResp) {
	if req == nil || resp == nil {
		blog.Errorf("create namespace quota failed, req or resp is empty")
		return
	}
	ca.ctx = ctx
	ca.req = req
	ca.resp = resp

	if err := ca.validate(); err != nil {
		ca.setResp(types.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	if err := ca.getNamespaceFromStore(); err != nil {
		ca.setResp(types.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	if len(ca.req.ClusterID) == 0 {
		if err := ca.allocateOneCluster(); err != nil {
			ca.setResp(types.BcsErrClusterManagerAllocateClusterInCreateQuota, err.Error())
			return
		}
	} else {
		// use requested cluster
		ca.allocatedCluster = ca.req.ClusterID
	}

	if err := ca.createQuotaToCluster(); err != nil {
		ca.setResp(types.BcsErrClusterManagerK8SOpsFailed, err.Error())
		return
	}
	if err := ca.createQuotaToStore(); err != nil {
		ca.setResp(types.BcsErrClusterManagerDBOperation, err.Error())
		return
	}
	ca.setResp(types.BcsErrClusterManagerSuccess, types.BcsErrClusterManagerSuccessStr)
	return
}
