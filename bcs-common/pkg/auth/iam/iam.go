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

// Package iam xxx
package iam

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/TencentBlueKing/iam-go-sdk"
	"github.com/TencentBlueKing/iam-go-sdk/logger"
	"github.com/TencentBlueKing/iam-go-sdk/metric"
	"github.com/golang-migrate/migrate/v4/source"
	"github.com/parnurzeal/gorequest"
	"github.com/sirupsen/logrus"
	"k8s.io/klog/v2"
)

// PermClient interface for IAM backend client
type PermClient interface {
	IsAllowedWithoutResource(actionID string, request PermissionRequest, cache bool) (bool, error)
	IsAllowedWithResource(actionID string, request PermissionRequest, nodes []ResourceNode, cache bool) (bool, error)
	BatchResourceIsAllowed(actionID string, request PermissionRequest, nodes [][]ResourceNode) (map[string]bool, error)
	MultiActionsAllowedWithoutResource(actions []string, request PermissionRequest) (map[string]bool, error)
	ResourceMultiActionsAllowed(actions []string, request PermissionRequest, nodes []ResourceNode) (map[string]bool, error)
	BatchResourceMultiActionsAllowed(actions []string, request PermissionRequest,
		nodes [][]ResourceNode) (map[string]map[string]bool, error)
	GetToken() (string, error)
	IsBasicAuthAllowed(user BkUser) error
	GetApplyURL(request ApplicationRequest, relatedResources []ApplicationAction, user BkUser) (string, error)
	// CreateGradeManagers xxx
	// perm management API
	CreateGradeManagers(ctx context.Context, request GradeManagerRequest) (uint64, error)
	CreateUserGroup(ctx context.Context, gradeManagerID uint64, request CreateUserGroupRequest) ([]uint64, error)
	DeleteUserGroup(ctx context.Context, groupID uint64) error
	AddUserGroupMembers(ctx context.Context, groupID uint64, request AddGroupMemberRequest) error
	DeleteUserGroupMembers(ctx context.Context, groupID uint64, request DeleteGroupMemberRequest) error
	CreateUserGroupPolicies(ctx context.Context, groupID uint64, request AuthorizationScope) error
	AuthResourceCreatorPerm(ctx context.Context, resource ResourceCreator, ancestors []Ancestor) error
}

// PermMigrateClient interface for IAM backend client
type PermMigrateClient interface {
	PermClient
	// Migrate xxx
	Migrate(db *sql.DB, driver source.Driver, migrateTable string, timeout time.Duration,
		templateVar interface{}) error
}

var (
	defaultTimeOut = time.Second * 60
)

// Options for init IAM client
type Options struct {
	// SystemID that bk_bcs used in auth center
	SystemID string
	// AppCode is code for authorize call iam
	AppCode string
	// AppSecret is secret for authorize call iam
	AppSecret string
	// External is false, use GateWayHost
	External bool
	// GateWay host
	GateWayHost string
	// IAM host
	IAMHost string
	// BkiIAM host
	BkiIAMHost string
	// Metrics
	Metric bool
	// Debug
	Debug bool
}

func (opt *Options) validate() error {
	if opt == nil {
		return ErrServerNotInit
	}

	if opt.SystemID == "" || opt.AppCode == "" || opt.AppSecret == "" {
		return fmt.Errorf("systemID/AppCode/AppSecret required")
	}

	if !opt.External && opt.GateWayHost == "" {
		return fmt.Errorf("BKAPIGatewayHost required when UseGateway flag set to true")
	}
	if opt.External && (opt.BkiIAMHost == "" || opt.IAMHost == "") {
		return fmt.Errorf("BKIAMHost and BKPAASHost required when UseGateway flag set to false")
	}

	return nil
}

type iamClient struct {
	cli *iam.IAM
	opt *Options
}

func setIAMLogger(opt *Options) {
	defaultLogLevel := logrus.ErrorLevel
	if opt.Debug {
		defaultLogLevel = logrus.DebugLevel
	}

	log := &logrus.Logger{
		Out:          os.Stderr,
		Formatter:    new(logrus.TextFormatter),
		Hooks:        make(logrus.LevelHooks),
		Level:        defaultLogLevel,
		ExitFunc:     os.Exit,
		ReportCaller: false,
	}

	logger.SetLogger(log)
}

