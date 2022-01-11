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

package util

import (
	"google.golang.org/protobuf/types/known/structpb"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	res "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/formatter"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util"
)

func BuildListApiResp(
	clusterID, resKind, groupVersion, namespace string, opts metav1.ListOptions,
) (*structpb.Struct, error) {
	clusterConf := res.NewClusterConfig(clusterID)
	k8sRes, err := res.GenGroupVersionResource(clusterConf, clusterID, resKind, groupVersion)
	if err != nil {
		return nil, err
	}

	// TODO 支持 namespace == "" 表示集群域资源
	ret, err := res.ListNamespaceScopedRes(clusterConf, namespace, k8sRes, opts)
	if err != nil {
		return nil, err
	}

	manifest := ret.UnstructuredContent()
	manifestExt := map[string]interface{}{}
	// 遍历列表中的每个资源，生成 manifestExt
	for _, item := range manifest["items"].([]interface{}) {
		uid, _ := util.GetItems(item.(map[string]interface{}), "metadata.uid")
		manifestExt[uid.(string)] = formatter.Kind2FormatFuncMap[resKind](item.(map[string]interface{}))
	}
	respData := map[string]interface{}{"manifest": manifest, "manifestExt": manifestExt}
	return util.Map2pbStruct(respData)
}

func BuildRetrieveApiResp(
	clusterID, resKind, groupVersion, namespace, name string, opts metav1.GetOptions,
) (*structpb.Struct, error) {
	clusterConf := res.NewClusterConfig(clusterID)
	k8sRes, err := res.GenGroupVersionResource(clusterConf, clusterID, resKind, groupVersion)
	if err != nil {
		return nil, err
	}

	// TODO 支持 namespace == "" 表示集群域资源
	ret, err := res.GetNamespaceScopedRes(clusterConf, namespace, name, k8sRes, opts)
	if err != nil {
		return nil, err
	}

	manifest := ret.UnstructuredContent()
	respData := map[string]interface{}{
		"manifest": manifest, "manifestExt": formatter.Kind2FormatFuncMap[resKind](manifest),
	}
	return util.Map2pbStruct(respData)
}

func BuildCreateApiResp(
	clusterID, resKind, groupVersion string,
	manifest *structpb.Struct,
	namespaceRequired bool,
	opts metav1.CreateOptions,
) (*structpb.Struct, error) {
	clusterConf := res.NewClusterConfig(clusterID)
	k8sRes, err := res.GenGroupVersionResource(clusterConf, clusterID, resKind, groupVersion)
	if err != nil {
		return nil, err
	}
	// TODO namespaceRequired == false 需要支持集群域资源
	ret, err := res.CreateNamespaceScopedRes(clusterConf, manifest.AsMap(), k8sRes, opts)
	if err != nil {
		return nil, err
	}
	return util.Unstructured2pbStruct(ret), nil
}

func BuildUpdateApiResp(
	clusterID, resKind, groupVersion, namespace, name string, manifest *structpb.Struct, opts metav1.UpdateOptions,
) (*structpb.Struct, error) {
	clusterConf := res.NewClusterConfig(clusterID)
	k8sRes, err := res.GenGroupVersionResource(clusterConf, clusterID, resKind, groupVersion)
	if err != nil {
		return nil, err
	}
	// TODO 支持 namespace == "" 表示集群域资源
	ret, err := res.UpdateNamespaceScopedRes(clusterConf, namespace, name, manifest.AsMap(), k8sRes, opts)
	if err != nil {
		return nil, err
	}
	return util.Unstructured2pbStruct(ret), nil
}

func BuildDeleteApiResp(
	clusterID, resKind, groupVersion, namespace, name string, opts metav1.DeleteOptions,
) error {
	clusterConf := res.NewClusterConfig(clusterID)
	k8sRes, err := res.GenGroupVersionResource(clusterConf, clusterID, resKind, groupVersion)
	if err != nil {
		return err
	}
	// TODO 支持 namespace == "" 表示集群域资源
	return res.DeleteNamespaceScopedRes(clusterConf, namespace, name, k8sRes, opts)
}
