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

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/cmd/util"
	clusterMgr "github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager/cluster"
	"github.com/spf13/cobra"
	"k8s.io/klog"
)

func newListCommonCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "listCommon",
		Short: "list common cluster from bcs-cluster-manager",
		Run:   listCommon,
	}

	return cmd
}

func listCommon(cmd *cobra.Command, args []string) {
	resp, err := clusterMgr.New(context.Background()).ListCommon()
	if err != nil {
		klog.Fatalf("list common cluster failed: %v", err)
	}

	util.Output2Json(resp.Data)
}
