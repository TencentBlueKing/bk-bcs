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

package retry

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	klog "k8s.io/klog/v2"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"

	clusterMgr "github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager/cluster"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager/types"
)

var (
	retryCreateClusterTaskExample = templates.Examples(i18n.T(`
	kubectl-bcs-cluster-manager retry createClusterTask --clusterID xxx`))
)

func newRetryCreateClusterTaskCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "createClusterTask",
		Short:   "retry create cluster from bcs-cluster-manager",
		Example: retryCreateClusterTaskExample,
		Run:     retryCreateClusterTask,
	}

	cmd.Flags().StringVarP(&clusterID, "clusterID", "c", "", "cluster ID (required)")
	_ = cmd.MarkFlagRequired("clusterID")

	return cmd
}

func retryCreateClusterTask(cmd *cobra.Command, args []string) {
	resp, err := clusterMgr.New(context.Background()).RetryCreate(types.RetryCreateClusterReq{
		ClusterID: clusterID,
	})
	if err != nil {
		klog.Fatalf("retry create cluster failed: %v", err)
	}

	fmt.Printf("retry create cluster succeed: clusterID: %v, taskID: %v\n", resp.ClusterID, resp.TaskID)
}