// NewIamClient create iam backend client
func NewIamClient(opt *Options) (PermClient, error) {
	err := opt.validate()
	if err != nil {
		return nil, fmt.Errorf("NewIamClient options invalid: %v", err)
	}

	// register interface metric
	if opt.Metric {
		metric.RegisterMetrics()
	}

	client := &iamClient{
		opt: opt,
	}

	if opt.External {
		// true directCAll + ESB API
		client.cli = iam.NewIAM(opt.SystemID, opt.AppCode, opt.AppSecret, opt.IAMHost, opt.BkiIAMHost)
	} else {
		// false APIGW
		client.cli = iam.NewAPIGatewayIAM(opt.SystemID, opt.AppCode, opt.AppSecret, opt.GateWayHost)
	}

	// init IAM logger
	setIAMLogger(opt)

	return client, nil
}

// NewIamMigrateClient create iam backend client
func NewIamMigrateClient(opt *Options) (PermMigrateClient, error) {
	err := opt.validate()
	if err != nil {
		return nil, fmt.Errorf("NewIamMigrateClient options invalid: %v", err)
	}

	// register interface metric
	if opt.Metric {
		metric.RegisterMetrics()
	}

	client := &iamClient{
		opt: opt,
	}

	if opt.External {
		// true directCAll + ESB API
		client.cli = iam.NewIAM(opt.SystemID, opt.AppCode, opt.AppSecret, opt.IAMHost, opt.BkiIAMHost)
	} else {
		// false APIGW
		client.cli = iam.NewAPIGatewayIAM(opt.SystemID, opt.AppCode, opt.AppSecret, opt.GateWayHost)
	}

	// init IAM logger
	setIAMLogger(opt)

	return client, nil
}

// BkUser user/token
type BkUser struct {
	BkToken    string
	BkUserName string
}

func (ic *iamClient) generateGateWayAuth(bkUserName string) (string, error) { // nolint
	if ic == nil {
		return "", ErrServerNotInit
	}

	auth := &AuthInfo{
		BkAppCode:   ic.opt.AppCode,
		BkAppSecret: ic.opt.AppSecret,
	}
	if len(bkUserName) > 0 {
		auth.BkUserName = bkUserName
	}

	userAuth, err := json.Marshal(auth)
	if err != nil {
		return "", err
	}

	return string(userAuth), nil
}

// IsAllowedWithoutResource query signal action permission without resource(cache use withoutResource or managerPerm)
func (ic *iamClient) IsAllowedWithoutResource(actionID string, request PermissionRequest, cache bool) (bool, error) {
	if ic == nil {
		return false, ErrServerNotInit
	}

	req := request.MakeRequestWithoutResources(actionID)
	if cache {
		return ic.cli.IsAllowedWithCache(req, defaultAllowTTL)
	}
	return ic.cli.IsAllowed(req)
}

// IsAllowedWithResource query signal action signal resource permission(cache use withoutResource or managerPerm)
func (ic *iamClient) IsAllowedWithResource(actionID string, request PermissionRequest, nodes []ResourceNode,
	cache bool) (bool, error) {
	if ic == nil {
		return false, ErrServerNotInit
	}

	req := request.MakeRequestWithResources(actionID, nodes)
	if cache {
		return ic.cli.IsAllowedWithCache(req, defaultAllowTTL)
	}
	return ic.cli.IsAllowed(req)
}

// BatchResourceIsAllowed batch resource check permission, signalAction multiResources
// resources []iam.ResourceNode: len=1, return node.ID; len > 1, node.Type:node.ID/node.Type:node.ID
func (ic *iamClient) BatchResourceIsAllowed(actionID string, request PermissionRequest,
	nodes [][]ResourceNode) (map[string]bool, error) {
	if ic == nil {
		return nil, ErrServerNotInit
	}

	req := request.MakeRequestWithoutResources(actionID)

	resourceList := make([]iam.Resources, 0)
	for _, nodeList := range nodes {
		iamNodes := make([]iam.ResourceNode, 0)
		for i := range nodeList {
			iamNodes = append(iamNodes, nodeList[i].BuildResourceNode())
		}

		resourceList = append(resourceList, iamNodes)
	}

	return ic.cli.BatchIsAllowed(req, resourceList)
}

