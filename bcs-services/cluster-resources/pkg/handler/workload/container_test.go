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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/envs"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/handler"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/client"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/formatter"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"
	clusterRes "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/proto/cluster-resources"
)

func TestContainer(t *testing.T) {
	h := New()
	ctx := handler.NewInjectedContext("", "", "")

	podName := getRunningPodNameFromCluster()
	assert.NotEqual(t, "", podName, "ensure has running pod")

	// List
	listReq := clusterRes.ContainerListReq{
		ProjectID: envs.TestProjectID,
		ClusterID: envs.TestClusterID,
		Namespace: envs.TestNamespace,
		PodName:   podName,
	}
	listResp := clusterRes.CommonListResp{}
	err := h.ListContainer(ctx, &listReq, &listResp)
	assert.Nil(t, err)

	listRespData := listResp.Data.AsSlice()
	for _, k := range []string{"containerID", "image", "name", "status", "message", "reason"} {
		_, exists := listRespData[0].(map[string]interface{})[k]
		assert.True(t, exists)
	}
	// 根据查询结果，找到第一个容器名，后续单测用
	containerName := listRespData[0].(map[string]interface{})["name"].(string)

	// Get
	getReq := clusterRes.ContainerGetReq{
		ProjectID:     envs.TestProjectID,
		ClusterID:     envs.TestClusterID,
		Namespace:     envs.TestNamespace,
		PodName:       podName,
		ContainerName: containerName,
	}
	getResp := clusterRes.CommonResp{}
	err = h.GetContainer(ctx, &getReq, &getResp)
	assert.Nil(t, err)

	getRespData := getResp.Data.AsMap()
	for _, k := range []string{
		"hostName", "hostIP", "containerIP", "containerID", "containerName",
		"image", "networkMode", "ports", "volumes", "resources", "command",
	} {
		_, exists := getRespData[k]
		assert.True(t, exists)
	}

	// Get EnvInfo
	getEnvResp := clusterRes.CommonListResp{}
	err = h.GetContainerEnvInfo(ctx, &getReq, &getEnvResp)
	assert.Nil(t, err)
}

// 取集群中已存在的，状态为 Running 的 Pod 用于测试（需确保 Pod 存在）
func getRunningPodNameFromCluster() string {
	ctx := handler.NewInjectedContext("", "", "")
	podCli := client.NewPodCliByClusterID(ctx, envs.TestClusterID)
	ret, _ := podCli.List(ctx, envs.TestNamespace, "", "", metav1.ListOptions{})

	for _, pod := range ret["items"].([]interface{}) {
		p, _ := pod.(map[string]interface{})
		parser := formatter.PodStatusParser{Manifest: p}
		if parser.Parse() == "Running" {
			return mapx.GetStr(p, "metadata.name")
		}
	}
	return ""
}
