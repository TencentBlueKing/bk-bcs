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
	"context"
	"fmt"
	"path"

	"github.com/spf13/cobra"

	"bk-bscp/cmd/bscp-client/cmd/utils"
	"bk-bscp/cmd/bscp-client/option"
	"bk-bscp/cmd/bscp-client/service"
	"bk-bscp/internal/protocol/common"
)

const (
	OnlineEffectReloadInstance  = "onlineEffectReloadInstance"
	OnlineEffectInstance        = "onlineEffectInstance"
	OnlineUnEffectInstance      = "onlineUnEffectInstance"
	OfflineEffectReloadInstance = "offlineEffectReloadInstance"
	OfflineEffectInstance       = "offlineEffectInstance"
)

//listAppInstCmd: client list release
func listAppInstCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "instance",
		Aliases: []string{"inst"},
		Short:   "List Application instance",
		Long:    "List all application instance information",
		Example: `
List all Application Instance information by Logic concept:
	bk-bscp-client list instance --cluster cname --zone zname --status 1

List all Application instance that specified strategy can match
	bk-bscp-client list instance --strategyid xxxxxxxx

List all Application instance that specified release can match and effected
	bk-bscp-client list instance --releaseid xxxxxxxx (--status 0)
		 `,
		RunE: handleListAppInst,
	}
	//simple list by logic concept
	cmd.Flags().StringP("app", "a", "", "application name for filter")
	cmd.Flags().StringP("cluster", "c", "", "cluster name for filter")
	cmd.Flags().StringP("zone", "z", "", "zone name for filter")
	cmd.Flags().Int32P("status", "s", 1, "instance status for filter,0 is All, 1 is ONLINE, 2 is OFFLINE, default 1 is ONLINE")

	//list instance by specified published release
	cmd.Flags().StringP("releaseid", "", "", "instance matching in the specified release")
	cmd.Flags().StringP("strategyid", "", "", "strategy matching in the specified release")
	return cmd
}

// handleListAppInst judge handle by strategy\release
func handleListAppInst(cmd *cobra.Command, args []string) error {
	//get global command info and create app operator
	operator := service.NewOperator(option.GlobalOptions)
	if err := operator.Init(option.GlobalOptions.ConfigFile); err != nil {
		return err
	}

	//check flags
	multiReleaseID, err := cmd.Flags().GetString("releaseid")
	if err != nil {
		return err
	}
	strategyID, err := cmd.Flags().GetString("strategyid")
	if err != nil {
		return err
	}
	if len(strategyID) != 0 && len(multiReleaseID) != 0 {
		return fmt.Errorf("Invalid params, only releaseid or strategyid")
	}

	// query match instances by strategy
	if len(strategyID) != 0 {
		return handleMatchAppInstByStrategy(cmd, operator, strategyID)
	}

	// query all instances by releaseId ( 1. online-effected-reload 2. online-effected 3. online 4. offline-effected-reload 5. offline-effected-reload)
	if len(multiReleaseID) != 0 {
		return handleAllStatusInstacesByRelease(cmd, operator, multiReleaseID)
	}

	// defult query insts by business and app
	return simpleListAppInstance(cmd, operator)
}

//
func handleAllStatusInstacesByRelease(cmd *cobra.Command, operator *service.AccessOperator, multiReleaseID string) error {
	// get multiRelease and check
	multiRelease, metaDatas, err := operator.GetMultiRelease(context.TODO(), multiReleaseID)
	if err != nil {
		return err
	}
	if multiRelease == nil {
		cmd.Printf("%s - release\n", option.SucMsg_DATA_NO_FOUNT)
		return nil
	}
	queryType, err := cmd.Flags().GetInt32("status")
	if err != nil {
		return err
	}

	// get application
	app, err := operator.GetAppByAppID(context.TODO(), multiRelease.Bid, multiRelease.Appid)
	if err != nil {
		return err
	}

	// judge instance is exist
	flag := true
	// query instance from match and effected interface
	for _, metaData := range metaDatas {
		// get match instance
		matchedInstances, err := operator.ListMatchedInstanceByReleaseId(context.TODO(), multiRelease.Bid, metaData.Releaseid)
		if err != nil {
			return err
		}
		// get effected instanct
		effectedInstances, err := operator.ListEffectedInstance(context.TODO(), multiRelease.Bid, metaData.Cfgsetid, metaData.Releaseid)
		if err != nil {
			return err
		}
		classificationInstance, categoryTotal := getInstanceAfterClassification(matchedInstances, effectedInstances)
		if classificationInstance != nil {
			flag = false
			configSet, err := operator.GetConfigSet(context.TODO(), app.Name, &common.ConfigSet{Cfgsetid: metaData.Cfgsetid})
			if err != nil {
				return err
			}
			clusterMap, zoneMap := handleInstanceClusterAndZone(operator, classificationInstance)
			//format output
			utils.PrintAppInstancesByRelease(classificationInstance, operator.Business, app.Name, clusterMap, zoneMap, queryType)
			fmt.Printf("    ConfigSet: %s    OnlineReloadInstance: %d    OnlineEffectInstance: %d    OnlineUnEffectInstance: %d    \n\t\t\tOfflineReloadInstance: %d    OfflineEffectInstance: %d\n", path.Clean(configSet.Fpath+"/"+configSet.Name),
				categoryTotal[OnlineEffectReloadInstance], categoryTotal[OnlineEffectInstance], categoryTotal[OnlineUnEffectInstance], categoryTotal[OfflineEffectReloadInstance], categoryTotal[OfflineEffectInstance])
			cmd.Println()
		}
	}
	if flag {
		fmt.Println("No uneffected instance resource found for the specified published releaseid")
	}
	return nil
}

