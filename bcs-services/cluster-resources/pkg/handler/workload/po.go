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
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/action/util/perm"
	respUtil "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/action/util/resp"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/errcode"
	res "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource"
	cli "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/client"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/errorx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/pbstruct"
	clusterRes "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/proto/cluster-resources"
)

// ListPo 获取 Pod 列表
func (h *Handler) ListPo(
	ctx context.Context, req *clusterRes.PodResListReq, resp *clusterRes.CommonResp,
) error {
	// 获取指定命名空间下的所有符合条件的 Pod
	ret, err := cli.NewPodCliByClusterID(ctx, req.ClusterID).List(
		ctx, req.Namespace, req.OwnerKind, req.OwnerName, metav1.ListOptions{LabelSelector: req.LabelSelector},
	)
	if err != nil {
		return err
	}

	respDataBuilder, err := respUtil.NewRespDataBuilder(ret, res.CRD, req.Format)
	if err != nil {
		return err
	}
	respData, err := respDataBuilder.BuildList()
	if err != nil {
		return err
	}

	resp.Data, err = pbstruct.Map2pbStruct(respData)
	return err
}

// GetPo 获取单个 Pod
func (h *Handler) GetPo(
	ctx context.Context, req *clusterRes.ResGetReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = resAction.NewResMgr(req.ClusterID, req.ApiVersion, res.Po).Get(
		ctx, req.Namespace, req.Name, req.Format, metav1.GetOptions{},
	)
	return err
}

// CreatePo 创建 Pod
func (h *Handler) CreatePo(
	ctx context.Context, req *clusterRes.ResCreateReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = resAction.NewResMgr(req.ClusterID, "", res.Po).Create(
		ctx, req.RawData, req.Format, true, metav1.CreateOptions{},
	)
	return err
}

// UpdatePo 更新 Pod
func (h *Handler) UpdatePo(
	ctx context.Context, req *clusterRes.ResUpdateReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = resAction.NewResMgr(req.ClusterID, "", res.Po).Update(
		ctx, req.Namespace, req.Name, req.RawData, req.Format, metav1.UpdateOptions{},
	)
	return err
}

// DeletePo 删除 Pod
func (h *Handler) DeletePo(
	ctx context.Context, req *clusterRes.ResDeleteReq, _ *clusterRes.CommonResp,
) error {
	return resAction.NewResMgr(req.ClusterID, "", res.Po).Delete(
		ctx, req.Namespace, req.Name, metav1.DeleteOptions{},
	)
}

// ListPoPVC 获取 Pod PVC 列表
func (h *Handler) ListPoPVC(
	ctx context.Context, req *clusterRes.ResGetReq, resp *clusterRes.CommonResp,
) (err error) {
	if err := perm.CheckNSAccess(ctx, req.ClusterID, req.Namespace); err != nil {
		return err
	}
	resp.Data, err = respUtil.BuildListPodRelatedResResp(
		ctx, req.ClusterID, req.Namespace, req.Name, req.Format, res.PVC,
	)
	return err
}

// ListPoCM 获取 Pod ConfigMap 列表
func (h *Handler) ListPoCM(
	ctx context.Context, req *clusterRes.ResGetReq, resp *clusterRes.CommonResp,
) (err error) {
	if err := perm.CheckNSAccess(ctx, req.ClusterID, req.Namespace); err != nil {
		return err
	}
	resp.Data, err = respUtil.BuildListPodRelatedResResp(
		ctx, req.ClusterID, req.Namespace, req.Name, req.Format, res.CM,
	)
	return err
}

// ListPoSecret 获取 Pod Secret 列表
func (h *Handler) ListPoSecret(
	ctx context.Context, req *clusterRes.ResGetReq, resp *clusterRes.CommonResp,
) (err error) {
	if err := perm.CheckNSAccess(ctx, req.ClusterID, req.Namespace); err != nil {
		return err
	}
	resp.Data, err = respUtil.BuildListPodRelatedResResp(
		ctx, req.ClusterID, req.Namespace, req.Name, req.Format, res.Secret,
	)
	return err
}

// ReschedulePo 重新调度 Pod
func (h *Handler) ReschedulePo(
	ctx context.Context, req *clusterRes.ResUpdateReq, _ *clusterRes.CommonResp,
) (err error) {
	if err := perm.CheckNSAccess(ctx, req.ClusterID, req.Namespace); err != nil {
		return err
	}

	podManifest, err := cli.NewPodCliByClusterID(ctx, req.ClusterID).GetManifest(ctx, req.Namespace, req.Name)
	if err != nil {
		return err
	}

	// 检查 Pod 配置，必须有父级资源且不为 Job 才可以重新调度
	ownerReferences, err := mapx.GetItems(podManifest, "metadata.ownerReferences")
	if err != nil {
		return errorx.New(errcode.Unsupported, "Pod %s/%s 不存在父级资源，不允许重新调度", req.Namespace, req.Name)
	}
	// 检查确保父级资源不为 Job
	for _, ref := range ownerReferences.([]interface{}) {
		if ref.(map[string]interface{})["kind"].(string) == res.Job {
			return errorx.New(errcode.Unsupported, "Pod %s/%s 父级资源存在 Job，不允许重新调度", req.Namespace, req.Name)
		}
	}

	// 重新调度的原理是直接删除 Pod，利用父级资源重新拉起服务
	return respUtil.BuildDeleteAPIResp(
		ctx, req.ClusterID, res.Po, "", req.Namespace, req.Name, metav1.DeleteOptions{},
	)
}
