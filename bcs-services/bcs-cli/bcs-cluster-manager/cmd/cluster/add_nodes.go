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

package cluster

import (
	"context"
	"fmt"

	clusterMgr "github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager/cluster"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager/types"
	"github.com/spf13/cobra"
	"k8s.io/klog"
)

func newAddNodesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "addNodes",
		Short: "add nodes to cluster from bcs-cluster-manager",
		Run:   addNodes,
	}

	cmd.Flags().StringVarP(&clusterID, "clusterID", "c", "", "cluster ID (required)")
	cmd.MarkFlagRequired("clusterID")

	cmd.Flags().StringSliceVarP(&nodes, "node", "n", []string{}, "node ip, for example: -n 47.43.47.100 -n 244.87.232.40")
	cmd.MarkFlagRequired("node")

	cmd.Flags().StringVarP(&initPassword, "initPassword", "p", "12345678", "init log password")

	return cmd
}

func addNodes(cmd *cobra.Command, args []string) {
	resp, err := clusterMgr.New(context.Background()).AddNodes(types.AddNodesClusterReq{
		ClusterID: clusterID,
		Nodes:     nodes,
	})
	if err != nil {
		klog.Fatalf("add nodes to cluster failed: %v", err)
	}

	fmt.Printf("add nodes to cluster succeed: taskID: %v", resp.TaskID)
}
