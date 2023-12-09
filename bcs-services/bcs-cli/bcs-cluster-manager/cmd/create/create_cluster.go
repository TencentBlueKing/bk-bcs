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

package create

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

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
	createClusterExample = templates.Examples(i18n.T(`create cluster from json file.file template:
	{"projectID":"b363e23b1b354928axxxxxxxxx","businessID":"3","engineType":"k8s","isExclusive":true,
	"clusterType":"single","creator":"bcs","manageType":"INDEPENDENT_CLUSTER","clusterName":"test001",
	"environment":"prod","provider":"bluekingCloud","description":"创建测试集群","clusterBasicSettings":
	{"version":"v1.20.xx"},"networkType":"overlay","region":"default","vpcID":"","networkSettings":{},
	"master":["xxx.xxx.xxx.xxx","xxx.xxx.xxx.xxx"]}.

	create virtual cluster json template: 
	{"clusterID":"","clusterName":"test-xxx","provider":"tencentCloud","region":"ap-xxx","vpcID":
	"vpc-xxx","projectID":"xxx","businessID":"xxx","environment":"debug","engineType":"k8s","isExclusive":
	true,"clusterType":"single","hostClusterID":"BCS-K8S-xxx","hostClusterNetwork":"devnet","labels":
	{"xxx":"xxx"},"creator":"bcs","onlyCreateInfo":false,"master":["xxx"],"networkSettings":{"cidrStep":
	1,"maxNodePodNum":1,"maxServiceNum":1},"clusterBasicSettings":{"version":"xxx"},"clusterAdvanceSettings":
	{"IPVS":false,"containerRuntime":"xxx","runtimeVersion":"xxx","extraArgs":{"xxx":"xxx"}},"nodeSettings":
	{"dockerGraphPath":"xxx","mountTarget":"xxx","unSchedulable":1,"labels":{"xxx":"xxx"},"extraArgs":{"xxx":
	"xxx"}},"extraInfo":{"xxx":"xxx"},"description":"xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx","ns":{"name":
	"ieg-bcs-xxxxxxxxxxxxxxx","labels":{"xxx":"xxx"},"annotations":{"xxx":"xxx"}}}`))
)

func newCreateClusterCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "cluster",
		Short:   "create cluster from bcs-cluster-manager",
		Example: createClusterExample,
		Run:     createCluster,
	}

	cmd.Flags().BoolVarP(&virtual, "virtual", "v", false, `create virtual cluster`)

	return cmd
}

func createCluster(cmd *cobra.Command, args []string) {
	data, err := os.ReadFile(filename)
	if err != nil {
		klog.Fatalf("read json file failed: %v", err)
	}

	if virtual {
		req := types.CreateVirtualClusterReq{}
		err = json.Unmarshal(data, &req)
		if err != nil {
			klog.Fatalf("unmarshal json file failed: %v", err)
		}

		resp := &types.CreateVirtualClusterResp{}
		err = clusterMgr.NewClusterMgrClient(&clusterMgr.Config{
			APIServer: viper.GetString("apiserver"),
			AuthToken: viper.GetString("authtoken"),
			Operator:  viper.GetString("operator"),
		}).Post("/bcsapi/v4/clustermanager/v1/vcluster", req, resp)
		if err != nil {
			klog.Fatalf("create virtual cluster failed: %v", err)
		}

		if resp != nil && resp.Code != 0 {
			klog.Fatalf("create virtual cluster failed: %s", resp.Message)
		}

		fmt.Printf("create virtual cluster succeed: clusterID: %v, taskID: %v\n", resp.Data.ClusterID, resp.Task.TaskID)

		return
	}

	req := types.CreateClusterReq{}
	err = json.Unmarshal(data, &req)
	if err != nil {
		klog.Fatalf("unmarshal json file failed: %v", err)
	}

	resp, err := cluster.New(context.Background()).Create(req)
	if err != nil {
		klog.Fatalf("create cluster failed: %v", err)
	}

	fmt.Printf("create cluster succeed: clusterID: %v, taskID: %v\n", resp.ClusterID, resp.TaskID)
}
