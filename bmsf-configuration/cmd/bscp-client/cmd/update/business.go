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

package update

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"bk-bscp/cmd/bscp-client/option"
	"bk-bscp/cmd/bscp-client/service"
)

func updateSharingDBCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "shardingdb",
		Aliases: []string{"db"},
		Short:   "Update shardingDB details",
		Long:    "Update shardingDB information",
		Hidden:  true,
		Example: `
	bk-bscp-client update shardingdb --dbid xxxxxxxx --host 127.0.0.1 --port 3306 --user guohu --password admin --memo "this is a example"
		 `,
		RunE: handleUpdateShardingDB,
	}
	cmd.Flags().StringP("dbid", "i", "", "the dbid of shardingDB")
	cmd.Flags().StringP("host", "", "", "the host of shardingDB")
	cmd.Flags().Int32P("port", "", 0, "the port of shardingDB")
	cmd.Flags().StringP("user", "", "", "the user of shardingDB")
	cmd.Flags().StringP("password", "", "", "the password of shardingDB")
	cmd.Flags().StringP("memo", "m", "", "the memo of shardingDB")
	cmd.MarkFlagRequired("dbid")
	return cmd
}

func handleUpdateShardingDB(cmd *cobra.Command, args []string) error {
	//get global command info and create business operator
	operator := service.NewOperator(option.GlobalOptions)
	if err := operator.Init(option.GlobalOptions.ConfigFile); err != nil {
		return err
	}
	// check flags
	dbid, err := cmd.Flags().GetString("dbid")
	if err != nil {
		return err
	}
	host, err := cmd.Flags().GetString("host")
	if err != nil {
		return err
	}
	port, err := cmd.Flags().GetInt32("port")
	if err != nil {
		return err
	}
	user, err := cmd.Flags().GetString("user")
	if err != nil {
		return err
	}
	password, err := cmd.Flags().GetString("password")
	if err != nil {
		return err
	}
	memo, err := cmd.Flags().GetString("memo")
	if err != nil {
		return err
	}
	// update
	request := &service.ShardingDBOption{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		Memo:     memo,
	}
	err = operator.UpdateShardingDB(context.TODO(), dbid, request)
	if err != nil {
		return err
	}
	fmt.Printf("%s\n", option.SucMsg_DATA_UPDATE)
	return nil
}

func updateSharingCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "sharding",
		Aliases: []string{"sd"},
		Short:   "Update sharding details",
		Long:    "Update sharding information",
		Hidden:  true,
		Example: `
	bk-bscp-client update sharding --key xxxxxxxx --dbid xxxxxxxx --dbname testDB --memo "this is a example"
		 `,
		RunE: handleUpdateSharding,
	}
	cmd.Flags().StringP("dbid", "", "", "the dbid of sharding")
	cmd.Flags().StringP("key", "", "", "the key of sharding")
	cmd.Flags().StringP("dbname", "", "", "the dbname of sharding")
	cmd.Flags().StringP("memo", "m", "", "the memo of sharding")
	cmd.MarkFlagRequired("key")
	return cmd
}

func handleUpdateSharding(cmd *cobra.Command, args []string) error {
	//get global command info and create business operator
	operator := service.NewOperator(option.GlobalOptions)
	if err := operator.Init(option.GlobalOptions.ConfigFile); err != nil {
		return err
	}
	// check flags
	dbid, err := cmd.Flags().GetString("dbid")
	if err != nil {
		return err
	}
	key, err := cmd.Flags().GetString("key")
	if err != nil {
		return err
	}
	dbname, err := cmd.Flags().GetString("dbname")
	if err != nil {
		return err
	}
	memo, err := cmd.Flags().GetString("memo")
	if err != nil {
		return err
	}
	// update
	request := &service.ShardingOption{
		DBID:   dbid,
		DbName: dbname,
		Memo:   memo,
	}
	err = operator.UpdateSharding(context.TODO(), key, request)
	if err != nil {
		return err
	}
	fmt.Printf("%s\n", option.SucMsg_DATA_UPDATE)
	return nil
}
