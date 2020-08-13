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

//getApplicationCmd: client get strategy
func getApplicationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "application",
		Aliases: []string{"app"},
		Short:   "Get application details",
		Long:    "Get application details under current business",
		Example: `
	bscp-client get application --id xxxxxxxx
	bscp-client get application --name gameserver
		`,
		RunE: handleGetApplication,
	}
	// --name is required
	cmd.Flags().StringP("id", "i", "", "the id of application")
	cmd.Flags().StringP("name", "n", "", "the name of application")
	return cmd
}

func handleGetApplication(cmd *cobra.Command, args []string) error {
	// get global command info and create business operator.
	operator := service.NewOperator(option.GlobalOptions)
	if err := operator.Init(option.GlobalOptions.ConfigFile); err != nil {
		return err
	}

	// check flags
	appId, err := cmd.Flags().GetString("id")
	if err != nil {
		return err
	}
	// check flags
	appName, err := cmd.Flags().GetString("name")
	if err != nil {
		return err
	}
	if len(appId) == 0 && len(appName) == 0 {
		return fmt.Errorf("query application, id or name is required")
	}
	business, err := operator.GetBusiness(context.TODO(), operator.Business)
	if err != nil {
		return err
	}
	if business == nil {
		return fmt.Errorf("No relative Business %s Resource", operator.Business)
	}
	var app *common.App
	if len(appId) != 0 {
		app, err = operator.GetAppByAppID(context.TODO(), business.Bid, appId)
	} else if len(appName) != 0 {
		app, err = operator.GetAppByID(context.TODO(), business.Bid, appName)
	}
	if err != nil {
		return err
	}
	if app != nil {
		utils.PrintApplication(operator.Business, app)
	} else {
		cmd.Printf("No application resource is queried by %s%s\n", appId, appName)
	}
	return nil
}

//getClusterCmd: client get strategy
func getClusterCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "cluster",
		Aliases: []string{"clu"},
		Short:   "Get cluster details",
		Long:    "Get cluster details under current application",
		Example: `
	bscp-client get cluster --id xxxxxxxx
	bscp-client get cluster --app game --name cluster-shenzhen
		`,
		RunE: handleGetCluster,
	}
	// --name is required
	cmd.Flags().StringP("id", "i", "", "the id of cluster")
	cmd.Flags().StringP("name", "n", "", "the name of cluster")
	cmd.Flags().StringP("app", "a", "", "application which the cluster belongs to")
	return cmd
}

func handleGetCluster(cmd *cobra.Command, args []string) error {
	option.SetGlobalVarByName(cmd, "app")
	// get global command info and create business operator.
	operator := service.NewOperator(option.GlobalOptions)
	if err := operator.Init(option.GlobalOptions.ConfigFile); err != nil {
		return err
	}
	// check flags
	clusterId, err := cmd.Flags().GetString("id")
	if err != nil {
		return err
	}
	clusterName, err := cmd.Flags().GetString("name")
	if err != nil {
		return err
	}
	if len(clusterId) == 0 && len(clusterName) == 0 {
		return fmt.Errorf("query cluster, id or name is required")
	}

	// query bus
	business, err := operator.GetBusiness(context.TODO(), operator.Business)
	if err != nil {
		return err
	}
	if business == nil {
		return fmt.Errorf("No relative Business %s Resource", operator.Business)
	}

	var cluster *common.Cluster
	if len(clusterId) != 0 {
		// query cluster
		cluster, err = operator.GetClusterAllByID(context.TODO(), business.Bid, "", clusterId)
	} else if len(clusterName) != 0 {
		appName, _ := cmd.Flags().GetString("app")
		if len(appName) == 0 {
			return fmt.Errorf("query cluster by name need the name of application which the cluster belongs to")
		}
		cluster, err = operator.GetLogicCluster(context.TODO(), appName, clusterName)
	}
	if err != nil {
		return err
	}
	if cluster != nil {
		app, _ := operator.GetAppByAppID(context.TODO(), cluster.Bid, cluster.Appid)
		utils.PrintCluster(operator.Business, app.Name, cluster)
	} else {
		cmd.Printf("No cluster resource is queried by %s%s\n", clusterId, clusterName)
	}
	return nil
}

//getClusterCmd: client get strategy
func getZoneCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "zone",
		Short: "Get zone details",
		Long:  "Get zone details under current cluster",
		Example: `
	bk-bscp-client get zone --name zone-tel-1
	bk-bscp-client get zone --id xxxxxxxx
		`,
		RunE: handleGetZone,
	}
	// --name is required
	cmd.Flags().StringP("id", "i", "", "the id of zone")
	cmd.Flags().StringP("name", "n", "", "the name of zone")
	cmd.Flags().StringP("app", "a", "", "application which the zone belongs to")
	return cmd
}

func handleGetZone(cmd *cobra.Command, args []string) error {
	// check flags judge query func
	zoneId, err := cmd.Flags().GetString("id")
	if err != nil {
		return err
	}
	zoneName, err := cmd.Flags().GetString("name")
	if err != nil {
		return err
	}

	if len(zoneId) == 0 && len(zoneName) == 0 {
		return fmt.Errorf("%s %s or %s", option.ErrMsg_PARAM_MISS, "id", "name")
	}
	// query By id
	if len(zoneId) != 0 {
		return handleGetZoneById(zoneId)
	}
	// query By Name
	if len(zoneName) != 0 {
		// 三级获取 appName
		option.SetGlobalVarByName(cmd, "app")
		appName, err := cmd.Flags().GetString("app")
		if err != nil {
			return nil
		}
		if len(appName) == 0 {
			return fmt.Errorf("%s %s", option.ErrMsg_PARAM_MISS, "app")
		}
		return handleGetZoneByName(appName, zoneName)
	}
	return nil
}

func handleGetZoneByName(appName, zoneName string) error {
	// get global command info and create business operator.
	operator := service.NewOperator(option.GlobalOptions)
	if err := operator.Init(option.GlobalOptions.ConfigFile); err != nil {
		return err
	}
	zone, err := operator.GetZoneByName(context.TODO(), appName, zoneName)
	if err != nil {
		return err
	}
	if zone == nil {
		fmt.Printf("%s\n", option.SucMsg_DATA_NO_FOUNT)
		return nil
	}
	cluster, _ := operator.GetClusterAllByID(context.TODO(), zone.Bid, "", zone.Clusterid)
	utils.PrintZone(operator.Business, appName, cluster.Name, zone)
	return nil
}

func handleGetZoneById(zoneId string) error {
	// get global command info and create business operator.
	operator := service.NewOperator(option.GlobalOptions)
	if err := operator.Init(option.GlobalOptions.ConfigFile); err != nil {
		return err
	}
	business, err := operator.GetBusiness(context.TODO(), operator.Business)
	if err != nil {
		return err
	}
	if business == nil {
		return fmt.Errorf("No relative Business %s Resource", operator.Business)
	}
	zone, err := operator.GetZoneAllByID(context.TODO(), business.Bid, "", zoneId)
	if err != nil {
		return err
	}
	if zone == nil {
		fmt.Printf("%s\n", option.SucMsg_DATA_NO_FOUNT)
		return nil
	}
	app, _ := operator.GetAppByAppID(context.TODO(), zone.Bid, zone.Appid)
	cluster, _ := operator.GetClusterAllByID(context.TODO(), business.Bid, "", zone.Clusterid)
	utils.PrintZone(operator.Business, app.Name, cluster.Name, zone)
	return nil
}
