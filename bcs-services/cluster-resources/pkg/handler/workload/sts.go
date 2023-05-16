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

package workload

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	resAction "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/action/resource"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/action/web"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/featureflag"
	resCsts "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/constants"
	clusterRes "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/proto/cluster-resources"
)

// ListSTS 获取 StatefulSet 列表
func (h *Handler) ListSTS(
	ctx context.Context, req *clusterRes.ResListReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = resAction.NewResMgr(req.ClusterID, req.ApiVersion, resCsts.STS).List(
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

// GetSTS 获取单个 StatefulSet
func (h *Handler) GetSTS(
	ctx context.Context, req *clusterRes.ResGetReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = resAction.NewResMgr(req.ClusterID, req.ApiVersion, resCsts.STS).Get(
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

// CreateSTS 创建 StatefulSet
func (h *Handler) CreateSTS(
	ctx context.Context, req *clusterRes.ResCreateReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = resAction.NewResMgr(req.ClusterID, "", resCsts.STS).Create(
		ctx, req.RawData, req.Format, true, metav1.CreateOptions{},
	)
	return err
}

// UpdateSTS 更新 StatefulSet
func (h *Handler) UpdateSTS(
	ctx context.Context, req *clusterRes.ResUpdateReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = resAction.NewResMgr(req.ClusterID, "", resCsts.STS).Update(
		ctx, req.Namespace, req.Name, req.RawData, req.Format, metav1.UpdateOptions{},
	)
	return err
}

// ScaleSTS StatefulSet 扩缩容
func (h *Handler) ScaleSTS(
	ctx context.Context, req *clusterRes.ResScaleReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = resAction.NewResMgr(req.ClusterID, "", resCsts.STS).Scale(
		ctx, req.Namespace, req.Name, req.Replicas, metav1.PatchOptions{},
	)
	return err
}

// RescheduleSTSPo 批量重新调度 StatefulSet 下的 Pod
func (h *Handler) RescheduleSTSPo(
	ctx context.Context, req *clusterRes.ResBatchRescheduleReq, _ *clusterRes.CommonResp,
) (err error) {
	return resAction.NewResMgr(req.ClusterID, "", resCsts.STS).Reschedule(
		ctx, req.Namespace, req.Name, req.LabelSelector, req.PodNames,
	)
}

// DeleteSTS 删除 StatefulSet
func (h *Handler) DeleteSTS(
	ctx context.Context, req *clusterRes.ResDeleteReq, _ *clusterRes.CommonResp,
) error {
	return resAction.NewResMgr(req.ClusterID, "", resCsts.STS).Delete(
		ctx, req.Namespace, req.Name, metav1.DeleteOptions{},
	)
}
