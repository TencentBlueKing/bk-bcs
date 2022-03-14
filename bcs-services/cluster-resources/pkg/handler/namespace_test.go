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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/envs"
	cli "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/client"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"
	clusterRes "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/proto/cluster-resources"
)

func TestNS(t *testing.T) {
	h := NewClusterResourcesHandler()

	// List
	listReq, listResp := genResListReq(), clusterRes.CommonResp{}
	err := h.ListNS(context.TODO(), &listReq, &listResp)
	assert.Nil(t, err)

	respData := listResp.Data.AsMap()
	assert.Equal(t, "NamespaceList", mapx.Get(respData, "manifest.kind", ""))
}

var nsManifest4Test = map[string]interface{}{
	"apiVersion": "v1",
	"kind":       "Namespace",
	"metadata": map[string]interface{}{
		"name":        "new_name_required",
		"annotations": map[string]interface{}{},
	},
}

// 在集群中初始化命名空间用于单元测试用
func getOrCreateNS(namespace string) error {
	if namespace == "" {
		namespace = envs.TestNamespace
	}
	nsCli := cli.NewNSCliByClusterID(envs.TestClusterID)
	_, err := nsCli.Get("", namespace, metav1.GetOptions{})
	if err != nil {
		_ = mapx.SetItems(nsManifest4Test, "metadata.name", namespace)
		if namespace == envs.TestSharedClusterNS {
			_ = mapx.SetItems(nsManifest4Test, []string{"metadata", "annotations", cli.ProjCodeAnnoKey}, envs.TestProjectCode)
		}
		_, err = nsCli.Create(nsManifest4Test, false, metav1.CreateOptions{})
	}
	return err
}

func TestNSInSharedCluster(t *testing.T) {
	// 初始化共享集群中的项目属命名空间
	err := getOrCreateNS(envs.TestSharedClusterNS)
	assert.Nil(t, err)

	h := NewClusterResourcesHandler()

	listReq := clusterRes.ResListReq{
		ProjectID: envs.TestProjectID,
		ClusterID: envs.TestSharedClusterID,
	}
	listResp := clusterRes.CommonResp{}
	err = h.ListNS(context.TODO(), &listReq, &listResp)
	assert.Nil(t, err)

	// 确保列出来的，都是共享集群中，属于项目的命名空间
	respData := listResp.Data.AsMap()
	for _, ns := range respData["manifest"].(map[string]interface{})["items"].([]interface{}) {
		name := mapx.Get(ns.(map[string]interface{}), "metadata.name", "")
		assert.True(t, strings.HasPrefix(name.(string), envs.TestProjectCode))
	}
}
