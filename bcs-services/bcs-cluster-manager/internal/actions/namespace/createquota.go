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

package namespace

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

// CreateQuotaAction action for creating namespace with quota
type CreateQuotaAction struct {
	ctx              context.Context
	model            store.ClusterManagerModel
	k8sop            *clusterops.K8SOperator
	req              *cmproto.CreateNamespaceWithQuotaReq
	resp             *cmproto.CreateNamespaceWithQuotaResp
	quota            *k8scorev1.ResourceQuota
	allocatedCluster string
}

// NewCreateQuotaAction create action for creating namespace with quota
func NewCreateQuotaAction(model store.ClusterManagerModel, k8sop *clusterops.K8SOperator) *CreateQuotaAction {
	return &CreateQuotaAction{
		model: model,
		k8sop: k8sop,
	}
}

func (cqa *CreateQuotaAction) validate() error {
	if err := cqa.req.Validate(); err != nil {
		return err
	}
	quota := &k8scorev1.ResourceQuota{}
	if err := json.Unmarshal([]byte(cqa.req.ResourceQuota), quota); err != nil {
		return fmt.Errorf("decode resourcequota failed, err %s", err)
	}
	if quota.Name != cqa.req.Name || quota.Namespace != cqa.req.Name {
		return fmt.Errorf("resource quota name and namespace should be the name of namespace %s", cqa.req.Name)
	}
	cqa.quota = quota
	return nil
}

func (cqa *CreateQuotaAction) createNamespace() error {
	_, err := cqa.model.GetCluster(cqa.ctx, cqa.req.FederationClusterID)
	if err != nil {
		return fmt.Errorf("failed to find federation cluster %s, err %s", cqa.req.FederationClusterID, err.Error())
	}
	if len(cqa.req.MaxQuota) != 0 {
		maxQuota := &k8scorev1.ResourceQuota{}
		if err := json.Unmarshal([]byte(cqa.req.MaxQuota), maxQuota); err != nil {
			blog.Warnf("decode max quota %s to k8s ResourceQuota failed, err %s", cqa.req.MaxQuota, err.Error())
			return fmt.Errorf("decode max quota %s to k8s ResourceQuota failed, err %s", cqa.req.MaxQuota, err.Error())
		}
	}
	now := time.Now().Format(time.RFC3339)
	newNs := &cmproto.Namespace{
		Name:                cqa.req.Name,
		FederationClusterID: cqa.req.FederationClusterID,
		ProjectID:           cqa.req.ProjectID,
		BusinessID:          cqa.req.BusinessID,
		Labels:              cqa.req.Labels,
		MaxQuota:            cqa.req.MaxQuota,
		CreateTime:          now,
		UpdateTime:          now,
	}
	return cqa.model.CreateNamespace(cqa.ctx, newNs)
}

func (cqa *CreateQuotaAction) deleteNamespace() error {
	return cqa.model.DeleteNamespace(cqa.ctx, cqa.req.Name, cqa.req.FederationClusterID)
}

func (cqa *CreateQuotaAction) listNodesFromCluster(cluster string) (*k8scorev1.NodeList, error) {
	kubeClient, err := cqa.k8sop.GetClusterClient(cluster)
	if err != nil {
		return nil, err
	}
	nodeList, err := kubeClient.CoreV1().Nodes().List(cqa.ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return nodeList, err
}

func (cqa *CreateQuotaAction) listQuotasByCluster(cluster string) ([]cmproto.ResourceQuota, error) {
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		"federationclusterid": cqa.req.FederationClusterID,
		"clusterid":           cluster,
	})
	quotaList, err := cqa.model.ListQuota(cqa.ctx, cond, &storeopt.ListOption{})
	if err != nil {
		return nil, err
	}
	return quotaList, nil
}