// MultiActionsAllowedWithoutResource for multiActions without resource
func (ic *iamClient) MultiActionsAllowedWithoutResource(actions []string, request PermissionRequest) (
	map[string]bool, error) {
	if ic == nil {
		return nil, ErrServerNotInit
	}

	req := request.MakeReqMultiActionsWithoutRes(actions)
	return ic.cli.ResourceMultiActionsAllowed(req)
}

// ResourceMultiActionsAllowed for multiActions signalResource
func (ic *iamClient) ResourceMultiActionsAllowed(actions []string, request PermissionRequest,
	nodes []ResourceNode) (map[string]bool, error) {
	if ic == nil {
		return nil, ErrServerNotInit
	}

	req := request.MakeRequestMultiActionResources(actions, nodes)
	return ic.cli.ResourceMultiActionsAllowed(req)
}

// BatchResourceMultiActionsAllowed will check the permissions of batch-resource with multi-actions,
// multi actions and multi resource
func (ic *iamClient) BatchResourceMultiActionsAllowed(actions []string, request PermissionRequest,
	nodes [][]ResourceNode) (map[string]map[string]bool, error) {
	if ic == nil {
		return nil, ErrServerNotInit
	}

	multiReq := request.MakeRequestMultiActionResources(actions, nil)

	resourceList := make([]iam.Resources, 0)
	for _, nodeList := range nodes {
		iamNodes := make([]iam.ResourceNode, 0)
		for i := range nodeList {
			iamNodes = append(iamNodes, nodeList[i].BuildResourceNode())
		}

		resourceList = append(resourceList, iamNodes)
	}

	return ic.cli.BatchResourceMultiActionsAllowed(multiReq, resourceList)
}

// GetToken will get the token of system
func (ic *iamClient) GetToken() (string, error) {
	if ic == nil {
		return "", ErrServerNotInit
	}

	return ic.cli.GetToken()
}

// IsBasicAuthAllowed xxx
// check iam callback request auth
func (ic *iamClient) IsBasicAuthAllowed(user BkUser) error {
	if ic == nil {
		return ErrServerNotInit
	}

	return ic.cli.IsBasicAuthAllowed(user.BkUserName, user.BkToken)
}

// GetApplyURL will generate the application URL
func (ic *iamClient) GetApplyURL(request ApplicationRequest, relatedResources []ApplicationAction, user BkUser) (string,
	error) {
	if ic == nil {
		return "", ErrServerNotInit
	}

	application := request.BuildApplication(relatedResources)

	url, err := ic.cli.GetApplyURL(application, user.BkToken, user.BkUserName)
	if err != nil {
		klog.Errorf("iam generate apply url failed: %s", err)
		return IamAppURL, err
	}

	return url, nil
}

// CreateGradeManagers 分级管理员相关接口
// CreateGradeManagers create gradeManagers
func (ic *iamClient) CreateGradeManagers(ctx context.Context, request GradeManagerRequest) (uint64, error) {
	if ic == nil {
		return 0, ErrServerNotInit
	}

	var (
		_    = "CreateGradeManagers"
		path = "/api/v1/open/management/grade_managers/"
	)

	var (
		url  = ic.opt.GateWayHost + path
		resp = &GradeManagerResponse{}
	)

	auth, err := ic.generateGateWayAuth("")
	if err != nil {
		klog.Errorf("CreateGradeManagers generateGateWayAuth failed: %v", err)
		return 0, err
	}

	result, body, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Post(url).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", auth).
		SetDebug(true).
		Send(&request).
		EndStruct(resp)

	if len(errs) != 0 {
		klog.Errorf("CreateGradeManagers gorequest errors=`%s`", errs)
		return 0, errs[0]
	}
	if result.StatusCode != http.StatusOK || resp.Code != 0 {
		errMsg := fmt.Errorf("CreateGradeManagers API error: code[%v], body[%v], err[%s]",
			result.StatusCode, string(body), resp.Message)
		return 0, errMsg
	}

	klog.Infof("CreateGradeManagers[%s:%s] successful", request.System, request.Name)
	return resp.Data.ID, nil
}

// UserGroup 用户组相关接口

