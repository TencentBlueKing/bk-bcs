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

package iam

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/metrics"

	"github.com/parnurzeal/gorequest"
)

const (
	codeNotFound = 1901404
)

var (
	// ErrNotFound not found
	ErrNotFound = errors.New("Not Found")
)

var (
	// ErrInitServerFail server error
	ErrInitServerFail = errors.New("init server failed")
	// ErrInvalidateOptions options error
	ErrInvalidateOptions = errors.New("options is nil")
	// ErrInvalidateServer server address error
	ErrInvalidateServer = errors.New("invalidate server address")
	// ErrInvalidateAuth appCode|appSecret error
	ErrInvalidateAuth = errors.New("invalidate appCode | appSecret")
)

// AuthConfig options for model server
type AuthConfig struct {
	Server string `json:"server"`

	SystemID  string `json:"systemID"`
	AppCode   string `json:"appCode"`
	AppSecret string `json:"appSecret"`

	ServerDebug bool `json:"serverDebug"`
}

// ValidateAuthConfig verify AuthConfig valid
func (config *AuthConfig) ValidateAuthConfig() error {
	if config == nil {
		return ErrInvalidateOptions
	}

	if !strings.HasPrefix(config.Server, "http") && !strings.HasPrefix(config.Server, "https") {
		return ErrInvalidateServer
	}

	if len(config.AppCode) == 0 || len(config.AppSecret) == 0 {
		return ErrInvalidateAuth
	}

	return nil
}

func newIamModelServer(config *AuthConfig) *iamModelServer {
	err := config.ValidateAuthConfig()
	if err != nil {
		blog.Errorf("NewIamModelServer validateAuthConfig failed: %v", err)
		return nil
	}

	modelServer := &iamModelServer{
		server:      config.Server,
		appCode:     config.AppCode,
		appSecret:   config.AppSecret,
		systemID:    config.SystemID,
		serverDebug: config.ServerDebug,
	}
	gateWayAuth, err := modelServer.generateGateWayAuth()
	if err != nil {
		blog.Errorf("NewIamModelServer generateGateWayAuth failed: %v", err)
		return nil
	}

	modelServer.gateWayAuth = gateWayAuth

	return modelServer
}

type iamModelServer struct {
	server      string
	appCode     string
	appSecret   string
	systemID    string
	gateWayAuth string

	testDebug   bool
	serverDebug bool
}

// AuthInfo auth
type AuthInfo struct {
	BkAppCode   string `json:"bk_app_code"`
	BkAppSecret string `json:"bk_app_secret"`
}

func (s *iamModelServer) generateGateWayAuth() (string, error) {
	if s == nil {
		return "", ErrServerNotInit
	}
	auth := &AuthInfo{
		BkAppCode:   s.appCode,
		BkAppSecret: s.appSecret,
	}
	userAuth, err := json.Marshal(auth)
	if err != nil {
		return "", err
	}

	return string(userAuth), nil
}

// curl -H 'X-Bkapi-Authorization: {"bk_app_code": "x", "bk_app_secret": "y"}' 'http://bk-iam.apigw.o.oa.com/stage/api/v1/XXXXX'
func (s *iamModelServer) RegisterSystem(timeout time.Duration, sys System) error {
	if s == nil {
		return ErrInitServerFail
	}
	const (
		apiName = "RegisterSystem"
		path    = "/api/v1/model/systems"
	)

	var (
		url      = s.server + path
		respData = &BaseResponse{}
		start    = time.Now()
	)

	resp, _, errs := gorequest.New().
		Timeout(timeout).
		Post(url).
		Set("Content-Type", "application/json").
		Set("Connection", "close").
		Set("X-Bk-App-Code", s.appCode).
		Set("X-Bk-App-Secret", s.appSecret).
		Set("X-Bkapi-Authorization", s.gateWayAuth).
		SetDebug(s.serverDebug).
		Send(sys).
		EndStruct(respData)
	if len(errs) > 0 {
		blog.Errorf("call api RegisterSystem failed: %v", errs[0])
		if !s.testDebug {
			metrics.ReportRequestAPIMetrics(apiName, http.MethodPost, metrics.ErrStatus, start)
		}
		return errs[0]
	}

	if respData.Code != 0 {
		if !s.testDebug {
			metrics.ReportRequestAPIMetrics(apiName, http.MethodPost, fmt.Sprintf("%d", respData.Code), start)
		}

		return &AuthError{
			RequestID: resp.Header.Get(IamRequestHeader),
			Reason:    fmt.Errorf("register system failed, code[%d], msg[%s]", respData.Code, respData.Message),
		}
	}

	if !s.testDebug {
		metrics.ReportRequestAPIMetrics(apiName, http.MethodPost, fmt.Sprintf("%d", respData.Code), start)
	}

	return nil
}

