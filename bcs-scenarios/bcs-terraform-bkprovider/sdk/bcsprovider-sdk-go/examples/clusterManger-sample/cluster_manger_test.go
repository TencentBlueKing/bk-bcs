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

// Package clusterMangerSample 测试
package clusterMangerSample

import (
	"context"
	"log"
	"testing"

	pb "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-bkprovider/sdk/bcsprovider-sdk-go/sdk/common/utils"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-bkprovider/sdk/bcsprovider-sdk-go/sdk/service/clusterManger"
)

func Test_ImportCluster(t *testing.T) {
	req := &pb.ImportClusterReq{
		// 必填
		ClusterName: "demo",
		// 必填
		Provider: clusterManger.TencentCloud,
		// 必填
		Region: region,
		// 必填
		ProjectID: projectID,
		// 必填
		BusinessID: "xxx",
		// 必填
		Environment: "prod",
		// 必填
		AccountID: accountID,
		CloudMode: &pb.ImportCloudMode{
			// 必填
			CloudID: clsIdCase,
			// true表示，内网方式导入；false表示，公网方式导入 --- 建议填写
			Inter: true,
		},
	}

	resp, err := service.ImportCluster(context.TODO(), req)
	if err != nil {
		t.Fatalf("import cluster failed, err: %s", err.Error())
	}

	log.Printf("import cluster success. resp: %s", utils.ObjToPrettyJson(resp))
}

func Test_CreateCluster(t *testing.T) {
	req := &pb.CreateClusterReq{
		// ---- 必填
		Region: region,
		// ---- 必填
		CloudAccountID: accountID,
		// ---- 必填
		ClusterName: "tf-create-test",
		// 集群环境 prod/debug ---- 必填
		Environment: "debug",
		// 集群类型 ---- 必填
		ManageType: clusterManger.ManagedCluster, // 托管集群
		// 项目id ---- 必填
		ProjectID: projectID,
		// vpc ---- 必填
		VpcID: vpcID,
		// Provider ---- 必填
		Provider: clusterManger.TencentCloud,
		// BusinessID ---- 必填
		BusinessID: "xxxx",
		// 网络配置 ---- 必填
		NetworkSettings: &pb.NetworkSetting{
			ClusterIPv4CIDR: "172.20.0.0/20",
			MaxNodePodNum:   64,
			MaxServiceNum:   1024,
			// 子网为20时，cird是4096
			CidrStep: 4096,
		},
		ClusterBasicSettings: &pb.ClusterBasicSetting{
			// 若为托管集群，则本字段必填
			ClusterLevel: "L20",
			// 指定k8s版本 ---- 必填
			Version:     "1.26.1",
			VersionName: "1.26.1",
			// 指定操作系统 ---- 若不指定，默认为tlinux3.2x86_64 (选填)
			OS: "tlinux3.2x86_64",
			Area: &pb.CloudArea{
				// 指定云区域id ---- 若不指定，默认为0 (选填)
				BkCloudID: 0,
			},
			Module: &pb.ClusterModule{
				// node节点cc模块 ---- 必填
				WorkerModuleID: "xxxxx",
			},
		},
		ClusterAdvanceSettings: &pb.ClusterAdvanceSetting{
			// 指定运行时和版本 ---- 必填
			ContainerRuntime: "containerd",
			RuntimeVersion:   "1.6.9",
			// 指定网络类型 Global Router=GR, VPC-CNI=VPC-CNI --- 若不指定，默认为GR
			NetworkType: "GR",
		},
		// worker节点 ---- 必填
		Nodes: []string{"127.0.0.1", "127.0.0.2", "127.0.0.3"},
		// worker节点密码 ---- 必填
		NodeSettings: &pb.NodeSetting{
			WorkerLogin: &pb.NodeLoginInfo{
				InitLoginPassword: "xxxxxx",
			},
		},
		// 选填，有需要则设置
		NodeTemplateID: "",
	}
	// 若为独立集群，需要额外配置，如下：
	// 设置类型为INDEPENDENT_CLUSTER;
	// 设置master节点及密码
	//req = &pb.CreateClusterReq{
	//	// 独立集群
	//	ManageType: clusterManger.IndependentCluster,
	//	// master节点
	//	Master: []string{"127.0.0.1", "127.0.0.2", "127.0.0.3"},
	//	// master登录密码
	//	NodeSettings: &pb.NodeSetting{
	//		MasterLogin: &pb.NodeLoginInfo{
	//			InitLoginPassword: ngPasswordCase,
	//		},
	//	},
	//}

	resp, err := service.CreateCluster(context.TODO(), req)
	if err != nil {
		t.Fatalf("create cluster failed, err: %s", err.Error())
	}

	log.Printf("create cluster success. resp: %s", utils.ObjToPrettyJson(resp))
}

// Test_DeleteCluster done
func Test_DeleteCluster(t *testing.T) {
	req := &pb.DeleteClusterReq{
		ClusterID: clusterID,
		//ClusterID: "BCS-K8S-15001", // todo: remove
		// OnlyDeleteInfo 仅删除集群信息，但是不删除云上集群l
		//OnlyDeleteInfo: true,
		// DeleteClusterRecord 管理员操作, 设置true时仅删除集群数据库记录
		//DeleteClusterRecord: true,
	}
	// 注意删除集群资源时，集群中不能保留节点
	// 如果仅删除集群信息时，集群中可以保留节点(需要加上OnlyDeleteInfo+DeleteClusterRecord)

	resp, err := service.DeleteCluster(context.TODO(), req)
	if err != nil {
		t.Fatalf("delete cluster failed, err: %s", err.Error())
	}

	log.Printf("delete cluster success. resp: %s", utils.ObjToPrettyJson(resp))
}

// Test_UpdateCluster done
func Test_UpdateCluster(t *testing.T) {
	req := &pb.UpdateClusterReq{
		ClusterID: clusterID,
		// 例如名称修改
		ClusterName: "test-tf-update-case",
	}

	resp, err := service.UpdateCluster(context.TODO(), req)
	if err != nil {
		t.Fatalf("update cluster failed, err: %s", err.Error())
	}

	log.Printf("update cluster success. resp: %s", utils.ObjToPrettyJson(resp))
}

// Test_GetCluster done
func Test_GetCluster(t *testing.T) {
	req := &pb.GetClusterReq{
		ClusterID: clusterID,
		//CloudInfo: true,
	}

	resp, err := service.GetCluster(context.TODO(), req)
	if err != nil {
		t.Fatalf("get cluster failed, err: %s", err.Error())
	}

	log.Printf("get cluster success. resp: %s\n", utils.ObjToPrettyJson(resp))
}

// Test_ListProjectCluster done
func Test_ListProjectCluster(t *testing.T) {
	req := &pb.ListProjectClusterReq{
		ProjectID: projectID,
	}

	resp, err := service.ListProjectCluster(context.TODO(), req)
	if err != nil {
		t.Fatalf("list project cluster failed, err: %s", err.Error())
	}

	log.Printf("list project cluster success. resp: %s\n", utils.ObjToPrettyJson(resp))
}
