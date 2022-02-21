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
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/envs"
	res "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource"
	cli "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/client"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/example"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util"
	clusterRes "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/proto/cluster-resources"
)

var crdName4Test = "crontabs.stable.example.com"

var crdManifest4Test = map[string]interface{}{
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

var cobjName4Test = "crontab-test-" + util.GenRandStr(example.RandomSuffixLength, example.SuffixCharset)

var cobjManifest4Test = map[string]interface{}{
	"apiVersion": "stable.example.com/v1",
	"kind":       "CronTab",
	"metadata": map[string]interface{}{
		"name":      cobjName4Test,
		"namespace": envs.TestNamespace,
	},
	"spec": map[string]interface{}{
		"cronSpec": "* * * * */10",
		"image":    "my-awesome-cron-image",
	},
}

// 在集群中初始化 CRD 用于单元测试用
func getOrCreateCRD() error {
	clusterConf := res.NewClusterConfig(envs.TestClusterID)
	crdRes, err := res.GetGroupVersionResource(clusterConf, res.CRD, "")
	if err != nil {
		return err
	}

	crdCli := cli.NewResClient(clusterConf, crdRes)
	_, err = crdCli.Get("", crdName4Test, metav1.GetOptions{})
	if err != nil {
		// TODO 这里认为出错就是不存在，可以做进一步的细化？
		_, err = crdCli.Create(crdManifest4Test, false, metav1.CreateOptions{})
	}
	return err
}

func TestCRD(t *testing.T) {
	// 在集群中初始化 CRD
	err := getOrCreateCRD()
	assert.Nil(t, err)

	crh := NewClusterResourcesHandler()
	ctx := context.TODO()

	// List
	listReq, listResp := genResListReq(), clusterRes.CommonResp{}
	err = crh.ListCRD(ctx, &listReq, &listResp)
	assert.Nil(t, err)

	respData := listResp.Data.AsMap()
	assert.Equal(t, "CustomResourceDefinitionList", util.GetWithDefault(respData, "manifest.kind", ""))

	// Get
	getReq, getResp := genResGetReq(crdName4Test), clusterRes.CommonResp{}
	err = crh.GetCRD(ctx, &getReq, &getResp)
	assert.Nil(t, err)

	respData = getResp.Data.AsMap()
	assert.Equal(t, "CustomResourceDefinition", util.GetWithDefault(respData, "manifest.kind", ""))
	assert.Equal(t, "Namespaced", util.GetWithDefault(respData, "manifest.spec.scope", ""))
}

func TestCObj(t *testing.T) {
	// 在集群中初始化 CRD
	err := getOrCreateCRD()
	assert.Nil(t, err)

	crh := NewClusterResourcesHandler()
	ctx := context.TODO()

	// Create
	createManifest, _ := util.Map2pbStruct(cobjManifest4Test)
	createReq := clusterRes.CObjCreateReq{
		ProjectID: envs.TestProjectID,
		ClusterID: envs.TestClusterID,
		CRDName:   crdName4Test,
		Manifest:  createManifest,
	}
	err = crh.CreateCObj(ctx, &createReq, &clusterRes.CommonResp{})
	assert.Nil(t, err)

	// List
	listReq := clusterRes.CObjListReq{
		ProjectID: envs.TestProjectID,
		ClusterID: envs.TestClusterID,
		CRDName:   crdName4Test,
		Namespace: envs.TestNamespace,
	}
	listResp := clusterRes.CommonResp{}
	err = crh.ListCObj(ctx, &listReq, &listResp)
	assert.Nil(t, err)

	respData := listResp.Data.AsMap()
	assert.Equal(t, "CronTabList", util.GetWithDefault(respData, "manifest.kind", ""))

	// Update
	_ = util.SetItems(cobjManifest4Test, "spec.cronSpec", "* * * * */5")
	updateManifest, _ := util.Map2pbStruct(cobjManifest4Test)
	updateReq := clusterRes.CObjUpdateReq{
		ProjectID: envs.TestProjectID,
		ClusterID: envs.TestClusterID,
		CRDName:   crdName4Test,
		CobjName:  cobjName4Test,
		Namespace: envs.TestNamespace,
		Manifest:  updateManifest,
	}
	err = crh.UpdateCObj(ctx, &updateReq, &clusterRes.CommonResp{})
	assert.Nil(t, err)

	// Get
	getReq := clusterRes.CObjGetReq{
		ProjectID: envs.TestProjectID,
		ClusterID: envs.TestClusterID,
		CRDName:   crdName4Test,
		CobjName:  cobjName4Test,
		Namespace: envs.TestNamespace,
	}
	getResp := clusterRes.CommonResp{}
	err = crh.GetCObj(ctx, &getReq, &getResp)
	assert.Nil(t, err)

	respData = getResp.Data.AsMap()
	assert.Equal(t, "CronTab", util.GetWithDefault(respData, "manifest.kind", ""))
	assert.Equal(t, "* * * * */5", util.GetWithDefault(respData, "manifest.spec.cronSpec", ""))

	// Delete
	deleteReq := clusterRes.CObjDeleteReq{
		ProjectID: envs.TestProjectID,
		ClusterID: envs.TestClusterID,
		CRDName:   crdName4Test,
		CobjName:  cobjName4Test,
		Namespace: envs.TestNamespace,
	}
	err = crh.DeleteCObj(ctx, &deleteReq, &clusterRes.CommonResp{})
	assert.Nil(t, err)
}
