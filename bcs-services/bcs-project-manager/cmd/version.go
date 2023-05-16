/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * 	http://opensource.org/licenses/MIT
 *
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/version"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of bcs-project",
	Run: func(cmd *cobra.Command, args []string) {
		info := []string{
			"Version:    " + version.Version,
			"Commit:     " + version.GitCommit,
			"Build Time: " + version.BuildTime,
			"Go Version: " + version.GoVersion,
		}
		fmt.Fprintf(cmd.OutOrStdout(), strings.Join(info, "\n"))
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