func (s *iamModelServer) UpdateSystemConfig(timeout time.Duration, config *SysConfig) error {
	if s == nil {
		return ErrInitServerFail
	}
	const (
		apiName = "UpdateSystemConfig"
		path    = "/api/v1/model/systems/%s"
	)

	system := new(System)
	system.ProviderConfig = config

	var (
		url      = s.server + fmt.Sprintf(path, s.systemID)
		respData = &BaseResponse{}
		start    = time.Now()
	)

	resp, _, errs := gorequest.New().
		Timeout(timeout).
		Put(url).
		Set("Content-Type", "application/json").
		Set("Connection", "close").
		Set("X-Bk-App-Code", s.appCode).
		Set("X-Bk-App-Secret", s.appSecret).
		Set("X-Bkapi-Authorization", s.gateWayAuth).
		SetDebug(s.serverDebug).
		Send(system).
		EndStruct(respData)
	if len(errs) > 0 {
		blog.Errorf("call api UpdateSystemConfig failed: %v", errs[0])
		if !s.testDebug {
			metrics.ReportRequestAPIMetrics(apiName, http.MethodPut, metrics.ErrStatus, start)
		}
		return errs[0]
	}

	if respData.Code != 0 {
		if !s.testDebug {
			metrics.ReportRequestAPIMetrics(apiName, http.MethodPut, fmt.Sprintf("%d", respData.Code), start)
		}

		return &AuthError{
			RequestID: resp.Header.Get(IamRequestHeader),
			Reason:    fmt.Errorf("update system failed, code[%d], msg[%s]", respData.Code, respData.Message),
		}
	}

	if !s.testDebug {
		metrics.ReportRequestAPIMetrics(apiName, http.MethodPut, fmt.Sprintf("%d", respData.Code), start)
	}

	return nil
}

func (s *iamModelServer) GetSystemInfo(timeout time.Duration) (*SystemResp, error) {
	if s == nil {
		return nil, ErrInitServerFail
	}
	const (
		apiName = "GetSystemInfo"
		path    = "/api/v1/model/systems/%s/query"
	)

	var (
		url      = s.server + fmt.Sprintf(path, s.systemID)
		respData = &SystemResp{}
		start    = time.Now()
	)

	resp, _, errs := gorequest.New().
		Timeout(timeout).
		Get(url).
		Set("Content-Type", "application/json").
		Set("Connection", "close").
		Set("X-Bk-App-Code", s.appCode).
		Set("X-Bk-App-Secret", s.appSecret).
		Set("X-Bkapi-Authorization", s.gateWayAuth).
		Param("fields", "base_info,resource_types,actions,action_groups,instance_selections,"+
			"resource_creator_actions,common_actions").
		SetDebug(s.serverDebug).
		EndStruct(respData)
	if len(errs) > 0 {
		blog.Errorf("call api GetSystemInfo failed: %v", errs[0])
		if !s.testDebug {
			metrics.ReportRequestAPIMetrics(apiName, http.MethodGet, metrics.ErrStatus, start)
		}
		return nil, errs[0]
	}

	if respData.Code != 0 {
		if !s.testDebug {
			metrics.ReportRequestAPIMetrics(apiName, http.MethodGet, fmt.Sprintf("%d", respData.Code), start)
		}

		if respData.Code == codeNotFound {
			return respData, ErrNotFound
		}

		return nil, &AuthError{
			RequestID: resp.Header.Get(IamRequestHeader),
			Reason:    fmt.Errorf("get system info failed, code[%d], msg[%s]", respData.Code, respData.Message),
		}
	}

	if !s.testDebug {
		metrics.ReportRequestAPIMetrics(apiName, http.MethodGet, fmt.Sprintf("%d", respData.Code), start)
	}

	return respData, nil
}

