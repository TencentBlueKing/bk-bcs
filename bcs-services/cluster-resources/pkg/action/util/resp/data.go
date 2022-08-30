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

package resp

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	res "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource"
	cli "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/client"
)

// BuildListAPIRespData xxx
func BuildListAPIRespData(
	ctx context.Context, clusterID, resKind, groupVersion, namespace, format string, opts metav1.ListOptions,
) (map[string]interface{}, error) {
	clusterConf := res.NewClusterConfig(clusterID)
	k8sRes, err := res.GetGroupVersionResource(ctx, clusterConf, resKind, groupVersion)
	if err != nil {
		return nil, err
	}

	var ret *unstructured.UnstructuredList
	ret, err = cli.NewResClient(clusterConf, k8sRes).List(ctx, namespace, opts)
	if err != nil {
		return nil, err
	}

	respDataBuilder, err := NewRespDataBuilder(ctx, ret.UnstructuredContent(), resKind, format)
	if err != nil {
		return nil, err
	}
	return respDataBuilder.BuildList()
}

// BuildRetrieveAPIRespData xxx
func BuildRetrieveAPIRespData(
	ctx context.Context, clusterID, resKind, groupVersion, namespace, name, format string, opts metav1.GetOptions,
) (map[string]interface{}, error) {
	clusterConf := res.NewClusterConfig(clusterID)
	k8sRes, err := res.GetGroupVersionResource(ctx, clusterConf, resKind, groupVersion)
	if err != nil {
		return nil, err
	}

	var ret *unstructured.Unstructured
	ret, err = cli.NewResClient(clusterConf, k8sRes).Get(ctx, namespace, name, opts)
	if err != nil {
		return nil, err
	}

	respDataBuilder, err := NewRespDataBuilder(ctx, ret.UnstructuredContent(), resKind, format)
	if err != nil {
		return nil, err
	}
	return respDataBuilder.Build()
}
