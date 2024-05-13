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

package get

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
	getNodeExample = templates.Examples(i18n.T(`
	kubectl-bcs-cluster-manager get node --innerIP xxx`))
)

func newGetNodeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "node",
		Short:   "get node from bcs-cluster-manager",
		Example: getNodeExample,
		Run:     getNode,
	}

	cmd.Flags().StringVarP(&innerIP, "innerIP", "i", "", "")
	_ = cmd.MarkFlagRequired("innerIP")

	return cmd
}

func getNode(cmd *cobra.Command, args []string) {
	resp, err := nodeMgr.New(context.Background()).Get(types.GetNodeReq{
		InnerIP: innerIP,
	})
	if err != nil {
		klog.Fatalf("get node failed: %v", err)
	}

	util.Output2Json(resp.Data)
}
