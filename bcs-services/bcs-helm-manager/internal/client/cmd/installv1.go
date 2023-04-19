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

var installV1CMD = &cobra.Command{
	Use:   "installv1",
	Short: "installv1",
	Long:  "install v1 chart release",
	Run:   InstallV1,
}

func init() {
	installV1CMD.PersistentFlags().StringVarP(
		&flagProject, "project", "p", "", "project id for operation")
	installV1CMD.PersistentFlags().StringVarP(
		&flagRepository, "repository", "r", "", "repository name for operation")
	installV1CMD.PersistentFlags().StringVarP(
		&flagNamespace, "namespace", "n", "", "release namespace for operation")
	installV1CMD.PersistentFlags().StringVarP(
		&flagCluster, "cluster", "", "", "release cluster id for operation")
	installV1CMD.PersistentFlags().StringSliceVarP(
		&flagValueFile, "file", "f", nil, "value file for installation")
	installV1CMD.PersistentFlags().StringVarP(
		&flagArgs, "args", "", "", "args to append to helm command")
	installV1CMD.PersistentFlags().StringVarP(
		&sysVarFile, "sysvar", "", "", "sys var file")
}

// InstallV1 provide the actions to do installV1CMD
func InstallV1(cmd *cobra.Command, args []string) {
	req := &helmmanager.InstallReleaseV1Req{}

	if len(args) < 3 {
		fmt.Printf("install args need at least 3, install [name] [chart] [version]\n")
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
	err = c.Release().InstallV1(cmd.Context(), req)
	if err != nil {
		fmt.Printf("install release failed, %s\n", err.Error())
		os.Exit(1)
	}

	fmt.Printf("success to install release %s in version %s namespace %s cluster %s "+
		"with appVersion %s revision %d\n",
		req.GetName(), req.GetVersion(), req.GetNamespace(), req.GetClusterID())
}
