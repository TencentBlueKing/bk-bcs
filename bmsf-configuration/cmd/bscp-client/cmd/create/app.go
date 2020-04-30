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
	"bk-bscp/cmd/bscp-client/cmd/utils"
	"bk-bscp/cmd/bscp-client/option"
	"bk-bscp/cmd/bscp-client/service"
	"bk-bscp/internal/protocol/accessserver"
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

//createAppCmd: client create app
func createAppCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "application",
		Aliases: []string{"app"},
		Short:   "create application",
		Long:    "create new application under business",
		Example: `
	bscp-client create application --name gamesvc --type 1
	bscp-client create app -c nobody -n gamesvc -t 0
		`,
		RunE: handleCreateApp,
	}
	//todo(DeveloperJim): --file is required
	cmd.Flags().StringP("name", "n", "", "settings new application name")
	cmd.Flags().Int32P("type", "t", 0, "settings new application type, 0 is container, 1 is GSE")
	cmd.MarkFlagRequired("name")
	return cmd
}

func logicClusterCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "cluster",
		Aliases: []string{"clu"},
		Short:   "create cluster",
		Long:    "create new logic cluster for application",
		Example: `
	bscp-client create logiccluster --operator nobody --name defaultcluster --app gameserver
	bscp-client create cluster -c nobody -n defaultcluster -a gameserver
		`,
		RunE: handleCreateLogicClucster,
	}
	//options
	cmd.Flags().StringP("name", "n", "", "settings new cluster name")
	cmd.Flags().StringP("app", "a", "", "settings app that cluster belongs to ")
	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("app")
	return cmd
}

func clusterListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "cluster-list",
		Aliases: []string{"clu-list"},
		Short:   "create cluster-list",
		Long:    "create same cluster for multiple application",
		Example: `
	create cluster "defaultcluster" for Application gameserver, db, proxy, new-module
	bscp-client create cluster-list --name defaultcluster --for-apps gameserver,db,proxy --for-apps new-module
		`,
		RunE: handleCreateClusterList,
	}
	//options
	cmd.Flags().StringP("name", "n", "", "settings new cluster name")
	cmd.Flags().StringSlice("for-apps", []string{}, "settings app that cluster belongs to ")
	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("for-apps")
	return cmd
}

func createZoneCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "zone",
		Short: "create zone",
		Long:  "create new zone for specified application",
		Example: `
	bscp-client create zone --operator nobody --cluster defaultcluster --app gameserver --name zoneName
	bscp-client create zone -c nobody -l defaultcluster -a gameserver -n zoneName
		`,
		RunE: handleCreateZone,
	}
	//options
	cmd.Flags().StringP("name", "n", "", "settings new zone name")
	cmd.Flags().StringP("cluster", "l", "", "settings cluster that zone belongs to ")
	cmd.Flags().StringP("app", "a", "", "settings app that zone belongs to ")
	cmd.MarkFlagRequired("cluster")
	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("app")
	return cmd
}

func createZoneListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "zone-list",
		Aliases: []string{"z-list"},
		Short:   "create zone-list",
		Long:    "create same zone for multiple cluster with specified application",
		Example: `
create zone "myzone" for multiple app c1, c2, c3 with same cluster
  > bscp-client create zone-list --cluster shenzhen --for-apps c1,c2,c3 --name myzone

create multipe zone "zone1","zone2" for single cluster
  > bscp-client creat zone-list --cluster onecluster --app gameserver --names zone1,zone2
		`,
		RunE: handleCreateZoneList,
	}
	//options
	cmd.Flags().String("name", "", "settings new zone name")
	cmd.Flags().StringSlice("for-apps", []string{}, "settings app list that zone belongs to ")

	cmd.Flags().String("app", "", "settings app that zone belongs to ")
	cmd.Flags().StringSlice("names", []string{}, "settings new zone name list ")

	cmd.Flags().String("cluster", "", "settings cluster that zone belongs to ")
	cmd.MarkFlagRequired("cluster")
	return cmd
}

func handleCreateApp(cmd *cobra.Command, args []string) error {
	//get global command info and create business operator
	operator := service.NewOperator(option.GlobalOptions)
	if err := operator.Init(option.GlobalOptions.ConfigFile); err != nil {
		return err
	}
	//check option
	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return err
	}
	apptype, err := cmd.Flags().GetInt32("type")
	if err != nil {
		return err
	}
	createOption := &service.CreateAppOption{
		Name:    name,
		Creator: option.GlobalOptions.Operator,
		Type:    apptype,
	}
	//create application and check result
	appID, err := operator.CreateApp(context.TODO(), createOption)
	if err != nil {
		return err
	}
	cmd.Printf("create Application %s successfully.\n", appID)
	return nil
}

