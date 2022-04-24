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

	"github.com/Tencent/bk-bcs/bcs-common/common/codec"
	helmmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/proto/bcs-helm-manager"

	"github.com/spf13/cobra"
)

var (
	createCMD = &cobra.Command{
		Use:   "create",
		Short: "create",
		Long:  "create resource",
	}
	createRepositoryCMD = &cobra.Command{
		Use:     "repository",
		Aliases: []string{"repo", "rp"},
		Short:   "create repository",
		Long:    "create repository",
		Run:     CreateRepository,
	}
)

// CreateRepository provide the actions to do createRepositoryCMD
func CreateRepository(cmd *cobra.Command, _ []string) {
	data, err := getInputData()
	if err != nil {
		fmt.Printf("create repository failed, specify data by -d or -f, parse data failed: %s\n", err.Error())
		os.Exit(1)
	}

	var req helmmanager.CreateRepositoryReq
	if err = codec.DecJson(data, &req); err != nil {
		fmt.Printf("create repository failed, parse data failed, %s\n", err.Error())
		os.Exit(1)
	}

	c := newClientWithConfiguration()
	if err = c.Repository().Create(cmd.Context(), &req); err != nil {
		fmt.Printf("create repository failed, %s\n", err.Error())
		os.Exit(1)
	}

	fmt.Printf("success to create repository %s under project %s\n", req.GetName(), req.GetProjectID())
}

func init() {
	createCMD.AddCommand(createRepositoryCMD)
	createCMD.PersistentFlags().StringVarP(&jsonData, "data", "d", "", "resource json data")
	createCMD.PersistentFlags().StringVarP(&jsonFile, "file", "f", "", "resource json file")
}
