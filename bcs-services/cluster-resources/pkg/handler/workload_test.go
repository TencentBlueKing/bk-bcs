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
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util"
	clusterRes "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/proto/cluster-resources"
)

func TestDeploy(t *testing.T) {
	crh := NewClusterResourcesHandler()
	ctx := context.TODO()

	manifest, _ := example.LoadDemoManifest("workload/simple_deployment")
	resName := util.GetWithDefault(manifest, "metadata.name", "")

	// Create
	createManifest, _ := util.Map2pbStruct(manifest)
	createReq := genResCreateReq(createManifest)
	err := crh.CreateDeploy(ctx, &createReq, &clusterRes.CommonResp{})
	assert.Nil(t, err)

	// List
	listReq, listResp := genResListReq(), clusterRes.CommonResp{}
	err = crh.ListDeploy(ctx, &listReq, &listResp)
	assert.Nil(t, err)

	respData := listResp.Data.AsMap()
	assert.Equal(t, "DeploymentList", util.GetWithDefault(respData, "manifest.kind", ""))

	// Update
	_ = util.SetItems(manifest, "spec.replicas", 5)
	updateManifest, _ := util.Map2pbStruct(manifest)
	updateReq := genResUpdateReq(updateManifest, resName.(string))
	err = crh.UpdateDeploy(ctx, &updateReq, &clusterRes.CommonResp{})
	assert.Nil(t, err)

	// Get
	getReq, getResp := genResGetReq(resName.(string)), clusterRes.CommonResp{}
	err = crh.GetDeploy(ctx, &getReq, &getResp)
	assert.Nil(t, err)

	respData = getResp.Data.AsMap()
	assert.Equal(t, "Deployment", util.GetWithDefault(respData, "manifest.kind", ""))
	assert.Equal(t, float64(5), util.GetWithDefault(respData, "manifest.spec.replicas", 0))

	// Delete
	deleteReq := genResDeleteReq(resName.(string))
	err = crh.DeleteDeploy(ctx, &deleteReq, &clusterRes.CommonResp{})
	assert.Nil(t, err)
}

func TestDS(t *testing.T) {
	crh := NewClusterResourcesHandler()
	ctx := context.TODO()

	manifest, _ := example.LoadDemoManifest("workload/simple_daemonset")
	resName := util.GetWithDefault(manifest, "metadata.name", "")

	// Create
	createManifest, _ := util.Map2pbStruct(manifest)
	createReq := genResCreateReq(createManifest)
	err := crh.CreateDS(ctx, &createReq, &clusterRes.CommonResp{})
	assert.Nil(t, err)

	// List
	listReq, listResp := genResListReq(), clusterRes.CommonResp{}
	err = crh.ListDS(ctx, &listReq, &listResp)
	assert.Nil(t, err)

	respData := listResp.Data.AsMap()
	assert.Equal(t, "DaemonSetList", util.GetWithDefault(respData, "manifest.kind", ""))

	// Update
	_ = util.SetItems(manifest, "spec.template.metadata.labels.tKey", "tVal")
	updateManifest, _ := util.Map2pbStruct(manifest)
	updateReq := genResUpdateReq(updateManifest, resName.(string))
	err = crh.UpdateDS(ctx, &updateReq, &clusterRes.CommonResp{})
	assert.Nil(t, err)

	// Get
	getReq, getResp := genResGetReq(resName.(string)), clusterRes.CommonResp{}
	err = crh.GetDS(ctx, &getReq, &getResp)
	assert.Nil(t, err)

	respData = getResp.Data.AsMap()
	assert.Equal(t, "DaemonSet", util.GetWithDefault(respData, "manifest.kind", ""))
	assert.Equal(t, "tVal", util.GetWithDefault(respData, "manifest.spec.template.metadata.labels.tKey", ""))

	// Delete
	deleteReq := genResDeleteReq(resName.(string))
	err = crh.DeleteDS(ctx, &deleteReq, &clusterRes.CommonResp{})
	assert.Nil(t, err)
}

