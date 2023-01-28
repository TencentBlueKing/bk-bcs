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
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/action"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/envs"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/handler"
	resCsts "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/constants"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/example"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/pbstruct"
	clusterRes "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/proto/cluster-resources"
)

func TestPod(t *testing.T) {
	h := New()
	ctx := handler.NewInjectedContext("", "", "")

	manifest, _ := example.LoadDemoManifest(ctx, "workload/simple_pod", "", "", resCsts.Po)
	resName := mapx.GetStr(manifest, "metadata.name")

	// Create
	createManifest, _ := pbstruct.Map2pbStruct(manifest)
	createReq := handler.GenResCreateReq(createManifest)
	err := h.CreatePo(ctx, &createReq, &clusterRes.CommonResp{})
	assert.Nil(t, err)

	// List
	podListReq := clusterRes.ResListReq{
		ProjectID: envs.TestProjectID,
		ClusterID: envs.TestClusterID,
		Namespace: envs.TestNamespace,
		Format:    action.ManifestFormat,
	}
	listResp := clusterRes.CommonResp{}
	err = h.ListPo(ctx, &podListReq, &listResp)
	assert.Nil(t, err)

	respData := listResp.Data.AsMap()
	assert.Equal(t, "PodList", mapx.GetStr(respData, "manifest.kind"))

	// Get
	getReq, getResp := handler.GenResGetReq(resName), clusterRes.CommonResp{}
	err = h.GetPo(ctx, &getReq, &getResp)
	assert.Nil(t, err)

	assert.Equal(t, "Pod", mapx.GetStr(getResp.Data.AsMap(), "manifest.kind"))

	// ListPodPVC
	err = h.ListPoPVC(ctx, &getReq, &getResp)
	assert.Nil(t, err)
	assert.Equal(t, "PersistentVolumeClaimList", mapx.GetStr(getResp.Data.AsMap(), "manifest.kind"))

	// ListPodCM
	err = h.ListPoCM(ctx, &getReq, &getResp)
	assert.Nil(t, err)
	assert.Equal(t, "ConfigMapList", mapx.GetStr(getResp.Data.AsMap(), "manifest.kind"))

	// ListPodSecret
	err = h.ListPoSecret(ctx, &getReq, &getResp)
	assert.Nil(t, err)
	assert.Equal(t, "SecretList", mapx.GetStr(getResp.Data.AsMap(), "manifest.kind"))

	// Delete
	deleteReq := handler.GenResDeleteReq(resName)
	err = h.DeletePo(ctx, &deleteReq, &clusterRes.CommonResp{})
	assert.Nil(t, err)
}
