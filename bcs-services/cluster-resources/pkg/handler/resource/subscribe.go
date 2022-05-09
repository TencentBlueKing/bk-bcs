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

package resource

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/watch"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/action/util/perm"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/cluster"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/ctxkey"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/errcode"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/project"
	res "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource"
	cli "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/client"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/formatter"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/errorx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/pbstruct"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/slice"
	clusterRes "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/proto/cluster-resources"
)

// Subscribe 集群资源事件订阅（websocket）
func (h *Handler) Subscribe(
	ctx context.Context, req *clusterRes.SubscribeReq, stream clusterRes.Resource_SubscribeStream,
) (err error) {
	// 注入项目，集群信息
	ctx, err = injectProjClusterInfo(ctx, req)
	if err != nil {
		return err
	}

	// 接口调用合法性校验
	if err = perm.CheckSubscribable(ctx, req); err != nil {
		return err
	}

	// 参数合法性校验
	if err = validateSubscribeParams(ctx, req); err != nil {
		return err
	}

	// 获取指定资源对应的 Watcher
	watcher, err := genResWatcher(ctx, req)
	if err != nil {
		return err
	}

	for event := range watcher.ResultChan() {
		resp := clusterRes.SubscribeResp{
			Kind: req.Kind,
			Type: string(event.Type),
		}

		var raw map[string]interface{}
		switch obj := event.Object.(type) {
		case *unstructured.Unstructured:
			raw = obj.UnstructuredContent()
			resp.Uid = mapx.GetStr(raw, "metadata.uid")
			resp.Manifest, err = pbstruct.Map2pbStruct(raw)
			if err != nil {
				return err
			}
			resp.ManifestExt, err = pbstruct.Map2pbStruct(formatter.GetFormatFunc(req.Kind)(raw))
			if err != nil {
				return err
			}
		case *metav1.Status:
			resp.Code = obj.Code
			resp.Message = obj.Message
		}

		if err = stream.Send(&resp); err != nil {
			return err
		}
	}
	return nil
}

var (
	// 支持订阅的 k8s 原生资源类型
	subscribableNativeKinds = []string{
		res.NS, res.Deploy, res.STS, res.DS, res.CJ, res.Job, res.Po, res.Ing, res.SVC,
		res.EP, res.CM, res.Secret, res.PV, res.PVC, res.SC, res.HPA, res.SA, res.CRD,
	}
	// 支持订阅的 k8s 原生资源类型（集群维度）
	subscribableClusterScopedKinds = []string{res.NS, res.PV, res.SC, res.CRD}
)

// 若不是指定订阅的原生类型，则假定其是自定义资源
func maybeCobjKind(kind string) bool {
	return !slice.StringInSlice(kind, subscribableNativeKinds)
}

// 在 Context 中注入 Project，Cluster 信息
func injectProjClusterInfo(ctx context.Context, req *clusterRes.SubscribeReq) (context.Context, error) {
	projInfo, err := project.GetProjectInfo(ctx, req.ProjectID)
	if err != nil {
		return nil, errorx.New(errcode.General, "获取项目 %s 信息失败：%v", req.ProjectID, err)
	}
	clusterInfo, err := cluster.GetClusterInfo(ctx, req.ClusterID)
	if err != nil {
		return nil, errorx.New(errcode.General, "获取集群 %s 信息失败：%v", req.ClusterID, err)
	}
	// 若集群类型非共享集群，则需确认集群的项目 ID 与请求参数中的一致
	if !slice.StringInSlice(clusterInfo.Type, cluster.SharedClusterTypes) && clusterInfo.ProjID != projInfo.ID {
		return nil, errorx.New(errcode.ValidateErr, "集群 %s 不属于指定项目!", req.ClusterID)
	}
	ctx = context.WithValue(ctx, ctxkey.ProjKey, projInfo)
	ctx = context.WithValue(ctx, ctxkey.ClusterKey, clusterInfo)
	return ctx, nil
}

// 订阅 API 参数校验
func validateSubscribeParams(ctx context.Context, req *clusterRes.SubscribeReq) error {
	if maybeCobjKind(req.Kind) {
		// 不支持订阅的原生资源，可以通过要求指定 ApiVersion，CRDName 等的后续检查限制住
		if req.ApiVersion == "" || req.CRDName == "" {
			return errorx.New(errcode.ValidateErr, "当资源类型为自定义对象时，需要指定 ApiVersion & CRDName")
		}
		crdInfo, err := cli.GetCRDInfo(ctx, req.ClusterID, req.CRDName)
		if err != nil {
			return err
		}
		// 优先检查 crdName 查询到的信息与指定的 kind 是否匹配
		if req.Kind != crdInfo["kind"].(string) {
			return errorx.New(errcode.ValidateErr, "CRD %s 的 Kind 与 %s 不匹配", req.CRDName, req.Kind)
		}
		// 自定义资源 & 没有指定命名空间则查询 CRD 检查配置
		if req.Namespace == "" && crdInfo["scope"].(string) == res.NamespacedScope {
			return errorx.New(errcode.ValidateErr, "查询当前自定义资源事件需要指定 Namespace")
		}
	} else if !slice.StringInSlice(req.Kind, subscribableClusterScopedKinds) && req.Namespace == "" {
		return errorx.New(errcode.ValidateErr, "查询当前资源事件需要指定 Namespace")
	}
	return nil
}

// 获取某类资源对应的 watcher
func genResWatcher(ctx context.Context, req *clusterRes.SubscribeReq) (watch.Interface, error) {
	clusterConf := res.NewClusterConfig(req.ClusterID)
	opts := metav1.ListOptions{ResourceVersion: req.ResourceVersion}
	clusterInfo, err := cluster.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	// 命名空间，CRD watcher 特殊处理
	if req.Kind == res.NS {
		projInfo, fetchProjErr := project.FromContext(ctx)
		if fetchProjErr != nil {
			return nil, err
		}
		return cli.NewNSClient(ctx, clusterConf).Watch(ctx, projInfo.Code, clusterInfo.Type, opts)
	}
	if req.Kind == res.CRD {
		return cli.NewCRDClient(ctx, clusterConf).Watch(ctx, clusterInfo.Type, opts)
	}
	k8sRes, err := res.GetGroupVersionResource(ctx, clusterConf, req.Kind, req.ApiVersion)
	if err != nil {
		return nil, err
	}
	return cli.NewResClient(clusterConf, k8sRes).Watch(ctx, req.Namespace, opts)
}
