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
	"bk-bscp/internal/protocol/common"
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

//listAppInstCmd: client list release
func listAppInstCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "instance",
		Aliases: []string{"inst"},
		Short:   "list Application instances",
		Long:    "list all application instance information",
		Example: `
List all Application Instance information by Logic concept:
	bscp-client list instance  --app gameserver --cluster name --zone zname

List Application instance that specified Release published to
	bscp-client list inst --app gamesvr --cfgset csetname --releaseId
		 `,
		RunE: handleListAppInst,
	}
	//simple list by logic concept
	cmd.Flags().String("app", "", "application name for filter")
	cmd.Flags().String("cluster", "", "cluster name for filter")
	cmd.Flags().String("zone", "", "zone name for filter")

	//list instance by specified published release
	cmd.Flags().String("cfgset", "", "ConfigSet Name that use for instance filter")
	cmd.Flags().String("releaseId", "", "release filter for instance")

	return cmd
}

func handleListAppInst(cmd *cobra.Command, args []string) error {
	//get global command info and create app operator
	operator := service.NewOperator(option.GlobalOptions)
	if err := operator.Init(option.GlobalOptions.ConfigFile); err != nil {
		return err
	}
	//check if list instance by Release
	releaseID, _ := cmd.Flags().GetString("releaseId")
	if len(releaseID) != 0 {
		return handleListInstByRelease(cmd, operator, releaseID)
	}
	//other list
	return simpleListAppInstance(cmd, operator)
}

func handleListInstByRelease(cmd *cobra.Command, operator *service.AccessOperator, releaseID string) error {
	cfgsetName, _ := cmd.Flags().GetString("cfgset")
	if len(cfgsetName) == 0 {
		return fmt.Errorf("Lost ConfigSetName when list by ReleaseId")
	}
	appName, _ := cmd.Flags().GetString("app")
	if len(appName) == 0 {
		return fmt.Errorf("Lost AppName when list by ReleaseId")
	}
	//list all datas if exists
	instances, err := operator.ListInstanceByRelease(context.TODO(), appName, cfgsetName, releaseID)
	if err != nil {
		return err
	}
	if instances == nil {
		cmd.Printf("Found no Application Instance resource.\n")
		return nil
	}
	clusterMap, zoneMap := handleInstanceClusterAndZone(operator, instances)
	//format output
	utils.PrintAppInstances(instances, operator.Business, appName, clusterMap, zoneMap)
	return nil
}

func simpleListAppInstance(cmd *cobra.Command, operator *service.AccessOperator) error {
	appName, _ := cmd.Flags().GetString("app")
	clusterName, _ := cmd.Flags().GetString("cluster")
	zoneName, _ := cmd.Flags().GetString("zone")
	request := &service.InstanceOption{
		AppName:     appName,
		ClusterName: clusterName,
		ZoneName:    zoneName,
	}
	//list all datas if exists
	instances, err := operator.ListAppInstance(context.TODO(), request)
	if err != nil {
		return err
	}
	if instances == nil {
		cmd.Printf("Found no Application Instance resource.\n")
		return nil
	}
	clusterMap, zoneMap := handleInstanceClusterAndZone(operator, instances)
	//format output
	utils.PrintAppInstances(instances, operator.Business, appName, clusterMap, zoneMap)
	return nil
}

func handleInstanceClusterAndZone(operator *service.AccessOperator, insts []*common.AppInstance) (map[string]*common.Cluster, map[string]*common.Zone) {
	clusterMap := make(map[string]*common.Cluster)
	zoneMap := make(map[string]*common.Zone)
	for _, inst := range insts {
		if _, ok := zoneMap[inst.Zoneid]; !ok {
			zone, err := operator.GetZoneAllByID(context.TODO(), inst.Bid, inst.Appid, inst.Zoneid)
			if err != nil {
				continue
			}
			if zone == nil {
				continue
			}
			zoneMap[inst.Zoneid] = zone
		}
		if _, ok := clusterMap[inst.Clusterid]; !ok {
			//search new cluster
			cluster, err := operator.GetClusterAllByID(context.TODO(), inst.Bid, inst.Appid, inst.Clusterid)
			if err != nil {
				continue
			}
			if cluster == nil {
				continue
			}
			clusterMap[inst.Clusterid] = cluster
		}
	}
	return clusterMap, zoneMap
}
