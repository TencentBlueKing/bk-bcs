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

package get

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"bk-bscp/cmd/bscp-client/cmd/utils"
	"bk-bscp/cmd/bscp-client/option"
	"bk-bscp/cmd/bscp-client/service"
	"bk-bscp/internal/protocol/common"
)

//getBusinessCmd: client create business
func getBusinessCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "business",
		Aliases: []string{"bus"},
		Hidden:  true,
		Short:   "Get business details",
		Long:    "Get business information",
		Example: `
	bk-bscp-client get business --id xxxxxxxx
	bk-bscp-client get business --name somegame
		`,
		RunE: handleGetBusiness,
	}
	// --name is required
	cmd.Flags().StringP("name", "n", "", "the name of business")
	cmd.Flags().StringP("id", "i", "", "the id of business")
	return cmd
}

func handleGetBusiness(cmd *cobra.Command, args []string) error {
	//get global command info and create business operator
	operator := service.NewOperator(option.GlobalOptions)
	if err := operator.Init(option.GlobalOptions.ConfigFile); err != nil {
		return err
	}
	// check flag
	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return err
	}
	id, err := cmd.Flags().GetString("id")
	if err != nil {
		return err
	}
	if len(name) == 0 && len(id) == 0 {
		return fmt.Errorf("%s %s or %s", option.ErrMsg_PARAM_MISS, "id", "name")
	}
	var business *common.Business
	// query by id
	if len(id) != 0 {
		business, err = operator.GetBusinessByID(context.TODO(), id)
	} else if len(name) != 0 { // query by name
		business, err = operator.GetBusiness(context.TODO(), name)
	}
	// check result
	if err != nil {
		return err
	}
	if business == nil {
		fmt.Printf("%s\n", option.SucMsg_DATA_NO_FOUNT)
		return nil
	}

	//format output
	utils.PrintBusiness(business)
	return nil
}

// getShardingDBCmd: client get shardingDB
func getShardingDBCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "shardingdb",
		Aliases: []string{"db"},
		Short:   "Get shardingDB details",
		Long:    "Get shardingDB information",
		Hidden:  true,
		Example: `
	bk-bscp-client get shardingdb --dbid xxxxxxxx
		 `,
		RunE: handleGetShardingDB,
	}
	cmd.Flags().StringP("dbid", "i", "", "the id of shardingDB")
	cmd.MarkFlagRequired("dbid")
	return cmd
}

func handleGetShardingDB(cmd *cobra.Command, args []string) error {
	//get global command info and create business operator
	operator := service.NewOperator(option.GlobalOptions)
	if err := operator.Init(option.GlobalOptions.ConfigFile); err != nil {
		return err
	}
	// check flag
	dbid, err := cmd.Flags().GetString("dbid")
	if err != nil {
		return err
	}
	shardingDB, err := operator.GetShardingDB(context.TODO(), dbid)
	if err != nil {
		return err
	}
	if shardingDB == nil {
		fmt.Printf("%s\n", option.SucMsg_DATA_NO_FOUNT)
		return nil
	}
	utils.PrintShardingDB(shardingDB)

	return nil
}

// getShardingCmd: client get sharding
func getShardingCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "sharding",
		Aliases: []string{"sd"},
		Short:   "Get sharding details",
		Long:    "Get sharding information",
		Hidden:  true,
		Example: `
	bk-bscp-client get sharding --key xxxxxxxx
		 `,
		RunE: handleGetSharding,
	}
	cmd.Flags().StringP("key", "k", "", "the key of sharding")
	cmd.MarkFlagRequired("key")
	return cmd
}

func handleGetSharding(cmd *cobra.Command, args []string) error {
	//get global command info and create business operator
	operator := service.NewOperator(option.GlobalOptions)
	if err := operator.Init(option.GlobalOptions.ConfigFile); err != nil {
		return err
	}
	// check flag
	key, err := cmd.Flags().GetString("key")
	if err != nil {
		return err
	}
	sharding, err := operator.GetSharding(context.TODO(), key)
	if err != nil {
		return err
	}
	if sharding == nil {
		fmt.Printf("%s\n", option.SucMsg_DATA_NO_FOUNT)
		return nil
	}
	utils.PrintSharding(sharding)

	return nil
}
