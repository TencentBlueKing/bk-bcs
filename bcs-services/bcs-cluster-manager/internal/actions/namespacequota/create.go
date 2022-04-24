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
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/clusterops"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	storeopt "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"

	k8scorev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CreateAction action for creating quota
type CreateAction struct {
	ctx              context.Context
	model            store.ClusterManagerModel
	k8sop            *clusterops.K8SOperator
	ns               *cmproto.Namespace
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
		return fmt.Errorf("decode resourcequota failed, err %s", err)
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

func (ca *CreateAction) isQuotaExisted(clusterID string) (bool, error) {
	quota, err := ca.model.GetQuota(ca.ctx, ca.req.Namespace, ca.req.FederationClusterID, clusterID)
	if err != nil {
		if errors.Is(err, drivers.ErrTableRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	if quota == nil {
		return false, fmt.Errorf("quota is nil")
	}
	return true, nil
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

func (ca *CreateAction) listQuotasByCluster(cluster string) ([]cmproto.ResourceQuota, error) {
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		"federationclusterid": ca.req.FederationClusterID,
		"clusterid":           cluster,
	})
	quotaList, err := ca.model.ListQuota(ca.ctx, cond, &storeopt.ListOption{})
	if err != nil {
		return nil, err
	}
	return quotaList, nil
}

// allocate one cluster for the request resource quota
func (ca *CreateAction) allocateOneCluster() error {
	// 计算已经分配的quota与集群总资源的比值，最后算出比值最小的集群
	condM := operator.M{
		"region":              ca.req.Region,
		"federationclusterid": ca.req.FederationClusterID,
		"enginetype":          common.ClusterEngineTypeK8s,
		"clustertype":         common.ClusterTypeSingle,
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
			blog.Warnf("failed to list nodes from cluster %s, continue to check next cluster, err %s",
				cluster.ClusterID, err.Error())
			continue
		}
		quotas, err := ca.listQuotasByCluster(cluster.ClusterID)
		if err != nil {
			blog.Warnf("failed to list  quotas by cluster %s, continue to check next cluster, err %s",
				cluster.ClusterID, err.Error())
			continue
		}
		tmpRate, err := utils.CalculateResourceAllocRate(quotas, nodes)
		if err != nil {
			blog.Warnf("failed to calculate rate of cluster %s, continue to check next cluster, err %s",
				cluster.ClusterID, err.Error())
			continue
		}
		if tmpRate <= minResRate {
			targetCluster = cluster.ClusterID
		}
	}
	if len(targetCluster) == 0 {
		return fmt.Errorf("can not find a suitable cluster")
	}
	ca.allocatedCluster = targetCluster
	return nil
}

// create namespace resoucequota to store
func (ca *CreateAction) createQuotaToStore() error {
	createTime := time.Now().Format(time.RFC3339)
	ca.quota.ClusterName = ca.allocatedCluster
	newQuota := &cmproto.ResourceQuota{
		Namespace:           ca.req.Namespace,
		FederationClusterID: ca.req.FederationClusterID,
		ClusterID:           ca.allocatedCluster,
		Region:              ca.req.Region,
		ResourceQuota:       ca.req.ResourceQuota,
		CreateTime:          createTime,
		UpdateTime:          createTime,
	}
	if err := ca.model.CreateQuota(ca.ctx, newQuota); err != nil {
		return err
	}
	return nil
}

func (ca *CreateAction) setResp(code uint32, msg string) {
	ca.resp.Code = code
	ca.resp.Message = msg
	ca.resp.Result = (code == common.BcsErrClusterManagerSuccess)
	ca.resp.Data = &cmproto.CreateNamespaceQuotaResp_CreateNamespaceQuotaRespData{
		ClusterID: ca.allocatedCluster,
	}
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
		ca.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	if err := ca.getNamespaceFromStore(); err != nil {
		ca.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	if len(ca.req.ClusterID) == 0 {
		if err := ca.allocateOneCluster(); err != nil {
			ca.setResp(common.BcsErrClusterManagerAllocateClusterInCreateQuota, err.Error())
			return
		}
	} else {
		// use requested cluster
		ca.allocatedCluster = ca.req.ClusterID
	}

	isExisted, err := ca.isQuotaExisted(ca.allocatedCluster)
	if err != nil {
		ca.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}
	if isExisted {
		ca.setResp(common.BcsErrClusterManagerResourceDuplicated,
			fmt.Sprintf("quota %s/%s/%s is duplicated",
				ca.req.Namespace, ca.req.FederationClusterID, ca.allocatedCluster))
		return
	}

	if err := utils.CreateQuotaToCluster(ca.ctx, ca.k8sop, ca.allocatedCluster, &k8scorev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   ca.ns.Name,
			Labels: ca.ns.Labels,
		},
	}, ca.quota); err != nil {
		ca.setResp(common.BcsErrClusterManagerK8SOpsFailed, err.Error())
		return
	}
	if err := ca.createQuotaToStore(); err != nil {
		ca.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}
	ca.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
	return
}
