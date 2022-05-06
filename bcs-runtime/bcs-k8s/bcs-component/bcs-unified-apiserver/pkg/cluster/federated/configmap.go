/*
 * Tencent is pleased to support the open source community by making Blueking Container ConfigMap available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package federated

import (
	"context"
	"time"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-unified-apiserver/pkg/clientutil"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
)

// ConfigMapStor
type ConfigMapStor struct {
	members      []string
	masterClient *kubernetes.Clientset
	k8sClientMap map[string]*kubernetes.Clientset
}

// NewConfigMapStor
func NewConfigMapStor(masterClientId string, members []string) (*ConfigMapStor, error) {
	stor := &ConfigMapStor{members: members, k8sClientMap: make(map[string]*kubernetes.Clientset)}
	for _, k := range members {
		k8sClient, err := clientutil.GetClusternetClientByClusterId(k)
		if err != nil {
			return nil, err
		}
		stor.k8sClientMap[k] = k8sClient
	}
	masterClient, err := clientutil.GetClusternetClientByClusterId(masterClientId)
	if err != nil {
		return nil, err
	}
	stor.masterClient = masterClient
	return stor, nil
}

// List 查询 ConfigMap 列表, Json格式返回
func (s *ConfigMapStor) List(ctx context.Context, namespace string, opts metav1.ListOptions) (*v1.ConfigMapList, error) {
	result, err := s.masterClient.CoreV1().ConfigMaps(namespace).List(ctx, opts)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// ListAsTable 查询 ConfigMap 列表, kubectl 格式返回
func (s *ConfigMapStor) ListAsTable(ctx context.Context, namespace string, acceptHeader string, opts metav1.ListOptions) (*metav1.Table, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}

	result := &metav1.Table{}
	err := s.masterClient.CoreV1().RESTClient().Get().
		Namespace(namespace).
		Resource("ConfigMaps").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		SetHeader("Accept", acceptHeader).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// Get 获取单个 ConfigMap, Json格式返回
func (s *ConfigMapStor) Get(ctx context.Context, namespace string, name string, opts metav1.GetOptions) (*v1.ConfigMap, error) {
	result, err := s.masterClient.CoreV1().ConfigMaps(namespace).Get(ctx, name, opts)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetAsTable 获取单个 ConfigMap, Table 格式返回
func (s *ConfigMapStor) GetAsTable(ctx context.Context, namespace string, name string, acceptHeader string, opts metav1.GetOptions) (*metav1.Table, error) {
	result := &metav1.Table{}
	err := s.masterClient.CoreV1().RESTClient().Get().
		Namespace(namespace).
		Resource("ConfigMaps").
		VersionedParams(&opts, scheme.ParameterCodec).
		SetHeader("Accept", acceptHeader).
		SubResource(name).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// Create 创建 ConfigMap
func (s *ConfigMapStor) Create(ctx context.Context, namespace string, ConfigMap *v1.ConfigMap, opts metav1.CreateOptions) (*v1.ConfigMap, error) {
	result, err := s.masterClient.CoreV1().ConfigMaps(namespace).Create(ctx, ConfigMap, opts)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Update 更新 ConfigMap
func (s *ConfigMapStor) Update(ctx context.Context, namespace string, ConfigMap *v1.ConfigMap, opts metav1.UpdateOptions) (*v1.ConfigMap, error) {
	result, err := s.masterClient.CoreV1().ConfigMaps(namespace).Update(ctx, ConfigMap, opts)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Patch Edit/Apply ConfigMap
func (s *ConfigMapStor) Patch(ctx context.Context, namespace string, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (*v1.ConfigMap, error) {
	result, err := s.masterClient.CoreV1().ConfigMaps(namespace).Patch(ctx, name, pt, data, opts, subresources...)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Delete 删除单个 ConfigMap
func (s *ConfigMapStor) Delete(ctx context.Context, namespace string, name string, opts metav1.DeleteOptions) (*metav1.Status, error) {
	result, err := s.masterClient.CoreV1().ConfigMaps(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	if err := s.masterClient.CoreV1().ConfigMaps(namespace).Delete(ctx, name, opts); err != nil {
		return nil, err
	}

	// ConfigMap 删除是返回标准 status 数据格式
	detailStatus := &metav1.StatusDetails{
		Name:  name,
		Group: "apps",
		Kind:  "ConfigMaps",
		UID:   result.UID,
	}
	status := &metav1.Status{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Status",
			APIVersion: "v1",
		},
		Status:  metav1.StatusSuccess,
		Details: detailStatus,
	}

	return status, nil
}
