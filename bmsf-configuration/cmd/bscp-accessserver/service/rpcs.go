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
	"fmt"
	"time"

	appinstanceaction "bk-bscp/cmd/bscp-accessserver/actions/appinstance"
	appaction "bk-bscp/cmd/bscp-accessserver/actions/application"
	auditaction "bk-bscp/cmd/bscp-accessserver/actions/audit"
	businessaction "bk-bscp/cmd/bscp-accessserver/actions/business"
	clusteraction "bk-bscp/cmd/bscp-accessserver/actions/cluster"
	commitaction "bk-bscp/cmd/bscp-accessserver/actions/commit"
	configsaction "bk-bscp/cmd/bscp-accessserver/actions/configs"
	cfgsetaction "bk-bscp/cmd/bscp-accessserver/actions/configset"
	itgaction "bk-bscp/cmd/bscp-accessserver/actions/integration"
	multicommitaction "bk-bscp/cmd/bscp-accessserver/actions/multi-commit"
	multireleaseaction "bk-bscp/cmd/bscp-accessserver/actions/multi-release"
	procattraction "bk-bscp/cmd/bscp-accessserver/actions/procattr"
	releaseaction "bk-bscp/cmd/bscp-accessserver/actions/release"
	reloadaction "bk-bscp/cmd/bscp-accessserver/actions/reload"
	shardingaction "bk-bscp/cmd/bscp-accessserver/actions/sharding"
	shardingdbaction "bk-bscp/cmd/bscp-accessserver/actions/shardingdb"
	strategyaction "bk-bscp/cmd/bscp-accessserver/actions/strategy"
	tplaction "bk-bscp/cmd/bscp-accessserver/actions/template"
	tplbindingaction "bk-bscp/cmd/bscp-accessserver/actions/templatebinding"
	tplsetaction "bk-bscp/cmd/bscp-accessserver/actions/templateset"
	tplversionaction "bk-bscp/cmd/bscp-accessserver/actions/templateversion"
	variableaction "bk-bscp/cmd/bscp-accessserver/actions/variable"
	zoneaction "bk-bscp/cmd/bscp-accessserver/actions/zone"
	pb "bk-bscp/internal/protocol/accessserver"
	pbbusinessserver "bk-bscp/internal/protocol/businessserver"
	pbcommon "bk-bscp/internal/protocol/common"
	"bk-bscp/pkg/common"
	"bk-bscp/pkg/logger"
)

// check auth in business level.
func (as *AccessServer) authCheck(ctx context.Context, bid string) (pbcommon.ErrCode, string) {
	if !as.viper.GetBool("auth.open") {
		logger.Warn("request auth check not open, bid[%s]", bid)
		return pbcommon.ErrCode_E_OK, ""
	}

	token, err := common.ParseHTTPBasicAuth(ctx)
	if err != nil {
		return pbcommon.ErrCode_E_AS_AUTH_FAILED, err.Error()
	}
	if common.VerifyUserPWD(token, as.viper.GetString("auth.admin")) {
		return pbcommon.ErrCode_E_OK, ""
	}

	// query business auth info.
	r := &pbbusinessserver.QueryAuthInfoReq{
		Seq: common.Sequence(),
		Bid: bid,
	}

	ctx, cancel := context.WithTimeout(context.Background(), as.viper.GetDuration("businessserver.calltimeout"))
	defer cancel()

	logger.V(2).Infof("request to businessserver QueryAuthInfo, %+v", r)

	resp, err := as.businessSvrCli.QueryAuthInfo(ctx, r)
	if err != nil {
		return pbcommon.ErrCode_E_AS_SYSTEM_UNKONW, fmt.Sprintf("request to businessserver QueryAuthInfo, %+v", err)
	}
	auth := resp.Auth

	if !common.VerifyUserPWD(token, auth) {
		return pbcommon.ErrCode_E_AS_NOT_AUTHED, fmt.Sprintf("request op on business[%s] not authed", bid)
	}
	return pbcommon.ErrCode_E_OK, ""
}

// check auth platform.
func (as *AccessServer) authPlatformCheck(ctx context.Context) (pbcommon.ErrCode, string) {
	if !as.viper.GetBool("auth.open") {
		logger.Warn("platform request auth check not open")
		return pbcommon.ErrCode_E_OK, ""
	}

	token, err := common.ParseHTTPBasicAuth(ctx)
	if err != nil {
		return pbcommon.ErrCode_E_AS_AUTH_FAILED, err.Error()
	}
	if !common.VerifyUserPWD(token, as.viper.GetString("auth.platform")) {
		return pbcommon.ErrCode_E_AS_NOT_AUTHED, "platform request op not authed"
	}
	return pbcommon.ErrCode_E_OK, ""
}

// check auth admin.
func (as *AccessServer) authAdminCheck(ctx context.Context) (pbcommon.ErrCode, string) {
	if !as.viper.GetBool("auth.open") {
		logger.Warn("admin request auth check not open")
		return pbcommon.ErrCode_E_OK, ""
	}

	token, err := common.ParseHTTPBasicAuth(ctx)
	if err != nil {
		return pbcommon.ErrCode_E_AS_AUTH_FAILED, err.Error()
	}
	if !common.VerifyUserPWD(token, as.viper.GetString("auth.admin")) {
		return pbcommon.ErrCode_E_AS_NOT_AUTHED, "admin request op not authed"
	}
	return pbcommon.ErrCode_E_OK, ""
}

