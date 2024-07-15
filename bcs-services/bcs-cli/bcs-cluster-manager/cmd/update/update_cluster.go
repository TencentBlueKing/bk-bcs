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

package update

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	klog "k8s.io/klog/v2"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"

	clusterMgr "github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager/cluster"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager/types"
)

var (
	updateClusterExample = templates.Examples(i18n.T(`create cluster from json file. file template:
	{"clusterID":"BCS-K8S-xxx","projectID":"b363e23b1b354928xxxxxxxxxxxxxxxxxxxxxxxx","businessID":"3",
	"engineType":"k8s","isExclusive":true,"clusterType":"single","creator":"bcs","manageType":"INDEPENDENT_CLUSTER",
	"clusterName":"test002","environment":"prod","provider":"bluekingCloud","description":"update创建测试集群",
	"clusterBasicSettings":{"version":"v1.20.xx"},"networkType":"overlay","region":"default","vpcID":"",
	"networkSettings":{},"master":["xxx.xxx.xxx.xxx","xxx.xxx.xxx.xxx"]}`))
)

func newUpdateClusterCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "cluster",
		Short:   "update cluster from bcs-cluster-manager",
		Example: updateClusterExample,
		Run:     updateCluster,
	}

	return cmd
}

func updateCluster(cmd *cobra.Command, args []string) {
	data, err := os.ReadFile(filename)
	if err != nil {
		klog.Fatalf("read json file failed: %v", err)
	}

	req := types.UpdateClusterReq{}
	err = json.Unmarshal(data, &req)
	if err != nil {
		klog.Fatalf("unmarshal json file failed: %v", err)
	}

	err = clusterMgr.New(context.Background()).Update(req)
	if err != nil {
		klog.Fatalf("update cluster failed: %v", err)
	}

	fmt.Println("update cluster succeed")
}
