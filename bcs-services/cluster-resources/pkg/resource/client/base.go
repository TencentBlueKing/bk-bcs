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

package client

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/cluster"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/action"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/errcode"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/project"
	res "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/perm"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/errorx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"
)

// NewDynamicClient ...
func NewDynamicClient(conf *res.ClusterConf) dynamic.Interface {
	dynamicCli, _ := dynamic.NewForConfig(conf.Rest)
	return dynamicCli
}

// ResClient K8S 集群资源管理客户端
type ResClient struct {
	cli  dynamic.Interface
	conf *res.ClusterConf
	res  schema.GroupVersionResource
}

// NewResClient ...
func NewResClient(conf *res.ClusterConf, resource schema.GroupVersionResource) *ResClient {
	return &ResClient{NewDynamicClient(conf), conf, resource}
}

// List 获取资源列表
func (c *ResClient) List(
	ctx context.Context, namespace string, opts metav1.ListOptions,
) (*unstructured.UnstructuredList, error) {
	if err := c.permValidate(ctx, action.List, namespace); err != nil {
		return nil, err
	}
	return c.cli.Resource(c.res).Namespace(namespace).List(ctx, opts)
}

// Get 获取单个资源
func (c *ResClient) Get(
	ctx context.Context, namespace, name string, opts metav1.GetOptions,
) (*unstructured.Unstructured, error) {
	if err := c.permValidate(ctx, action.View, namespace); err != nil {
		return nil, err
	}
	return c.cli.Resource(c.res).Namespace(namespace).Get(ctx, name, opts)
}

// Create 创建资源
func (c *ResClient) Create(
	ctx context.Context, manifest map[string]interface{}, isNSScoped bool, opts metav1.CreateOptions,
) (*unstructured.Unstructured, error) {
	namespace := ""
	if isNSScoped {
		namespace = mapx.GetStr(manifest, "metadata.namespace")
		if namespace == "" {
			return nil, errorx.New(errcode.ValidateErr, "创建 %s 需要指定 metadata.namespace", c.res.Resource)
		}
	}
	if err := c.permValidate(ctx, action.Create, namespace); err != nil {
		return nil, err
	}
	return c.cli.Resource(c.res).Namespace(namespace).Create(ctx, &unstructured.Unstructured{Object: manifest}, opts)
}

// Update 更新单个资源
func (c *ResClient) Update(
	ctx context.Context, namespace, name string, manifest map[string]interface{}, opts metav1.UpdateOptions,
) (*unstructured.Unstructured, error) {
	// 检查 name 与 manifest.metadata.name 是否一致
	manifestName, err := mapx.GetItems(manifest, "metadata.name")
	if err != nil || name != manifestName {
		return nil, errorx.New(errcode.ValidateErr, "metadata.name 必须指定且与准备编辑的资源名保持一致")
	}
	if err = c.permValidate(ctx, action.Update, namespace); err != nil {
		return nil, err
	}
	return c.cli.Resource(c.res).Namespace(namespace).Update(ctx, &unstructured.Unstructured{Object: manifest}, opts)
}

// Delete 删除单个资源
func (c *ResClient) Delete(ctx context.Context, namespace, name string, opts metav1.DeleteOptions) error {
	if err := c.permValidate(ctx, action.Delete, namespace); err != nil {
		return err
	}
	return c.cli.Resource(c.res).Namespace(namespace).Delete(ctx, name, opts)
}

// Watch 获取某类资源 watcher
func (c *ResClient) Watch(ctx context.Context, namespace string, opts metav1.ListOptions) (watch.Interface, error) {
	return c.cli.Resource(c.res).Namespace(namespace).Watch(ctx, opts)
}

// IAM 权限校验
func (c *ResClient) permValidate(ctx context.Context, action, namespace string) error {
	projInfo, err := project.FromContext(ctx)
	if err != nil {
		return errorx.New(errcode.General, "由 Context 获取项目信息失败")
	}
	clusterInfo, err := cluster.FromContext(ctx)
	if err != nil {
		return errorx.New(errcode.General, "由 Context 获取集群信息失败")
	}
	return perm.Validate(ctx, c.res.Resource, action, projInfo.ID, clusterInfo.ID, namespace)
}
