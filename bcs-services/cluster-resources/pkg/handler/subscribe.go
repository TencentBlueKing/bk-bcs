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

// Package handler subscribe.go 订阅相关逻辑实现
package handler

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	handlerUtil "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/handler/util"
	res "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource"
	cli "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/client"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/formatter"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util"
	clusterRes "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/proto/cluster-resources"
)

// Subscribe 集群资源事件订阅（websocket）
func (crh *ClusterResourcesHandler) Subscribe(
	ctx context.Context, req *clusterRes.SubscribeReq, stream clusterRes.ClusterResources_SubscribeStream,
) error {
	// 参数合法性校验
	if err := validateSubscribeParams(req); err != nil {
		return err
	}

	clusterConf := res.NewClusterConfig(req.ClusterID)
	k8sRes, err := res.GetGroupVersionResource(clusterConf, req.Kind, req.ApiVersion)
	if err != nil {
		return err
	}
	// 获取指定资源类型对应的 watcher
	watcher, err := cli.NewResClient(clusterConf, k8sRes).Watch(
		req.Namespace, metav1.ListOptions{ResourceVersion: req.ResourceVersion},
	)
	if err != nil {
		return err
	}

	for event := range watcher.ResultChan() {
		resp := clusterRes.SubscribeResp{
			Kind:    req.Kind,
			Operate: string(event.Type),
		}

		var raw map[string]interface{}
		switch obj := event.Object.(type) {
		case *unstructured.Unstructured:
			raw = obj.UnstructuredContent()
			resp.Uid = util.GetWithDefault(raw, "metadata.uid", "--").(string)
			resp.Manifest, err = util.Map2pbStruct(raw)
			if err != nil {
				return err
			}
			resp.ManifestExt, err = util.Map2pbStruct(formatter.GetFormatFunc(req.Kind)(raw))
			if err != nil {
				return err
			}
		case *metav1.Status:
			resp.Code = obj.Code
			resp.Message = obj.Message
		}

		if err := stream.Send(&resp); err != nil {
			return err
		}
	}
	return nil
}

var (
	// 支持订阅的 k8s 原生资源类型
	subscribableK8sNaiveKinds = []string{
		res.NS, res.Deploy, res.STS, res.DS, res.CJ, res.Job, res.Po, res.Ing, res.SVC,
		res.EP, res.CM, res.Secret, res.PV, res.PVC, res.SC, res.HPA, res.SA, res.CRD,
	}
	// 支持订阅的 k8s 原生资源类型（集群维度）
	subscribableClusterScopedResKinds = []string{res.NS, res.PV, res.SC, res.CRD}
)

// maybeCobjKind 若不是指定订阅的原生类型，则假定其是自定义资源
func maybeCobjKind(kind string) bool {
	return !util.StringInSlice(kind, subscribableK8sNaiveKinds)
}

// 订阅 API 参数校验
func validateSubscribeParams(req *clusterRes.SubscribeReq) error {
	if maybeCobjKind(req.Kind) {
		// 不支持订阅的原生资源，可以通过要求指定 ApiVersion，CRDName 等的后续检查限制住
		if req.ApiVersion == "" || req.CrdName == "" {
			return fmt.Errorf("当资源类型为自定义对象时，需要指定 ApiVersion & CrdName")
		}
		crdInfo, err := handlerUtil.GetCrdInfo(req.ClusterID, req.CrdName)
		if err != nil {
			return err
		}
		// 优先检查 crdName 查询到的信息与指定的 kind 是否匹配
		if req.Kind != crdInfo["kind"].(string) {
			return fmt.Errorf("CRD %s 的 Kind 与 %s 不匹配", req.CrdName, req.Kind)
		}
		// 自定义资源 & 没有指定命名空间则查询 CRD 检查配置
		if req.Namespace == "" && crdInfo["scope"].(string) == res.NamespacedScope {
			return fmt.Errorf("查询当前自定义资源事件需要指定 Namespace")
		}
	} else if !util.StringInSlice(req.Kind, subscribableClusterScopedResKinds) && req.Namespace == "" {
		return fmt.Errorf("查询当前资源事件需要指定 Namespace")
	}
	return nil
}
