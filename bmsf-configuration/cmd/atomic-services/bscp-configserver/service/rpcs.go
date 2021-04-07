/*
Tencent is pleased to support the open source community by making Blueking Container Service available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "as IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package service

import (
	"context"
	"time"

	appinstanceaction "bk-bscp/cmd/atomic-services/bscp-configserver/actions/appinstance"
	appaction "bk-bscp/cmd/atomic-services/bscp-configserver/actions/application"
	auditaction "bk-bscp/cmd/atomic-services/bscp-configserver/actions/audit"
	commitaction "bk-bscp/cmd/atomic-services/bscp-configserver/actions/commit"
	configaction "bk-bscp/cmd/atomic-services/bscp-configserver/actions/config"
	contentaction "bk-bscp/cmd/atomic-services/bscp-configserver/actions/content"
	healthzaction "bk-bscp/cmd/atomic-services/bscp-configserver/actions/healthz"
	multicommitaction "bk-bscp/cmd/atomic-services/bscp-configserver/actions/multi-commit"
	multireleaseaction "bk-bscp/cmd/atomic-services/bscp-configserver/actions/multi-release"
	procattraction "bk-bscp/cmd/atomic-services/bscp-configserver/actions/procattr"
	releaseaction "bk-bscp/cmd/atomic-services/bscp-configserver/actions/release"
	strategyaction "bk-bscp/cmd/atomic-services/bscp-configserver/actions/strategy"
	pb "bk-bscp/internal/protocol/configserver"
	"bk-bscp/pkg/common"
	"bk-bscp/pkg/logger"
)

// CreateApp creates new app.
func (cs *ConfigServer) CreateApp(ctx context.Context, req *pb.CreateAppReq) (*pb.CreateAppResp, error) {
	rtime := time.Now()
	kit := common.RequestKit(ctx)
	logger.V(2).Infof("%s[%s]| appcode: %s, user: %s, input[%+v]", kit.Method, kit.Rid, kit.AppCode, kit.User, req)

	response := new(pb.CreateAppResp)

	defer func() {
		cost := cs.collector.StatRequest(kit.Method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", kit.Method, kit.Rid, cost, response)
	}()

	action := appaction.NewCreateAction(kit, cs.viper, cs.dataMgrCli, req, response)
	if err := cs.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", kit.Method, kit.Rid, err)
	}

	return response, nil
}

// QueryApp returns target app.
func (cs *ConfigServer) QueryApp(ctx context.Context, req *pb.QueryAppReq) (*pb.QueryAppResp, error) {
	rtime := time.Now()
	kit := common.RequestKit(ctx)
	logger.V(2).Infof("%s[%s]| appcode: %s, user: %s, input[%+v]", kit.Method, kit.Rid, kit.AppCode, kit.User, req)

	response := new(pb.QueryAppResp)

	defer func() {
		cost := cs.collector.StatRequest(kit.Method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", kit.Method, kit.Rid, cost, response)
	}()

	action := appaction.NewQueryAction(kit, cs.viper, cs.dataMgrCli, req, response)
	if err := cs.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", kit.Method, kit.Rid, err)
	}

	return response, nil
}

// QueryAppList returns all apps.
func (cs *ConfigServer) QueryAppList(ctx context.Context, req *pb.QueryAppListReq) (*pb.QueryAppListResp, error) {
	rtime := time.Now()
	kit := common.RequestKit(ctx)
	logger.V(2).Infof("%s[%s]| appcode: %s, user: %s, input[%+v]", kit.Method, kit.Rid, kit.AppCode, kit.User, req)

	response := new(pb.QueryAppListResp)

	defer func() {
		cost := cs.collector.StatRequest(kit.Method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", kit.Method, kit.Rid, cost, response)
	}()

	action := appaction.NewListAction(kit, cs.viper, cs.dataMgrCli, req, response)
	if err := cs.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", kit.Method, kit.Rid, err)
	}

	return response, nil
}

// UpdateApp updates target app.
func (cs *ConfigServer) UpdateApp(ctx context.Context, req *pb.UpdateAppReq) (*pb.UpdateAppResp, error) {
	rtime := time.Now()
	kit := common.RequestKit(ctx)
	logger.V(2).Infof("%s[%s]| appcode: %s, user: %s, input[%+v]", kit.Method, kit.Rid, kit.AppCode, kit.User, req)

	response := new(pb.UpdateAppResp)

	defer func() {
		cost := cs.collector.StatRequest(kit.Method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", kit.Method, kit.Rid, cost, response)
	}()

	action := appaction.NewUpdateAction(kit, cs.viper, cs.authSvrCli, cs.dataMgrCli, req, response)
	if err := cs.executor.ExecuteWithAuth(action); err != nil {
		logger.Errorf("%s[%s]| %+v", kit.Method, kit.Rid, err)
	}

	return response, nil
}

// DeleteApp deletes target app.
func (cs *ConfigServer) DeleteApp(ctx context.Context, req *pb.DeleteAppReq) (*pb.DeleteAppResp, error) {
	rtime := time.Now()
	kit := common.RequestKit(ctx)
	logger.V(2).Infof("%s[%s]| appcode: %s, user: %s, input[%+v]", kit.Method, kit.Rid, kit.AppCode, kit.User, req)

	response := new(pb.DeleteAppResp)

	defer func() {
		cost := cs.collector.StatRequest(kit.Method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", kit.Method, kit.Rid, cost, response)
	}()

	action := appaction.NewDeleteAction(kit, cs.viper, cs.authSvrCli, cs.dataMgrCli, req, response)
	if err := cs.executor.ExecuteWithAuth(action); err != nil {
		logger.Errorf("%s[%s]| %+v", kit.Method, kit.Rid, err)
	}

	return response, nil
}

// CreateConfig creates new config.
func (cs *ConfigServer) CreateConfig(ctx context.Context, req *pb.CreateConfigReq) (*pb.CreateConfigResp, error) {
	rtime := time.Now()
	kit := common.RequestKit(ctx)
	logger.V(2).Infof("%s[%s]| appcode: %s, user: %s, input[%+v]", kit.Method, kit.Rid, kit.AppCode, kit.User, req)

	response := new(pb.CreateConfigResp)

	defer func() {
		cost := cs.collector.StatRequest(kit.Method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", kit.Method, kit.Rid, cost, response)
	}()

	action := configaction.NewCreateAction(kit, cs.viper, cs.authSvrCli, cs.dataMgrCli, req, response)
	if err := cs.executor.ExecuteWithAuth(action); err != nil {
		logger.Errorf("%s[%s]| %+v", kit.Method, kit.Rid, err)
	}

	return response, nil
}

// QueryConfig returns target config.
func (cs *ConfigServer) QueryConfig(ctx context.Context, req *pb.QueryConfigReq) (*pb.QueryConfigResp, error) {
	rtime := time.Now()
	kit := common.RequestKit(ctx)
	logger.V(2).Infof("%s[%s]| appcode: %s, user: %s, input[%+v]", kit.Method, kit.Rid, kit.AppCode, kit.User, req)

	response := new(pb.QueryConfigResp)

	defer func() {
		cost := cs.collector.StatRequest(kit.Method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", kit.Method, kit.Rid, cost, response)
	}()

	action := configaction.NewQueryAction(kit, cs.viper, cs.dataMgrCli, req, response)
	if err := cs.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", kit.Method, kit.Rid, err)
	}

	return response, nil
}

// QueryConfigList returns all configs.
func (cs *ConfigServer) QueryConfigList(ctx context.Context,
	req *pb.QueryConfigListReq) (*pb.QueryConfigListResp, error) {

	rtime := time.Now()
	kit := common.RequestKit(ctx)
	logger.V(2).Infof("%s[%s]| appcode: %s, user: %s, input[%+v]", kit.Method, kit.Rid, kit.AppCode, kit.User, req)

	response := new(pb.QueryConfigListResp)

	defer func() {
		cost := cs.collector.StatRequest(kit.Method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", kit.Method, kit.Rid, cost, response)
	}()

	action := configaction.NewListAction(kit, cs.viper, cs.dataMgrCli, req, response)
	if err := cs.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", kit.Method, kit.Rid, err)
	}

	return response, nil
}

// UpdateConfig updates target config.
func (cs *ConfigServer) UpdateConfig(ctx context.Context, req *pb.UpdateConfigReq) (*pb.UpdateConfigResp, error) {
	rtime := time.Now()
	kit := common.RequestKit(ctx)
	logger.V(2).Infof("%s[%s]| appcode: %s, user: %s, input[%+v]", kit.Method, kit.Rid, kit.AppCode, kit.User, req)

	response := new(pb.UpdateConfigResp)

	defer func() {
		cost := cs.collector.StatRequest(kit.Method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", kit.Method, kit.Rid, cost, response)
	}()

	action := configaction.NewUpdateAction(kit, cs.viper, cs.authSvrCli, cs.dataMgrCli, req, response)
	if err := cs.executor.ExecuteWithAuth(action); err != nil {
		logger.Errorf("%s[%s]| %+v", kit.Method, kit.Rid, err)
	}

	return response, nil
}

// DeleteConfig deletes target config.
func (cs *ConfigServer) DeleteConfig(ctx context.Context, req *pb.DeleteConfigReq) (*pb.DeleteConfigResp, error) {
	rtime := time.Now()
	kit := common.RequestKit(ctx)
	logger.V(2).Infof("%s[%s]| appcode: %s, user: %s, input[%+v]", kit.Method, kit.Rid, kit.AppCode, kit.User, req)

	response := new(pb.DeleteConfigResp)

	defer func() {
		cost := cs.collector.StatRequest(kit.Method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", kit.Method, kit.Rid, cost, response)
	}()

	action := configaction.NewDeleteAction(kit, cs.viper, cs.authSvrCli, cs.dataMgrCli, req, response)
	if err := cs.executor.ExecuteWithAuth(action); err != nil {
		logger.Errorf("%s[%s]| %+v", kit.Method, kit.Rid, err)
	}

	return response, nil
}

// CreateConfigContent creates config content base on cid and index info.
func (cs *ConfigServer) CreateConfigContent(ctx context.Context,
	req *pb.CreateConfigContentReq) (*pb.CreateConfigContentResp, error) {

	rtime := time.Now()
	kit := common.RequestKit(ctx)
	logger.V(2).Infof("%s[%s]| appcode: %s, user: %s, input[%+v]", kit.Method, kit.Rid, kit.AppCode, kit.User, req)

	response := new(pb.CreateConfigContentResp)

	defer func() {
		cost := cs.collector.StatRequest(kit.Method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s`[%s]| output[%dms][%+v]", kit.Method, kit.Rid, cost, response)
	}()

	action := contentaction.NewCreateAction(kit, cs.viper, cs.authSvrCli, cs.dataMgrCli, req, response)
	if err := cs.executor.ExecuteWithAuth(action); err != nil {
		logger.Errorf("%s[%s]| %+v", kit.Method, kit.Rid, err)
	}

	return response, nil
}

// QueryConfigContent returns target config content base on labels.
func (cs *ConfigServer) QueryConfigContent(ctx context.Context,
	req *pb.QueryConfigContentReq) (*pb.QueryConfigContentResp, error) {

	rtime := time.Now()
	kit := common.RequestKit(ctx)
	logger.V(2).Infof("%s[%s]| appcode: %s, user: %s, input[%+v]", kit.Method, kit.Rid, kit.AppCode, kit.User, req)

	response := new(pb.QueryConfigContentResp)

	defer func() {
		cost := cs.collector.StatRequest(kit.Method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s`[%s]| output[%dms][%+v]", kit.Method, kit.Rid, cost, response)
	}()

	action := contentaction.NewQueryAction(kit, cs.viper, cs.dataMgrCli, req, response)
	if err := cs.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", kit.Method, kit.Rid, err)
	}

	return response, nil
}

// QueryConfigContentList returns all config contents.
func (cs *ConfigServer) QueryConfigContentList(ctx context.Context,
	req *pb.QueryConfigContentListReq) (*pb.QueryConfigContentListResp, error) {

	rtime := time.Now()
	kit := common.RequestKit(ctx)
	logger.V(2).Infof("%s[%s]| appcode: %s, user: %s, input[%+v]", kit.Method, kit.Rid, kit.AppCode, kit.User, req)

	response := new(pb.QueryConfigContentListResp)

	defer func() {
		cost := cs.collector.StatRequest(kit.Method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s`[%s]| output[%dms][%+v]", kit.Method, kit.Rid, cost, response)
	}()

	action := contentaction.NewListAction(kit, cs.viper, cs.dataMgrCli, req, response)
	if err := cs.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", kit.Method, kit.Rid, err)
	}

	return response, nil
}

// QueryReleaseConfigContent returns config contents of target release.
func (cs *ConfigServer) QueryReleaseConfigContent(ctx context.Context,
	req *pb.QueryReleaseConfigContentReq) (*pb.QueryReleaseConfigContentResp, error) {

	rtime := time.Now()
	kit := common.RequestKit(ctx)
	logger.V(2).Infof("%s[%s]| appcode: %s, user: %s, input[%+v]", kit.Method, kit.Rid, kit.AppCode, kit.User, req)

	response := new(pb.QueryReleaseConfigContentResp)

	defer func() {
		cost := cs.collector.StatRequest(kit.Method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", kit.Method, kit.Rid, cost, response)
	}()

	action := contentaction.NewReleaseAction(kit, cs.viper, cs.dataMgrCli, req, response)
	if err := cs.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", kit.Method, kit.Rid, err)
	}

	return response, nil
}

// CreateCommit creates new commit.
func (cs *ConfigServer) CreateCommit(ctx context.Context, req *pb.CreateCommitReq) (*pb.CreateCommitResp, error) {
	rtime := time.Now()
	kit := common.RequestKit(ctx)
	logger.V(2).Infof("%s[%s]| appcode: %s, user: %s, input[%+v]", kit.Method, kit.Rid, kit.AppCode, kit.User, req)

	response := new(pb.CreateCommitResp)

	defer func() {
		cost := cs.collector.StatRequest(kit.Method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", kit.Method, kit.Rid, cost, response)
	}()

	action := commitaction.NewCreateAction(kit, cs.viper, cs.authSvrCli, cs.dataMgrCli, req, response)
	if err := cs.executor.ExecuteWithAuth(action); err != nil {
		logger.Errorf("%s[%s]| %+v", kit.Method, kit.Rid, err)
	}

	return response, nil
}

// QueryCommit returns target commit.
func (cs *ConfigServer) QueryCommit(ctx context.Context, req *pb.QueryCommitReq) (*pb.QueryCommitResp, error) {
	rtime := time.Now()
	kit := common.RequestKit(ctx)
	logger.V(2).Infof("%s[%s]| appcode: %s, user: %s, input[%+v]", kit.Method, kit.Rid, kit.AppCode, kit.User, req)

	response := new(pb.QueryCommitResp)

	defer func() {
		cost := cs.collector.StatRequest(kit.Method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", kit.Method, kit.Rid, cost, response)
	}()

	action := commitaction.NewQueryAction(kit, cs.viper, cs.dataMgrCli, req, response)
	if err := cs.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", kit.Method, kit.Rid, err)
	}

	return response, nil
}

// QueryHistoryCommits returns all history commits.
func (cs *ConfigServer) QueryHistoryCommits(ctx context.Context,
	req *pb.QueryHistoryCommitsReq) (*pb.QueryHistoryCommitsResp, error) {

	rtime := time.Now()
	kit := common.RequestKit(ctx)
	logger.V(2).Infof("%s[%s]| appcode: %s, user: %s, input[%+v]", kit.Method, kit.Rid, kit.AppCode, kit.User, req)

	response := new(pb.QueryHistoryCommitsResp)

	defer func() {
		cost := cs.collector.StatRequest(kit.Method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", kit.Method, kit.Rid, cost, response)
	}()

	action := commitaction.NewListAction(kit, cs.viper, cs.dataMgrCli, req, response)
	if err := cs.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", kit.Method, kit.Rid, err)
	}

	return response, nil
}

// UpdateCommit updates target commit.
func (cs *ConfigServer) UpdateCommit(ctx context.Context, req *pb.UpdateCommitReq) (*pb.UpdateCommitResp, error) {
	rtime := time.Now()
	kit := common.RequestKit(ctx)
	logger.V(2).Infof("%s[%s]| appcode: %s, user: %s, input[%+v]", kit.Method, kit.Rid, kit.AppCode, kit.User, req)

	response := new(pb.UpdateCommitResp)

	defer func() {
		cost := cs.collector.StatRequest(kit.Method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", kit.Method, kit.Rid, cost, response)
	}()

	action := commitaction.NewUpdateAction(kit, cs.viper, cs.authSvrCli, cs.dataMgrCli, req, response)
	if err := cs.executor.ExecuteWithAuth(action); err != nil {
		logger.Errorf("%s[%s]| %+v", kit.Method, kit.Rid, err)
	}

	return response, nil
}

// CancelCommit cancels target commit.
func (cs *ConfigServer) CancelCommit(ctx context.Context, req *pb.CancelCommitReq) (*pb.CancelCommitResp, error) {
	rtime := time.Now()
	kit := common.RequestKit(ctx)
	logger.V(2).Infof("%s[%s]| appcode: %s, user: %s, input[%+v]", kit.Method, kit.Rid, kit.AppCode, kit.User, req)

	response := new(pb.CancelCommitResp)

	defer func() {
		cost := cs.collector.StatRequest(kit.Method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", kit.Method, kit.Rid, cost, response)
	}()

	action := commitaction.NewCancelAction(kit, cs.viper, cs.authSvrCli, cs.dataMgrCli, req, response)
	if err := cs.executor.ExecuteWithAuth(action); err != nil {
		logger.Errorf("%s[%s]| %+v", kit.Method, kit.Rid, err)
	}

	return response, nil
}

// ConfirmCommit confirms target commit.
func (cs *ConfigServer) ConfirmCommit(ctx context.Context, req *pb.ConfirmCommitReq) (*pb.ConfirmCommitResp, error) {
	rtime := time.Now()
	kit := common.RequestKit(ctx)
	logger.V(2).Infof("%s[%s]| appcode: %s, user: %s, input[%+v]", kit.Method, kit.Rid, kit.AppCode, kit.User, req)

	response := new(pb.ConfirmCommitResp)

	defer func() {
		cost := cs.collector.StatRequest(kit.Method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", kit.Method, kit.Rid, cost, response)
	}()

	action := commitaction.NewConfirmAction(kit, cs.viper, cs.authSvrCli, cs.dataMgrCli, req, response)
	if err := cs.executor.ExecuteWithAuth(action); err != nil {
		logger.Errorf("%s[%s]| %+v", kit.Method, kit.Rid, err)
	}

	return response, nil
}

// CreateMultiCommitWithContent creates new multi commit with contents.
func (cs *ConfigServer) CreateMultiCommitWithContent(ctx context.Context,
	req *pb.CreateMultiCommitWithContentReq) (*pb.CreateMultiCommitWithContentResp, error) {

	rtime := time.Now()
	kit := common.RequestKit(ctx)
	logger.V(2).Infof("%s[%s]| appcode: %s, user: %s, input[%+v]", kit.Method, kit.Rid, kit.AppCode, kit.User, req)

	response := new(pb.CreateMultiCommitWithContentResp)

	defer func() {
		cost := cs.collector.StatRequest(kit.Method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", kit.Method, kit.Rid, cost, response)
	}()

	action := multicommitaction.NewCreateWithContentAction(kit, cs.viper, cs.authSvrCli, cs.dataMgrCli, req, response)
	if err := cs.executor.ExecuteWithAuth(action); err != nil {
		logger.Errorf("%s[%s]| %+v", kit.Method, kit.Rid, err)
	}

	return response, nil
}

// CreateMultiCommit creates new multi commit.
func (cs *ConfigServer) CreateMultiCommit(ctx context.Context,
	req *pb.CreateMultiCommitReq) (*pb.CreateMultiCommitResp, error) {

	rtime := time.Now()
	kit := common.RequestKit(ctx)
	logger.V(2).Infof("%s[%s]| appcode: %s, user: %s, input[%+v]", kit.Method, kit.Rid, kit.AppCode, kit.User, req)

	response := new(pb.CreateMultiCommitResp)

	defer func() {
		cost := cs.collector.StatRequest(kit.Method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", kit.Method, kit.Rid, cost, response)
	}()

	action := multicommitaction.NewCreateAction(kit, cs.viper, cs.authSvrCli, cs.dataMgrCli, req, response)
	if err := cs.executor.ExecuteWithAuth(action); err != nil {
		logger.Errorf("%s[%s]| %+v", kit.Method, kit.Rid, err)
	}

	return response, nil
}

// QueryMultiCommit returns target multi commit.
func (cs *ConfigServer) QueryMultiCommit(ctx context.Context,
	req *pb.QueryMultiCommitReq) (*pb.QueryMultiCommitResp, error) {

	rtime := time.Now()
	kit := common.RequestKit(ctx)
	logger.V(2).Infof("%s[%s]| appcode: %s, user: %s, input[%+v]", kit.Method, kit.Rid, kit.AppCode, kit.User, req)

	response := new(pb.QueryMultiCommitResp)

	defer func() {
		cost := cs.collector.StatRequest(kit.Method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", kit.Method, kit.Rid, cost, response)
	}()

	action := multicommitaction.NewQueryAction(kit, cs.viper, cs.dataMgrCli, req, response)
	if err := cs.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", kit.Method, kit.Rid, err)
	}

	return response, nil
}

// QueryHistoryMultiCommits returns all history multi commits.
func (cs *ConfigServer) QueryHistoryMultiCommits(ctx context.Context,
	req *pb.QueryHistoryMultiCommitsReq) (*pb.QueryHistoryMultiCommitsResp, error) {

	rtime := time.Now()
	kit := common.RequestKit(ctx)
	logger.V(2).Infof("%s[%s]| appcode: %s, user: %s, input[%+v]", kit.Method, kit.Rid, kit.AppCode, kit.User, req)

	response := new(pb.QueryHistoryMultiCommitsResp)

	defer func() {
		cost := cs.collector.StatRequest(kit.Method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", kit.Method, kit.Rid, cost, response)
	}()

	action := multicommitaction.NewListAction(kit, cs.viper, cs.dataMgrCli, req, response)
	if err := cs.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", kit.Method, kit.Rid, err)
	}

	return response, nil
}

// UpdateMultiCommit updates target multi commit.
func (cs *ConfigServer) UpdateMultiCommit(ctx context.Context,
	req *pb.UpdateMultiCommitReq) (*pb.UpdateMultiCommitResp, error) {

	rtime := time.Now()
	kit := common.RequestKit(ctx)
	logger.V(2).Infof("%s[%s]| appcode: %s, user: %s, input[%+v]", kit.Method, kit.Rid, kit.AppCode, kit.User, req)

	response := new(pb.UpdateMultiCommitResp)

	defer func() {
		cost := cs.collector.StatRequest(kit.Method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", kit.Method, kit.Rid, cost, response)
	}()

	action := multicommitaction.NewUpdateAction(kit, cs.viper, cs.authSvrCli, cs.dataMgrCli, req, response)
	if err := cs.executor.ExecuteWithAuth(action); err != nil {
		logger.Errorf("%s[%s]| %+v", kit.Method, kit.Rid, err)
	}

	return response, nil
}

// CancelMultiCommit cancels target multi commit.
func (cs *ConfigServer) CancelMultiCommit(ctx context.Context,
	req *pb.CancelMultiCommitReq) (*pb.CancelMultiCommitResp, error) {

	rtime := time.Now()
	kit := common.RequestKit(ctx)
	logger.V(2).Infof("%s[%s]| appcode: %s, user: %s, input[%+v]", kit.Method, kit.Rid, kit.AppCode, kit.User, req)

	response := new(pb.CancelMultiCommitResp)

	defer func() {
		cost := cs.collector.StatRequest(kit.Method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", kit.Method, kit.Rid, cost, response)
	}()

	action := multicommitaction.NewCancelAction(kit, cs.viper, cs.authSvrCli, cs.dataMgrCli, req, response)
	if err := cs.executor.ExecuteWithAuth(action); err != nil {
		logger.Errorf("%s[%s]| %+v", kit.Method, kit.Rid, err)
	}

	return response, nil
}

// ConfirmMultiCommit confirms target multi commit.
func (cs *ConfigServer) ConfirmMultiCommit(ctx context.Context,
	req *pb.ConfirmMultiCommitReq) (*pb.ConfirmMultiCommitResp, error) {

	rtime := time.Now()
	kit := common.RequestKit(ctx)
	logger.V(2).Infof("%s[%s]| appcode: %s, user: %s, input[%+v]", kit.Method, kit.Rid, kit.AppCode, kit.User, req)

	response := new(pb.ConfirmMultiCommitResp)

	defer func() {
		cost := cs.collector.StatRequest(kit.Method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", kit.Method, kit.Rid, cost, response)
	}()

	action := multicommitaction.NewConfirmAction(kit, cs.viper, cs.authSvrCli, cs.dataMgrCli, req, response)
	if err := cs.executor.ExecuteWithAuth(action); err != nil {
		logger.Errorf("%s[%s]| %+v", kit.Method, kit.Rid, err)
	}

	return response, nil
}

// CreateRelease creates new release.
func (cs *ConfigServer) CreateRelease(ctx context.Context, req *pb.CreateReleaseReq) (*pb.CreateReleaseResp, error) {
	rtime := time.Now()
	kit := common.RequestKit(ctx)
	logger.V(2).Infof("%s[%s]| appcode: %s, user: %s, input[%+v]", kit.Method, kit.Rid, kit.AppCode, kit.User, req)

	response := new(pb.CreateReleaseResp)

	defer func() {
		cost := cs.collector.StatRequest(kit.Method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", kit.Method, kit.Rid, cost, response)
	}()

	action := releaseaction.NewCreateAction(kit, cs.viper, cs.authSvrCli, cs.dataMgrCli, req, response)
	if err := cs.executor.ExecuteWithAuth(action); err != nil {
		logger.Errorf("%s[%s]| %+v", kit.Method, kit.Rid, err)
	}

	return response, nil
}

// QueryRelease returns target release.
func (cs *ConfigServer) QueryRelease(ctx context.Context, req *pb.QueryReleaseReq) (*pb.QueryReleaseResp, error) {
	rtime := time.Now()
	kit := common.RequestKit(ctx)
	logger.V(2).Infof("%s[%s]| appcode: %s, user: %s, input[%+v]", kit.Method, kit.Rid, kit.AppCode, kit.User, req)

	response := new(pb.QueryReleaseResp)

	defer func() {
		cost := cs.collector.StatRequest(kit.Method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", kit.Method, kit.Rid, cost, response)
	}()

	action := releaseaction.NewQueryAction(kit, cs.viper, cs.dataMgrCli, req, response)
	if err := cs.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", kit.Method, kit.Rid, err)
	}

	return response, nil
}

// QueryHistoryReleases returns history releases.
func (cs *ConfigServer) QueryHistoryReleases(ctx context.Context,
	req *pb.QueryHistoryReleasesReq) (*pb.QueryHistoryReleasesResp, error) {

	rtime := time.Now()
	kit := common.RequestKit(ctx)
	logger.V(2).Infof("%s[%s]| appcode: %s, user: %s, input[%+v]", kit.Method, kit.Rid, kit.AppCode, kit.User, req)

	response := new(pb.QueryHistoryReleasesResp)

	defer func() {
		cost := cs.collector.StatRequest(kit.Method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", kit.Method, kit.Rid, cost, response)
	}()

	action := releaseaction.NewListAction(kit, cs.viper, cs.dataMgrCli, req, response)
	if err := cs.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", kit.Method, kit.Rid, err)
	}

	return response, nil
}

// UpdateRelease updates target release.
func (cs *ConfigServer) UpdateRelease(ctx context.Context, req *pb.UpdateReleaseReq) (*pb.UpdateReleaseResp, error) {
	rtime := time.Now()
	kit := common.RequestKit(ctx)
	logger.V(2).Infof("%s[%s]| appcode: %s, user: %s, input[%+v]", kit.Method, kit.Rid, kit.AppCode, kit.User, req)

	response := new(pb.UpdateReleaseResp)

	defer func() {
		cost := cs.collector.StatRequest(kit.Method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", kit.Method, kit.Rid, cost, response)
	}()

	action := releaseaction.NewUpdateAction(kit, cs.viper, cs.authSvrCli, cs.dataMgrCli, req, response)
	if err := cs.executor.ExecuteWithAuth(action); err != nil {
		logger.Errorf("%s[%s]| %+v", kit.Method, kit.Rid, err)
	}

	return response, nil
}

// CancelRelease cancels target release.
func (cs *ConfigServer) CancelRelease(ctx context.Context, req *pb.CancelReleaseReq) (*pb.CancelReleaseResp, error) {
	rtime := time.Now()
	kit := common.RequestKit(ctx)
	logger.V(2).Infof("%s[%s]| appcode: %s, user: %s, input[%+v]", kit.Method, kit.Rid, kit.AppCode, kit.User, req)

	response := new(pb.CancelReleaseResp)

	defer func() {
		cost := cs.collector.StatRequest(kit.Method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", kit.Method, kit.Rid, cost, response)
	}()

	action := releaseaction.NewCancelAction(kit, cs.viper, cs.authSvrCli, cs.dataMgrCli, req, response)
	if err := cs.executor.ExecuteWithAuth(action); err != nil {
		logger.Errorf("%s[%s]| %+v", kit.Method, kit.Rid, err)
	}

	return response, nil
}

// PublishRelease publishes target release.
func (cs *ConfigServer) PublishRelease(ctx context.Context,
	req *pb.PublishReleaseReq) (*pb.PublishReleaseResp, error) {

	rtime := time.Now()
	kit := common.RequestKit(ctx)
	logger.V(2).Infof("%s[%s]| appcode: %s, user: %s, input[%+v]", kit.Method, kit.Rid, kit.AppCode, kit.User, req)

	response := new(pb.PublishReleaseResp)

	defer func() {
		cost := cs.collector.StatRequest(kit.Method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", kit.Method, kit.Rid, cost, response)
	}()

	action := releaseaction.NewPublishAction(kit, cs.viper, cs.authSvrCli, cs.dataMgrCli,
		cs.gseControllerCli, req, response)

	if err := cs.executor.ExecuteWithAuth(action); err != nil {
		logger.Errorf("%s[%s]| %+v", kit.Method, kit.Rid, err)
	}

	return response, nil
}

// RollbackRelease rollbacks target release.
func (cs *ConfigServer) RollbackRelease(ctx context.Context,
	req *pb.RollbackReleaseReq) (*pb.RollbackReleaseResp, error) {

	rtime := time.Now()
	kit := common.RequestKit(ctx)
	logger.V(2).Infof("%s[%s]| appcode: %s, user: %s, input[%+v]", kit.Method, kit.Rid, kit.AppCode, kit.User, req)

	response := new(pb.RollbackReleaseResp)

	defer func() {
		cost := cs.collector.StatRequest(kit.Method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", kit.Method, kit.Rid, cost, response)
	}()

	action := releaseaction.NewRollbackAction(kit, cs.viper, cs.authSvrCli, cs.dataMgrCli,
		cs.gseControllerCli, req, response)

	if err := cs.executor.ExecuteWithAuth(action); err != nil {
		logger.Errorf("%s[%s]| %+v", kit.Method, kit.Rid, err)
	}

	return response, nil
}

// CreateMultiRelease creates new multi release.
func (cs *ConfigServer) CreateMultiRelease(ctx context.Context,
	req *pb.CreateMultiReleaseReq) (*pb.CreateMultiReleaseResp, error) {

	rtime := time.Now()
	kit := common.RequestKit(ctx)
	logger.V(2).Infof("%s[%s]| appcode: %s, user: %s, input[%+v]", kit.Method, kit.Rid, kit.AppCode, kit.User, req)

	response := new(pb.CreateMultiReleaseResp)

	defer func() {
		cost := cs.collector.StatRequest(kit.Method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", kit.Method, kit.Rid, cost, response)
	}()

	action := multireleaseaction.NewCreateAction(kit, cs.viper, cs.authSvrCli, cs.dataMgrCli, req, response)
	if err := cs.executor.ExecuteWithAuth(action); err != nil {
		logger.Errorf("%s[%s]| %+v", kit.Method, kit.Rid, err)
	}

	return response, nil
}

// QueryMultiRelease returns target multi release.
func (cs *ConfigServer) QueryMultiRelease(ctx context.Context,
	req *pb.QueryMultiReleaseReq) (*pb.QueryMultiReleaseResp, error) {

	rtime := time.Now()
	kit := common.RequestKit(ctx)
	logger.V(2).Infof("%s[%s]| appcode: %s, user: %s, input[%+v]", kit.Method, kit.Rid, kit.AppCode, kit.User, req)

	response := new(pb.QueryMultiReleaseResp)

	defer func() {
		cost := cs.collector.StatRequest(kit.Method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", kit.Method, kit.Rid, cost, response)
	}()

	action := multireleaseaction.NewQueryAction(kit, cs.viper, cs.dataMgrCli, req, response)
	if err := cs.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", kit.Method, kit.Rid, err)
	}

	return response, nil
}

// QueryHistoryMultiReleases returns history releases.
func (cs *ConfigServer) QueryHistoryMultiReleases(ctx context.Context,
	req *pb.QueryHistoryMultiReleasesReq) (*pb.QueryHistoryMultiReleasesResp, error) {

	rtime := time.Now()
	kit := common.RequestKit(ctx)
	logger.V(2).Infof("%s[%s]| appcode: %s, user: %s, input[%+v]", kit.Method, kit.Rid, kit.AppCode, kit.User, req)

	response := new(pb.QueryHistoryMultiReleasesResp)

	defer func() {
		cost := cs.collector.StatRequest(kit.Method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", kit.Method, kit.Rid, cost, response)
	}()

	action := multireleaseaction.NewListAction(kit, cs.viper, cs.dataMgrCli, req, response)
	if err := cs.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", kit.Method, kit.Rid, err)
	}

	return response, nil
}

// UpdateMultiRelease updates target multi release.
func (cs *ConfigServer) UpdateMultiRelease(ctx context.Context,
	req *pb.UpdateMultiReleaseReq) (*pb.UpdateMultiReleaseResp, error) {

	rtime := time.Now()
	kit := common.RequestKit(ctx)
	logger.V(2).Infof("%s[%s]| appcode: %s, user: %s, input[%+v]", kit.Method, kit.Rid, kit.AppCode, kit.User, req)

	response := new(pb.UpdateMultiReleaseResp)

	defer func() {
		cost := cs.collector.StatRequest(kit.Method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", kit.Method, kit.Rid, cost, response)
	}()

	action := multireleaseaction.NewUpdateAction(kit, cs.viper, cs.authSvrCli, cs.dataMgrCli, req, response)
	if err := cs.executor.ExecuteWithAuth(action); err != nil {
		logger.Errorf("%s[%s]| %+v", kit.Method, kit.Rid, err)
	}

	return response, nil
}

// CancelMultiRelease cancels target multi release.
func (cs *ConfigServer) CancelMultiRelease(ctx context.Context,
	req *pb.CancelMultiReleaseReq) (*pb.CancelMultiReleaseResp, error) {

	rtime := time.Now()
	kit := common.RequestKit(ctx)
	logger.V(2).Infof("%s[%s]| appcode: %s, user: %s, input[%+v]", kit.Method, kit.Rid, kit.AppCode, kit.User, req)

	response := new(pb.CancelMultiReleaseResp)

	defer func() {
		cost := cs.collector.StatRequest(kit.Method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", kit.Method, kit.Rid, cost, response)
	}()

	action := multireleaseaction.NewCancelAction(kit, cs.viper, cs.authSvrCli, cs.dataMgrCli, req, response)
	if err := cs.executor.ExecuteWithAuth(action); err != nil {
		logger.Errorf("%s[%s]| %+v", kit.Method, kit.Rid, err)
	}

	return response, nil
}

// PublishMultiRelease publishes target multi release.
func (cs *ConfigServer) PublishMultiRelease(ctx context.Context,
	req *pb.PublishMultiReleaseReq) (*pb.PublishMultiReleaseResp, error) {

	rtime := time.Now()
	kit := common.RequestKit(ctx)
	logger.V(2).Infof("%s[%s]| appcode: %s, user: %s, input[%+v]", kit.Method, kit.Rid, kit.AppCode, kit.User, req)

	response := new(pb.PublishMultiReleaseResp)

	defer func() {
		cost := cs.collector.StatRequest(kit.Method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", kit.Method, kit.Rid, cost, response)
	}()

	action := multireleaseaction.NewPublishAction(kit, cs.viper, cs.authSvrCli, cs.dataMgrCli,
		cs.gseControllerCli, req, response)

	if err := cs.executor.ExecuteWithAuth(action); err != nil {
		logger.Errorf("%s[%s]| %+v", kit.Method, kit.Rid, err)
	}

	return response, nil
}

// RollbackMultiRelease rollbacks target multi release.
func (cs *ConfigServer) RollbackMultiRelease(ctx context.Context,
	req *pb.RollbackMultiReleaseReq) (*pb.RollbackMultiReleaseResp, error) {

	rtime := time.Now()
	kit := common.RequestKit(ctx)
	logger.V(2).Infof("%s[%s]| appcode: %s, user: %s, input[%+v]", kit.Method, kit.Rid, kit.AppCode, kit.User, req)

	response := new(pb.RollbackMultiReleaseResp)

	defer func() {
		cost := cs.collector.StatRequest(kit.Method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", kit.Method, kit.Rid, cost, response)
	}()

	action := multireleaseaction.NewRollbackAction(kit, cs.viper, cs.authSvrCli, cs.dataMgrCli,
		cs.gseControllerCli, req, response)

	if err := cs.executor.ExecuteWithAuth(action); err != nil {
		logger.Errorf("%s[%s]| %+v", kit.Method, kit.Rid, err)
	}

	return response, nil
}

// QueryEffectedAppInstances returns sidecar instances which effected target release of the config.
func (cs *ConfigServer) QueryEffectedAppInstances(ctx context.Context,
	req *pb.QueryEffectedAppInstancesReq) (*pb.QueryEffectedAppInstancesResp, error) {

	rtime := time.Now()
	kit := common.RequestKit(ctx)
	logger.V(2).Infof("%s[%s]| appcode: %s, user: %s, input[%+v]", kit.Method, kit.Rid, kit.AppCode, kit.User, req)

	response := new(pb.QueryEffectedAppInstancesResp)

	defer func() {
		cost := cs.collector.StatRequest(kit.Method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", kit.Method, kit.Rid, cost, response)
	}()

	action := appinstanceaction.NewEffectedAction(kit, cs.viper, cs.dataMgrCli, req, response)
	if err := cs.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", kit.Method, kit.Rid, err)
	}

	return response, nil
}

// QueryMatchedAppInstances returns sidecar instances which matched target release or strategy.
func (cs *ConfigServer) QueryMatchedAppInstances(ctx context.Context,
	req *pb.QueryMatchedAppInstancesReq) (*pb.QueryMatchedAppInstancesResp, error) {

	rtime := time.Now()
	kit := common.RequestKit(ctx)
	logger.V(2).Infof("%s[%s]| appcode: %s, user: %s, input[%+v]", kit.Method, kit.Rid, kit.AppCode, kit.User, req)

	response := new(pb.QueryMatchedAppInstancesResp)

	defer func() {
		cost := cs.collector.StatRequest(kit.Method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", kit.Method, kit.Rid, cost, response)
	}()

	action := appinstanceaction.NewMatchedAction(kit, cs.viper, cs.dataMgrCli, req, response)
	if err := cs.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", kit.Method, kit.Rid, err)
	}

	return response, nil
}

// QueryReachableAppInstances returns sidecar instances which reachable of the app/cluster/zone.
func (cs *ConfigServer) QueryReachableAppInstances(ctx context.Context,
	req *pb.QueryReachableAppInstancesReq) (*pb.QueryReachableAppInstancesResp, error) {

	rtime := time.Now()
	kit := common.RequestKit(ctx)
	logger.V(2).Infof("%s[%s]| appcode: %s, user: %s, input[%+v]", kit.Method, kit.Rid, kit.AppCode, kit.User, req)

	response := new(pb.QueryReachableAppInstancesResp)

	defer func() {
		cost := cs.collector.StatRequest(kit.Method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", kit.Method, kit.Rid, cost, response)
	}()

	action := appinstanceaction.NewReachableAction(kit, cs.viper, cs.dataMgrCli, req, response)
	if err := cs.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", kit.Method, kit.Rid, err)
	}

	return response, nil
}

// QueryAppInstanceRelease returns release of target app instance.
func (cs *ConfigServer) QueryAppInstanceRelease(ctx context.Context,
	req *pb.QueryAppInstanceReleaseReq) (*pb.QueryAppInstanceReleaseResp, error) {

	rtime := time.Now()
	kit := common.RequestKit(ctx)
	logger.V(2).Infof("%s[%s]| appcode: %s, user: %s, input[%+v]", kit.Method, kit.Rid, kit.AppCode, kit.User, req)

	response := new(pb.QueryAppInstanceReleaseResp)

	defer func() {
		cost := cs.collector.StatRequest(kit.Method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", kit.Method, kit.Rid, cost, response)
	}()

	action := appinstanceaction.NewReleaseAction(kit, cs.viper, cs.dataMgrCli, req, response)
	if err := cs.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", kit.Method, kit.Rid, err)
	}

	return response, nil
}

// CreateStrategy creates new strategy.
func (cs *ConfigServer) CreateStrategy(ctx context.Context, req *pb.CreateStrategyReq) (*pb.CreateStrategyResp, error) {
	rtime := time.Now()
	kit := common.RequestKit(ctx)
	logger.V(2).Infof("%s[%s]| appcode: %s, user: %s, input[%+v]", kit.Method, kit.Rid, kit.AppCode, kit.User, req)

	response := new(pb.CreateStrategyResp)

	defer func() {
		cost := cs.collector.StatRequest(kit.Method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", kit.Method, kit.Rid, cost, response)
	}()

	action := strategyaction.NewCreateAction(kit, cs.viper, cs.authSvrCli, cs.dataMgrCli, req, response)
	if err := cs.executor.ExecuteWithAuth(action); err != nil {
		logger.Errorf("%s[%s]| %+v", kit.Method, kit.Rid, err)
	}

	return response, nil
}

// QueryStrategy returns target strategy.
func (cs *ConfigServer) QueryStrategy(ctx context.Context, req *pb.QueryStrategyReq) (*pb.QueryStrategyResp, error) {
	rtime := time.Now()
	kit := common.RequestKit(ctx)
	logger.V(2).Infof("%s[%s]| appcode: %s, user: %s, input[%+v]", kit.Method, kit.Rid, kit.AppCode, kit.User, req)

	response := new(pb.QueryStrategyResp)

	defer func() {
		cost := cs.collector.StatRequest(kit.Method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", kit.Method, kit.Rid, cost, response)
	}()

	action := strategyaction.NewQueryAction(kit, cs.viper, cs.dataMgrCli, req, response)
	if err := cs.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", kit.Method, kit.Rid, err)
	}

	return response, nil
}

// QueryStrategyList returns all strategies.
func (cs *ConfigServer) QueryStrategyList(ctx context.Context,
	req *pb.QueryStrategyListReq) (*pb.QueryStrategyListResp, error) {

	rtime := time.Now()
	kit := common.RequestKit(ctx)
	logger.V(2).Infof("%s[%s]| appcode: %s, user: %s, input[%+v]", kit.Method, kit.Rid, kit.AppCode, kit.User, req)

	response := new(pb.QueryStrategyListResp)

	defer func() {
		cost := cs.collector.StatRequest(kit.Method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", kit.Method, kit.Rid, cost, response)
	}()

	action := strategyaction.NewListAction(kit, cs.viper, cs.dataMgrCli, req, response)
	if err := cs.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", kit.Method, kit.Rid, err)
	}

	return response, nil
}

// DeleteStrategy deletes target strategy.
func (cs *ConfigServer) DeleteStrategy(ctx context.Context, req *pb.DeleteStrategyReq) (*pb.DeleteStrategyResp, error) {
	rtime := time.Now()
	kit := common.RequestKit(ctx)
	logger.V(2).Infof("%s[%s]| appcode: %s, user: %s, input[%+v]", kit.Method, kit.Rid, kit.AppCode, kit.User, req)

	response := new(pb.DeleteStrategyResp)

	defer func() {
		cost := cs.collector.StatRequest(kit.Method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", kit.Method, kit.Rid, cost, response)
	}()

	action := strategyaction.NewDeleteAction(kit, cs.viper, cs.authSvrCli, cs.dataMgrCli, req, response)
	if err := cs.executor.ExecuteWithAuth(action); err != nil {
		logger.Errorf("%s[%s]| %+v", kit.Method, kit.Rid, err)
	}

	return response, nil
}

// CreateProcAttr creates new ProcAttr.
func (cs *ConfigServer) CreateProcAttr(ctx context.Context, req *pb.CreateProcAttrReq) (*pb.CreateProcAttrResp, error) {
	rtime := time.Now()
	kit := common.RequestKit(ctx)
	logger.V(2).Infof("%s[%s]| appcode: %s, user: %s, input[%+v]", kit.Method, kit.Rid, kit.AppCode, kit.User, req)

	response := new(pb.CreateProcAttrResp)

	defer func() {
		cost := cs.collector.StatRequest(kit.Method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", kit.Method, kit.Rid, cost, response)
	}()

	action := procattraction.NewCreateAction(kit, cs.viper, cs.authSvrCli, cs.dataMgrCli, req, response)
	if err := cs.executor.ExecuteWithAuth(action); err != nil {
		logger.Errorf("%s[%s]| %+v", kit.Method, kit.Rid, err)
	}

	return response, nil
}

// CreateProcAttrBatch creates new ProcAttrs in batch mode.
func (cs *ConfigServer) CreateProcAttrBatch(ctx context.Context, req *pb.CreateProcAttrBatchReq) (*pb.CreateProcAttrBatchResp, error) {
	rtime := time.Now()
	kit := common.RequestKit(ctx)
	logger.V(2).Infof("%s[%s]| appcode: %s, user: %s, input[%+v]", kit.Method, kit.Rid, kit.AppCode, kit.User, req)

	response := new(pb.CreateProcAttrBatchResp)

	defer func() {
		cost := cs.collector.StatRequest(kit.Method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", kit.Method, kit.Rid, cost, response)
	}()

	action := procattraction.NewCreateBatchAction(kit, cs.viper, cs.authSvrCli, cs.dataMgrCli, req, response)
	if err := cs.executor.ExecuteWithAuth(action); err != nil {
		logger.Errorf("%s[%s]| %+v", kit.Method, kit.Rid, err)
	}

	return response, nil
}

// QueryHostProcAttr returns ProcAttr of target app on the host.
func (cs *ConfigServer) QueryHostProcAttr(ctx context.Context,
	req *pb.QueryHostProcAttrReq) (*pb.QueryHostProcAttrResp, error) {

	rtime := time.Now()
	kit := common.RequestKit(ctx)
	logger.V(2).Infof("%s[%s]| appcode: %s, user: %s, input[%+v]", kit.Method, kit.Rid, kit.AppCode, kit.User, req)

	response := new(pb.QueryHostProcAttrResp)

	defer func() {
		cost := cs.collector.StatRequest(kit.Method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", kit.Method, kit.Rid, cost, response)
	}()

	action := procattraction.NewQueryAction(kit, cs.viper, cs.dataMgrCli, req, response)
	if err := cs.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", kit.Method, kit.Rid, err)
	}

	return response, nil
}

// QueryHostProcAttrList returns ProcAttr list on target host.
func (cs *ConfigServer) QueryHostProcAttrList(ctx context.Context,
	req *pb.QueryHostProcAttrListReq) (*pb.QueryHostProcAttrListResp, error) {

	rtime := time.Now()
	kit := common.RequestKit(ctx)
	logger.V(2).Infof("%s[%s]| appcode: %s, user: %s, input[%+v]", kit.Method, kit.Rid, kit.AppCode, kit.User, req)

	response := new(pb.QueryHostProcAttrListResp)

	defer func() {
		cost := cs.collector.StatRequest(kit.Method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", kit.Method, kit.Rid, cost, response)
	}()

	action := procattraction.NewHostListAction(kit, cs.viper, cs.dataMgrCli, req, response)
	if err := cs.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", kit.Method, kit.Rid, err)
	}

	return response, nil
}

// QueryAppProcAttrList returns ProcAttr list of target app.
func (cs *ConfigServer) QueryAppProcAttrList(ctx context.Context,
	req *pb.QueryAppProcAttrListReq) (*pb.QueryAppProcAttrListResp, error) {

	rtime := time.Now()
	kit := common.RequestKit(ctx)
	logger.V(2).Infof("%s[%s]| appcode: %s, user: %s, input[%+v]", kit.Method, kit.Rid, kit.AppCode, kit.User, req)

	response := new(pb.QueryAppProcAttrListResp)

	defer func() {
		cost := cs.collector.StatRequest(kit.Method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", kit.Method, kit.Rid, cost, response)
	}()

	action := procattraction.NewAppListAction(kit, cs.viper, cs.dataMgrCli, req, response)
	if err := cs.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", kit.Method, kit.Rid, err)
	}

	return response, nil
}

// UpdateProcAttr updates target app ProcAttr on the host.
func (cs *ConfigServer) UpdateProcAttr(ctx context.Context, req *pb.UpdateProcAttrReq) (*pb.UpdateProcAttrResp, error) {
	rtime := time.Now()
	kit := common.RequestKit(ctx)
	logger.V(2).Infof("%s[%s]| appcode: %s, user: %s, input[%+v]", kit.Method, kit.Rid, kit.AppCode, kit.User, req)

	response := new(pb.UpdateProcAttrResp)

	defer func() {
		cost := cs.collector.StatRequest(kit.Method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", kit.Method, kit.Rid, cost, response)
	}()

	action := procattraction.NewUpdateAction(kit, cs.viper, cs.authSvrCli, cs.dataMgrCli, req, response)
	if err := cs.executor.ExecuteWithAuth(action); err != nil {
		logger.Errorf("%s[%s]| %+v", kit.Method, kit.Rid, err)
	}

	return response, nil
}

// DeleteProcAttr deletes target app ProcAttr on the host.
func (cs *ConfigServer) DeleteProcAttr(ctx context.Context, req *pb.DeleteProcAttrReq) (*pb.DeleteProcAttrResp, error) {
	rtime := time.Now()
	kit := common.RequestKit(ctx)
	logger.V(2).Infof("%s[%s]| appcode: %s, user: %s, input[%+v]", kit.Method, kit.Rid, kit.AppCode, kit.User, req)

	response := new(pb.DeleteProcAttrResp)

	defer func() {
		cost := cs.collector.StatRequest(kit.Method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", kit.Method, kit.Rid, cost, response)
	}()

	action := procattraction.NewDeleteAction(kit, cs.viper, cs.authSvrCli, cs.dataMgrCli, req, response)
	if err := cs.executor.ExecuteWithAuth(action); err != nil {
		logger.Errorf("%s[%s]| %+v", kit.Method, kit.Rid, err)
	}

	return response, nil
}

// QueryAuditList returns history audits.
func (cs *ConfigServer) QueryAuditList(ctx context.Context, req *pb.QueryAuditListReq) (*pb.QueryAuditListResp, error) {
	rtime := time.Now()
	kit := common.RequestKit(ctx)
	logger.V(2).Infof("%s[%s]| appcode: %s, user: %s, input[%+v]", kit.Method, kit.Rid, kit.AppCode, kit.User, req)

	response := new(pb.QueryAuditListResp)

	defer func() {
		cost := cs.collector.StatRequest(kit.Method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", kit.Method, kit.Rid, cost, response)
	}()

	action := auditaction.NewListAction(kit, cs.viper, cs.dataMgrCli, req, response)
	if err := cs.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", kit.Method, kit.Rid, err)
	}

	return response, nil
}

// Reload reloads target release or multi release.
func (cs *ConfigServer) Reload(ctx context.Context, req *pb.ReloadReq) (*pb.ReloadResp, error) {
	rtime := time.Now()
	kit := common.RequestKit(ctx)
	logger.V(2).Infof("%s[%s]| appcode: %s, user: %s, input[%+v]", kit.Method, kit.Rid, kit.AppCode, kit.User, req)

	response := new(pb.ReloadResp)

	defer func() {
		cost := cs.collector.StatRequest(kit.Method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", kit.Method, kit.Rid, cost, response)
	}()

	action := releaseaction.NewReloadAction(kit, cs.viper, cs.authSvrCli, cs.dataMgrCli,
		cs.gseControllerCli, req, response)

	if err := cs.executor.ExecuteWithAuth(action); err != nil {
		logger.Errorf("%s[%s]| %+v", kit.Method, kit.Rid, err)
	}

	return response, nil
}

// Healthz returns server health informations.
func (cs *ConfigServer) Healthz(ctx context.Context, req *pb.HealthzReq) (*pb.HealthzResp, error) {
	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.HealthzResp)

	defer func() {
		cost := cs.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := healthzaction.NewAction(ctx, cs.viper, req, response)
	if err := cs.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}