func (s *iamModelServer) RegisterResourceTypes(timeout time.Duration, resTypes []ResourceType) error {
	if s == nil {
		return ErrInitServerFail
	}
	const (
		apiName = "RegisterResourceTypes"
		path    = "/api/v1/model/systems/%s/resource-types"
	)

	var (
		url      = s.server + fmt.Sprintf(path, s.systemID)
		respData = &BaseResponse{}
		start    = time.Now()
	)

	resp, _, errs := gorequest.New().
		Timeout(timeout).
		Post(url).
		Set("Content-Type", "application/json").
		Set("Connection", "close").
		Set("X-Bk-App-Code", s.appCode).
		Set("X-Bk-App-Secret", s.appSecret).
		Set("X-Bkapi-Authorization", s.gateWayAuth).
		SetDebug(s.serverDebug).
		Send(resTypes).
		EndStruct(respData)
	if len(errs) > 0 {
		blog.Errorf("call api RegisterResourceTypes failed: %v", errs[0])
		if !s.testDebug {
			metrics.ReportRequestAPIMetrics(apiName, http.MethodPost, metrics.ErrStatus, start)
		}
		return errs[0]
	}

	if respData.Code != 0 {
		if !s.testDebug {
			metrics.ReportRequestAPIMetrics(apiName, http.MethodPost, fmt.Sprintf("%d", respData.Code), start)
		}

		return &AuthError{
			RequestID: resp.Header.Get(IamRequestHeader),
			Reason:    fmt.Errorf("register resourceTypes failed, code[%d], msg[%s]", respData.Code, respData.Message),
		}
	}

	if !s.testDebug {
		metrics.ReportRequestAPIMetrics(apiName, http.MethodPost, fmt.Sprintf("%d", respData.Code), start)
	}

	return nil
}

func (s *iamModelServer) UpdateResourceTypes(timeout time.Duration, resType ResourceType) error {
	if s == nil {
		return ErrInitServerFail
	}
	const (
		apiName = "UpdateResourceTypes"
		path    = "/api/v1/model/systems/%s/resource-types/%s"
	)

	var (
		url      = s.server + fmt.Sprintf(path, s.systemID, resType.ID)
		respData = &BaseResponse{}
		start    = time.Now()
	)

	resp, _, errs := gorequest.New().
		Timeout(timeout).
		Put(url).
		Set("Content-Type", "application/json").
		Set("Connection", "close").
		Set("X-Bk-App-Code", s.appCode).
		Set("X-Bk-App-Secret", s.appSecret).
		Set("X-Bkapi-Authorization", s.gateWayAuth).
		SetDebug(s.serverDebug).
		Send(resType).
		EndStruct(respData)
	if len(errs) > 0 {
		blog.Errorf("call api UpdateResourceTypes failed: %v", errs[0])
		if !s.testDebug {
			metrics.ReportRequestAPIMetrics(apiName, http.MethodPut, metrics.ErrStatus, start)
		}
		return errs[0]
	}

	if respData.Code != 0 {
		if !s.testDebug {
			metrics.ReportRequestAPIMetrics(apiName, http.MethodPut, fmt.Sprintf("%d", respData.Code), start)
		}

		return &AuthError{
			RequestID: resp.Header.Get(IamRequestHeader),
			Reason:    fmt.Errorf("update resourceTypes %s failed, code[%d], msg[%s]", resType.ID, respData.Code, respData.Message),
		}
	}

	if !s.testDebug {
		metrics.ReportRequestAPIMetrics(apiName, http.MethodPut, fmt.Sprintf("%d", respData.Code), start)
	}

	return nil
}

func (s *iamModelServer) DeleteResourceTypes(timeout time.Duration, resourceTypeIDs []TypeID) error {
	if s == nil {
		return ErrInitServerFail
	}
	const (
		apiName = "DeleteResourceTypes"
		path    = "/api/v1/model/systems/%s/resource-types"
	)

	ids := make([]struct {
		ID TypeID `json:"id"`
	}, len(resourceTypeIDs))
	for index := range resourceTypeIDs {
		ids[index].ID = resourceTypeIDs[index]
	}

	var (
		url      = s.server + fmt.Sprintf(path, s.systemID)
		respData = &BaseResponse{}
		start    = time.Now()
	)

	resp, _, errs := gorequest.New().
		Timeout(timeout).
		Delete(url).
		Set("Content-Type", "application/json").
		Set("Connection", "close").
		Set("X-Bk-App-Code", s.appCode).
		Set("X-Bk-App-Secret", s.appSecret).
		Set("X-Bkapi-Authorization", s.gateWayAuth).
		SetDebug(s.serverDebug).
		Send(ids).
		EndStruct(respData)
	if len(errs) > 0 {
		blog.Errorf("call api DeleteResourceTypes failed: %v", errs[0])
		if !s.testDebug {
			metrics.ReportRequestAPIMetrics(apiName, http.MethodDelete, metrics.ErrStatus, start)
		}
		return errs[0]
	}

	if respData.Code != 0 {
		if !s.testDebug {
			metrics.ReportRequestAPIMetrics(apiName, http.MethodDelete, fmt.Sprintf("%d", respData.Code), start)
		}

		return &AuthError{
			RequestID: resp.Header.Get(IamRequestHeader),
			Reason:    fmt.Errorf("delete resourceTypes %v failed, code[%d], msg[%s]", resourceTypeIDs, respData.Code, respData.Message),
		}
	}

	if !s.testDebug {
		metrics.ReportRequestAPIMetrics(apiName, http.MethodDelete, fmt.Sprintf("%d", respData.Code), start)
	}

	return nil
}

