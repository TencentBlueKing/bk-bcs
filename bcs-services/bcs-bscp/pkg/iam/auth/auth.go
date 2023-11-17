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

	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"k8s.io/klog/v2"

	"bscp.io/pkg/cc"
	"bscp.io/pkg/components/bkpaas"
	"bscp.io/pkg/criteria/errf"
	"bscp.io/pkg/iam/client"
	"bscp.io/pkg/iam/meta"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	pbas "bscp.io/pkg/protocol/auth-server"
	"bscp.io/pkg/rest"
	"bscp.io/pkg/runtime/gwparser"
	"bscp.io/pkg/serviced"
	"bscp.io/pkg/tools"
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
	resp, err := authClient.GetAuthLoginConf(context.Background(), &pbas.GetAuthLoginConfReq{})
	if err != nil {
		return nil, errors.Wrap(err, "get authlogin conf")
	}

	conf := &cc.LoginAuthSettings{
		Host:      resp.Host,
		InnerHost: resp.InnerHost,
		Provider:  resp.Provider,
	}
	authLoginClient := bkpaas.NewAuthLoginClient(conf)
	klog.InfoS("init authlogin client done", "host", conf.Host, "inner_host", conf.InnerHost, "provider", conf.Provider)

	authz := &authorizer{
		authClient:      authClient,
		authLoginClient: authLoginClient,
		gwParser:        nil,
	}

	// 如果有公钥，初始化配置
	if resp.GwPubkey != "" {
		gwParser, err := gwparser.NewJWTParser(resp.GwPubkey)
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
		return errf.New(errf.DoAuthorizeFailed, "authorize failed")
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
		return errf.New(errf.DoAuthorizeFailed, "get permission to apply failed")
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
		return errf.New(errf.PermissionDenied, "grpc status with details failed")
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
