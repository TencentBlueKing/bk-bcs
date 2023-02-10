/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

// Package service NOTES
package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"bscp.io/cmd/auth-server/options"
	"bscp.io/cmd/auth-server/service/auth"
	"bscp.io/cmd/auth-server/service/iam"
	"bscp.io/cmd/auth-server/service/initial"
	"bscp.io/pkg/cc"
	"bscp.io/pkg/components/bkpaas"
	"bscp.io/pkg/iam/client"
	pkgauth "bscp.io/pkg/iam/sdk/auth"
	"bscp.io/pkg/iam/sys"
	"bscp.io/pkg/logs"
	"bscp.io/pkg/metrics"
	pbas "bscp.io/pkg/protocol/auth-server"
	pbds "bscp.io/pkg/protocol/data-service"
	"bscp.io/pkg/serviced"
	"bscp.io/pkg/tools"
)

// Service do all the data service's work
type Service struct {
	client  *ClientSet
	gateway *gateway
	// disableAuth defines whether iam authorization is disabled
	disableAuth bool
	// disableWriteOpt defines which biz's write operation needs to be disabled
	disableWriteOpt *options.DisableWriteOption

	// iam logic module.
	iam *iam.IAM
	// initial logic module.
	initial *initial.Initial
	// auth logic module.
	auth *auth.Auth
}

// NewService create a service instance.
func NewService(sd serviced.Discover, iamSettings cc.IAM, disableAuth bool,
	disableWriteOpt *options.DisableWriteOption) (*Service, error) {

	client, err := newClientSet(sd, cc.AuthServer().Network.TLS, iamSettings, disableAuth)
	if err != nil {
		return nil, fmt.Errorf("new client set failed, err: %v", err)
	}

	state, ok := sd.(serviced.State)
	if !ok {
		return nil, errors.New("discover convert state failed")
	}
	gateway, err := newGateway(state, client.sys)
	if err != nil {
		return nil, fmt.Errorf("new gateway failed, err: %v", err)
	}

	s := &Service{
		client:          client,
		gateway:         gateway,
		disableAuth:     disableAuth,
		disableWriteOpt: disableWriteOpt,
	}

	if err = s.initLogicModule(); err != nil {
		return nil, err
	}

	return s, nil
}

// Handler return service's handler.
func (s *Service) Handler() (http.Handler, error) {
	if s.gateway == nil {
		return nil, errors.New("gateway is nil")
	}

	return s.gateway.handler(), nil
}

func newClientSet(sd serviced.Discover, tls cc.TLSConfig, iamSettings cc.IAM, disableAuth bool) (
	*ClientSet, error) {

	logs.Infof("start initialize the client set.")

	opts := make([]grpc.DialOption, 0)

	// add dial load balancer.
	opts = append(opts, sd.LBRoundRobin())

	if !tls.Enable() {
		// dial without ssl
		opts = append(opts, grpc.WithInsecure())
	} else {
		// dial with ssl.
		tlsC, err := tools.ClientTLSConfVerify(tls.InsecureSkipVerify, tls.CAFile, tls.CertFile, tls.KeyFile,
			tls.Password)
		if err != nil {
			return nil, fmt.Errorf("init etcd tls config failed, err: %v", err)
		}

		cred := credentials.NewTLS(tlsC)
		opts = append(opts, grpc.WithTransportCredentials(cred))
	}

	// connect data service.
	dsConn, err := grpc.Dial(serviced.GrpcServiceDiscoveryName(cc.DataServiceName), opts...)
	if err != nil {
		return nil, fmt.Errorf("dial data service failed, err: %v", err)
	}
	ds := pbds.NewDataClient(dsConn)
	logs.Infof("initialize data service client success.")

	tlsConfig := new(tools.TLSConfig)
	if iamSettings.TLS.Enable() {
		tlsConfig = &tools.TLSConfig{
			InsecureSkipVerify: iamSettings.TLS.InsecureSkipVerify,
			CertFile:           iamSettings.TLS.CertFile,
			KeyFile:            iamSettings.TLS.KeyFile,
			CAFile:             iamSettings.TLS.CAFile,
			Password:           iamSettings.TLS.Password,
		}
	}
	cfg := &client.Config{
		Address:   iamSettings.Endpoints,
		AppCode:   iamSettings.AppCode,
		AppSecret: iamSettings.AppSecret,
		SystemID:  sys.SystemIDBSCP,
		TLS:       tlsConfig,
	}
	iamCli, err := client.NewClient(cfg, metrics.Register())
	if err != nil {
		return nil, err
	}

	iamSys, err := sys.NewSys(iamCli)
	if err != nil {
		return nil, fmt.Errorf("new iam sys failed, err: %v", err)
	}
	logs.Infof("initialize iam sys success.")

	// initialize iam auth sdk
	iamLgc, err := iam.NewIAM(ds, iamSys, disableAuth)
	if err != nil {
		return nil, fmt.Errorf("new iam logics failed, err: %v", err)
	}

	authSdk, err := pkgauth.NewAuth(iamCli, iamLgc)
	if err != nil {
		return nil, fmt.Errorf("new iam auth sdk failed, err: %v", err)
	}
	logs.Infof("initialize iam auth sdk success.")

	cs := &ClientSet{
		DS:   ds,
		sys:  iamSys,
		auth: authSdk,
	}
	logs.Infof("initialize the client set success.")
	return cs, nil
}

