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

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/action/perm"
	resAction "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/action/resource"
	respUtil "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/action/resp"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/action/web"
	resCsts "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/constants"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/pbstruct"
	clusterRes "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/proto/cluster-resources"
)

// Handler ...
type Handler struct{}

// New ...
func New() *Handler {
	return &Handler{}
}

// ListCM 获取 ConfigMap 列表
func (h *Handler) ListCM(
	ctx context.Context, req *clusterRes.ResListReq, resp *clusterRes.CommonResp,
) (err error) {
	if err = perm.CheckNSAccess(ctx, req.ClusterID, req.Namespace); err != nil {
		return err
	}
	respData, err := respUtil.BuildListAPIRespData(
		ctx, respUtil.ListParams{
			req.ClusterID, resCsts.CM, req.ApiVersion, req.Namespace, req.Format, req.Scene,
		}, metav1.ListOptions{LabelSelector: req.LabelSelector},
	)
	if err != nil {
		return err
	}
	if resp.Data, err = pbstruct.Map2pbStruct(respData); err != nil {
		return err
	}
	resp.WebAnnotations, err = web.GenListConfigWebAnnos(ctx, respData)
	return err
}

// GetCM 获取单个 ConfigMap
func (h *Handler) GetCM(
	ctx context.Context, req *clusterRes.ResGetReq, resp *clusterRes.CommonResp,
) (err error) {
	if err = perm.CheckNSAccess(ctx, req.ClusterID, req.Namespace); err != nil {
		return err
	}
	respData, err := respUtil.BuildRetrieveAPIRespData(
		ctx, respUtil.GetParams{
			req.ClusterID, resCsts.CM, req.ApiVersion, req.Namespace, req.Name, req.Format,
		}, metav1.GetOptions{},
	)
	if err != nil {
		return err
	}
	if resp.Data, err = pbstruct.Map2pbStruct(respData); err != nil {
		return err
	}
	resp.WebAnnotations, err = web.GenRetrieveConfigWebAnnos(ctx, respData)
	return err
}

// CreateCM 创建 ConfigMap
func (h *Handler) CreateCM(
	ctx context.Context, req *clusterRes.ResCreateReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = resAction.NewResMgr(req.ClusterID, "", resCsts.CM).Create(
		ctx, req.RawData, req.Format, true, metav1.CreateOptions{},
	)
	return err
}

// UpdateCM 更新 ConfigMap
func (h *Handler) UpdateCM(
	ctx context.Context, req *clusterRes.ResUpdateReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = resAction.NewResMgr(req.ClusterID, "", resCsts.CM).Update(
		ctx, req.Namespace, req.Name, req.RawData, req.Format, metav1.UpdateOptions{},
	)
	return err
}

// DeleteCM 删除 ConfigMap
func (h *Handler) DeleteCM(
	ctx context.Context, req *clusterRes.ResDeleteReq, _ *clusterRes.CommonResp,
) error {
	return resAction.NewResMgr(req.ClusterID, "", resCsts.CM).Delete(
		ctx, req.Namespace, req.Name, metav1.DeleteOptions{},
	)
}
