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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/envs"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/client"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/example"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/formatter"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/pbstruct"
	clusterRes "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/proto/cluster-resources"
)

func TestDeploy(t *testing.T) {
	h := NewClusterResourcesHandler()
	ctx := context.TODO()

	manifest, _ := example.LoadDemoManifest("workload/simple_deployment")
	resName := mapx.Get(manifest, "metadata.name", "")

	// Create
	createManifest, _ := pbstruct.Map2pbStruct(manifest)
	createReq := genResCreateReq(createManifest)
	err := h.CreateDeploy(ctx, &createReq, &clusterRes.CommonResp{})
	assert.Nil(t, err)

	// List
	listReq, listResp := genResListReq(), clusterRes.CommonResp{}
	err = h.ListDeploy(ctx, &listReq, &listResp)
	assert.Nil(t, err)

	respData := listResp.Data.AsMap()
	assert.Equal(t, "DeploymentList", mapx.Get(respData, "manifest.kind", ""))

	// Update
	_ = mapx.SetItems(manifest, "spec.replicas", 5)
	updateManifest, _ := pbstruct.Map2pbStruct(manifest)
	updateReq := genResUpdateReq(updateManifest, resName.(string))
	err = h.UpdateDeploy(ctx, &updateReq, &clusterRes.CommonResp{})
	assert.Nil(t, err)

	// Get
	getReq, getResp := genResGetReq(resName.(string)), clusterRes.CommonResp{}
	err = h.GetDeploy(ctx, &getReq, &getResp)
	assert.Nil(t, err)

	respData = getResp.Data.AsMap()
	assert.Equal(t, "Deployment", mapx.Get(respData, "manifest.kind", ""))
	assert.Equal(t, float64(5), mapx.Get(respData, "manifest.spec.replicas", 0))

	// Delete
	deleteReq := genResDeleteReq(resName.(string))
	err = h.DeleteDeploy(ctx, &deleteReq, &clusterRes.CommonResp{})
	assert.Nil(t, err)
}

