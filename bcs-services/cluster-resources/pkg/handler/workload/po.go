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
	"sort"

	"github.com/TencentBlueKing/gopkg/collection/set"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/action/perm"
	resAction "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/action/resource"
	respUtil "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/action/resp"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/action/web"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/errcode"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/featureflag"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/i18n"
	cli "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/client"
	resCsts "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/constants"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/formatter"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/errorx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/pbstruct"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/slice"
	clusterRes "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/proto/cluster-resources"
)

// ListPo 获取 Pod 列表
func (h *Handler) ListPo(
	ctx context.Context, req *clusterRes.ResListReq, resp *clusterRes.CommonResp,
) error {
	// 获取指定命名空间下的所有符合条件的 Pod
	ret, err := cli.NewPodCliByClusterID(ctx, req.ClusterID).List(
		ctx, req.Namespace, req.OwnerKind, req.OwnerName, metav1.ListOptions{LabelSelector: req.LabelSelector},
	)
	if err != nil {
		return err
	}

	respDataBuilder, err := respUtil.NewRespDataBuilder(
		ctx, respUtil.DataBuilderParams{Manifest: ret, Kind: resCsts.Po, Format: req.Format, Scene: req.Scene},
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
		web.NewFeatureFlag(featureflag.FormCreate, true),
	).ToPbStruct()
	return err
}

