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

package workload

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/action"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/envs"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/handler"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/example"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/form/parser/workload"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/pbstruct"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/stringx"
	clusterRes "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/proto/cluster-resources"
)

func TestDeploy(t *testing.T) {
	h := New()
	ctx := context.TODO()

	manifest, _ := example.LoadDemoManifest("workload/simple_deployment", "")
	resName := mapx.Get(manifest, "metadata.name", "")

	// Create
	createManifest, _ := pbstruct.Map2pbStruct(manifest)
	createReq := handler.GenResCreateReq(createManifest)
	err := h.CreateDeploy(ctx, &createReq, &clusterRes.CommonResp{})
	assert.Nil(t, err)

	// List
	listReq, listResp := handler.GenResListReq(), clusterRes.CommonResp{}
	err = h.ListDeploy(ctx, &listReq, &listResp)
	assert.Nil(t, err)

	respData := listResp.Data.AsMap()
	assert.Equal(t, "DeploymentList", mapx.Get(respData, "manifest.kind", ""))

	// Update
	_ = mapx.SetItems(manifest, "spec.replicas", 5)
	updateManifest, _ := pbstruct.Map2pbStruct(manifest)
	updateReq := handler.GenResUpdateReq(updateManifest, resName.(string))
	err = h.UpdateDeploy(ctx, &updateReq, &clusterRes.CommonResp{})
	assert.Nil(t, err)

	// Get
	getReq, getResp := handler.GenResGetReq(resName.(string)), clusterRes.CommonResp{}
	err = h.GetDeploy(ctx, &getReq, &getResp)
	assert.Nil(t, err)

	respData = getResp.Data.AsMap()
	assert.Equal(t, "Deployment", mapx.Get(respData, "manifest.kind", ""))
	assert.Equal(t, float64(5), mapx.Get(respData, "manifest.spec.replicas", 0))

	// Delete
	deleteReq := handler.GenResDeleteReq(resName.(string))
	err = h.DeleteDeploy(ctx, &deleteReq, &clusterRes.CommonResp{})
	assert.Nil(t, err)
}

var deployManifest4FormTest = map[string]interface{}{
	"apiVersion": "apps/v1",
	"kind":       "Deployment",
	"metadata": map[string]interface{}{
		"name":      "deployment-test-" + stringx.Rand(example.RandomSuffixLength, example.SuffixCharset),
		"namespace": envs.TestNamespace,
		"labels": map[string]interface{}{
			"app": "busybox",
		},
	},
	"spec": map[string]interface{}{
		"replicas": int64(2),
		"selector": map[string]interface{}{
			"matchLabels": map[string]interface{}{
				"app": "busybox",
			},
		},
		"template": map[string]interface{}{
			"metadata": map[string]interface{}{
				"labels": map[string]interface{}{
					"app": "busybox",
				},
			},
			"spec": map[string]interface{}{
				"containers": []interface{}{
					map[string]interface{}{
						"name":  "busybox",
						"image": "busybox:latest",
						"ports": []interface{}{
							map[string]interface{}{
								"containerPort": int64(80),
							},
						},
						"command": []interface{}{
							"/bin/sh",
							"-c",
						},
						"args": []interface{}{
							"echo hello",
						},
					},
				},
			},
		},
	},
}

func TestDeployWithForm(t *testing.T) {
	h := New()
	ctx := context.TODO()

	resName := mapx.Get(deployManifest4FormTest, "metadata.name", "")

	// Create by form data
	formData, _ := pbstruct.Map2pbStruct(workload.ParseDeploy(deployManifest4FormTest))
	createReq := clusterRes.ResCreateReq{
		ProjectID: envs.TestProjectID,
		ClusterID: envs.TestClusterID,
		RawData:   formData,
		Format:    action.FormDataFormat,
	}
	err := h.CreateDeploy(ctx, &createReq, &clusterRes.CommonResp{})
	assert.Nil(t, err)

	// Update
	_ = mapx.SetItems(deployManifest4FormTest, "spec.replicas", int64(3))
	formData, _ = pbstruct.Map2pbStruct(workload.ParseDeploy(deployManifest4FormTest))
	updateReq := clusterRes.ResUpdateReq{
		ProjectID: envs.TestProjectID,
		ClusterID: envs.TestClusterID,
		Namespace: envs.TestNamespace,
		Name:      resName.(string),
		RawData:   formData,
		Format:    action.FormDataFormat,
	}
	err = h.UpdateDeploy(ctx, &updateReq, &clusterRes.CommonResp{})
	assert.Nil(t, err)

	// Get FormData
	getReq := clusterRes.ResGetReq{
		ProjectID: envs.TestProjectID,
		ClusterID: envs.TestClusterID,
		Namespace: envs.TestNamespace,
		Name:      resName.(string),
		Format:    action.FormDataFormat,
	}
	getResp := clusterRes.CommonResp{}
	err = h.GetDeploy(ctx, &getReq, &getResp)
	assert.Nil(t, err)

	// Get Manifest
	getReq, getResp = handler.GenResGetReq(resName.(string)), clusterRes.CommonResp{}
	err = h.GetDeploy(ctx, &getReq, &getResp)
	assert.Nil(t, err)

	respData := getResp.Data.AsMap()
	assert.Equal(t, "Deployment", mapx.Get(respData, "manifest.kind", ""))
	assert.Equal(t, float64(3), mapx.Get(respData, "manifest.spec.replicas", 0))

	// Delete
	deleteReq := handler.GenResDeleteReq(resName.(string))
	err = h.DeleteDeploy(ctx, &deleteReq, &clusterRes.CommonResp{})
	assert.Nil(t, err)
}

