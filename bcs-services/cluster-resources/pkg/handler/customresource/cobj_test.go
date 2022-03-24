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

package customresource

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/envs"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/handler"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/example"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/pbstruct"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/stringx"
	clusterRes "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/proto/cluster-resources"
)

var cobjName4Test = "crontab-test-" + stringx.Rand(example.RandomSuffixLength, example.SuffixCharset)

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

func TestCObj(t *testing.T) {
	// 在集群中初始化 CRD
	err := handler.GetOrCreateCRD()
	assert.Nil(t, err)

	h := New()
	ctx := context.TODO()

	// Create
	createManifest, _ := pbstruct.Map2pbStruct(cobjManifest4Test)
	createReq := clusterRes.CObjCreateReq{
		ProjectID: envs.TestProjectID,
		ClusterID: envs.TestClusterID,
		CRDName:   handler.CRDName4Test,
		Manifest:  createManifest,
	}
	err = h.CreateCObj(ctx, &createReq, &clusterRes.CommonResp{})
	assert.Nil(t, err)

	// List
	listReq := clusterRes.CObjListReq{
		ProjectID: envs.TestProjectID,
		ClusterID: envs.TestClusterID,
		CRDName:   handler.CRDName4Test,
		Namespace: envs.TestNamespace,
	}
	listResp := clusterRes.CommonResp{}
	err = h.ListCObj(ctx, &listReq, &listResp)
	assert.Nil(t, err)

	respData := listResp.Data.AsMap()
	assert.Equal(t, "CronTabList", mapx.Get(respData, "manifest.kind", ""))

	// Update
	_ = mapx.SetItems(cobjManifest4Test, "spec.cronSpec", "* * * * */5")
	updateManifest, _ := pbstruct.Map2pbStruct(cobjManifest4Test)
	updateReq := clusterRes.CObjUpdateReq{
		ProjectID: envs.TestProjectID,
		ClusterID: envs.TestClusterID,
		CRDName:   handler.CRDName4Test,
		CobjName:  cobjName4Test,
		Namespace: envs.TestNamespace,
		Manifest:  updateManifest,
	}
	err = h.UpdateCObj(ctx, &updateReq, &clusterRes.CommonResp{})
	assert.Nil(t, err)

	// Get
	getReq := clusterRes.CObjGetReq{
		ProjectID: envs.TestProjectID,
		ClusterID: envs.TestClusterID,
		CRDName:   handler.CRDName4Test,
		CobjName:  cobjName4Test,
		Namespace: envs.TestNamespace,
	}
	getResp := clusterRes.CommonResp{}
	err = h.GetCObj(ctx, &getReq, &getResp)
	assert.Nil(t, err)

	respData = getResp.Data.AsMap()
	assert.Equal(t, "CronTab", mapx.Get(respData, "manifest.kind", ""))
	assert.Equal(t, "* * * * */5", mapx.Get(respData, "manifest.spec.cronSpec", ""))

	// Delete
	deleteReq := clusterRes.CObjDeleteReq{
		ProjectID: envs.TestProjectID,
		ClusterID: envs.TestClusterID,
		CRDName:   handler.CRDName4Test,
		CobjName:  cobjName4Test,
		Namespace: envs.TestNamespace,
	}
	err = h.DeleteCObj(ctx, &deleteReq, &clusterRes.CommonResp{})
	assert.Nil(t, err)
}

func TestCObjInSharedCluster(t *testing.T) {
	// 在集群中初始化 CRD
	err := handler.GetOrCreateCRD()
	assert.Nil(t, err)

	hdlr := New()

	listReq := clusterRes.CObjListReq{
		ProjectID: envs.TestProjectID,
		ClusterID: envs.TestSharedClusterID,
		CRDName:   handler.CRDName4Test,
		Namespace: envs.TestSharedClusterNS,
	}
	listResp := clusterRes.CommonResp{}
	err = hdlr.ListCObj(context.TODO(), &listReq, &listResp)
	// 新创建的 CRD 对应的 CObj 不被共享集群支持
	assert.NotNil(t, err)
}
