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

package handler

import (
	"context"

	spb "google.golang.org/protobuf/types/known/structpb"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/action"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/envs"
	cli "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/client"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"
	clusterRes "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/proto/cluster-resources"
)

// GenResListReq ...
func GenResListReq() clusterRes.ResListReq {
	return clusterRes.ResListReq{
		ProjectID: envs.TestProjectID,
		ClusterID: envs.TestClusterID,
		Namespace: envs.TestNamespace,
	}
}

// GenResCreateReq ...
func GenResCreateReq(manifest *spb.Struct) clusterRes.ResCreateReq {
	return clusterRes.ResCreateReq{
		ProjectID: envs.TestProjectID,
		ClusterID: envs.TestClusterID,
		RawData:   manifest,
		Format:    action.ManifestFormat,
	}
}

// GenResUpdateReq ...
func GenResUpdateReq(manifest *spb.Struct, name string) clusterRes.ResUpdateReq {
	return clusterRes.ResUpdateReq{
		ProjectID: envs.TestProjectID,
		ClusterID: envs.TestClusterID,
		Namespace: envs.TestNamespace,
		Name:      name,
		RawData:   manifest,
		Format:    action.ManifestFormat,
	}
}

// GenResGetReq ...
func GenResGetReq(name string) clusterRes.ResGetReq {
	return clusterRes.ResGetReq{
		ProjectID: envs.TestProjectID,
		ClusterID: envs.TestClusterID,
		Namespace: envs.TestNamespace,
		Name:      name,
		Format:    action.ManifestFormat,
	}
}

// GenResDeleteReq ...
func GenResDeleteReq(name string) clusterRes.ResDeleteReq {
	return clusterRes.ResDeleteReq{
		ProjectID: envs.TestProjectID,
		ClusterID: envs.TestClusterID,
		Namespace: envs.TestNamespace,
		Name:      name,
	}
}

var nsManifest4Test = map[string]interface{}{
	"apiVersion": "v1",
	"kind":       "Namespace",
	"metadata": map[string]interface{}{
		"name":        "new_name_required",
		"annotations": map[string]interface{}{},
	},
}

// GetOrCreateNS 在集群中初始化命名空间用于单元测试用
func GetOrCreateNS(namespace string) error {
	if namespace == "" {
		namespace = envs.TestNamespace
	}
	ctx := context.TODO()
	nsCli := cli.NewNSCliByClusterID(ctx, envs.TestClusterID)
	_, err := nsCli.Get(ctx, "", namespace, metav1.GetOptions{})
	if err != nil {
		_ = mapx.SetItems(nsManifest4Test, "metadata.name", namespace)
		if namespace == envs.TestSharedClusterNS {
			_ = mapx.SetItems(nsManifest4Test, []string{"metadata", "annotations", cli.ProjCodeAnnoKey}, envs.TestProjectCode)
		}
		_, err = nsCli.Create(ctx, nsManifest4Test, false, metav1.CreateOptions{})
	}
	return err
}

// CRDName4Test ...
var CRDName4Test = "crontabs.stable.example.com"

// CRDManifest4Test ...
var CRDManifest4Test = map[string]interface{}{
	"apiVersion": "apiextensions.k8s.io/v1",
	"kind":       "CustomResourceDefinition",
	"metadata": map[string]interface{}{
		"name": "crontabs.stable.example.com",
	},
	"spec": map[string]interface{}{
		"group": "stable.example.com",
		"versions": []interface{}{
			map[string]interface{}{
				"name":    "v1",
				"served":  true,
				"storage": true,
				"schema": map[string]interface{}{
					"openAPIV3Schema": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"spec": map[string]interface{}{
								"type": "object",
								"properties": map[string]interface{}{
									"cronSpec": map[string]interface{}{
										"type": "string",
									},
									"image": map[string]interface{}{
										"type": "string",
									},
									"replicas": map[string]interface{}{
										"type": "integer",
									},
								},
							},
						},
					},
				},
			},
		},
		"scope": "Namespaced",
		"names": map[string]interface{}{
			"plural":   "crontabs",
			"singular": "crontab",
			"kind":     "CronTab",
			"shortNames": []interface{}{
				"ct",
			},
		},
	},
}

// GetOrCreateCRD 在集群中初始化 CRD 用于单元测试用
func GetOrCreateCRD() error {
	ctx := context.TODO()
	crdCli := cli.NewCRDCliByClusterID(ctx, envs.TestClusterID)
	_, err := crdCli.Get(ctx, "", CRDName4Test, metav1.GetOptions{})
	if err != nil {
		// TODO 这里认为出错就是不存在，可以做进一步的细化？
		_, err = crdCli.Create(ctx, CRDManifest4Test, false, metav1.CreateOptions{})
	}
	return err
}
