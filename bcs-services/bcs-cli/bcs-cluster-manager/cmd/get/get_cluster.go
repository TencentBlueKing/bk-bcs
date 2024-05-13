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
	clusterMgr "github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager/cluster"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager/types"
)

var (
	getClusterExample = templates.Examples(i18n.T(`
	kubectl-bcs-cluster-manager get cluster --clusterID xxx`))
)

func newGetClusterCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "cluster",
		Short:   "get cluster from bcs-cluster-manager",
		Example: getClusterExample,
		Run:     getCluster,
	}

	cmd.Flags().StringVarP(&clusterID, "clusterID", "c", "", "cluster ID (required)")
	_ = cmd.MarkFlagRequired("clusterID")

	return cmd
}

func getCluster(cmd *cobra.Command, args []string) {
	resp, err := clusterMgr.New(context.Background()).Get(types.GetClusterReq{ClusterID: clusterID})
	if err != nil {
		klog.Fatalf("get cluster failed: %v", err)
	}

	util.Output2Json(resp.Data)
}
