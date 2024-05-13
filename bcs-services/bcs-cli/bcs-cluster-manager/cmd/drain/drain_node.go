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

package drain

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	klog "k8s.io/klog/v2"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/cmd/util"
	nodeMgr "github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager/node"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager/types"
)

var (
	drainNodeExample = templates.Examples(i18n.T(`
	kubectl-bcs-cluster-manager drain node --clusterID xxx --innerIPs xxx`))
)

func newDrainNodeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "node",
		Short:   "drain node from bcs-cluster-manager",
		Example: drainNodeExample,
		Run:     drainNode,
	}

	cmd.Flags().StringSliceVarP(&innerIPs, "innerIPs", "i", []string{},
		"node inner ip, for example: -i xxx.xxx.xxx.xxx -i xxx.xxx.xxx.xxx")
	_ = cmd.MarkFlagRequired("innerIPs")

	cmd.Flags().StringVarP(&clusterID, "clusterID", "c", "", "更新节点所属的clusterID")

	return cmd
}

func drainNode(cmd *cobra.Command, args []string) {
	resp, err := nodeMgr.New(context.Background()).Drain(types.DrainNodeReq{
		InnerIPs:  innerIPs,
		ClusterID: clusterID,
	})
	if err != nil {
		klog.Fatalf("drain node failed: %v", err)
	}

	if len(resp.Data) == 0 {
		fmt.Println("drain node succeed")
		return
	}

	fmt.Println("drain the following node failed:")
	util.Output2Json(resp.Data)
}
