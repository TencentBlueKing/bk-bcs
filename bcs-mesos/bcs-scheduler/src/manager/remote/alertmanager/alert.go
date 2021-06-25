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

package alertmanager

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/encrypt"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-scheduler/src/util"

	"github.com/parnurzeal/gorequest"
)

var (
	errInitSchemaServer     = errors.New("server schema err: http or https")
	errInitClientAuthServer = errors.New("server client auth err: clientAuth true when https")
	errInitTLSServer        = errors.New("server tls err: config tls config when https")
	errNotInitServer        = errors.New("server not init")
)

// AlertManageInterface for alertmanager interface
type AlertManageInterface interface {
	CreateAlertInfoToAlertManager(req *CreateBusinessAlertInfoReq, timeout time.Duration) error
}

// Options for alert-manager server conf
type Options struct {
	Server     string
	Token      string
	ClientAuth bool
	Debug      bool
}

type alertManager struct {
	opt Options
}

// NewAlertManager init alertmanager server
func NewAlertManager(opt Options) (AlertManageInterface, error) {
	_, err := validateOptions(opt)
	if err != nil {
		return nil, err
	}

	return &alertManager{
		opt: opt,
	}, nil
}

// am.opt.clientAuth = true
func (am *alertManager) getAPIGatewayToken() (string, error) {
	if am == nil {
		return "", errNotInitServer
	}

	password := am.opt.Token
	if password != "" {
		realPwd, _ := encrypt.DesDecryptFromBase([]byte(password))
		password = string(realPwd)
	}

	return password, nil
}

func (am *alertManager) CreateAlertInfoToAlertManager(req *CreateBusinessAlertInfoReq, timeout time.Duration) error {
	if am == nil {
		return nil
	}

	const (
		apiName = "CreateAlertInfoToAlertManager"
		path    = "/alertmanager/v1/businessalerts"
	)

	var (
		url      = am.opt.Server + path
		start    = time.Now()
		respData = &CreateBusinessAlertInfoResp{}
		token    string
		err      error
	)

	superAgent := gorequest.New().Timeout(timeout).Post(url).
		Set("Content-Type", "application/json").
		Set("Connection", "close")

	if am.opt.ClientAuth {
		token, err = am.getAPIGatewayToken()
		if err != nil {
			blog.Errorf("getAPIGatewayToken err: %v", err)
			return err
		}

		superAgent = superAgent.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	}

	resp, body, errs := superAgent.SetDebug(am.opt.Debug).Send(req).EndStruct(respData)
	if len(errs) > 0 {
		blog.Errorf("call api CreateAlertInfoToAlertManager failed: %v", errs[0])
		util.ReportLibAlertManagerAPIMetrics(apiName, http.MethodPost, util.ErrStatus, start)
		return errs[0]
	}

	if resp.StatusCode != http.StatusOK {
		errMsg := fmt.Errorf("call bcs-alert-manager API error: code[%v], body[%v], err[%s]",
			resp.StatusCode, string(body), respData.ErrMsg)
		util.ReportLibAlertManagerAPIMetrics(apiName, http.MethodPost, fmt.Sprintf("%d", resp.StatusCode), start)
		return errMsg
	}

	util.ReportLibAlertManagerAPIMetrics(apiName, http.MethodPost, fmt.Sprintf("%d", resp.StatusCode), start)
	return nil
}

func validateOptions(opt Options) (bool, error) {
	if !strings.HasPrefix(opt.Server, "http://") && !strings.HasPrefix(opt.Server, "https://") {
		return false, errInitSchemaServer
	}

	if strings.HasPrefix(opt.Server, "https") {
		if !opt.ClientAuth {
			return false, errInitClientAuthServer
		}
		if len(opt.Token) == 0 {
			return false, errInitTLSServer
		}
	}

	return true, nil
}
