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
	procattraction "bk-bscp/cmd/bscp-datamanager/actions/procattr"
	releaseaction "bk-bscp/cmd/bscp-datamanager/actions/release"
	shardingaction "bk-bscp/cmd/bscp-datamanager/actions/sharding"
	shardingdbaction "bk-bscp/cmd/bscp-datamanager/actions/shardingdb"
	strategyaction "bk-bscp/cmd/bscp-datamanager/actions/strategy"
	templateaction "bk-bscp/cmd/bscp-datamanager/actions/template"
	templatebindingaction "bk-bscp/cmd/bscp-datamanager/actions/templatebinding"
	templatesetaction "bk-bscp/cmd/bscp-datamanager/actions/templateset"
	templateversionaction "bk-bscp/cmd/bscp-datamanager/actions/templateversion"
	variableaction "bk-bscp/cmd/bscp-datamanager/actions/variable"
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

	action := releaseaction.NewPublishAction(dm.viper, dm.smgr, dm.releaseCache, dm.configsCache, req, response)
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

// CreateProcAttr creates new ProcAttr.
func (dm *DataManager) CreateProcAttr(ctx context.Context, req *pb.CreateProcAttrReq) (*pb.CreateProcAttrResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("CreateProcAttr[%d]| input[%+v]", req.Seq, req)
	response := &pb.CreateProcAttrResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := dm.collector.StatRequest("CreateProcAttr", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("CreateProcAttr[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := procattraction.NewCreateAction(dm.viper, dm.smgr, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// QueryHostProcAttr returns ProcAttr of target app on the host.
func (dm *DataManager) QueryHostProcAttr(ctx context.Context, req *pb.QueryHostProcAttrReq) (*pb.QueryHostProcAttrResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryHostProcAttr[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryHostProcAttrResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := dm.collector.StatRequest("QueryHostProcAttr", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryHostProcAttr[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := procattraction.NewQueryAction(dm.viper, dm.smgr, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// QueryHostProcAttrList returns ProcAttr list on target host.
func (dm *DataManager) QueryHostProcAttrList(ctx context.Context, req *pb.QueryHostProcAttrListReq) (*pb.QueryHostProcAttrListResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryHostProcAttrList[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryHostProcAttrListResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := dm.collector.StatRequest("QueryHostProcAttrList", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryHostProcAttrList[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := procattraction.NewHostListAction(dm.viper, dm.smgr, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// QueryAppProcAttrList returns ProcAttr list of target app.
func (dm *DataManager) QueryAppProcAttrList(ctx context.Context, req *pb.QueryAppProcAttrListReq) (*pb.QueryAppProcAttrListResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryAppProcAttrList[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryAppProcAttrListResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := dm.collector.StatRequest("QueryAppProcAttrList", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryAppProcAttrList[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := procattraction.NewAppListAction(dm.viper, dm.smgr, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// UpdateProcAttr updates target app ProcAttr on the host.
func (dm *DataManager) UpdateProcAttr(ctx context.Context, req *pb.UpdateProcAttrReq) (*pb.UpdateProcAttrResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("UpdateProcAttr[%d]| input[%+v]", req.Seq, req)
	response := &pb.UpdateProcAttrResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := dm.collector.StatRequest("UpdateProcAttr", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("UpdateProcAttr[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := procattraction.NewUpdateAction(dm.viper, dm.smgr, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// DeleteProcAttr deletes target app ProcAttr on the host.
func (dm *DataManager) DeleteProcAttr(ctx context.Context, req *pb.DeleteProcAttrReq) (*pb.DeleteProcAttrResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("DeleteProcAttr[%d]| input[%+v]", req.Seq, req)
	response := &pb.DeleteProcAttrResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := dm.collector.StatRequest("DeleteProcAttr", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("DeleteProcAttr[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := procattraction.NewDeleteAction(dm.viper, dm.smgr, req, response)
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

// CreateConfigTemplateSet create config template set
func (dm *DataManager) CreateConfigTemplateSet(ctx context.Context, req *pb.CreateConfigTemplateSetReq) (*pb.CreateConfigTemplateSetResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("CreateConfigTemplateSet[%d]| input[%+v]", req.Seq, req)
	response := &pb.CreateConfigTemplateSetResp{}

	defer func() {
		cost := dm.collector.StatRequest("CreateConfigTemplateSet", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("CreateConfigTemplateSet[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := templatesetaction.NewCreateAction(dm.viper, dm.smgr, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// DeleteConfigTemplateSet delete config template set
func (dm *DataManager) DeleteConfigTemplateSet(ctx context.Context, req *pb.DeleteConfigTemplateSetReq) (*pb.DeleteConfigTemplateSetResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("DeleteConfigTemplateSet[%d]| input[%+v]", req.Seq, req)
	response := &pb.DeleteConfigTemplateSetResp{}

	defer func() {
		cost := dm.collector.StatRequest("DeleteConfigTemplateSet", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("DeleteConfigTemplateSet[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := templatesetaction.NewDeleteAction(dm.viper, dm.smgr, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// UpdateConfigTemplateSet update config template set
func (dm *DataManager) UpdateConfigTemplateSet(ctx context.Context, req *pb.UpdateConfigTemplateSetReq) (*pb.UpdateConfigTemplateSetResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("UpdateConfigTemplateSet[%d]| input[%+v]", req.Seq, req)
	response := &pb.UpdateConfigTemplateSetResp{}

	defer func() {
		cost := dm.collector.StatRequest("UpdateConfigTemplateSet", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("UpdateConfigTemplateSet[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := templatesetaction.NewUpdateAction(dm.viper, dm.smgr, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// QueryConfigTemplateSet query config template set
func (dm *DataManager) QueryConfigTemplateSet(ctx context.Context, req *pb.QueryConfigTemplateSetReq) (*pb.QueryConfigTemplateSetResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryConfigTemplateSet[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryConfigTemplateSetResp{}

	defer func() {
		cost := dm.collector.StatRequest("QueryConfigTemplateSet", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryConfigTemplateSet[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := templatesetaction.NewQueryAction(dm.viper, dm.smgr, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// QueryConfigTemplateSetList query config template set list
func (dm *DataManager) QueryConfigTemplateSetList(ctx context.Context, req *pb.QueryConfigTemplateSetListReq) (*pb.QueryConfigTemplateSetListResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryConfigTemplateSetList[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryConfigTemplateSetListResp{}

	defer func() {
		cost := dm.collector.StatRequest("QueryConfigTemplateSetList", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryConfigTemplateSetList[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := templatesetaction.NewListAction(dm.viper, dm.smgr, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// CreateConfigTemplate create config template
func (dm *DataManager) CreateConfigTemplate(ctx context.Context, req *pb.CreateConfigTemplateReq) (*pb.CreateConfigTemplateResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("CreateConfigTemplate[%d]| input[%+v]", req.Seq, req)
	response := &pb.CreateConfigTemplateResp{}

	defer func() {
		cost := dm.collector.StatRequest("CreateConfigTemplate", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("CreateConfigTemplate[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := templateaction.NewCreateAction(dm.viper, dm.smgr, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// UpdateConfigTemplate create config template
func (dm *DataManager) UpdateConfigTemplate(ctx context.Context, req *pb.UpdateConfigTemplateReq) (*pb.UpdateConfigTemplateResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("UpdateConfigTemplate[%d]| input[%+v]", req.Seq, req)
	response := &pb.UpdateConfigTemplateResp{}

	defer func() {
		cost := dm.collector.StatRequest("UpdateConfigTemplate", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("UpdateConfigTemplate[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := templateaction.NewUpdateAction(dm.viper, dm.smgr, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// DeleteConfigTemplate delete config template
func (dm *DataManager) DeleteConfigTemplate(ctx context.Context, req *pb.DeleteConfigTemplateReq) (*pb.DeleteConfigTemplateResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("DeleteConfigTemplate[%d]| input[%+v]", req.Seq, req)
	response := &pb.DeleteConfigTemplateResp{}

	defer func() {
		cost := dm.collector.StatRequest("DeleteConfigTemplate", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("DeleteConfigTemplate[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := templateaction.NewDeleteAction(dm.viper, dm.smgr, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// QueryConfigTemplate query config template
func (dm *DataManager) QueryConfigTemplate(ctx context.Context, req *pb.QueryConfigTemplateReq) (*pb.QueryConfigTemplateResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryConfigTemplate[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryConfigTemplateResp{}

	defer func() {
		cost := dm.collector.StatRequest("QueryConfigTemplate", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryConfigTemplate[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := templateaction.NewQueryAction(dm.viper, dm.smgr, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// QueryConfigTemplateList query config template list
func (dm *DataManager) QueryConfigTemplateList(ctx context.Context, req *pb.QueryConfigTemplateListReq) (*pb.QueryConfigTemplateListResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryConfigTemplateList[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryConfigTemplateListResp{}

	defer func() {
		cost := dm.collector.StatRequest("QueryConfigTemplateList", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryConfigTemplateList[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := templateaction.NewListAction(dm.viper, dm.smgr, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// CreateTemplateVersion create template version
func (dm *DataManager) CreateTemplateVersion(ctx context.Context, req *pb.CreateTemplateVersionReq) (*pb.CreateTemplateVersionResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("CreateTemplateVersion[%d]| input[%+v]", req.Seq, req)
	response := &pb.CreateTemplateVersionResp{}

	defer func() {
		cost := dm.collector.StatRequest("CreateTemplateVersion", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("CreateTemplateVersion[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := templateversionaction.NewCreateAction(dm.viper, dm.smgr, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// DeleteTemplateVersion delete template version
func (dm *DataManager) DeleteTemplateVersion(ctx context.Context, req *pb.DeleteTemplateVersionReq) (*pb.DeleteTemplateVersionResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("DeleteTemplateVersion[%d]| input[%+v]", req.Seq, req)
	response := &pb.DeleteTemplateVersionResp{}

	defer func() {
		cost := dm.collector.StatRequest("DeleteTemplateVersion", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("DeleteTemplateVersion[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := templateversionaction.NewDeleteAction(dm.viper, dm.smgr, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// UpdateTemplateVersion update template version
func (dm *DataManager) UpdateTemplateVersion(ctx context.Context, req *pb.UpdateTemplateVersionReq) (*pb.UpdateTemplateVersionResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("UpdateTemplateVersion[%d]| input[%+v]", req.Seq, req)
	response := &pb.UpdateTemplateVersionResp{}

	defer func() {
		cost := dm.collector.StatRequest("UpdateTemplateVersion", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("UpdateTemplateVersion[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := templateversionaction.NewUpdateAction(dm.viper, dm.smgr, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// QueryTemplateVersion query template version
func (dm *DataManager) QueryTemplateVersion(ctx context.Context, req *pb.QueryTemplateVersionReq) (*pb.QueryTemplateVersionResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryTemplateVersion[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryTemplateVersionResp{}

	defer func() {
		cost := dm.collector.StatRequest("QueryTemplateVersion", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryTemplateVersion[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := templateversionaction.NewQueryAction(dm.viper, dm.smgr, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// QueryTemplateVersionList query template version list
func (dm *DataManager) QueryTemplateVersionList(ctx context.Context, req *pb.QueryTemplateVersionListReq) (*pb.QueryTemplateVersionListResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryTemplateVersionList[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryTemplateVersionListResp{}

	defer func() {
		cost := dm.collector.StatRequest("QueryTemplateVersionList", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryTemplateVersionList[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := templateversionaction.NewListAction(dm.viper, dm.smgr, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// CreateConfigTemplateBinding create template binding
func (dm *DataManager) CreateConfigTemplateBinding(ctx context.Context, req *pb.CreateConfigTemplateBindingReq) (*pb.CreateConfigTemplateBindingResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("CreateConfigTemplateBinding[%d]| input[%+v]", req.Seq, req)
	response := &pb.CreateConfigTemplateBindingResp{}

	defer func() {
		cost := dm.collector.StatRequest("CreateConfigTemplateBinding", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("CreateConfigTemplateBinding[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := templatebindingaction.NewCreateAction(dm.viper, dm.smgr, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// DeleteConfigTemplateBinding delete template binding
func (dm *DataManager) DeleteConfigTemplateBinding(ctx context.Context, req *pb.DeleteConfigTemplateBindingReq) (*pb.DeleteConfigTemplateBindingResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("DeleteConfigTemplateBinding[%d]| input[%+v]", req.Seq, req)
	response := &pb.DeleteConfigTemplateBindingResp{}

	defer func() {
		cost := dm.collector.StatRequest("DeleteConfigTemplateBinding", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("DeleteConfigTemplateBinding[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := templatebindingaction.NewDeleteAction(dm.viper, dm.smgr, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// UpdateConfigTemplateBinding update template binding
func (dm *DataManager) UpdateConfigTemplateBinding(ctx context.Context, req *pb.UpdateConfigTemplateBindingReq) (*pb.UpdateConfigTemplateBindingResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("UpdateConfigTemplateBinding[%d]| input[%+v]", req.Seq, req)
	response := &pb.UpdateConfigTemplateBindingResp{}

	defer func() {
		cost := dm.collector.StatRequest("UpdateConfigTemplateBinding", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("UpdateConfigTemplateBinding[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := templatebindingaction.NewUpdateAction(dm.viper, dm.smgr, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// QueryConfigTemplateBinding query template binding
func (dm *DataManager) QueryConfigTemplateBinding(ctx context.Context, req *pb.QueryConfigTemplateBindingReq) (*pb.QueryConfigTemplateBindingResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryConfigTemplateBinding[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryConfigTemplateBindingResp{}

	defer func() {
		cost := dm.collector.StatRequest("QueryConfigTemplateBinding", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryConfigTemplateBinding[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := templatebindingaction.NewQueryAction(dm.viper, dm.smgr, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// QueryConfigTemplateBindingList query template binding list
func (dm *DataManager) QueryConfigTemplateBindingList(ctx context.Context, req *pb.QueryConfigTemplateBindingListReq) (*pb.QueryConfigTemplateBindingListResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryConfigTemplateBindingList[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryConfigTemplateBindingListResp{}

	defer func() {
		cost := dm.collector.StatRequest("QueryConfigTemplateBindingList", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryConfigTemplateBindingList[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := templatebindingaction.NewListAction(dm.viper, dm.smgr, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// CreateVariable create variable
func (dm *DataManager) CreateVariable(ctx context.Context, req *pb.CreateVariableReq) (*pb.CreateVariableResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("CreateVariable[%d]| input[%+v]", req.Seq, req)
	response := &pb.CreateVariableResp{}

	defer func() {
		cost := dm.collector.StatRequest("CreateVariable", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("CreateVariable[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := variableaction.NewCreateAction(dm.viper, dm.smgr, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// UpdateVariable update variable
func (dm *DataManager) UpdateVariable(ctx context.Context, req *pb.UpdateVariableReq) (*pb.UpdateVariableResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("UpdateVariable[%d]| input[%+v]", req.Seq, req)
	response := &pb.UpdateVariableResp{}

	defer func() {
		cost := dm.collector.StatRequest("UpdateVariable", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("UpdateVariable[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := variableaction.NewUpdateAction(dm.viper, dm.smgr, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// DeleteVariable delete variable
func (dm *DataManager) DeleteVariable(ctx context.Context, req *pb.DeleteVariableReq) (*pb.DeleteVariableResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("DeleteVariable[%d]| input[%+v]", req.Seq, req)
	response := &pb.DeleteVariableResp{}

	defer func() {
		cost := dm.collector.StatRequest("DeleteVariable", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("DeleteVariable[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := variableaction.NewDeleteAction(dm.viper, dm.smgr, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// QueryVariable query variable
func (dm *DataManager) QueryVariable(ctx context.Context, req *pb.QueryVariableReq) (*pb.QueryVariableResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryVariable[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryVariableResp{}

	defer func() {
		cost := dm.collector.StatRequest("QueryVariable", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryVariable[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := variableaction.NewQueryAction(dm.viper, dm.smgr, req, response)
	dm.executor.Execute(action)

	return response, nil
}

// QueryVariableList query variable list
func (dm *DataManager) QueryVariableList(ctx context.Context, req *pb.QueryVariableListReq) (*pb.QueryVariableListResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryVariableList[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryVariableListResp{}

	defer func() {
		cost := dm.collector.StatRequest("QueryVariableList", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryVariableList[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := variableaction.NewListAction(dm.viper, dm.smgr, req, response)
	dm.executor.Execute(action)

	return response, nil
}
