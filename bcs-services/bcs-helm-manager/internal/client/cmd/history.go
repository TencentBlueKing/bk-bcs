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

	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/client/cmd/printer"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/common"
	helmmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/proto/bcs-helm-manager"

	"github.com/spf13/cobra"
)

var (
	historyCMD = &cobra.Command{
		Use:   "history",
		Short: "get release history",
		Run:   GetReleaseHistory,
	}
)

// GetReleaseHistory provide the actions to do getReleaseHistoryCMD
func GetReleaseHistory(cmd *cobra.Command, args []string) {
	req := &helmmanager.GetReleaseHistoryReq{}

	if len(args) != 1 {
		fmt.Printf("get release history need release name\nExample: helmctl history -p " +
			"<project_code> -c <cluster_id> -n <namespace> <release_name>\n")
		os.Exit(1)
	}

	req.ProjectCode = &flagProject
	req.Name = common.GetStringP(args[0])
	req.Namespace = &flagNamespace
	req.ClusterID = &flagCluster

	c := newClientWithConfiguration()
	r, err := c.Release().GetReleaseHistory(cmd.Context(), req)
	if err != nil {
		fmt.Printf("get release history failed, %s\n", err.Error())
		os.Exit(1)
	}
	if flagOutput == outputTypeJSON {
		printer.PrintResultInJSON(r)
		return
	}

	printer.PrintReleaseHistoryInTable(flagOutput == outputTypeWide, r)
}

func init() {
	historyCMD.PersistentFlags().StringVarP(
		&flagProject, "project", "p", "", "project code")
	historyCMD.PersistentFlags().StringVarP(
		&flagCluster, "cluster", "c", "", "release cluster id")
	historyCMD.PersistentFlags().StringVarP(
		&flagNamespace, "namespace", "n", "", "release namespace")
	historyCMD.MarkPersistentFlagRequired("project")
	historyCMD.MarkPersistentFlagRequired("cluster")
	historyCMD.MarkPersistentFlagRequired("namespace")
}
