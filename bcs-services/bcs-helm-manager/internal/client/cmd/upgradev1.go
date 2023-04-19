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
	"strings"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/common"
	helmmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/proto/bcs-helm-manager"

	"github.com/spf13/cobra"
)

var upgradeV1CMD = &cobra.Command{
	Use:   "upgradev1",
	Short: "upgradev1",
	Long:  "upgradev1 chart release",
	Run:   UpgradeV1,
}

// UpgradeV1 provide the actions to do upgradeV1CMD
func UpgradeV1(cmd *cobra.Command, args []string) {
	req := &helmmanager.UpgradeReleaseV1Req{}

	if len(args) < 3 {
		fmt.Printf("upgradev1 args need at least 3, install [name] [chart] [version]\n")
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
	req.Repository = &flagRepository
	req.Chart = common.GetStringP(args[1])
	req.Version = common.GetStringP(args[2])
	req.Values = values
	req.ValueFile = &flagValueFile[0]
	if flagArgs != "" {
		req.Args = strings.Split(flagArgs, " ")
	}

	c := newClientWithConfiguration()
	err = c.Release().UpgradeV1(cmd.Context(), req)
	if err != nil {
		fmt.Printf("upgrade release failed, %s\n", err.Error())
		os.Exit(1)
	}

	fmt.Printf("success to upgrade release %s in version %s namespace %s cluster %s "+
		"with appVersion %s revision %d\n",
		req.GetName(), req.GetVersion(), req.GetNamespace(), req.GetClusterID())
}

func init() {
	upgradeV1CMD.PersistentFlags().StringVarP(
		&flagProject, "project", "p", "", "project id for operation")
	upgradeV1CMD.PersistentFlags().StringVarP(
		&flagRepository, "repository", "r", "", "repository name for operation")
	upgradeV1CMD.PersistentFlags().StringVarP(
		&flagNamespace, "namespace", "n", "", "release namespace for operation")
	upgradeV1CMD.PersistentFlags().StringVarP(
		&flagCluster, "cluster", "", "", "release cluster id for operation")
	upgradeV1CMD.PersistentFlags().StringSliceVarP(
		&flagValueFile, "file", "f", nil, "value file for installation")
	upgradeV1CMD.PersistentFlags().StringVarP(
		&flagArgs, "args", "", "", "args to append to helm command")
	upgradeV1CMD.PersistentFlags().StringVarP(
		&sysVarFile, "sysvar", "", "", "sys var file")
}
