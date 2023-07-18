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
	"testing"
	"time"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/auth"
)

func getPermServer() *auth.ClientAuth {
	cli := auth.NewAccessClient(auth.Options{
		Server: "http://xxx",
		Debug:  true,
	})

	return cli
}

var server = &ClientConfig{
	server:    "http://xxx",
	appCode:   "bcs-xxx",
	appSecret: "xxx",
	debug:     true,
}

func TestCreatePassCCClusterSnap(t *testing.T) {
	token, err := server.getAccessToken(getPermServer())
	if err != nil {
		t.Fatal(err)
	}
	t.Log(token)

	namespaces, err := server.GetProjectSharedNamespaces("xxx", "BCS-K8S-xxx", getPermServer())
	if err != nil {
		t.Fatal(err)
	}

	for _, ns := range namespaces {
		t.Logf(ns.Name)
	}
}

func TestDeletePassCCCluster(t *testing.T) {
	token, err := server.getAccessToken(getPermServer())
	if err != nil {
		t.Fatal(err)
	}
	t.Log(token)

	err = server.DeletePassCCCluster("xxx", "BCS-K8S-xxx")
	if err != nil {
		t.Fatal(err)
	}

	t.Log("success")
}

func TestUpdatePassCCCluster(t *testing.T) {
	token, err := server.getAccessToken(getPermServer())
	if err != nil {
		t.Fatal(err)
	}
	t.Log(token)

	cls := &proto.Cluster{
		ClusterID:   "BCS-K8S-xxx",
		ClusterName: "xxx",
		Region:      "xxx",
		VpcID:       "xxx",
		ProjectID:   "xxx",
		BusinessID:  "xxx",
		Environment: "stag",
		EngineType:  "k8s",
		IsExclusive: false,
		ClusterType: "single",
		Creator:     "xxxx",
		CreateTime:  time.Now().String(),
		UpdateTime:  time.Now().String(),
		Master: map[string]*proto.Node{
			"xxx": {
				NodeID:  "",
				InnerIP: "",
			},
		},
		SystemID:    "xxx",
		ManageType:  "INDEPENDENT_CLUSTER",
		Description: "测试集群",
	}

	err = server.UpdatePassCCCluster(cls)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("success")
}

func TestCreatePassCCCluster(t *testing.T) {
	token, err := server.getAccessToken(getPermServer())
	if err != nil {
		t.Fatal(err)
	}
	t.Log(token)

	cls := &proto.Cluster{
		ClusterID:   "BCS-K8S-xxx",
		ClusterName: "Job - xxx",
		Region:      "xxx",
		VpcID:       "xxx",
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
			"127.0.0.1": {
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
