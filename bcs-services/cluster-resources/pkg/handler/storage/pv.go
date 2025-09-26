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

// Package storage 存储类接口实现
package storage

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	resAction "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/action/resource"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/action/web"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/featureflag"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/component/cluster"
	resCsts "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/constants"
	clusterRes "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/proto/cluster-resources"
)

// Handler xxx
type Handler struct{}

// New xxx
func New() *Handler {
	return &Handler{}
}

// ListPV xxx
func (h *Handler) ListPV(
	ctx context.Context, req *clusterRes.ResListReq, resp *clusterRes.CommonResp,
) (err error) {
	clusterInfo, err := cluster.GetClusterInfo(ctx, req.ClusterID)
	if err != nil {
		return err
	}
	// 共享集群 PV 返回空列表
	if !clusterInfo.IsShared {
		resp.Data, err = resAction.NewResMgr(req.ClusterID, req.ApiVersion, resCsts.PV).List(
			ctx, "", req.Format, req.Scene, metav1.ListOptions{LabelSelector: req.LabelSelector},
		)
		if err != nil {
			return err
		}
	}

	resp.WebAnnotations, err = web.NewAnnos(
		web.NewFeatureFlag(featureflag.FormCreate, false),
	).ToPbStruct()
	return err
}

// GetPV xxx
func (h *Handler) GetPV(
	ctx context.Context, req *clusterRes.ResGetReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = resAction.NewResMgr(req.ClusterID, req.ApiVersion, resCsts.PV).Get(
		ctx, "", req.Name, req.Format, metav1.GetOptions{},
	)
	if err != nil {
		return err
	}
	resp.WebAnnotations, err = web.NewAnnos(
		web.NewFeatureFlag(featureflag.FormUpdate, false),
	).ToPbStruct()
	return err
}

// CreatePV xxx
func (h *Handler) CreatePV(
	ctx context.Context, req *clusterRes.ResCreateReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = resAction.NewResMgr(req.ClusterID, "", resCsts.PV).Create(
		ctx, req.RawData, req.Format, false, metav1.CreateOptions{},
	)
	return err
}

// UpdatePV xxx
func (h *Handler) UpdatePV(
	ctx context.Context, req *clusterRes.ResUpdateReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = resAction.NewResMgr(req.ClusterID, "", resCsts.PV).Update(
		ctx, "", req.Name, req.RawData, req.Format, metav1.UpdateOptions{},
	)
	return err
}

// DeletePV xxx
func (h *Handler) DeletePV(
	ctx context.Context, req *clusterRes.ResDeleteReq, _ *clusterRes.CommonResp,
) error {
	return resAction.NewResMgr(req.ClusterID, "", resCsts.PV).Delete(
		ctx, "", req.Name, metav1.DeleteOptions{},
	)
}