func (s *iamModelServer) CreateAction(timeout time.Duration, actions []ResourceAction) error {
	if s == nil {
		return ErrInitServerFail
	}
	const (
		apiName = "CreateAction"
		path    = "/api/v1/model/systems/%s/actions"
	)

	var (
		url      = s.server + fmt.Sprintf(path, s.systemID)
		respData = &BaseResponse{}
		start    = time.Now()
	)

	resp, _, errs := gorequest.New().
		Timeout(timeout).
		Post(url).
		Set("Content-Type", "application/json").
		Set("Connection", "close").
		Set("X-Bk-App-Code", s.appCode).
		Set("X-Bk-App-Secret", s.appSecret).
		Set("X-Bkapi-Authorization", s.gateWayAuth).
		SetDebug(s.serverDebug).
		Send(actions).
		EndStruct(respData)
	if len(errs) > 0 {
		blog.Errorf("call api CreateAction failed: %v", errs[0])
		if !s.testDebug {
			metrics.ReportRequestAPIMetrics(apiName, http.MethodPost, metrics.ErrStatus, start)
		}
		return errs[0]
	}

	if respData.Code != 0 {
		if !s.testDebug {
			metrics.ReportRequestAPIMetrics(apiName, http.MethodPost, fmt.Sprintf("%d", respData.Code), start)
		}

		return &AuthError{
			RequestID: resp.Header.Get(IamRequestHeader),
			Reason:    fmt.Errorf("add resource action %v failed, code[%d], msg[%s]", actions, respData.Code, respData.Message),
		}
	}

	if !s.testDebug {
		metrics.ReportRequestAPIMetrics(apiName, http.MethodPost, fmt.Sprintf("%d", respData.Code), start)
	}

	return nil
}

func (s *iamModelServer) UpdateAction(timeout time.Duration, action ResourceAction) error {
	if s == nil {
		return ErrInitServerFail
	}
	const (
		apiName = "UpdateAction"
		path    = "/api/v1/model/systems/%s/actions/%s"
	)

	var (
		url      = s.server + fmt.Sprintf(path, s.systemID, action.ID)
		respData = &BaseResponse{}
		start    = time.Now()
	)

	resp, _, errs := gorequest.New().
		Timeout(timeout).
		Put(url).
		Set("Content-Type", "application/json").
		Set("Connection", "close").
		Set("X-Bk-App-Code", s.appCode).
		Set("X-Bk-App-Secret", s.appSecret).
		Set("X-Bkapi-Authorization", s.gateWayAuth).
		SetDebug(s.serverDebug).
		Send(action).
		EndStruct(respData)
	if len(errs) > 0 {
		blog.Errorf("call api UpdateAction failed: %v", errs[0])
		if !s.testDebug {
			metrics.ReportRequestAPIMetrics(apiName, http.MethodPut, metrics.ErrStatus, start)
		}
		return errs[0]
	}

	if respData.Code != 0 {
		if !s.testDebug {
			metrics.ReportRequestAPIMetrics(apiName, http.MethodPut, fmt.Sprintf("%d", respData.Code), start)
		}

		return &AuthError{
			RequestID: resp.Header.Get(IamRequestHeader),
			Reason:    fmt.Errorf("update resource action %v failed, code[%d], msg[%s]", action, respData.Code, respData.Message),
		}
	}

	if !s.testDebug {
		metrics.ReportRequestAPIMetrics(apiName, http.MethodPut, fmt.Sprintf("%d", respData.Code), start)
	}

	return nil
}

