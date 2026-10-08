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
	limitrangeutils "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/limitrange"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/proto/bcsproject"
)

// ListPodLimitRanges lists LimitRanges containing a Pod item in a namespace.
func (c *IndependentNamespaceAction) ListPodLimitRanges(ctx context.Context,
	req *proto.ListPodLimitRangesRequest, resp *proto.ListPodLimitRangesResponse) error {
	client, err := clientset.GetClientGroup().Client(req.GetClusterID())
	if err != nil {
		logging.Error("get clientset for cluster %s failed, err: %s", req.GetClusterID(), err.Error())
		return err
	}
	limitRangeList, err := client.CoreV1().LimitRanges(req.GetNamespace()).List(ctx, metav1.ListOptions{})
	if err != nil {
		logging.Error("list Pod LimitRanges %s/%s failed, err: %s", req.GetClusterID(), req.GetNamespace(),
			err.Error())
		return errorx.NewClusterErr(err.Error())
	}
	resp.Data = make([]*proto.PodLimitRange, 0, len(limitRangeList.Items))
	for index := range limitRangeList.Items {
		if limitRange := limitrangeutils.TransferToProto(&limitRangeList.Items[index]); limitRange != nil {
			resp.Data = append(resp.Data, limitRange)
		}
	}
	return nil
}

// CreatePodLimitRange creates a LimitRange containing a Pod item.
func (c *IndependentNamespaceAction) CreatePodLimitRange(ctx context.Context,
	req *proto.CreatePodLimitRangeRequest, resp *proto.PodLimitRangeResponse) error {
	limitRange := &corev1.LimitRange{
		ObjectMeta: metav1.ObjectMeta{Name: req.GetLimitRangeName(), Namespace: req.GetNamespace()},
	}
	if err := limitrangeutils.UpsertPodItem(limitRange, req.GetMin(), req.GetMax()); err != nil {
		return err
	}
	client, err := clientset.GetClientGroup().Client(req.GetClusterID())
	if err != nil {
		logging.Error("get clientset for cluster %s failed, err: %s", req.GetClusterID(), err.Error())
		return err
	}
	created, err := client.CoreV1().LimitRanges(req.GetNamespace()).Create(ctx, limitRange, metav1.CreateOptions{})
	if errors.IsAlreadyExists(err) {
		return errorx.NewReadableErr(errorx.ParamErr,
			fmt.Sprintf("Pod LimitRange [%s] 已存在", req.GetLimitRangeName()))
	}
	if err != nil {
		logging.Error("create Pod LimitRange %s/%s/%s failed, err: %s", req.GetClusterID(),
			req.GetNamespace(), req.GetLimitRangeName(), err.Error())
		return errorx.NewClusterErr(err.Error())
	}
	resp.Data = limitrangeutils.TransferToProto(created)
	return nil
}

// UpdatePodLimitRange updates only the Pod item and preserves other LimitRange items and metadata.
func (c *IndependentNamespaceAction) UpdatePodLimitRange(ctx context.Context,
	req *proto.UpdatePodLimitRangeRequest, resp *proto.PodLimitRangeResponse) error {
	client, err := clientset.GetClientGroup().Client(req.GetClusterID())
	if err != nil {
		logging.Error("get clientset for cluster %s failed, err: %s", req.GetClusterID(), err.Error())
		return err
	}
	limitRange, err := client.CoreV1().LimitRanges(req.GetNamespace()).
		Get(ctx, req.GetLimitRangeName(), metav1.GetOptions{})
	if errors.IsNotFound(err) {
		return errorx.NewReadableErr(errorx.ParamErr,
			fmt.Sprintf("Pod LimitRange [%s] 不存在", req.GetLimitRangeName()))
	}
	if err != nil {
		logging.Error("get Pod LimitRange %s/%s/%s failed, err: %s", req.GetClusterID(),
			req.GetNamespace(), req.GetLimitRangeName(), err.Error())
		return errorx.NewClusterErr(err.Error())
	}
	if limitrangeutils.TransferToProto(limitRange) == nil {
		return errorx.NewReadableErr(errorx.ParamErr,
			fmt.Sprintf("LimitRange [%s] 不包含 Pod 限制", req.GetLimitRangeName()))
	}
	if err = limitrangeutils.UpsertPodItem(limitRange, req.GetMin(), req.GetMax()); err != nil {
		return err
	}
	updated, err := client.CoreV1().LimitRanges(req.GetNamespace()).Update(ctx, limitRange, metav1.UpdateOptions{})
	if err != nil {
		logging.Error("update Pod LimitRange %s/%s/%s failed, err: %s", req.GetClusterID(),
			req.GetNamespace(), req.GetLimitRangeName(), err.Error())
		return errorx.NewClusterErr(err.Error())
	}
	resp.Data = limitrangeutils.TransferToProto(updated)
	return nil
}

// DeletePodLimitRange removes only the Pod item, deleting the object when no other items remain.
func (c *IndependentNamespaceAction) DeletePodLimitRange(ctx context.Context,
	req *proto.DeletePodLimitRangeRequest, resp *proto.PodLimitRangeResponse) error {
	client, err := clientset.GetClientGroup().Client(req.GetClusterID())
	if err != nil {
		logging.Error("get clientset for cluster %s failed, err: %s", req.GetClusterID(), err.Error())
		return err
	}
	limitRange, err := client.CoreV1().LimitRanges(req.GetNamespace()).
		Get(ctx, req.GetLimitRangeName(), metav1.GetOptions{})
	if errors.IsNotFound(err) {
		return errorx.NewReadableErr(errorx.ParamErr,
			fmt.Sprintf("Pod LimitRange [%s] 不存在", req.GetLimitRangeName()))
	}
	if err != nil {
		logging.Error("get Pod LimitRange %s/%s/%s failed, err: %s", req.GetClusterID(),
			req.GetNamespace(), req.GetLimitRangeName(), err.Error())
		return errorx.NewClusterErr(err.Error())
	}
	if !limitrangeutils.RemovePodItem(limitRange) {
		return errorx.NewReadableErr(errorx.ParamErr,
			fmt.Sprintf("LimitRange [%s] 不包含 Pod 限制", req.GetLimitRangeName()))
	}
	if len(limitRange.Spec.Limits) == 0 {
		err = client.CoreV1().LimitRanges(req.GetNamespace()).
			Delete(ctx, req.GetLimitRangeName(), metav1.DeleteOptions{})
	} else {
		_, err = client.CoreV1().LimitRanges(req.GetNamespace()).Update(ctx, limitRange, metav1.UpdateOptions{})
	}
	if err != nil {
		logging.Error("delete Pod item from LimitRange %s/%s/%s failed, err: %s", req.GetClusterID(),
			req.GetNamespace(), req.GetLimitRangeName(), err.Error())
		return errorx.NewClusterErr(err.Error())
	}
	return nil
}
