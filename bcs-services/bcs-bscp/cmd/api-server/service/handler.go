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
	"net/http"

	"github.com/go-chi/render"

	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/cc"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/errf"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/rest"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/runtime/shutdown"
)

// HealthyHandler livenessProbe 健康检查
func (p *proxy) HealthyHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

// ReadyHandler ReadinessProbe 健康检查
func (p *proxy) ReadyHandler(w http.ResponseWriter, r *http.Request) {
	p.Healthz(w, r)
}

// Healthz service health check.
func (p *proxy) Healthz(w http.ResponseWriter, r *http.Request) {
	if shutdown.IsShuttingDown() {
		logs.Errorf("service healthz check failed, current service is shutting down")
		w.WriteHeader(http.StatusServiceUnavailable)
		rest.WriteResp(w, rest.NewBaseResp(errf.UnHealth, "current service is shutting down"))
		return
	}

	if err := p.state.Healthz(); err != nil {
		logs.Errorf("etcd healthz check failed, err: %v", err)
		rest.WriteResp(w, rest.NewBaseResp(errf.UnHealth, "etcd healthz error, "+err.Error()))
		return
	}

	rest.WriteResp(w, rest.NewBaseResp(errf.OK, "healthy"))
}

// LogoutHandler return redirect url
func (p *proxy) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	data := p.authorizer.LogOut(r)
	render.Render(w, r, rest.OKRender(data))
}

// UserInfoHandler 鉴权后的用户信息接口
func UserInfoHandler(w http.ResponseWriter, r *http.Request) {
	k := kit.MustGetKit(r.Context())

	user := kit.User{Username: k.User}
	render.Render(w, r, rest.OKRender(user))
}

// FeatureFlags map of feature flags
type FeatureFlags map[cc.FeatureFlag]bool

// FeatureFlagsHandler 特性开关接口
func FeatureFlagsHandler(w http.ResponseWriter, r *http.Request) {
	featureFlags := FeatureFlags{}

	biz := r.URL.Query().Get("biz")
	for k, v := range cc.ApiServer().FeatureFlags {
		// 默认和开关开启保持一致
		featureFlags[k] = v.Enabled

		if biz == "" {
			continue
		}

		// 默认未开启, 设置是白名单模式，否则取反
		for _, w := range v.List {
			if biz == w {
				featureFlags[k] = !v.Enabled
			}
		}
	}

	render.Render(w, r, rest.OKRender(featureFlags))
}
