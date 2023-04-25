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
	"encoding/json"
	"fmt"
	"os"

	clusterMgr "github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager/cluster"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager/types"
	"github.com/spf13/cobra"
	"k8s.io/klog"
)

func newCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "create cluster from bcs-cluster-manager",
		Run:   create,
	}

	cmd.Flags().StringVarP(&file, "file", "f", "./create_cluster.json", `create cluster from json file.
file template:{"projectID":"b363e23b1b354928axxxxxxxxx","businessID":"3","engineType":"k8s",
"isExclusive":true,"clusterType":"single","creator":"bcs","manageType":"INDEPENDENT_CLUSTER",
"clusterName":"test001","environment":"prod","provider":"bluekingCloud","description":"创建测试集群",
"clusterBasicSettings":{"version":"v1.20.11"},"networkType":"overlay","region":"default","vpcID":"",
"networkSettings":{},"master":["11.144.254.xx","11.144.254.xx"]}`)

	return cmd
}

func create(cmd *cobra.Command, args []string) {
	data, err := os.ReadFile(file)
	if err != nil {
		klog.Fatalf("read json file failed: %v", err)
	}

	req := types.CreateClusterReq{}
	err = json.Unmarshal(data, &req)
	if err != nil {
		klog.Fatalf("unmarshal json file failed: %v", err)
	}

	resp, err := clusterMgr.New(context.Background()).Create(req)
	if err != nil {
		klog.Fatalf("create cluster failed: %v", err)
	}

	fmt.Printf("create cluster succeed: clusterID: %v, taskID: %v\n", resp.ClusterID, resp.TaskID)
}
