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
	conf "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/config"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/handler"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/slice"
	clusterRes "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/proto/cluster-resources"
)

func TestCRD(t *testing.T) {
	// 在集群中初始化 CRD
	err := handler.GetOrCreateCRD()
	assert.Nil(t, err)

	h := New()
	ctx := context.TODO()

	// List
	listReq, listResp := handler.GenResListReq(), clusterRes.CommonResp{}
	err = h.ListCRD(ctx, &listReq, &listResp)
	assert.Nil(t, err)

	respData := listResp.Data.AsMap()
	assert.Equal(t, "CustomResourceDefinitionList", mapx.Get(respData, "manifest.kind", ""))

	// Get
	getReq, getResp := handler.GenResGetReq(handler.CRDName4Test), clusterRes.CommonResp{}
	err = h.GetCRD(ctx, &getReq, &getResp)
	assert.Nil(t, err)

	respData = getResp.Data.AsMap()
	assert.Equal(t, "CustomResourceDefinition", mapx.Get(respData, "manifest.kind", ""))
	assert.Equal(t, "Namespaced", mapx.Get(respData, "manifest.spec.scope", ""))
}

func TestCRDInSharedCluster(t *testing.T) {
	// 在集群中初始化 CRD
	err := handler.GetOrCreateCRD()
	assert.Nil(t, err)

	h := New()

	listReq := clusterRes.ResListReq{
		ProjectID: envs.TestProjectID,
		ClusterID: envs.TestSharedClusterID,
	}
	listResp := clusterRes.CommonResp{}
	err = h.ListCRD(context.TODO(), &listReq, &listResp)
	assert.Nil(t, err)

	// 确保共享集群中查出的 CRD 都是共享集群允许的
	respData := listResp.Data.AsMap()
	for _, crdInfo := range respData["manifestExt"].(map[string]interface{}) {
		assert.True(t, slice.StringInSlice(crdInfo.(map[string]interface{})["name"].(string), conf.G.SharedCluster.EnabledCRDs))
	}
}
