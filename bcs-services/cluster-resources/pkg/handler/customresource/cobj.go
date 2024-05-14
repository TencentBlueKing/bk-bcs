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

package customresource

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/action/perm"
	resAction "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/action/resource"
	respUtil "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/action/resp"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/action/trans"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/action/web"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/errcode"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/i18n"
	cli "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/client"
	resCsts "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/constants"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/errorx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/pbstruct"
	clusterRes "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/proto/cluster-resources"
)

// ListCObj 获取自定义资源列表
func (h *Handler) ListCObj(
	ctx context.Context, req *clusterRes.CObjListReq, resp *clusterRes.CommonResp,
) error {
	crdInfo, err := cli.GetCRDInfo(ctx, req.ClusterID, req.CRDName)
	if err != nil {
		return err
	}
	if err = perm.CheckCObjAccess(ctx, req.ClusterID, req.CRDName, req.Namespace); err != nil {
		return err
	}
	kind, apiVersion := crdInfo["kind"].(string), crdInfo["apiVersion"].(string)
	respData, err := respUtil.BuildListAPIRespData(
		ctx, respUtil.ListParams{
			ClusterID: req.ClusterID, ResKind: kind, GroupVersion: apiVersion, Namespace: req.Namespace,
			Format: req.Format, Scene: req.Scene,
		}, metav1.ListOptions{},
	)
	if err != nil {
		return err
	}
	if resp.Data, err = pbstruct.Map2pbStruct(respData); err != nil {
		return err
	}

	resp.WebAnnotations, err = web.GenListCObjWebAnnos(ctx, respData, crdInfo, req.Format)
	return err
}

// GetCObj 获取某个自定义资源
func (h *Handler) GetCObj(
	ctx context.Context, req *clusterRes.CObjGetReq, resp *clusterRes.CommonResp,
) error {
	crdInfo, err := cli.GetCRDInfo(ctx, req.ClusterID, req.CRDName)
	if err != nil {
		return err
	}
	if err = validateNSParam(ctx, crdInfo, req.Namespace); err != nil {
		return err
	}
	if err = perm.CheckCObjAccess(ctx, req.ClusterID, req.CRDName, req.Namespace); err != nil {
		return err
	}
	kind, apiVersion := crdInfo["kind"].(string), crdInfo["apiVersion"].(string)
	respData, err := respUtil.BuildRetrieveAPIRespData(
		ctx, respUtil.GetParams{
			ClusterID: req.ClusterID, ResKind: kind, GroupVersion: apiVersion, Namespace: req.Namespace,
			Name: req.CobjName, Format: req.Format,
		}, metav1.GetOptions{},
	)
	if err != nil {
		return err
	}
	if resp.Data, err = pbstruct.Map2pbStruct(respData); err != nil {
		return err
	}

	resp.WebAnnotations, err = web.GenRetrieveCObjWebAnnos(ctx, respData, crdInfo, req.Format)
	return err
}

// CreateCObj 创建自定义资源
func (h *Handler) CreateCObj(
	ctx context.Context, req *clusterRes.CObjCreateReq, resp *clusterRes.CommonResp,
) (err error) {
	crdInfo, err := cli.GetCRDInfo(ctx, req.ClusterID, req.CRDName)
	if err != nil {
		return err
	}
	kind, apiVersion := crdInfo["kind"].(string), crdInfo["apiVersion"].(string)

	transformer, err := trans.New(ctx, req.RawData.AsMap(), req.ClusterID, kind, resCsts.CreateAction, req.Format)
	if err != nil {
		return err
	}
	manifest, err := transformer.ToManifest()
	if err != nil {
		return err
	}
	namespace := mapx.GetStr(manifest, "metadata.namespace")

	if err = validateNSParam(ctx, crdInfo, namespace); err != nil {
		return err
	}
	if err = perm.CheckCObjAccess(ctx, req.ClusterID, req.CRDName, namespace); err != nil {
		return err
	}
	// 经过命名空间检查后，若不需要指定命名空间，则认为是集群维度的
	resp.Data, err = respUtil.BuildCreateAPIResp(
		ctx, req.ClusterID, kind, apiVersion, manifest, namespace != "", metav1.CreateOptions{},
	)
	return err
}

