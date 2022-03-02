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
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/clusterops"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"

	k8scorev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DeleteAction action for delete action
type DeleteAction struct {
	ctx   context.Context
	model store.ClusterManagerModel
	k8sop *clusterops.K8SOperator
	req   *cmproto.DeleteNamespaceQuotaReq
	resp  *cmproto.DeleteNamespaceQuotaResp
	ns    *cmproto.Namespace
}

// NewDeleteAction delete action for delete quota
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

func (da *DeleteAction) getNamespaceFromStore() error {
	ns, err := da.model.GetNamespace(da.ctx, da.req.Namespace, da.req.FederationClusterID)
	if err != nil {
		return err
	}
	da.ns = ns
	return nil
}

func (da *DeleteAction) listPodsFromCluster() (*k8scorev1.PodList, error) {
	kubeClient, err := da.k8sop.GetClusterClient(da.req.ClusterID)
	if err != nil {
		return nil, err
	}
	podList, err := kubeClient.CoreV1().Pods(da.req.Namespace).List(da.ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return podList, nil
}

func (da *DeleteAction) deleteFromCluster() error {
	kubeClient, err := da.k8sop.GetClusterClient(da.req.ClusterID)
	if err != nil {
		return err
	}
	err = kubeClient.CoreV1().ResourceQuotas(da.req.Namespace).Delete(da.ctx, da.req.Namespace, metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	err = kubeClient.CoreV1().Namespaces().Delete(da.ctx, da.req.Namespace, metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (da *DeleteAction) deleteFromStore() error {
	return da.model.DeleteQuota(da.ctx, da.req.Namespace, da.req.FederationClusterID, da.req.ClusterID)
}

func (da *DeleteAction) setResp(code uint32, msg string) {
	da.resp.Code = code
	da.resp.Message = msg
	da.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

// Handle handle namespace quota delete request
func (da *DeleteAction) Handle(
	ctx context.Context, req *cmproto.DeleteNamespaceQuotaReq, resp *cmproto.DeleteNamespaceQuotaResp) {
	if req == nil || resp == nil {
		blog.Errorf("delete namespace quota failed, req or resp is empty")
		return
	}
	da.ctx = ctx
	da.req = req
	da.resp = resp

	if err := da.validate(); err != nil {
		da.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}
	if err := da.getNamespaceFromStore(); err != nil {
		da.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}
	if !req.IsForced {
		podList, err := da.listPodsFromCluster()
		if err != nil {
			da.setResp(common.BcsErrClusterManagerK8SOpsFailed, err.Error())
			return
		}
		if podList != nil && len(podList.Items) != 0 {
			da.setResp(common.BcsErrClusterManagerK8SOpsFailed, fmt.Sprintf(
				"there is still pods in namespace %s/%s, cannot delete quota",
				req.ClusterID, req.Namespace))
			return
		}
	}
	if err := da.deleteFromCluster(); err != nil {
		da.setResp(common.BcsErrClusterManagerK8SOpsFailed, err.Error())
		return
	}
	if err := da.deleteFromStore(); err != nil {
		da.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}
	da.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
	return
}
