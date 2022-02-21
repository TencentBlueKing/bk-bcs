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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"

	res "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/formatter"
)

// GetCRDInfo 获取 CRD 基础信息
func GetCRDInfo(clusterID, crdName string) (map[string]interface{}, error) {
	clusterConf := res.NewClusterConfig(clusterID)
	crdRes, err := res.GetGroupVersionResource(clusterConf, res.CRD, "")
	if err != nil {
		return nil, err
	}

	var ret *unstructured.Unstructured
	ret, err = NewResClient(clusterConf, crdRes).Get("", crdName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	manifest := ret.UnstructuredContent()
	return formatter.FormatCRD(manifest), nil
}

// GetCObjManifest 获取自定义资源信息
func GetCObjManifest(
	clusterConf *res.ClusterConf, cobjRes schema.GroupVersionResource, namespace, cobjName string,
) (manifest map[string]interface{}, err error) {
	var ret *unstructured.Unstructured
	ret, err = NewResClient(clusterConf, cobjRes).Get(namespace, cobjName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return ret.UnstructuredContent(), nil
}