// getInstanceAfterClassification Sort by category:
// 1. onlineEffectReloadInstance
// 2. onlineEffectInstance
// 3. onlineUneffectInstance
// 4. offlineEffectReloadInstance
// 5. offlineEffectInstance
func getInstanceAfterClassification(matchInstances []*common.AppInstance, effectedInstances []*common.AppInstance) ([]*common.AppInstance, map[string]int) {
	var result []*common.AppInstance
	categoryTotal := make(map[string]int)
	matchInstancesmap := sliceToMap(matchInstances)
	effectedInstancesMap := sliceToMap(effectedInstances)

	// 通过 matchInstance and effectedInstace 划分为 未生效、生效在线、生效下线 三类
	onlineEffectedInstances := make(map[uint64]*common.AppInstance)
	for instanceID, instance := range matchInstancesmap {
		_, ok := effectedInstancesMap[instanceID]
		if ok {
			effectedInstance := effectedInstancesMap[instanceID]
			// 将 match 查到的实例状态 status 传给 effected 的预留字段 status，用于之后展示的实例状态判断
			effectedInstance.State = instance.State
			onlineEffectedInstances[instanceID] = effectedInstance
			delete(matchInstancesmap, instanceID)
			delete(effectedInstancesMap, instanceID)
		}
	}

	// onlineEffectedInstances 分为 onlineEffectReloadInstance 和 onlineEffectInstance 2种
	count := 0
	for _, instance := range onlineEffectedInstances {
		if instance.ReloadCode != 0 {
			count++
			result = append(result, instance)
			delete(onlineEffectedInstances, instance.Instanceid)
		}
	}
	categoryTotal[OnlineEffectReloadInstance] = count

	count = 0
	for _, instance := range onlineEffectedInstances {
		count++
		result = append(result, instance)
	}
	categoryTotal[OnlineEffectInstance] = count

	// onlineUneffectInstance
	count = 0
	for _, instance := range matchInstancesmap {
		count++
		result = append(result, instance)
	}
	categoryTotal[OnlineUnEffectInstance] = count

	// effectedInstancesMap 分为 offlineEffectReloadInstance and offlineEffectInstance
	count = 0
	for _, instance := range effectedInstancesMap {
		if instance.ReloadCode != 0 {
			count++
			result = append(result, instance)
			delete(effectedInstancesMap, instance.Instanceid)
		}
	}
	categoryTotal[OfflineEffectReloadInstance] = count

	count = 0
	for _, instance := range effectedInstancesMap {
		count++
		result = append(result, instance)
	}
	categoryTotal[OfflineEffectInstance] = count

	return result, categoryTotal
}

// handleMatchAppInstByStrategy query match instances by strategy
func handleMatchAppInstByStrategy(cmd *cobra.Command, operator *service.AccessOperator, strategyID string) error {
	business, err := operator.GetBusiness(context.TODO(), operator.Business)
	if err != nil {
		return err
	}
	if business == nil {
		return fmt.Errorf("No relative Business %s Resource", operator.Business)
	}
	strategy, err := operator.GetStrategyById(context.TODO(), strategyID)
	if err != nil {
		return err
	}
	if strategy == nil {
		return fmt.Errorf("No strategy %s resource!", strategyID)
	}
	instances, err := operator.ListMatchedInstanceByStrategyId(context.TODO(), strategyID)
	if err != nil {
		return err
	}
	if len(instances) == 0 {
		cmd.Printf("No instance resource found for the specified strategyId\n")
		return nil
	}
	clusterMap, zoneMap := handleInstanceClusterAndZone(operator, instances)
	app, err := operator.GetAppByAppID(context.TODO(), business.Bid, strategy.Appid)
	if err != nil {
		return err
	}
	//format output
	utils.PrintAppInstances(instances, operator.Business, app.Name, clusterMap, zoneMap)
	return nil
}

// simpleListAppInstance query instance filer by (business app cluster zone status)
func simpleListAppInstance(cmd *cobra.Command, operator *service.AccessOperator) error {
	err := option.SetGlobalVarByName(cmd, "app")
	if err != nil {
		return err
	}
	appName, _ := cmd.Flags().GetString("app")
	clusterName, _ := cmd.Flags().GetString("cluster")
	zoneName, _ := cmd.Flags().GetString("zone")
	queryType, _ := cmd.Flags().GetInt32("status")
	request := &service.InstanceOption{
		AppName:     appName,
		ClusterName: clusterName,
		ZoneName:    zoneName,
		QueryType:   queryType,
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

func sliceToMap(instances []*common.AppInstance) map[uint64]*common.AppInstance {
	mapData := make(map[uint64]*common.AppInstance)
	for _, instance := range instances {
		mapData[instance.Instanceid] = instance
	}
	return mapData
}
