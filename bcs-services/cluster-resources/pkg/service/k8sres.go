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

// Package service k8sres.go k8s 资源管理相关逻辑
package service

import (
	"fmt"

	structpb "github.com/golang/protobuf/ptypes/struct"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/cluster"
	cli "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/client"
	respUtil "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/service/util/resp"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util"
)

// K8SResMgr k8s 资源管理器，包含命名空间校验，集群操作下发，构建响应内容等功能
type K8SResMgr struct {
	ProjectID    string
	ClusterID    string
	GroupVersion string
	Kind         string
}

// NewK8SResMgr 创建 K8SResMgr 并初始化
func NewK8SResMgr(projectID, clusterID, groupVersion, kind string) *K8SResMgr {
	return &K8SResMgr{
		ProjectID:    projectID,
		ClusterID:    clusterID,
		GroupVersion: groupVersion,
		Kind:         kind,
	}
}

// List ...
func (m *K8SResMgr) List(namespace string, opts metav1.ListOptions) (*structpb.Struct, error) {
	if err := m.accessibleCheck(namespace, nil); err != nil {
		return nil, err
	}
	return respUtil.BuildListAPIResp(m.ClusterID, m.Kind, m.GroupVersion, namespace, opts)
}

// Get ...
func (m *K8SResMgr) Get(namespace, name string, opts metav1.GetOptions) (*structpb.Struct, error) {
	if err := m.accessibleCheck(namespace, nil); err != nil {
		return nil, err
	}
	return respUtil.BuildRetrieveAPIResp(m.ClusterID, m.Kind, m.GroupVersion, namespace, name, opts)
}

// Create ...
func (m *K8SResMgr) Create(manifest *structpb.Struct, isNSScoped bool, opts metav1.CreateOptions) (*structpb.Struct, error) {
	if err := m.accessibleCheck("", manifest); err != nil {
		return nil, err
	}
	return respUtil.BuildCreateAPIResp(m.ClusterID, m.Kind, m.GroupVersion, manifest, isNSScoped, opts)
}

// Update ...
func (m *K8SResMgr) Update(namespace, name string, manifest *structpb.Struct, opts metav1.UpdateOptions) (*structpb.Struct, error) {
	if err := m.accessibleCheck(namespace, manifest); err != nil {
		return nil, err
	}
	return respUtil.BuildUpdateAPIResp(m.ClusterID, m.Kind, m.GroupVersion, namespace, name, manifest, opts)
}

// Delete ...
func (m *K8SResMgr) Delete(namespace, name string, opts metav1.DeleteOptions) error {
	if err := m.accessibleCheck(namespace, nil); err != nil {
		return err
	}
	return respUtil.BuildDeleteAPIResp(m.ClusterID, m.Kind, m.GroupVersion, namespace, name, opts)
}

// 访问权限检查（如共享集群禁用等）
func (m *K8SResMgr) accessibleCheck(namespace string, manifest *structpb.Struct) error {
	clusterInfo, err := cluster.GetClusterInfo(m.ClusterID)
	if err != nil {
		return err
	}
	// 独立集群中，不需要做类似校验
	if clusterInfo.Type == cluster.ClusterTypeSingle {
		return nil
	}
	// 不允许的资源类型，直接抛出错误
	if !util.StringInSlice(m.Kind, cluster.SharedClusterAccessibleResKinds) {
		return fmt.Errorf("该请求资源类型在共享集群中不可用")
	}
	// 对命名空间进行检查，确保是属于项目的
	if manifest != nil {
		namespace = util.GetWithDefault(manifest.AsMap(), "metadata.namespace", "").(string)
	}
	if !cli.IsProjNSinSharedCluster(m.ProjectID, m.ClusterID, namespace) {
		return fmt.Errorf("命名空间 %s 在该共享集群中不属于指定项目", namespace)
	}
	return nil
}