// CreateBusiness creates new business.
func (as *AccessServer) CreateBusiness(ctx context.Context, req *pb.CreateBusinessReq) (*pb.CreateBusinessResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("CreateBusiness[%d]| input[%+v]", req.Seq, req)
	response := &pb.CreateBusinessResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("CreateBusiness", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("CreateBusiness[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := businessaction.NewCreateAction(as.viper, as.businessSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// QueryBusiness returns target business.
func (as *AccessServer) QueryBusiness(ctx context.Context, req *pb.QueryBusinessReq) (*pb.QueryBusinessResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryBusiness[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryBusinessResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("QueryBusiness", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryBusiness[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := businessaction.NewQueryAction(as.viper, as.businessSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// QueryBusinessList returns all businesses.
func (as *AccessServer) QueryBusinessList(ctx context.Context, req *pb.QueryBusinessListReq) (*pb.QueryBusinessListResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryBusinessList[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryBusinessListResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("QueryBusinessList", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryBusinessList[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := businessaction.NewListAction(as.viper, as.businessSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// UpdateBusiness updates target business.
func (as *AccessServer) UpdateBusiness(ctx context.Context, req *pb.UpdateBusinessReq) (*pb.UpdateBusinessResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("UpdateBusiness[%d]| input[%+v]", req.Seq, req)
	response := &pb.UpdateBusinessResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("UpdateBusiness", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("UpdateBusiness[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authCheck(ctx, req.Bid); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := businessaction.NewUpdateAction(as.viper, as.businessSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// CreateApp creates a new app.
func (as *AccessServer) CreateApp(ctx context.Context, req *pb.CreateAppReq) (*pb.CreateAppResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("CreateApp[%d]| input[%+v]", req.Seq, req)
	response := &pb.CreateAppResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("CreateApp", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("CreateApp[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authCheck(ctx, req.Bid); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := appaction.NewCreateAction(as.viper, as.businessSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// QueryApp returns target app.
func (as *AccessServer) QueryApp(ctx context.Context, req *pb.QueryAppReq) (*pb.QueryAppResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryApp[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryAppResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("QueryApp", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryApp[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authCheck(ctx, req.Bid); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := appaction.NewQueryAction(as.viper, as.businessSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// QueryAppList returns all apps.
func (as *AccessServer) QueryAppList(ctx context.Context, req *pb.QueryAppListReq) (*pb.QueryAppListResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryAppList[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryAppListResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("QueryAppList", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryAppList[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authCheck(ctx, req.Bid); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := appaction.NewListAction(as.viper, as.businessSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// UpdateApp updates target app.
func (as *AccessServer) UpdateApp(ctx context.Context, req *pb.UpdateAppReq) (*pb.UpdateAppResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("UpdateApp[%d]| input[%+v]", req.Seq, req)
	response := &pb.UpdateAppResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("UpdateApp", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("UpdateApp[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authCheck(ctx, req.Bid); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := appaction.NewUpdateAction(as.viper, as.businessSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// DeleteApp deletes target app.
func (as *AccessServer) DeleteApp(ctx context.Context, req *pb.DeleteAppReq) (*pb.DeleteAppResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("DeleteApp[%d]| input[%+v]", req.Seq, req)
	response := &pb.DeleteAppResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("DeleteApp", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("DeleteApp[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authCheck(ctx, req.Bid); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := appaction.NewDeleteAction(as.viper, as.businessSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// CreateCluster creates new cluster.
func (as *AccessServer) CreateCluster(ctx context.Context, req *pb.CreateClusterReq) (*pb.CreateClusterResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("CreateCluster[%d]| input[%+v]", req.Seq, req)
	response := &pb.CreateClusterResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("CreateCluster", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("CreateCluster[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authCheck(ctx, req.Bid); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := clusteraction.NewCreateAction(as.viper, as.businessSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// QueryCluster returns target cluster.
func (as *AccessServer) QueryCluster(ctx context.Context, req *pb.QueryClusterReq) (*pb.QueryClusterResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryCluster[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryClusterResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("QueryCluster", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryCluster[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authCheck(ctx, req.Bid); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := clusteraction.NewQueryAction(as.viper, as.businessSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// QueryClusterList returns all clusters.
func (as *AccessServer) QueryClusterList(ctx context.Context, req *pb.QueryClusterListReq) (*pb.QueryClusterListResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryClusterList[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryClusterListResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("QueryClusterList", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryClusterList[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authCheck(ctx, req.Bid); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := clusteraction.NewListAction(as.viper, as.businessSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// UpdateCluster updates target cluster.
func (as *AccessServer) UpdateCluster(ctx context.Context, req *pb.UpdateClusterReq) (*pb.UpdateClusterResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("UpdateCluster[%d]| input[%+v]", req.Seq, req)
	response := &pb.UpdateClusterResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("UpdateCluster", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("UpdateCluster[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authCheck(ctx, req.Bid); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := clusteraction.NewUpdateAction(as.viper, as.businessSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// DeleteCluster deletes target cluster.
func (as *AccessServer) DeleteCluster(ctx context.Context, req *pb.DeleteClusterReq) (*pb.DeleteClusterResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("DeleteCluster[%d]| input[%+v]", req.Seq, req)
	response := &pb.DeleteClusterResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("DeleteCluster", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("DeleteCluster[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authCheck(ctx, req.Bid); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := clusteraction.NewDeleteAction(as.viper, as.businessSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// CreateZone creates new zone.
func (as *AccessServer) CreateZone(ctx context.Context, req *pb.CreateZoneReq) (*pb.CreateZoneResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("CreateZone[%d]| input[%+v]", req.Seq, req)
	response := &pb.CreateZoneResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("CreateZone", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("CreateZone[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authCheck(ctx, req.Bid); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := zoneaction.NewCreateAction(as.viper, as.businessSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// QueryZone returns target zone.
func (as *AccessServer) QueryZone(ctx context.Context, req *pb.QueryZoneReq) (*pb.QueryZoneResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryZone[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryZoneResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("QueryZone", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryZone[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authCheck(ctx, req.Bid); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := zoneaction.NewQueryAction(as.viper, as.businessSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// QueryZoneList returns all zones.
func (as *AccessServer) QueryZoneList(ctx context.Context, req *pb.QueryZoneListReq) (*pb.QueryZoneListResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryZoneList[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryZoneListResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("QueryZoneList", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryZoneList[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authCheck(ctx, req.Bid); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := zoneaction.NewListAction(as.viper, as.businessSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// UpdateZone updates target zone.
func (as *AccessServer) UpdateZone(ctx context.Context, req *pb.UpdateZoneReq) (*pb.UpdateZoneResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("UpdateZone[%d]| input[%+v]", req.Seq, req)
	response := &pb.UpdateZoneResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("UpdateZone", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("UpdateZone[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authCheck(ctx, req.Bid); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := zoneaction.NewUpdateAction(as.viper, as.businessSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// DeleteZone deletes target zone.
func (as *AccessServer) DeleteZone(ctx context.Context, req *pb.DeleteZoneReq) (*pb.DeleteZoneResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("DeleteZone[%d]| input[%+v]", req.Seq, req)
	response := &pb.DeleteZoneResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("DeleteZone", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("DeleteZone[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authCheck(ctx, req.Bid); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := zoneaction.NewDeleteAction(as.viper, as.businessSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// CreateConfigSet creates new configset.
func (as *AccessServer) CreateConfigSet(ctx context.Context, req *pb.CreateConfigSetReq) (*pb.CreateConfigSetResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("CreateConfigSet[%d]| input[%+v]", req.Seq, req)
	response := &pb.CreateConfigSetResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("CreateConfigSet", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("CreateConfigSet[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authCheck(ctx, req.Bid); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := cfgsetaction.NewCreateAction(as.viper, as.businessSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// QueryConfigSet returns target configset.
func (as *AccessServer) QueryConfigSet(ctx context.Context, req *pb.QueryConfigSetReq) (*pb.QueryConfigSetResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryConfigSet[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryConfigSetResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("QueryConfigSet", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryConfigSet[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authCheck(ctx, req.Bid); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := cfgsetaction.NewQueryAction(as.viper, as.businessSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// QueryConfigSetList returns all configsets.
func (as *AccessServer) QueryConfigSetList(ctx context.Context, req *pb.QueryConfigSetListReq) (*pb.QueryConfigSetListResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryConfigSetList[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryConfigSetListResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("QueryConfigSetList", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryConfigSetList[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authCheck(ctx, req.Bid); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := cfgsetaction.NewListAction(as.viper, as.businessSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// UpdateConfigSet updates target configset.
func (as *AccessServer) UpdateConfigSet(ctx context.Context, req *pb.UpdateConfigSetReq) (*pb.UpdateConfigSetResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("UpdateConfigSet[%d]| input[%+v]", req.Seq, req)
	response := &pb.UpdateConfigSetResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("UpdateConfigSet", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("UpdateConfigSet[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authCheck(ctx, req.Bid); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := cfgsetaction.NewUpdateAction(as.viper, as.businessSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// DeleteConfigSet deletes target configset.
func (as *AccessServer) DeleteConfigSet(ctx context.Context, req *pb.DeleteConfigSetReq) (*pb.DeleteConfigSetResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("DeleteConfigSet[%d]| input[%+v]", req.Seq, req)
	response := &pb.DeleteConfigSetResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("DeleteConfigSet", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("DeleteConfigSet[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authCheck(ctx, req.Bid); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := cfgsetaction.NewDeleteAction(as.viper, as.businessSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// LockConfigSet locks target configset.
func (as *AccessServer) LockConfigSet(ctx context.Context, req *pb.LockConfigSetReq) (*pb.LockConfigSetResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("LockConfigSet[%d]| input[%+v]", req.Seq, req)
	response := &pb.LockConfigSetResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("LockConfigSet", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("LockConfigSet[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authCheck(ctx, req.Bid); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := cfgsetaction.NewLockAction(as.viper, as.businessSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// UnlockConfigSet unlocks target configset.
func (as *AccessServer) UnlockConfigSet(ctx context.Context, req *pb.UnlockConfigSetReq) (*pb.UnlockConfigSetResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("UnlockConfigSet[%d]| input[%+v]", req.Seq, req)
	response := &pb.UnlockConfigSetResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("UnlockConfigSet", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("UnlockConfigSet[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authCheck(ctx, req.Bid); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := cfgsetaction.NewUnlockAction(as.viper, as.businessSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// QueryConfigs returns target configs.
func (as *AccessServer) QueryConfigs(ctx context.Context, req *pb.QueryConfigsReq) (*pb.QueryConfigsResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryConfigs[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryConfigsResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("QueryConfigs", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryConfigs[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authCheck(ctx, req.Bid); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := configsaction.NewQueryAction(as.viper, as.businessSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// QueryConfigsList returns all configs.
func (as *AccessServer) QueryConfigsList(ctx context.Context, req *pb.QueryConfigsListReq) (*pb.QueryConfigsListResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryConfigsList[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryConfigsListResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("QueryConfigsList", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryConfigsList[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authCheck(ctx, req.Bid); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := configsaction.NewListAction(as.viper, as.businessSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// QueryReleaseConfigs returns configs of target release.
func (as *AccessServer) QueryReleaseConfigs(ctx context.Context, req *pb.QueryReleaseConfigsReq) (*pb.QueryReleaseConfigsResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryReleaseConfigs[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryReleaseConfigsResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("QueryReleaseConfigs", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryReleaseConfigs[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authCheck(ctx, req.Bid); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := configsaction.NewReleaseAction(as.viper, as.businessSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// CreateCommit creates new commit.
func (as *AccessServer) CreateCommit(ctx context.Context, req *pb.CreateCommitReq) (*pb.CreateCommitResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("CreateCommit[%d]| input[%+v]", req.Seq, req)
	response := &pb.CreateCommitResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("CreateCommit", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("CreateCommit[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authCheck(ctx, req.Bid); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := commitaction.NewCreateAction(as.viper, as.businessSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// QueryCommit returns target commit.
func (as *AccessServer) QueryCommit(ctx context.Context, req *pb.QueryCommitReq) (*pb.QueryCommitResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryCommit[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryCommitResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("QueryCommit", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryCommit[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authCheck(ctx, req.Bid); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := commitaction.NewQueryAction(as.viper, as.businessSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// QueryHistoryCommits returns history commits.
func (as *AccessServer) QueryHistoryCommits(ctx context.Context, req *pb.QueryHistoryCommitsReq) (*pb.QueryHistoryCommitsResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryHistoryCommits[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryHistoryCommitsResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("QueryHistoryCommits", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryHistoryCommits[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authCheck(ctx, req.Bid); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := commitaction.NewListAction(as.viper, as.businessSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// UpdateCommit updates target commit.
func (as *AccessServer) UpdateCommit(ctx context.Context, req *pb.UpdateCommitReq) (*pb.UpdateCommitResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("UpdateCommit[%d]| input[%+v]", req.Seq, req)
	response := &pb.UpdateCommitResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("UpdateCommit", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("UpdateCommit[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authCheck(ctx, req.Bid); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := commitaction.NewUpdateAction(as.viper, as.businessSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// CancelCommit cancels target commit.
func (as *AccessServer) CancelCommit(ctx context.Context, req *pb.CancelCommitReq) (*pb.CancelCommitResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("CancelCommit[%d]| input[%+v]", req.Seq, req)
	response := &pb.CancelCommitResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("CancelCommit", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("CancelCommit[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authCheck(ctx, req.Bid); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := commitaction.NewCancelAction(as.viper, as.businessSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// ConfirmCommit confirms target commit.
func (as *AccessServer) ConfirmCommit(ctx context.Context, req *pb.ConfirmCommitReq) (*pb.ConfirmCommitResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("ConfirmCommit[%d]| input[%+v]", req.Seq, req)
	response := &pb.ConfirmCommitResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("ConfirmCommit", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("ConfirmCommit[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authCheck(ctx, req.Bid); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := commitaction.NewConfirmAction(as.viper, as.businessSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// PreviewCommit confirms target commit.
func (as *AccessServer) PreviewCommit(ctx context.Context, req *pb.PreviewCommitReq) (*pb.PreviewCommitResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("PreviewCommit[%d]| input[%+v]", req.Seq, req)
	response := &pb.PreviewCommitResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("PreviewCommit", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("PreviewCommit[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	action := commitaction.NewPreviewAction(as.viper, as.businessSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// CreateMultiCommit creates new multi commit.
func (as *AccessServer) CreateMultiCommit(ctx context.Context, req *pb.CreateMultiCommitReq) (*pb.CreateMultiCommitResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("CreateMultiCommit[%d]| input[%+v]", req.Seq, req)
	response := &pb.CreateMultiCommitResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("CreateMultiCommit", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("CreateMultiCommit[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authCheck(ctx, req.Bid); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := multicommitaction.NewCreateAction(as.viper, as.businessSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// QueryMultiCommit returns target multi commit.
func (as *AccessServer) QueryMultiCommit(ctx context.Context, req *pb.QueryMultiCommitReq) (*pb.QueryMultiCommitResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryMultiCommit[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryMultiCommitResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("QueryMultiCommit", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryMultiCommit[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authCheck(ctx, req.Bid); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := multicommitaction.NewQueryAction(as.viper, as.businessSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// QueryHistoryMultiCommits returns history multi commits.
func (as *AccessServer) QueryHistoryMultiCommits(ctx context.Context, req *pb.QueryHistoryMultiCommitsReq) (*pb.QueryHistoryMultiCommitsResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryHistoryMultiCommits[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryHistoryMultiCommitsResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("QueryHistoryMultiCommits", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryHistoryMultiCommits[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authCheck(ctx, req.Bid); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := multicommitaction.NewListAction(as.viper, as.businessSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// UpdateMultiCommit updates target multi commit.
func (as *AccessServer) UpdateMultiCommit(ctx context.Context, req *pb.UpdateMultiCommitReq) (*pb.UpdateMultiCommitResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("UpdateMultiCommit[%d]| input[%+v]", req.Seq, req)
	response := &pb.UpdateMultiCommitResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("UpdateMultiCommit", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("UpdateMultiCommit[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authCheck(ctx, req.Bid); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := multicommitaction.NewUpdateAction(as.viper, as.businessSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// CancelMultiCommit cancels target multi commit.
func (as *AccessServer) CancelMultiCommit(ctx context.Context, req *pb.CancelMultiCommitReq) (*pb.CancelMultiCommitResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("CancelMultiCommit[%d]| input[%+v]", req.Seq, req)
	response := &pb.CancelMultiCommitResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("CancelMultiCommit", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("CancelMultiCommit[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authCheck(ctx, req.Bid); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := multicommitaction.NewCancelAction(as.viper, as.businessSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// ConfirmMultiCommit confirms target multi commit.
func (as *AccessServer) ConfirmMultiCommit(ctx context.Context, req *pb.ConfirmMultiCommitReq) (*pb.ConfirmMultiCommitResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("ConfirmMultiCommit[%d]| input[%+v]", req.Seq, req)
	response := &pb.ConfirmMultiCommitResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("ConfirmMultiCommit", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("ConfirmMultiCommit[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authCheck(ctx, req.Bid); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := multicommitaction.NewConfirmAction(as.viper, as.businessSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// CreateRelease creates new release.
func (as *AccessServer) CreateRelease(ctx context.Context, req *pb.CreateReleaseReq) (*pb.CreateReleaseResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("CreateRelease[%d]| input[%+v]", req.Seq, req)
	response := &pb.CreateReleaseResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("CreateRelease", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("CreateRelease[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authCheck(ctx, req.Bid); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := releaseaction.NewCreateAction(as.viper, as.businessSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// QueryRelease returns target release.
func (as *AccessServer) QueryRelease(ctx context.Context, req *pb.QueryReleaseReq) (*pb.QueryReleaseResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryRelease[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryReleaseResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("QueryRelease", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryRelease[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authCheck(ctx, req.Bid); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := releaseaction.NewQueryAction(as.viper, as.businessSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// UpdateRelease updates tergate release.
func (as *AccessServer) UpdateRelease(ctx context.Context, req *pb.UpdateReleaseReq) (*pb.UpdateReleaseResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("UpdateRelease[%d]| input[%+v]", req.Seq, req)
	response := &pb.UpdateReleaseResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("UpdateRelease", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("UpdateRelease[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authCheck(ctx, req.Bid); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := releaseaction.NewUpdateAction(as.viper, as.businessSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// CancelRelease cancels tergate release.
func (as *AccessServer) CancelRelease(ctx context.Context, req *pb.CancelReleaseReq) (*pb.CancelReleaseResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("CancelRelease[%d]| input[%+v]", req.Seq, req)
	response := &pb.CancelReleaseResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("CancelRelease", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("CancelRelease[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authCheck(ctx, req.Bid); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := releaseaction.NewCancelAction(as.viper, as.businessSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// PublishRelease publishs target release.
func (as *AccessServer) PublishRelease(ctx context.Context, req *pb.PublishReleaseReq) (*pb.PublishReleaseResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("PublishRelease[%d]| input[%+v]", req.Seq, req)
	response := &pb.PublishReleaseResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("PublishRelease", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("PublishRelease[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authCheck(ctx, req.Bid); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := releaseaction.NewPublishAction(as.viper, as.businessSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// RollbackRelease rollbacks target release.
func (as *AccessServer) RollbackRelease(ctx context.Context, req *pb.RollbackReleaseReq) (*pb.RollbackReleaseResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("RollbackRelease[%d]| input[%+v]", req.Seq, req)
	response := &pb.RollbackReleaseResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("RollbackRelease", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("RollbackRelease[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authCheck(ctx, req.Bid); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := releaseaction.NewRollbackAction(as.viper, as.businessSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// QueryHistoryReleases returns history releases.
func (as *AccessServer) QueryHistoryReleases(ctx context.Context, req *pb.QueryHistoryReleasesReq) (*pb.QueryHistoryReleasesResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryHistoryReleases[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryHistoryReleasesResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("QueryHistoryReleases", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryHistoryReleases[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authCheck(ctx, req.Bid); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := releaseaction.NewListAction(as.viper, as.businessSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// CreateMultiRelease creates new multi release.
func (as *AccessServer) CreateMultiRelease(ctx context.Context, req *pb.CreateMultiReleaseReq) (*pb.CreateMultiReleaseResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("CreateMultiRelease[%d]| input[%+v]", req.Seq, req)
	response := &pb.CreateMultiReleaseResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("CreateMultiRelease", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("CreateMultiRelease[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authCheck(ctx, req.Bid); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := multireleaseaction.NewCreateAction(as.viper, as.businessSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// QueryMultiRelease returns target multi release.
func (as *AccessServer) QueryMultiRelease(ctx context.Context, req *pb.QueryMultiReleaseReq) (*pb.QueryMultiReleaseResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryMultiRelease[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryMultiReleaseResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("QueryMultiRelease", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryMultiRelease[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authCheck(ctx, req.Bid); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := multireleaseaction.NewQueryAction(as.viper, as.businessSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// UpdateMultiRelease updates tergate multi release.
func (as *AccessServer) UpdateMultiRelease(ctx context.Context, req *pb.UpdateMultiReleaseReq) (*pb.UpdateMultiReleaseResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("UpdateMultiRelease[%d]| input[%+v]", req.Seq, req)
	response := &pb.UpdateMultiReleaseResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("UpdateMultiRelease", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("UpdateMultiRelease[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authCheck(ctx, req.Bid); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := multireleaseaction.NewUpdateAction(as.viper, as.businessSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// CancelMultiRelease cancels tergate multi release.
func (as *AccessServer) CancelMultiRelease(ctx context.Context, req *pb.CancelMultiReleaseReq) (*pb.CancelMultiReleaseResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("CancelMultiRelease[%d]| input[%+v]", req.Seq, req)
	response := &pb.CancelMultiReleaseResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("CancelMultiRelease", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("CancelMultiRelease[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authCheck(ctx, req.Bid); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := multireleaseaction.NewCancelAction(as.viper, as.businessSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// PublishMultiRelease publishs target multi release.
func (as *AccessServer) PublishMultiRelease(ctx context.Context, req *pb.PublishMultiReleaseReq) (*pb.PublishMultiReleaseResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("PublishMultiRelease[%d]| input[%+v]", req.Seq, req)
	response := &pb.PublishMultiReleaseResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("PublishMultiRelease", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("PublishMultiRelease[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authCheck(ctx, req.Bid); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := multireleaseaction.NewPublishAction(as.viper, as.businessSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// RollbackMultiRelease rollbacks target multi release.
func (as *AccessServer) RollbackMultiRelease(ctx context.Context, req *pb.RollbackMultiReleaseReq) (*pb.RollbackMultiReleaseResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("RollbackMultiRelease[%d]| input[%+v]", req.Seq, req)
	response := &pb.RollbackMultiReleaseResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("RollbackMultiRelease", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("RollbackMultiRelease[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authCheck(ctx, req.Bid); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := multireleaseaction.NewRollbackAction(as.viper, as.businessSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// QueryHistoryMultiReleases returns history multi releases.
func (as *AccessServer) QueryHistoryMultiReleases(ctx context.Context, req *pb.QueryHistoryMultiReleasesReq) (*pb.QueryHistoryMultiReleasesResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryHistoryMultiReleases[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryHistoryMultiReleasesResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("QueryHistoryMultiReleases", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryHistoryMultiReleases[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authCheck(ctx, req.Bid); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := multireleaseaction.NewListAction(as.viper, as.businessSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// QueryHistoryAppInstances returns history sidecar instances.
func (as *AccessServer) QueryHistoryAppInstances(ctx context.Context, req *pb.QueryHistoryAppInstancesReq) (*pb.QueryHistoryAppInstancesResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryHistoryAppInstances[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryHistoryAppInstancesResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("QueryHistoryAppInstances", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryHistoryAppInstances[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authCheck(ctx, req.Bid); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := appinstanceaction.NewHistoryAction(as.viper, as.businessSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// QueryEffectedAppInstances returns sidecar instances which effected target release of the configset.
func (as *AccessServer) QueryEffectedAppInstances(ctx context.Context, req *pb.QueryEffectedAppInstancesReq) (*pb.QueryEffectedAppInstancesResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryEffectedAppInstances[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryEffectedAppInstancesResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("QueryEffectedAppInstances", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryEffectedAppInstances[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authCheck(ctx, req.Bid); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := appinstanceaction.NewEffectedAction(as.viper, as.businessSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// QueryMatchedAppInstances returns sidecar instances which matched target release or strategy.
func (as *AccessServer) QueryMatchedAppInstances(ctx context.Context, req *pb.QueryMatchedAppInstancesReq) (*pb.QueryMatchedAppInstancesResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryMatchedAppInstances[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryMatchedAppInstancesResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("QueryMatchedAppInstances", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryMatchedAppInstances[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authCheck(ctx, req.Bid); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := appinstanceaction.NewMatchedAction(as.viper, as.businessSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// QueryReachableAppInstances returns sidecar instances which reachable of the app/cluster/zone.
func (as *AccessServer) QueryReachableAppInstances(ctx context.Context, req *pb.QueryReachableAppInstancesReq) (*pb.QueryReachableAppInstancesResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryReachableAppInstances[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryReachableAppInstancesResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("QueryReachableAppInstances", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryReachableAppInstances[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authCheck(ctx, req.Bid); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := appinstanceaction.NewReachableAction(as.viper, as.businessSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// QueryAppInstanceRelease returns release of target app instance.
func (as *AccessServer) QueryAppInstanceRelease(ctx context.Context, req *pb.QueryAppInstanceReleaseReq) (*pb.QueryAppInstanceReleaseResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryAppInstanceRelease[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryAppInstanceReleaseResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("QueryAppInstanceRelease", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryAppInstanceRelease[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authCheck(ctx, req.Bid); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := appinstanceaction.NewReleaseAction(as.viper, as.businessSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// CreateStrategy creates new strategy.
func (as *AccessServer) CreateStrategy(ctx context.Context, req *pb.CreateStrategyReq) (*pb.CreateStrategyResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("CreateStrategy[%d]| input[%+v]", req.Seq, req)
	response := &pb.CreateStrategyResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("CreateStrategy", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("CreateStrategy[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authCheck(ctx, req.Bid); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := strategyaction.NewCreateAction(as.viper, as.businessSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// QueryStrategy returns target strategy.
func (as *AccessServer) QueryStrategy(ctx context.Context, req *pb.QueryStrategyReq) (*pb.QueryStrategyResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryStrategy[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryStrategyResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("QueryStrategy", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryStrategy[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authCheck(ctx, req.Bid); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := strategyaction.NewQueryAction(as.viper, as.businessSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// QueryStrategyList returns all strategies of target app.
func (as *AccessServer) QueryStrategyList(ctx context.Context, req *pb.QueryStrategyListReq) (*pb.QueryStrategyListResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryStrategyList[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryStrategyListResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("QueryStrategyList", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryStrategyList[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authCheck(ctx, req.Bid); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := strategyaction.NewListAction(as.viper, as.businessSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// DeleteStrategy deletes target strategy.
func (as *AccessServer) DeleteStrategy(ctx context.Context, req *pb.DeleteStrategyReq) (*pb.DeleteStrategyResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("DeleteStrategy[%d]| input[%+v]", req.Seq, req)
	response := &pb.DeleteStrategyResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("DeleteStrategy", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("DeleteStrategy[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authCheck(ctx, req.Bid); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := strategyaction.NewDeleteAction(as.viper, as.businessSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// CreateProcAttr creates new ProcAttr.
func (as *AccessServer) CreateProcAttr(ctx context.Context, req *pb.CreateProcAttrReq) (*pb.CreateProcAttrResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("CreateProcAttr[%d]| input[%+v]", req.Seq, req)
	response := &pb.CreateProcAttrResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("CreateProcAttr", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("CreateProcAttr[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authPlatformCheck(ctx); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := procattraction.NewCreateAction(as.viper, as.businessSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// QueryHostProcAttr returns ProcAttr of target app on the host.
func (as *AccessServer) QueryHostProcAttr(ctx context.Context, req *pb.QueryHostProcAttrReq) (*pb.QueryHostProcAttrResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryHostProcAttr[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryHostProcAttrResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("QueryHostProcAttr", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryHostProcAttr[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authPlatformCheck(ctx); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := procattraction.NewQueryAction(as.viper, as.businessSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// QueryHostProcAttrList returns ProcAttr list on target host.
func (as *AccessServer) QueryHostProcAttrList(ctx context.Context, req *pb.QueryHostProcAttrListReq) (*pb.QueryHostProcAttrListResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryHostProcAttrList[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryHostProcAttrListResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("QueryHostProcAttrList", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryHostProcAttrList[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authPlatformCheck(ctx); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := procattraction.NewHostListAction(as.viper, as.businessSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// QueryAppProcAttrList returns ProcAttr list of target app.
func (as *AccessServer) QueryAppProcAttrList(ctx context.Context, req *pb.QueryAppProcAttrListReq) (*pb.QueryAppProcAttrListResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryAppProcAttrList[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryAppProcAttrListResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("QueryAppProcAttrList", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryAppProcAttrList[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authPlatformCheck(ctx); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := procattraction.NewAppListAction(as.viper, as.businessSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// UpdateProcAttr updates target app ProcAttr on the host.
func (as *AccessServer) UpdateProcAttr(ctx context.Context, req *pb.UpdateProcAttrReq) (*pb.UpdateProcAttrResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("UpdateProcAttr[%d]| input[%+v]", req.Seq, req)
	response := &pb.UpdateProcAttrResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("UpdateProcAttr", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("UpdateProcAttr[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authPlatformCheck(ctx); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := procattraction.NewUpdateAction(as.viper, as.businessSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// DeleteProcAttr deletes target app ProcAttr on the host.
func (as *AccessServer) DeleteProcAttr(ctx context.Context, req *pb.DeleteProcAttrReq) (*pb.DeleteProcAttrResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("DeleteProcAttr[%d]| input[%+v]", req.Seq, req)
	response := &pb.DeleteProcAttrResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("DeleteProcAttr", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("DeleteProcAttr[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authPlatformCheck(ctx); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := procattraction.NewDeleteAction(as.viper, as.businessSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// CreateShardingDB registers new sharding database instance.
func (as *AccessServer) CreateShardingDB(ctx context.Context, req *pb.CreateShardingDBReq) (*pb.CreateShardingDBResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("CreateShardingDB[%d]| input[%+v]", req.Seq, req)
	response := &pb.CreateShardingDBResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("CreateShardingDB", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("CreateShardingDB[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authAdminCheck(ctx); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := shardingdbaction.NewCreateAction(as.viper, as.businessSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// QueryShardingDB returns target sharding database.
func (as *AccessServer) QueryShardingDB(ctx context.Context, req *pb.QueryShardingDBReq) (*pb.QueryShardingDBResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryShardingDB[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryShardingDBResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("QueryShardingDB", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryShardingDB[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authAdminCheck(ctx); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := shardingdbaction.NewQueryAction(as.viper, as.businessSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// QueryShardingDBList returns all sharding databases.
func (as *AccessServer) QueryShardingDBList(ctx context.Context, req *pb.QueryShardingDBListReq) (*pb.QueryShardingDBListResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryShardingDBList[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryShardingDBListResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("QueryShardingDBList", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryShardingDBList[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authAdminCheck(ctx); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := shardingdbaction.NewListAction(as.viper, as.businessSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// UpdateShardingDB updates target sharding database.
func (as *AccessServer) UpdateShardingDB(ctx context.Context, req *pb.UpdateShardingDBReq) (*pb.UpdateShardingDBResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("UpdateShardingDB[%d]| input[%+v]", req.Seq, req)
	response := &pb.UpdateShardingDBResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("UpdateShardingDB", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("UpdateShardingDB[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authAdminCheck(ctx); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := shardingdbaction.NewUpdateAction(as.viper, as.businessSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// CreateSharding creates new sharding.
func (as *AccessServer) CreateSharding(ctx context.Context, req *pb.CreateShardingReq) (*pb.CreateShardingResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("CreateSharding[%d]| input[%+v]", req.Seq, req)
	response := &pb.CreateShardingResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("CreateSharding", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("CreateSharding[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authAdminCheck(ctx); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := shardingaction.NewCreateAction(as.viper, as.businessSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// QuerySharding returns target sharding.
func (as *AccessServer) QuerySharding(ctx context.Context, req *pb.QueryShardingReq) (*pb.QueryShardingResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QuerySharding[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryShardingResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("QuerySharding", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QuerySharding[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authAdminCheck(ctx); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := shardingaction.NewQueryAction(as.viper, as.businessSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// UpdateSharding updates target sharding.
func (as *AccessServer) UpdateSharding(ctx context.Context, req *pb.UpdateShardingReq) (*pb.UpdateShardingResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("UpdateSharding[%d]| input[%+v]", req.Seq, req)
	response := &pb.UpdateShardingResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("UpdateSharding", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("UpdateSharding[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authAdminCheck(ctx); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := shardingaction.NewUpdateAction(as.viper, as.businessSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// QueryAuditList returns history audits.
func (as *AccessServer) QueryAuditList(ctx context.Context, req *pb.QueryAuditListReq) (*pb.QueryAuditListResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryAuditList[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryAuditListResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("QueryAuditList", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryAuditList[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authCheck(ctx, req.Bid); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := auditaction.NewListAction(as.viper, as.businessSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// Integrate handles logic integrations.
func (as *AccessServer) Integrate(ctx context.Context, req *pb.IntegrateReq) (*pb.IntegrateResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("Integrate[%d]| input[%+v]", req.Seq, req)
	response := &pb.IntegrateResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("Integrate", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("Integrate[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	// TODO add auth in itg logic.

	action := itgaction.NewIntegrateAction(as.viper, as.itgCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// CreateConfigTemplateSet create config template set
func (as *AccessServer) CreateConfigTemplateSet(ctx context.Context, req *pb.CreateConfigTemplateSetReq) (*pb.CreateConfigTemplateSetResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("CreateConfigTemplateSet[%d]| input[%+v]", req.Seq, req)
	response := &pb.CreateConfigTemplateSetResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("CreateConfigTemplateSet", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("CreateConfigTemplateSet[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authCheck(ctx, req.Bid); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := tplsetaction.NewCreateAction(as.viper, as.templateSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// DeleteConfigTemplateSet delete config template set
func (as *AccessServer) DeleteConfigTemplateSet(ctx context.Context, req *pb.DeleteConfigTemplateSetReq) (*pb.DeleteConfigTemplateSetResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("DeleteConfigTemplateSet[%d]| input[%+v]", req.Seq, req)
	response := &pb.DeleteConfigTemplateSetResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("DeleteConfigTemplateSet", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("DeleteConfigTemplateSet[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authCheck(ctx, req.Bid); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := tplsetaction.NewDeleteAction(as.viper, as.templateSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// UpdateConfigTemplateSet update config template set
func (as *AccessServer) UpdateConfigTemplateSet(ctx context.Context, req *pb.UpdateConfigTemplateSetReq) (*pb.UpdateConfigTemplateSetResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("UpdateConfigTemplateSet[%d]| input[%+v]", req.Seq, req)
	response := &pb.UpdateConfigTemplateSetResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("UpdateConfigTemplateSet", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("UpdateConfigTemplateSet[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authCheck(ctx, req.Bid); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := tplsetaction.NewUpdateAction(as.viper, as.templateSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// QueryConfigTemplateSet query config template set
func (as *AccessServer) QueryConfigTemplateSet(ctx context.Context, req *pb.QueryConfigTemplateSetReq) (*pb.QueryConfigTemplateSetResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryConfigTemplateSet[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryConfigTemplateSetResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("QueryConfigTemplateSet", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryConfigTemplateSet[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authCheck(ctx, req.Bid); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := tplsetaction.NewQueryAction(as.viper, as.templateSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// QueryConfigTemplateSetList delete config template set
func (as *AccessServer) QueryConfigTemplateSetList(ctx context.Context, req *pb.QueryConfigTemplateSetListReq) (*pb.QueryConfigTemplateSetListResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryConfigTemplateSetList[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryConfigTemplateSetListResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("QueryConfigTemplateSetList", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryConfigTemplateSetList[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authCheck(ctx, req.Bid); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := tplsetaction.NewListAction(as.viper, as.templateSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// CreateConfigTemplate create config template
func (as *AccessServer) CreateConfigTemplate(ctx context.Context, req *pb.CreateConfigTemplateReq) (*pb.CreateConfigTemplateResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("CreateConfigTemplate[%d]| input[%+v]", req.Seq, req)
	response := &pb.CreateConfigTemplateResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("CreateConfigTemplate", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("CreateConfigTemplate[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authCheck(ctx, req.Bid); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := tplaction.NewCreateAction(as.viper, as.templateSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// DeleteConfigTemplate create config template
func (as *AccessServer) DeleteConfigTemplate(ctx context.Context, req *pb.DeleteConfigTemplateReq) (*pb.DeleteConfigTemplateResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("DeleteConfigTemplate[%d]| input[%+v]", req.Seq, req)
	response := &pb.DeleteConfigTemplateResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("DeleteConfigTemplate", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("DeleteConfigTemplate[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authCheck(ctx, req.Bid); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := tplaction.NewDeleteAction(as.viper, as.templateSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// UpdateConfigTemplate update config template
func (as *AccessServer) UpdateConfigTemplate(ctx context.Context, req *pb.UpdateConfigTemplateReq) (*pb.UpdateConfigTemplateResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("UpdateConfigTemplate[%d]| input[%+v]", req.Seq, req)
	response := &pb.UpdateConfigTemplateResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("UpdateConfigTemplate", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("UpdateConfigTemplate[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authCheck(ctx, req.Bid); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := tplaction.NewUpdateAction(as.viper, as.templateSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// QueryConfigTemplate update config template
func (as *AccessServer) QueryConfigTemplate(ctx context.Context, req *pb.QueryConfigTemplateReq) (*pb.QueryConfigTemplateResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryConfigTemplate[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryConfigTemplateResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("QueryConfigTemplate", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryConfigTemplate[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authCheck(ctx, req.Bid); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := tplaction.NewQueryAction(as.viper, as.templateSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// QueryConfigTemplateList update config template
func (as *AccessServer) QueryConfigTemplateList(ctx context.Context, req *pb.QueryConfigTemplateListReq) (*pb.QueryConfigTemplateListResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryConfigTemplateList[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryConfigTemplateListResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("QueryConfigTemplateList", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryConfigTemplateList[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authCheck(ctx, req.Bid); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := tplaction.NewListAction(as.viper, as.templateSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// CreateTemplateVersion create config template version
func (as *AccessServer) CreateTemplateVersion(ctx context.Context, req *pb.CreateTemplateVersionReq) (*pb.CreateTemplateVersionResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("CreateTemplateVersion[%d]| input[%+v]", req.Seq, req)
	response := &pb.CreateTemplateVersionResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("CreateTemplateVersion", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("CreateTemplateVersion[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authCheck(ctx, req.Bid); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := tplversionaction.NewCreateAction(as.viper, as.templateSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// DeleteTemplateVersion create config template version
func (as *AccessServer) DeleteTemplateVersion(ctx context.Context, req *pb.DeleteTemplateVersionReq) (*pb.DeleteTemplateVersionResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("DeleteTemplateVersion[%d]| input[%+v]", req.Seq, req)
	response := &pb.DeleteTemplateVersionResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("DeleteTemplateVersion", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("DeleteTemplateVersion[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authCheck(ctx, req.Bid); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := tplversionaction.NewDeleteAction(as.viper, as.templateSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// UpdateTemplateVersion create config template version
func (as *AccessServer) UpdateTemplateVersion(ctx context.Context, req *pb.UpdateTemplateVersionReq) (*pb.UpdateTemplateVersionResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("UpdateTemplateVersion[%d]| input[%+v]", req.Seq, req)
	response := &pb.UpdateTemplateVersionResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("UpdateTemplateVersion", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("UpdateTemplateVersion[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authCheck(ctx, req.Bid); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := tplversionaction.NewUpdateAction(as.viper, as.templateSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// QueryTemplateVersion query config template version
func (as *AccessServer) QueryTemplateVersion(ctx context.Context, req *pb.QueryTemplateVersionReq) (*pb.QueryTemplateVersionResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryTemplateVersion[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryTemplateVersionResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("QueryTemplateVersion", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryTemplateVersion[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authCheck(ctx, req.Bid); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := tplversionaction.NewQueryAction(as.viper, as.templateSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// QueryTemplateVersionList list config template version
func (as *AccessServer) QueryTemplateVersionList(ctx context.Context, req *pb.QueryTemplateVersionListReq) (*pb.QueryTemplateVersionListResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryTemplateVersionList[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryTemplateVersionListResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("QueryTemplateVersionList", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryTemplateVersionList[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authCheck(ctx, req.Bid); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := tplversionaction.NewListAction(as.viper, as.templateSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// CreateConfigTemplateBinding create config template binding
func (as *AccessServer) CreateConfigTemplateBinding(ctx context.Context, req *pb.CreateConfigTemplateBindingReq) (*pb.CreateConfigTemplateBindingResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("CreateConfigTemplateBinding[%d]| input[%+v]", req.Seq, req)
	response := &pb.CreateConfigTemplateBindingResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("CreateConfigTemplateBinding", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("CreateConfigTemplateBinding[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authCheck(ctx, req.Bid); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := tplbindingaction.NewCreateAction(as.viper, as.templateSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// DeleteConfigTemplateBinding delete config template binding
func (as *AccessServer) DeleteConfigTemplateBinding(ctx context.Context, req *pb.DeleteConfigTemplateBindingReq) (*pb.DeleteConfigTemplateBindingResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("DeleteConfigTemplateBinding[%d]| input[%+v]", req.Seq, req)
	response := &pb.DeleteConfigTemplateBindingResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("DeleteConfigTemplateBinding", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("DeleteConfigTemplateBinding[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authCheck(ctx, req.Bid); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := tplbindingaction.NewDeleteAction(as.viper, as.templateSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// SyncConfigTemplateBinding delete config template binding
func (as *AccessServer) SyncConfigTemplateBinding(ctx context.Context, req *pb.SyncConfigTemplateBindingReq) (*pb.SyncConfigTemplateBindingResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("SyncConfigTemplateBinding[%d]| input[%+v]", req.Seq, req)
	response := &pb.SyncConfigTemplateBindingResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("SyncConfigTemplateBinding", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("SyncConfigTemplateBinding[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authCheck(ctx, req.Bid); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := tplbindingaction.NewUpdateAction(as.viper, as.templateSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// QueryConfigTemplateBinding query config template binding
func (as *AccessServer) QueryConfigTemplateBinding(ctx context.Context, req *pb.QueryConfigTemplateBindingReq) (*pb.QueryConfigTemplateBindingResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryConfigTemplateBinding[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryConfigTemplateBindingResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("QueryConfigTemplateBinding", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryConfigTemplateBinding[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authCheck(ctx, req.Bid); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := tplbindingaction.NewQueryAction(as.viper, as.templateSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// QueryConfigTemplateBindingList query config template binding list
func (as *AccessServer) QueryConfigTemplateBindingList(ctx context.Context, req *pb.QueryConfigTemplateBindingListReq) (*pb.QueryConfigTemplateBindingListResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryConfigTemplateBindingList[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryConfigTemplateBindingListResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("QueryConfigTemplateBindingList", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryConfigTemplateBindingList[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authCheck(ctx, req.Bid); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := tplbindingaction.NewListAction(as.viper, as.templateSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// CreateVariable create variable
func (as *AccessServer) CreateVariable(ctx context.Context, req *pb.CreateVariableReq) (*pb.CreateVariableResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("CreateVariable[%d]| input[%+v]", req.Seq, req)
	response := &pb.CreateVariableResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("CreateVariable", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("CreateVariable[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authCheck(ctx, req.Bid); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := variableaction.NewCreateAction(as.viper, as.templateSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// DeleteVariable delete variable
func (as *AccessServer) DeleteVariable(ctx context.Context, req *pb.DeleteVariableReq) (*pb.DeleteVariableResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("DeleteVariable[%d]| input[%+v]", req.Seq, req)
	response := &pb.DeleteVariableResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("DeleteVariable", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("DeleteVariable[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authCheck(ctx, req.Bid); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := variableaction.NewDeleteAction(as.viper, as.templateSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// UpdateVariable update variable
func (as *AccessServer) UpdateVariable(ctx context.Context, req *pb.UpdateVariableReq) (*pb.UpdateVariableResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("UpdateVariable[%d]| input[%+v]", req.Seq, req)
	response := &pb.UpdateVariableResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("UpdateVariable", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("UpdateVariable[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authCheck(ctx, req.Bid); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := variableaction.NewUpdateAction(as.viper, as.templateSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// QueryVariable query variable
func (as *AccessServer) QueryVariable(ctx context.Context, req *pb.QueryVariableReq) (*pb.QueryVariableResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryVariable[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryVariableResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("QueryVariable", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryVariable[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authCheck(ctx, req.Bid); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := variableaction.NewQueryAction(as.viper, as.templateSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// QueryVariableList query variable
func (as *AccessServer) QueryVariableList(ctx context.Context, req *pb.QueryVariableListReq) (*pb.QueryVariableListResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("QueryVariableList[%d]| input[%+v]", req.Seq, req)
	response := &pb.QueryVariableListResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("QueryVariableList", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("QueryVariableList[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authCheck(ctx, req.Bid); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := variableaction.NewListAction(as.viper, as.templateSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}

// Reload reloads target release or multi release.
func (as *AccessServer) Reload(ctx context.Context, req *pb.ReloadReq) (*pb.ReloadResp, error) {
	rtime := time.Now()
	logger.V(2).Infof("Reload[%d]| input[%+v]", req.Seq, req)
	response := &pb.ReloadResp{Seq: req.Seq, ErrCode: pbcommon.ErrCode_E_OK, ErrMsg: "OK"}

	defer func() {
		cost := as.collector.StatRequest("Reload", response.ErrCode, rtime, time.Now())
		logger.V(2).Infof("Reload[%d]| output[%dms][%+v]", req.Seq, cost, response)
	}()

	if errCode, errMsg := as.authCheck(ctx, req.Bid); errCode != pbcommon.ErrCode_E_OK {
		response.ErrCode = errCode
		response.ErrMsg = errMsg
		return response, nil
	}

	action := reloadaction.NewReloadAction(as.viper, as.businessSvrCli, req, response)
	as.executor.Execute(action)

	return response, nil
}
