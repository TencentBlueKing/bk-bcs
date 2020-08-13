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
 *
 */

package list

import (
	"bk-bscp/cmd/bscp-client/cmd/utils"
	"bk-bscp/cmd/bscp-client/option"
	"bk-bscp/cmd/bscp-client/service"
	"context"

	"github.com/spf13/cobra"
)

//listBusinessCmd: client create business
func listBusinessCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "business",
		Aliases: []string{"bus"},
		Long:    "List all business information",
		Hidden:  true,
		Example: `
	bk-bscp-client list business
		 `,
		RunE: handleListBusiness,
	}
	return cmd
}

func handleListBusiness(cmd *cobra.Command, args []string) error {
	//get global command info and create business operator
	operator := service.NewOperator(option.GlobalOptions)
	if err := operator.Init(option.GlobalOptions.ConfigFile); err != nil {
		return err
	}
	//create business and check result
	businesses, err := operator.ListBusiness(context.TODO())
	if err != nil {
		return err
	}
	if businesses == nil {
		cmd.Printf("Found no Business resource.\n")
		return nil
	}
	//format output
	utils.PrintBusinesses(businesses)
	return nil
}

//listShardingDBCmd: list all shardingDB information
func listShardingDBCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "shardingdb",
		Aliases: []string{"db"},
		Short:   "List shardingDB",
		Long:    "List all shardingDB information",
		Hidden:  true,
		Example: `
	bk-bscp-client list shardingdb
		 `,
		RunE: handleListShardingDB,
	}
	return cmd
}

func handleListShardingDB(cmd *cobra.Command, args []string) error {
	//get global command info and create business operator
	operator := service.NewOperator(option.GlobalOptions)
	if err := operator.Init(option.GlobalOptions.ConfigFile); err != nil {
		return err
	}
	//create business and check result
	dbList, err := operator.ListShardingDB(context.TODO())
	if err != nil {
		return err
	}
	if dbList == nil {
		cmd.Printf("Found no ShardingDB resource.\n")
		return nil
	}
	//format output
	utils.PrintShardingDBList(dbList)
	return nil
}
