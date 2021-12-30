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

package resources

import (
	"context"
	"errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/utils"
)

// ListNamespaceScopedRes 获取命名空间维度资源列表
func ListNamespaceScopedRes(
	namespace string,
	resource schema.GroupVersionResource,
	opts metav1.ListOptions,
) (*unstructured.UnstructuredList, error) {
	client := newResourceClient()
	return client.Resource(resource).Namespace(namespace).List(context.TODO(), opts)
}

// GetNamespaceScopedRes 获取单个命名空间维度资源
func GetNamespaceScopedRes(
	namespace string,
	name string,
	resource schema.GroupVersionResource,
	opts metav1.GetOptions,
) (*unstructured.Unstructured, error) {
	client := newResourceClient()
	return client.Resource(resource).Namespace(namespace).Get(context.TODO(), name, opts)
}

// CreateNamespaceScopedRes 创建命名空间维度资源
func CreateNamespaceScopedRes(
	manifest map[string]interface{},
	resource schema.GroupVersionResource,
	opts metav1.CreateOptions,
) (*unstructured.Unstructured, error) {
	client := newResourceClient()
	namespace, err := utils.GetItems(manifest, []string{"metadata", "namespace"})
	if err != nil {
		return nil, errors.New("创建 Deployment 需要指定 metadata.namespace")
	}
	return client.Resource(resource).Namespace(namespace.(string)).Create(
		context.TODO(), &unstructured.Unstructured{Object: manifest}, opts,
	)
}

// UpdateNamespaceScopedRes 更新单个命名空间维度资源
func UpdateNamespaceScopedRes(
	namespace string,
	name string,
	manifest map[string]interface{},
	resource schema.GroupVersionResource,
	opts metav1.UpdateOptions,
) (*unstructured.Unstructured, error) {
	client := newResourceClient()
	// 检查 name 与 manifest.metadata.name 是否一致
	manifestName, err := utils.GetItems(manifest, []string{"metadata", "name"})
	if err != nil || name != manifestName {
		return nil, errors.New("metadata.name 必须指定且与准备编辑的资源名保持一致")
	}
	return client.Resource(resource).Namespace(namespace).Update(
		context.TODO(), &unstructured.Unstructured{Object: manifest}, opts,
	)
}

// DeleteNamespaceScopedRes 删除单个命名空间维度资源
func DeleteNamespaceScopedRes(
	namespace string,
	name string,
	resource schema.GroupVersionResource,
	opts metav1.DeleteOptions,
) error {
	client := newResourceClient()
	return client.Resource(resource).Namespace(namespace).Delete(context.TODO(), name, opts)
}
