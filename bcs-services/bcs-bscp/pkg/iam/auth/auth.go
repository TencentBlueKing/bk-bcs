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

// Package auth NOTES
package auth

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"k8s.io/klog/v2"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/cc"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/components/bkpaas"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/errf"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/i18n"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/iam/client"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/iam/meta"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/metrics"
	pbas "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/auth-server"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/rest"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/runtime/gwparser"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/serviced"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/space"
	esbcli "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/thirdparty/esb/client"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/tools"
)

// Authorizer defines all the supported functionalities to do auth operation.
type Authorizer interface {
	// AuthorizeDecision if user has permission to the resources, returns auth status per resource and for all.
	AuthorizeDecision(kt *kit.Kit, resources ...*meta.ResourceAttribute) ([]*meta.Decision, bool, error)
	// Authorize authorize if user has permission to the resources.
	// If user is unauthorized, assign apply url and resources into error.
	Authorize(kt *kit.Kit, resources ...*meta.ResourceAttribute) error
	// UnifiedAuthentication API 鉴权中间件
	UnifiedAuthentication(next http.Handler) http.Handler
	// GrantResourceCreatorAction grant a user's resource creator action.
	GrantResourceCreatorAction(kt *kit.Kit, opts *client.GrantResourceCreatorActionOption) error
	// WebAuthentication 网页鉴权中间件
	WebAuthentication(webHost string) func(http.Handler) http.Handler
	// AppVerified App校验中间件, 需要放到 UnifiedAuthentication 后面, url 需要添加 {app_id} 变量
	AppVerified(next http.Handler) http.Handler
	// BizVerified 业务鉴权
	BizVerified(next http.Handler) http.Handler
	// ContentVerified 内容(上传下载)鉴权
	ContentVerified(next http.Handler) http.Handler
	// LogOut handler will build login url, client should make redirect
	LogOut(r *http.Request) *rest.UnauthorizedData
	// HasBiz 业务是否存在
	HasBiz(bizID uint32) bool
}

// NewAuthorizer create an authorizer for iam authorize related operation.
func NewAuthorizer(sd serviced.Discover, tls cc.TLSConfig) (Authorizer, error) {
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
			return nil, fmt.Errorf("init client set tls config failed, err: %v", err)
		}

		cred := credentials.NewTLS(tlsC)
		opts = append(opts, grpc.WithTransportCredentials(cred))
	}

	// connect auth server.
	asConn, err := grpc.Dial(serviced.GrpcServiceDiscoveryName(cc.AuthServerName), opts...)
	if err != nil {
		logs.Errorf("dial auth server failed, err: %v", err)
		return nil, errf.New(errf.Unknown, fmt.Sprintf("dial auth server failed, err: %v", err))
	}

	authClient := pbas.NewAuthClient(asConn)
	resp, err := authClient.GetAuthConf(context.Background(), &pbas.GetAuthConfReq{})
	if err != nil {
		return nil, errors.Wrap(err, "get auth conf")
	}

	conf := &cc.LoginAuthSettings{
		Host:      resp.LoginAuth.Host,
		InnerHost: resp.LoginAuth.InnerHost,
		Provider:  resp.LoginAuth.Provider,
	}
	authLoginClient := bkpaas.NewAuthLoginClient(conf)
	klog.InfoS("init authlogin client done", "host", conf.Host, "inner_host", conf.InnerHost, "provider", conf.Provider)

	// init space manager
	esbSetting := &cc.Esb{
		Endpoints: resp.Esb.Endpoints,
		AppCode:   resp.Esb.AppCode,
		AppSecret: resp.Esb.AppSecret,
		User:      resp.Esb.User,
		TLS: cc.TLSConfig{
			InsecureSkipVerify: resp.Esb.Tls.InsecureSkipVerify,
			CertFile:           resp.Esb.Tls.CertFile,
			KeyFile:            resp.Esb.Tls.KeyFile,
			CAFile:             resp.Esb.Tls.CaFile,
			Password:           resp.Esb.Tls.Password,
		},
	}
	esbCli, err := esbcli.NewClient(esbSetting, metrics.Register())
	if err != nil {
		return nil, fmt.Errorf("new esb cleint failed, err: %v", err)
	}
	spaceMgr, err := space.NewSpaceMgr(context.Background(), esbCli)
	if err != nil {
		return nil, fmt.Errorf("init space manager failed, err: %v", err)
	}

	authz := &authorizer{
		authClient:      authClient,
		authLoginClient: authLoginClient,
		gwParser:        nil,
		spaceMgr:        spaceMgr,
	}

	// 如果有公钥，初始化配置
	if resp.LoginAuth.GwPubkey != "" {
		gwParser, err := gwparser.NewJWTParser(resp.LoginAuth.GwPubkey)
		if err != nil {
			return nil, errors.Wrap(err, "init gw parser")
		}

		authz.gwParser = gwParser
		klog.InfoS("init gw parser done", "fingerprint", gwParser.Fingerprint())
	}

	return authz, nil
}

