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

package check

import (
	"context"
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
	checkCloudKubeConfigExample = templates.Examples(i18n.T(`
	kubectl-bcs-cluster-manager check cloudKubeConfig --filename xxx`))
)

func newCheckCloudKubeConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "cloudKubeConfig",
		Short:   "check cloud kube config from bcs-cluster-manager",
		Example: checkCloudKubeConfigExample,
		Run:     checkCloudKubeconfig,
	}

	return cmd
}

func checkCloudKubeconfig(cmd *cobra.Command, args []string) {
	data, err := os.ReadFile(filename)
	if err != nil {
		klog.Fatalf("read file failed: %v", err)
	}

	err = clusterMgr.New(context.Background()).CheckCloudKubeConfig(types.CheckCloudKubeConfigReq{
		Kubeconfig: string(data),
	})
	if err != nil {
		klog.Fatalf("check cloud kube config failed: %v", err)
	}

	fmt.Println("check cloud kube config succeed")
}