// ClientSet defines configure server's all the depends api client.
type ClientSet struct {
	// data service's sys api
	DS pbds.DataClient
	// iam sys related operate.
	sys *sys.Sys
	// auth related operate.
	auth pkgauth.Authorizer
}

// PullResource init auth center's auth model.
func (s *Service) PullResource(ctx context.Context, req *pbas.PullResourceReq) (*pbas.PullResourceResp, error) {
	return s.iam.PullResource(ctx, req)
}

// InitAuthCenter init auth center's auth model.
func (s *Service) InitAuthCenter(ctx context.Context, req *pbas.InitAuthCenterReq) (*pbas.InitAuthCenterResp, error) {
	return s.initial.InitAuthCenter(ctx, req)
}

// AuthorizeBatch authorize resource batch.
func (s *Service) AuthorizeBatch(ctx context.Context, req *pbas.AuthorizeBatchReq) (*pbas.AuthorizeBatchResp, error) {
	return s.auth.AuthorizeBatch(ctx, req)
}

// GetPermissionToApply get iam permission to apply.
func (s *Service) GetPermissionToApply(ctx context.Context, req *pbas.GetPermissionToApplyReq) (
	*pbas.GetPermissionToApplyResp, error) {

	return s.auth.GetPermissionToApply(ctx, req)
}

// initLogicModule init logic module.
func (s *Service) initLogicModule() error {
	var err error

	s.initial, err = initial.NewInitial(s.client.sys, s.disableAuth)
	if err != nil {
		return err
	}

	s.iam, err = iam.NewIAM(s.client.DS, s.client.sys, s.disableAuth)
	if err != nil {
		return err
	}

	s.auth, err = auth.NewAuth(s.client.auth, s.client.DS, s.disableAuth, s.disableWriteOpt)
	if err != nil {
		return err
	}

	return nil
}

// GetUserInfo 获取用户信息
func (s *Service) GetUserInfo(ctx context.Context, req *pbas.UserCredentialReq) (*pbas.UserInfoResp, error) {
	token := req.GetToken()
	if token == "" {
		return nil, errors.New("token not provided")
	}

	// 优先使用 InnerHost
	host := cc.AuthServer().LoginAuth.Host
	if cc.AuthServer().LoginAuth.InnerHost != "" {
		host = cc.AuthServer().LoginAuth.InnerHost
	}

	username, err := bkpaas.GetUserInfoByToken(ctx, host, req.GetUid(), token)
	if err != nil {
		return nil, err
	}

	return &pbas.UserInfoResp{Username: username, AvatarUrl: ""}, nil
}
