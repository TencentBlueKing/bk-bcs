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

// Package config 配置类接口实现
package config

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	resAction "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/action/resource"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/action/util/web"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/featureflag"
	res "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource"
	clusterRes "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/proto/cluster-resources"
)

// Handler xxx
type Handler struct{}

// New xxx
func New() *Handler {
	return &Handler{}
}

// ListCM xxx
func (h *Handler) ListCM(
	ctx context.Context, req *clusterRes.ResListReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = resAction.NewResMgr(req.ClusterID, req.ApiVersion, res.CM).List(
		ctx, req.Namespace, req.Format, metav1.ListOptions{LabelSelector: req.LabelSelector},
	)
	if err != nil {
		return err
	}
	resp.WebAnnotations, err = web.NewAnnos(
		web.NewFeatureFlag(featureflag.FormCreate, false),
	).ToPbStruct()
	return err
}

// GetCM xxx
func (h *Handler) GetCM(
	ctx context.Context, req *clusterRes.ResGetReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = resAction.NewResMgr(req.ClusterID, req.ApiVersion, res.CM).Get(
		ctx, req.Namespace, req.Name, req.Format, metav1.GetOptions{},
	)
	if err != nil {
		return err
	}
	resp.WebAnnotations, err = web.NewAnnos(
		web.NewFeatureFlag(featureflag.FormUpdate, false),
	).ToPbStruct()
	return err
}

// CreateCM xxx
func (h *Handler) CreateCM(
	ctx context.Context, req *clusterRes.ResCreateReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = resAction.NewResMgr(req.ClusterID, "", res.CM).Create(
		ctx, req.RawData, req.Format, true, metav1.CreateOptions{},
	)
	return err
}

// UpdateCM xxx
func (h *Handler) UpdateCM(
	ctx context.Context, req *clusterRes.ResUpdateReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = resAction.NewResMgr(req.ClusterID, "", res.CM).Update(
		ctx, req.Namespace, req.Name, req.RawData, req.Format, metav1.UpdateOptions{},
	)
	return err
}

// DeleteCM xxx
func (h *Handler) DeleteCM(
	ctx context.Context, req *clusterRes.ResDeleteReq, _ *clusterRes.CommonResp,
) error {
	return resAction.NewResMgr(req.ClusterID, "", res.CM).Delete(
		ctx, req.Namespace, req.Name, metav1.DeleteOptions{},
	)
}