func TestSTS(t *testing.T) {
	crh := NewClusterResourcesHandler()
	ctx := context.TODO()

	manifest, _ := example.LoadDemoManifest("workload/simple_statefulset")
	resName := util.GetWithDefault(manifest, "metadata.name", "")

	// Create
	createManifest, _ := util.Map2pbStruct(manifest)
	createReq := genResCreateReq(createManifest)
	err := crh.CreateSTS(ctx, &createReq, &clusterRes.CommonResp{})
	assert.Nil(t, err)

	// List
	listReq, listResp := genResListReq(), clusterRes.CommonResp{}
	err = crh.ListSTS(ctx, &listReq, &listResp)
	assert.Nil(t, err)

	respData := listResp.Data.AsMap()
	assert.Equal(t, "StatefulSetList", util.GetWithDefault(respData, "manifest.kind", ""))

	// Update
	_ = util.SetItems(manifest, "spec.replicas", 3)
	updateManifest, _ := util.Map2pbStruct(manifest)
	updateReq := genResUpdateReq(updateManifest, resName.(string))
	err = crh.UpdateSTS(ctx, &updateReq, &clusterRes.CommonResp{})
	assert.Nil(t, err)

	// Get
	getReq, getResp := genResGetReq(resName.(string)), clusterRes.CommonResp{}
	err = crh.GetSTS(ctx, &getReq, &getResp)
	assert.Nil(t, err)

	respData = getResp.Data.AsMap()
	assert.Equal(t, "StatefulSet", util.GetWithDefault(respData, "manifest.kind", ""))
	assert.Equal(t, float64(3), util.GetWithDefault(respData, "manifest.spec.replicas", 0))

	// Delete
	deleteReq := genResDeleteReq(resName.(string))
	err = crh.DeleteSTS(ctx, &deleteReq, &clusterRes.CommonResp{})
	assert.Nil(t, err)
}

func TestCJ(t *testing.T) {
	crh := NewClusterResourcesHandler()
	ctx := context.TODO()

	manifest, _ := example.LoadDemoManifest("workload/simple_cronjob")
	resName := util.GetWithDefault(manifest, "metadata.name", "")

	// Create
	createManifest, _ := util.Map2pbStruct(manifest)
	createReq := genResCreateReq(createManifest)
	err := crh.CreateCJ(ctx, &createReq, &clusterRes.CommonResp{})
	assert.Nil(t, err)

	// List
	listReq, listResp := genResListReq(), clusterRes.CommonResp{}
	err = crh.ListCJ(ctx, &listReq, &listResp)
	assert.Nil(t, err)

	respData := listResp.Data.AsMap()
	assert.Equal(t, "CronJobList", util.GetWithDefault(respData, "manifest.kind", ""))

	// Update
	_ = util.SetItems(manifest, "spec.schedule", "*/5 * * * *")
	updateManifest, _ := util.Map2pbStruct(manifest)
	updateReq := genResUpdateReq(updateManifest, resName.(string))
	err = crh.UpdateCJ(ctx, &updateReq, &clusterRes.CommonResp{})
	assert.Nil(t, err)

	// Get
	getReq, getResp := genResGetReq(resName.(string)), clusterRes.CommonResp{}
	err = crh.GetCJ(ctx, &getReq, &getResp)
	assert.Nil(t, err)

	respData = getResp.Data.AsMap()
	assert.Equal(t, "CronJob", util.GetWithDefault(respData, "manifest.kind", ""))
	assert.Equal(t, "*/5 * * * *", util.GetWithDefault(respData, "manifest.spec.schedule", ""))

	// Delete
	deleteReq := genResDeleteReq(resName.(string))
	err = crh.DeleteCJ(ctx, &deleteReq, &clusterRes.CommonResp{})
	assert.Nil(t, err)
}