func (s *iamModelServer) DeleteAction(timeout time.Duration, actionsIDs []ActionID) error {
	if s == nil {
		return ErrInitServerFail
	}
	const (
		apiName = "DeleteAction"
		path    = "/api/v1/model/systems/%s/actions"
	)

	ids := make([]struct {
		ID ActionID `json:"id"`
	}, len(actionsIDs))
	for index := range actionsIDs {
		ids[index].ID = actionsIDs[index]
	}

	var (
		url      = s.server + fmt.Sprintf(path, s.systemID)
		respData = &BaseResponse{}
		start    = time.Now()
	)

	resp, _, errs := gorequest.New().
		Timeout(timeout).
		Delete(url).
		Set("Content-Type", "application/json").
		Set("Connection", "close").
		Set("X-Bk-App-Code", s.appCode).
		Set("X-Bk-App-Secret", s.appSecret).
		Set("X-Bkapi-Authorization", s.gateWayAuth).
		SetDebug(s.serverDebug).
		Send(ids).
		EndStruct(respData)
	if len(errs) > 0 {
		blog.Errorf("call api DeleteAction failed: %v", errs[0])
		if !s.testDebug {
			metrics.ReportRequestAPIMetrics(apiName, http.MethodDelete, metrics.ErrStatus, start)
		}
		return errs[0]
	}

	if respData.Code != 0 {
		if !s.testDebug {
			metrics.ReportRequestAPIMetrics(apiName, http.MethodDelete, fmt.Sprintf("%d", respData.Code), start)
		}

		return &AuthError{
			RequestID: resp.Header.Get(IamRequestHeader),
			Reason:    fmt.Errorf("delete resource actions %v failed, code[%d], msg[%s]", actionsIDs, respData.Code, respData.Message),
		}
	}

	if !s.testDebug {
		metrics.ReportRequestAPIMetrics(apiName, http.MethodDelete, fmt.Sprintf("%d", respData.Code), start)
	}

	return nil
}

func (s *iamModelServer) RegisterActionGroup(timeout time.Duration, actionGroups []ActionGroup) error {
	if s == nil {
		return ErrInitServerFail
	}
	const (
		apiName = "RegisterActionGroup"
		path    = "/api/v1/model/systems/%s/configs/action_groups"
	)

	var (
		url      = s.server + fmt.Sprintf(path, s.systemID)
		respData = &BaseResponse{}
		start    = time.Now()
	)

	resp, _, errs := gorequest.New().
		Timeout(timeout).
		Post(url).
		Set("Content-Type", "application/json").
		Set("Connection", "close").
		Set("X-Bk-App-Code", s.appCode).
		Set("X-Bk-App-Secret", s.appSecret).
		Set("X-Bkapi-Authorization", s.gateWayAuth).
		SetDebug(s.serverDebug).
		Send(actionGroups).
		EndStruct(respData)
	if len(errs) > 0 {
		blog.Errorf("call api RegisterActionGroup failed: %v", errs[0])
		if !s.testDebug {
			metrics.ReportRequestAPIMetrics(apiName, http.MethodPost, metrics.ErrStatus, start)
		}
		return errs[0]
	}

	if respData.Code != 0 {
		if !s.testDebug {
			metrics.ReportRequestAPIMetrics(apiName, http.MethodPost, fmt.Sprintf("%d", respData.Code), start)
		}

		return &AuthError{
			RequestID: resp.Header.Get(IamRequestHeader),
			Reason: fmt.Errorf("register action groups %v failed, code[%d], msg[%s]", actionGroups,
				respData.Code, respData.Message),
		}
	}

	if !s.testDebug {
		metrics.ReportRequestAPIMetrics(apiName, http.MethodPost, fmt.Sprintf("%d", respData.Code), start)
	}

	return nil
}

