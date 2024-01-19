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

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/utils/chart"
)

var (
	flagVersion string
	flagForce   bool

	pushCMD = &cobra.Command{
		Use:     "push",
		Short:   "push",
		Long:    "push chart",
		Run:     pushChart,
		Example: "helmctl push <file_or_dir> -p <project_code> -r <repo_name> -V <version> -f",
	}
)

// pushChart provide the actions to do pushChart
func pushChart(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		fmt.Printf("push chart need specific file or dir\n")
		os.Exit(1)
	}
	req := chart.UploadChart{
		RepoName:    flagRepository,
		FilePath:    args[0],
		ProjectCode: flagProject,
		Version:     flagVersion,
		Force:       flagForce,
	}

	c := newClientWithConfiguration()
	if err := c.Chart().Create(cmd.Context(), &req); err != nil {
		fmt.Printf("push chart failed, %s\n", err.Error())
		os.Exit(1)
	}

	fmt.Printf("success to create chart %s under project %s\n", req.RepoName, req.ProjectCode)
}

func init() {
	pushCMD.PersistentFlags().StringVarP(
		&flagProject, "project", "p", "", "project code")
	pushCMD.PersistentFlags().StringVarP(
		&flagRepository, "repo", "r", "", "repo name")
	pushCMD.PersistentFlags().StringVarP(
		&flagVersion, "version", "V", "", "version")
	pushCMD.PersistentFlags().BoolVarP(
		&flagForce, "force", "f", false, "force")
	_ = pushCMD.MarkPersistentFlagRequired("project")
	_ = pushCMD.MarkPersistentFlagRequired("repo")
}
