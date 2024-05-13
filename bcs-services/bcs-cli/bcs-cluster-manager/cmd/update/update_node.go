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
	"fmt"

	"github.com/spf13/cobra"
	klog "k8s.io/klog/v2"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"

	nodeMgr "github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager/node"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager/types"
)

var (
	updateNodeExample = templates.Examples(i18n.T(`
	kubectl-bcs-cluster-manager update node --status xxx --innerIPs xxx.xxx.xxx.xxx --innerIPs xxx.xxx.xxx.xxx`))
)

func newUpdateNodeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "node",
		Short:   "update node from bcs-cluster-manager",
		Example: updateNodeExample,
		Run:     updateNode,
	}

	cmd.Flags().StringSliceVarP(&innerIPs, "innerIPs", "i", []string{},
		"node inner ip, for example: -i xxx.xxx.xxx.xxx -i xxx.xxx.xxx.xxx")
	_ = cmd.MarkFlagRequired("innerIPs")

	cmd.Flags().StringVarP(&status, "status", "s", "",
		"更新节点状态(INITIALIZATION/RUNNING/DELETING/ADD-FAILURE/REMOVE-FAILURE)")
	cmd.Flags().StringVarP(&nodeGroupID, "nodeGroupID", "n", "", "更新节点所属的node group ID")
	cmd.Flags().StringVarP(&clusterID, "clusterID", "c", "", "更新节点所属的clusterID")

	return cmd
}

func updateNode(cmd *cobra.Command, args []string) {
	err := nodeMgr.New(context.Background()).Update(types.UpdateNodeReq{
		InnerIPs:    innerIPs,
		Status:      status,
		NodeGroupID: nodeGroupID,
		ClusterID:   clusterID,
	})
	if err != nil {
		klog.Fatalf("get node failed: %v", err)
	}

	fmt.Println("update node succeed")
}
