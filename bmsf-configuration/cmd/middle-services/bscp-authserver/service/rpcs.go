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

	authaction "bk-bscp/cmd/middle-services/bscp-authserver/actions/auth"
	healthzaction "bk-bscp/cmd/middle-services/bscp-authserver/actions/healthz"
	policyaction "bk-bscp/cmd/middle-services/bscp-authserver/actions/policy"
	pb "bk-bscp/internal/protocol/authserver"
	"bk-bscp/pkg/common"
	"bk-bscp/pkg/logger"
)

// Authorize authorizes the request base on local or bkiam mode.
func (as *AuthServer) Authorize(ctx context.Context, req *pb.AuthorizeReq) (*pb.AuthorizeResp, error) {
	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.AuthorizeResp)

	defer func() {
		cost := as.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := authaction.NewAuthorizeAction(ctx, as.viper, as.authMode,
		as.localAuthController, as.bkiamAuthController, req, response)

	if err := as.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// AddPolicy adds a new policy to local auth or bkiam.
func (as *AuthServer) AddPolicy(ctx context.Context, req *pb.AddPolicyReq) (*pb.AddPolicyResp, error) {
	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.AddPolicyResp)

	defer func() {
		cost := as.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := policyaction.NewAddAction(ctx, as.viper, as.authMode,
		as.localAuthController, as.bkiamAuthController, req, response)

	if err := as.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// RemovePolicy removes a new policy from local auth or bkiam.
func (as *AuthServer) RemovePolicy(ctx context.Context, req *pb.RemovePolicyReq) (*pb.RemovePolicyResp, error) {
	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.RemovePolicyResp)

	defer func() {
		cost := as.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := policyaction.NewRemoveAction(ctx, as.viper, as.authMode,
		as.localAuthController, as.bkiamAuthController, req, response)

	if err := as.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}

// Healthz returns server health informations.
func (as *AuthServer) Healthz(ctx context.Context, req *pb.HealthzReq) (*pb.HealthzResp, error) {
	rtime := time.Now()
	method := common.GRPCMethod(ctx)
	logger.V(2).Infof("%s[%s]| input[%+v]", method, req.Seq, req)

	response := new(pb.HealthzResp)

	defer func() {
		cost := as.collector.StatRequest(method, response.Code, rtime, time.Now())
		logger.V(2).Infof("%s[%s]| output[%dms][%+v]", method, req.Seq, cost, response)
	}()

	action := healthzaction.NewAction(ctx, as.viper, as.authMode, req, response)
	if err := as.executor.Execute(action); err != nil {
		logger.Errorf("%s[%s]| %+v", method, req.Seq, err)
	}

	return response, nil
}
