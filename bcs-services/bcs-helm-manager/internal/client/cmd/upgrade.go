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

var upgradeCMD = &cobra.Command{
	Use:   "upgrade",
	Short: "upgrade",
	Long:  "upgrade chart release",
	Run:   Upgrade,
	Example: "helmctl upgrade -p <project_code> -c <cluster_id> -n <namespace> <release_name> <chart_name> " +
		"<version> -f values.yaml",
}

// Upgrade provide the actions to do upgradeCMD
func Upgrade(cmd *cobra.Command, args []string) {
	req := &helmmanager.UpgradeReleaseV1Req{}

	if len(args) < 3 {
		fmt.Printf("upgrade args need at least 3, upgrade [name] [chart] [version]\n")
		os.Exit(1)
	}
	values, err := getValues()
	if err != nil {
		fmt.Printf("read values file failed, %s\n", err.Error())
		os.Exit(1)
	}

	req.Name = common.GetStringP(args[0])
	req.Namespace = &flagNamespace
	req.ClusterID = &flagCluster
	req.ProjectCode = &flagProject
	if flagRepository == "" {
		flagRepository = flagProject
	}
	req.Repository = &flagRepository
	req.Chart = common.GetStringP(args[1])
	req.Version = common.GetStringP(args[2])
	req.Values = values
	req.Args = flagArgs

	c := newClientWithConfiguration()
	err = c.Release().Upgrade(cmd.Context(), req)
	if err != nil {
		fmt.Printf("upgrade release failed, %s\n", err.Error())
		os.Exit(1)
	}

	fmt.Printf("success to upgrade release %s", req.GetName())
}

func init() {
	upgradeCMD.PersistentFlags().StringVarP(
		&flagProject, "project", "p", "", "project code")
	upgradeCMD.PersistentFlags().StringVarP(
		&flagRepository, "repo", "r", "", "repository name")
	upgradeCMD.PersistentFlags().StringVarP(
		&flagCluster, "cluster", "c", "", "release cluster id")
	upgradeCMD.PersistentFlags().StringVarP(
		&flagNamespace, "namespace", "n", "", "release namespace")
	upgradeCMD.PersistentFlags().StringSliceVarP(
		&flagValueFile, "file", "f", nil, "value file for installation, -f values.yaml")
	upgradeCMD.PersistentFlags().StringSliceVarP(
		&flagArgs, "args", "", nil, "--args=--wait=true --args=--timeout=600s")
	upgradeCMD.MarkPersistentFlagRequired("project")
	upgradeCMD.MarkPersistentFlagRequired("cluster")
	upgradeCMD.MarkPersistentFlagRequired("namespace")
}
