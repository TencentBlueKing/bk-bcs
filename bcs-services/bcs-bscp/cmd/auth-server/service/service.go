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
 */

// Package service NOTES
package service

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"

	bkiam "github.com/TencentBlueKing/iam-go-sdk"
	bkiamlogger "github.com/TencentBlueKing/iam-go-sdk/logger"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	structpb "google.golang.org/protobuf/types/known/structpb"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/cmd/auth-server/options"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/cmd/auth-server/service/auth"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/cmd/auth-server/service/iam"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/cmd/auth-server/service/initial"
	confsvc "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/cmd/config-server/service"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/cc"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/components/bkpaas"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/errf"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/iam/apigw"
	iamauth "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/iam/auth"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/iam/client"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/iam/meta"
	pkgauth "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/iam/sdk/auth"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/iam/sys"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/metrics"
	pbas "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/auth-server"
	pbcs "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/config-server"
	base "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/base"
	basepb "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/base"
	pbds "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/data-service"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/rest/view/webannotation"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/serviced"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/space"
	esbcli "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/thirdparty/esb/client"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/tools"
)

// Service do all the data service's work
type Service struct {
	client  *ClientSet
	gateway *gateway
	// disableAuth defines whether iam authorization is disabled
	disableAuth bool
	// disableWriteOpt defines which biz's write operation needs to be disabled
	disableWriteOpt *options.DisableWriteOption
	iamSettings     cc.IAM
	// iam logic module.
	iam *iam.IAM
	// initial logic module.
	initial *initial.Initial
	// auth logic module.
	auth     *auth.Auth
	spaceMgr *space.Manager
	pubKey   string
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

	spaceMgr, err := space.NewSpaceMgr(context.Background(), client.Esb)
	if err != nil {
		return nil, errors.Wrap(err, "init space mgr")
	}

	s := &Service{
		client:          client,
		gateway:         gateway,
		disableAuth:     disableAuth,
		disableWriteOpt: disableWriteOpt,
		iamSettings:     iamSettings,
		spaceMgr:        spaceMgr,
	}

	if errH := s.handlerAutoRegister(); errH != nil {
		return nil, errH
	}

	if err = s.initLogicModule(); err != nil {
		return nil, err
	}

	return s, nil
}

// 注册网关
func (s *Service) handlerAutoRegister() error {
	s.pubKey = cc.AuthServer().ApiGateway.GWPubKey
	if cc.AuthServer().ApiGateway.AutoRegister {
		gw, err := apigw.NewApiGw(cc.AuthServer().Esb, cc.AuthServer().ApiGateway)
		if err != nil {
			return err
		}

		result, err := gw.GetApigwPublicKey(apigw.Name)
		if err != nil {
			return err
		}
		if result.Code != 0 && result.Data.PublicKey == "" {
			return fmt.Errorf("get the gateway public key failed, err: %s", result.Message)
		}
		s.pubKey = result.Data.PublicKey
	}

	return nil
}

// Handler return service's handler.
func (s *Service) Handler() (http.Handler, error) {
	if s.gateway == nil {
		return nil, errors.New("gateway is nil")
	}

	return s.gateway.handler(), nil
}

