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

package client

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	res "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource"
)

// ApiResourceClient xxx
type ApiResourceClient struct {
	ResClient
}

// NewApiResourceClient xxx
func NewApiResourceClient(
	ctx context.Context, kind, apiVersion string, conf *res.ClusterConf) *ApiResourceClient {
	gvr, _ := res.GetGroupVersionResource(ctx, conf, kind, apiVersion)
	return &ApiResourceClient{ResClient{NewDynamicClient(conf), conf, gvr}}
}

// NewApiResourceClientByClusterID xxx
func NewApiResourceClientByClusterID(
	ctx context.Context, clusterID, kind, apiVersion string) *ApiResourceClient {
	return NewApiResourceClient(ctx, kind, apiVersion, res.NewClusterConf(clusterID))
}

// GetResObjectInfo 获取 api-resources object 基础信息
func GetResObjectInfo(
	ctx context.Context, clusterID, namespace, name, kind, apiVersion string) (*unstructured.Unstructured, error) {
	return NewApiResourceClientByClusterID(ctx, clusterID, kind, apiVersion).Get(ctx, namespace, name, metav1.GetOptions{})
}

// CreateResObjectInfo 创建 api-resources object 基础信息
func CreateResObjectInfo(ctx context.Context, clusterID, kind, apiVersion string, namespaced bool,
	manifest map[string]interface{}) (*unstructured.Unstructured, error) {
	return NewApiResourceClientByClusterID(ctx, clusterID, kind, apiVersion).
		Create(ctx, manifest, namespaced, metav1.CreateOptions{})
}

// UpdateResObjectInfo 更新 api-resources object 基础信息
func UpdateResObjectInfo(ctx context.Context, clusterID, kind, apiVersion string,
	manifest map[string]interface{}) (*unstructured.Unstructured, error) {
	return NewApiResourceClientByClusterID(ctx, clusterID, kind, apiVersion).
		ApplyWithoutPerm(ctx, manifest, metav1.CreateOptions{})
}

// DeleteResObjectInfo 删除 api-resources object 基础信息
func DeleteResObjectInfo(ctx context.Context, clusterID, namespace, name, kind, apiVersion string) error {
	return NewApiResourceClientByClusterID(ctx, clusterID, kind, apiVersion).
		Delete(ctx, namespace, name, metav1.DeleteOptions{})
}
