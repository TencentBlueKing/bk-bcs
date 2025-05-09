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

// Package resource k8s 资源管理相关逻辑
package resource

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/TencentBlueKing/gopkg/collection/set"
	"google.golang.org/protobuf/types/known/structpb"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/action/resp"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/action/trans"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/cluster"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/ctxkey"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/errcode"
	conf "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/config"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/i18n"
	cli "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/client"
	resCsts "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/constants"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/formatter"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/errorx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/slice"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/stringx"
)

// ResMgr k8s 资源管理器，包含命名空间校验，集群操作下发，构建响应内容等功能
type ResMgr struct {
	ClusterID    string
	GroupVersion string
	Kind         string
}

// NewResMgr 创建 ResMgr 并初始化
func NewResMgr(clusterID, groupVersion, kind string) *ResMgr {
	return &ResMgr{
		ClusterID:    clusterID,
		GroupVersion: groupVersion,
		Kind:         kind,
	}
}

/*
目前存在一个问题，大部分资源使用到了 ResMgr，但其 List，Get 方法返回格式固定为了 structpb.Struct，
导致在 handler 层如果要使用数据计算 WebAnnotations（如 ConfigMap, GameDeploy 等），则要么解包 xxx.AsMap()，
要么不使用 ResMgr，直接使用 resp.BuildListApiRespData，这样会导致调用关系比较乱，有些走 ResMgr 有些不走。

考虑的解决方案是：ResMgr 不返回具体的数据，而是一个数据构造方法（比如扩展的 DataBuilder?），根据实际需要的格式获取数据：
比如要获取 Map，则 NewResMgr().List().AsMap()，要获取 structpb.Struct 则 NewResMgr().List().AsPbStruct()
*/

// List 请求某类资源（指定命名空间）下的所有资源列表，按指定 format 格式化后返回
func (m *ResMgr) List(
	ctx context.Context, namespace, format, scene string, opts metav1.ListOptions,
) (*structpb.Struct, error) {
	if err := m.checkAccess(ctx, namespace, nil); err != nil {
		return nil, err
	}
	return resp.BuildListAPIResp(ctx, resp.ListParams{
		ClusterID: m.ClusterID, ResKind: m.Kind, GroupVersion: m.GroupVersion, Namespace: namespace,
		Format: format, Scene: scene,
	}, opts)
}

// Get 请求某个资源详情，按指定 Format 格式化后返回
func (m *ResMgr) Get(
	ctx context.Context, namespace, name, format string, opts metav1.GetOptions,
) (*structpb.Struct, error) {
	if err := m.checkAccess(ctx, namespace, nil); err != nil {
		return nil, err
	}
	return resp.BuildRetrieveAPIResp(ctx, resp.GetParams{
		ClusterID: m.ClusterID, ResKind: m.Kind, GroupVersion: m.GroupVersion, Namespace: namespace,
		Name: name, Format: format,
	}, opts)
}

// Create 创建 k8s 资源，支持以 manifest / formData 格式创建
func (m *ResMgr) Create(
	ctx context.Context, rawData *structpb.Struct, format string, isNSScoped bool, opts metav1.CreateOptions,
) (*structpb.Struct, error) {
	transformer, err := trans.New(ctx, rawData.AsMap(), m.ClusterID, m.Kind, resCsts.CreateAction, format)
	if err != nil {
		return nil, err
	}
	manifest, err := transformer.ToManifest()
	if err != nil {
		return nil, err
	}
	if err = m.checkAccess(ctx, "", manifest); err != nil {
		return nil, err
	}
	// apiVersion 以 manifest 中的为准，不强制要求 preferred
	m.GroupVersion = mapx.GetStr(manifest, "apiVersion")
	return resp.BuildCreateAPIResp(ctx, m.ClusterID, m.Kind, m.GroupVersion, manifest, isNSScoped, opts)
}

// Update 更新 k8s 资源，支持以 manifest / formData 格式更新
func (m *ResMgr) Update(
	ctx context.Context, namespace, name string, rawData *structpb.Struct, format string, opts metav1.UpdateOptions,
) (*structpb.Struct, error) {
	transformer, err := trans.New(ctx, rawData.AsMap(), m.ClusterID, m.Kind, resCsts.UpdateAction, format)
	if err != nil {
		return nil, err
	}
	manifest, err := transformer.ToManifest()
	if err != nil {
		return nil, err
	}
	if err = m.checkAccess(ctx, namespace, manifest); err != nil {
		return nil, err
	}
	// apiVersion 以 manifest 中的为准，不强制要求 preferred
	m.GroupVersion = mapx.GetStr(manifest, "apiVersion")
	return resp.BuildUpdateAPIResp(ctx, m.ClusterID, m.Kind, m.GroupVersion, namespace, name, manifest, opts)
}

// Scale 对某个资源进行扩缩容
func (m *ResMgr) Scale(
	ctx context.Context, namespace, name string, replicas int64, opts metav1.PatchOptions,
) (*structpb.Struct, error) {
	if !isScalable(m.Kind) {
		return nil, errorx.New(errcode.Unsupported, i18n.GetMsg(ctx, "资源类型 %s 不支持扩缩容"), m.Kind)
	}
	patchByte, _ := json.Marshal(
		map[string]interface{}{
			"spec": map[string]interface{}{
				"replicas": replicas,
			},
		},
	)
	return resp.BuildPatchAPIResp(
		ctx, m.ClusterID, m.Kind, m.GroupVersion, namespace, name, types.MergePatchType, patchByte, opts,
	)
}

