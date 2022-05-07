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

// Package resource k8s 资源管理相关逻辑
package resource

import (
	"context"

	"google.golang.org/protobuf/types/known/structpb"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/action/util/resp"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/action/util/trans"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/cluster"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/errcode"
	cli "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/client"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/errorx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/slice"
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

// List 请求某类资源（指定命名空间）下的所有资源列表，按指定 format 格式化后返回
func (m *ResMgr) List(ctx context.Context, namespace, format string, opts metav1.ListOptions) (*structpb.Struct, error) {
	if err := m.checkAccess(ctx, namespace, nil); err != nil {
		return nil, err
	}
	return resp.BuildListAPIResp(ctx, m.ClusterID, m.Kind, m.GroupVersion, namespace, format, opts)
}

// Get 请求某个资源详情，按指定 Format 格式化后返回
func (m *ResMgr) Get(ctx context.Context, namespace, name, format string, opts metav1.GetOptions) (*structpb.Struct, error) {
	if err := m.checkAccess(ctx, namespace, nil); err != nil {
		return nil, err
	}
	return resp.BuildRetrieveAPIResp(ctx, m.ClusterID, m.Kind, m.GroupVersion, namespace, name, format, opts)
}

// Create 创建 k8s 资源，支持以 manifest / formData 格式创建
func (m *ResMgr) Create(
	ctx context.Context, rawData *structpb.Struct, format string, isNSScoped bool, opts metav1.CreateOptions,
) (*structpb.Struct, error) {
	transformer, err := trans.New(ctx, rawData.AsMap(), m.ClusterID, m.Kind, format)
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
	transformer, err := trans.New(ctx, rawData.AsMap(), m.ClusterID, m.Kind, format)
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

// Delete 删除某个 k8s 资源
func (m *ResMgr) Delete(ctx context.Context, namespace, name string, opts metav1.DeleteOptions) error {
	if err := m.checkAccess(ctx, namespace, nil); err != nil {
		return err
	}
	return resp.BuildDeleteAPIResp(ctx, m.ClusterID, m.Kind, m.GroupVersion, namespace, name, opts)
}

// 访问权限检查（如共享集群禁用等）
func (m *ResMgr) checkAccess(ctx context.Context, namespace string, manifest map[string]interface{}) error {
	clusterInfo, err := cluster.GetClusterInfo(ctx, m.ClusterID)
	if err != nil {
		return err
	}
	// 独立集群中，不需要做类似校验
	if clusterInfo.Type == cluster.ClusterTypeSingle {
		return nil
	}
	// 不允许的资源类型，直接抛出错误
	if !slice.StringInSlice(m.Kind, cluster.SharedClusterAccessibleResKinds) {
		return errorx.New(errcode.NoPerm, "该请求资源类型在共享集群中不可用")
	}
	// 对命名空间进行检查，确保是属于项目的，命名空间以 manifest 中的为准
	if manifest != nil {
		namespace = mapx.GetStr(manifest, "metadata.namespace")
	}
	if !cli.IsProjNSinSharedCluster(ctx, m.ClusterID, namespace) {
		return errorx.New(errcode.NoPerm, "命名空间 %s 在该共享集群中不属于指定项目", namespace)
	}
	return nil
}
