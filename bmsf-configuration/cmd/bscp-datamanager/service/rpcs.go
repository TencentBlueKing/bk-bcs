/*
Tencent is pleased to support the open source community by making Blueking Container Service available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package service

import (
	"context"
	"time"

	appinstanceaction "bk-bscp/cmd/bscp-datamanager/actions/appinstance"
	appaction "bk-bscp/cmd/bscp-datamanager/actions/application"
	auditaction "bk-bscp/cmd/bscp-datamanager/actions/audit"
	businessaction "bk-bscp/cmd/bscp-datamanager/actions/business"
	clusteraction "bk-bscp/cmd/bscp-datamanager/actions/cluster"
	commitaction "bk-bscp/cmd/bscp-datamanager/actions/commit"
	configsaction "bk-bscp/cmd/bscp-datamanager/actions/configs"
	configsetaction "bk-bscp/cmd/bscp-datamanager/actions/configset"
	metadataaction "bk-bscp/cmd/bscp-datamanager/actions/metadata"
	multicommitaction "bk-bscp/cmd/bscp-datamanager/actions/multi-commit"
	multireleaseaction "bk-bscp/cmd/bscp-datamanager/actions/multi-release"
	releaseaction "bk-bscp/cmd/bscp-datamanager/actions/release"
	shardingaction "bk-bscp/cmd/bscp-datamanager/actions/sharding"
	shardingdbaction "bk-bscp/cmd/bscp-datamanager/actions/shardingdb"
	strategyaction "bk-bscp/cmd/bscp-datamanager/actions/strategy"
	zoneaction "bk-bscp/cmd/bscp-datamanager/actions/zone"
	pbcommon "bk-bscp/internal/protocol/common"
	pb "bk-bscp/internal/protocol/datamanager"
	"bk-bscp/pkg/logger"
)

// QueryAuthInfo returns auth information of target business.
func (dm *DataManager) QueryAuthInfo(ctx context.Context, req *pb.QueryAuthInfoReq) (*pb.QueryAuthInfoResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryAuthInfo[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryAuthInfoResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := dm.collector.StatRequest("QueryAuthInfo", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryAuthInfo[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := businessaction.NewAuthInfoAction(dm.viper, dm.smgr, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// QueryAppMetadata returns metadata informations of target app.
func (dm *DataManager) QueryAppMetadata(ctx context.Context, req *pb.QueryAppMetadataReq) (*pb.QueryAppMetadataResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryAppMetadata[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryAppMetadataResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := dm.collector.StatRequest("QueryAppMetadata", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryAppMetadata[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := metadataaction.NewQueryAction(dm.viper, dm.smgr, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// CreateBusiness creates new business.
func (dm *DataManager) CreateBusiness(ctx context.Context, req *pb.CreateBusinessReq) (*pb.CreateBusinessResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("CreateBusiness[%d]| input[%+v]", req.Seq, req)
	response := &pb.CreateBusinessResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := dm.collector.StatRequest("CreateBusiness", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("CreateBusiness[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := businessaction.NewCreateAction(dm.viper, dm.smgr, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// QueryBusiness returns target business.
func (dm *DataManager) QueryBusiness(ctx context.Context, req *pb.QueryBusinessReq) (*pb.QueryBusinessResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryBusiness[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryBusinessResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := dm.collector.StatRequest("QueryBusiness", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryBusiness[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := businessaction.NewQueryAction(dm.viper, dm.smgr, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// QueryBusinessList returns all business.
func (dm *DataManager) QueryBusinessList(ctx context.Context, req *pb.QueryBusinessListReq) (*pb.QueryBusinessListResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryBusinessList[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryBusinessListResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := dm.collector.StatRequest("QueryBusinessList", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryBusinessList[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := businessaction.NewListAction(dm.viper, dm.smgr, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// UpdateBusiness updates target business.
func (dm *DataManager) UpdateBusiness(ctx context.Context, req *pb.UpdateBusinessReq) (*pb.UpdateBusinessResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("UpdateBusiness[%d]| input[%+v]", req.Seq, req)
	response := &pb.UpdateBusinessResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := dm.collector.StatRequest("UpdateBusiness", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("UpdateBusiness[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := businessaction.NewUpdateAction(dm.viper, dm.smgr, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// CreateApp creates new app.
func (dm *DataManager) CreateApp(ctx context.Context, req *pb.CreateAppReq) (*pb.CreateAppResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("CreateApp[%d]| input[%+v]", req.Seq, req)
	response := &pb.CreateAppResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := dm.collector.StatRequest("CreateApp", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("CreateApp[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := appaction.NewCreateAction(dm.viper, dm.smgr, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// QueryApp returns target app.
func (dm *DataManager) QueryApp(ctx context.Context, req *pb.QueryAppReq) (*pb.QueryAppResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryApp[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryAppResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := dm.collector.StatRequest("QueryApp", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryApp[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := appaction.NewQueryAction(dm.viper, dm.smgr, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// QueryAppList returns all apps.
func (dm *DataManager) QueryAppList(ctx context.Context, req *pb.QueryAppListReq) (*pb.QueryAppListResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryAppList[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryAppListResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := dm.collector.StatRequest("QueryAppList", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryAppList[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := appaction.NewListAction(dm.viper, dm.smgr, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// UpdateApp updates target app.
func (dm *DataManager) UpdateApp(ctx context.Context, req *pb.UpdateAppReq) (*pb.UpdateAppResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("UpdateApp[%d]| input[%+v]", req.Seq, req)
	response := &pb.UpdateAppResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := dm.collector.StatRequest("UpdateApp", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("UpdateApp[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := appaction.NewUpdateAction(dm.viper, dm.smgr, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// DeleteApp deletes target app.
func (dm *DataManager) DeleteApp(ctx context.Context, req *pb.DeleteAppReq) (*pb.DeleteAppResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("DeleteApp[%d]| input[%+v]", req.Seq, req)
	response := &pb.DeleteAppResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := dm.collector.StatRequest("DeleteApp", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("DeleteApp[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := appaction.NewDeleteAction(dm.viper, dm.smgr, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// CreateCluster creates new cluster.
func (dm *DataManager) CreateCluster(ctx context.Context, req *pb.CreateClusterReq) (*pb.CreateClusterResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("CreateCluster[%d]| input[%+v]", req.Seq, req)
	response := &pb.CreateClusterResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := dm.collector.StatRequest("CreateCluster", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("CreateCluster[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := clusteraction.NewCreateAction(dm.viper, dm.smgr, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// QueryCluster returns target cluster.
func (dm *DataManager) QueryCluster(ctx context.Context, req *pb.QueryClusterReq) (*pb.QueryClusterResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryCluster[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryClusterResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := dm.collector.StatRequest("QueryCluster", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryCluster[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := clusteraction.NewQueryAction(dm.viper, dm.smgr, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// QueryClusterList returns all clusters.
func (dm *DataManager) QueryClusterList(ctx context.Context, req *pb.QueryClusterListReq) (*pb.QueryClusterListResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryClusterList[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryClusterListResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := dm.collector.StatRequest("QueryClusterList", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryClusterList[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := clusteraction.NewListAction(dm.viper, dm.smgr, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// UpdateCluster updates target cluster.
func (dm *DataManager) UpdateCluster(ctx context.Context, req *pb.UpdateClusterReq) (*pb.UpdateClusterResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("UpdateCluster[%d]| input[%+v]", req.Seq, req)
	response := &pb.UpdateClusterResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := dm.collector.StatRequest("UpdateCluster", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("UpdateCluster[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := clusteraction.NewUpdateAction(dm.viper, dm.smgr, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// DeleteCluster deletes target cluster.
func (dm *DataManager) DeleteCluster(ctx context.Context, req *pb.DeleteClusterReq) (*pb.DeleteClusterResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("DeleteCluster[%d]| input[%+v]", req.Seq, req)
	response := &pb.DeleteClusterResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := dm.collector.StatRequest("DeleteCluster", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("DeleteCluster[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := clusteraction.NewDeleteAction(dm.viper, dm.smgr, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// CreateZone creates new zone.
func (dm *DataManager) CreateZone(ctx context.Context, req *pb.CreateZoneReq) (*pb.CreateZoneResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("CreateZone[%d]| input[%+v]", req.Seq, req)
	response := &pb.CreateZoneResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := dm.collector.StatRequest("CreateZone", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("CreateZone[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := zoneaction.NewCreateAction(dm.viper, dm.smgr, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// QueryZone returns target zone.
func (dm *DataManager) QueryZone(ctx context.Context, req *pb.QueryZoneReq) (*pb.QueryZoneResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryZone[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryZoneResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := dm.collector.StatRequest("QueryZone", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryZone[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := zoneaction.NewQueryAction(dm.viper, dm.smgr, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// QueryZoneList returns all zones.
func (dm *DataManager) QueryZoneList(ctx context.Context, req *pb.QueryZoneListReq) (*pb.QueryZoneListResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryZoneList[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryZoneListResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := dm.collector.StatRequest("QueryZoneList", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryZoneList[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := zoneaction.NewListAction(dm.viper, dm.smgr, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// UpdateZone updates target zone.
func (dm *DataManager) UpdateZone(ctx context.Context, req *pb.UpdateZoneReq) (*pb.UpdateZoneResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("UpdateZone[%d]| input[%+v]", req.Seq, req)
	response := &pb.UpdateZoneResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := dm.collector.StatRequest("UpdateZone", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("UpdateZone[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := zoneaction.NewUpdateAction(dm.viper, dm.smgr, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// DeleteZone deletes target zone.
func (dm *DataManager) DeleteZone(ctx context.Context, req *pb.DeleteZoneReq) (*pb.DeleteZoneResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("DeleteZone[%d]| input[%+v]", req.Seq, req)
	response := &pb.DeleteZoneResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := dm.collector.StatRequest("DeleteZone", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("DeleteZone[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := zoneaction.NewDeleteAction(dm.viper, dm.smgr, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// CreateConfigSet creates new configset.
func (dm *DataManager) CreateConfigSet(ctx context.Context, req *pb.CreateConfigSetReq) (*pb.CreateConfigSetResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("CreateConfigSet[%d]| input[%+v]", req.Seq, req)
	response := &pb.CreateConfigSetResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := dm.collector.StatRequest("CreateConfigSet", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("CreateConfigSet[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := configsetaction.NewCreateAction(dm.viper, dm.smgr, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// QueryConfigSet returns target configset.
func (dm *DataManager) QueryConfigSet(ctx context.Context, req *pb.QueryConfigSetReq) (*pb.QueryConfigSetResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryConfigSet[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryConfigSetResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := dm.collector.StatRequest("QueryConfigSet", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryConfigSet[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := configsetaction.NewQueryAction(dm.viper, dm.smgr, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// QueryConfigSetList returns all configsets.
func (dm *DataManager) QueryConfigSetList(ctx context.Context, req *pb.QueryConfigSetListReq) (*pb.QueryConfigSetListResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryConfigSetList[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryConfigSetListResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := dm.collector.StatRequest("QueryConfigSetList", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryConfigSetList[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := configsetaction.NewListAction(dm.viper, dm.smgr, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// UpdateConfigSet updates target configset.
func (dm *DataManager) UpdateConfigSet(ctx context.Context, req *pb.UpdateConfigSetReq) (*pb.UpdateConfigSetResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("UpdateConfigSet[%d]| input[%+v]", req.Seq, req)
	response := &pb.UpdateConfigSetResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := dm.collector.StatRequest("UpdateConfigSet", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("UpdateConfigSet[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := configsetaction.NewUpdateAction(dm.viper, dm.smgr, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// DeleteConfigSet deletes target configset.
func (dm *DataManager) DeleteConfigSet(ctx context.Context, req *pb.DeleteConfigSetReq) (*pb.DeleteConfigSetResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("DeleteConfigSet[%d]| input[%+v]", req.Seq, req)
	response := &pb.DeleteConfigSetResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := dm.collector.StatRequest("DeleteConfigSet", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("DeleteConfigSet[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := configsetaction.NewDeleteAction(dm.viper, dm.smgr, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// LockConfigSet locks target configset.
func (dm *DataManager) LockConfigSet(ctx context.Context, req *pb.LockConfigSetReq) (*pb.LockConfigSetResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("LockConfigSet[%d]| input[%+v]", req.Seq, req)
	response := &pb.LockConfigSetResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := dm.collector.StatRequest("LockConfigSet", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("LockConfigSet[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := configsetaction.NewLockAction(dm.viper, dm.smgr, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// UnlockConfigSet unlocks target configset.
func (dm *DataManager) UnlockConfigSet(ctx context.Context, req *pb.UnlockConfigSetReq) (*pb.UnlockConfigSetResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("UnlockConfigSet[%d]| input[%+v]", req.Seq, req)
	response := &pb.UnlockConfigSetResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := dm.collector.StatRequest("UnlockConfigSet", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("UnlockConfigSet[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := configsetaction.NewUnlockAction(dm.viper, dm.smgr, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// CreateConfigs creates new configs.
func (dm *DataManager) CreateConfigs(ctx context.Context, req *pb.CreateConfigsReq) (*pb.CreateConfigsResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("CreateConfigs[%d]| input[%+v]", req.Seq, req)
	response := &pb.CreateConfigsResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := dm.collector.StatRequest("CreateConfigs", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("CreateConfigs[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := configsaction.NewCreateAction(dm.viper, dm.smgr, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// QueryConfigs returns target configs.
func (dm *DataManager) QueryConfigs(ctx context.Context, req *pb.QueryConfigsReq) (*pb.QueryConfigsResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryConfigs[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryConfigsResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := dm.collector.StatRequest("QueryConfigs", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryConfigs[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := configsaction.NewQueryAction(dm.viper, dm.smgr, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// QueryConfigsList returns configs list.
func (dm *DataManager) QueryConfigsList(ctx context.Context, req *pb.QueryConfigsListReq) (*pb.QueryConfigsListResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryConfigsList[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryConfigsListResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := dm.collector.StatRequest("QueryConfigsList", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryConfigsList[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := configsaction.NewListAction(dm.viper, dm.smgr, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// QueryReleaseConfigs returns configs of target relase.
func (dm *DataManager) QueryReleaseConfigs(ctx context.Context, req *pb.QueryReleaseConfigsReq) (*pb.QueryReleaseConfigsResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryReleaseConfigs[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryReleaseConfigsResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := dm.collector.StatRequest("QueryReleaseConfigs", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryReleaseConfigs[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := configsaction.NewReleaseConfigsAction(dm.viper, dm.smgr, dm.collector, dm.commitCache, dm.configsCache, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// CreateCommit creates new commit.
func (dm *DataManager) CreateCommit(ctx context.Context, req *pb.CreateCommitReq) (*pb.CreateCommitResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("CreateCommit[%d]| input[%+v]", req.Seq, req)
	response := &pb.CreateCommitResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := dm.collector.StatRequest("CreateCommit", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("CreateCommit[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := commitaction.NewCreateAction(dm.viper, dm.smgr, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// QueryCommit returns target commit.
func (dm *DataManager) QueryCommit(ctx context.Context, req *pb.QueryCommitReq) (*pb.QueryCommitResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryCommit[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryCommitResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := dm.collector.StatRequest("QueryCommit", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryCommit[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := commitaction.NewQueryAction(dm.viper, dm.smgr, dm.collector, dm.commitCache, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// QueryHistoryCommits returns history commits of target configset.
func (dm *DataManager) QueryHistoryCommits(ctx context.Context, req *pb.QueryHistoryCommitsReq) (*pb.QueryHistoryCommitsResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryHistoryCommits[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryHistoryCommitsResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := dm.collector.StatRequest("QueryHistoryCommits", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryHistoryCommits[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := commitaction.NewListAction(dm.viper, dm.smgr, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// UpdateCommit updates target commit.
func (dm *DataManager) UpdateCommit(ctx context.Context, req *pb.UpdateCommitReq) (*pb.UpdateCommitResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("UpdateCommit[%d]| input[%+v]", req.Seq, req)
	response := &pb.UpdateCommitResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := dm.collector.StatRequest("UpdateCommit", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("UpdateCommit[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := commitaction.NewUpdateAction(dm.viper, dm.smgr, dm.commitCache, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// CancelCommit cancels target commit.
func (dm *DataManager) CancelCommit(ctx context.Context, req *pb.CancelCommitReq) (*pb.CancelCommitResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("CancelCommit[%d]| input[%+v]", req.Seq, req)
	response := &pb.CancelCommitResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := dm.collector.StatRequest("CancelCommit", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("CancelCommit[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := commitaction.NewCancelAction(dm.viper, dm.smgr, dm.commitCache, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// ConfirmCommit confirms target commit.
func (dm *DataManager) ConfirmCommit(ctx context.Context, req *pb.ConfirmCommitReq) (*pb.ConfirmCommitResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("ConfirmCommit[%d]| input[%+v]", req.Seq, req)
	response := &pb.ConfirmCommitResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := dm.collector.StatRequest("ConfirmCommit", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("ConfirmCommit[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := commitaction.NewConfirmAction(dm.viper, dm.smgr, dm.commitCache, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// CreateMultiCommit creates new multi commit.
func (dm *DataManager) CreateMultiCommit(ctx context.Context, req *pb.CreateMultiCommitReq) (*pb.CreateMultiCommitResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("CreateMultiCommit[%d]| input[%+v]", req.Seq, req)
	response := &pb.CreateMultiCommitResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := dm.collector.StatRequest("CreateMultiCommit", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("CreateMultiCommit[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := multicommitaction.NewCreateAction(dm.viper, dm.smgr, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// QueryMultiCommit returns target multi commit.
func (dm *DataManager) QueryMultiCommit(ctx context.Context, req *pb.QueryMultiCommitReq) (*pb.QueryMultiCommitResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryMultiCommit[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryMultiCommitResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := dm.collector.StatRequest("QueryMultiCommit", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryMultiCommit[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := multicommitaction.NewQueryAction(dm.viper, dm.smgr, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// QueryHistoryMultiCommits returns history multi commits of target configset.
func (dm *DataManager) QueryHistoryMultiCommits(ctx context.Context, req *pb.QueryHistoryMultiCommitsReq) (*pb.QueryHistoryMultiCommitsResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryHistoryMultiCommits[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryHistoryMultiCommitsResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := dm.collector.StatRequest("QueryHistoryMultiCommits", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryHistoryMultiCommits[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := multicommitaction.NewListAction(dm.viper, dm.smgr, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// QueryMultiCommitSubList returns multi commit sub list.
func (dm *DataManager) QueryMultiCommitSubList(ctx context.Context, req *pb.QueryMultiCommitSubListReq) (*pb.QueryMultiCommitSubListResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryMultiCommitSubList[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryMultiCommitSubListResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := dm.collector.StatRequest("QueryMultiCommitSubList", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryMultiCommitSubList[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := multicommitaction.NewSubListAction(dm.viper, dm.smgr, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// UpdateMultiCommit updates target multi commit.
func (dm *DataManager) UpdateMultiCommit(ctx context.Context, req *pb.UpdateMultiCommitReq) (*pb.UpdateMultiCommitResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("UpdateMultiCommit[%d]| input[%+v]", req.Seq, req)
	response := &pb.UpdateMultiCommitResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := dm.collector.StatRequest("UpdateMultiCommit", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("UpdateMultiCommit[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := multicommitaction.NewUpdateAction(dm.viper, dm.smgr, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// CancelMultiCommit cancels target multi commit.
func (dm *DataManager) CancelMultiCommit(ctx context.Context, req *pb.CancelMultiCommitReq) (*pb.CancelMultiCommitResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("CancelMultiCommit[%d]| input[%+v]", req.Seq, req)
	response := &pb.CancelMultiCommitResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := dm.collector.StatRequest("CancelMultiCommit", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("CancelMultiCommit[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := multicommitaction.NewCancelAction(dm.viper, dm.smgr, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// ConfirmMultiCommit confirms target multi commit.
func (dm *DataManager) ConfirmMultiCommit(ctx context.Context, req *pb.ConfirmMultiCommitReq) (*pb.ConfirmMultiCommitResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("ConfirmMultiCommit[%d]| input[%+v]", req.Seq, req)
	response := &pb.ConfirmMultiCommitResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := dm.collector.StatRequest("ConfirmMultiCommit", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("ConfirmMultiCommit[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := multicommitaction.NewConfirmAction(dm.viper, dm.smgr, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// CreateRelease creates new release.
func (dm *DataManager) CreateRelease(ctx context.Context, req *pb.CreateReleaseReq) (*pb.CreateReleaseResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("CreateRelease[%d]| input[%+v]", req.Seq, req)
	response := &pb.CreateReleaseResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := dm.collector.StatRequest("CreateRelease", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("CreateRelease[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := releaseaction.NewCreateAction(dm.viper, dm.smgr, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// QueryRelease returns target release.
func (dm *DataManager) QueryRelease(ctx context.Context, req *pb.QueryReleaseReq) (*pb.QueryReleaseResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryRelease[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryReleaseResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := dm.collector.StatRequest("QueryRelease", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryRelease[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := releaseaction.NewQueryAction(dm.viper, dm.smgr, dm.collector, dm.releaseCache, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// QueryNewestReleases returns newest releases.
func (dm *DataManager) QueryNewestReleases(ctx context.Context, req *pb.QueryNewestReleasesReq) (*pb.QueryNewestReleasesResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryNewestReleases[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryNewestReleasesResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := dm.collector.StatRequest("QueryNewestReleases", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryNewestReleases[%d] output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := releaseaction.NewNewestAction(dm.viper, dm.smgr, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// QueryHistoryReleases returns history releases.
func (dm *DataManager) QueryHistoryReleases(ctx context.Context, req *pb.QueryHistoryReleasesReq) (*pb.QueryHistoryReleasesResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryHistoryReleases[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryHistoryReleasesResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := dm.collector.StatRequest("QueryHistoryReleases", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryHistoryReleases[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := releaseaction.NewListAction(dm.viper, dm.smgr, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// UpdateRelease updates target release.
func (dm *DataManager) UpdateRelease(ctx context.Context, req *pb.UpdateReleaseReq) (*pb.UpdateReleaseResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("UpdateRelease[%d]| input[%+v]", req.Seq, req)
	response := &pb.UpdateReleaseResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := dm.collector.StatRequest("UpdateRelease", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("UpdateRelease[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := releaseaction.NewUpdateAction(dm.viper, dm.smgr, dm.releaseCache, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// CancelRelease cancels target release.
func (dm *DataManager) CancelRelease(ctx context.Context, req *pb.CancelReleaseReq) (*pb.CancelReleaseResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("CancelRelease[%d]| input[%+v]", req.Seq, req)
	response := &pb.CancelReleaseResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := dm.collector.StatRequest("CancelRelease", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("CancelRelease[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := releaseaction.NewCancelAction(dm.viper, dm.smgr, dm.releaseCache, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// RollbackRelease rollbacks target release.
func (dm *DataManager) RollbackRelease(ctx context.Context, req *pb.RollbackReleaseReq) (*pb.RollbackReleaseResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("RollbackRelease[%d]| input[%+v]", req.Seq, req)
	response := &pb.RollbackReleaseResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := dm.collector.StatRequest("RollbackRelease", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("RollbackRelease[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := releaseaction.NewRollbackAction(dm.viper, dm.smgr, dm.releaseCache, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// PublishRelease publishs target release.
func (dm *DataManager) PublishRelease(ctx context.Context, req *pb.PublishReleaseReq) (*pb.PublishReleaseResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("PublishRelease[%d]| input[%+v]", req.Seq, req)
	response := &pb.PublishReleaseResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := dm.collector.StatRequest("PublishRelease", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("PublishRelease[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := releaseaction.NewPublishAction(dm.viper, dm.smgr, dm.releaseCache, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// CreateMultiRelease creates new multi release.
func (dm *DataManager) CreateMultiRelease(ctx context.Context, req *pb.CreateMultiReleaseReq) (*pb.CreateMultiReleaseResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("CreateMultiRelease[%d]| input[%+v]", req.Seq, req)
	response := &pb.CreateMultiReleaseResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := dm.collector.StatRequest("CreateMultiRelease", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("CreateMultiRelease[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := multireleaseaction.NewCreateAction(dm.viper, dm.smgr, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// QueryMultiRelease returns target multi release.
func (dm *DataManager) QueryMultiRelease(ctx context.Context, req *pb.QueryMultiReleaseReq) (*pb.QueryMultiReleaseResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryMultiRelease[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryMultiReleaseResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := dm.collector.StatRequest("QueryMultiRelease", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryMultiRelease[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := multireleaseaction.NewQueryAction(dm.viper, dm.smgr, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// QueryMultiReleaseSubList returns target multi release sub list.
func (dm *DataManager) QueryMultiReleaseSubList(ctx context.Context, req *pb.QueryMultiReleaseSubListReq) (*pb.QueryMultiReleaseSubListResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryMultiReleaseSubList[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryMultiReleaseSubListResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := dm.collector.StatRequest("QueryMultiReleaseSubList", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryMultiReleaseSubList[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := multireleaseaction.NewSubListAction(dm.viper, dm.smgr, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// QueryHistoryMultiReleases returns history multi releases.
func (dm *DataManager) QueryHistoryMultiReleases(ctx context.Context, req *pb.QueryHistoryMultiReleasesReq) (*pb.QueryHistoryMultiReleasesResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryHistoryMultiReleases[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryHistoryMultiReleasesResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := dm.collector.StatRequest("QueryHistoryMultiReleases", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryHistoryMultiReleases[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := multireleaseaction.NewListAction(dm.viper, dm.smgr, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// UpdateMultiRelease updates target multi release.
func (dm *DataManager) UpdateMultiRelease(ctx context.Context, req *pb.UpdateMultiReleaseReq) (*pb.UpdateMultiReleaseResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("UpdateMultiRelease[%d]| input[%+v]", req.Seq, req)
	response := &pb.UpdateMultiReleaseResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := dm.collector.StatRequest("UpdateMultiRelease", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("UpdateMultiRelease[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := multireleaseaction.NewUpdateAction(dm.viper, dm.smgr, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// CancelMultiRelease cancels target multi release.
func (dm *DataManager) CancelMultiRelease(ctx context.Context, req *pb.CancelMultiReleaseReq) (*pb.CancelMultiReleaseResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("CancelMultiRelease[%d]| input[%+v]", req.Seq, req)
	response := &pb.CancelMultiReleaseResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := dm.collector.StatRequest("CancelMultiRelease", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("CancelMultiRelease[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := multireleaseaction.NewCancelAction(dm.viper, dm.smgr, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// RollbackMultiRelease rollbacks target multi release.
func (dm *DataManager) RollbackMultiRelease(ctx context.Context, req *pb.RollbackMultiReleaseReq) (*pb.RollbackMultiReleaseResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("RollbackMultiRelease[%d]| input[%+v]", req.Seq, req)
	response := &pb.RollbackMultiReleaseResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := dm.collector.StatRequest("RollbackMultiRelease", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("RollbackMultiRelease[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := multireleaseaction.NewRollbackAction(dm.viper, dm.smgr, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// PublishMultiRelease publishs target multi release.
func (dm *DataManager) PublishMultiRelease(ctx context.Context, req *pb.PublishMultiReleaseReq) (*pb.PublishMultiReleaseResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("PublishMultiRelease[%d]| input[%+v]", req.Seq, req)
	response := &pb.PublishMultiReleaseResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := dm.collector.StatRequest("PublishMultiRelease", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("PublishMultiRelease[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := multireleaseaction.NewPublishAction(dm.viper, dm.smgr, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// CreateAppInstance creates new app instance.
func (dm *DataManager) CreateAppInstance(ctx context.Context, req *pb.CreateAppInstanceReq) (*pb.CreateAppInstanceResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("CreateAppInstance[%d]| input[%+v]", req.Seq, req)
	response := &pb.CreateAppInstanceResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := dm.collector.StatRequest("CreateAppInstance", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("CreateAppInstance[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := appinstanceaction.NewCreateAction(dm.viper, dm.smgr, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// QueryHistoryAppInstances returns history app instances.
func (dm *DataManager) QueryHistoryAppInstances(ctx context.Context, req *pb.QueryHistoryAppInstancesReq) (*pb.QueryHistoryAppInstancesResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryHistoryAppInstances[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryHistoryAppInstancesResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := dm.collector.StatRequest("QueryHistoryAppInstances", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryHistoryAppInstances[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := appinstanceaction.NewHistoryAction(dm.viper, dm.smgr, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// QueryReachableAppInstances returns reachable app instances of app/cluster/zone.
func (dm *DataManager) QueryReachableAppInstances(ctx context.Context, req *pb.QueryReachableAppInstancesReq) (*pb.QueryReachableAppInstancesResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryReachableAppInstances[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryReachableAppInstancesResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := dm.collector.StatRequest("QueryReachableAppInstances", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryReachableAppInstances[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := appinstanceaction.NewReachableAction(dm.viper, dm.smgr, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// UpdateAppInstance updates target app instances.
func (dm *DataManager) UpdateAppInstance(ctx context.Context, req *pb.UpdateAppInstanceReq) (*pb.UpdateAppInstanceResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("UpdateAppInstance[%d]| input[%+v]", req.Seq, req)
	response := &pb.UpdateAppInstanceResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := dm.collector.StatRequest("UpdateAppInstance", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("UpdateAppInstance[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := appinstanceaction.NewUpdateAction(dm.viper, dm.smgr, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// QueryMatchedAppInstances returns app instances which matched target strategy.
func (dm *DataManager) QueryMatchedAppInstances(ctx context.Context, req *pb.QueryMatchedAppInstancesReq) (*pb.QueryMatchedAppInstancesResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryMatchedAppInstances[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryMatchedAppInstancesResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := dm.collector.StatRequest("QueryMatchedAppInstances", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryMatchedAppInstances[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := appinstanceaction.NewMatchedAction(dm.viper, dm.smgr, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// QueryEffectedAppInstances returns app instances which effected target release.
func (dm *DataManager) QueryEffectedAppInstances(ctx context.Context, req *pb.QueryEffectedAppInstancesReq) (*pb.QueryEffectedAppInstancesResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryEffectedAppInstances[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryEffectedAppInstancesResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := dm.collector.StatRequest("QueryEffectedAppInstances", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryEffectedAppInstances[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := appinstanceaction.NewEffectedAction(dm.viper, dm.smgr, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// CreateAppInstanceRelease creates new app instance release.
func (dm *DataManager) CreateAppInstanceRelease(ctx context.Context, req *pb.CreateAppInstanceReleaseReq) (*pb.CreateAppInstanceReleaseResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("CreateAppInstanceRelease[%d]| input[%+v]", req.Seq, req)
	response := &pb.CreateAppInstanceReleaseResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := dm.collector.StatRequest("CreateAppInstanceRelease", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("CreateAppInstanceRelease[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := appinstanceaction.NewCreateReleaseAction(dm.viper, dm.smgr, dm.collector, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// QueryAppInstanceRelease returns release of target app instance.
func (dm *DataManager) QueryAppInstanceRelease(ctx context.Context, req *pb.QueryAppInstanceReleaseReq) (*pb.QueryAppInstanceReleaseResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryAppInstanceRelease[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryAppInstanceReleaseResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := dm.collector.StatRequest("QueryAppInstanceRelease", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryAppInstanceRelease[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := appinstanceaction.NewQueryReleaseAction(dm.viper, dm.smgr, dm.collector, dm.configsCache, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// CreateStrategy creates new strategy.
func (dm *DataManager) CreateStrategy(ctx context.Context, req *pb.CreateStrategyReq) (*pb.CreateStrategyResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("CreateStrategy[%d]| input[%+v]", req.Seq, req)
	response := &pb.CreateStrategyResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := dm.collector.StatRequest("CreateStrategy", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("CreateStrategy[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := strategyaction.NewCreateAction(dm.viper, dm.smgr, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// QueryStrategy returns target strategy.
func (dm *DataManager) QueryStrategy(ctx context.Context, req *pb.QueryStrategyReq) (*pb.QueryStrategyResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryStrategy[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryStrategyResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := dm.collector.StatRequest("QueryStrategy", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryStrategy[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := strategyaction.NewQueryAction(dm.viper, dm.smgr, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// QueryStrategyList returns all strategies of target app.
func (dm *DataManager) QueryStrategyList(ctx context.Context, req *pb.QueryStrategyListReq) (*pb.QueryStrategyListResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryStrategyList[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryStrategyListResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := dm.collector.StatRequest("QueryStrategyList", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryStrategyList[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := strategyaction.NewListAction(dm.viper, dm.smgr, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// DeleteStrategy deletes target strategy.
func (dm *DataManager) DeleteStrategy(ctx context.Context, req *pb.DeleteStrategyReq) (*pb.DeleteStrategyResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("DeleteStrategy[%d]| input[%+v]", req.Seq, req)
	response := &pb.DeleteStrategyResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := dm.collector.StatRequest("DeleteStrategy", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("DeleteStrategy[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := strategyaction.NewDeleteAction(dm.viper, dm.smgr, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// CreateShardingDB registers new sharding database instance.
func (dm *DataManager) CreateShardingDB(ctx context.Context, req *pb.CreateShardingDBReq) (*pb.CreateShardingDBResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("CreateShardingDB[%d]| input[%+v]", req.Seq, req)
	response := &pb.CreateShardingDBResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := dm.collector.StatRequest("CreateShardingDB", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("CreateShardingDB[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := shardingdbaction.NewCreateAction(dm.viper, dm.smgr, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// QueryShardingDB returns target sharding database information.
func (dm *DataManager) QueryShardingDB(ctx context.Context, req *pb.QueryShardingDBReq) (*pb.QueryShardingDBResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryShardingDB[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryShardingDBResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := dm.collector.StatRequest("QueryShardingDB", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryShardingDB[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := shardingdbaction.NewQueryAction(dm.viper, dm.smgr, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// QueryShardingDBList returns all sharding databases.
func (dm *DataManager) QueryShardingDBList(ctx context.Context, req *pb.QueryShardingDBListReq) (*pb.QueryShardingDBListResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryShardingDBList[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryShardingDBListResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := dm.collector.StatRequest("QueryShardingDBList", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryShardingDBList[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := shardingdbaction.NewListAction(dm.viper, dm.smgr, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// UpdateShardingDB updates target sharding database.
func (dm *DataManager) UpdateShardingDB(ctx context.Context, req *pb.UpdateShardingDBReq) (*pb.UpdateShardingDBResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("UpdateShardingDB[%d]| input[%+v]", req.Seq, req)
	response := &pb.UpdateShardingDBResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := dm.collector.StatRequest("UpdateShardingDB", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("UpdateShardingDB[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := shardingdbaction.NewUpdateAction(dm.viper, dm.smgr, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// CreateSharding registers new sharding relation.
func (dm *DataManager) CreateSharding(ctx context.Context, req *pb.CreateShardingReq) (*pb.CreateShardingResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("CreateSharding[%d]| input[%+v]", req.Seq, req)
	response := &pb.CreateShardingResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := dm.collector.StatRequest("CreateSharding", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("CreateSharding[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := shardingaction.NewCreateAction(dm.viper, dm.smgr, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// QuerySharding returns target sharding relation.
func (dm *DataManager) QuerySharding(ctx context.Context, req *pb.QueryShardingReq) (*pb.QueryShardingResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QuerySharding[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryShardingResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := dm.collector.StatRequest("QuerySharding", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QuerySharding[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := shardingaction.NewQueryAction(dm.viper, dm.smgr, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// UpdateSharding updates target sharding relation.
func (dm *DataManager) UpdateSharding(ctx context.Context, req *pb.UpdateShardingReq) (*pb.UpdateShardingResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("UpdateSharding[%d]| input[%+v]", req.Seq, req)
	response := &pb.UpdateShardingResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := dm.collector.StatRequest("UpdateSharding", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("UpdateSharding[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := shardingaction.NewUpdateAction(dm.viper, dm.smgr, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// CreateAudit creates new audit.
func (dm *DataManager) CreateAudit(ctx context.Context, req *pb.CreateAuditReq) (*pb.CreateAuditResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("CreateAudit[%d]| input[%+v]", req.Seq, req)
	response := &pb.CreateAuditResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := dm.collector.StatRequest("CreateAudit", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("CreateAudit[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := auditaction.NewCreateAction(dm.viper, dm.smgr, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// QueryAuditList returns history audits.
func (dm *DataManager) QueryAuditList(ctx context.Context, req *pb.QueryAuditListReq) (*pb.QueryAuditListResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryAuditList[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryAuditListResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := dm.collector.StatRequest("QueryAuditList", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryAuditList[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := auditaction.NewListAction(dm.viper, dm.smgr, req, response)
	dm.executor.Execute(action)

	return response, nil
}