func (s *iamModelServer) UpdateActionGroups(timeout time.Duration, actionGroups []ActionGroup) error {
	if s == nil {
		return ErrInitServerFail
	}
	const (
		apiName = "UpdateActionGroups"
		path    = "/api/v1/model/systems/%s/configs/action_groups"
	)

	var (
		url      = s.server + fmt.Sprintf(path, s.systemID)
		respData = &BaseResponse{}
		start    = time.Now()
	)

	resp, _, errs := gorequest.New().
		Timeout(timeout).
		Put(url).
		Set("Content-Type", "application/json").
		Set("Connection", "close").
		Set("X-Bk-App-Code", s.appCode).
		Set("X-Bk-App-Secret", s.appSecret).
		Set("X-Bkapi-Authorization", s.gateWayAuth).
		SetDebug(s.serverDebug).
		Send(actionGroups).
		EndStruct(respData)
	if len(errs) > 0 {
		blog.Errorf("call api UpdateActionGroups failed: %v", errs[0])
		if !s.testDebug {
			metrics.ReportRequestAPIMetrics(apiName, http.MethodPut, metrics.ErrStatus, start)
		}
		return errs[0]
	}

	if respData.Code != 0 {
		if !s.testDebug {
			metrics.ReportRequestAPIMetrics(apiName, http.MethodPut, fmt.Sprintf("%d", respData.Code), start)
		}

		return &AuthError{
			RequestID: resp.Header.Get(IamRequestHeader),
			Reason: fmt.Errorf("update action groups %v failed, code[%d], msg[%s]", actionGroups,
				respData.Code, respData.Message),
		}
	}

	if !s.testDebug {
		metrics.ReportRequestAPIMetrics(apiName, http.MethodPut, fmt.Sprintf("%d", respData.Code), start)
	}

	return nil
}

func (s *iamModelServer) CreateInstanceSelection(timeout time.Duration, instanceSelections []InstanceSelection) error {
	if s == nil {
		return ErrInitServerFail
	}
	const (
		apiName = "CreateInstanceSelection"
		path    = "/api/v1/model/systems/%s/instance-selections"
	)

	var (
		url      = s.server + fmt.Sprintf(path, s.systemID)
		respData = &BaseResponse{}
		start    = time.Now()
	)

	resp, _, errs := gorequest.New().
		Timeout(timeout).
		Post(url).
		Set("Content-Type", "application/json").
		Set("Connection", "close").
		Set("X-Bk-App-Code", s.appCode).
		Set("X-Bk-App-Secret", s.appSecret).
		Set("X-Bkapi-Authorization", s.gateWayAuth).
		SetDebug(s.serverDebug).
		Send(instanceSelections).
		EndStruct(respData)
	if len(errs) > 0 {
		blog.Errorf("call api CreateInstanceSelection failed: %v", errs[0])
		if !s.testDebug {
			metrics.ReportRequestAPIMetrics(apiName, http.MethodPost, metrics.ErrStatus, start)
		}
		return errs[0]
	}

	if respData.Code != 0 {
		if !s.testDebug {
			metrics.ReportRequestAPIMetrics(apiName, http.MethodPost, fmt.Sprintf("%d", respData.Code), start)
		}

		return &AuthError{
			RequestID: resp.Header.Get(IamRequestHeader),
			Reason: fmt.Errorf("add instance selections %v failed, code[%d], msg[%s]", instanceSelections,
				respData.Code, respData.Message),
		}
	}

	if !s.testDebug {
		metrics.ReportRequestAPIMetrics(apiName, http.MethodPost, fmt.Sprintf("%d", respData.Code), start)
	}

	return nil
}

func (s *iamModelServer) UpdateInstanceSelection(timeout time.Duration, instanceSelection InstanceSelection) error {
	if s == nil {
		return ErrInitServerFail
	}
	const (
		apiName = "UpdateInstanceSelection"
		path    = "/api/v1/model/systems/%s/instance-selections/%s"
	)

	var (
		url      = s.server + fmt.Sprintf(path, s.systemID, instanceSelection.ID)
		respData = &BaseResponse{}
		start    = time.Now()
	)

	resp, _, errs := gorequest.New().
		Timeout(timeout).
		Put(url).
		Set("Content-Type", "application/json").
		Set("Connection", "close").
		Set("X-Bk-App-Code", s.appCode).
		Set("X-Bk-App-Secret", s.appSecret).
		Set("X-Bkapi-Authorization", s.gateWayAuth).
		SetDebug(s.serverDebug).
		Send(instanceSelection).
		EndStruct(respData)
	if len(errs) > 0 {
		blog.Errorf("call api UpdateAction failed: %v", errs[0])
		if !s.testDebug {
			metrics.ReportRequestAPIMetrics(apiName, http.MethodPut, metrics.ErrStatus, start)
		}
		return errs[0]
	}

	if respData.Code != 0 {
		if !s.testDebug {
			metrics.ReportRequestAPIMetrics(apiName, http.MethodPut, fmt.Sprintf("%d", respData.Code), start)
		}

		return &AuthError{
			RequestID: resp.Header.Get(IamRequestHeader),
			Reason: fmt.Errorf("update instance selection %v failed, code[%d], msg[%s]", instanceSelection,
				respData.Code, respData.Message),
		}
	}

	if !s.testDebug {
		metrics.ReportRequestAPIMetrics(apiName, http.MethodPut, fmt.Sprintf("%d", respData.Code), start)
	}

	return nil
}

