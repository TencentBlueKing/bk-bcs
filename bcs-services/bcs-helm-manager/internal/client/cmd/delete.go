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

var (
	deleteCMD = &cobra.Command{
		Use:   "delete",
		Short: "delete",
		Long:  "delete resource",
	}
	deleteRepositoryCMD = &cobra.Command{
		Use:     "repository",
		Aliases: []string{"repo", "rp"},
		Short:   "delete repository",
		Long:    "delete repository",
		Run:     DeleteRepository,
	}
)

// DeleteRepository provide the actions to do deleteRepositoryCMD
func DeleteRepository(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		fmt.Printf("delete repository need specific repo name\n")
		os.Exit(1)
	}

	req := &helmmanager.DeleteRepositoryReq{
		ProjectID: &flagProject,
		Name:      common.GetStringP(args[0]),
	}

	c := newClientWithConfiguration()
	if err := c.Repository().Delete(cmd.Context(), req); err != nil {
		fmt.Printf("delete repository failed, %s\n", err.Error())
		os.Exit(1)
	}

	fmt.Printf("success to delete repository %s under project %s\n", req.GetName(), req.GetProjectID())
}

func init() {
	deleteCMD.AddCommand(deleteRepositoryCMD)
	deleteCMD.PersistentFlags().StringVarP(
		&flagProject, "project", "p", "", "project id for operation")
	deleteCMD.PersistentFlags().StringVarP(&jsonData, "data", "d", "", "resource json data")
	deleteCMD.PersistentFlags().StringVarP(&jsonFile, "file", "f", "", "resource json file")
}
