/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package api

import (
	"fmt"
	"testing"
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/regions"
)

func getClient(region string) *TkeClient {
	cli, _ := NewTkeClient(&cloudprovider.CommonOption{
		Key:    "xxx",
		Secret: "xxx",
		Region: region,
	})

	return cli
}

func generateClusterCIDRInfo() *ClusterCIDRSettings {
	cidrInfo := &ClusterCIDRSettings{
		ClusterCIDR:          "xxx",
		MaxNodePodNum:        32,
		MaxClusterServiceNum: 1024,
	}

	return cidrInfo
}

func generateClusterBasicInfo() *ClusterBasicSettings {
	basicInfo := &ClusterBasicSettings{
		ClusterOS:      "",
		ClusterVersion: "",
		ClusterName:    "BCS-K8S-xxxxx",
		VpcID:          "vpc-xxxx",
	}

	tagTemplate := map[string]string{}

	basicInfo.TagSpecification = make([]*TagSpecification, 0)
	tags := make([]*Tag, 0)
	for k, v := range tagTemplate {
		tags = append(tags, &Tag{
			Key:   common.StringPtr(k),
			Value: common.StringPtr(v),
		})
	}
	basicInfo.TagSpecification = append(basicInfo.TagSpecification, &TagSpecification{
		ResourceType: "cluster",
		Tags:         tags,
	})

	return basicInfo
}

func generateClusterAdvancedInfo() *ClusterAdvancedSettings {
	advancedInfo := &ClusterAdvancedSettings{
		IPVS:             true,
		ContainerRuntime: "docker",
		RuntimeVersion:   "19.3",
	}

	if advancedInfo.ExtraArgs == nil {
		advancedInfo.ExtraArgs = &ClusterExtraArgs{}
	}

	advancedInfo.ExtraArgs.Etcd = []*string{
		common.StringPtr("node-data-dir=/data/bcs/lib/etcd"),
	}

	return advancedInfo
}

func generateInstanceAdvanceInfo() *InstanceAdvancedSettings {
	advanceInfo := &InstanceAdvancedSettings{
		MountTarget:     "/data",
		DockerGraphPath: "/data/bcs/service/docker",
		Unschedulable:   common.Int64Ptr(1),
	}

	return advanceInfo
}

func generateExistedInstance() *ExistedInstancesForNode {
	passwd := utils.BuildInstancePwd()

	masterInstanceIDs := []string{"ins-xxx", "ins-xxx", "ins-xxx"}
	existedInstance := &ExistedInstancesForNode{
		NodeRole: MASTER_ETCD.String(),
		ExistedInstancesPara: &ExistedInstancesPara{
			InstanceIDs:   masterInstanceIDs,
			LoginSettings: &LoginSettings{Password: passwd},
		},
	}

	return existedInstance
}

func TestTkeClient_CreateTKECluster(t *testing.T) {
	cli := getClient("ap-nanjing")
	req := &CreateClusterRequest{
		AddNodeMode:      false,
		Region:           "ap-nanjing",
		ClusterType:      "INDEPENDENT_CLUSTER",
		ClusterCIDR:      generateClusterCIDRInfo(),
		ClusterBasic:     generateClusterBasicInfo(),
		ClusterAdvanced:  generateClusterAdvancedInfo(),
		InstanceAdvanced: generateInstanceAdvanceInfo(),
	}

	req.ExistedInstancesForNode = []*ExistedInstancesForNode{
		generateExistedInstance(),
	}

	clusterRsp, err := cli.CreateTKECluster(req)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(clusterRsp.ClusterID)
}

func TestTkeClient_GetTKECluster(t *testing.T) {
	cli := getClient("ap-guangzhou")

	cluster, err := cli.GetTKECluster("cls-xxx")
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%+v", *cluster.ClusterStatus)
	t.Logf("%+v", *cluster.ClusterNetworkSettings.VpcId)
}

func TestTkeClient_ListTKECluster(t *testing.T) {
	cli := getClient(regions.Guangzhou)

	clusterList, err := cli.ListTKECluster()
	if err != nil {
		t.Fatal(err)
	}

	for i := range clusterList {
		t.Logf("%v\n", *clusterList[i].ClusterId)
	}
}

func TestTkeClient_AddExistedInstancesToCluster(t *testing.T) {
	cli := getClient("ap-nanjing")
	passwd := utils.BuildInstancePwd()

	req := &AddExistedInstanceReq{
		ClusterID:   "cls-xxx",
		InstanceIDs: []string{"ins-xxx"},
		AdvancedSetting: &InstanceAdvancedSettings{
			MountTarget:     MountTarget,
			DockerGraphPath: DockerGraphPath,
			Unschedulable:   common.Int64Ptr(1),
		},
		LoginSetting: &LoginSettings{Password: passwd},
	}
	resp, err := cli.AddExistedInstancesToCluster(req)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(resp.SuccessInstanceIDs, resp.FailedInstanceIDs)
}

