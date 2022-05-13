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
 *
 */

package user

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/discovery"

	"github.com/micro/go-micro/v2/registry"
	"github.com/parnurzeal/gorequest"
)

const (
	usermanagerPrefixV1 = types.BCS_MODULE_USERMANAGER + "/v1"
	usermanagerPrefixV2 = types.BCS_MODULE_USERMANAGER + "/v2"
)

// UserManager http API definition
type UserManager interface {
	// CreateUserToken create user token and return token
	CreateUserToken(user CreateTokenReq) (string, error)
	// GetUserToken get user token
	GetUserToken(user string) (string, error)
	// DeleteUserToken delete user token
	DeleteUserToken(token string) error
	// GrantUserPermission grant user permission
	GrantUserPermission(permissions []types.Permission) error
	// RevokeUserPermission revoke user permission
	RevokeUserPermission(permissions []types.Permission) error
	// VerifyUserPermission verify user permission
	VerifyUserPermission(permReq VerifyPermissionReq) (bool, error)
}

var (
	// errNotInited err server not init
	errNotInited = errors.New("server not init")
)

const (
	defaultTimeOut = time.Second * 60
)

// userManagerClient user-manager client
var userManagerClient *UserManagerClient

// SetUserManagerClient set global user-manager client
func SetUserManagerClient(opts *Options) {
	userManagerClient = NewUserManagerClient(opts)
}

// GetUserManagerClient get user-manager client
func GetUserManagerClient() *UserManagerClient {
	return userManagerClient
}

// UserManagerClient client for usermanager
type UserManagerClient struct {
	opts      *Options
	discovery *discovery.ModuleDiscovery
	ctx       context.Context
	cancel    context.CancelFunc
}

// Options for init clusterManager
type Options struct {
	Enable bool
	// GateWay address
	GateWay         string
	IsVerifyTLS     bool
	Token           string
	Module          string
	EtcdRegistry    registry.Registry
	ClientTLSConfig *tls.Config
}

func (o *Options) validate() bool {
	if o == nil {
		return false
	}
	if !o.Enable {
		return false
	}

	if o.Module == "" {
		o.Module = ModuleUserManager
	}

	return true
}

// NewUserManagerClient init user manager and start discovery module(usermanager)
func NewUserManagerClient(opts *Options) *UserManagerClient {
	ok := opts.validate()
	if !ok {
		return nil
	}

	userClient := &UserManagerClient{
		opts: opts,
	}
	userClient.ctx, userClient.cancel = context.WithCancel(context.Background())

	if len(opts.GateWay) == 0 {
		userClient.discovery = discovery.NewModuleDiscovery(opts.Module, opts.EtcdRegistry)
		err := userClient.discovery.Start()
		if err != nil {
			blog.Errorf("start discovery client failed: %v", err)
			return nil
		}
	}

	return userClient
}

func (um *UserManagerClient) getUserManagerServerPath(url string) (string, error) {
	// call server by gateway
	if len(um.opts.GateWay) != 0 {
		return um.opts.GateWay + url, nil
	}

	// get bcs-user-manager server from etcd registry
	node, err := um.discovery.GetRandomServiceNode()
	if err != nil {
		blog.Errorf("module[%s] GetRandomServiceInstance failed: %v", um.opts.Module, err)
		return "", err
	}
	blog.V(4).Infof("get random user-manager instance [%s] from etcd registry successful", node.Address)

	if um.opts.IsVerifyTLS {
		return fmt.Sprintf("https://%s/%s", node.Address, url), nil
	}

	return fmt.Sprintf("http://%s/%s", node.Address, url), nil
}

// CreateUserToken create user token and return token
func (um *UserManagerClient) CreateUserToken(user CreateTokenReq) (string, error) {
	if um == nil {
		return "", errNotInited
	}

	var (
		_    = "CreateUserToken"
		path = usermanagerPrefixV1 + "/tokens"
		resp = &CreateTokenResp{}
	)

	url, err := um.getUserManagerServerPath(path)
	if err != nil {
		return "", err
	}

	result, body, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Post(url).
		TLSClientConfig(um.opts.ClientTLSConfig).
		Set("Authorization", fmt.Sprintf("Bearer %s", um.opts.Token)).
		Set("Content-Type", "application/json").
		Set("Connection", "close").
		SetDebug(true).
		Send(&user).
		EndStruct(resp)

	if len(errs) > 0 {
		blog.Errorf("call api CreateUserToken failed: %v", errs[0])
		return "", errs[0]
	}

	if resp.Code != 0 || !resp.Result {
		errMsg := fmt.Errorf("call CreateUserToken API error: code[%v], body[%v], err[%s]",
			result.StatusCode, string(body), resp.Message)
		return "", errMsg
	}

	return resp.Data.Token, nil
}

// DeleteUserToken delete user token
func (um *UserManagerClient) DeleteUserToken(token string) error {
	if um == nil {
		return errNotInited
	}

	var (
		_    = "DeleteUserToken"
		path = fmt.Sprintf(usermanagerPrefixV1+"/tokens/%s", token)
		resp = &GetTokenResp{}
	)

	url, err := um.getUserManagerServerPath(path)
	if err != nil {
		return err
	}

	result, body, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Delete(url).
		TLSClientConfig(um.opts.ClientTLSConfig).
		Set("Authorization", fmt.Sprintf("Bearer %s", um.opts.Token)).
		Set("Content-Type", "application/json").
		Set("Connection", "close").
		SetDebug(true).
		EndStruct(resp)

	if len(errs) > 0 {
		blog.Errorf("call api GetUserToken failed: %v", errs[0])
		return errs[0]
	}

	if resp.Code != 0 || !resp.Result {
		errMsg := fmt.Errorf("call GetUserToken API error: code[%v], body[%v], err[%s]",
			result.StatusCode, string(body), resp.Message)
		return errMsg
	}

	return nil
}

