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

package list

import (
	"context"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/cmd/printer"
	nodegroup "github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager/node_group"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager/types"
	"github.com/spf13/cobra"
	"k8s.io/klog"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	listNodeGroupExample = templates.Examples(i18n.T(`
	kubectl-bcs-cluster-manager list nodeGroup`))
)

func newListNodeGroupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "nodeGroup",
		Short:   "list node group from bcs-cluster-manager",
		Example: listNodeGroupExample,
		Run:     listNodeGroup,
	}

	return cmd
}

func listNodeGroup(cmd *cobra.Command, args []string) {
	resp, err := nodegroup.New(context.Background()).List(types.ListNodeGroupReq{})
	if err != nil {
		klog.Fatalf("list node group failed: %v", err)
	}

	header := []string{"NODE_GROUP_ID", "NAME", "CLUSTER_ID", "REGION", "ENABLE_AUTOSCALE", "PROJECT_ID", "NODE_OS",
		"PROVIDER", "CONSUMER_ID", "STATUS", "BK_CLOUD_ID", "BK_CLOUD_NAME", "CREATOR", "UPDATER"}
	data := make([][]string, len(resp.Data))
	for key, item := range resp.Data {
		data[key] = []string{
			item.NodeGroupID,
			item.Name,
			item.ClusterID,
			item.Region,
			fmt.Sprintf("%t", item.EnableAutoscale),
			item.ProjectID,
			item.NodeOS,
			item.Provider,
			item.ConsumerID,
			item.Status,
			fmt.Sprintf("%d", item.BkCloudID),
			item.BkCloudName,
			item.Creator,
			item.Updater,
		}
	}

	err = printer.PrintList(flagOutput, resp.Data, header, data)
	if err != nil {
		klog.Fatalf("list cloud account to perm failed: %v", err)
	}
}
