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

	appinstanceaction "bk-bscp/cmd/bscp-businessserver/actions/appinstance"
	appaction "bk-bscp/cmd/bscp-businessserver/actions/application"
	auditaction "bk-bscp/cmd/bscp-businessserver/actions/audit"
	businessaction "bk-bscp/cmd/bscp-businessserver/actions/business"
	clusteraction "bk-bscp/cmd/bscp-businessserver/actions/cluster"
	commitaction "bk-bscp/cmd/bscp-businessserver/actions/commit"
	configsaction "bk-bscp/cmd/bscp-businessserver/actions/configs"
	configsetaction "bk-bscp/cmd/bscp-businessserver/actions/configset"
	multicommitaction "bk-bscp/cmd/bscp-businessserver/actions/multi-commit"
	multireleaseaction "bk-bscp/cmd/bscp-businessserver/actions/multi-release"
	procattraction "bk-bscp/cmd/bscp-businessserver/actions/procattr"
	releaseaction "bk-bscp/cmd/bscp-businessserver/actions/release"
	reloadaction "bk-bscp/cmd/bscp-businessserver/actions/reload"
	shardingaction "bk-bscp/cmd/bscp-businessserver/actions/sharding"
	shardingdbaction "bk-bscp/cmd/bscp-businessserver/actions/shardingdb"
	strategyaction "bk-bscp/cmd/bscp-businessserver/actions/strategy"
	zoneaction "bk-bscp/cmd/bscp-businessserver/actions/zone"
	pb "bk-bscp/internal/protocol/businessserver"
	pbcommon "bk-bscp/internal/protocol/common"
	"bk-bscp/pkg/logger"
)