func TestTkeClient_DeleteTkeClusterInstance(t *testing.T) {
	cli := getClient("ap-nanjing")

	resp, err := cli.DeleteTkeClusterInstance(&DeleteInstancesRequest{
		ClusterID:  "cls-xxx",
		Instances:  []string{"ins-xxx"},
		DeleteMode: Retain,
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Log(resp.Success, resp.Failure)
}

func TestTkeClient_DeleteTKECluster(t *testing.T) {
	cli := getClient("ap-nanjing")

	err := cli.DeleteTKECluster("cls-xxx", Retain)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("success")
}

func TestTkeClient_QueryTkeClusterAllInstances(t *testing.T) {
	cli := getClient("ap-nanjing")
	instances, err := cli.QueryTkeClusterAllInstances("cls-xxx")
	if err != nil {
		t.Fatal(err)
	}

	for i := range instances {
		t.Log(instances[i].InstanceID, instances[i].InstanceIP, instances[i].InstanceRole)
	}
	t.Log(len(instances))
}

func TestTkeClient_QueryTkeClusterInstances(t *testing.T) {
	cli := getClient("ap-nanjing")
	instances, err := cli.QueryTkeClusterInstances(&DescribeClusterInstances{
		ClusterID:    "cls-xxx",
		InstanceRole: NodePoolInstanceAll,
		Offset:       0,
		Limit:        100,
	})
	if err != nil {
		t.Fatal(err)
	}

	for i := range instances {
		t.Log(*instances[i].InstanceId, *instances[i].InstanceRole, *instances[i].LanIP,
			*instances[i].DrainStatus, *instances[i].InstanceState)
	}
}

func TestTkeClient_GetTKEClusterVersions(t *testing.T) {
	cli := getClient("ap-shenzhen")
	versions, err := cli.GetTKEClusterVersions()
	if err != nil {
		t.Fatal(err)
	}

	for _, version := range versions {
		t.Log(version.Name, version.Version)
	}
}

func TestTkeClient_GetTKEClusterImages(t *testing.T) {
	cli := getClient("ap-nanjing")
	images, err := cli.GetTKEClusterImages()
	if err != nil {
		t.Fatal(err)
	}

	for _, image := range images {
		t.Log(image.OsName, image.ImageID)
	}
}

func TestTkeClient_GetTKEClusterKubeConfig(t *testing.T) {
	cli := getClient("ap-guangzhou")
	kubeBytes, err := cli.GetTKEClusterKubeConfig("cls-xxx", true)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(kubeBytes)
}

func TestTkeClient_GetClusterEndpointStatus(t *testing.T) {
	cli := getClient("ap-guangzhou")
	status, err := cli.GetClusterEndpointStatus("cls-xxx", true)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(status)
}

func TestTkeClient_CreateClusterEndpoint(t *testing.T) {
	cli := getClient("ap-guangzhou")
	err := cli.CreateClusterEndpoint("cls-xxx")
	if err != nil {
		t.Fatal(err)
	}

	t.Log("success")
}

func TestTkeClient_DeleteClusterEndpoint(t *testing.T) {
	cli := getClient("ap-guangzhou")
	err := cli.DeleteClusterEndpoint("cls-xxx")
	if err != nil {
		t.Fatal(err)
	}

	t.Log("success")
}

func TestTkeClient_EnableTKEVpcCniMode(t *testing.T) {
	cli := getClient("ap-nanjing")
	err := cli.EnableTKEVpcCniMode(&EnableVpcCniInput{
		TkeClusterID:   "cls-xxx",
		VpcCniType:     "tke-direct-eni",
		SubnetsIDs:     []string{"subnet-xxx"},
		EnableStaticIP: true,
		ExpiredSeconds: 500,
	})
	if err != nil {
		t.Fatal(err)
	}

	for {
		time.Sleep(time.Second * 5)
		status, err := cli.GetEnableVpcCniProgress("cls-xxx")
		if err != nil {
			t.Fatal(err)
		}

		fmt.Println(status.Status, status.RequestID)
		if status.Status == string(Succeed) || status.Status == string(Failed) {
			fmt.Println("return")
			break
		}
	}
}

func TestTkeClient_GetEnableVpcCniProgress(t *testing.T) {
	cli := getClient("ap-nanjing")
	for {
		time.Sleep(time.Second * 5)
		status, err := cli.GetEnableVpcCniProgress("cls-xxx")
		if err != nil {
			t.Fatal(err)
		}

		fmt.Println(status.Status, status.RequestID)
		if status.Status == string(Succeed) || status.Status == string(Failed) {
			fmt.Println("return")
			break
		}
	}
}
