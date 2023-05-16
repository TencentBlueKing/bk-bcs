/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 *  Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 *  Licensed under the MIT License (the "License"); you may not use this file except
 *  in compliance with the License. You may obtain a copy of the License at
 *  http://opensource.org/licenses/MIT
 *  Unless required by applicable law or agreed to in writing, software distributed under
 *  the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 *  either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cmd

import (
	"fmt"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-nodegroup-manager/cmd/client/pkg"
	nodegroupmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-nodegroup-manager/proto"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

// NewDeleteCmd new delete cmd
func NewDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "delete",
		Long:  "delete resource",
	}
	cmd.PersistentFlags().BoolVarP(&flagErrIfNotExist, "errIfNotExist", "", false,
		"If true, return error when result is empty")

	cmd.AddCommand(NewDeleteStrategyCmd())
	return cmd
}

// NewDeleteStrategyCmd new deleteStrategy cmd
func NewDeleteStrategyCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "strategy",
		Short: "delete strategy",
		Long:  "delete strategy",
		Run: func(cmd *cobra.Command, args []string) {
			DeleteStrategy(cmd, args)
		},
	}
}

// DeleteStrategy delete strategy by name
func DeleteStrategy(cmd *cobra.Command, args []string) {
	operator := viper.GetString("config.operator")
	if operator == "" {
		fmt.Println("config.operator cannot be empty")
		os.Exit(1)
	}
	cli, ctx, err := pkg.NewClientWithConfiguration(cmd.Context())
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if len(args) == 0 || args[0] == "" {
		fmt.Println("strategy name must be specific")
		os.Exit(1)
	}
	req := &nodegroupmanager.DeleteNodePoolMgrStrategyReq{
		Operator: operator,
		Name:     args[0],
	}
	rsp, err := cli.DeleteNodePoolMgrStrategy(ctx, req)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println(rsp.Message)
}
