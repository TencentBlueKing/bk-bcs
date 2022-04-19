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

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/handler"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/example"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/pbstruct"
	clusterRes "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/proto/cluster-resources"
)

func TestSTS(t *testing.T) {
	h := New()
	ctx := context.TODO()

	manifest, _ := example.LoadDemoManifest("workload/simple_statefulset")
	resName := mapx.Get(manifest, "metadata.name", "")

	// Create
	createManifest, _ := pbstruct.Map2pbStruct(manifest)
	createReq := handler.GenResCreateReq(createManifest)
	err := h.CreateSTS(ctx, &createReq, &clusterRes.CommonResp{})
	assert.Nil(t, err)

	// List
	listReq, listResp := handler.GenResListReq(), clusterRes.CommonResp{}
	err = h.ListSTS(ctx, &listReq, &listResp)
	assert.Nil(t, err)

	respData := listResp.Data.AsMap()
	assert.Equal(t, "StatefulSetList", mapx.Get(respData, "manifest.kind", ""))

	// Update
	_ = mapx.SetItems(manifest, "spec.replicas", 3)
	updateManifest, _ := pbstruct.Map2pbStruct(manifest)
	updateReq := handler.GenResUpdateReq(updateManifest, resName.(string))
	err = h.UpdateSTS(ctx, &updateReq, &clusterRes.CommonResp{})
	assert.Nil(t, err)

	// Get
	getReq, getResp := handler.GenResGetReq(resName.(string)), clusterRes.CommonResp{}
	err = h.GetSTS(ctx, &getReq, &getResp)
	assert.Nil(t, err)

	respData = getResp.Data.AsMap()
	assert.Equal(t, "StatefulSet", mapx.Get(respData, "manifest.kind", ""))
	assert.Equal(t, float64(3), mapx.Get(respData, "manifest.spec.replicas", 0))

	// Delete
	deleteReq := handler.GenResDeleteReq(resName.(string))
	err = h.DeleteSTS(ctx, &deleteReq, &clusterRes.CommonResp{})
	assert.Nil(t, err)
}
