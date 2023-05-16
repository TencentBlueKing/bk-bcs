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

package nodegroup

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	nodegroup "github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager/node_group"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager/types"
	"github.com/spf13/cobra"
	"k8s.io/klog"
)

func newCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "create node group from bcs-cluster-manager",
		Run:   create,
	}

	cmd.Flags().StringVarP(&file, "file", "f", "./create_node_group.json", `create node group from json file.
file template: {"name":"test001","autoScaling":{"maxSize":10,"minSize":0,"scalingMode":"CLASSIC_SCALING",
"multiZoneSubnetPolicy":"PRIORITY","retryPolicy":"IMMEDIATE_RETRY","subnetIDs":["subnet-5zem7xxx"]},
"enableAutoscale":true,"nodeTemplate":{"unSchedulable":0,"labels":{},"taints":[],"dataDisks":[],
"dockerGraphPath":"/var/lib/docker","runtime":{"containerRuntime":"docker","runtimeVersion":"19.3"}},
"launchTemplate":{"imageInfo":{"imageID":"img-fv2263iz"},"CPU":2,"Mem":2,"instanceType":"S4.MEDIUM2",
"systemDisk":{"diskType":"CLOUD_PREMIUM","diskSize":"50"},"internetAccess":{"internetChargeType":
"TRAFFIC_POSTPAID_BY_HOUR","internetMaxBandwidth":"0","publicIPAssigned":false},"initLoginPassword":"123456",
"securityGroupIDs":["sg-dhjkgqo4"],"dataDisks":[],"isSecurityService":true,"isMonitorService":true},
"clusterID":"BCS-K8S-xxxxx","region":"ap-shanghai","creator":"bcs"}`)

	return cmd
}

func create(cmd *cobra.Command, args []string) {
	data, err := os.ReadFile(file)
	if err != nil {
		klog.Fatalf("read json file failed: %v", err)
	}

	req := types.CreateNodeGroupReq{}

	err = json.Unmarshal(data, &req)
	if err != nil {
		klog.Fatalf("unmarshal json file failed: %v", err)
	}

	resp, err := nodegroup.New(context.Background()).Create(req)
	if err != nil {
		klog.Fatalf("create node group failed: %v", err)
	}

	fmt.Printf("create node group succeed: nodeGroupID: %v, taskID: %v", resp.NodeGroupID, resp.TaskID)
}
