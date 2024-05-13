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

package main

import (
	"context"
	"flag"
	"fmt"

	"google.golang.org/grpc"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	pbnetagent "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/api/protocol/cloudnetagent"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/pkg/common"
)

func main() {
	var endpoint string
	var podname string
	var podns string
	var containerID string
	var address string
	var action string
	flag.StringVar(&endpoint, "endpoint", "", "endpoint")
	flag.StringVar(&action, "action", "", "action")
	flag.StringVar(&podname, "podname", "", "pod name")
	flag.StringVar(&podns, "podns", "", "pod ns")
	flag.StringVar(&containerID, "containerID", "", "container id")
	flag.StringVar(&address, "address", "", "ip address")

	flag.Parse()

	conn, err := grpc.Dial(endpoint, grpc.WithInsecure())
	if err != nil {
		blog.Fatalf("grpc dial failed, err %s", err.Error())
	}
	agentClient := pbnetagent.NewCloudNetagentClient(conn)
	switch action {
	case "alloc":
		resp, err := agentClient.AllocIP(context.Background(), &pbnetagent.AllocIPReq{
			Seq:          common.TimeSequence(),
			ContainerID:  containerID,
			PodName:      podname,
			PodNamespace: podns,
			IpAddr:       address,
		})
		if err != nil {
			blog.Fatalf("alloc ip failed, err %s", err.Error())
		}
		fmt.Printf("alloc resp: %+v", resp)
	case "release":
		resp, err := agentClient.ReleaseIP(context.Background(), &pbnetagent.ReleaseIPReq{
			Seq:          common.TimeSequence(),
			ContainerID:  containerID,
			PodName:      podname,
			PodNamespace: podns,
			IpAddr:       address,
		})
		if err != nil {
			blog.Fatalf("release ip failed, err %s", err.Error())
		}
		fmt.Printf("release resp: %+v", resp)
	}
}