func handleCreateLogicClucster(cmd *cobra.Command, args []string) error {
	//get global command info and create business operator
	operator := service.NewOperator(option.GlobalOptions)
	if err := operator.Init(option.GlobalOptions.ConfigFile); err != nil {
		return err
	}
	//check option
	//cluster name
	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return err
	}
	//appName from command line flags
	appName, err := cmd.Flags().GetString("app")
	if err != nil {
		return err
	}
	business, app, err := utils.GetBusinessAndApp(operator, operator.Business, appName)
	if err != nil {
		return err
	}
	//check
	createOption := &accessserver.CreateClusterReq{
		Name:    name,
		Bid:     business.Bid,
		Appid:   app.Appid,
		Creator: option.GlobalOptions.Operator,
	}
	//create application and check result
	clusterID, err := operator.CreateLogicCluster(context.TODO(), createOption)
	if err != nil {
		return err
	}
	cmd.Printf("create Cluster %s/%s successfully.\n", clusterID, name)
	return nil
}

func handleCreateClusterList(cmd *cobra.Command, args []string) error {
	//get global command info and create business operator
	operator := service.NewOperator(option.GlobalOptions)
	if err := operator.Init(option.GlobalOptions.ConfigFile); err != nil {
		return err
	}
	//check option
	//cluster name
	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return err
	}
	//appName from command line flags
	appNameList, err := cmd.Flags().GetStringSlice("for-apps")
	if err != nil {
		return err
	}
	business, err := operator.GetBusiness(context.TODO(), operator.Business)
	if err != nil {
		return err
	}
	if business == nil {
		return fmt.Errorf("No relative Business %s", operator.Business)
	}
	for index, appName := range appNameList {
		app, err := operator.GetApp(context.TODO(), appName)
		if err != nil {
			cmd.Printf("%d: Create Cluster %s for App %s failed, %s\n", index, name, appName, err.Error())
			continue
		}
		if app == nil {
			cmd.Printf("%d: Create Cluster %s for App %s failed, No relative App!\n", index, name, appName)
			continue
		}
		//check
		createOption := &accessserver.CreateClusterReq{
			Name:    name,
			Bid:     business.Bid,
			Appid:   app.Appid,
			Creator: option.GlobalOptions.Operator,
		}
		//create application and check result
		clusterID, err := operator.CreateLogicCluster(context.TODO(), createOption)
		if err != nil {
			cmd.Printf("%d: Create Cluster %s for App %s failed, %s\n", index, name, appName, err.Error())
			return err
		}
		cmd.Printf("%d: create Cluster %s/%s for App %s successfully.\n", index, clusterID, name, appName)
	}
	return nil
}

func handleCreateZone(cmd *cobra.Command, args []string) error {
	//get global command info and create business operator
	operator := service.NewOperator(option.GlobalOptions)
	if err := operator.Init(option.GlobalOptions.ConfigFile); err != nil {
		return err
	}
	//zone name
	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return err
	}
	clusterName, err := cmd.Flags().GetString("cluster")
	if err != nil {
		return err
	}
	//appName from command line flags
	appName, err := cmd.Flags().GetString("app")
	if err != nil {
		return err
	}
	business, app, err := utils.GetBusinessAndApp(operator, operator.Business, appName)
	if err != nil {
		return err
	}
	cluster, err := operator.GetLogicCluster(context.TODO(), appName, clusterName)
	if err != nil {
		return err
	}
	if cluster == nil {
		return fmt.Errorf("No relative Cluster %s", clusterName)
	}
	request := &accessserver.CreateZoneReq{
		Bid:       business.Bid,
		Appid:     app.Appid,
		Clusterid: cluster.Clusterid,
		Name:      name,
		Creator:   option.GlobalOptions.Operator,
	}
	zoneID, err := operator.CreateZone(context.TODO(), request)
	if err != nil {
		return err
	}
	cmd.Printf("create Zone %s/%s successfully.\n", zoneID, name)
	return nil
}