// UpdateCObj 更新某个自定义资源（replace）
func (h *Handler) UpdateCObj(
	ctx context.Context, req *clusterRes.CObjUpdateReq, resp *clusterRes.CommonResp,
) error {
	crdInfo, err := cli.GetCRDInfo(ctx, req.ClusterID, req.CRDName)
	if err != nil {
		return err
	}
	kind, apiVersion := crdInfo["kind"].(string), crdInfo["apiVersion"].(string)

	transformer, err := trans.New(ctx, req.RawData.AsMap(), req.ClusterID, kind, resCsts.UpdateAction, req.Format)
	if err != nil {
		return err
	}
	manifest, err := transformer.ToManifest()
	if err != nil {
		return err
	}
	if err = validateNSParam(ctx, crdInfo, req.Namespace); err != nil {
		return err
	}
	if err = perm.CheckCObjAccess(ctx, req.ClusterID, req.CRDName, req.Namespace); err != nil {
		return err
	}
	resp.Data, err = respUtil.BuildUpdateCObjAPIResp(
		ctx, req.ClusterID, kind, apiVersion, req.Namespace, req.CobjName, manifest, metav1.UpdateOptions{},
	)
	return err
}

// ScaleCObj 自定义资源扩缩容（仅 GameDeployment, GameStatefulSet 可用）
func (h *Handler) ScaleCObj(
	ctx context.Context, req *clusterRes.CObjScaleReq, resp *clusterRes.CommonResp,
) error {
	crdInfo, err := cli.GetCRDInfo(ctx, req.ClusterID, req.CRDName)
	if err != nil {
		return err
	}

	resp.Data, err = resAction.NewResMgr(req.ClusterID, crdInfo["apiVersion"].(string), crdInfo["kind"].(string)).Scale(
		ctx, req.Namespace, req.CobjName, req.Replicas, metav1.PatchOptions{},
	)
	return err
}

// DeleteCObj 删除某个自定义资源
func (h *Handler) DeleteCObj(
	ctx context.Context, req *clusterRes.CObjDeleteReq, _ *clusterRes.CommonResp,
) error {
	crdInfo, err := cli.GetCRDInfo(ctx, req.ClusterID, req.CRDName)
	if err != nil {
		return err
	}
	if err = validateNSParam(ctx, crdInfo, req.Namespace); err != nil {
		return err
	}
	if err = perm.CheckCObjAccess(ctx, req.ClusterID, req.CRDName, req.Namespace); err != nil {
		return err
	}
	kind, apiVersion := crdInfo["kind"].(string), crdInfo["apiVersion"].(string)
	return respUtil.BuildDeleteAPIResp(
		ctx, req.ClusterID, kind, apiVersion, req.Namespace, req.CobjName, metav1.DeleteOptions{},
	)
}

// RescheduleCObjPo 重新调度自定义资源下属的 Pod（仅 GameDeployment, GameStatefulSet 可用）
func (h *Handler) RescheduleCObjPo(
	ctx context.Context, req *clusterRes.CObjBatchRescheduleReq, _ *clusterRes.CommonResp,
) error {
	crdInfo, err := cli.GetCRDInfo(ctx, req.ClusterID, req.CRDName)
	if err != nil {
		return err
	}

	return resAction.NewResMgr(req.ClusterID, crdInfo["apiVersion"].(string), crdInfo["kind"].(string)).Reschedule(
		ctx, req.Namespace, req.CobjName, req.LabelSelector, req.PodNames,
	)
}

// validateNSParam 校验 CObj 相关请求中命名空间参数，若 CRD 中定义为集群维度，则不需要，否则需要指定命名空间
func validateNSParam(ctx context.Context, crdInfo map[string]interface{}, namespace string) error {
	if namespace == "" && crdInfo["scope"].(string) == resCsts.NamespacedScope {
		return errorx.New(errcode.ValidateErr, i18n.GetMsg(ctx, "查看/操作自定义资源 %s 需要指定命名空间"), crdInfo["name"])
	}
	return nil
}