// Reschedule 对某类资源下属的 Pod 进行重新调度
func (m *ResMgr) Reschedule(ctx context.Context, namespace, name, labelSelector string, podNames []string) error {
	// 1. 父资源类型检查，仅几类可以批量重新调度
	if !isReschedulable(m.Kind) {
		return errorx.New(errcode.Unsupported, i18n.GetMsg(ctx, "资源类型 %s 不支持重新调度下属 Pod"), m.Kind)
	}

	// 2. 获取父资源下属 Pod，与准备重新调度的 Pod 名称列表做比较，确保都属于指定父资源
	podCli := cli.NewPodCliByClusterID(ctx, m.ClusterID)
	podList, err := podCli.List(
		ctx, namespace, m.Kind, name, metav1.ListOptions{LabelSelector: labelSelector},
	)
	if err != nil {
		return err
	}
	ownerPodNames := set.NewStringSet()
	for _, po := range mapx.GetList(podList, "items") {
		ownerPodNames.Add(mapx.GetStr(po.(map[string]interface{}), "metadata.name"))
	}
	// 过滤出属于指定资源的，可以重新调度的 Pod
	allowReschedulePodNames := []string{}
	for _, pn := range podNames {
		if ownerPodNames.Has(pn) {
			allowReschedulePodNames = append(allowReschedulePodNames, pn)
		}
	}

	// 3. 通过批量删除的方式，对下属的 Pod 进行重新调度
	return podCli.BatchDelete(ctx, namespace, allowReschedulePodNames, metav1.DeleteOptions{})
}

// Delete 删除某个 k8s 资源
func (m *ResMgr) Delete(ctx context.Context, namespace, name string, opts metav1.DeleteOptions) error {
	if err := m.checkAccess(ctx, namespace, nil); err != nil {
		return err
	}
	return resp.BuildDeleteAPIResp(ctx, m.ClusterID, m.Kind, m.GroupVersion, namespace, name, opts)
}

// checkAccess 访问权限检查（如共享集群禁用等）
func (m *ResMgr) checkAccess(ctx context.Context, namespace string, manifest map[string]interface{}) error {
	clusterInfo, err := cluster.GetClusterInfo(ctx, m.ClusterID)
	if err != nil {
		return err
	}
	// 独立集群中，不需要做类似校验
	if clusterInfo.Type == cluster.ClusterTypeSingle {
		return nil
	}
	// SC 允许用户查看，PV 返回空，不报错
	if slice.StringInSlice(m.Kind, cluster.SharedClusterBypassClusterScopedKinds) {
		return nil
	}
	// 不允许的资源类型，直接抛出错误
	if !slice.StringInSlice(m.Kind, cluster.SharedClusterEnabledNativeKinds) &&
		!slice.StringInSlice(m.Kind, conf.G.SharedCluster.EnabledCObjKinds) {
		return errorx.New(errcode.NoPerm, i18n.GetMsg(ctx, "该请求资源类型 %s 在共享集群中不可用"), m.Kind)
	}
	// 对命名空间进行检查，确保是属于项目的，命名空间以 manifest 中的为准
	if manifest != nil {
		namespace = mapx.GetStr(manifest, "metadata.namespace")
	}
	if err = cli.CheckIsProjNSinSharedCluster(ctx, m.ClusterID, namespace); err != nil {
		return err
	}
	return nil
}

// Restart 对某个资源进行调度
func (m *ResMgr) Restart(
	ctx context.Context, namespace, name string, generation int64, opts metav1.PatchOptions,
) (*structpb.Struct, error) {
	username := ctxkey.GetUsernameFromCtx(ctx)
	username = stringx.ReplaceIllegalChars(username)
	patchByte := fmt.Sprintf(
		`{"metadata":{"annotations":{"%s":"%s"}},"spec":{"template":{"metadata":{"annotations":{"%s":"%s","%s":"%d"}}}}}`,
		resCsts.UpdaterAnnoKey, username,
		formatter.WorkloadRestartAnnotationKey, metav1.Now().Format(time.RFC3339),
		formatter.WorkloadRestartVersionAnnotationKey, generation)
	pt := types.StrategicMergePatchType
	// 自定义资源的调度(patchType diff)
	if m.GroupVersion == "tkex.tencent.com/v1alpha1" {
		pt = types.MergePatchType
	}
	return resp.BuildPatchAPIResp(
		ctx, m.ClusterID, m.Kind, m.GroupVersion, namespace, name, pt, []byte(patchByte),
		opts,
	)
}

// PauseOrResume 对某个资源进行暂停或恢复
func (m *ResMgr) PauseOrResume(
	ctx context.Context, namespace, name string, paused bool, opts metav1.PatchOptions,
) (*structpb.Struct, error) {
	opts.FieldManager = "kubectl-resume"
	if paused {
		opts.FieldManager = "kubectl-pause"
	}
	patchByte := fmt.Sprintf(`{"spec":{"paused":%v}}`, paused)
	return resp.BuildPatchAPIResp(
		ctx, m.ClusterID, m.Kind, m.GroupVersion, namespace, name, types.StrategicMergePatchType, []byte(patchByte),
		opts,
	)
}