// nolint: funlen
func newClientSet(sd serviced.Discover, tls cc.TLSConfig, iamSettings cc.IAM, disableAuth bool) (
	*ClientSet, error) {

	logs.Infof("start initialize the client set.")

	opts := make([]grpc.DialOption, 0)

	// add dial load balancer.
	opts = append(opts, sd.LBRoundRobin())

	if !tls.Enable() {
		// dial without ssl
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
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
		Address:   []string{iamSettings.APIURL},
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

	esbSetting := cc.AuthServer().Esb
	esbCli, err := esbcli.NewClient(&esbSetting, metrics.Register())
	if err != nil {
		return nil, err
	}

	log := &logrus.Logger{
		Out:          os.Stderr,
		Formatter:    new(logrus.TextFormatter),
		Hooks:        make(logrus.LevelHooks),
		Level:        logrus.DebugLevel,
		ExitFunc:     os.Exit,
		ReportCaller: false,
	}
	bkiamlogger.SetLogger(log)
	apiGatewayIAM := bkiam.NewAPIGatewayIAM(
		sys.SystemIDBSCP, iamSettings.AppCode, iamSettings.AppSecret, iamSettings.APIURL)

	cs := &ClientSet{
		DS:        ds,
		sys:       iamSys,
		auth:      authSdk,
		Esb:       esbCli,
		iamClient: apiGatewayIAM,
	}
	logs.Infof("initialize the client set success.")
	return cs, nil
}

// ClientSet defines configure server's all the depends api client.
type ClientSet struct {
	// data service's sys api
	DS pbds.DataClient
	// iam sys related operate.
	iamClient *bkiam.IAM
	sys       *sys.Sys
	// auth related operate.
	auth pkgauth.Authorizer
	// Esb Esb client api
	Esb esbcli.Client
}

// PullResource init auth center's auth model.
func (s *Service) PullResource(ctx context.Context, req *pbas.PullResourceReq) (*structpb.Struct, error) {
	return s.iam.PullResource(ctx, req)
}

// InitAuthCenter init auth center's auth model.
func (s *Service) InitAuthCenter(ctx context.Context, req *pbas.InitAuthCenterReq) (*pbas.InitAuthCenterResp, error) {
	return s.initial.InitAuthCenter(ctx, req)
}

// GetAuthConf get auth login conf
func (s *Service) GetAuthConf(_ context.Context,
	_ *pbas.GetAuthConfReq) (*pbas.GetAuthConfResp, error) {

	resp := &pbas.GetAuthConfResp{
		LoginAuth: &pbas.LoginAuth{
			Host:      cc.AuthServer().LoginAuth.Host,
			InnerHost: cc.AuthServer().LoginAuth.InnerHost,
			Provider:  cc.AuthServer().LoginAuth.Provider,
			GwPubkey:  s.pubKey,
			UseEsb:    false,
		},
		Esb: &pbas.ESB{
			Endpoints: cc.AuthServer().Esb.Endpoints,
			AppCode:   cc.AuthServer().Esb.AppCode,
			AppSecret: cc.AuthServer().Esb.AppSecret,
			User:      cc.AuthServer().Esb.User,
			Tls: &pbas.TLS{
				InsecureSkipVerify: cc.AuthServer().Esb.TLS.InsecureSkipVerify,
				CertFile:           cc.AuthServer().Esb.TLS.CertFile,
				KeyFile:            cc.AuthServer().Esb.TLS.KeyFile,
				CaFile:             cc.AuthServer().Esb.TLS.CAFile,
				Password:           cc.AuthServer().Esb.TLS.Password,
			},
		},
	}
	return resp, nil
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

// GrantResourceCreatorAction GetPermissionToApply get iam permission to apply.
func (s *Service) GrantResourceCreatorAction(ctx context.Context, req *pbas.
	GrantResourceCreatorActionReq) (*base.EmptyResp, error) {

	err := s.auth.GrantResourceCreatorAction(ctx, pbas.GrantResourceCreatorAction(req))
	return nil, err

}

// CheckPermission grpc check permission
func (s *Service) CheckPermission(ctx context.Context, req *pbas.CheckPermissionReq) (
	*pbas.CheckPermissionResp, error) {
	kt := kit.FromGrpcContext(ctx)

	resp := &pbas.CheckPermissionResp{
		IsAllowed: false,
		ApplyUrl:  "",
		Resources: []*pbas.BasicDetail{},
	}

	userInfo := &meta.UserInfo{UserName: kt.User}
	abReq := &pbas.AuthorizeBatchReq{
		User:      pbas.PbUserInfo(userInfo),
		Resources: req.Resources,
	}

	abResp, err := s.AuthorizeBatch(kt.RpcCtx(), abReq)
	if err != nil {
		logs.Errorf("authorize failed, req: %#v, err: %v, rid: %s", req, err, kt.Rid)
		return nil, err
	}

	authorized := true
	for _, decision := range abResp.Decisions {
		if !decision.Authorized {
			authorized = false
			break
		}
	}

	if authorized {
		resp.IsAllowed = true
		return resp, nil
	}

	gpReq := &pbas.GetPermissionToApplyReq{
		Resources: req.Resources,
	}

	permResp, err := s.GetPermissionToApply(kt.RpcCtx(), gpReq)
	if err != nil {
		logs.Errorf("get permission to apply failed, req: %#v, err: %v, rid: %s", req, err, kt.Rid)
		return nil, errf.New(errf.DoAuthorizeFailed, "get permission to apply failed")
	}
	resp.ApplyUrl = permResp.ApplyUrl
	for _, action := range permResp.Permission.Actions {
		for _, resourceType := range action.RelatedResourceTypes {
			for _, instance := range resourceType.Instances {
				for _, i := range instance.Instances {
					if i.Type != resourceType.Type {
						continue
					}
					resp.Resources = append(resp.Resources, &pbas.BasicDetail{
						Type:         resourceType.Type,
						TypeName:     resourceType.TypeName,
						Action:       action.Id,
						ActionName:   action.Name,
						ResourceId:   i.Id,
						ResourceName: i.Name,
					})
				}
			}
		}
	}
	return resp, nil
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

	s.auth, err = auth.NewAuth(s.client.auth, s.client.DS, s.disableAuth, s.client.iamClient, s.disableWriteOpt,
		s.spaceMgr)
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

	conf := cc.AuthServer().LoginAuth
	authLoginClient := bkpaas.NewAuthLoginClient(&conf)

	var (
		username string
		err      error
	)

	if cc.AuthServer().LoginAuth.UseESB && cc.AuthServer().LoginAuth.Provider != bkpaas.BKLoginProvider {
		username, err = s.client.Esb.BKLogin().IsLogin(ctx, token)
	} else {
		username, err = authLoginClient.GetUserInfoByToken(ctx, host, req.GetUid(), token)
	}

	if err != nil {
		if errors.Is(err, errf.ErrPermissionDenied) {
			return nil, status.New(codes.PermissionDenied, errf.GetErrMsg(err)).Err()
		}
		return nil, err
	}

	return &pbas.UserInfoResp{Username: username, AvatarUrl: ""}, nil
}

// ListUserSpaceAnnotation list user space permission annotations
func ListUserSpaceAnnotation(ctx context.Context, kt *kit.Kit, authorizer iamauth.Authorizer,
	msg proto.Message) (*webannotation.Annotation, error) {

	resp, ok := msg.(*pbas.ListUserSpaceResp)
	if !ok {
		return nil, nil
	}

	perms := map[string]webannotation.Perm{}
	authRes := make([]*meta.ResourceAttribute, 0, len(resp.GetItems()))
	for _, v := range resp.GetItems() {
		bID, _ := strconv.ParseInt(v.SpaceId, 10, 64)
		authRes = append(authRes, &meta.ResourceAttribute{
			Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource, ResourceID: uint32(bID)}, BizID: uint32(bID)},
		)

	}

	authResp, _, err := authorizer.AuthorizeDecision(kt, authRes...)
	if err != nil {
		return nil, err
	}

	for idx, v := range resp.GetItems() {
		perms[v.SpaceId] = webannotation.Perm{string(meta.FindBusinessResource): authResp[idx].Authorized}
	}

	return &webannotation.Annotation{Perms: perms}, nil
}

func init() {
	webannotation.Register(&pbas.ListUserSpaceResp{}, ListUserSpaceAnnotation)
	webannotation.Register(&pbcs.ListAppsResp{}, confsvc.ListAppsAnnotation)
}

// ListUserSpace 获取用户信息
func (s *Service) ListUserSpace(ctx context.Context, req *pbas.ListUserSpaceReq) (*pbas.ListUserSpaceResp, error) {
	kt := kit.FromGrpcContext(ctx)
	if kt.User == "" {
		err := basepb.InvalidArgumentsErr(&basepb.InvalidArgument{
			Field:   "kit.user",
			Message: "kit.user not found in metadata",
		})

		return nil, err
	}

	// 定期同步
	spaceList := s.spaceMgr.AllSpaces()

	items := make([]*pbas.Space, 0, len(spaceList))
	for _, space := range spaceList {
		items = append(items, &pbas.Space{
			SpaceId:       space.SpaceId,
			SpaceName:     space.SpaceName,
			SpaceTypeId:   space.SpaceTypeID,
			SpaceTypeName: space.SpaceTypeName,
			SpaceUid:      space.SpaceUid,
			SpaceEnName:   space.SpaceEnName,
		})
	}

	return &pbas.ListUserSpaceResp{Items: items}, nil
}

// QuerySpace 查询 space 信息
func (s *Service) QuerySpace(ctx context.Context, req *pbas.QuerySpaceReq) (*pbas.QuerySpaceResp, error) {
	uidList := req.GetSpaceUid()
	if len(uidList) == 0 {
		return &pbas.QuerySpaceResp{}, nil
	}

	spaceList, err := s.spaceMgr.QuerySpace(uidList)
	if err != nil {
		return nil, err
	}

	items := make([]*pbas.Space, 0, len(spaceList))
	for _, space := range spaceList {
		items = append(items, &pbas.Space{
			SpaceId:       space.SpaceId,
			SpaceName:     space.SpaceName,
			SpaceTypeId:   space.SpaceTypeID,
			SpaceTypeName: space.SpaceTypeName,
			SpaceUid:      space.SpaceUid,
		})
	}

	return &pbas.QuerySpaceResp{Items: items}, nil
}

// QuerySpaceByAppID 查询space
func (s *Service) QuerySpaceByAppID(ctx context.Context, req *pbas.QuerySpaceByAppIDReq) (*pbas.Space, error) {
	kt := kit.FromGrpcContext(ctx)
	appID := req.GetAppId()
	if appID == 0 {
		return nil, errors.New("app_id is required")
	}

	app, err := s.client.DS.GetAppByID(kt.RpcCtx(), &pbds.GetAppByIDReq{AppId: appID})
	if err != nil {
		return nil, err
	}

	resp := &pbas.Space{
		SpaceId:       strconv.Itoa(int(app.BizId)),
		SpaceTypeId:   space.BK_CMDB.ID,
		SpaceTypeName: space.BK_CMDB.Name,
	}
	return resp, nil
}
