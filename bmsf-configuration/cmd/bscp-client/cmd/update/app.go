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

func updateAppCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "application",
		Aliases: []string{"app"},
		Short:   "Update application",
		Long:    "Update specific application information by the id of application",
		Example: `
	bk-bscp-client update application --id xxxxxxxx --name game --type 1 --memo "update app info"
	bk-bscp-client update application -i xxxxxxxx -n game -t 1 -m "update app info"
		`,
		RunE: handleUpdateApp,
	}
	cmd.Flags().StringP("id", "i", "", "the id of application to update")
	cmd.Flags().StringP("name", "n", "", "the new name of the application update")
	cmd.Flags().Int32P("type", "t", -1, "the new type of the application update, 0 is container, 1 is process")
	//cmd.Flags().Int32P("status", "s", -1, "the new status of the application update, 0 is Affectived, 1 is Deleted")
	cmd.Flags().StringP("memo", "m", "", "the memo of application update")
	cmd.MarkFlagRequired("id")
	return cmd
}

func handleUpdateApp(cmd *cobra.Command, args []string) error {
	//get global command info and create business operator
	operator := service.NewOperator(option.GlobalOptions)
	if err := operator.Init(option.GlobalOptions.ConfigFile); err != nil {
		return err
	}
	// check flags
	appId, err := cmd.Flags().GetString("id")
	if err != nil {
		return err
	}
	appName, err := cmd.Flags().GetString("name")
	if err != nil {
		return err
	}
	appType, err := cmd.Flags().GetInt32("type")
	if err != nil {
		return err
	}
	//appStatus, err := cmd.Flags().GetInt32("status")
	//if err != nil {
	//	return err
	//}
	memo, err := cmd.Flags().GetString("memo")
	if err != nil {
		return err
	}

	if len(appName) == 0 && len(memo) == 0 && appType == -1 {
		return fmt.Errorf("%s", option.ErrMsg_PARAM_MISS)
	}
	// update
	request := &service.UpdateAppOption{
		Name: appName,
		Type: appType,
		//Status: appStatus,
		Memo: memo,
	}
	err = operator.UpdateApp(context.TODO(), appId, request)
	// check result
	if err != nil {
		return err
	}
	fmt.Printf("%s\n", option.SucMsg_DATA_UPDATE)
	return nil
}

func updateClusterCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "cluster",
		Aliases: []string{"clu"},
		Short:   "Update cluster",
		Long:    "Update specific cluster information by the id of cluster",
		Example: `
	bk-bscp-client update cluster --id xxxxxxxx --name cluster1 --memo "update cluster info"
	bk-bscp-client update cluster -i xxxxxxxx -n cluster1 -m "update cluster info"
		`,
		RunE: handleUpdateCluster,
	}
	cmd.Flags().StringP("id", "i", "", "the id of cluster to update")
	cmd.Flags().StringP("name", "n", "", "the new name of the cluster update")
	//cmd.Flags().Int32P("status", "s", -1, "the new status of the cluster update, 0 is Affectived, 1 is Deleted")
	cmd.Flags().StringP("memo", "m", "", "the memo of cluster update")
	// labels and rclusterid
	cmd.MarkFlagRequired("id")
	return cmd
}

func handleUpdateCluster(cmd *cobra.Command, args []string) error {
	//get global command info and create business operator
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
	//clusterStatus, err := cmd.Flags().GetInt32("status")
	//if err != nil {
	//	return err
	//}
	memo, err := cmd.Flags().GetString("memo")
	if err != nil {
		return err
	}
	if len(clusterName) == 0 && len(memo) == 0 {
		return fmt.Errorf("%s", option.ErrMsg_PARAM_MISS)
	}

	// update
	request := &service.UpdateClusterOption{
		Name: clusterName,
		Memo: memo,
		//Status: clusterStatus,
	}
	err = operator.UpdateCluster(context.TODO(), clusterId, request)
	// check result
	if err != nil {
		return err
	}
	fmt.Printf("%s\n", option.SucMsg_DATA_UPDATE)
	return nil
}

func updateZoneCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "zone",
		Short: "Update zone",
		Long:  "Update specific zone information by the id of zone",
		Example: `
	bk-bscp-client update zone --id xxxxxxxx --name zone-tel-1 --memo "update zone info"
	bk-bscp-client update zone -i xxxxxxxx -n zone-tel-1 -m "update zone info"
		`,
		RunE: handleUpdateZone,
	}
	cmd.Flags().StringP("id", "i", "", "the id of zone to update")
	cmd.Flags().StringP("name", "n", "", "the new name of the zone update")
	//cmd.Flags().Int32P("status", "s", -1, "the new status of the zone update, 0 is Affectived, 1 is Deleted")
	cmd.Flags().StringP("memo", "m", "", "the memo of zone update")
	// labels and rclusterid
	cmd.MarkFlagRequired("id")
	return cmd
}

func handleUpdateZone(cmd *cobra.Command, args []string) error {
	//get global command info and create business operator
	operator := service.NewOperator(option.GlobalOptions)
	if err := operator.Init(option.GlobalOptions.ConfigFile); err != nil {
		return err
	}
	// check flags
	zoneId, err := cmd.Flags().GetString("id")
	if err != nil {
		return err
	}
	zoneName, err := cmd.Flags().GetString("name")
	if err != nil {
		return err
	}
	//zoneStatus, err := cmd.Flags().GetInt32("status")
	//if err != nil {
	//	return err
	//}
	memo, err := cmd.Flags().GetString("memo")
	if err != nil {
		return err
	}

	if len(zoneName) == 0 && len(memo) == 0 {
		return fmt.Errorf("%s", option.ErrMsg_PARAM_MISS)
	}
	// update
	request := &service.UpdateZoneOption{
		Name: zoneName,
		Memo: memo,
		//Status: zoneStatus,
	}
	err = operator.UpdateZone(context.TODO(), zoneId, request)
	// check result
	if err != nil {
		return err
	}
	fmt.Printf("%s\n", option.SucMsg_DATA_UPDATE)
	return nil
}
