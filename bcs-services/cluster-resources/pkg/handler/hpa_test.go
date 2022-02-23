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
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util"
	clusterRes "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/proto/cluster-resources"
)

func TestHPA(t *testing.T) {
	crh := NewClusterResourcesHandler()
	ctx := context.TODO()

	manifest, _ := example.LoadDemoManifest("hpa/simple_hpa")
	resName := util.GetWithDefault(manifest, "metadata.name", "")

	// Create
	createManifest, _ := util.Map2pbStruct(manifest)
	createReq := genResCreateReq(createManifest)
	err := crh.CreateHPA(ctx, &createReq, &clusterRes.CommonResp{})
	assert.Nil(t, err)

	// List
	listReq, listResp := genResListReq(), clusterRes.CommonResp{}
	err = crh.ListHPA(ctx, &listReq, &listResp)
	assert.Nil(t, err)

	respData := listResp.Data.AsMap()
	assert.Equal(t, "HorizontalPodAutoscalerList", util.GetWithDefault(respData, "manifest.kind", ""))

	// Update
	_ = util.SetItems(manifest, "spec.minReplicas", 2)
	updateManifest, _ := util.Map2pbStruct(manifest)
	updateReq := genResUpdateReq(updateManifest, resName.(string))
	err = crh.UpdateHPA(ctx, &updateReq, &clusterRes.CommonResp{})
	assert.Nil(t, err)

	// Get
	getReq, getResp := genResGetReq(resName.(string)), clusterRes.CommonResp{}
	err = crh.GetHPA(ctx, &getReq, &getResp)
	assert.Nil(t, err)

	respData = getResp.Data.AsMap()
	assert.Equal(t, "HorizontalPodAutoscaler", util.GetWithDefault(respData, "manifest.kind", ""))
	assert.Equal(t, float64(2), util.GetWithDefault(respData, "manifest.spec.minReplicas", 0))

	// Delete
	deleteReq := genResDeleteReq(resName.(string))
	err = crh.DeleteHPA(ctx, &deleteReq, &clusterRes.CommonResp{})
	assert.Nil(t, err)
}