// GetCObjHistoryRevision 获取自定义资源 history revision
func (h *Handler) GetCObjHistoryRevision(ctx context.Context, req *clusterRes.CObjHistoryReq,
	resp *clusterRes.CommonListResp) error {

	crdInfo, err := cli.GetCRDInfo(ctx, req.ClusterID, req.CRDName)
	if err != nil {
		return err
	}
	if err = validateNSParam(ctx, crdInfo, req.Namespace); err != nil {
		return err
	}
	if err = perm.CheckCObjAccess(ctx, req.ClusterID, req.CRDName, req.Namespace); err != nil {
		return err
	}
	kind := crdInfo["kind"].(string)

	ret, err := cli.NewCRDCliByClusterID(ctx, req.ClusterID).HistoryRevision(ctx, kind,
		req.Namespace, req.CobjName)
	if err != nil {
		return err
	}

	resp.Data, err = pbstruct.MapSlice2ListValue(ret)
	if err != nil {
		return err
	}

	return nil
}

// GetCObjRevisionDiff 获取自定义资源 revision差异信息
func (h *Handler) GetCObjRevisionDiff(ctx context.Context, req *clusterRes.CObjRolloutReq,
	resp *clusterRes.CommonResp) error {

	crdInfo, err := cli.GetCRDInfo(ctx, req.ClusterID, req.CRDName)
	if err != nil {
		return err
	}
	if err = validateNSParam(ctx, crdInfo, req.Namespace); err != nil {
		return err
	}
	if err = perm.CheckCObjAccess(ctx, req.ClusterID, req.CRDName, req.Namespace); err != nil {
		return err
	}
	kind := crdInfo["kind"].(string)

	ret, err := cli.NewCRDCliByClusterID(ctx, req.ClusterID).GetResRevisionDiff(ctx, kind,
		req.Namespace, req.CobjName, req.Revision)
	if err != nil {
		return err
	}

	resp.Data, err = pbstruct.Map2pbStruct(ret)
	if err != nil {
		return err
	}
	return nil
}

// RestartCObj 重新调度单个自定义资源
func (h *Handler) RestartCObj(ctx context.Context, req *clusterRes.CObjRestartReq, resp *clusterRes.CommonResp) error {
	crdInfo, err := cli.GetCRDInfo(ctx, req.ClusterID, req.CRDName)
	if err != nil {
		return err
	}
	if err = validateNSParam(ctx, crdInfo, req.Namespace); err != nil {
		return err
	}
	if err = perm.CheckCObjAccess(ctx, req.ClusterID, req.CRDName, req.Namespace); err != nil {
		return err
	}
	kind, apiVersion := crdInfo["kind"].(string), crdInfo["apiVersion"].(string)

	currentManifest, err := respUtil.BuildRetrieveAPIRespData(ctx, respUtil.GetParams{
		ClusterID: req.ClusterID, ResKind: kind, Namespace: req.Namespace, Name: req.CobjName,
	}, metav1.GetOptions{})
	if err != nil {
		return err
	}
	revision := mapx.GetInt64(currentManifest, "manifest.metadata.generation")
	// 标记 revision 用来标识应用是否在重启状态
	resp.Data, err = resAction.NewResMgr(req.ClusterID, apiVersion, kind).Restart(
		ctx, req.Namespace, req.CobjName, revision+1, metav1.PatchOptions{FieldManager: "kubectl-rollout"},
	)
	return err
}

// RolloutCObj 回滚自定义资源
func (h *Handler) RolloutCObj(ctx context.Context, req *clusterRes.CObjRolloutReq, _ *clusterRes.CommonResp) error {
	crdInfo, err := cli.GetCRDInfo(ctx, req.ClusterID, req.CRDName)
	if err != nil {
		return err
	}
	if err = validateNSParam(ctx, crdInfo, req.Namespace); err != nil {
		return err
	}
	if err = perm.CheckCObjAccess(ctx, req.ClusterID, req.CRDName, req.Namespace); err != nil {
		return err
	}
	kind := crdInfo["kind"].(string)

	return cli.NewCRDCliByClusterID(ctx, req.ClusterID).RolloutRevision(ctx, req.Namespace,
		req.CobjName, kind, req.Revision)
}
