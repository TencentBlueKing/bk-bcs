/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cmd

import (
	"fmt"
	"os"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/common"
	helmmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/proto/bcs-helm-manager"

	"github.com/spf13/cobra"
)

var uninstallCMD = &cobra.Command{
	Use:     "uninstall",
	Aliases: []string{"un"},
	Short:   "uninstall chart release",
	Long:    "uninstall chart release",
	Run:     Uninstall,
	Example: "helmctl uninstall -p <project_code> -c <cluster_id> -n <namespace> <release_name>",
}

// Uninstall provide the actions to do uninstallCMD
func Uninstall(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		fmt.Printf("uninstall release need specific release name\n")
		os.Exit(1)
	}

	req := &helmmanager.UninstallReleaseV1Req{}
	req.ProjectCode = &flagProject
	req.ClusterID = &flagCluster
	req.Namespace = &flagNamespace
	req.Name = common.GetStringP(args[0])

	c := newClientWithConfiguration()
	if err := c.Release().Uninstall(cmd.Context(), req); err != nil {
		fmt.Printf("uninstall release failed, %s\n", err.Error())
		os.Exit(1)
	}

	fmt.Printf("success to uninstall release %s\n", req.GetName())
}

func init() {
	uninstallCMD.PersistentFlags().StringVarP(
		&flagProject, "project", "p", "", "project code")
	uninstallCMD.PersistentFlags().StringVarP(
		&flagCluster, "cluster", "c", "", "release cluster id")
	uninstallCMD.PersistentFlags().StringVarP(
		&flagNamespace, "namespace", "n", "", "release namespace")
	uninstallCMD.MarkPersistentFlagRequired("project")
	uninstallCMD.MarkPersistentFlagRequired("cluster")
	uninstallCMD.MarkPersistentFlagRequired("namespace")
}
