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

package namespace

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/action"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/envs"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/handler"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"
	clusterRes "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/proto/cluster-resources"
)

func TestNS(t *testing.T) {
	h := New()
	ctx := handler.NewInjectedContext("", "", "")

	// List
	listReq, listResp := handler.GenResListReq(), clusterRes.CommonResp{}
	err := h.ListNS(ctx, &listReq, &listResp)
	assert.Nil(t, err)

	respData := listResp.Data.AsMap()
	assert.Equal(t, "NamespaceList", mapx.Get(respData, "manifest.kind", ""))
}

func TestNSInSharedCluster(t *testing.T) {
	// 初始化共享集群中的项目属命名空间
	err := handler.GetOrCreateNS(envs.TestSharedClusterNS)
	assert.Nil(t, err)

	h := New()
	ctx := handler.NewInjectedContext("", "", envs.TestSharedClusterID)

	listReq := clusterRes.ResListReq{
		ProjectID: envs.TestProjectID,
		ClusterID: envs.TestSharedClusterID,
		Format:    action.ManifestFormat,
	}
	listResp := clusterRes.CommonResp{}
	err = h.ListNS(ctx, &listReq, &listResp)
	assert.Nil(t, err)

	// 确保列出来的，都是共享集群中，属于项目的命名空间
	respData := listResp.Data.AsMap()
	for _, ns := range respData["manifest"].(map[string]interface{})["items"].([]interface{}) {
		name := mapx.Get(ns.(map[string]interface{}), "metadata.name", "")
		assert.True(t, strings.Contains(name.(string), envs.TestProjectCode))
	}
}
