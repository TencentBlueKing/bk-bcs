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

package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/cmd/auth-server/types"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/constant"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/errf"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/uuid"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/iam/client"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/iam/sys"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/rest"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/runtime/shutdown"
)

// moduleType auth logic module type.
// nolint: unused
type moduleType string

const (
	// auth module.
	//nolint:unused
	authModule moduleType = "auth"
	// initial bscp auth model in iam module.
	//nolint:unused
	initialModule moduleType = "init"
	// iam callback module.
	//nolint:unused
	iamModule moduleType = "iam"
	//nolint:unused
	userModule moduleType = "user"
	//nolint:unused
	spaceModule moduleType = "space"
)

// setFilter set mux request filter.
// nolint: unused
func (g *gateway) setFilter(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		var module string
		// path format: /api/{api_version}/{service}/{module}/other
		paths := strings.Split(r.URL.Path, "/")
		if len(paths) <= 4 {
			logs.Errorf("received url path length not conform to the regulations, path: %s", r.URL.Path)
			fmt.Fprint(w, errf.New(http.StatusNotFound, "Not Found").Error())
			return
		}
		module = paths[4]

		switch moduleType(module) {
		case iamModule:
			if err := iamRequestFilter(g.iamSys, w, r); err != nil {
				fmt.Fprint(w, errf.Error(err).Error())
				return
			}

		case authModule:
			if err := authRequestFilter(w, r); err != nil {
				fmt.Fprint(w, errf.Error(err).Error())
				return
			}

		case initialModule, userModule, spaceModule:

		default:
			logs.Errorf("received unknown module's request req: %v", r)
			fmt.Fprint(w, errf.New(http.StatusNotFound, "Not Found").Error())
			return
		}

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

// HealthyHandler livenessProbe 健康检查
func (g *gateway) HealthyHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

// ReadyHandler ReadinessProbe 健康检查
func (g *gateway) ReadyHandler(w http.ResponseWriter, r *http.Request) {
	g.Healthz(w, r)
}

// Healthz service health check.
func (g *gateway) Healthz(w http.ResponseWriter, r *http.Request) {
	if shutdown.IsShuttingDown() {
		logs.Errorf("service healthz check failed, current service is shutting down")
		w.WriteHeader(http.StatusServiceUnavailable)
		rest.WriteResp(w, rest.NewBaseResp(errf.UnHealth, "current service is shutting down"))
		return
	}

	if err := g.state.Healthz(); err != nil {
		logs.Errorf("etcd healthz check failed, err: %v", err)
		rest.WriteResp(w, rest.NewBaseResp(errf.UnHealth, "etcd healthz error, "+err.Error()))
		return
	}

	rest.WriteResp(w, rest.NewBaseResp(errf.OK, "healthy"))
}

// iamRequestFilter setups all api filters here. All request would cross here, and we filter request base on URL.
// nolint: unused
func iamRequestFilter(sysCli *sys.Sys, w http.ResponseWriter, req *http.Request) error {
	isAuthorized, err := checkRequestAuthorization(sysCli, req)
	if err != nil {
		return errf.New(http.StatusInternalServerError, err.Error())
	}
	if !isAuthorized {
		return errf.New(types.UnauthorizedErrorCode, "authorized failed")
	}

	rid := getRid(req.Header)
	req.Header.Set(constant.RidKey, rid)

	// set rid to response header, used to troubleshoot the problem.
	w.Header().Set(client.RequestIDHeader, rid)

	// use sys language as bscp language
	req.Header.Set(constant.LanguageKey, req.Header.Get("Blueking-Language"))

	user := req.Header.Get(constant.UserKey)
	if len(user) == 0 {
		req.Header.Set(constant.UserKey, "auth")
	}

	appCode := req.Header.Get(constant.AppCodeKey)
	if len(appCode) == 0 {
		req.Header.Set(constant.AppCodeKey, client.SystemIDIAM)
	}

	return nil
}

// getRid get request id from header. if rid is empty, generate a rid to return.
//
//lint:ignore U1000 getRid is unused
func getRid(h http.Header) string {
	if rid := h.Get(client.RequestIDHeader); len(rid) != 0 {
		return rid
	}

	if rid := h.Get(constant.RidKey); len(rid) != 0 {
		return rid
	}

	return uuid.UUID()
}

// authRequestFilter set auth request filter.
// nolint: unused
func authRequestFilter(w http.ResponseWriter, req *http.Request) error {
	// Note: set auth request filter.

	return nil
}

// nolint: unused
var iamToken = struct {
	token            string
	tokenRefreshTime time.Time
}{}

// nolint: unused
func checkRequestAuthorization(cli *sys.Sys, req *http.Request) (bool, error) {
	rid := req.Header.Get(client.RequestIDHeader)
	name, pwd, ok := req.BasicAuth()
	if !ok || name != client.SystemIDIAM {
		logs.Errorf("request have no basic authorization, rid: %s", rid)
		return false, nil
	}

	// if cached token is set within a minute, use it to check request authorization
	if iamToken.token != "" && time.Since(iamToken.tokenRefreshTime) <= time.Minute && pwd == iamToken.token {
		return true, nil
	}

	var err error
	iamToken.token, err = cli.GetSystemToken(context.Background())
	if err != nil {
		logs.Errorf("check request authorization get system token failed, error: %s, rid: %s", err.Error(), rid)
		return false, err
	}

	iamToken.tokenRefreshTime = time.Now()
	if pwd != iamToken.token {
		return false, errors.New("request password not match system token")
	}

	return true, nil
}
