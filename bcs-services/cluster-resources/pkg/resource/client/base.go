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
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"

	res "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util"
)

// NewDynamicClient ...
func NewDynamicClient(conf *res.ClusterConf) dynamic.Interface {
	client, _ := dynamic.NewForConfig(conf.Rest)
	return client
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
func (c *ResClient) List(namespace string, opts metav1.ListOptions) (*unstructured.UnstructuredList, error) {
	return c.cli.Resource(c.res).Namespace(namespace).List(context.TODO(), opts)
}

// Get 获取单个资源
func (c *ResClient) Get(namespace, name string, opts metav1.GetOptions) (*unstructured.Unstructured, error) {
	return c.cli.Resource(c.res).Namespace(namespace).Get(context.TODO(), name, opts)
}

// Create 创建资源
func (c *ResClient) Create(
	manifest map[string]interface{}, isNamespaceScoped bool, opts metav1.CreateOptions,
) (*unstructured.Unstructured, error) {
	namespace := ""
	if isNamespaceScoped {
		namespace = util.GetWithDefault(manifest, "metadata.namespace", "").(string)
		if namespace == "" {
			return nil, fmt.Errorf("创建 %s 需要指定 metadata.namespace", c.res.Resource)
		}
	}
	return c.cli.Resource(c.res).Namespace(namespace).Create(
		context.TODO(), &unstructured.Unstructured{Object: manifest}, opts,
	)
}

// Update 更新单个资源
func (c *ResClient) Update(
	namespace, name string, manifest map[string]interface{}, opts metav1.UpdateOptions,
) (*unstructured.Unstructured, error) {
	// 检查 name 与 manifest.metadata.name 是否一致
	manifestName, err := util.GetItems(manifest, "metadata.name")
	if err != nil || name != manifestName {
		return nil, fmt.Errorf("metadata.name 必须指定且与准备编辑的资源名保持一致")
	}
	return c.cli.Resource(c.res).Namespace(namespace).Update(
		context.TODO(), &unstructured.Unstructured{Object: manifest}, opts,
	)
}

// Delete 删除单个资源
func (c *ResClient) Delete(namespace, name string, opts metav1.DeleteOptions) error {
	return c.cli.Resource(c.res).Namespace(namespace).Delete(context.TODO(), name, opts)
}
