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

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/cmd/printer"
	clusterMgr "github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager/cluster"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager/types"
	"github.com/spf13/cobra"
	"k8s.io/klog"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	listNodesInClusterExample = templates.Examples(i18n.T(`
	kubectl-bcs-cluster-manager list nodesInCluster --clusterID xxx --offset 0 --limit 10`))
)

func newListNodesInClusterCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "nodesInCluster",
		Short:   "list cluster nodes from bcs-cluster-manager",
		Example: listNodesInClusterExample,
		Run:     listNodesInCluster,
	}

	cmd.Flags().StringVarP(&clusterID, "clusterID", "c", "", "cluster ID (required)")
	cmd.MarkFlagRequired("clusterID")
	cmd.Flags().Uint32VarP(&offset, "offset", "s", 0, "offset number of queries")
	cmd.Flags().Uint32VarP(&limit, "limit", "l", 10, "limit number of queries")

	return cmd
}

func listNodesInCluster(cmd *cobra.Command, args []string) {
	resp, err := clusterMgr.New(context.Background()).ListNodes(types.ListClusterNodesReq{
		ClusterID: clusterID,
		Offset:    offset,
		Limit:     limit,
	})
	if err != nil {
		klog.Fatalf("list cluster nodes failed: %v", err)
	}

	header := []string{"NODE_ID", "INNER_IP"}
	data := make([][]string, len(resp.Data))
	for key, item := range resp.Data {
		data[key] = []string{
			item.NodeID,
			item.InnerIP,
		}
	}

	err = printer.PrintList(flagOutput, resp.Data, header, data)
	if err != nil {
		klog.Fatalf("list cloud account to perm failed: %v", err)
	}
}