func handleCreateZoneList(cmd *cobra.Command, args []string) error {
	//get global command info and create business operator
	operator := service.NewOperator(option.GlobalOptions)
	if err := operator.Init(option.GlobalOptions.ConfigFile); err != nil {
		return err
	}
	//check option
	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return err
	}
	//name is not empty, create zone for multiple Cluster
	if len(name) != 0 {
		return createZoneForMultipleAppInOneCluster(operator, cmd, name)
	}
	//create multiple zone for one cluster
	//appName from command line flags
	appName, err := cmd.Flags().GetString("app")
	if err != nil {
		return err
	}
	return createMultipleZoneForCluster(operator, cmd, appName)
}

func createZoneForMultipleAppInOneCluster(operator *service.AccessOperator, cmd *cobra.Command, name string) error {
	//check option
	clusterName, err := cmd.Flags().GetString("cluster")
	if err != nil {
		return err
	}
	if len(clusterName) == 0 {
		return fmt.Errorf("Lost params `cluster` in when using param `name`")
	}
	business, err := operator.GetBusiness(context.TODO(), operator.Business)
	if err != nil {
		return err
	}
	if business == nil {
		return fmt.Errorf("No relative Business %s", operator.Business)
	}
	appNameList, err := cmd.Flags().GetStringSlice("for-apps")
	if err != nil {
		return err
	}
	if len(appNameList) == 0 {
		return fmt.Errorf("Lost params `for-apps` in when using param `name`")
	}
	for index, appName := range appNameList {
		app, err := operator.GetApp(context.TODO(), appName)
		if err != nil {
			cmd.Printf("%d: Create Zone %s for App %s with cluster %s failed, %s\n", index, name, appName, clusterName, err.Error())
			continue
		}
		if app == nil {
			cmd.Printf("%d: Create Zone %s for App %s with cluster %s failed, No relative Application\n", index, name, appName, clusterName)
			continue
		}
		cluster, err := operator.GetLogicCluster(context.TODO(), appName, clusterName)
		if err != nil {
			cmd.Printf("%d: Create Zone %s for App %s with cluster %s failed, %s\n", index, name, appName, clusterName, err.Error())
			continue
		}
		if cluster == nil {
			cmd.Printf("%d: Create Zone %s for App %s with cluster %s failed, No relative Cluster\n", index, name, appName, clusterName)
			continue
		}
		request := &accessserver.CreateZoneReq{
			Bid:       business.Bid,
			Appid:     app.Appid,
			Clusterid: cluster.Clusterid,
			Name:      name,
			Creator:   option.GlobalOptions.Operator,
		}
		zoneID, err := operator.CreateZone(context.TODO(), request)
		if err != nil {
			cmd.Printf("%d: Create Zone %s for App %s failed, %s\n", index, name, appName, err.Error())
			continue
		}
		cmd.Printf("%d: create Zone %s/%s for App %s under cluster %s successfully.\n", index, zoneID, name, appName, clusterName)
	}
	return nil
}

func createMultipleZoneForCluster(operator *service.AccessOperator, cmd *cobra.Command, appName string) error {
	//check option
	nameList, err := cmd.Flags().GetStringSlice("names")
	if err != nil {
		return err
	}
	if len(nameList) == 0 {
		return fmt.Errorf("Lost params `names` list")
	}
	clusterName, err := cmd.Flags().GetString("cluster")
	if err != nil {
		return err
	}
	if len(clusterName) == 0 {
		return fmt.Errorf("Lost Params cluster")
	}
	business, err := operator.GetBusiness(context.TODO(), operator.Business)
	if err != nil {
		return err
	}
	if business == nil {
		return fmt.Errorf("No relative Business %s", operator.Business)
	}
	app, err := operator.GetApp(context.TODO(), appName)
	if err != nil {
		return err
	}
	if app == nil {
		return fmt.Errorf("No relative App %s", appName)
	}
	cluster, err := operator.GetLogicCluster(context.TODO(), appName, clusterName)
	if err != nil {
		return err
	}
	if cluster == nil {
		return fmt.Errorf("No relative Cluster %s", clusterName)
	}
	for index, zoneName := range nameList {
		request := &accessserver.CreateZoneReq{
			Bid:       business.Bid,
			Appid:     app.Appid,
			Clusterid: cluster.Clusterid,
			Name:      zoneName,
			Creator:   option.GlobalOptions.Operator,
		}
		zoneID, err := operator.CreateZone(context.TODO(), request)
		if err != nil {
			cmd.Printf("%d: Create Zone %s for App %s failed, %s\n", index, zoneName, appName, err.Error())
			continue
		}
		cmd.Printf("%d: create Zone %s/%s for App %s under cluster %s successfully.\n", index, zoneID, zoneName, appName, clusterName)
	}
	return nil
}
