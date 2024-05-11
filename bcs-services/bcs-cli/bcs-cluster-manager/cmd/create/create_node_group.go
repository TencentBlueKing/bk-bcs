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

package create

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	klog "k8s.io/klog/v2"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"

	nodegroup "github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager/node_group"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager/types"
)

var (
	createNodeGroupExample = templates.Examples(i18n.T(`create node group from json file.file template:
	{"name":"test001","autoScaling":{"maxSize":10,"minSize":0,"scalingMode":"CLASSIC_SCALING",
	"multiZoneSubnetPolicy":"PRIORITY","retryPolicy":"IMMEDIATE_RETRY","subnetIDs":["subnet-xxxxxxx"]},
	"enableAutoscale":true,"nodeTemplate":{"unSchedulable":0,"labels":{},"taints":[],"dataDisks":[],
	"dockerGraphPath":"/var/lib/docker","runtime":{"containerRuntime":"docker","runtimeVersion":"19.x"}},
	"launchTemplate":{"imageInfo":{"imageID":"img-xxxxx"},"CPU":2,"Mem":2,"instanceType":"S4.MEDIUM2",
	"systemDisk":{"diskType":"CLOUD_PREMIUM","diskSize":"50"},"internetAccess":{"internetChargeType":
	"TRAFFIC_POSTPAID_BY_HOUR","internetMaxBandwidth":"0","publicIPAssigned":false},"initLoginPassword":"123456",
	"securityGroupIDs":["sg-xxx"],"dataDisks":[],"isSecurityService":true,"isMonitorService":true},
	"clusterID":"BCS-K8S-xxxxx","region":"ap-shanghai","creator":"bcs"}`))
)

func newCreateNodeGroupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "nodeGroup",
		Short:   "create node group from bcs-cluster-manager",
		Example: createNodeGroupExample,
		Run:     createNodeGroup,
	}

	return cmd
}

func createNodeGroup(cmd *cobra.Command, args []string) {
	data, err := os.ReadFile(filename)
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

	fmt.Printf("create node group succeed: nodeGroupID: %v, taskID: %v\n", resp.NodeGroupID, resp.TaskID)
}