// CreateUserGroup xxx
// CreateGradeManagers create gradeManagers
func (ic *iamClient) CreateUserGroup(ctx context.Context, gradeManagerID uint64,
	request CreateUserGroupRequest) ([]uint64, error) {
	if ic == nil {
		return nil, ErrServerNotInit
	}

	var (
		_    = "CreateUserGroup"
		path = fmt.Sprintf("/api/v1/open/management/grade_managers/%v/groups/", gradeManagerID)
	)

	var (
		url  = ic.opt.GateWayHost + path
		resp = &CreateUserGroupResponse{}
	)

	auth, err := ic.generateGateWayAuth("")
	if err != nil {
		klog.Errorf("CreateUserGroup generateGateWayAuth failed: %v", err)
		return nil, err
	}

	result, body, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Post(url).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", auth).
		SetDebug(true).
		Send(&request).
		EndStruct(resp)

	if len(errs) != 0 {
		klog.Errorf("CreateUserGroup gorequest errors=`%s`", errs)
		return nil, errs[0]
	}
	if result.StatusCode != http.StatusOK || resp.Code != 0 {
		errMsg := fmt.Errorf("CreateUserGroup API error: code[%v], body[%v], err[%s]",
			result.StatusCode, string(body), resp.Message)
		return nil, errMsg
	}

	klog.Infof("CreateUserGroup[%s:%v] successful", ic.opt.SystemID, gradeManagerID)
	return resp.Data, nil
}

// DeleteUserGroup delete userGroup
func (ic *iamClient) DeleteUserGroup(ctx context.Context, groupID uint64) error {
	if ic == nil {
		return ErrServerNotInit
	}

	var (
		_    = "DeleteUserGroup"
		path = fmt.Sprintf("/api/v1/open/management/groups/%v/", groupID)
	)

	var (
		url  = ic.opt.GateWayHost + path
		resp = &BaseResponse{}
	)

	auth, err := ic.generateGateWayAuth("")
	if err != nil {
		klog.Errorf("DeleteUserGroup generateGateWayAuth failed: %v", err)
		return err
	}

	result, body, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Delete(url).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", auth).
		SetDebug(true).
		EndStruct(resp)

	if len(errs) != 0 {
		klog.Errorf("DeleteUserGroup gorequest errors=`%s`", errs)
		return errs[0]
	}
	if result.StatusCode != http.StatusOK || resp.Code != 0 {
		errMsg := fmt.Errorf("DeleteUserGroup API error: code[%v], body[%v], err[%s]",
			result.StatusCode, string(body), resp.Message)
		return errMsg
	}

	klog.Infof("DeleteUserGroup[%s:%v] successful", ic.opt.SystemID, groupID)
	return nil
}

// AddUserGroupMembers add user group members
func (ic *iamClient) AddUserGroupMembers(ctx context.Context, groupID uint64, request AddGroupMemberRequest) error {
	if ic == nil {
		return ErrServerNotInit
	}

	var (
		_    = "AddUserGroupMembers"
		path = fmt.Sprintf("/api/v1/open/management/groups/%v/members/", groupID)
	)

	var (
		url  = ic.opt.GateWayHost + path
		resp = &AddGroupMemberResponse{}
	)

	auth, err := ic.generateGateWayAuth("")
	if err != nil {
		klog.Errorf("AddUserGroupMembers generateGateWayAuth failed: %v", err)
		return err
	}

	result, body, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Post(url).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", auth).
		SetDebug(true).
		Send(&request).
		EndStruct(resp)

	if len(errs) != 0 {
		klog.Errorf("AddUserGroupMembers gorequest errors=`%s`", errs)
		return errs[0]
	}
	if result.StatusCode != http.StatusOK || resp.Code != 0 {
		errMsg := fmt.Errorf("AddUserGroupMembers API error: code[%v], body[%v], err[%s]",
			result.StatusCode, string(body), resp.Message)
		return errMsg
	}

	klog.Infof("AddUserGroupMembers[%s:%v] successful", ic.opt.SystemID, groupID)
	return nil
}

