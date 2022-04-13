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

// Package workload 工作负载类接口实现
package workload

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	resAction "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/action/resource"
	res "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource"
	clusterRes "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/proto/cluster-resources"
)

// ListDS 获取 DaemonSet 列表
func (h *Handler) ListDS(
	ctx context.Context, req *clusterRes.ResListReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = resAction.NewResMgr(req.ProjectID, req.ClusterID, "", res.DS).List(
		ctx, req.Namespace, metav1.ListOptions{LabelSelector: req.LabelSelector},
	)
	return err
}

// GetDS 获取单个 DaemonSet
func (h *Handler) GetDS(
	ctx context.Context, req *clusterRes.ResGetReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = resAction.NewResMgr(req.ProjectID, req.ClusterID, "", res.DS).Get(
		ctx, req.Namespace, req.Name, req.Format, metav1.GetOptions{},
	)
	return err
}

// CreateDS 创建 DaemonSet
func (h *Handler) CreateDS(
	ctx context.Context, req *clusterRes.ResCreateReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = resAction.NewResMgr(req.ProjectID, req.ClusterID, "", res.DS).Create(
		ctx, req.RawData, req.Format, true, metav1.CreateOptions{},
	)
	return err
}

// UpdateDS 更新 DaemonSet
func (h *Handler) UpdateDS(
	ctx context.Context, req *clusterRes.ResUpdateReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = resAction.NewResMgr(req.ProjectID, req.ClusterID, "", res.DS).Update(
		ctx, req.Namespace, req.Name, req.RawData, req.Format, metav1.UpdateOptions{},
	)
	return err
}

// DeleteDS 删除 DaemonSet
func (h *Handler) DeleteDS(
	ctx context.Context, req *clusterRes.ResDeleteReq, _ *clusterRes.CommonResp,
) error {
	return resAction.NewResMgr(req.ProjectID, req.ClusterID, "", res.DS).Delete(
		ctx, req.Namespace, req.Name, metav1.DeleteOptions{},
	)
}
