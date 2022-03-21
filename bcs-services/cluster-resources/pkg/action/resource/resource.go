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
	structpb "github.com/golang/protobuf/ptypes/struct"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/action/util/resp"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/cluster"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/errcode"
	cli "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/client"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/errorx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/slice"
)

// ResMgr k8s 资源管理器，包含命名空间校验，集群操作下发，构建响应内容等功能
type ResMgr struct {
	ProjectID    string
	ClusterID    string
	GroupVersion string
	Kind         string
}

// NewResMgr 创建 ResMgr 并初始化
func NewResMgr(projectID, clusterID, groupVersion, kind string) *ResMgr {
	return &ResMgr{
		ProjectID:    projectID,
		ClusterID:    clusterID,
		GroupVersion: groupVersion,
		Kind:         kind,
	}
}

// List ...
func (m *ResMgr) List(namespace string, opts metav1.ListOptions) (*structpb.Struct, error) {
	if err := m.checkAccess(namespace, nil); err != nil {
		return nil, err
	}
	return resp.BuildListAPIResp(m.ClusterID, m.Kind, m.GroupVersion, namespace, opts)
}

// Get ...
func (m *ResMgr) Get(namespace, name string, opts metav1.GetOptions) (*structpb.Struct, error) {
	if err := m.checkAccess(namespace, nil); err != nil {
		return nil, err
	}
	return resp.BuildRetrieveAPIResp(m.ClusterID, m.Kind, m.GroupVersion, namespace, name, opts)
}

// Create ...
func (m *ResMgr) Create(manifest *structpb.Struct, isNSScoped bool, opts metav1.CreateOptions) (*structpb.Struct, error) {
	if err := m.checkAccess("", manifest); err != nil {
		return nil, err
	}
	return resp.BuildCreateAPIResp(m.ClusterID, m.Kind, m.GroupVersion, manifest, isNSScoped, opts)
}

// Update ...
func (m *ResMgr) Update(namespace, name string, manifest *structpb.Struct, opts metav1.UpdateOptions) (*structpb.Struct, error) {
	if err := m.checkAccess(namespace, manifest); err != nil {
		return nil, err
	}
	return resp.BuildUpdateAPIResp(m.ClusterID, m.Kind, m.GroupVersion, namespace, name, manifest, opts)
}

// Delete ...
func (m *ResMgr) Delete(namespace, name string, opts metav1.DeleteOptions) error {
	if err := m.checkAccess(namespace, nil); err != nil {
		return err
	}
	return resp.BuildDeleteAPIResp(m.ClusterID, m.Kind, m.GroupVersion, namespace, name, opts)
}

// 访问权限检查（如共享集群禁用等）
func (m *ResMgr) checkAccess(namespace string, manifest *structpb.Struct) error {
	clusterInfo, err := cluster.GetClusterInfo(m.ClusterID)
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
		namespace = mapx.Get(manifest.AsMap(), "metadata.namespace", "").(string)
	}
	if !cli.IsProjNSinSharedCluster(m.ProjectID, m.ClusterID, namespace) {
		return errorx.New(errcode.NoPerm, "命名空间 %s 在该共享集群中不属于指定项目", namespace)
	}
	return nil
}