// DeleteUserGroupMembers delete user group members
func (ic *iamClient) DeleteUserGroupMembers(ctx context.Context, groupID uint64,
	request DeleteGroupMemberRequest) error {
	if ic == nil {
		return ErrServerNotInit
	}

	var (
		_    = "DeleteUserGroupMembers"
		path = fmt.Sprintf("/api/v1/open/management/groups/%v/members/", groupID)
	)

	var (
		url  = ic.opt.GateWayHost + path
		resp = &BaseResponse{}
	)

	auth, err := ic.generateGateWayAuth("")
	if err != nil {
		klog.Errorf("DeleteUserGroupMembers generateGateWayAuth failed: %v", err)
		return err
	}
	if request.Type == "" {
		request.Type = string(User)
	}
	if len(request.IDs) == 0 {
		return fmt.Errorf("DeleteUserGroupMembers paras IDs empty")
	}

	result, _, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Delete(url).
		Query(fmt.Sprintf("type=%s", request.Type)).
		Query(fmt.Sprintf("ids=%s", strings.Join(request.IDs, ","))).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", auth).
		SetDebug(true).
		EndStruct(resp)

	if len(errs) != 0 {
		klog.Errorf("DeleteUserGroupMembers gorequest errors=`%s`", errs)
		return errs[0]
	}
	if result.StatusCode != http.StatusOK || resp.Code != 0 {
		errMsg := fmt.Errorf("DeleteUserGroupMembers API error: code[%v], err[%s]",
			result.StatusCode, resp.Message)
		return errMsg
	}

	klog.Infof("DeleteUserGroupMembers[%s:%v] successful", ic.opt.SystemID, groupID)
	return nil
}

// CreateUserGroupPolicies create group policies
func (ic *iamClient) CreateUserGroupPolicies(ctx context.Context, groupID uint64, request AuthorizationScope) error {
	if ic == nil {
		return ErrServerNotInit
	}

	var (
		_    = "CreateUserGroupPolicies"
		path = fmt.Sprintf("/api/v1/open/management/groups/%v/policies/", groupID)
	)

	var (
		url  = ic.opt.GateWayHost + path
		resp = &BaseResponse{}
	)

	auth, err := ic.generateGateWayAuth("")
	if err != nil {
		klog.Errorf("CreateUserGroupPolicies generateGateWayAuth failed: %v", err)
		return err
	}

	result, body, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Post(url).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", auth).
		SetDebug(true).
		Send(&request).
		EndStruct(resp)

	if len(errs) != 0 {
		klog.Errorf("CreateUserGroupPolicies gorequest errors=`%s`", errs)
		return errs[0]
	}
	if result.StatusCode != http.StatusOK || resp.Code != 0 {
		errMsg := fmt.Errorf("CreateUserGroupPolicies API error: code[%v], body[%v], err[%s]",
			result.StatusCode, string(body), resp.Message)
		return errMsg
	}

	klog.Infof("CreateUserGroupPolicies[%s:%s] successful", request.System, groupID)
	return nil
}

// AuthResourceCreatorPerm authorize creator resource perm
func (ic *iamClient) AuthResourceCreatorPerm(ctx context.Context, resource ResourceCreator,
	ancestors []Ancestor) error {
	var (
		_    = "AuthResourceCreatorPerm"
		path = "/api/v1/open/authorization/resource_creator_action/"
	)

	var (
		url  = ic.opt.GateWayHost + path
		resp = &ResourceCreatorActionResponse{}
	)

	auth, err := ic.generateGateWayAuth("")
	if err != nil {
		klog.Errorf("AuthResourceCreatorPerm generateGateWayAuth failed: %v", err)
		return err
	}

	request := buildResourceCreatorActionRequest(resource, ancestors)

	result, body, errs := gorequest.New().
		Timeout(defaultTimeOut).
		Post(url).
		Set("Content-Type", "application/json").
		Set("Accept", "application/json").
		Set("X-Bkapi-Authorization", auth).
		SetDebug(true).
		Send(&request).
		EndStruct(resp)

	if len(errs) != 0 {
		klog.Errorf("AuthResourceCreatorPerm gorequest errors=`%s`", errs)
		return errs[0]
	}
	if result.StatusCode != http.StatusOK || resp.Code != 0 {
		errMsg := fmt.Errorf("AuthResourceCreatorPerm API error: code[%v], body[%v], err[%s]",
			result.StatusCode, string(body), resp.Message)
		return errMsg
	}

	klog.Infof("AuthResourceCreatorPerm[%s:%s] successful[%+v]", request.System, resource.Creator, resp.Data)

	return nil
}

// Migrate migrate iam db
func (ic *iamClient) Migrate(db *sql.DB, driver source.Driver, migrateTable string, timeout time.Duration,
	templateVar interface{}) error {
	return ic.cli.Migrate(db, driver, migrateTable, timeout, templateVar)
}
