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

package independent

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component/clientset"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/logging"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/errorx"
	quotautils "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/quota"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/proto/bcsproject"
)

// UpdateNamespace implement for UpdateNamespace interface
func (c *IndependentNamespaceAction) UpdateNamespace(ctx context.Context,
	req *proto.UpdateNamespaceRequest, resp *proto.UpdateNamespaceResponse) error {
	if err := quotautils.ValidateResourceQuota(req.Quota); err != nil {
		return err
	}
	client, err := clientset.GetClientGroup().Client(req.GetClusterID())
	if err != nil {
		logging.Error("get clientset for cluster %s failed, err: %s", req.GetClusterID(), err.Error())
		return err
	}
	// get namespace and update
	namespace, err := client.CoreV1().Namespaces().Get(ctx, req.GetNamespace(), metav1.GetOptions{})
	if err != nil {
		logging.Error("get namespace %s/%s failed, err: %s", req.GetClusterID(), req.GetNamespace(), err.Error())
		return errorx.NewClusterErr(err.Error())
	}
	labels := map[string]string{}
	for _, label := range req.GetLabels() {
		labels[label.GetKey()] = label.GetValue()
	}
	namespace.SetLabels(labels)
	annotations := map[string]string{}
	for _, annotation := range req.GetAnnotations() {
		annotations[annotation.GetKey()] = annotation.GetValue()
	}
	namespace.SetAnnotations(annotations)
	_, err = client.CoreV1().Namespaces().Update(ctx, namespace, metav1.UpdateOptions{})
	if err != nil {
		logging.Error("update namespace %s/%s failed, err: %s", req.GetClusterID(), req.GetNamespace(), err.Error())
		return errorx.NewClusterErr(err.Error())
	}

	// get old quota
	oldQuota, err := client.CoreV1().ResourceQuotas(req.GetNamespace()).Get(ctx, req.GetNamespace(), metav1.GetOptions{})
	if err != nil && !errors.IsNotFound(err) {
		logging.Error("get resourceQuota %s/%s failed, err: %s", req.GetClusterID(), req.GetNamespace(), err.Error())
		return errorx.NewClusterErr(err.Error())
	}

	// update for create
	if errors.IsNotFound(err) {
		if req.GetQuota() == nil {
			return nil
		}
		quota := &corev1.ResourceQuota{
			Spec: corev1.ResourceQuotaSpec{
				Hard: corev1.ResourceList{},
			},
		}
		quota.SetName(req.GetNamespace())
		quota.SetNamespace(req.GetNamespace())

		if lErr := quotautils.LoadFromProto(quota, req.GetQuota()); lErr != nil {
			return err
		}

		_, err = client.CoreV1().ResourceQuotas(req.GetNamespace()).Create(ctx, quota, metav1.CreateOptions{})
		if err != nil {
			logging.Error("create resourceQuota %s/%s failed, err: %s", req.GetClusterID(), req.GetNamespace(), err.Error())
			return errorx.NewClusterErr(err.Error())
		}
		return nil
	}
	// update for delete
	if req.GetQuota() == nil {
		return client.CoreV1().ResourceQuotas(req.GetNamespace()).Delete(ctx, req.GetNamespace(), metav1.DeleteOptions{})
	}

	// update for update
	if lErr := quotautils.LoadFromProto(oldQuota, req.GetQuota()); lErr != nil {
		return err
	}

	_, err = client.CoreV1().ResourceQuotas(req.GetNamespace()).Update(ctx, oldQuota, metav1.UpdateOptions{})
	if err != nil {
		logging.Error("update resourceQuota %s/%s failed, err: %s", req.GetClusterID(), req.GetNamespace(), err.Error())
		return errorx.NewClusterErr(err.Error())
	}
	return nil
}