// GetUserToken get user token and return token
func (um *UserManagerClient) GetUserToken(user string) (string, error) {
	if um == nil {
		return "", errNotInited
	}

	var (
		_    = "GetUserToken"
		path = fmt.Sprintf(usermanagerPrefixV1+"/users/%s/tokens", user)
		resp = &GetTokenResp{}
	)

	url, err := um.getUserManagerServerPath(path)
	if err != nil {
		return "", err
	}

	result, body, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Get(url).
		TLSClientConfig(um.opts.ClientTLSConfig).
		Set("Authorization", fmt.Sprintf("Bearer %s", um.opts.Token)).
		Set("Content-Type", "application/json").
		Set("Connection", "close").
		SetDebug(true).
		EndStruct(resp)

	if len(errs) > 0 {
		blog.Errorf("call api GetUserToken failed: %v", errs[0])
		return "", errs[0]
	}

	if resp.Code != 0 || !resp.Result {
		errMsg := fmt.Errorf("call GetUserToken API error: code[%v], body[%v], err[%s]",
			result.StatusCode, string(body), resp.Message)
		return "", errMsg
	}

	// user not exist token
	if len(resp.Data) == 0 {
		return "", nil
	}

	return resp.Data[0].Token, nil
}

// GrantUserPermission grant user permission
func (um *UserManagerClient) GrantUserPermission(permissions []types.Permission) error {
	if um == nil {
		return errNotInited
	}

	var (
		_    = "GrantUserPermission"
		path = usermanagerPrefixV1 + "/permissions"
		resp = &CommonResp{}
	)

	url, err := um.getUserManagerServerPath(path)
	if err != nil {
		return err
	}

	perm := buildUserPermission(permissions)
	result, body, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Post(url).
		TLSClientConfig(um.opts.ClientTLSConfig).
		Set("Authorization", fmt.Sprintf("Bearer %s", um.opts.Token)).
		Set("Content-Type", "application/json").
		Set("Connection", "close").
		SetDebug(true).
		Send(&perm).
		EndStruct(resp)

	if len(errs) > 0 {
		blog.Errorf("call api GrantUserPermission failed: %v", errs[0])
		return errs[0]
	}

	if resp.Code != 0 || !resp.Result {
		errMsg := fmt.Errorf("call GrantUserPermission API error: code[%v], body[%v], err[%s]",
			result.StatusCode, string(body), resp.Message)
		return errMsg
	}

	return nil
}

// RevokeUserPermission revoke user permission
func (um *UserManagerClient) RevokeUserPermission(permissions []types.Permission) error {
	if um == nil {
		return errNotInited
	}

	var (
		_    = "RevokeUserPermission"
		path = usermanagerPrefixV1 + "/permissions"
		resp = &CommonResp{}
	)

	url, err := um.getUserManagerServerPath(path)
	if err != nil {
		return err
	}

	perm := buildUserPermission(permissions)
	result, body, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Delete(url).
		TLSClientConfig(um.opts.ClientTLSConfig).
		Set("Authorization", fmt.Sprintf("Bearer %s", um.opts.Token)).
		Set("Content-Type", "application/json").
		Set("Connection", "close").
		SetDebug(true).
		Send(&perm).
		EndStruct(resp)

	if len(errs) > 0 {
		blog.Errorf("call api GrantUserPermission failed: %v", errs[0])
		return errs[0]
	}

	if resp.Code != 0 || !resp.Result {
		errMsg := fmt.Errorf("call GrantUserPermission API error: code[%v], body[%v], err[%s]",
			result.StatusCode, string(body), resp.Message)
		return errMsg
	}

	return nil
}

// VerifyUserPermission verify user permission
func (um *UserManagerClient) VerifyUserPermission(permReq VerifyPermissionReq) (bool, error) {
	if um == nil {
		return false, errNotInited
	}

	var (
		_    = "VerifyUserPermission"
		path = usermanagerPrefixV2 + "/permissions/verify"
		resp = &VerifyPermissionResponse{}
	)

	url, err := um.getUserManagerServerPath(path)
	if err != nil {
		return false, err
	}

	result, body, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Get(url).
		TLSClientConfig(um.opts.ClientTLSConfig).
		Set("Authorization", fmt.Sprintf("Bearer %s", um.opts.Token)).
		Set("Content-Type", "application/json").
		Set("Connection", "close").
		SetDebug(true).
		Send(&permReq).
		EndStruct(resp)

	if len(errs) > 0 {
		blog.Errorf("call api VerifyUserPermission failed: %v", errs[0])
		return false, errs[0]
	}

	if resp.Code != 0 || !resp.Result {
		errMsg := fmt.Errorf("call VerifyUserPermission API error: code[%v], body[%v], err[%s]",
			result.StatusCode, string(body), resp.Message)
		return false, errMsg
	}

	return resp.Data.Allowed, nil
}

// Stop stop UserManagerClient
func (um *UserManagerClient) Stop() {
	if um == nil {
		return
	}
	if um.discovery != nil {
		um.discovery.Stop()
	}
	um.cancel()
}
