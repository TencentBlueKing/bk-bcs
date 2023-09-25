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
	respAction "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/action/resp"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/action/web"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/featureflag"
	cli "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/client"
	resCsts "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/constants"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/pbstruct"
	clusterRes "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/proto/cluster-resources"
)

// ListDS 获取 DaemonSet 列表
func (h *Handler) ListDS(
	ctx context.Context, req *clusterRes.ResListReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = resAction.NewResMgr(req.ClusterID, req.ApiVersion, resCsts.DS).List(
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

// GetDS 获取单个 DaemonSet
func (h *Handler) GetDS(
	ctx context.Context, req *clusterRes.ResGetReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = resAction.NewResMgr(req.ClusterID, req.ApiVersion, resCsts.DS).Get(
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

// CreateDS 创建 DaemonSet
func (h *Handler) CreateDS(
	ctx context.Context, req *clusterRes.ResCreateReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = resAction.NewResMgr(req.ClusterID, "", resCsts.DS).Create(
		ctx, req.RawData, req.Format, true, metav1.CreateOptions{},
	)
	return err
}

// UpdateDS 更新 DaemonSet
func (h *Handler) UpdateDS(
	ctx context.Context, req *clusterRes.ResUpdateReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = resAction.NewResMgr(req.ClusterID, "", resCsts.DS).Update(
		ctx, req.Namespace, req.Name, req.RawData, req.Format, metav1.UpdateOptions{},
	)
	return err
}

// DeleteDS 删除 DaemonSet
func (h *Handler) DeleteDS(
	ctx context.Context, req *clusterRes.ResDeleteReq, _ *clusterRes.CommonResp,
) error {
	return resAction.NewResMgr(req.ClusterID, "", resCsts.DS).Delete(
		ctx, req.Namespace, req.Name, metav1.DeleteOptions{},
	)
}

// RestartDS 重新调度 DaemonSet
func (h *Handler) RestartDS(
	ctx context.Context, req *clusterRes.ResRestartReq, resp *clusterRes.CommonResp,
) (err error) {
	currentManifest, err := respAction.BuildRetrieveAPIRespData(ctx, respAction.GetParams{
		ClusterID: req.ClusterID, ResKind: resCsts.DS, Namespace: req.Namespace, Name: req.Name,
	}, metav1.GetOptions{})
	if err != nil {
		return err
	}
	revision := mapx.GetInt64(currentManifest, "manifest.metadata.generation")
	// 标记 revision 用来标识应用是否在重启状态
	resp.Data, err = resAction.NewResMgr(req.ClusterID, "", resCsts.DS).Restart(
		ctx, req.Namespace, req.Name, revision+1, metav1.PatchOptions{FieldManager: "kubectl-rollout"},
	)
	return err
}

// GetDSHistoryRevision 获取 DaemonSet history revision
func (h *Handler) GetDSHistoryRevision(ctx context.Context, req *clusterRes.GetResHistoryReq,
	resp *clusterRes.CommonListResp) error {
	ret, err := cli.NewRSCliByClusterID(ctx, req.ClusterID).GetResHistoryRevision(
		ctx, resCsts.DS, req.Namespace, req.Name)

	if err != nil {
		return err
	}
	resp.Data, err = pbstruct.MapSlice2ListValue(ret)
	return err
}

// GetDSRevisionDiff 获取 DaemonSet revision差异信息
func (h *Handler) GetDSRevisionDiff(ctx context.Context, req *clusterRes.RolloutRevisionReq,
	resp *clusterRes.CommonResp) error {
	ret, err := cli.NewRSCliByClusterID(ctx, req.ClusterID).GetResRevisionDiff(
		ctx, resCsts.DS, req.Namespace, req.Name, req.Revision)
	if err != nil {
		return err
	}

	resp.Data, err = pbstruct.Map2pbStruct(ret)
	if err != nil {
		return err
	}
	return nil
}

// RolloutDSRevision 回滚DaemonSet history revision
func (h *Handler) RolloutDSRevision(ctx context.Context, req *clusterRes.RolloutRevisionReq,
	_ *clusterRes.CommonResp) error {
	return cli.NewRSCliByClusterID(ctx, req.ClusterID).RolloutResRevision(
		ctx, req.Namespace, req.Name, resCsts.DS, req.Revision)
}