type authorizer struct {
	// authClient auth server's client api
	authClient      pbas.AuthClient
	authLoginClient bkpaas.AuthLoginClient
	gwParser        gwparser.Parser
	spaceMgr        *space.Manager
}

// AuthorizeDecision if user has permission to the resources, returns auth status per resource and for all.
func (a authorizer) AuthorizeDecision(kt *kit.Kit, resources ...*meta.ResourceAttribute) (
	[]*meta.Decision, bool, error) {
	userInfo := &meta.UserInfo{UserName: kt.User}

	req := &pbas.AuthorizeBatchReq{
		User:      pbas.PbUserInfo(userInfo),
		Resources: pbas.PbResourceAttributes(resources),
	}

	resp, err := a.authClient.AuthorizeBatch(kt.RpcCtx(), req)
	if err != nil {
		logs.Errorf("authorize failed, req: %#v, err: %v, rid: %s", req, err, kt.Rid)
		return nil, false, err
	}

	authorized := true
	for _, decision := range resp.Decisions {
		if !decision.Authorized {
			authorized = false
			break
		}
	}

	return pbas.Decisions(resp.Decisions), authorized, nil
}

// Authorize authorize if user has permission to the resources.
// If user is unauthorized, assign apply url and resources into error.
func (a authorizer) Authorize(kt *kit.Kit, resources ...*meta.ResourceAttribute) error {
	_, authorized, err := a.AuthorizeDecision(kt, resources...)
	if err != nil {
		return errf.New(errf.DoAuthorizeFailed, i18n.T(kt, "authorize failed"))
	}

	if authorized {
		return nil
	}

	req := &pbas.GetPermissionToApplyReq{
		Resources: pbas.PbResourceAttributes(resources),
	}

	permResp, err := a.authClient.GetPermissionToApply(kt.RpcCtx(), req)
	if err != nil {
		logs.Errorf("get permission to apply failed, req: %#v, err: %v, rid: %s", req, err, kt.Rid)
		return errf.New(errf.DoAuthorizeFailed, i18n.T(kt, "get permission to apply failed, err: %v", err))
	}

	st := status.New(codes.PermissionDenied, "permission denied")
	details := pbas.ApplyDetail{
		Resources: []*pbas.BasicDetail{},
		ApplyUrl:  permResp.ApplyUrl,
	}
	for _, action := range permResp.Permission.Actions {
		for _, resourceType := range action.RelatedResourceTypes {
			for _, instance := range resourceType.Instances {
				for _, i := range instance.Instances {
					if i.Type != resourceType.Type {
						continue
					}
					details.Resources = append(details.Resources, &pbas.BasicDetail{
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
	st, err = st.WithDetails(&details)
	if err != nil {
		logs.Errorf("with details failed, err: %v", err)
		return errf.New(errf.PermissionDenied, i18n.T(kt, "grpc status with details failed, err: %v", err))
	}
	return st.Err()
}

// GrantResourceCreatorAction grant a user's resource creator action.
func (a authorizer) GrantResourceCreatorAction(kt *kit.Kit, opts *client.GrantResourceCreatorActionOption) error {
	req := pbas.PbGrantResourceCreatorActionOption(opts)
	_, err := a.authClient.GrantResourceCreatorAction(kt.RpcCtx(), req)
	return err
}

// LogOut handler will build login url, client should make redirect
func (a authorizer) LogOut(r *http.Request) *rest.UnauthorizedData {
	loginURL, loginPlainURL := a.authLoginClient.BuildLoginURL(r)
	return &rest.UnauthorizedData{LoginURL: loginURL, LoginPlainURL: loginPlainURL}
}

// HasBiz 业务是否存在
func (a authorizer) HasBiz(bizID uint32) bool {
	return a.spaceMgr.HasCMDBSpace(strconv.Itoa(int(bizID)))
}