func TestDeployInSharedCluster(t *testing.T) {
	// 在共享集群中新建命名空间
	err := getOrCreateNS(envs.TestSharedClusterNS)
	assert.Nil(t, err)

	h := NewClusterResourcesHandler()
	ctx := context.TODO()

	manifest, _ := example.LoadDemoManifest("workload/simple_deployment")
	resName := mapx.Get(manifest, "metadata.name", "")
	// 设置为共享集群项目属命名空间
	err = mapx.SetItems(manifest, "metadata.namespace", envs.TestSharedClusterNS)
	assert.Nil(t, err)

	// Create
	createManifest, _ := pbstruct.Map2pbStruct(manifest)
	createReq := clusterRes.ResCreateReq{
		ProjectID: envs.TestProjectID,
		ClusterID: envs.TestSharedClusterID,
		Manifest:  createManifest,
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
		Manifest:  createManifest,
	}
	err = h.UpdateDeploy(ctx, &updateReq, &clusterRes.CommonResp{})
	assert.Nil(t, err)

	// Get
	getReq := clusterRes.ResGetReq{
		ProjectID: envs.TestProjectID,
		ClusterID: envs.TestSharedClusterID,
		Namespace: envs.TestSharedClusterNS,
		Name:      resName.(string),
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

func TestDeployInSharedClusterNotPerm(t *testing.T) {
	h := NewClusterResourcesHandler()
	ctx := context.TODO()

	manifest, _ := example.LoadDemoManifest("workload/simple_deployment")
	resName := mapx.Get(manifest, "metadata.name", "")

	// Create
	createManifest, _ := pbstruct.Map2pbStruct(manifest)
	createReq := clusterRes.ResCreateReq{
		ProjectID: envs.TestProjectID,
		ClusterID: envs.TestSharedClusterID,
		Manifest:  createManifest,
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
		Manifest:  createManifest,
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

func TestDS(t *testing.T) {
	h := NewClusterResourcesHandler()
	ctx := context.TODO()

	manifest, _ := example.LoadDemoManifest("workload/simple_daemonset")
	resName := mapx.Get(manifest, "metadata.name", "")

	// Create
	createManifest, _ := pbstruct.Map2pbStruct(manifest)
	createReq := genResCreateReq(createManifest)
	err := h.CreateDS(ctx, &createReq, &clusterRes.CommonResp{})
	assert.Nil(t, err)

	// List
	listReq, listResp := genResListReq(), clusterRes.CommonResp{}
	err = h.ListDS(ctx, &listReq, &listResp)
	assert.Nil(t, err)

	respData := listResp.Data.AsMap()
	assert.Equal(t, "DaemonSetList", mapx.Get(respData, "manifest.kind", ""))

	// Update
	_ = mapx.SetItems(manifest, "spec.template.metadata.labels.tKey", "tVal")
	updateManifest, _ := pbstruct.Map2pbStruct(manifest)
	updateReq := genResUpdateReq(updateManifest, resName.(string))
	err = h.UpdateDS(ctx, &updateReq, &clusterRes.CommonResp{})
	assert.Nil(t, err)

	// Get
	getReq, getResp := genResGetReq(resName.(string)), clusterRes.CommonResp{}
	err = h.GetDS(ctx, &getReq, &getResp)
	assert.Nil(t, err)

	respData = getResp.Data.AsMap()
	assert.Equal(t, "DaemonSet", mapx.Get(respData, "manifest.kind", ""))
	assert.Equal(t, "tVal", mapx.Get(respData, "manifest.spec.template.metadata.labels.tKey", ""))

	// Delete
	deleteReq := genResDeleteReq(resName.(string))
	err = h.DeleteDS(ctx, &deleteReq, &clusterRes.CommonResp{})
	assert.Nil(t, err)
}

func TestSTS(t *testing.T) {
	h := NewClusterResourcesHandler()
	ctx := context.TODO()

	manifest, _ := example.LoadDemoManifest("workload/simple_statefulset")
	resName := mapx.Get(manifest, "metadata.name", "")

	// Create
	createManifest, _ := pbstruct.Map2pbStruct(manifest)
	createReq := genResCreateReq(createManifest)
	err := h.CreateSTS(ctx, &createReq, &clusterRes.CommonResp{})
	assert.Nil(t, err)

	// List
	listReq, listResp := genResListReq(), clusterRes.CommonResp{}
	err = h.ListSTS(ctx, &listReq, &listResp)
	assert.Nil(t, err)

	respData := listResp.Data.AsMap()
	assert.Equal(t, "StatefulSetList", mapx.Get(respData, "manifest.kind", ""))

	// Update
	_ = mapx.SetItems(manifest, "spec.replicas", 3)
	updateManifest, _ := pbstruct.Map2pbStruct(manifest)
	updateReq := genResUpdateReq(updateManifest, resName.(string))
	err = h.UpdateSTS(ctx, &updateReq, &clusterRes.CommonResp{})
	assert.Nil(t, err)

	// Get
	getReq, getResp := genResGetReq(resName.(string)), clusterRes.CommonResp{}
	err = h.GetSTS(ctx, &getReq, &getResp)
	assert.Nil(t, err)

	respData = getResp.Data.AsMap()
	assert.Equal(t, "StatefulSet", mapx.Get(respData, "manifest.kind", ""))
	assert.Equal(t, float64(3), mapx.Get(respData, "manifest.spec.replicas", 0))

	// Delete
	deleteReq := genResDeleteReq(resName.(string))
	err = h.DeleteSTS(ctx, &deleteReq, &clusterRes.CommonResp{})
	assert.Nil(t, err)
}

func TestCJ(t *testing.T) {
	h := NewClusterResourcesHandler()
	ctx := context.TODO()

	manifest, _ := example.LoadDemoManifest("workload/simple_cronjob")
	resName := mapx.Get(manifest, "metadata.name", "")

	// Create
	createManifest, _ := pbstruct.Map2pbStruct(manifest)
	createReq := genResCreateReq(createManifest)
	err := h.CreateCJ(ctx, &createReq, &clusterRes.CommonResp{})
	assert.Nil(t, err)

	// List
	listReq, listResp := genResListReq(), clusterRes.CommonResp{}
	err = h.ListCJ(ctx, &listReq, &listResp)
	assert.Nil(t, err)

	respData := listResp.Data.AsMap()
	assert.Equal(t, "CronJobList", mapx.Get(respData, "manifest.kind", ""))

	// Update
	_ = mapx.SetItems(manifest, "spec.schedule", "*/5 * * * *")
	updateManifest, _ := pbstruct.Map2pbStruct(manifest)
	updateReq := genResUpdateReq(updateManifest, resName.(string))
	err = h.UpdateCJ(ctx, &updateReq, &clusterRes.CommonResp{})
	assert.Nil(t, err)

	// Get
	getReq, getResp := genResGetReq(resName.(string)), clusterRes.CommonResp{}
	err = h.GetCJ(ctx, &getReq, &getResp)
	assert.Nil(t, err)

	respData = getResp.Data.AsMap()
	assert.Equal(t, "CronJob", mapx.Get(respData, "manifest.kind", ""))
	assert.Equal(t, "*/5 * * * *", mapx.Get(respData, "manifest.spec.schedule", ""))

	// Delete
	deleteReq := genResDeleteReq(resName.(string))
	err = h.DeleteCJ(ctx, &deleteReq, &clusterRes.CommonResp{})
	assert.Nil(t, err)
}

func TestJob(t *testing.T) {
	h := NewClusterResourcesHandler()
	ctx := context.TODO()

	manifest, _ := example.LoadDemoManifest("workload/simple_job")
	resName := mapx.Get(manifest, "metadata.name", "")

	// Create
	createManifest, _ := pbstruct.Map2pbStruct(manifest)
	createReq := genResCreateReq(createManifest)
	err := h.CreateJob(ctx, &createReq, &clusterRes.CommonResp{})
	assert.Nil(t, err)

	// List
	listReq, listResp := genResListReq(), clusterRes.CommonResp{}
	err = h.ListJob(ctx, &listReq, &listResp)
	assert.Nil(t, err)

	respData := listResp.Data.AsMap()
	assert.Equal(t, "JobList", mapx.Get(respData, "manifest.kind", ""))

	// Get
	getReq, getResp := genResGetReq(resName.(string)), clusterRes.CommonResp{}
	err = h.GetJob(ctx, &getReq, &getResp)
	assert.Nil(t, err)

	respData = getResp.Data.AsMap()
	assert.Equal(t, "Job", mapx.Get(respData, "manifest.kind", ""))
	assert.Equal(t, float64(4), mapx.Get(respData, "manifest.spec.backoffLimit", 0))

	// Delete
	deleteReq := genResDeleteReq(resName.(string))
	err = h.DeleteJob(ctx, &deleteReq, &clusterRes.CommonResp{})
	assert.Nil(t, err)
}

func TestPod(t *testing.T) {
	h := NewClusterResourcesHandler()
	ctx := context.TODO()

	manifest, _ := example.LoadDemoManifest("workload/simple_pod")
	resName := mapx.Get(manifest, "metadata.name", "")

	// Create
	createManifest, _ := pbstruct.Map2pbStruct(manifest)
	createReq := genResCreateReq(createManifest)
	err := h.CreatePo(ctx, &createReq, &clusterRes.CommonResp{})
	assert.Nil(t, err)

	// List
	podListReq := clusterRes.PodResListReq{
		ProjectID: envs.TestProjectID,
		ClusterID: envs.TestClusterID,
		Namespace: envs.TestNamespace,
	}
	listResp := clusterRes.CommonResp{}
	err = h.ListPo(ctx, &podListReq, &listResp)
	assert.Nil(t, err)

	respData := listResp.Data.AsMap()
	assert.Equal(t, "PodList", mapx.Get(respData, "manifest.kind", ""))

	// Get
	getReq, getResp := genResGetReq(resName.(string)), clusterRes.CommonResp{}
	err = h.GetPo(ctx, &getReq, &getResp)
	assert.Nil(t, err)

	assert.Equal(t, "Pod", mapx.Get(getResp.Data.AsMap(), "manifest.kind", ""))

	// ListPodPVC
	err = h.ListPoPVC(ctx, &getReq, &getResp)
	assert.Nil(t, err)
	assert.Equal(t, "PersistentVolumeClaimList", mapx.Get(getResp.Data.AsMap(), "manifest.kind", ""))

	// ListPodCM
	err = h.ListPoCM(ctx, &getReq, &getResp)
	assert.Nil(t, err)
	assert.Equal(t, "ConfigMapList", mapx.Get(getResp.Data.AsMap(), "manifest.kind", ""))

	// ListPodSecret
	err = h.ListPoSecret(ctx, &getReq, &getResp)
	assert.Nil(t, err)
	assert.Equal(t, "SecretList", mapx.Get(getResp.Data.AsMap(), "manifest.kind", ""))

	// Delete
	deleteReq := genResDeleteReq(resName.(string))
	err = h.DeletePo(ctx, &deleteReq, &clusterRes.CommonResp{})
	assert.Nil(t, err)
}

func TestContainer(t *testing.T) {
	h := NewClusterResourcesHandler()
	ctx := context.TODO()

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

// 取集群 kube-system 命名空间中已存在的，状态为 Running 的 Pod 用于测试（需确保 Pod 存在）
func getRunningPodNameFromCluster() string {
	podCli := client.NewPodCliByClusterID(envs.TestClusterID)
	ret, _ := podCli.List(envs.TestNamespace, "", "", metav1.ListOptions{})

	for _, pod := range ret["items"].([]interface{}) {
		p, _ := pod.(map[string]interface{})
		parser := formatter.PodStatusParser{Manifest: p}
		if parser.Parse() == "Running" {
			return mapx.Get(p, "metadata.name", "").(string)
		}
	}
	return ""
}
