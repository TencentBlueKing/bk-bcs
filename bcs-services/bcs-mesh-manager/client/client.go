package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"google.golang.org/grpc"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/proto/meshmanager"
)

func main() {
	conn, err := grpc.Dial("127.0.0.1:8899", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	cli := meshmanager.NewMeshManagerClient(conn)
	resp,err := cli.CreateMeshCluster(context.TODO(), &meshmanager.CreateMeshClusterReq{Clusterid: "BCS-K8S-15091"})
	if err!=nil {
		fmt.Println(err.Error())
		return
	}
	by,_ := json.Marshal(resp)
	fmt.Println(string(by))
}