// GetPo 获取单个 Pod
func (h *Handler) GetPo(
	ctx context.Context, req *clusterRes.ResGetReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = resAction.NewResMgr(req.ClusterID, req.ApiVersion, resCsts.Po).Get(
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

// CreatePo 创建 Pod
func (h *Handler) CreatePo(
	ctx context.Context, req *clusterRes.ResCreateReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = resAction.NewResMgr(req.ClusterID, "", resCsts.Po).Create(
		ctx, req.RawData, req.Format, true, metav1.CreateOptions{},
	)
	return err
}

// UpdatePo 更新 Pod
func (h *Handler) UpdatePo(
	ctx context.Context, req *clusterRes.ResUpdateReq, resp *clusterRes.CommonResp,
) (err error) {
	resp.Data, err = resAction.NewResMgr(req.ClusterID, "", resCsts.Po).Update(
		ctx, req.Namespace, req.Name, req.RawData, req.Format, metav1.UpdateOptions{},
	)
	return err
}

// DeletePo 删除 Pod
func (h *Handler) DeletePo(
	ctx context.Context, req *clusterRes.ResDeleteReq, _ *clusterRes.CommonResp,
) error {
	return resAction.NewResMgr(req.ClusterID, "", resCsts.Po).Delete(
		ctx, req.Namespace, req.Name, metav1.DeleteOptions{},
	)
}

// ListPoPVC 获取 Pod PVC 列表
func (h *Handler) ListPoPVC(
	ctx context.Context, req *clusterRes.ResGetReq, resp *clusterRes.CommonResp,
) (err error) {
	if err = perm.CheckNSAccess(ctx, req.ClusterID, req.Namespace); err != nil {
		return err
	}
	resp.Data, err = respUtil.BuildListPodRelatedResResp(
		ctx, req.ClusterID, req.Namespace, req.Name, req.Format, resCsts.PVC,
	)
	return err
}

// ListPoCM 获取 Pod ConfigMap 列表
func (h *Handler) ListPoCM(
	ctx context.Context, req *clusterRes.ResGetReq, resp *clusterRes.CommonResp,
) (err error) {
	if err = perm.CheckNSAccess(ctx, req.ClusterID, req.Namespace); err != nil {
		return err
	}
	resp.Data, err = respUtil.BuildListPodRelatedResResp(
		ctx, req.ClusterID, req.Namespace, req.Name, req.Format, resCsts.CM,
	)
	return err
}

// ListPoSecret 获取 Pod Secret 列表
func (h *Handler) ListPoSecret(
	ctx context.Context, req *clusterRes.ResGetReq, resp *clusterRes.CommonResp,
) (err error) {
	if err = perm.CheckNSAccess(ctx, req.ClusterID, req.Namespace); err != nil {
		return err
	}
	resp.Data, err = respUtil.BuildListPodRelatedResResp(
		ctx, req.ClusterID, req.Namespace, req.Name, req.Format, resCsts.Secret,
	)
	return err
}

// ReschedulePo 重新调度 Pod
func (h *Handler) ReschedulePo(
	ctx context.Context, req *clusterRes.ResUpdateReq, _ *clusterRes.CommonResp,
) (err error) {
	if err = perm.CheckNSAccess(ctx, req.ClusterID, req.Namespace); err != nil {
		return err
	}

	podManifest, err := cli.NewPodCliByClusterID(ctx, req.ClusterID).GetManifest(ctx, req.Namespace, req.Name)
	if err != nil {
		return err
	}

	// 检查 Pod 配置，必须有父级资源且不为 Job 才可以重新调度
	ownerReferences, err := mapx.GetItems(podManifest, "metadata.ownerReferences")
	if err != nil {
		return errorx.New(
			errcode.Unsupported,
			i18n.GetMsg(ctx, "Pod %s/%s 不存在父级资源，不允许重新调度"),
			req.Namespace, req.Name,
		)
	}
	// 检查确保父级资源不为 Job
	for _, ref := range ownerReferences.([]interface{}) {
		if ref.(map[string]interface{})["kind"].(string) == resCsts.Job {
			return errorx.New(
				errcode.Unsupported,
				i18n.GetMsg(ctx, "Pod %s/%s 父级资源存在 Job，不允许重新调度"),
				req.Namespace, req.Name,
			)
		}
	}

	// 重新调度的原理是直接删除 Pod，利用父级资源重新拉起服务
	return respUtil.BuildDeleteAPIResp(
		ctx, req.ClusterID, resCsts.Po, "", req.Namespace, req.Name, metav1.DeleteOptions{},
	)
}

// ListPoByNode 获取指定集群运行于某 Node 上的 Pod
// 注意，该接口权限为 "集群查看" 权限，而非命名空间域资源查看权限，返回的数据也不是 manifest，仅包含列表展示需要的数据
// 返回仅 pod 部分数据，不需要集群管理权限，很多用户只有集群查看权限，没有集群管理权限
func (h *Handler) ListPoByNode(
	ctx context.Context, req *clusterRes.ListPoByNodeReq, resp *clusterRes.CommonListResp,
) error {
	podCli := cli.NewPodCliByClusterID(ctx, req.ClusterID)
	ret, err := podCli.ListAllPods(
		ctx, req.ProjectID, req.ClusterID, metav1.ListOptions{FieldSelector: "spec.nodeName=" + req.NodeName},
	)
	if err != nil {
		return err
	}

	podList := []map[string]interface{}{}
	namespaces := set.NewStringSet()

	for _, po := range mapx.GetList(ret, "items") {
		p := po.(map[string]interface{})
		ns := mapx.GetStr(p, "metadata.namespace")
		pod := formatter.FormatPo(p)
		pod["namespace"] = ns
		pod["uid"] = mapx.GetStr(p, "metadata.uid")
		pod["name"] = mapx.GetStr(p, "metadata.name")
		pod["hostIP"] = mapx.Get(p, "status.hostIP", "N/A").(string)
		pod["podIP"] = mapx.Get(p, "status.podIP", "N/A").(string)
		pod["node"] = mapx.Get(p, "spec.nodeName", "N/A").(string)
		podList = append(podList, pod)
		namespaces.Add(ns)
	}

	if resp.Data, err = pbstruct.MapSlice2ListValue(podList); err != nil {
		return err
	}
	// 提供列表行权限的 WebAnnotations，用于前端跳转禁用
	resp.WebAnnotations, err = web.GenNodePodListWebAnnos(
		ctx, podList, req.ProjectID, req.ClusterID, namespaces.ToSlice(),
	)
	return err
}

// ListPoLabelsByNode 获取指定集群运行于某 Node 上的 Pod labels
// 注意，该接口权限为 "集群查看" 权限，而非命名空间域资源查看权限，返回的数据也不是 manifest，仅包含列表展示需要的数据
// 返回仅 pod 部分数据，不需要集群管理权限，很多用户只有集群查看权限，没有集群管理权限
func (h *Handler) ListPoLabelsByNode(
	ctx context.Context, req *clusterRes.ListPoByNodeReq, resp *clusterRes.CommonResp,
) error {
	podCli := cli.NewPodCliByClusterID(ctx, req.ClusterID)
	ret, err := podCli.ListAllPods(
		ctx, req.ProjectID, req.ClusterID, metav1.ListOptions{FieldSelector: "spec.nodeName=" + req.NodeName},
	)
	if err != nil {
		return err
	}

	labels := parseResourceLabels(mapx.GetList(ret, "items"))

	labels = slice.RemoveDuplicateValues(labels)
	sort.Strings(labels)

	if resp.Data, err = pbstruct.Map2pbStruct(map[string]interface{}{"values": labels}); err != nil {
		return err
	}
	return nil
}

// 根据集群命名空间查询资源标签
func parseResourceLabels(items []interface{}) []string {
	labels := make([]string, 0)
	for _, item := range items {
		v := item.(map[string]interface{})
		for k := range mapx.GetMap(v, "metadata.labels") {
			labels = append(labels, k)
		}
	}
	return labels
}