func TestDeployInSharedCluster(t *testing.T) {
	// 在共享集群中新建命名空间
	err := handler.GetOrCreateNS(envs.TestSharedClusterNS)
	assert.Nil(t, err)

	h := New()
	ctx := context.TODO()

	manifest, _ := example.LoadDemoManifest("workload/simple_deployment", "")
	resName := mapx.Get(manifest, "metadata.name", "")
	// 设置为共享集群项目属命名空间
	err = mapx.SetItems(manifest, "metadata.namespace", envs.TestSharedClusterNS)
	assert.Nil(t, err)

	// Create
	createManifest, _ := pbstruct.Map2pbStruct(manifest)
	createReq := clusterRes.ResCreateReq{
		ProjectID: envs.TestProjectID,
		ClusterID: envs.TestSharedClusterID,
		RawData:   createManifest,
		Format:    action.ManifestFormat,
	}
	err = h.CreateDeploy(ctx, &createReq, &clusterRes.CommonResp{})
	assert.Nil(t, err)

	// List
	listReq := clusterRes.ResListReq{
		ProjectID: envs.TestProjectID,
		ClusterID: envs.TestSharedClusterID,
		Namespace: envs.TestSharedClusterNS,
	}
	err = h.ListDeploy(ctx, &listReq, &clusterRes.CommonResp{})
	assert.Nil(t, err)

	// Update
	updateReq := clusterRes.ResUpdateReq{
		ProjectID: envs.TestProjectID,
		ClusterID: envs.TestSharedClusterID,
		Namespace: envs.TestSharedClusterNS,
		Name:      resName.(string),
		RawData:   createManifest,
		Format:    action.ManifestFormat,
	}
	err = h.UpdateDeploy(ctx, &updateReq, &clusterRes.CommonResp{})
	assert.Nil(t, err)

	// Get
	getReq := clusterRes.ResGetReq{
		ProjectID: envs.TestProjectID,
		ClusterID: envs.TestSharedClusterID,
		Namespace: envs.TestSharedClusterNS,
		Name:      resName.(string),
		Format:    action.ManifestFormat,
	}
	err = h.GetDeploy(ctx, &getReq, &clusterRes.CommonResp{})
	assert.Nil(t, err)

	// Delete
	deleteReq := clusterRes.ResDeleteReq{
		ProjectID: envs.TestProjectID,
		ClusterID: envs.TestSharedClusterID,
		Namespace: envs.TestSharedClusterNS,
		Name:      resName.(string),
	}
	err = h.DeleteDeploy(ctx, &deleteReq, &clusterRes.CommonResp{})
	assert.Nil(t, err)
}

func TestDeployInSharedClusterNoPerm(t *testing.T) {
	h := New()
	ctx := context.TODO()

	manifest, _ := example.LoadDemoManifest("workload/simple_deployment", "")
	resName := mapx.Get(manifest, "metadata.name", "")

	// Create
	createManifest, _ := pbstruct.Map2pbStruct(manifest)
	createReq := clusterRes.ResCreateReq{
		ProjectID: envs.TestProjectID,
		ClusterID: envs.TestSharedClusterID,
		RawData:   createManifest,
		Format:    action.ManifestFormat,
	}
	err := h.CreateDeploy(ctx, &createReq, &clusterRes.CommonResp{})
	assert.NotNil(t, err)

	// List
	listReq := clusterRes.ResListReq{
		ProjectID: envs.TestProjectID,
		ClusterID: envs.TestSharedClusterID,
		Namespace: envs.TestNamespace,
	}
	err = h.ListDeploy(ctx, &listReq, &clusterRes.CommonResp{})
	assert.NotNil(t, err)

	// Update
	updateReq := clusterRes.ResUpdateReq{
		ProjectID: envs.TestProjectID,
		ClusterID: envs.TestSharedClusterID,
		Namespace: envs.TestNamespace,
		Name:      resName.(string),
		RawData:   createManifest,
		Format:    action.ManifestFormat,
	}
	err = h.UpdateDeploy(ctx, &updateReq, &clusterRes.CommonResp{})
	assert.NotNil(t, err)

	// Get
	getReq := clusterRes.ResGetReq{
		ProjectID: envs.TestProjectID,
		ClusterID: envs.TestSharedClusterID,
		Namespace: envs.TestNamespace,
		Name:      resName.(string),
	}
	err = h.GetDeploy(ctx, &getReq, &clusterRes.CommonResp{})
	assert.NotNil(t, err)

	// Delete
	deleteReq := clusterRes.ResDeleteReq{
		ProjectID: envs.TestProjectID,
		ClusterID: envs.TestSharedClusterID,
		Namespace: envs.TestNamespace,
		Name:      resName.(string),
	}
	err = h.DeleteDeploy(ctx, &deleteReq, &clusterRes.CommonResp{})
	assert.NotNil(t, err)
}