// allocate one cluster for the request resource quota
func (cqa *CreateQuotaAction) allocateOneCluster() error {
	// 计算已经分配的quota与集群总资源的比值，最后算出比值最小的集群
	condM := operator.M{
		"region":              cqa.req.Region,
		"federationclusterid": cqa.req.FederationClusterID,
		"enginetype":          common.ClusterEngineTypeK8s,
		"clustertype":         common.ClusterTypeSingle,
	}
	cond := operator.NewLeafCondition(operator.Eq, condM)
	clusterList, err := cqa.model.ListCluster(cqa.ctx, cond, &storeopt.ListOption{})
	if err != nil {
		return err
	}
	minResRate := float32(math.MaxFloat32)
	targetCluster := ""
	for _, cluster := range clusterList {
		nodes, err := cqa.listNodesFromCluster(cluster.ClusterID)
		if err != nil {
			blog.Warnf("failed to list nodes from cluster %s, continue to check next cluster, err %s",
				cluster.ClusterID, err.Error())
			continue
		}
		quotas, err := cqa.listQuotasByCluster(cluster.ClusterID)
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
	cqa.allocatedCluster = targetCluster
	return nil
}

func (cqa *CreateQuotaAction) isQuotaExisted(clusterID string) (bool, error) {
	quota, err := cqa.model.GetQuota(cqa.ctx, cqa.req.Name, cqa.req.FederationClusterID, clusterID)
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

// create namespace resoucequota to store
func (cqa *CreateQuotaAction) createQuotaToStore() error {
	createTime := time.Now().Format(time.RFC3339)
	cqa.quota.ClusterName = cqa.allocatedCluster
	newQuota := &cmproto.ResourceQuota{
		Namespace:           cqa.req.Name,
		FederationClusterID: cqa.req.FederationClusterID,
		ClusterID:           cqa.allocatedCluster,
		Region:              cqa.req.Region,
		ResourceQuota:       cqa.req.ResourceQuota,
		CreateTime:          createTime,
		UpdateTime:          createTime,
	}
	if err := cqa.model.CreateQuota(cqa.ctx, newQuota); err != nil {
		return err
	}
	return nil
}

func (cqa *CreateQuotaAction) setResp(code uint32, msg string) {
	cqa.resp.Code = code
	cqa.resp.Message = msg
	cqa.resp.Result = (code == common.BcsErrClusterManagerSuccess)
	cqa.resp.Data = &cmproto.CreateNamespaceWithQuotaResp_CreateNamespaceWithQuotaRespData{
		ClusterID: cqa.allocatedCluster,
	}
}

// Handle handle namespace with quota request
func (cqa *CreateQuotaAction) Handle(
	ctx context.Context, req *cmproto.CreateNamespaceWithQuotaReq, resp *cmproto.CreateNamespaceWithQuotaResp) {
	if req == nil || resp == nil {
		blog.Errorf("create namespace with quota failed, req or resp is empty")
		return
	}
	cqa.ctx = ctx
	cqa.req = req
	cqa.resp = resp

	if err := cqa.validate(); err != nil {
		cqa.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}
	if len(cqa.req.ClusterID) == 0 {
		if err := cqa.allocateOneCluster(); err != nil {
			cqa.setResp(common.BcsErrClusterManagerAllocateClusterInCreateQuota, err.Error())
			return
		}
	} else {
		// use requested cluster
		cqa.allocatedCluster = cqa.req.ClusterID
	}

	isExisted, err := cqa.isQuotaExisted(cqa.allocatedCluster)
	if err != nil {
		cqa.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}
	if isExisted {
		cqa.setResp(common.BcsErrClusterManagerResourceDuplicated,
			fmt.Sprintf("quota %s/%s/%s is duplicated",
				cqa.req.Name, cqa.req.FederationClusterID, cqa.allocatedCluster))
		return
	}

	if err := cqa.createNamespace(); err != nil {
		if errors.Is(err, drivers.ErrTableRecordDuplicateKey) {
			cqa.setResp(common.BcsErrClusterManagerDatabaseRecordDuplicateKey, err.Error())
			return
		}
		cqa.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	if err := utils.CreateQuotaToCluster(cqa.ctx, cqa.k8sop, cqa.allocatedCluster, &k8scorev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   cqa.req.Name,
			Labels: cqa.req.Labels,
		},
	}, cqa.quota); err != nil {
		if inErr := cqa.deleteNamespace(); inErr != nil {
			blog.Warnf("rollback namespace from store failed, err %s", inErr.Error())
		}
		cqa.setResp(common.BcsErrClusterManagerK8SOpsFailed, err.Error())
		return
	}
	if err := cqa.createQuotaToStore(); err != nil {
		cqa.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}
	cqa.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
	return
}
