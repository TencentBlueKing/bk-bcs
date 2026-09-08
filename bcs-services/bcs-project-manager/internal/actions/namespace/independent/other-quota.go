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
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component/clientset"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/logging"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/errorx"
	quotautils "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/quota"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/proto/bcsproject"
)

// CreateOtherQuota creates a ResourceQuota whose name differs from the namespace name.
func (c *IndependentNamespaceAction) CreateOtherQuota(ctx context.Context,
	req *proto.CreateOtherQuotaRequest, resp *proto.OtherQuotaResponse) error {
	if err := validateOtherQuotaRequest(req.GetNamespace(), req.GetQuotaName(), req.GetQuota()); err != nil {
		return err
	}
	client, err := clientset.GetClientGroup().Client(req.GetClusterID())
	if err != nil {
		logging.Error("get clientset for cluster %s failed, err: %s", req.GetClusterID(), err.Error())
		return err
	}
	quota := &corev1.ResourceQuota{
		ObjectMeta: metav1.ObjectMeta{Name: req.GetQuotaName(), Namespace: req.GetNamespace()},
		Spec:       corev1.ResourceQuotaSpec{Hard: corev1.ResourceList{}},
	}
	if err = quotautils.LoadFromProto(quota, req.GetQuota()); err != nil {
		return errorx.NewParamErr(err.Error())
	}
	created, err := client.CoreV1().ResourceQuotas(req.GetNamespace()).Create(ctx, quota, metav1.CreateOptions{})
	if errors.IsAlreadyExists(err) {
		return errorx.NewReadableErr(errorx.ParamErr, fmt.Sprintf("配额 [%s] 已存在", req.GetQuotaName()))
	}
	if err != nil {
		logging.Error("create resourceQuota %s/%s/%s failed, err: %s", req.GetClusterID(),
			req.GetNamespace(), req.GetQuotaName(), err.Error())
		return errorx.NewClusterErr(err.Error())
	}
	resp.Data = quotautils.TransferToProtoOtherQuota(created)
	return nil
}

// UpdateOtherQuota updates the supported hard limits and preserves other ResourceQuota fields.
func (c *IndependentNamespaceAction) UpdateOtherQuota(ctx context.Context,
	req *proto.UpdateOtherQuotaRequest, resp *proto.OtherQuotaResponse) error {
	if err := validateOtherQuotaRequest(req.GetNamespace(), req.GetQuotaName(), req.GetQuota()); err != nil {
		return err
	}
	client, err := clientset.GetClientGroup().Client(req.GetClusterID())
	if err != nil {
		logging.Error("get clientset for cluster %s failed, err: %s", req.GetClusterID(), err.Error())
		return err
	}
	quota, err := client.CoreV1().ResourceQuotas(req.GetNamespace()).Get(ctx, req.GetQuotaName(), metav1.GetOptions{})
	if errors.IsNotFound(err) {
		return errorx.NewReadableErr(errorx.ParamErr, fmt.Sprintf("配额 [%s] 不存在", req.GetQuotaName()))
	}
	if err != nil {
		logging.Error("get resourceQuota %s/%s/%s failed, err: %s", req.GetClusterID(),
			req.GetNamespace(), req.GetQuotaName(), err.Error())
		return errorx.NewClusterErr(err.Error())
	}
	if err = quotautils.LoadFromProto(quota, req.GetQuota()); err != nil {
		return errorx.NewParamErr(err.Error())
	}
	updated, err := client.CoreV1().ResourceQuotas(req.GetNamespace()).Update(ctx, quota, metav1.UpdateOptions{})
	if err != nil {
		logging.Error("update resourceQuota %s/%s/%s failed, err: %s", req.GetClusterID(),
			req.GetNamespace(), req.GetQuotaName(), err.Error())
		return errorx.NewClusterErr(err.Error())
	}
	resp.Data = quotautils.TransferToProtoOtherQuota(updated)
	return nil
}

// DeleteOtherQuota deletes a ResourceQuota whose name differs from the namespace name.
func (c *IndependentNamespaceAction) DeleteOtherQuota(ctx context.Context,
	req *proto.DeleteOtherQuotaRequest, resp *proto.OtherQuotaResponse) error {
	if err := validateOtherQuotaName(req.GetNamespace(), req.GetQuotaName()); err != nil {
		return err
	}
	client, err := clientset.GetClientGroup().Client(req.GetClusterID())
	if err != nil {
		logging.Error("get clientset for cluster %s failed, err: %s", req.GetClusterID(), err.Error())
		return err
	}
	err = client.CoreV1().ResourceQuotas(req.GetNamespace()).Delete(ctx, req.GetQuotaName(), metav1.DeleteOptions{})
	if errors.IsNotFound(err) {
		return errorx.NewReadableErr(errorx.ParamErr, fmt.Sprintf("配额 [%s] 不存在", req.GetQuotaName()))
	}
	if err != nil {
		logging.Error("delete resourceQuota %s/%s/%s failed, err: %s", req.GetClusterID(),
			req.GetNamespace(), req.GetQuotaName(), err.Error())
		return errorx.NewClusterErr(err.Error())
	}
	return nil
}

func validateOtherQuotaRequest(namespace, quotaName string, quota *proto.ResourceQuota) error {
	if err := validateOtherQuotaName(namespace, quotaName); err != nil {
		return err
	}
	if quota == nil {
		return errorx.NewParamErr("quota cannot be empty")
	}
	if quota.GetCpuLimits() == "" && quota.GetCpuRequests() == "" &&
		quota.GetMemoryLimits() == "" && quota.GetMemoryRequests() == "" {
		return errorx.NewParamErr("quota resources cannot all be empty")
	}
	return nil
}

func validateOtherQuotaName(namespace, quotaName string) error {
	if namespace == quotaName {
		return errorx.NewReadableErr(errorx.ParamErr, "其他配额名称不能与命名空间名称相同")
	}
	return nil
}
