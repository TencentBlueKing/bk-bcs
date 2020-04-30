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

package create

import (
	"bk-bscp/cmd/bscp-client/option"
	"bk-bscp/cmd/bscp-client/service"
	"context"

	"github.com/spf13/cobra"
)

//createBusinessCmd: client create business
func createBusinessCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "business",
		Aliases: []string{"bu", "bi"},
		Short:   "create business",
		Long:    "create new business, only affected in administrator mode",
		Hidden:  true,
		Example: `
	bscp-client create business --file newbusiness.yaml
	bscp-client create business -f business.yaml
	yaml format template as followed:
		kind: bscp-business
		version: 0.1.1
		spec:
		name: lol
		deptID: 1201
		creator: MrMGXXXX
		memo: annotation
		db:
		#dbID sharding index,对应mysql instance
		dbID: bscp-default-sharding
		#对应mysql中不同数据库
		dbName: bscp-default
		#如果有以下信息说明是新建sharedingDB
		host: 127.0.0.1
		port: 3306
		user: mysql
		password: ${pwd}
		memo: information
		`,
		RunE: handleCreateBusiness,
	}
	// --file is required
	cmd.Flags().StringP("file", "f", "", "settings new business yaml file")
	cmd.MarkFlagRequired("file")
	return cmd
}

func handleCreateBusiness(cmd *cobra.Command, args []string) error {
	//get global command info and create business operator
	operator := service.NewOperator(option.GlobalOptions)
	if err := operator.Init(option.GlobalOptions.ConfigFile); err != nil {
		return err
	}
	//check --file option
	cfgFile, cfgErr := cmd.Flags().GetString("file")
	if cfgErr != nil {
		return cfgErr
	}
	createOption := &service.CreateBusinessOption{}
	if err := createOption.LoadConfig(cfgFile); err != nil {
		return err
	}
	//create business and check result
	businessID, err := operator.CreateBusiness(context.TODO(), createOption)
	if err != nil {
		return err
	}
	cmd.Printf("create business %s successfully.\n", businessID)
	return nil
}
