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

//listAppCmd: client list app
func listAppCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "application",
		Aliases: []string{"app"},
		Short:   "List application",
		Long:    "List all application information under business",
		RunE:    handleListApp,
	}
	return cmd
}

//listLogicClusterCmd: client list cluster
func listClusterCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "cluster",
		Aliases: []string{"clu"},
		Short:   "List cluster",
		Long:    "List all logic cluster information under application",
		Example: `
	bscp-client list cluster
	bscp-client list clu
		 `,
		RunE: handleListCluster,
	}
	cmd.Flags().StringP("app", "a", "", "settings app name for logic cluster")
	return cmd
}

//listZoneCmd: client list cluster
func listZoneCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "zone",
		Short: "List zone",
		Long:  "List all zone information under application, requires parameter cluster",
		Example: `
	bscp-client list zone --cluster logic
		`,
		RunE: handleListZone,
	}
	cmd.Flags().StringP("app", "a", "", "settings app name")
	cmd.Flags().StringP("cluster", "c", "", "settings cluster name")
	cmd.MarkFlagRequired("cluster")
	return cmd
}

func handleListApp(cmd *cobra.Command, args []string) error {
	//get global command info and create app operator
	operator := service.NewOperator(option.GlobalOptions)
	if err := operator.Init(option.GlobalOptions.ConfigFile); err != nil {
		return err
	}
	//create business and check result
	apps, err := operator.ListApps(context.TODO())
	if err != nil {
		return err
	}
	if apps == nil {
		cmd.Printf("Found no Apps resource.\n")
		return nil
	}
	//format output
	utils.PrintAppList(apps, operator.Business)
	return nil
}

func handleListCluster(cmd *cobra.Command, args []string) error {
	err := option.SetGlobalVarByName(cmd, "app")
	if err != nil {
		return err
	}
	//get global command info and create app operator
	operator := service.NewOperator(option.GlobalOptions)
	if err := operator.Init(option.GlobalOptions.ConfigFile); err != nil {
		return err
	}
	appName, _ := cmd.Flags().GetString("app")
	//create logic cluster and check result
	clusters, err := operator.ListLogicClusterByApp(context.TODO(), appName)
	if err != nil {
		return err
	}
	if clusters == nil {
		cmd.Printf("Found no cluster resource.\n")
		return nil
	}
	//format output
	utils.PrintClusters(clusters, operator.Business, appName)
	return nil
}

func handleListZone(cmd *cobra.Command, args []string) error {
	err := option.SetGlobalVarByName(cmd, "app")
	if err != nil {
		return err
	}
	//get global command info and create app operator
	operator := service.NewOperator(option.GlobalOptions)
	if err := operator.Init(option.GlobalOptions.ConfigFile); err != nil {
		return err
	}
	appName, _ := cmd.Flags().GetString("app")
	clusterName, _ := cmd.Flags().GetString("cluster")
	zones, err := operator.ListZones(context.TODO(), appName, clusterName)
	if err != nil {
		return err
	}
	if zones == nil {
		cmd.Printf("Found no zone resource.\n")
		return nil
	}
	//format output
	utils.PrintZones(zones, operator.Business, appName, clusterName)
	return nil
}
