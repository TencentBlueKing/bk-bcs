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

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/example"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/pbstruct"
	clusterRes "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/proto/cluster-resources"
)

func TestPV(t *testing.T) {
	h := NewClusterResourcesHandler()
	ctx := context.TODO()

	manifest, _ := example.LoadDemoManifest("storage/simple_persistent_volume")
	resName := mapx.Get(manifest, "metadata.name", "")

	// Create
	createManifest, _ := pbstruct.Map2pbStruct(manifest)
	createReq := genResCreateReq(createManifest)
	err := h.CreatePV(ctx, &createReq, &clusterRes.CommonResp{})
	assert.Nil(t, err)

	// List
	listReq, listResp := genResListReq(), clusterRes.CommonResp{}
	err = h.ListPV(ctx, &listReq, &listResp)
	assert.Nil(t, err)

	respData := listResp.Data.AsMap()
	assert.Equal(t, "PersistentVolumeList", mapx.Get(respData, "manifest.kind", ""))

	// Update
	_ = mapx.SetItems(manifest, "spec.capacity.storage", "2Gi")
	updateManifest, _ := pbstruct.Map2pbStruct(manifest)
	updateReq := genResUpdateReq(updateManifest, resName.(string))
	err = h.UpdatePV(ctx, &updateReq, &clusterRes.CommonResp{})
	assert.Nil(t, err)

	// Get
	getReq, getResp := genResGetReq(resName.(string)), clusterRes.CommonResp{}
	err = h.GetPV(ctx, &getReq, &getResp)
	assert.Nil(t, err)

	respData = getResp.Data.AsMap()
	assert.Equal(t, "PersistentVolume", mapx.Get(respData, "manifest.kind", ""))
	assert.Equal(t, "2Gi", mapx.Get(respData, "manifest.spec.capacity.storage", 0))

	// Delete
	deleteReq := genResDeleteReq(resName.(string))
	err = h.DeletePV(ctx, &deleteReq, &clusterRes.CommonResp{})
	assert.Nil(t, err)
}

func TestPVC(t *testing.T) {
	h := NewClusterResourcesHandler()
	ctx := context.TODO()

	manifest, _ := example.LoadDemoManifest("storage/simple_persistent_volume_claim")
	resName := mapx.Get(manifest, "metadata.name", "")

	// Create
	createManifest, _ := pbstruct.Map2pbStruct(manifest)
	createReq := genResCreateReq(createManifest)
	err := h.CreatePVC(ctx, &createReq, &clusterRes.CommonResp{})
	assert.Nil(t, err)

	// List
	listReq, listResp := genResListReq(), clusterRes.CommonResp{}
	err = h.ListPVC(ctx, &listReq, &listResp)
	assert.Nil(t, err)

	respData := listResp.Data.AsMap()
	assert.Equal(t, "PersistentVolumeClaimList", mapx.Get(respData, "manifest.kind", ""))

	// Update
	_ = mapx.SetItems(manifest, "metadata.annotations", map[string]interface{}{"tKey": "tVal"})
	updateManifest, _ := pbstruct.Map2pbStruct(manifest)
	updateReq := genResUpdateReq(updateManifest, resName.(string))
	err = h.UpdatePVC(ctx, &updateReq, &clusterRes.CommonResp{})
	assert.Nil(t, err)

	// Get
	getReq, getResp := genResGetReq(resName.(string)), clusterRes.CommonResp{}
	err = h.GetPVC(ctx, &getReq, &getResp)
	assert.Nil(t, err)

	respData = getResp.Data.AsMap()
	assert.Equal(t, "PersistentVolumeClaim", mapx.Get(respData, "manifest.kind", ""))
	assert.Equal(t, "tVal", mapx.Get(respData, "manifest.metadata.annotations.tKey", ""))

	// Delete
	deleteReq := genResDeleteReq(resName.(string))
	err = h.DeletePVC(ctx, &deleteReq, &clusterRes.CommonResp{})
	assert.Nil(t, err)
}

func TestSC(t *testing.T) {
	h := NewClusterResourcesHandler()
	ctx := context.TODO()

	manifest, _ := example.LoadDemoManifest("storage/simple_storage_class")
	resName := mapx.Get(manifest, "metadata.name", "")

	// Create
	createManifest, _ := pbstruct.Map2pbStruct(manifest)
	createReq := genResCreateReq(createManifest)
	err := h.CreateSC(ctx, &createReq, &clusterRes.CommonResp{})
	assert.Nil(t, err)

	// List
	listReq, listResp := genResListReq(), clusterRes.CommonResp{}
	err = h.ListSC(ctx, &listReq, &listResp)
	assert.Nil(t, err)

	respData := listResp.Data.AsMap()
	assert.Equal(t, "StorageClassList", mapx.Get(respData, "manifest.kind", ""))

	// Update
	_ = mapx.SetItems(manifest, "metadata.annotations", map[string]interface{}{"tKey": "tVal"})
	updateManifest, _ := pbstruct.Map2pbStruct(manifest)
	updateReq := genResUpdateReq(updateManifest, resName.(string))
	err = h.UpdateSC(ctx, &updateReq, &clusterRes.CommonResp{})
	assert.Nil(t, err)

	// Get
	getReq, getResp := genResGetReq(resName.(string)), clusterRes.CommonResp{}
	err = h.GetSC(ctx, &getReq, &getResp)
	assert.Nil(t, err)

	respData = getResp.Data.AsMap()
	assert.Equal(t, "StorageClass", mapx.Get(respData, "manifest.kind", ""))
	assert.Equal(t, "tVal", mapx.Get(respData, "manifest.metadata.annotations.tKey", ""))

	// Delete
	deleteReq := genResDeleteReq(resName.(string))
	err = h.DeleteSC(ctx, &deleteReq, &clusterRes.CommonResp{})
	assert.Nil(t, err)
}
