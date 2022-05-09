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
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/clusterops"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	storeopt "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/options"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DeleteAction action for delete namespace
type DeleteAction struct {
	ctx       context.Context
	model     store.ClusterManagerModel
	k8sop     *clusterops.K8SOperator
	req       *cmproto.DeleteNamespaceReq
	resp      *cmproto.DeleteNamespaceResp
	quotaList []cmproto.ResourceQuota
}

// NewDeleteAction delete namespace
func NewDeleteAction(model store.ClusterManagerModel, k8sop *clusterops.K8SOperator) *DeleteAction {
	return &DeleteAction{
		model: model,
		k8sop: k8sop,
	}
}

func (da *DeleteAction) validate() error {
	if err := da.req.Validate(); err != nil {
		return err
	}
	return nil
}

func (da *DeleteAction) listQuotas() error {
	quotaList, err := da.model.ListQuota(da.ctx, operator.NewLeafCondition(operator.Eq, operator.M{
		"namespace":           da.req.Name,
		"federationClusterID": da.req.FederationClusterID,
	}), &storeopt.ListOption{})
	if err != nil {
		return err
	}
	da.quotaList = quotaList
	return nil
}

func (da *DeleteAction) deleteQuotaFromCluster(quota *cmproto.ResourceQuota) error {
	kubeClient, err := da.k8sop.GetClusterClient(quota.ClusterID)
	if err != nil {
		return err
	}
	err = kubeClient.CoreV1().ResourceQuotas(quota.Namespace).Delete(da.ctx, quota.Namespace, metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	err = kubeClient.CoreV1().Namespaces().Delete(da.ctx, quota.Namespace, metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (da *DeleteAction) deleteQuotaFromStore(quota *cmproto.ResourceQuota) error {
	return da.model.DeleteQuota(da.ctx, quota.Namespace, quota.FederationClusterID, quota.ClusterID)
}

func (da *DeleteAction) deleteQuotaList() error {
	for _, tmpQuota := range da.quotaList {
		if err := da.deleteQuotaFromCluster(&tmpQuota); err != nil {
			blog.Errorf("delete quota %s/%s/%s from cluster failed, err %s",
				tmpQuota.Namespace, tmpQuota.FederationClusterID, tmpQuota.ClusterID, err.Error())
			return fmt.Errorf("delete quota %s/%s/%s from cluster failed, err %s",
				tmpQuota.Namespace, tmpQuota.FederationClusterID, tmpQuota.ClusterID, err.Error())
		}
		if err := da.deleteQuotaFromStore(&tmpQuota); err != nil {
			blog.Errorf("delete quota %s/%s/%s from store failed, err %s",
				tmpQuota.Namespace, tmpQuota.FederationClusterID, tmpQuota.ClusterID, err.Error())
			return fmt.Errorf("delete quota %s/%s/%s from store failed, err %s",
				tmpQuota.Namespace, tmpQuota.FederationClusterID, tmpQuota.ClusterID, err.Error())
		}
	}
	return nil
}

func (da *DeleteAction) deleteNamespaceFromStore() error {
	return da.model.DeleteNamespace(da.ctx, da.req.Name, da.req.FederationClusterID)
}

func (da *DeleteAction) setResp(code uint32, msg string) {
	da.resp.Code = code
	da.resp.Message = msg
	da.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

// Handle handle delete namespace reqeust
func (da *DeleteAction) Handle(ctx context.Context,
	req *cmproto.DeleteNamespaceReq, resp *cmproto.DeleteNamespaceResp) {
	if req == nil || resp == nil {
		blog.Errorf("delete namespace failed, req or resp is empty")
		return
	}
	da.ctx = ctx
	da.req = req
	da.resp = resp

	if err := da.validate(); err != nil {
		da.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}
	if err := da.listQuotas(); err != nil {
		da.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}
	if !da.req.IsForced {
		if len(da.quotaList) != 0 {
			da.setResp(common.BcsErrClusterManagerDBOperation,
				"cannot delete namespace which has resourcequota")
			return
		}
	} else {
		if err := da.deleteQuotaList(); err != nil {
			da.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
			return
		}
	}
	if err := da.deleteNamespaceFromStore(); err != nil {
		da.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}
	da.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
	return
}
