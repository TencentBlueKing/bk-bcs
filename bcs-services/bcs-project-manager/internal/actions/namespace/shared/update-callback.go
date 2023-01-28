/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * 	http://opensource.org/licenses/MIT
 *
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package shared

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component/clientset"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/logging"
	nsm "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/namespace"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/errorx"
	quotautils "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/quota"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/proto/bcsproject"
)

// UpdateNamespaceCallback implement for UpdateNamespaceCallback interface
func (a *SharedNamespaceAction) UpdateNamespaceCallback(ctx context.Context,
	req *proto.NamespaceCallbackRequest, resp *proto.NamespaceCallbackResponse) error {
	if !req.GetApproveResult() {
		return a.model.DeleteNamespace(ctx, req.GetProjectCode(), req.GetClusterID(), req.GetNamespace())
	}
	namespace, err := a.model.GetNamespaceByItsmTicketType(ctx, req.GetProjectCode(), req.GetClusterID(),
		req.GetNamespace(), nsm.ItsmTicketTypeUpdate)
	if err != nil {
		logging.Error("get namespace %s/%s from db failed, err: %s", req.GetClusterID(), req.GetNamespace(), err.Error())
		return errorx.NewDBErr(err.Error())
	}
	if req.GetApplyInCluster() {
		if err := updateNamespaceQuotaInCluster(ctx, namespace); err != nil {
			return err
		}
	}

	// delete namespace in db
	if err := a.model.DeleteNamespace(ctx, req.GetProjectCode(), req.GetClusterID(), req.GetNamespace()); err != nil {
		logging.Error("delete namespace %s/%s from db failed, err: %s", req.GetClusterID(), req.GetNamespace(), err.Error())
		return errorx.NewDBErr(err.Error())
	}
	resp.Code = 0
	resp.Message = "ok"
	return nil
}

func updateNamespaceQuotaInCluster(ctx context.Context, ns *nsm.Namespace) error {
	client, err := clientset.GetClientGroup().Client(ns.ClusterID)
	if err != nil {
		logging.Error("get clientset for cluster %s failed, err: %s", ns.ClusterID, err.Error())
		return err
	}
	// get old quota
	oldQuota, err := client.CoreV1().ResourceQuotas(ns.Name).Get(ctx, ns.Name, metav1.GetOptions{})
	if err != nil && !errors.IsNotFound(err) {
		logging.Error("get resourceQuota %s/%s failed, err: %s", ns.ClusterID, ns.Name, err.Error())
		return errorx.NewClusterErr(err.Error())
	}
	// update for create in cluster
	if errors.IsNotFound(err) {
		if ns.ResourceQuota == nil {
			return nil
		}
		quota := &corev1.ResourceQuota{
			Spec: corev1.ResourceQuotaSpec{
				Hard: corev1.ResourceList{},
			},
		}
		quota.SetName(ns.Name)
		quota.SetNamespace(ns.Name)

		if lErr := quotautils.LoadFromModel(quota, ns.ResourceQuota); lErr != nil {
			return err
		}

		_, err = client.CoreV1().ResourceQuotas(ns.Name).Create(ctx, quota, metav1.CreateOptions{})
		if err != nil {
			logging.Error("create resourceQuota %s/%s failed, err: %s", ns.ClusterID, ns.Name, err.Error())
			return errorx.NewClusterErr(err.Error())
		}
		return nil
	}
	// update for delete in cluster
	if ns.ResourceQuota == nil {
		return client.CoreV1().ResourceQuotas(ns.Name).Delete(ctx, ns.Name, metav1.DeleteOptions{})
	}
	// update for update in cluster
	if lErr := quotautils.LoadFromModel(oldQuota, ns.ResourceQuota); lErr != nil {
		return err
	}
	_, err = client.CoreV1().ResourceQuotas(ns.Name).Update(ctx, oldQuota, metav1.UpdateOptions{})
	if err != nil {
		logging.Error("update resourceQuota %s/%s failed, err: %s", ns.ClusterID, ns.Name, err.Error())
		return errorx.NewClusterErr(err.Error())
	}
	return nil
}
