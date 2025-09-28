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

package storage

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	resAction "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/action/resource"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/action/web"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/errcode"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/featureflag"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/component/cluster"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/i18n"
	resCsts "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/constants"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/errorx"
	clusterRes "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/proto/cluster-resources"
)

// ListSC xxx
func (h *Handler) ListSC(
	ctx context.Context, req *clusterRes.ResListReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = resAction.NewResMgr(req.ClusterID, req.ApiVersion, resCsts.SC).List(
		ctx, "", req.Format, req.Scene, metav1.ListOptions{LabelSelector: req.LabelSelector},
	)
	if err != nil {
		return err
	}
	resp.WebAnnotations, err = web.NewAnnos(
		web.NewFeatureFlag(featureflag.FormCreate, false),
	).ToPbStruct()
	return err
}

// GetSC xxx
func (h *Handler) GetSC(
	ctx context.Context, req *clusterRes.ResGetReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = resAction.NewResMgr(req.ClusterID, req.ApiVersion, resCsts.SC).Get(
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

// CreateSC xxx
func (h *Handler) CreateSC(
	ctx context.Context, req *clusterRes.ResCreateReq, resp *clusterRes.CommonResp,
) (err error) {
	clusterInfo, err := cluster.GetClusterInfo(ctx, req.ClusterID)
	if err != nil {
		return err
	}
	if clusterInfo.IsShared {
		return errorx.New(errcode.NoPerm, i18n.GetMsg(ctx, "该请求资源类型 %s 在共享集群中不可用"), resCsts.SC)
	}
	resp.Data, err = resAction.NewResMgr(req.ClusterID, "", resCsts.SC).Create(
		ctx, req.RawData, req.Format, false, metav1.CreateOptions{},
	)
	return err
}

// UpdateSC xxx
func (h *Handler) UpdateSC(
	ctx context.Context, req *clusterRes.ResUpdateReq, resp *clusterRes.CommonResp,
) (err error) {
	clusterInfo, err := cluster.GetClusterInfo(ctx, req.ClusterID)
	if err != nil {
		return err
	}
	if clusterInfo.IsShared {
		return errorx.New(errcode.NoPerm, i18n.GetMsg(ctx, "该请求资源类型 %s 在共享集群中不可用"), resCsts.SC)
	}
	resp.Data, err = resAction.NewResMgr(req.ClusterID, "", resCsts.SC).Update(
		ctx, "", req.Name, req.RawData, req.Format, metav1.UpdateOptions{},
	)
	return err
}

// DeleteSC xxx
func (h *Handler) DeleteSC(
	ctx context.Context, req *clusterRes.ResDeleteReq, _ *clusterRes.CommonResp,
) error {
	clusterInfo, err := cluster.GetClusterInfo(ctx, req.ClusterID)
	if err != nil {
		return err
	}
	if clusterInfo.IsShared {
		return errorx.New(errcode.NoPerm, i18n.GetMsg(ctx, "该请求资源类型 %s 在共享集群中不可用"), resCsts.SC)
	}
	return resAction.NewResMgr(req.ClusterID, "", resCsts.SC).Delete(
		ctx, "", req.Name, metav1.DeleteOptions{},
	)
}
