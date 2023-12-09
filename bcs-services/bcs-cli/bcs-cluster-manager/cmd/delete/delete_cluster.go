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

package delete

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/klog"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"

	clusterMgr "github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager/cluster"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager/types"
)

var (
	deleteClusterExample = templates.Examples(i18n.T(`
	kubectl-bcs-cluster-manager delete cluster --clusterID xxx`))
)

func newDeleteClusterCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "cluster",
		Short:   "delete cluster from bcs-cluster-manager",
		Example: deleteClusterExample,
		Run:     deleteCluster,
	}

	cmd.Flags().BoolVarP(&virtual, "virtual", "v", false, `delete virtual cluster`)
	cmd.Flags().StringVarP(&clusterID, "clusterID", "c", "", "cluster ID (required)")
	_ = cmd.MarkFlagRequired("clusterID")

	return cmd
}

func deleteCluster(cmd *cobra.Command, args []string) {
	if virtual {
		resp := &types.CreateVirtualClusterResp{}
		url := fmt.Sprintf("/bcsapi/v4/clustermanager/v1/vcluster/%s?operator=bcs&onlyDeleteInfo=false", clusterID)

		err := clusterMgr.NewClusterMgrClient(&clusterMgr.Config{
			APIServer: viper.GetString("apiserver"),
			AuthToken: viper.GetString("authtoken"),
			Operator:  viper.GetString("operator"),
		}).Delete(url, resp)
		if err != nil {
			klog.Fatalf("delete virtual cluster failed: %v", err)
		}

		if resp != nil && resp.Code != 0 {
			klog.Fatalf("delete virtual cluster failed: %s", resp.Message)
		}

		fmt.Printf("delete virtual cluster succeed: clusterID: %v, taskID: %v\n", resp.Data.ClusterID, resp.Task.TaskID)

		return
	}

	resp, err := cluster.New(context.Background()).Delete(types.DeleteClusterReq{ClusterID: clusterID})
	if err != nil {
		klog.Fatalf("delete cluster failed: %v", err)
	}

	fmt.Printf("create cluster succeed: clusterID: %v, taskID: %v\n", resp.ClusterID, resp.TaskID)
}
