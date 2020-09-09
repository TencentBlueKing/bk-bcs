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

package main

import (
	"context"
	"encoding/json"

	"github.com/Tencent/bk-bcs/bcs-common/common/ssl"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/proto/meshmanager"

	"k8s.io/klog"
	"github.com/micro/go-micro/v2/registry/etcd"
	"github.com/micro/go-micro/v2/service"
	"github.com/micro/go-micro/v2/service/grpc"
	"github.com/micro/go-micro/v2/registry"
)

func main(){
	conf := config.Config{}
	conf.EtcdCaFile = "/data/bcs/cert/k8s/etcd-ca.pem"
	conf.EtcdCertFile = "/data/bcs/cert/k8s/bcs-etcd.pem"
	conf.EtcdKeyFile = "/data/bcs/cert/k8s/bcs-etcd-key.pem"
	tlsConf,err := ssl.ClientTslConfVerity(conf.EtcdCaFile, conf.EtcdCertFile, conf.EtcdKeyFile, "")
	if err!=nil {
		klog.Errorf("new client tsl conf failed: %s", err.Error())
		return
	}
	// New Service
	regOption := func(e *registry.Options){
		e.Addrs = []string{"https://127.0.0.1:2379"}
		e.TLSConfig = tlsConf
	}
	svc := grpc.NewService(
		service.Registry(etcd.NewRegistry(regOption)),
	)
	svc.Client().Init()
	svc.Init()
	cli := meshmanager.NewMeshManagerService("meshmanager.bkbcs.tencent.com", svc.Client())
	req := &meshmanager.ListMeshClusterReq{
		Clusterid: "BCS-K8S-15091",
	}
	resp,err := cli.ListMeshCluster(context.Background(), req)
	if err!=nil {
		klog.Errorf("ListMeshCluster failed: %s", err.Error())
		return
	}
	by,_ := json.Marshal(resp)
	klog.Infof("resp %s", string(by))
}

