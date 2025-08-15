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
 */

package api

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/regions"
	tke "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/tke/v20180525"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	icommon "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

func getClient(region string) *TkeClient {
	cli, _ := NewTkeClient(&cloudprovider.CommonOption{
		Account: &cmproto.Account{
			SecretID:  os.Getenv(TencentCloudSecretIDClusterEnv),
			SecretKey: os.Getenv(TencentCloudSecretKeyClusterEnv),
		},
		Region: region,
		CommonConf: cloudprovider.CloudConf{
			CloudInternalEnable: true,
			CloudDomain:         "tke.internal.tencentcloudapi.com",
		},
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
		ClusterOS:      "img-xxx",
		ClusterVersion: "1.20.6",
		ClusterName:    "xxx",
		VpcID:          "vpc-xxx",
		SubnetID:       "subnet-xxx",
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
		IPVS:             false,
		ContainerRuntime: "docker",
		RuntimeVersion:   "19.3",
		NetworkType:      "CiliumOverlay",
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
		Unschedulable:   common.Int64Ptr(0),
	}

	return advanceInfo
}

func generateExistedInstance() *ExistedInstancesForNode {
	passwd := utils.BuildInstancePwd()
	fmt.Println(passwd)

	// masterInstanceIDs := []string{"ins-xxx", "ins-xxx", "ins-xxx"}

	nodeInstance := []string{"ins-xxx"}
	existedInstance := &ExistedInstancesForNode{
		NodeRole: WORKER.String(),
		ExistedInstancesPara: &ExistedInstancesPara{
			InstanceIDs:   nodeInstance,
			LoginSettings: &LoginSettings{Password: passwd},
		},
	}

	return existedInstance
}

func TestTkeClient_CreateTKECluster(t *testing.T) {
	cli := getClient("ap-xxx")
	req := &CreateClusterRequest{
		AddNodeMode:      false,
		Region:           "ap-xxx",
		ClusterType:      "MANAGED_CLUSTER", // "INDEPENDENT_CLUSTER",
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
	cli := getClient(regions.Nanjing)

	cluster, err := cli.GetTKECluster("cls-xxx")
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%+v", *cluster.ClusterStatus)
	t.Logf("%+v", *cluster.ClusterNetworkSettings.VpcId)
	fmt.Println(*cluster.ClusterNetworkSettings.MaxNodePodNum)
	fmt.Println(*cluster.ClusterNetworkSettings.MaxClusterServiceNum)
	fmt.Println(*cluster.ClusterNetworkSettings.Ipvs)

	fmt.Println(*cluster.ClusterType)
	fmt.Println(*cluster.ClusterVersion)
	fmt.Println(*cluster.ClusterOs)
	fmt.Println(*cluster.ContainerRuntime)
	fmt.Println(*cluster.EnableExternalNode)
	fmt.Println(*cluster.ImageId)

	// t.Logf("%+v", *cluster.ClusterNetworkSettings.Subnets[0])
	fmt.Println(*cluster.ClusterNetworkSettings.ServiceCIDR)
	fmt.Println(*cluster.ClusterNetworkSettings.ClusterCIDR)
	fmt.Println(*cluster.ClusterNetworkSettings.Cni)

	fmt.Println(*cluster.Property)
	fmt.Println(*cluster.RuntimeVersion)
}

func TestGetTKEClusterKubeConfig(t *testing.T) {
	cli := getClient("ap-guangzhou")
	kubeBytes, err := cli.GetTKEClusterKubeConfig("cls-xxx", false)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(kubeBytes)
}

func TestGetClusterEndpointStatus(t *testing.T) {
	cli := getClient("ap-guangzhou")
	status, err := cli.GetClusterEndpointStatus("cls-xxx", true)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(status)
}

func TestCreateClusterEndpoint(t *testing.T) {
	cli := getClient("ap-guangzhou")
	err := cli.CreateClusterEndpoint("cls-xxx", ClusterEndpointConfig{IsExtranet: false})
	if err != nil {
		t.Fatal(err)
	}

	t.Log("success")
}

func TestDeleteClusterEndpoint(t *testing.T) {
	cli := getClient("ap-nanjing")
	err := cli.DeleteClusterEndpoint("cls-xxx", false)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("success")
}

func TestAddExistedInstancesToCluster(t *testing.T) {
	cli := getClient("ap-nanjing")

	passwd := utils.BuildInstancePwd()
	fmt.Println(passwd)

	req := &AddExistedInstanceReq{
		ClusterID:   "cls-xxx",
		InstanceIDs: []string{"ins-xxx"},
		AdvancedSetting: &InstanceAdvancedSettings{
			MountTarget:     MountTarget,
			DockerGraphPath: DockerGraphPath,
			Unschedulable:   common.Int64Ptr(1),
			Labels: []*KeyValue{
				{
					Name:  "1",
					Value: "2",
				},
				{
					Name:  "3",
					Value: "4",
				},
			},
			TaintList: MapToTaints([]*cmproto.Taint{
				{
					Key:    "5",
					Value:  "6",
					Effect: "NoSchedule",
				},
			}),
		},
		LoginSetting: &LoginSettings{Password: passwd},
	}
	resp, err := cli.AddExistedInstancesToCluster(req)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(resp.SuccessInstanceIDs, resp.FailedInstanceIDs)
}

func TestDeleteTkeClusterInstance(t *testing.T) {
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

func TestQueryTkeClusterAllInstances(t *testing.T) {
	cli := getClient("ap-guangzhou")
	instances, err := cli.QueryTkeClusterAllInstances(context.Background(), "cls-xxx", QueryClusterInstanceFilter{
		NodePoolID:           "",
		NodePoolInstanceType: "",
	})
	if err != nil {
		t.Fatal(err)
	}

	for i := range instances {
		t.Log(instances[i].InstanceID, instances[i].InstanceIP, instances[i].InstanceRole, instances[i].InstanceState,
			instances[i].NodePoolId, instances[i].AutoscalingGroupId)
	}
	t.Log(len(instances))
}

func TestQueryTkeClusterInstances(t *testing.T) {
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

func TestTkeClient_CloseVpcCniMode(t *testing.T) {
	cli := getClient("ap-nanjing")

	err := cli.CloseVpcCniMode("cls-xxx")
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
		EnableStaticIp: true,
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

func TestTkeClient_AddVpcCniSubnets(t *testing.T) {
	cli := getClient("ap-nanjing")

	err := cli.AddVpcCniSubnets(&AddVpcCniSubnetsInput{
		ClusterID: "cls-xxx",
		VpcID:     "vpc-xxx",
		SubnetIDs: []string{"subnet-xxx"},
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Log("success")
}

func TestGetEnableVpcCniProgress(t *testing.T) {
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

func TestCreateClusterNodePool(t *testing.T) {
	cli := getClient("ap-guangzhou")
	input := &CreateNodePoolInput{
		ClusterID:       common.StringPtr("cls-xxx"),
		EnableAutoscale: common.BoolPtr(false),
		Name:            common.StringPtr("test-node-pool"),
		AutoScalingGroupPara: &AutoScalingGroup{
			MaxSize:         common.Uint64Ptr(2),
			MinSize:         common.Uint64Ptr(0),
			DesiredCapacity: common.Uint64Ptr(0),
			VpcID:           common.StringPtr("vpc-xxx"),
			SubnetIds:       common.StringPtrs([]string{"subnet-xxx"}),
			RetryPolicy:     common.StringPtr("IMMEDIATE_RETRY"),
			ServiceSettings: &ServiceSettings{ScalingMode: common.StringPtr("CLASSIC_SCALING")},
		},
		LaunchConfigurePara: &LaunchConfiguration{
			InstanceType: common.StringPtr("SA2.MEDIUM2"),
			SystemDisk: &SystemDisk{
				DiskType: common.StringPtr("CLOUD_PREMIUM"),
				DiskSize: common.Uint64Ptr(50),
			},
			InternetAccessible: &InternetAccessible{
				InternetChargeType:      common.StringPtr("TRAFFIC_POSTPAID_BY_HOUR"),
				InternetMaxBandwidthOut: common.Uint64Ptr(0),
			},
			SecurityGroupIds:   common.StringPtrs([]string{"sg-xxx"}),
			InstanceChargeType: common.StringPtr("POSTPAID_BY_HOUR"),
		},
		InstanceAdvancedSettings: generateInstanceAdvanceInfo(),
	}
	np, err := cli.CreateClusterNodePool(input)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(np)
}

func TestDescribeClusterNodePools(t *testing.T) {
	cli := getClient("ap-guangzhou")
	np, total, err := cli.DescribeClusterNodePools("cls-xxx", nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(utils.ToJSONString(np), total)
}

func TestDescribeClusterNodePoolDetail(t *testing.T) {
	cli := getClient("ap-guangzhou")
	np, err := cli.DescribeClusterNodePoolDetail("cls-xxx", "np-xxx")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(utils.ToJSONString(np))
}

func TestModifyClusterNodePool(t *testing.T) {
	cli := getClient("ap-guangzhou")
	err := cli.ModifyClusterNodePool(&tke.ModifyClusterNodePoolRequest{
		ClusterId:   common.StringPtr("cls-xxx"),
		NodePoolId:  common.StringPtr("np-xxx"),
		Name:        common.StringPtr("test-node-pool"),
		MaxNodesNum: common.Int64Ptr(3),
		MinNodesNum: common.Int64Ptr(0),
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestDeleteClusterNodePool(t *testing.T) {
	cli := getClient("ap-guangzhou")
	err := cli.DeleteClusterNodePool("cls-xxx", []string{"np-xxx"}, false)
	if err != nil {
		t.Fatal(err)
	}
}

func TestModifyNodePoolCapacityByAsg(t *testing.T) {
	cli := getClient("ap-guangzhou")
	err := cli.ModifyNodePoolCapacity("cls-xxx", "np-xxx", 1)
	if err != nil {
		t.Fatal(err)
	}
}

func TestModifyNodePoolInstanceTypes(t *testing.T) {
	cli := getClient("ap-guangzhou")
	err := cli.ModifyNodePoolInstanceTypes("cls-xxx", "np-xxx", []string{"SA2.MEDIUM2", "S5.MEDIUM2", "S4.MEDIUM2"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestRemoveNodeFromNodePool(t *testing.T) {
	cli := getClient("ap-guangzhou")
	err := cli.RemoveNodeFromNodePool("cls-xxx", "np-xxx", []string{"ins-xxx"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestEnableExternalNodeSupport(t *testing.T) {
	cli := getClient(regions.Nanjing)

	err := cli.EnableExternalNodeSupport("cls-xxx", EnableExternalNodeConfig{
		NetworkType: "Cilium VXLan",
		ClusterCIDR: "xxx/20",
		SubnetId:    "subnet-xxx",
		Enabled:     true,
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Log("success")
}

func TestCreateExternalNodePool(t *testing.T) {
	cli := getClient(regions.Nanjing)

	nodePoolID, err := cli.CreateExternalNodePool("cls-xxx", CreateExternalNodePoolConfig{
		Name:             "xxx",
		ContainerRuntime: "docker",
		RuntimeVersion:   "19.3",
		Labels: []*Label{
			{
				Name:  common.StringPtr("xxx"),
				Value: common.StringPtr("yyy"),
			},
		},
		Taints:                   nil,
		InstanceAdvancedSettings: nil,
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Log(nodePoolID)
}

func TestModifyExternalNodePool(t *testing.T) {
	cli := getClient(regions.Nanjing)

	err := cli.ModifyExternalNodePool("cls-xxx", ModifyExternalNodePoolConfig{
		NodePoolId: "np-xxx",
		Labels: []*Label{
			{
				Name:  common.StringPtr("xxx"),
				Value: common.StringPtr("xxx"),
			},
		},
		Taints: []*Taint{
			{
				Key:    common.StringPtr("xxx"),
				Value:  common.StringPtr("1"),
				Effect: common.StringPtr("PreferNoSchedule"),
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Log("success")
}

func TestDescribeExternalNodePools(t *testing.T) {
	cli := getClient(regions.Shanghai)

	nodePools, err := cli.DescribeExternalNodePools("cls-xxx")
	if err != nil {
		t.Fatal(err)
	}

	for _, pool := range nodePools {
		t.Log(*pool.Name, *pool.NodePoolId, *pool.LifeState)
	}
}

func TestDeleteExternalNodePool(t *testing.T) {
	cli := getClient(regions.Nanjing)

	err := cli.DeleteExternalNodePool("cls-xxx", DeleteExternalNodePoolConfig{
		NodePoolIds: []string{"np-xxx"},
		Force:       true,
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Log("success")
}

func TestDescribeExternalNode(t *testing.T) {
	cli := getClient(regions.Shanghai)

	nodes, err := cli.DescribeExternalNode("cls-xxx", DescribeExternalNodeConfig{
		NodePoolId: "np-xxx",
	})
	if err != nil {
		t.Fatal(err)
	}

	for _, node := range nodes {
		t.Log(node.Name, node.NodePoolId, node.IP, node.Location, node.Status)
	}
}

func TestDescribeExternalNodeScript(t *testing.T) {
	cli := getClient(regions.Nanjing)

	scriptInfo, err := cli.DescribeExternalNodeScript("cls-xxx", DescribeExternalNodeScriptConfig{
		NodePoolId: "np-xxx",
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Log(*scriptInfo.Link, "\n", *scriptInfo.Token, "\n", *scriptInfo.Command)

	cmd := base64.StdEncoding.EncodeToString([]byte(*scriptInfo.Command))
	t.Log(cmd)

	src, _ := base64.StdEncoding.DecodeString(cmd)
	t.Log(string(src))
}

func TestTkeClient_DeleteExternalNode(t *testing.T) {
	cli := getClient(regions.Nanjing)

	err := cli.DeleteExternalNode("cls-xxx", DeleteExternalNodeConfig{
		Names: []string{"node-xxx"},
		Force: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Log("success")
}

func TestDescribeNodeSupportConfig(t *testing.T) {
	cli := getClient(regions.Nanjing)

	resp, err := cli.DescribeExternalNodeSupportConfig("cls-xxx")
	if err != nil {
		t.Fatal(err)
	}

	t.Log(resp.Proxy)
	t.Log(resp.Master)
	t.Log(resp.FailedReason)
	t.Log(resp.SwitchIP)
	t.Log(resp.Enabled)
	t.Log(resp.Status)
	t.Log(resp.SubnetId, resp.NetworkType, resp.ClusterCIDR)
}

func TestDescribeVpcCniPodLimits(t *testing.T) {
	cli := getClient(regions.Nanjing)

	limits, err := cli.DescribeVpcCniPodLimits("ap-nanjing-1", "S5.LARGE16")
	if err != nil {
		t.Fatal(err)
	}

	t.Log(limits.Limits.RouterEniNonStaticIP, limits.Limits.RouterEniStaticIP, limits.Limits.directEni)
}

func TestClient_DescribeOSImages(t *testing.T) {
	cli := getClient(regions.Nanjing)

	defaultCommonOption.Region = regions.Nanjing
	images, err := cli.DescribeOsImages(icommon.PrivateImageProvider, "", nil, defaultCommonOption)
	if err != nil {
		t.Fatal(err)
	}

	for _, image := range images {
		fmt.Printf("%+v %+v %+v %+v\n", image.OsName, image.Status, image.ImageId, image.Arch)
	}

	t.Log(len(images))
}

func TestAcquireClusterAdminRole(t *testing.T) {
	cli := getClient(regions.Nanjing)

	clusterID := "cls-xxx"
	err := cli.AcquireClusterAdminRole(clusterID)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("success")
}

func TestDescribeClusterKubeconfig(t *testing.T) {
	cli := getClient(regions.Nanjing)

	clusterID := "cls-xxx"
	kube, err := cli.GetTKEClusterKubeConfig(clusterID, false)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(kube)
}

func TestNewTkeClient_GetTkeAppChartList(t *testing.T) {
	cli := getClient(regions.Nanjing)

	version, err := cli.GetTkeAppChartVersionByName("", "cos")
	if err != nil {
		t.Fatal(err)
	}

	t.Log(version)
}
