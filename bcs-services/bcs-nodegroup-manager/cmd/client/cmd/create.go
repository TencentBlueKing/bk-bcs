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
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-nodegroup-manager/cmd/client/pkg"
	nodegroupmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-nodegroup-manager/proto"
)

// NewCreateCmd new create cmd
func NewCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "create",
		Long:  "create resource",
	}
	cmd.PersistentFlags().StringVarP(&flagFile, "file", "f", "", "create item by file")

	cmd.AddCommand(NewCreateStrategyCmd())
	return cmd
}

// NewCreateStrategyCmd new createStrategy cmd
func NewCreateStrategyCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "strategy",
		Short: "create strategy -f FILENAME",
		Long:  "create strategy --file FILENAME",
		Run: func(cmd *cobra.Command, args []string) {
			CreateStrategy(cmd, args)
		},
	}
}

// CreateStrategy create strategy by file
func CreateStrategy(cmd *cobra.Command, args []string) {
	operator := viper.GetString("config.operator")
	if operator == "" {
		fmt.Println("config.operator cannot be empty")
		os.Exit(1)
	}
	file, err := os.ReadFile(flagFile)
	if err != nil {
		fmt.Printf("read file error:%v\n", err)
		os.Exit(1)
	}
	strategy := &nodegroupmanager.NodeGroupStrategy{}
	err = json.Unmarshal(file, strategy)
	if err != nil {
		fmt.Printf("unmarshal strategy error:%v\n", err)
		os.Exit(1)
	}
	req := &nodegroupmanager.CreateNodePoolMgrStrategyReq{
		Option: &nodegroupmanager.CreateOptions{
			OverWriteIfExist: false,
			Operator:         operator,
		},
		Strategy: strategy,
	}
	cli, ctx, err := pkg.NewClientWithConfiguration(cmd.Context())
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	rsp, err := cli.CreateNodePoolMgrStrategy(ctx, req)
	if err != nil {
		fmt.Printf("create strategy error:%v\n", err)
		os.Exit(1)
	}
	fmt.Printf(rsp.Message)
}
