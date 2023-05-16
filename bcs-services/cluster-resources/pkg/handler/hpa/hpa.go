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

// Package hpa HPA 接口实现
package hpa

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	resAction "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/action/resource"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/action/web"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/featureflag"
	resCsts "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/constants"
	clusterRes "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/proto/cluster-resources"
)

// Handler xxx
type Handler struct{}

// New xxx
func New() *Handler {
	return &Handler{}
}

// ListHPA xxx
func (h *Handler) ListHPA(
	ctx context.Context, req *clusterRes.ResListReq, resp *clusterRes.CommonResp,
) (err error) {
	if req.ApiVersion == "" {
		req.ApiVersion = resCsts.DefaultHPAGroupVersion
	}
	resp.Data, err = resAction.NewResMgr(req.ClusterID, req.ApiVersion, resCsts.HPA).List(
		ctx, req.Namespace, req.Format, req.Scene, metav1.ListOptions{LabelSelector: req.LabelSelector},
	)
	if err != nil {
		return err
	}
	resp.WebAnnotations, err = web.NewAnnos(
		web.NewFeatureFlag(featureflag.FormCreate, true),
	).ToPbStruct()
	return err
}

// GetHPA xxx
func (h *Handler) GetHPA(
	ctx context.Context, req *clusterRes.ResGetReq, resp *clusterRes.CommonResp,
) (err error) {
	if req.ApiVersion == "" {
		req.ApiVersion = resCsts.DefaultHPAGroupVersion
	}
	resp.Data, err = resAction.NewResMgr(req.ClusterID, req.ApiVersion, resCsts.HPA).Get(
		ctx, req.Namespace, req.Name, req.Format, metav1.GetOptions{},
	)
	if err != nil {
		return err
	}
	resp.WebAnnotations, err = web.NewAnnos(
		web.NewFeatureFlag(featureflag.FormUpdate, true),
	).ToPbStruct()
	return err
}

// CreateHPA xxx
func (h *Handler) CreateHPA(
	ctx context.Context, req *clusterRes.ResCreateReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = resAction.NewResMgr(req.ClusterID, "", resCsts.HPA).Create(
		ctx, req.RawData, req.Format, true, metav1.CreateOptions{},
	)
	return err
}

// UpdateHPA xxx
func (h *Handler) UpdateHPA(
	ctx context.Context, req *clusterRes.ResUpdateReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = resAction.NewResMgr(req.ClusterID, "", resCsts.HPA).Update(
		ctx, req.Namespace, req.Name, req.RawData, req.Format, metav1.UpdateOptions{},
	)
	return err
}

// DeleteHPA xxx
func (h *Handler) DeleteHPA(
	ctx context.Context, req *clusterRes.ResDeleteReq, _ *clusterRes.CommonResp,
) error {
	return resAction.NewResMgr(req.ClusterID, "", resCsts.HPA).Delete(
		ctx, req.Namespace, req.Name, metav1.DeleteOptions{},
	)
}
