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

// Package customresource CRD，自定义资源接口实现
package customresource

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	respUtil "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/action/resp"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/action/web"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/errcode"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/featureflag"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/component/cluster"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/i18n"
	cli "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/client"
	resCsts "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/constants"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/errorx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/pbstruct"
	clusterRes "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/proto/cluster-resources"
)

// Handler xxx
type Handler struct{}

// New xxx
func New() *Handler {
	return &Handler{}
}

// ListCRD xxx
func (h *Handler) ListCRD(
	ctx context.Context, req *clusterRes.ResListReq, resp *clusterRes.CommonResp,
) error {
	ret, err := cli.NewCRDCliByClusterID(ctx, req.ClusterID).List(
		ctx, metav1.ListOptions{LabelSelector: req.LabelSelector},
	)
	if err != nil {
		return err
	}

	respDataBuilder, err := respUtil.NewRespDataBuilder(
		ctx, respUtil.DataBuilderParams{Manifest: ret, Kind: resCsts.CRD, Format: req.Format, Scene: req.Scene},
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
	resp.WebAnnotations, err = web.NewAnnos(
		web.NewFeatureFlag(featureflag.FormCreate, false),
	).ToPbStruct()
	return err
}

// GetCRD xxx
func (h *Handler) GetCRD(
	ctx context.Context, req *clusterRes.ResGetReq, resp *clusterRes.CommonResp,
) error {
	clusterInfo, err := cluster.FromContext(ctx)
	if err != nil {
		return err
	}
	if clusterInfo.Type == cluster.ClusterTypeShared && !cli.IsSharedClusterEnabledCRD(req.Name) {
		return errorx.New(errcode.NoPerm, i18n.GetMsg(ctx, "共享集群中不支持查看 CRD %s 信息"), req.Name)
	}
	resp.Data, err = respUtil.BuildRetrieveAPIResp(
		ctx, respUtil.GetParams{ClusterID: req.ClusterID, ResKind: resCsts.CRD, GroupVersion: "", Namespace: "",
			Name: req.Name, Format: req.Format}, metav1.GetOptions{},
	)
	if err != nil {
		return err
	}
	resp.WebAnnotations, err = web.NewAnnos(
		web.NewFeatureFlag(featureflag.FormUpdate, false),
	).ToPbStruct()
	return err
}
