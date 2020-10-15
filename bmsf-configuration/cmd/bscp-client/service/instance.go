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
	"context"
	"fmt"

	"google.golang.org/grpc"

	"bk-bscp/internal/protocol/accessserver"
	"bk-bscp/internal/protocol/common"
	pkgcommon "bk-bscp/pkg/common"
	"bk-bscp/pkg/logger"
)

//InstanceOption for instance list
type InstanceOption struct {
	AppName     string
	ClusterName string
	ZoneName    string
	QueryType   int32
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
	request := &accessserver.QueryHistoryAppInstancesReq{
		Seq:       pkgcommon.Sequence(),
		Bid:       business.Bid,
		QueryType: option.QueryType,
		Index:     operator.index,
		Limit:     operator.limit,
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
		zone, zoneErr := operator.innerGetZone(cxt, business.Bid, request.Appid, "", option.ZoneName)
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
	response, err := operator.Client.QueryHistoryAppInstances(cxt, request, grpcOptions...)
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

//ListEffectedInstance get effected Instance List which release published to
//Args:
//	release
//return:
//  appInstance list
//	error: error info if that happened
func (operator *AccessOperator) ListEffectedInstance(cxt context.Context, bID, cfgSetID, releaseID string) ([]*common.AppInstance, error) {
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
		logger.V(3).Infof("ListEffectedInstance %s details failed, %s", releaseID, err.Error())
		return nil, err
	}
	if response.ErrCode == common.ErrCode_E_DM_NOT_FOUND {
		logger.V(3).Infof("ListEffectedInstance %s: resource Not Found.", releaseID)
		return nil, nil
	}
	if response.ErrCode != common.ErrCode_E_OK {
		logger.V(3).Infof("ListEffectedInstance %s successfully, but response Err, %s", releaseID, response.ErrMsg)
		return nil, fmt.Errorf("%s", response.ErrMsg)
	}
	return response.Instances, nil
}

//GetAppinstanceReleasees get Instance release by cfgsetId and instance info
//Args:
//	cfgsetId,instance
//return:
//  releaseId
func (operator *AccessOperator) GetAppinstanceReleasees(cxt context.Context, cfgSetId string, instance *common.AppInstance) (string, error) {
	if instance == nil {
		return "", fmt.Errorf("instance is nil")
	}

	// query appinstance by cfgsetid and instance info
	request := &accessserver.QueryAppInstanceReleaseReq{
		Seq:       pkgcommon.Sequence(),
		Bid:       instance.Bid,
		Appid:     instance.Appid,
		Clusterid: instance.Clusterid,
		Zoneid:    instance.Zoneid,
		Dc:        instance.Dc,
		IP:        instance.IP,
		Cfgsetid:  cfgSetId,
	}
	grpcOptions := []grpc.CallOption{
		grpc.WaitForReady(true),
	}
	response, err := operator.Client.QueryAppInstanceRelease(cxt, request, grpcOptions...)

	// check result
	if err != nil {
		logger.V(3).Infof("GetAppinstanceReleasees %s details failed, %s", instance, err.Error())
		return "", err
	}
	if response.ErrCode == common.ErrCode_E_DM_NOT_FOUND {
		logger.V(3).Infof("GetAppinstanceReleasees %s: resource Not Found.", instance)
		return "", nil
	}
	if response.ErrCode != common.ErrCode_E_OK {
		logger.V(3).Infof("GetAppinstanceReleasees %s successfully, but response Err, %s", instance, response.ErrMsg)
		return "", fmt.Errorf("%s", response.ErrMsg)
	}

	return response.Releaseid, nil
}

func (operator *AccessOperator) innerGetMatchedInstance(cxt context.Context, bID, strategyId, releasesId string) ([]*common.AppInstance, error) {
	request := &accessserver.QueryMatchedAppInstancesReq{
		Seq:        pkgcommon.Sequence(),
		Bid:        bID,
		Strategyid: strategyId,
		Releaseid:  releasesId,
		Index:      operator.index,
		Limit:      operator.limit,
	}
	grpcOptions := []grpc.CallOption{
		grpc.WaitForReady(true),
	}
	response, err := operator.Client.QueryMatchedAppInstances(cxt, request, grpcOptions...)
	if err != nil {
		logger.V(3).Infof("innerGetMatchedInstance strategyId[%s] releasesId[%s] details failed, %s", strategyId, releasesId, err.Error())
		return nil, err
	}
	if response.ErrCode == common.ErrCode_E_DM_NOT_FOUND {
		logger.V(3).Infof("innerGetMatchedInstance strategyId[%s] releasesId[%s] : resource Not Found.", strategyId, releasesId)
		return nil, nil
	}
	if response.ErrCode != common.ErrCode_E_OK {
		logger.V(3).Infof("innerGetMatchedInstance strategyId[%s] releasesId[%s] or  successfully, but response Err, %s", strategyId, releasesId, response.ErrMsg)
		return nil, fmt.Errorf("%s", response.ErrMsg)
	}
	return response.Instances, nil
}

func (operator *AccessOperator) ListMatchedInstanceByReleaseId(cxt context.Context, bID, releaseId string) ([]*common.AppInstance, error) {
	return operator.innerGetMatchedInstance(cxt, bID, "", releaseId)
}

func (operator *AccessOperator) ListMatchedInstanceByStrategyId(cxt context.Context, strategyId string) ([]*common.AppInstance, error) {
	business, err := operator.GetBusiness(cxt, operator.Business)
	if err != nil {
		logger.V(3).Infof("Query relative business %s failed, %s", operator.Business, err.Error())
		return nil, err
	}
	if business == nil {
		return nil, fmt.Errorf("No relative Business %s Resource", operator.Business)
	}
	return operator.innerGetMatchedInstance(cxt, business.Bid, strategyId, "")
}
