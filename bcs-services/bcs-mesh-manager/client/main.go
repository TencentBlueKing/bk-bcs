package main

import (
	"context"
	"encoding/json"
	"github.com/Tencent/bk-bcs/bcs-common/common/ssl"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/proto/meshmicro"

	"k8s.io/klog"
	"github.com/micro/go-micro/v2/registry/etcd"
	"github.com/micro/go-micro/v2/service"
	"github.com/micro/go-micro/v2/service/grpc"
	"github.com/micro/go-micro/v2/registry"
)

func main(){
	conf := config.Config{}
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

	cli := meshmicro.NewMeshManagerService("bcs-mesh-manager.bkbcs.tencent.com", svc.Client())
	req := &meshmicro.ListMeshClusterReq{
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

