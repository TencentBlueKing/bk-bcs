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
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-nodegroup-manager/cmd/client/cmd/printer"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-nodegroup-manager/cmd/client/pkg"
	nodegroupmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-nodegroup-manager/proto"
)

// NewGetCmd new get cmd
func NewGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "get",
		Long:  "get resources",
	}
	cmd.PersistentFlags().StringVarP(&flagOutput, "output", "o", "", "output format, one of json|wide")
	cmd.AddCommand(NewGetStrategyCmd())
	cmd.AddCommand(NewGetCaReviewCmd())
	return cmd
}

// NewGetStrategyCmd new getStrategy cmd
func NewGetStrategyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "strategy",
		Short: "get strategy",
		Long:  "get strategy",
		Run: func(cmd *cobra.Command, args []string) {
			GetStrategy(cmd, args)
		},
	}
	cmd.Flags().BoolVarP(&flagErrIfNotExist, "errIfNotExist", "", false,
		"If true, return error when result is empty")
	cmd.Flags().BoolVarP(&flagGetSoftDeleted, "getSoftDeleted", "", false,
		"If true, return the soft deleted item")
	cmd.Flags().Uint32VarP(&flagPage, "page", "p", 0, "list page")
	cmd.Flags().Uint32VarP(&flagLimit, "limit", "l", 10, "list limit")
	return cmd
}

// NewGetCaReviewCmd xxx
func NewGetCaReviewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "caReview",
		Short: "get caReview",
		Long:  "get caReview",
		Run: func(cmd *cobra.Command, args []string) {
			GetClusterAutoscalerReview(cmd, args)
		},
	}
	cmd.Flags().StringVarP(&flagFile, "file", "f", "", "get cluster autoscaler review by file")
	return cmd
}

// GetStrategy get strategy
func GetStrategy(cmd *cobra.Command, args []string) {
	cli, ctx, err := pkg.NewClientWithConfiguration(cmd.Context())
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	var response []*nodegroupmanager.NodeGroupStrategy
	switch len(args) {
	case 0:
		response = listStrategy(ctx, cli)
	default:
		response = getStrategy(ctx, cli, args[0])
	}
	if flagOutput == outputTypeJSON {
		printer.PrintStrategyInJSON(response)
		return
	}
	printer.PrintStrategyInTable(flagOutput == outputTypeWide, response)
}

func listStrategy(ctx context.Context,
	cli nodegroupmanager.NodegroupManagerClient) []*nodegroupmanager.NodeGroupStrategy {
	if flagLimit == 0 {
		flagLimit = 10
	}
	req := &nodegroupmanager.ListNodePoolMgrStrategyReq{
		Limit: flagLimit,
		Page:  flagPage,
	}
	rsp, err := cli.ListNodePoolMgrStrategies(ctx, req)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return rsp.Data
}

func getStrategy(ctx context.Context, cli nodegroupmanager.NodegroupManagerClient,
	name string) []*nodegroupmanager.NodeGroupStrategy {
	req := &nodegroupmanager.GetNodePoolMgrStrategyReq{
		Name: name,
	}
	rsp, err := cli.GetNodePoolMgrStrategy(ctx, req)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return []*nodegroupmanager.NodeGroupStrategy{rsp.Data}
}

// GetClusterAutoscalerReview get caReview
func GetClusterAutoscalerReview(cmd *cobra.Command, args []string) {
	file, err := os.ReadFile(flagFile)
	if err != nil {
		fmt.Printf("read file error:%v\n", err)
		os.Exit(1)
	}
	caReview := &nodegroupmanager.ClusterAutoscalerReview{}
	err = json.Unmarshal(file, caReview)
	if err != nil {
		fmt.Printf("unmarshal ca review error:%v\n", err)
		os.Exit(1)
	}
	cli, ctx, err := pkg.NewClientWithConfiguration(cmd.Context())
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	rsp, err := cli.GetClusterAutoscalerReview(ctx, caReview)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	printer.PrintCaReviewInJSON(rsp)
}
