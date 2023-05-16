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

package network

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	resAction "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/action/resource"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/action/web"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/featureflag"
	resCsts "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/constants"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/pbstruct"
	clusterRes "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/proto/cluster-resources"
)

// ListEP ...
func (h *Handler) ListEP(
	ctx context.Context, req *clusterRes.ResListReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = resAction.NewResMgr(req.ClusterID, req.ApiVersion, resCsts.EP).List(
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

// GetEP ...
func (h *Handler) GetEP(
	ctx context.Context, req *clusterRes.ResGetReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = resAction.NewResMgr(req.ClusterID, req.ApiVersion, resCsts.EP).Get(
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

// GetEPStatus 特殊接口，当能够获取指定的 endpoints 信息时，返回中 epReady 为 true
func (h *Handler) GetEPStatus(
	ctx context.Context, req *clusterRes.ResGetReq, resp *clusterRes.CommonResp,
) (err error) {
	epReady := true
	if _, err = resAction.NewResMgr(req.ClusterID, req.ApiVersion, resCsts.EP).Get(
		ctx, req.Namespace, req.Name, req.Format, metav1.GetOptions{},
	); err != nil {
		epReady = false
	}

	resp.Data, err = pbstruct.Map2pbStruct(map[string]interface{}{"epReady": epReady})
	return err
}

// CreateEP ...
func (h *Handler) CreateEP(
	ctx context.Context, req *clusterRes.ResCreateReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = resAction.NewResMgr(req.ClusterID, "", resCsts.EP).Create(
		ctx, req.RawData, req.Format, true, metav1.CreateOptions{},
	)
	return err
}

// UpdateEP ...
func (h *Handler) UpdateEP(
	ctx context.Context, req *clusterRes.ResUpdateReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = resAction.NewResMgr(req.ClusterID, "", resCsts.EP).Update(
		ctx, req.Namespace, req.Name, req.RawData, req.Format, metav1.UpdateOptions{},
	)
	return err
}

// DeleteEP ...
func (h *Handler) DeleteEP(
	ctx context.Context, req *clusterRes.ResDeleteReq, _ *clusterRes.CommonResp,
) error {
	return resAction.NewResMgr(req.ClusterID, "", resCsts.EP).Delete(
		ctx, req.Namespace, req.Name, metav1.DeleteOptions{},
	)
}
