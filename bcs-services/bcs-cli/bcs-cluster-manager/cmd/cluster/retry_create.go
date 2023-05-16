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

func newRetryCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "retryCreate",
		Short: "retry create cluster from bcs-cluster-manager",
		Run:   retry,
	}

	cmd.Flags().StringVarP(&clusterID, "clusterID", "c", "", "cluster ID (required)")
	cmd.MarkFlagRequired("clusterID")

	return cmd
}

func retry(cmd *cobra.Command, args []string) {
	resp, err := clusterMgr.New(context.Background()).RetryCreate(types.RetryCreateClusterReq{
		ClusterID: clusterID,
	})
	if err != nil {
		klog.Fatalf("retry create cluster failed: %v", err)
	}

	fmt.Printf("retry create cluster succeed: clusterID: %v, taskID: %v", resp.ClusterID, resp.TaskID)
}
