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

func newImportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import",
		Short: "import cluster from bcs-cluster-manager",
		Run:   importFunc,
	}

	cmd.Flags().StringVarP(&file, "file", "f", "./import.json", `import cluster from json file. file template: 
	{"clusterID":"","projectID":"","businessID":"100001","engineType":"k8s","isExclusive":false,
	"clusterType":"single","clusterName":"ceshi","environment":"stag","provider":"tencentCloud"}`)
	cmd.MarkFlagRequired("file")

	return cmd
}

func importFunc(cmd *cobra.Command, args []string) {
	data, err := os.ReadFile(file)
	if err != nil {
		klog.Fatalf("read file failed: %v", err)
	}

	req := types.ImportClusterReq{}
	err = json.Unmarshal(data, &req)
	if err != nil {
		klog.Fatalf("unmarshal json file failed: %v", err)
	}

	err = clusterMgr.New(context.Background()).Import(req)
	if err != nil {
		klog.Fatalf("import cluster failed: %v", err)
	}

	fmt.Printf("import cluster succeed")
}