func (s *iamModelServer) DeleteInstanceSelection(timeout time.Duration, instanceSelectionIDs []InstanceSelectionID) error {
	if s == nil {
		return ErrInitServerFail
	}
	const (
		apiName = "DeleteInstanceSelection"
		path    = "/api/v1/model/systems/%s/instance-selections"
	)

	ids := make([]struct {
		ID InstanceSelectionID `json:"id"`
	}, len(instanceSelectionIDs))
	for index := range instanceSelectionIDs {
		ids[index].ID = instanceSelectionIDs[index]
	}

	var (
		url      = s.server + fmt.Sprintf(path, s.systemID)
		respData = &BaseResponse{}
		start    = time.Now()
	)

	resp, _, errs := gorequest.New().
		Timeout(timeout).
		Delete(url).
		Set("Content-Type", "application/json").
		Set("Connection", "close").
		Set("X-Bk-App-Code", s.appCode).
		Set("X-Bk-App-Secret", s.appSecret).
		Set("X-Bkapi-Authorization", s.gateWayAuth).
		SetDebug(s.serverDebug).
		Send(ids).
		EndStruct(respData)
	if len(errs) > 0 {
		blog.Errorf("call api DeleteInstanceSelection failed: %v", errs[0])
		if !s.testDebug {
			metrics.ReportRequestAPIMetrics(apiName, http.MethodDelete, metrics.ErrStatus, start)
		}
		return errs[0]
	}

	if respData.Code != 0 {
		if !s.testDebug {
			metrics.ReportRequestAPIMetrics(apiName, http.MethodDelete, fmt.Sprintf("%d", respData.Code), start)
		}

		return &AuthError{
			RequestID: resp.Header.Get(IamRequestHeader),
			Reason:    fmt.Errorf("delete instance selections %v failed, code[%d], msg[%s]", instanceSelectionIDs, respData.Code, respData.Message),
		}
	}

	if !s.testDebug {
		metrics.ReportRequestAPIMetrics(apiName, http.MethodDelete, fmt.Sprintf("%d", respData.Code), start)
	}

	return nil
}

func (s *iamModelServer) GetSystemToken(timeout time.Duration) (string, error) {
	if s == nil {
		return "", ErrInitServerFail
	}
	const (
		apiName = "GetSystemToken"
		path    = "/api/v1/model/systems/%s/token"
	)

	var (
		url      = s.server + fmt.Sprintf(path, s.systemID)
		respData = &TokenResponse{}
		start    = time.Now()
	)

	resp, _, errs := gorequest.New().
		Timeout(timeout).
		Get(url).
		Set("Content-Type", "application/json").
		Set("Connection", "close").
		Set("X-Bk-App-Code", s.appCode).
		Set("X-Bk-App-Secret", s.appSecret).
		Set("X-Bkapi-Authorization", s.gateWayAuth).
		SetDebug(s.serverDebug).
		EndStruct(respData)
	if len(errs) > 0 {
		blog.Errorf("call api GetSystemToken failed: %v", errs[0])
		if !s.testDebug {
			metrics.ReportRequestAPIMetrics(apiName, http.MethodGet, metrics.ErrStatus, start)
		}
		return "", errs[0]
	}

	if respData.Code != 0 {
		if !s.testDebug {
			metrics.ReportRequestAPIMetrics(apiName, http.MethodGet, fmt.Sprintf("%d", respData.Code), start)
		}

		return "", &AuthError{
			RequestID: resp.Header.Get(IamRequestHeader),
			Reason:    fmt.Errorf("get system token failed, code[%d], msg[%s]", respData.Code, respData.Message),
		}
	}

	if !s.testDebug {
		metrics.ReportRequestAPIMetrics(apiName, http.MethodGet, fmt.Sprintf("%d", respData.Code), start)
	}

	return respData.Data.Token, nil
}
