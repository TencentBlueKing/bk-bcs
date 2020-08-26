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
	"github.com/spf13/cobra"

	"bk-bscp/cmd/bscp-client/option"
)

//getInstanceCmd: client get strategy
func getInstanceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "instance",
		Aliases: []string{"inst"},
		Short:   "Get instance details",
		Long:    "Get instance relative detail information",
		Example: `
	bscp-client get instance
		`,
		RunE: handleGetInstance,
	}
	// --name is required
	cmd.Flags().String("app", "", "the name of application")
	return cmd
}

func handleGetInstance(cmd *cobra.Command, args []string) error {
	err := option.SetGlobalVarByName(cmd, "app")
	if err != nil {
		return err
	}
	//get global command info and create business operator
	// operator := service.NewOperator(option.GlobalOptions)
	// if err := operator.Init(option.GlobalOptions.ConfigFile); err != nil {
	// 	return err
	// }
	// //check --name option
	// name, _ := cmd.Flags().GetString("name")
	// appName, _ := cmd.Flags().GetString("app")
	cmd.Printf("Do not implemented\n")
	// //create business and check result
	// strategy, err := operator.GetStrategy(context.TODO(), appName, name)
	// if err != nil {
	// 	return err
	// }
	// if strategy == nil {
	// 	cmd.Printf("Found no strategy resource.")
	// 	return nil
	// }
	// //format output
	// utils.PrintStrategy(strategy)
	return nil
}
