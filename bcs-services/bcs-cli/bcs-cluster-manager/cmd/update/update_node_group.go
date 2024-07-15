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

package update

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
	updateNodeGroupExample = templates.Examples(i18n.T(`update node group from json file. file template: 
	{"nodeGroupID":"BCS-ng-xxxxx","name":"evan测试","clusterID":"BCS-K8S-xxx","region":"ap-shanghai",
	"enableAutoscale":true,"autoScaling":{"autoScalingID":"asg-xxxxxx","autoScalingName":"tke-np-xxxxx",
	"minSize":0,"maxSize":10,"desiredSize":0,"vpcID":"vpc-xxx","defaultCooldown":300,"subnetIDs":
	["subnet-xxxxxx"],"zones":[],"retryPolicy":"IMMEDIATE_RETRY","multiZoneSubnetPolicy":"PRIORITY",
	"replaceUnhealthy":false,"scalingMode":"CLASSIC_SCALING","timeRanges":[]},"launchTemplate":
	{"launchConfigurationID":"asc-xxxxxx","launchConfigureName":"tke-np-xxxxx","projectID":"0","CPU":4,
	"Mem":8,"GPU":0,"instanceType":"S4.LARGE8","instanceChargeType":"POSTPAID_BY_HOUR","systemDisk":
	{"diskType":"CLOUD_PREMIUM","diskSize":"50","fileSystem":"","autoFormatAndMount":false,"mountTarget":""},
	"dataDisks":[{"diskType":"CLOUD_PREMIUM","diskSize":"50","fileSystem":"ext4","autoFormatAndMount":false,
	"mountTarget":"/var/lib/docker"}],"internetAccess":{"internetChargeType":"","internetMaxBandwidth":"0",
	"publicIPAssigned":false},"initLoginPassword":"","securityGroupIDs":["sg-xxxx"],"imageInfo":{"imageID":
	"img-eb30mz89","imageName":"TencentOS Server 3.1 (TK4)"},"isSecurityService":true,"isMonitorService":true,
	"userData":"xxxxxxxx"},"labels":{},"taints":{},"nodeOS":"","creator":"bcs","updater":"bcs",
	"createTime":"2022-11-18T14:28:06+08:00","updateTime":"2022-11-18T14:28:06+08:00","projectID":"xxxxxxxxx",
	"provider":"tencentCloud","status":"RUNNING","consumerID":"","nodeTemplate":{"nodeTemplateID":"","name":"",
	"projectID":"","labels":{},"taints":[],"dockerGraphPath":"/var/lib/docker","mountTarget":"","userScript":"",
	"unSchedulable":0,"dataDisks":[{"diskType":"CLOUD_PREMIUM","diskSize":"50","fileSystem":"ext4",
	"autoFormatAndMount":false,"mountTarget":"/var/lib/docker"}],"extraArgs":{},"preStartUserScript":"",
	"bcsScaleOutAddons":null,"bcsScaleInAddons":null,"scaleOutExtraAddons":null,"scaleInExtraAddons":null,"nodeOS":"",
	"moduleID":"","creator":"","updater":"","createTime":"","updateTime":"","desc":"","runtime":{"containerRuntime":
	"docker","runtimeVersion":"19.x"},"module":null},"cloudNodeGroupID":"np-xxxxx","tags":{}}`))
)

func newUpdateNodeGroupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "nodeGroup",
		Short:   "update node group from bcs-cluster-manager",
		Example: updateNodeGroupExample,
		Run:     updateNodeGroup,
	}

	return cmd
}

func updateNodeGroup(cmd *cobra.Command, args []string) {
	data, err := os.ReadFile(filename)
	if err != nil {
		klog.Fatalf("read json file failed: %v", err)
	}

	req := types.UpdateNodeGroupReq{}

	err = json.Unmarshal(data, &req)
	if err != nil {
		klog.Fatalf("unmarshal json file failed: %v", err)
	}

	err = nodegroup.New(context.Background()).Update(req)
	if err != nil {
		klog.Fatalf("update node group failed: %v", err)
	}

	fmt.Println("update node group succeed")
}