func TestJob(t *testing.T) {
	crh := NewClusterResourcesHandler()
	ctx := context.TODO()

	manifest, _ := example.LoadDemoManifest("workload/simple_job")
	resName := util.GetWithDefault(manifest, "metadata.name", "")

	// Create
	createManifest, _ := util.Map2pbStruct(manifest)
	createReq := genResCreateReq(createManifest)
	err := crh.CreateJob(ctx, &createReq, &clusterRes.CommonResp{})
	assert.Nil(t, err)

	// List
	listReq, listResp := genResListReq(), clusterRes.CommonResp{}
	err = crh.ListJob(ctx, &listReq, &listResp)
	assert.Nil(t, err)

	respData := listResp.Data.AsMap()
	assert.Equal(t, "JobList", util.GetWithDefault(respData, "manifest.kind", ""))

	// Get
	getReq, getResp := genResGetReq(resName.(string)), clusterRes.CommonResp{}
	err = crh.GetJob(ctx, &getReq, &getResp)
	assert.Nil(t, err)

	respData = getResp.Data.AsMap()
	assert.Equal(t, "Job", util.GetWithDefault(respData, "manifest.kind", ""))
	assert.Equal(t, float64(4), util.GetWithDefault(respData, "manifest.spec.backoffLimit", 0))

	// Delete
	deleteReq := genResDeleteReq(resName.(string))
	err = crh.DeleteJob(ctx, &deleteReq, &clusterRes.CommonResp{})
	assert.Nil(t, err)
}

func TestPod(t *testing.T) {
	crh := NewClusterResourcesHandler()
	ctx := context.TODO()

	manifest, _ := example.LoadDemoManifest("workload/simple_pod")
	resName := util.GetWithDefault(manifest, "metadata.name", "")

	// Create
	createManifest, _ := util.Map2pbStruct(manifest)
	createReq := genResCreateReq(createManifest)
	err := crh.CreatePo(ctx, &createReq, &clusterRes.CommonResp{})
	assert.Nil(t, err)

	// List
	podListReq := clusterRes.PodResListReq{
		ProjectID: envs.TestProjectID,
		ClusterID: envs.TestClusterID,
		Namespace: envs.TestNamespace,
	}
	listResp := clusterRes.CommonResp{}
	err = crh.ListPo(ctx, &podListReq, &listResp)
	assert.Nil(t, err)

	respData := listResp.Data.AsMap()
	assert.Equal(t, "PodList", util.GetWithDefault(respData, "manifest.kind", ""))

	// Get
	getReq, getResp := genResGetReq(resName.(string)), clusterRes.CommonResp{}
	err = crh.GetPo(ctx, &getReq, &getResp)
	assert.Nil(t, err)

	assert.Equal(t, "Pod", util.GetWithDefault(getResp.Data.AsMap(), "manifest.kind", ""))

	// ListPodPVC
	err = crh.ListPoPVC(ctx, &getReq, &getResp)
	assert.Nil(t, err)
	assert.Equal(t, "PersistentVolumeClaimList", util.GetWithDefault(getResp.Data.AsMap(), "manifest.kind", ""))

	// ListPodCM
	err = crh.ListPoCM(ctx, &getReq, &getResp)
	assert.Nil(t, err)
	assert.Equal(t, "ConfigMapList", util.GetWithDefault(getResp.Data.AsMap(), "manifest.kind", ""))

	// ListPodSecret
	err = crh.ListPoSecret(ctx, &getReq, &getResp)
	assert.Nil(t, err)
	assert.Equal(t, "SecretList", util.GetWithDefault(getResp.Data.AsMap(), "manifest.kind", ""))

	// Delete
	deleteReq := genResDeleteReq(resName.(string))
	err = crh.DeletePo(ctx, &deleteReq, &clusterRes.CommonResp{})
	assert.Nil(t, err)
}

func TestContainer(t *testing.T) {
	crh := NewClusterResourcesHandler()
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
	err := crh.ListContainer(ctx, &listReq, &listResp)
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
	err = crh.GetContainer(ctx, &getReq, &getResp)
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
	err = crh.GetContainerEnvInfo(ctx, &getReq, &getEnvResp)
	assert.Nil(t, err)
}

// 取集群 default 命名空间中已存在的，状态为 Running 的 Pod 用于测试（需确保 Pod 存在）
func getRunningPodNameFromCluster() string {
	podCli := client.NewPodCliByClusterID(envs.TestClusterID)
	ret, _ := podCli.List("default", "", "", metav1.ListOptions{})

	for _, po := range ret["items"].([]interface{}) {
		po, _ := po.(map[string]interface{})
		parser := formatter.PodStatusParser{Manifest: po}
		if parser.Parse() == "Running" {
			return util.GetWithDefault(po, "metadata.name", "").(string)
		}
	}
	return ""
}
