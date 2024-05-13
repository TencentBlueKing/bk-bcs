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

package check

import (
	"context"

	"github.com/spf13/cobra"
	klog "k8s.io/klog/v2"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/cmd/util"
	nodeMgr "github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager/node"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager/types"
)

var (
	checkNodeInClusterExample = templates.Examples(i18n.T(`
	kubectl-bcs-cluster-manager check nodeInCluster --innerIPs xxx`))
)

func newCheckNodeInClusterCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "nodeInCluster",
		Short:   "check node in cluster from bcs-cluster-manager",
		Example: checkNodeInClusterExample,
		Run:     checkNodeInCluster,
	}

	cmd.Flags().StringSliceVarP(&innerIPs, "innerIPs", "i", []string{},
		"node inner ip, for example: -i xxx.xxx.xxx.xxx -i xxx.xxx.xxx.xxx")
	_ = cmd.MarkFlagRequired("innerIPs")

	return cmd
}

func checkNodeInCluster(cmd *cobra.Command, args []string) {
	resp, err := nodeMgr.New(context.Background()).CheckNodeInCluster(types.CheckNodeInClusterReq{
		InnerIPs: innerIPs,
	})
	if err != nil {
		klog.Fatalf("check node in cluster failed: %v", err)
	}

	util.Output2Json(resp.Data)
}
