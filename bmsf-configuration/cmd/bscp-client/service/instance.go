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

package service

import (
	"bk-bscp/internal/protocol/accessserver"
	"bk-bscp/internal/protocol/common"
	pkgcommon "bk-bscp/pkg/common"
	"bk-bscp/pkg/logger"
	"context"
	"fmt"

	"google.golang.org/grpc"
)

//InstanceOption for instance list
type InstanceOption struct {
	AppName     string
	ClusterName string
	ZoneName    string
}

//ListAppInstance list all App instance information
//return:
//	strategies: all strategies, nil if not exist
//	error: error info if that happened
func (operator *AccessOperator) ListAppInstance(cxt context.Context, option *InstanceOption) ([]*common.AppInstance, error) {
	//query business and app first
	if option == nil {
		return nil, fmt.Errorf("Lost option information")
	}
	business, err := operator.GetBusiness(cxt, operator.Business)
	if err != nil {
		logger.V(3).Infof("Query relative business %s failed, %s", operator.Business, err.Error())
		return nil, err
	}
	if business == nil {
		return nil, fmt.Errorf("No relative Business %s Resource", operator.Business)
	}
	request := &accessserver.QueryReachableAppInstancesReq{
		Seq:   pkgcommon.Sequence(),
		Bid:   business.Bid,
		Index: operator.index,
		Limit: operator.limit,
	}
	if len(option.AppName) != 0 {
		app, appErr := operator.GetApp(cxt, option.AppName)
		if appErr != nil {
			logger.V(3).Infof("Query relative App %s failed, %s", option.AppName, appErr.Error())
			return nil, appErr
		}
		if app == nil {
			return nil, fmt.Errorf("No relative Application %s Resource", option.AppName)
		}
		request.Appid = app.Appid
	}
	if len(option.ClusterName) != 0 {
		cluster, clusterErr := operator.GetLogicCluster(cxt, option.AppName, option.ClusterName)
		if clusterErr != nil {
			logger.V(3).Infof("Query relative Cluster %s failed, %s", option.ClusterName, clusterErr.Error())
			return nil, clusterErr
		}
		if cluster == nil {
			return nil, fmt.Errorf("No relative Cluster %s Resource", option.ClusterName)
		}
		request.Clusterid = cluster.Clusterid
	}
	if len(option.ZoneName) != 0 {
		zone, zoneErr := operator.GetZone(cxt, option.AppName, option.ClusterName, option.ZoneName)
		if err != nil {
			logger.V(3).Infof("Query relative Zone %s failed, %s", option.ZoneName, zoneErr.Error())
			return nil, zoneErr
		}
		if zone == nil {
			return nil, fmt.Errorf("No relative Zone %s Resource", option.ZoneName)
		}
		request.Zoneid = zone.Zoneid
	}
	grpcOptions := []grpc.CallOption{
		grpc.WaitForReady(true),
	}
	response, err := operator.Client.QueryReachableAppInstances(cxt, request, grpcOptions...)
	if err != nil {
		logger.V(3).Infof("ListAppInstance failed, %s", err.Error())
		return nil, err
	}
	if response.ErrCode != common.ErrCode_E_OK {
		logger.V(3).Infof("ListAppInstance all successfully, but response Err, %s", response.ErrMsg)
		return nil, fmt.Errorf("%s", response.ErrMsg)
	}
	return response.Instances, nil
}

//ListInstanceByRelease get Instance List which release published to
//Args:
//	name: strategy name
//	ID: strategy ID
//return:
//	strategy: specified configset, nil if not exist
//	error: error info if that happened
func (operator *AccessOperator) ListInstanceByRelease(cxt context.Context, appName, cfgsetName, releaseID string) ([]*common.AppInstance, error) {
	//query business and app first
	business, app, err := getBusinessAndApp(operator, operator.Business, appName)
	if err != nil {
		return nil, err
	}
	configSet, err := operator.innergetConfigSet(cxt, business.Bid, app.Appid, cfgsetName)
	if err != nil {
		return nil, err
	}
	if configSet == nil {
		return nil, nil
	}
	return operator.innerGetInstanceByRelease(cxt, business.Bid, configSet.Cfgsetid, releaseID)
}

func (operator *AccessOperator) innerGetInstanceByRelease(cxt context.Context, bID, cfgSetID, releaseID string) ([]*common.AppInstance, error) {
	request := &accessserver.QueryEffectedAppInstancesReq{
		Seq:       pkgcommon.Sequence(),
		Bid:       bID,
		Cfgsetid:  cfgSetID,
		Releaseid: releaseID,
		Index:     operator.index,
		Limit:     operator.limit,
	}
	grpcOptions := []grpc.CallOption{
		grpc.WaitForReady(true),
	}
	response, err := operator.Client.QueryEffectedAppInstances(cxt, request, grpcOptions...)
	if err != nil {
		logger.V(3).Infof("ListInstanceByRelease %s details failed, %s", releaseID, err.Error())
		return nil, err
	}
	if response.ErrCode == common.ErrCode_E_DM_NOT_FOUND {
		logger.V(3).Infof("ListInstanceByRelease %s: resource Not Found.", releaseID)
		return nil, nil
	}
	if response.ErrCode != common.ErrCode_E_OK {
		logger.V(3).Infof("ListInstanceByRelease %s successfully, but response Err, %s", releaseID, response.ErrMsg)
		return nil, fmt.Errorf("%s", response.ErrMsg)
	}
	return response.Instances, nil
}
