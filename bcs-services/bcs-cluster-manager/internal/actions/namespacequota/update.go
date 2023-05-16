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
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/clusterops"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"

	k8scorev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// UpdateAction action for updating namespace quota
type UpdateAction struct {
	ctx     context.Context
	model   store.ClusterManagerModel
	k8sop   *clusterops.K8SOperator
	req     *cmproto.UpdateNamespaceQuotaReq
	resp    *cmproto.UpdateNamespaceQuotaResp
	ns      *cmproto.Namespace
	dbQuota *cmproto.ResourceQuota
	quota   *k8scorev1.ResourceQuota
}

// NewUpdateAction create action for updating namespace quota
func NewUpdateAction(model store.ClusterManagerModel, k8sop *clusterops.K8SOperator) *UpdateAction {
	return &UpdateAction{
		model: model,
		k8sop: k8sop,
	}
}

func (ua *UpdateAction) validate() error {
	if err := ua.req.Validate(); err != nil {
		return err
	}
	quota := &k8scorev1.ResourceQuota{}
	if err := json.Unmarshal([]byte(ua.req.ResourceQuota), quota); err != nil {
		return fmt.Errorf("decode resourcequota failed, err %s", err)
	}
	if quota.Name != ua.req.Namespace || quota.Namespace != ua.req.Namespace {
		return fmt.Errorf("resource quota name and namespace should be the name of namespace %s", ua.req.Namespace)
	}
	ua.quota = quota
	return nil
}

func (ua *UpdateAction) getNamespaceFromStore() error {
	ns, err := ua.model.GetNamespace(ua.ctx, ua.req.Namespace, ua.req.FederationClusterID)
	if err != nil {
		return err
	}
	ua.ns = ns
	return nil
}

func (ua *UpdateAction) getQuotaFromStore() error {
	quota, err := ua.model.GetQuota(ua.ctx, ua.req.Namespace, ua.req.FederationClusterID, ua.req.ClusterID)
	if err != nil {
		return err
	}
	ua.dbQuota = quota
	return nil
}

func (ua *UpdateAction) updateQuotaToCluster() error {
	kubeClient, err := ua.k8sop.GetClusterClient(ua.req.ClusterID)
	if err != nil {
		return err
	}
	existedQuota, err := kubeClient.CoreV1().ResourceQuotas(ua.req.Namespace).Get(
		ua.ctx, ua.quota.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	ua.quota.SetResourceVersion(existedQuota.GetResourceVersion())
	_, err = kubeClient.CoreV1().ResourceQuotas(ua.req.Namespace).Update(ua.ctx, ua.quota, metav1.UpdateOptions{})
	return err
}

func (ua *UpdateAction) updateQuotaToStore() error {
	newQuota := &cmproto.ResourceQuota{
		Namespace:           ua.req.Namespace,
		FederationClusterID: ua.req.FederationClusterID,
		ClusterID:           ua.req.ClusterID,
		Region:              ua.dbQuota.Region,
		ResourceQuota:       ua.req.ResourceQuota,
		CreateTime:          ua.dbQuota.CreateTime,
		UpdateTime:          time.Now().Format(time.RFC3339),
	}
	if err := ua.model.UpdateQuota(ua.ctx, newQuota); err != nil {
		return err
	}
	return nil
}

func (ua *UpdateAction) setResp(code uint32, msg string) {
	ua.resp.Code = code
	ua.resp.Message = msg
	ua.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

// Handle handle updating quota request
func (ua *UpdateAction) Handle(
	ctx context.Context, req *cmproto.UpdateNamespaceQuotaReq, resp *cmproto.UpdateNamespaceQuotaResp) {
	if req == nil || resp == nil {
		blog.Errorf("update namespace quota failed, req or resp is empty")
		return
	}
	ua.ctx = ctx
	ua.req = req
	ua.resp = resp

	if err := ua.validate(); err != nil {
		ua.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}
	if err := ua.getNamespaceFromStore(); err != nil {
		ua.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}
	if err := ua.getQuotaFromStore(); err != nil {
		ua.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}
	if err := ua.updateQuotaToCluster(); err != nil {
		ua.setResp(common.BcsErrClusterManagerK8SOpsFailed, err.Error())
		return
	}
	if err := ua.updateQuotaToStore(); err != nil {
		ua.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}
	ua.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
	return
}