// QueryAuthInfo returns auth information of target business.
func (bs *BusinessServer) QueryAuthInfo(ctx context.Context, req *pb.QueryAuthInfoReq) (*pb.QueryAuthInfoResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryAuthInfo[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryAuthInfoResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := bs.collector.StatRequest("QueryAuthInfo", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryAuthInfo[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := businessaction.NewAuthInfoAction(bs.viper, bs.dataMgrCli, req, response)
	bs.executor.Execute(action)

	return response, nil
}

// CreateBusiness creates new business.
func (bs *BusinessServer) CreateBusiness(ctx context.Context, req *pb.CreateBusinessReq) (*pb.CreateBusinessResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("CreateBusiness[%d]| input[%+v]", req.Seq, req)
	response := &pb.CreateBusinessResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := bs.collector.StatRequest("CreateBusiness", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("CreateBusiness[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := businessaction.NewCreateAction(bs.viper, bs.dataMgrCli, req, response)
	bs.executor.Execute(action)

	return response, nil
}

// QueryBusiness returns target business.
func (bs *BusinessServer) QueryBusiness(ctx context.Context, req *pb.QueryBusinessReq) (*pb.QueryBusinessResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryBusiness[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryBusinessResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := bs.collector.StatRequest("QueryBusiness", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryBusiness[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := businessaction.NewQueryAction(bs.viper, bs.dataMgrCli, req, response)
	bs.executor.Execute(action)

	return response, nil
}

// QueryBusinessList returns all businesses.
func (bs *BusinessServer) QueryBusinessList(ctx context.Context, req *pb.QueryBusinessListReq) (*pb.QueryBusinessListResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryBusinessList[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryBusinessListResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := bs.collector.StatRequest("QueryBusinessList", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryBusinessList[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := businessaction.NewListAction(bs.viper, bs.dataMgrCli, req, response)
	bs.executor.Execute(action)

	return response, nil
}

// UpdateBusiness updates target business.
func (bs *BusinessServer) UpdateBusiness(ctx context.Context, req *pb.UpdateBusinessReq) (*pb.UpdateBusinessResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("UpdateBusiness[%d]| input[%+v]", req.Seq, req)
	response := &pb.UpdateBusinessResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := bs.collector.StatRequest("UpdateBusiness", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("UpdateBusiness[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := businessaction.NewUpdateAction(bs.viper, bs.dataMgrCli, req, response)
	bs.executor.Execute(action)

	return response, nil
}

// CreateApp creates new app.
func (bs *BusinessServer) CreateApp(ctx context.Context, req *pb.CreateAppReq) (*pb.CreateAppResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("CreateApp[%d]| input[%+v]", req.Seq, req)
	response := &pb.CreateAppResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := bs.collector.StatRequest("CreateApp", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("CreateApp[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := appaction.NewCreateAction(bs.viper, bs.dataMgrCli, req, response)
	bs.executor.Execute(action)

	return response, nil
}

// QueryApp returns target app.
func (bs *BusinessServer) QueryApp(ctx context.Context, req *pb.QueryAppReq) (*pb.QueryAppResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryApp[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryAppResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := bs.collector.StatRequest("QueryApp", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryApp[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := appaction.NewQueryAction(bs.viper, bs.dataMgrCli, req, response)
	bs.executor.Execute(action)

	return response, nil
}

// QueryAppList returns all apps.
func (bs *BusinessServer) QueryAppList(ctx context.Context, req *pb.QueryAppListReq) (*pb.QueryAppListResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryAppList[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryAppListResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := bs.collector.StatRequest("QueryAppList", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryAppList[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := appaction.NewListAction(bs.viper, bs.dataMgrCli, req, response)
	bs.executor.Execute(action)

	return response, nil
}

// UpdateApp updates target app.
func (bs *BusinessServer) UpdateApp(ctx context.Context, req *pb.UpdateAppReq) (*pb.UpdateAppResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("UpdateApp[%d]| input[%+v]", req.Seq, req)
	response := &pb.UpdateAppResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := bs.collector.StatRequest("UpdateApp", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("UpdateApp[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := appaction.NewUpdateAction(bs.viper, bs.dataMgrCli, req, response)
	bs.executor.Execute(action)

	return response, nil
}

// DeleteApp deletes target app.
func (bs *BusinessServer) DeleteApp(ctx context.Context, req *pb.DeleteAppReq) (*pb.DeleteAppResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("DeleteApp[%d]| input[%+v]", req.Seq, req)
	response := &pb.DeleteAppResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := bs.collector.StatRequest("DeleteApp", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("DeleteApp[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := appaction.NewDeleteAction(bs.viper, bs.dataMgrCli, req, response)
	bs.executor.Execute(action)

	return response, nil
}

// CreateCluster creates new cluster.
func (bs *BusinessServer) CreateCluster(ctx context.Context, req *pb.CreateClusterReq) (*pb.CreateClusterResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("CreateCluster[%d]| input[%+v]", req.Seq, req)
	response := &pb.CreateClusterResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := bs.collector.StatRequest("CreateCluster", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("CreateCluster[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := clusteraction.NewCreateAction(bs.viper, bs.dataMgrCli, req, response)
	bs.executor.Execute(action)

	return response, nil
}

// QueryCluster returns target cluster.
func (bs *BusinessServer) QueryCluster(ctx context.Context, req *pb.QueryClusterReq) (*pb.QueryClusterResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryCluster[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryClusterResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := bs.collector.StatRequest("QueryCluster", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryCluster[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := clusteraction.NewQueryAction(bs.viper, bs.dataMgrCli, req, response)
	bs.executor.Execute(action)

	return response, nil
}

// QueryClusterList returns all clusters.
func (bs *BusinessServer) QueryClusterList(ctx context.Context, req *pb.QueryClusterListReq) (*pb.QueryClusterListResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryClusterList[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryClusterListResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := bs.collector.StatRequest("QueryClusterList", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryClusterList[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := clusteraction.NewListAction(bs.viper, bs.dataMgrCli, req, response)
	bs.executor.Execute(action)

	return response, nil
}

// UpdateCluster updates target cluster.
func (bs *BusinessServer) UpdateCluster(ctx context.Context, req *pb.UpdateClusterReq) (*pb.UpdateClusterResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("UpdateCluster[%d]| input[%+v]", req.Seq, req)
	response := &pb.UpdateClusterResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := bs.collector.StatRequest("UpdateCluster", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("UpdateCluster[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := clusteraction.NewUpdateAction(bs.viper, bs.dataMgrCli, req, response)
	bs.executor.Execute(action)

	return response, nil
}

// DeleteCluster deletes target cluster.
func (bs *BusinessServer) DeleteCluster(ctx context.Context, req *pb.DeleteClusterReq) (*pb.DeleteClusterResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("DeleteCluster[%d]| input[%+v]", req.Seq, req)
	response := &pb.DeleteClusterResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := bs.collector.StatRequest("DeleteCluster", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("DeleteCluster[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := clusteraction.NewDeleteAction(bs.viper, bs.dataMgrCli, req, response)
	bs.executor.Execute(action)

	return response, nil
}

// CreateZone creates new zone.
func (bs *BusinessServer) CreateZone(ctx context.Context, req *pb.CreateZoneReq) (*pb.CreateZoneResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("CreateZone[%d]| input[%+v]", req.Seq, req)
	response := &pb.CreateZoneResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := bs.collector.StatRequest("CreateZone", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("CreateZone[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := zoneaction.NewCreateAction(bs.viper, bs.dataMgrCli, req, response)
	bs.executor.Execute(action)

	return response, nil
}

// QueryZone returns target zone.
func (bs *BusinessServer) QueryZone(ctx context.Context, req *pb.QueryZoneReq) (*pb.QueryZoneResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryZone[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryZoneResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := bs.collector.StatRequest("QueryZone", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryZone[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := zoneaction.NewQueryAction(bs.viper, bs.dataMgrCli, req, response)
	bs.executor.Execute(action)

	return response, nil
}

// QueryZoneList returns all zones.
func (bs *BusinessServer) QueryZoneList(ctx context.Context, req *pb.QueryZoneListReq) (*pb.QueryZoneListResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryZoneList[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryZoneListResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := bs.collector.StatRequest("QueryZoneList", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryZoneList[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := zoneaction.NewListAction(bs.viper, bs.dataMgrCli, req, response)
	bs.executor.Execute(action)

	return response, nil
}

// UpdateZone updates target zone.
func (bs *BusinessServer) UpdateZone(ctx context.Context, req *pb.UpdateZoneReq) (*pb.UpdateZoneResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("UpdateZone[%d]| input[%+v]", req.Seq, req)
	response := &pb.UpdateZoneResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := bs.collector.StatRequest("UpdateZone", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("UpdateZone[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := zoneaction.NewUpdateAction(bs.viper, bs.dataMgrCli, req, response)
	bs.executor.Execute(action)

	return response, nil
}

// DeleteZone deletes target zone.
func (bs *BusinessServer) DeleteZone(ctx context.Context, req *pb.DeleteZoneReq) (*pb.DeleteZoneResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("DeleteZone[%d]| input[%+v]", req.Seq, req)
	response := &pb.DeleteZoneResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := bs.collector.StatRequest("DeleteZone", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("DeleteZone[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := zoneaction.NewDeleteAction(bs.viper, bs.dataMgrCli, req, response)
	bs.executor.Execute(action)

	return response, nil
}

// CreateConfigSet creates new configset.
func (bs *BusinessServer) CreateConfigSet(ctx context.Context, req *pb.CreateConfigSetReq) (*pb.CreateConfigSetResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("CreateConfigSet[%d]| input[%+v]", req.Seq, req)
	response := &pb.CreateConfigSetResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := bs.collector.StatRequest("CreateConfigSet", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("CreateConfigSet[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := configsetaction.NewCreateAction(bs.viper, bs.dataMgrCli, req, response)
	bs.executor.Execute(action)

	return response, nil
}

// QueryConfigSet returns target configset.
func (bs *BusinessServer) QueryConfigSet(ctx context.Context, req *pb.QueryConfigSetReq) (*pb.QueryConfigSetResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryConfigSet[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryConfigSetResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := bs.collector.StatRequest("QueryConfigSet", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryConfigSet[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := configsetaction.NewQueryAction(bs.viper, bs.dataMgrCli, req, response)
	bs.executor.Execute(action)

	return response, nil
}

// QueryConfigSetList returns all configsets.
func (bs *BusinessServer) QueryConfigSetList(ctx context.Context, req *pb.QueryConfigSetListReq) (*pb.QueryConfigSetListResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryConfigSetList[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryConfigSetListResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := bs.collector.StatRequest("QueryConfigSetList", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryConfigSetList[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := configsetaction.NewListAction(bs.viper, bs.dataMgrCli, req, response)
	bs.executor.Execute(action)

	return response, nil
}

// UpdateConfigSet updates target configset.
func (bs *BusinessServer) UpdateConfigSet(ctx context.Context, req *pb.UpdateConfigSetReq) (*pb.UpdateConfigSetResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("UpdateConfigSet[%d]| input[%+v]", req.Seq, req)
	response := &pb.UpdateConfigSetResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := bs.collector.StatRequest("UpdateConfigSet", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("UpdateConfigSet[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := configsetaction.NewUpdateAction(bs.viper, bs.dataMgrCli, req, response)
	bs.executor.Execute(action)

	return response, nil
}

// DeleteConfigSet deletes target configset.
func (bs *BusinessServer) DeleteConfigSet(ctx context.Context, req *pb.DeleteConfigSetReq) (*pb.DeleteConfigSetResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("DeleteConfigSet[%d]| input[%+v]", req.Seq, req)
	response := &pb.DeleteConfigSetResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := bs.collector.StatRequest("DeleteConfigSet", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("DeleteConfigSet[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := configsetaction.NewDeleteAction(bs.viper, bs.dataMgrCli, req, response)
	bs.executor.Execute(action)

	return response, nil
}

// LockConfigSet locks target configset.
func (bs *BusinessServer) LockConfigSet(ctx context.Context, req *pb.LockConfigSetReq) (*pb.LockConfigSetResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("LockConfigSet[%d]| input[%+v]", req.Seq, req)
	response := &pb.LockConfigSetResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := bs.collector.StatRequest("LockConfigSet", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("LockConfigSet[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := configsetaction.NewLockAction(bs.viper, bs.dataMgrCli, req, response)
	bs.executor.Execute(action)

	return response, nil
}

// UnlockConfigSet unlocks target configset.
func (bs *BusinessServer) UnlockConfigSet(ctx context.Context, req *pb.UnlockConfigSetReq) (*pb.UnlockConfigSetResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("UnlockConfigSet[%d]| input[%+v]", req.Seq, req)
	response := &pb.UnlockConfigSetResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := bs.collector.StatRequest("UnlockConfigSet", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("UnlockConfigSet[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := configsetaction.NewUnlockAction(bs.viper, bs.dataMgrCli, req, response)
	bs.executor.Execute(action)

	return response, nil
}

// QueryConfigs returns target configs.
func (bs *BusinessServer) QueryConfigs(ctx context.Context, req *pb.QueryConfigsReq) (*pb.QueryConfigsResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryConfigs[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryConfigsResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := bs.collector.StatRequest("QueryConfigs", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryConfigs[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := configsaction.NewQueryAction(bs.viper, bs.dataMgrCli, req, response)
	bs.executor.Execute(action)

	return response, nil
}

// QueryConfigsList returns all configs.
func (bs *BusinessServer) QueryConfigsList(ctx context.Context, req *pb.QueryConfigsListReq) (*pb.QueryConfigsListResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryConfigsList[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryConfigsListResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := bs.collector.StatRequest("QueryConfigsList", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryConfigsList[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := configsaction.NewListAction(bs.viper, bs.dataMgrCli, req, response)
	bs.executor.Execute(action)

	return response, nil
}

// QueryReleaseConfigs returns configs of target release.
func (bs *BusinessServer) QueryReleaseConfigs(ctx context.Context, req *pb.QueryReleaseConfigsReq) (*pb.QueryReleaseConfigsResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryReleaseConfigs[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryReleaseConfigsResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := bs.collector.StatRequest("QueryReleaseConfigs", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryReleaseConfigs[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := configsaction.NewReleaseAction(bs.viper, bs.dataMgrCli, req, response)
	bs.executor.Execute(action)

	return response, nil
}

// CreateCommit creates new commit.
func (bs *BusinessServer) CreateCommit(ctx context.Context, req *pb.CreateCommitReq) (*pb.CreateCommitResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("CreateCommit[%d]| input[%+v]", req.Seq, req)
	response := &pb.CreateCommitResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := bs.collector.StatRequest("CreateCommit", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("CreateCommit[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := commitaction.NewCreateAction(bs.viper, bs.dataMgrCli, req, response)
	bs.executor.Execute(action)

	return response, nil
}

// QueryCommit returns target commit.
func (bs *BusinessServer) QueryCommit(ctx context.Context, req *pb.QueryCommitReq) (*pb.QueryCommitResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryCommit[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryCommitResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := bs.collector.StatRequest("QueryCommit", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryCommit[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := commitaction.NewQueryAction(bs.viper, bs.dataMgrCli, req, response)
	bs.executor.Execute(action)

	return response, nil
}

// QueryHistoryCommits returns all history commits.
func (bs *BusinessServer) QueryHistoryCommits(ctx context.Context, req *pb.QueryHistoryCommitsReq) (*pb.QueryHistoryCommitsResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryHistoryCommits[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryHistoryCommitsResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := bs.collector.StatRequest("QueryHistoryCommits", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryHistoryCommits[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := commitaction.NewListAction(bs.viper, bs.dataMgrCli, req, response)
	bs.executor.Execute(action)

	return response, nil
}

// UpdateCommit updates target commit.
func (bs *BusinessServer) UpdateCommit(ctx context.Context, req *pb.UpdateCommitReq) (*pb.UpdateCommitResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("UpdateCommit[%d]| input[%+v]", req.Seq, req)
	response := &pb.UpdateCommitResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := bs.collector.StatRequest("UpdateCommit", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("UpdateCommit[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := commitaction.NewUpdateAction(bs.viper, bs.dataMgrCli, req, response)
	bs.executor.Execute(action)

	return response, nil
}

// CancelCommit cancels target commit.
func (bs *BusinessServer) CancelCommit(ctx context.Context, req *pb.CancelCommitReq) (*pb.CancelCommitResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("CancelCommit[%d]| input[%+v]", req.Seq, req)
	response := &pb.CancelCommitResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := bs.collector.StatRequest("CancelCommit", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("CancelCommit[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := commitaction.NewCancelAction(bs.viper, bs.dataMgrCli, req, response)
	bs.executor.Execute(action)

	return response, nil
}

// ConfirmCommit confirms target commit.
func (bs *BusinessServer) ConfirmCommit(ctx context.Context, req *pb.ConfirmCommitReq) (*pb.ConfirmCommitResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("ConfirmCommit[%d]| input[%+v]", req.Seq, req)
	response := &pb.ConfirmCommitResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := bs.collector.StatRequest("ConfirmCommit", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("ConfirmCommit[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := commitaction.NewConfirmAction(bs.viper, bs.dataMgrCli, bs.templateSvrCli, req, response)
	bs.executor.Execute(action)

	return response, nil
}

// PreviewCommit confirms target commit.
func (bs *BusinessServer) PreviewCommit(ctx context.Context, req *pb.PreviewCommitReq) (*pb.PreviewCommitResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("PreviewCommit[%d]| input[%+v]", req.Seq, req)
	response := &pb.PreviewCommitResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := bs.collector.StatRequest("PreviewCommit", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("PreviewCommit[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := commitaction.NewPreviewAction(bs.viper, bs.templateSvrCli, bs.dataMgrCli, req, response)
	bs.executor.Execute(action)

	return response, nil
}

// CreateMultiCommit creates new multi commit.
func (bs *BusinessServer) CreateMultiCommit(ctx context.Context, req *pb.CreateMultiCommitReq) (*pb.CreateMultiCommitResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("CreateMultiCommit[%d]| input[%+v]", req.Seq, req)
	response := &pb.CreateMultiCommitResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := bs.collector.StatRequest("CreateMultiCommit", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("CreateMultiCommit[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := multicommitaction.NewCreateAction(bs.viper, bs.dataMgrCli, req, response)
	bs.executor.Execute(action)

	return response, nil
}

// QueryMultiCommit returns target multi commit.
func (bs *BusinessServer) QueryMultiCommit(ctx context.Context, req *pb.QueryMultiCommitReq) (*pb.QueryMultiCommitResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryMultiCommit[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryMultiCommitResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := bs.collector.StatRequest("QueryMultiCommit", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryMultiCommit[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := multicommitaction.NewQueryAction(bs.viper, bs.dataMgrCli, req, response)
	bs.executor.Execute(action)

	return response, nil
}

// QueryHistoryMultiCommits returns all history multi commits.
func (bs *BusinessServer) QueryHistoryMultiCommits(ctx context.Context, req *pb.QueryHistoryMultiCommitsReq) (*pb.QueryHistoryMultiCommitsResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryHistoryMultiCommits[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryHistoryMultiCommitsResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := bs.collector.StatRequest("QueryHistoryMultiCommits", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryHistoryMultiCommits[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := multicommitaction.NewListAction(bs.viper, bs.dataMgrCli, req, response)
	bs.executor.Execute(action)

	return response, nil
}

// UpdateMultiCommit updates target multi commit.
func (bs *BusinessServer) UpdateMultiCommit(ctx context.Context, req *pb.UpdateMultiCommitReq) (*pb.UpdateMultiCommitResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("UpdateMultiCommit[%d]| input[%+v]", req.Seq, req)
	response := &pb.UpdateMultiCommitResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := bs.collector.StatRequest("UpdateMultiCommit", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("UpdateMultiCommit[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := multicommitaction.NewUpdateAction(bs.viper, bs.dataMgrCli, req, response)
	bs.executor.Execute(action)

	return response, nil
}

// CancelMultiCommit cancels target multi commit.
func (bs *BusinessServer) CancelMultiCommit(ctx context.Context, req *pb.CancelMultiCommitReq) (*pb.CancelMultiCommitResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("CancelMultiCommit[%d]| input[%+v]", req.Seq, req)
	response := &pb.CancelMultiCommitResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := bs.collector.StatRequest("CancelMultiCommit", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("CancelMultiCommit[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := multicommitaction.NewCancelAction(bs.viper, bs.dataMgrCli, req, response)
	bs.executor.Execute(action)

	return response, nil
}

// ConfirmMultiCommit confirms target multi commit.
func (bs *BusinessServer) ConfirmMultiCommit(ctx context.Context, req *pb.ConfirmMultiCommitReq) (*pb.ConfirmMultiCommitResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("ConfirmMultiCommit[%d]| input[%+v]", req.Seq, req)
	response := &pb.ConfirmMultiCommitResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := bs.collector.StatRequest("ConfirmMultiCommit", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("ConfirmMultiCommit[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := multicommitaction.NewConfirmAction(bs.viper, bs.dataMgrCli, bs.templateSvrCli, req, response)
	bs.executor.Execute(action)

	return response, nil
}

// CreateRelease creates new release.
func (bs *BusinessServer) CreateRelease(ctx context.Context, req *pb.CreateReleaseReq) (*pb.CreateReleaseResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("CreateRelease[%d]| input[%+v]", req.Seq, req)
	response := &pb.CreateReleaseResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := bs.collector.StatRequest("CreateRelease", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("CreateRelease[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := releaseaction.NewCreateAction(bs.viper, bs.dataMgrCli, req, response)
	bs.executor.Execute(action)

	return response, nil
}

// QueryRelease returns target release.
func (bs *BusinessServer) QueryRelease(ctx context.Context, req *pb.QueryReleaseReq) (*pb.QueryReleaseResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryRelease[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryReleaseResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := bs.collector.StatRequest("QueryRelease", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryRelease[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := releaseaction.NewQueryAction(bs.viper, bs.dataMgrCli, req, response)
	bs.executor.Execute(action)

	return response, nil
}

// QueryHistoryReleases returns history releases.
func (bs *BusinessServer) QueryHistoryReleases(ctx context.Context, req *pb.QueryHistoryReleasesReq) (*pb.QueryHistoryReleasesResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryHistoryReleases[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryHistoryReleasesResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := bs.collector.StatRequest("QueryHistoryReleases", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryHistoryReleases[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := releaseaction.NewListAction(bs.viper, bs.dataMgrCli, req, response)
	bs.executor.Execute(action)

	return response, nil
}

// UpdateRelease updates target release.
func (bs *BusinessServer) UpdateRelease(ctx context.Context, req *pb.UpdateReleaseReq) (*pb.UpdateReleaseResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("UpdateRelease[%d]| input[%+v]", req.Seq, req)
	response := &pb.UpdateReleaseResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := bs.collector.StatRequest("UpdateRelease", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("UpdateRelease[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := releaseaction.NewUpdateAction(bs.viper, bs.dataMgrCli, req, response)
	bs.executor.Execute(action)

	return response, nil
}

// CancelRelease cancels target release.
func (bs *BusinessServer) CancelRelease(ctx context.Context, req *pb.CancelReleaseReq) (*pb.CancelReleaseResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("CancelRelease[%d]| input[%+v]", req.Seq, req)
	response := &pb.CancelReleaseResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := bs.collector.StatRequest("CancelRelease", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("CancelRelease[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := releaseaction.NewCancelAction(bs.viper, bs.dataMgrCli, req, response)
	bs.executor.Execute(action)

	return response, nil
}

// PublishRelease publishes target release.
func (bs *BusinessServer) PublishRelease(ctx context.Context, req *pb.PublishReleaseReq) (*pb.PublishReleaseResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("PublishRelease[%d]| input[%+v]", req.Seq, req)
	response := &pb.PublishReleaseResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := bs.collector.StatRequest("PublishRelease", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("PublishRelease[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := releaseaction.NewPublishAction(bs.viper, bs.dataMgrCli, bs.bcsControllerCli, bs.gseControllerCli, req, response)
	bs.executor.Execute(action)

	return response, nil
}

// RollbackRelease rollbacks target release.
func (bs *BusinessServer) RollbackRelease(ctx context.Context, req *pb.RollbackReleaseReq) (*pb.RollbackReleaseResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("RollbackRelease[%d]| input[%+v]", req.Seq, req)
	response := &pb.RollbackReleaseResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := bs.collector.StatRequest("RollbackRelease", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("RollbackRelease[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := releaseaction.NewRollbackAction(bs.viper, bs.dataMgrCli, bs.bcsControllerCli, bs.gseControllerCli, req, response)
	bs.executor.Execute(action)

	return response, nil
}

// CreateMultiRelease creates new multi release.
func (bs *BusinessServer) CreateMultiRelease(ctx context.Context, req *pb.CreateMultiReleaseReq) (*pb.CreateMultiReleaseResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("CreateMultiRelease[%d]| input[%+v]", req.Seq, req)
	response := &pb.CreateMultiReleaseResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := bs.collector.StatRequest("CreateMultiRelease", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("CreateMultiRelease[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := multireleaseaction.NewCreateAction(bs.viper, bs.dataMgrCli, req, response)
	bs.executor.Execute(action)

	return response, nil
}

// QueryMultiRelease returns target multi release.
func (bs *BusinessServer) QueryMultiRelease(ctx context.Context, req *pb.QueryMultiReleaseReq) (*pb.QueryMultiReleaseResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryMultiRelease[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryMultiReleaseResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := bs.collector.StatRequest("QueryMultiRelease", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryMultiRelease[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := multireleaseaction.NewQueryAction(bs.viper, bs.dataMgrCli, req, response)
	bs.executor.Execute(action)

	return response, nil
}

// QueryHistoryMultiReleases returns history releases.
func (bs *BusinessServer) QueryHistoryMultiReleases(ctx context.Context, req *pb.QueryHistoryMultiReleasesReq) (*pb.QueryHistoryMultiReleasesResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryHistoryMultiReleases[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryHistoryMultiReleasesResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := bs.collector.StatRequest("QueryHistoryMultiReleases", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryHistoryMultiReleases[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := multireleaseaction.NewListAction(bs.viper, bs.dataMgrCli, req, response)
	bs.executor.Execute(action)

	return response, nil
}

// UpdateMultiRelease updates target multi release.
func (bs *BusinessServer) UpdateMultiRelease(ctx context.Context, req *pb.UpdateMultiReleaseReq) (*pb.UpdateMultiReleaseResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("UpdateMultiRelease[%d]| input[%+v]", req.Seq, req)
	response := &pb.UpdateMultiReleaseResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := bs.collector.StatRequest("UpdateMultiRelease", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("UpdateMultiRelease[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := multireleaseaction.NewUpdateAction(bs.viper, bs.dataMgrCli, req, response)
	bs.executor.Execute(action)

	return response, nil
}

// CancelMultiRelease cancels target multi release.
func (bs *BusinessServer) CancelMultiRelease(ctx context.Context, req *pb.CancelMultiReleaseReq) (*pb.CancelMultiReleaseResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("CancelMultiRelease[%d]| input[%+v]", req.Seq, req)
	response := &pb.CancelMultiReleaseResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := bs.collector.StatRequest("CancelMultiRelease", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("CancelMultiRelease[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := multireleaseaction.NewCancelAction(bs.viper, bs.dataMgrCli, req, response)
	bs.executor.Execute(action)

	return response, nil
}

// PublishMultiRelease publishes target multi release.
func (bs *BusinessServer) PublishMultiRelease(ctx context.Context, req *pb.PublishMultiReleaseReq) (*pb.PublishMultiReleaseResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("PublishMultiRelease[%d]| input[%+v]", req.Seq, req)
	response := &pb.PublishMultiReleaseResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := bs.collector.StatRequest("PublishMultiRelease", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("PublishMultiRelease[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := multireleaseaction.NewPublishAction(bs.viper, bs.dataMgrCli, bs.bcsControllerCli, bs.gseControllerCli, req, response)
	bs.executor.Execute(action)

	return response, nil
}

// RollbackMultiRelease rollbacks target multi release.
func (bs *BusinessServer) RollbackMultiRelease(ctx context.Context, req *pb.RollbackMultiReleaseReq) (*pb.RollbackMultiReleaseResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("RollbackMultiRelease[%d]| input[%+v]", req.Seq, req)
	response := &pb.RollbackMultiReleaseResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := bs.collector.StatRequest("RollbackMultiRelease", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("RollbackMultiRelease[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := multireleaseaction.NewRollbackAction(bs.viper, bs.dataMgrCli, bs.bcsControllerCli, bs.gseControllerCli, req, response)
	bs.executor.Execute(action)

	return response, nil
}

// QueryHistoryAppInstances returns history app instances.
func (bs *BusinessServer) QueryHistoryAppInstances(ctx context.Context, req *pb.QueryHistoryAppInstancesReq) (*pb.QueryHistoryAppInstancesResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryHistoryAppInstances[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryHistoryAppInstancesResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := bs.collector.StatRequest("QueryHistoryAppInstances", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryHistoryAppInstances[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := appinstanceaction.NewHistoryAction(bs.viper, bs.dataMgrCli, req, response)
	bs.executor.Execute(action)

	return response, nil
}

// QueryEffectedAppInstances returns sidecar instances which effected target release of the configset.
func (bs *BusinessServer) QueryEffectedAppInstances(ctx context.Context, req *pb.QueryEffectedAppInstancesReq) (*pb.QueryEffectedAppInstancesResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryEffectedAppInstances[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryEffectedAppInstancesResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := bs.collector.StatRequest("QueryEffectedAppInstances", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryEffectedAppInstances[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := appinstanceaction.NewEffectedAction(bs.viper, bs.dataMgrCli, req, response)
	bs.executor.Execute(action)

	return response, nil
}

// QueryMatchedAppInstances returns sidecar instances which matched target release or strategy.
func (bs *BusinessServer) QueryMatchedAppInstances(ctx context.Context, req *pb.QueryMatchedAppInstancesReq) (*pb.QueryMatchedAppInstancesResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryMatchedAppInstances[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryMatchedAppInstancesResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := bs.collector.StatRequest("QueryMatchedAppInstances", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryMatchedAppInstances[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := appinstanceaction.NewMatchedAction(bs.viper, bs.dataMgrCli, req, response)
	bs.executor.Execute(action)

	return response, nil
}

// QueryReachableAppInstances returns sidecar instances which reachable of the app/cluster/zone.
func (bs *BusinessServer) QueryReachableAppInstances(ctx context.Context, req *pb.QueryReachableAppInstancesReq) (*pb.QueryReachableAppInstancesResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryReachableAppInstances[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryReachableAppInstancesResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := bs.collector.StatRequest("QueryReachableAppInstances", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryReachableAppInstances[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := appinstanceaction.NewReachableAction(bs.viper, bs.dataMgrCli, req, response)
	bs.executor.Execute(action)

	return response, nil
}

// QueryAppInstanceRelease returns release of target app instance.
func (bs *BusinessServer) QueryAppInstanceRelease(ctx context.Context, req *pb.QueryAppInstanceReleaseReq) (*pb.QueryAppInstanceReleaseResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryAppInstanceRelease[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryAppInstanceReleaseResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := bs.collector.StatRequest("QueryAppInstanceRelease", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryAppInstanceRelease[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := appinstanceaction.NewReleaseAction(bs.viper, bs.dataMgrCli, req, response)
	bs.executor.Execute(action)

	return response, nil
}

// CreateStrategy creates new strategy.
func (bs *BusinessServer) CreateStrategy(ctx context.Context, req *pb.CreateStrategyReq) (*pb.CreateStrategyResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("CreateStrategy[%d]| input[%+v]", req.Seq, req)
	response := &pb.CreateStrategyResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := bs.collector.StatRequest("CreateStrategy", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("CreateStrategy[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := strategyaction.NewCreateAction(bs.viper, bs.dataMgrCli, req, response)
	bs.executor.Execute(action)

	return response, nil
}

// QueryStrategy returns target strategy.
func (bs *BusinessServer) QueryStrategy(ctx context.Context, req *pb.QueryStrategyReq) (*pb.QueryStrategyResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryStrategy[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryStrategyResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := bs.collector.StatRequest("QueryStrategy", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryStrategy[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := strategyaction.NewQueryAction(bs.viper, bs.dataMgrCli, req, response)
	bs.executor.Execute(action)

	return response, nil
}

// QueryStrategyList returns all strategies.
func (bs *BusinessServer) QueryStrategyList(ctx context.Context, req *pb.QueryStrategyListReq) (*pb.QueryStrategyListResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryStrategyList[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryStrategyListResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := bs.collector.StatRequest("QueryStrategyList", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryStrategyList[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := strategyaction.NewListAction(bs.viper, bs.dataMgrCli, req, response)
	bs.executor.Execute(action)

	return response, nil
}

// DeleteStrategy deletes target strategy.
func (bs *BusinessServer) DeleteStrategy(ctx context.Context, req *pb.DeleteStrategyReq) (*pb.DeleteStrategyResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("DeleteStrategy[%d]| input[%+v]", req.Seq, req)
	response := &pb.DeleteStrategyResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := bs.collector.StatRequest("DeleteStrategy", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("DeleteStrategy[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := strategyaction.NewDeleteAction(bs.viper, bs.dataMgrCli, req, response)
	bs.executor.Execute(action)

	return response, nil
}

// CreateProcAttr creates new ProcAttr.
func (bs *BusinessServer) CreateProcAttr(ctx context.Context, req *pb.CreateProcAttrReq) (*pb.CreateProcAttrResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("CreateProcAttr[%d]| input[%+v]", req.Seq, req)
	response := &pb.CreateProcAttrResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := bs.collector.StatRequest("CreateProcAttr", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("CreateProcAttr[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := procattraction.NewCreateAction(bs.viper, bs.dataMgrCli, req, response)
	bs.executor.Execute(action)

	return response, nil
}

// QueryHostProcAttr returns ProcAttr of target app on the host.
func (bs *BusinessServer) QueryHostProcAttr(ctx context.Context, req *pb.QueryHostProcAttrReq) (*pb.QueryHostProcAttrResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryHostProcAttr[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryHostProcAttrResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := bs.collector.StatRequest("QueryHostProcAttr", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryHostProcAttr[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := procattraction.NewQueryAction(bs.viper, bs.dataMgrCli, req, response)
	bs.executor.Execute(action)

	return response, nil
}

// QueryHostProcAttrList returns ProcAttr list on target host.
func (bs *BusinessServer) QueryHostProcAttrList(ctx context.Context, req *pb.QueryHostProcAttrListReq) (*pb.QueryHostProcAttrListResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryHostProcAttrList[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryHostProcAttrListResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := bs.collector.StatRequest("QueryHostProcAttrList", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryHostProcAttrList[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := procattraction.NewHostListAction(bs.viper, bs.dataMgrCli, req, response)
	bs.executor.Execute(action)

	return response, nil
}

// QueryAppProcAttrList returns ProcAttr list of target app.
func (bs *BusinessServer) QueryAppProcAttrList(ctx context.Context, req *pb.QueryAppProcAttrListReq) (*pb.QueryAppProcAttrListResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryAppProcAttrList[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryAppProcAttrListResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := bs.collector.StatRequest("QueryAppProcAttrList", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryAppProcAttrList[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := procattraction.NewAppListAction(bs.viper, bs.dataMgrCli, req, response)
	bs.executor.Execute(action)

	return response, nil
}

// UpdateProcAttr updates target app ProcAttr on the host.
func (bs *BusinessServer) UpdateProcAttr(ctx context.Context, req *pb.UpdateProcAttrReq) (*pb.UpdateProcAttrResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("UpdateProcAttr[%d]| input[%+v]", req.Seq, req)
	response := &pb.UpdateProcAttrResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := bs.collector.StatRequest("UpdateProcAttr", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("UpdateProcAttr[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := procattraction.NewUpdateAction(bs.viper, bs.dataMgrCli, req, response)
	bs.executor.Execute(action)

	return response, nil
}

// DeleteProcAttr deletes target app ProcAttr on the host.
func (bs *BusinessServer) DeleteProcAttr(ctx context.Context, req *pb.DeleteProcAttrReq) (*pb.DeleteProcAttrResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("DeleteProcAttr[%d]| input[%+v]", req.Seq, req)
	response := &pb.DeleteProcAttrResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := bs.collector.StatRequest("DeleteProcAttr", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("DeleteProcAttr[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := procattraction.NewDeleteAction(bs.viper, bs.dataMgrCli, req, response)
	bs.executor.Execute(action)

	return response, nil
}

// CreateShardingDB registers new sharding database instance.
func (bs *BusinessServer) CreateShardingDB(ctx context.Context, req *pb.CreateShardingDBReq) (*pb.CreateShardingDBResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("CreateShardingDB[%d]| input[%+v]", req.Seq, req)
	response := &pb.CreateShardingDBResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := bs.collector.StatRequest("CreateShardingDB", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("CreateShardingDB[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := shardingdbaction.NewCreateAction(bs.viper, bs.dataMgrCli, req, response)
	bs.executor.Execute(action)

	return response, nil
}

// QueryShardingDB returns target sharding database.
func (bs *BusinessServer) QueryShardingDB(ctx context.Context, req *pb.QueryShardingDBReq) (*pb.QueryShardingDBResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryShardingDB[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryShardingDBResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := bs.collector.StatRequest("QueryShardingDB", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryShardingDB[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := shardingdbaction.NewQueryAction(bs.viper, bs.dataMgrCli, req, response)
	bs.executor.Execute(action)

	return response, nil
}

// QueryShardingDBList returns all sharding databases.
func (bs *BusinessServer) QueryShardingDBList(ctx context.Context, req *pb.QueryShardingDBListReq) (*pb.QueryShardingDBListResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryShardingDBList[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryShardingDBListResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := bs.collector.StatRequest("QueryShardingDBList", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryShardingDBList[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := shardingdbaction.NewListAction(bs.viper, bs.dataMgrCli, req, response)
	bs.executor.Execute(action)

	return response, nil
}

// UpdateShardingDB updates target sharding database.
func (bs *BusinessServer) UpdateShardingDB(ctx context.Context, req *pb.UpdateShardingDBReq) (*pb.UpdateShardingDBResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("UpdateShardingDB[%d]| input[%+v]", req.Seq, req)
	response := &pb.UpdateShardingDBResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := bs.collector.StatRequest("UpdateShardingDB", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("UpdateShardingDB[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := shardingdbaction.NewUpdateAction(bs.viper, bs.dataMgrCli, req, response)
	bs.executor.Execute(action)

	return response, nil
}

// CreateSharding registers new sharding relation.
func (bs *BusinessServer) CreateSharding(ctx context.Context, req *pb.CreateShardingReq) (*pb.CreateShardingResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("CreateSharding[%d]| input[%+v]", req.Seq, req)
	response := &pb.CreateShardingResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := bs.collector.StatRequest("CreateSharding", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("CreateSharding[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := shardingaction.NewCreateAction(bs.viper, bs.dataMgrCli, req, response)
	bs.executor.Execute(action)

	return response, nil
}

// QuerySharding returns target sharding relation.
func (bs *BusinessServer) QuerySharding(ctx context.Context, req *pb.QueryShardingReq) (*pb.QueryShardingResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QuerySharding[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryShardingResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := bs.collector.StatRequest("QuerySharding", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QuerySharding[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := shardingaction.NewQueryAction(bs.viper, bs.dataMgrCli, req, response)
	bs.executor.Execute(action)

	return response, nil
}

// UpdateSharding updates target sharding relation.
func (bs *BusinessServer) UpdateSharding(ctx context.Context, req *pb.UpdateShardingReq) (*pb.UpdateShardingResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("UpdateSharding[%d]| input[%+v]", req.Seq, req)
	response := &pb.UpdateShardingResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := bs.collector.StatRequest("UpdateSharding", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("UpdateSharding[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := shardingaction.NewUpdateAction(bs.viper, bs.dataMgrCli, req, response)
	bs.executor.Execute(action)

	return response, nil
}

// QueryAuditList returns history audits.
func (bs *BusinessServer) QueryAuditList(ctx context.Context, req *pb.QueryAuditListReq) (*pb.QueryAuditListResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryAuditList[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryAuditListResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := bs.collector.StatRequest("QueryAuditList", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryAuditList[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := auditaction.NewListAction(bs.viper, bs.dataMgrCli, req, response)
	bs.executor.Execute(action)

	return response, nil
}

// Reload reloads target release or multi release.
func (bs *BusinessServer) Reload(ctx context.Context, req *pb.ReloadReq) (*pb.ReloadResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("Reload[%d]| input[%+v]", req.Seq, req)
	response := &pb.ReloadResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := bs.collector.StatRequest("Reload", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("Reload[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := reloadaction.NewReloadAction(bs.viper, bs.dataMgrCli, bs.bcsControllerCli, bs.gseControllerCli, req, response)
	bs.executor.Execute(action)

	return response, nil
}
