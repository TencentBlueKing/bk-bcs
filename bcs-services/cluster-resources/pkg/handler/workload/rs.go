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

package workload

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	respUtil "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/action/resp"
	cli "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/client"
	resCsts "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/constants"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/pbstruct"
	clusterRes "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/proto/cluster-resources"
)

// ListRS 获取 ReplicaSet 列表
func (h *Handler) ListRS(
	ctx context.Context, req *clusterRes.ResListReq, resp *clusterRes.CommonResp,
) error {
	ret, err := cli.NewRSCliByClusterID(ctx, req.ClusterID).List(
		ctx, req.Namespace, req.OwnerName, metav1.ListOptions{LabelSelector: req.LabelSelector},
	)
	if err != nil {
		return err
	}

	respDataBuilder, err := respUtil.NewRespDataBuilder(
		ctx, respUtil.DataBuilderParams{Manifest: ret, Kind: resCsts.RS, Format: req.Format, Scene: req.Scene},
	)
	if err != nil {
		return err
	}
	respData, err := respDataBuilder.BuildList()
	if err != nil {
		return err
	}

	resp.Data, err = pbstruct.Map2pbStruct(respData)
	if err != nil {
		return err
	}
	return err
}

// GetDeployHistoryRevision 获取deployment history revision
func (h *Handler) GetDeployHistoryRevision(ctx context.Context, req *clusterRes.GetResHistoryReq,
	resp *clusterRes.CommonListResp) error {

	// 根据deployment name namespace筛选
	ret, err := cli.NewRSCliByClusterID(ctx, req.ClusterID).GetResHistoryRevision(
		ctx, resCsts.Deploy, req.Namespace, req.Name)
	if err != nil {
		return err
	}

	resp.Data, err = pbstruct.MapSlice2ListValue(ret)
	if err != nil {
		return err
	}

	return nil
}

// GetDeployRevisionDiff 获取deployment revision差异信息
func (h *Handler) GetDeployRevisionDiff(ctx context.Context, req *clusterRes.RolloutRevisionReq,
	resp *clusterRes.CommonResp) error {
	ret, err := cli.NewRSCliByClusterID(ctx, req.ClusterID).GetResRevisionDiff(
		ctx, resCsts.Deploy, req.Namespace, req.Name, req.Revision)
	if err != nil {
		return err
	}

	resp.Data, err = pbstruct.Map2pbStruct(ret)
	if err != nil {
		return err
	}
	return nil
}

// RolloutDeployRevision 回滚deployment history revision
func (h *Handler) RolloutDeployRevision(ctx context.Context, req *clusterRes.RolloutRevisionReq,
	resp *clusterRes.CommonResp) error {
	return cli.NewRSCliByClusterID(ctx, req.ClusterID).RolloutResRevision(
		ctx, req.Namespace, req.Name, resCsts.Deploy, req.Revision,
	)
}
