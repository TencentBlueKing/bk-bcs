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

package passcc

import (
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/auth"
	"testing"
	"time"
)

func getPermServer() *auth.ClientSSM {
	cli := auth.NewSSMClient(auth.Options{
		Server:    "http://xxx.com",
		AppCode:   "xxx",
		AppSecret: "xxx",
		Debug:     true,
	})

	return cli
}

var server = &ClientConfig{
	server:    "xxx",
	appCode:   "xxx",
	appSecret: "xxx",
	debug:     true,
}

func TestClientConfig_CreatePassCCClusterSnapshoot(t *testing.T) {
	token, err := server.getAccessToken(getPermServer())
	if err != nil {
		t.Fatal(err)
	}
	t.Log(token)

	cls := &proto.Cluster{
		ClusterID:   "BCS-K8S-xxxxx",
		ClusterName: "xx",
		Region:      "ap-nanjing",
		VpcID:       "vpc-xxx",
		ProjectID:   "xxx",
		BusinessID:  "xxx",
		Environment: "debug",
		EngineType:  "k8s",
		IsExclusive: false,
		ClusterType: "single",
		Creator:     "xxx",
		CreateTime:  time.Now().String(),
		UpdateTime:  time.Now().String(),
		Master: map[string]*proto.Node{
			"127.0.0.1": &proto.Node{
				NodeID:  "",
				InnerIP: "",
			},
		},
		SystemID:    "cls-xxx",
		ManageType:  "INDEPENDENT_CLUSTER",
		Description: "xxx",
	}

	err = server.CreatePassCCClusterSnapshoot(cls)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("success")
}

func TestClientConfig_DeletePassCCCluster(t *testing.T) {
	token, err := server.getAccessToken(getPermServer())
	if err != nil {
		t.Fatal(err)
	}
	t.Log(token)

	err = server.DeletePassCCCluster("xxx", "BCS-K8S-xxxxx")
	if err != nil {
		t.Fatal(err)
	}

	t.Log("success")
}

func TestClientConfig_CreatePassCCCluster(t *testing.T) {
	token, err := server.getAccessToken(getPermServer())
	if err != nil {
		t.Fatal(err)
	}
	t.Log(token)

	cls := &proto.Cluster{
		ClusterID:   "BCS-K8S-xxxxx",
		ClusterName: "Job - xxx",
		Region:      "ap-guangzhou",
		VpcID:       "vpc-xxx",
		ProjectID:   "xxx",
		BusinessID:  "xxx",
		Environment: "debug",
		EngineType:  "k8s",
		IsExclusive: false,
		ClusterType: "single",
		Creator:     "xxx",
		CreateTime:  time.Now().String(),
		UpdateTime:  time.Now().String(),
		Master: map[string]*proto.Node{
			"127.0.0.1": &proto.Node{
				NodeID:  "",
				InnerIP: "",
			},
		},
		SystemID:    "cls-xxx",
		ManageType:  "INDEPENDENT_CLUSTER",
		Description: "xxx",
	}

	err = server.CreatePassCCCluster(cls)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("success")
}
