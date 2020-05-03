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
	"bk-bscp/cmd/bscp-client/cmd/utils"
	"bk-bscp/cmd/bscp-client/option"
	"bk-bscp/cmd/bscp-client/service"
	"bk-bscp/internal/protocol/common"
	"context"

	"github.com/spf13/cobra"
)

//getBusinessCmd: client create business
func getBusinessCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "business",
		Aliases: []string{"bu", "bi"},
		Short:   "get business details",
		Long:    "get business information",
		Hidden:  true,
		Example: `
	bscp-client get business --name somegame
	bscp-client get business -n somegame
		`,
		RunE: handleGetBusiness,
	}
	// --name is required
	cmd.Flags().StringP("name", "n", "", "the name of business")
	cmd.MarkFlagRequired("name")
	return cmd
}

func handleGetBusiness(cmd *cobra.Command, args []string) error {
	//get global command info and create business operator
	operator := service.NewOperator(option.GlobalOptions)
	if err := operator.Init(option.GlobalOptions.ConfigFile); err != nil {
		return err
	}
	//check --file option
	businessName, cfgErr := cmd.Flags().GetString("name")
	if cfgErr != nil {
		return cfgErr
	}
	//create business and check result
	business, err := operator.GetBusiness(context.TODO(), businessName)
	if err != nil {
		return err
	}
	if business == nil {
		cmd.Printf("Found no Business resource.\n")
		return nil
	}
	//format output
	utils.PrintBusiness([]*common.Business{business})
	return nil
}
