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

package argocd

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/project"
)

// WebhookPlugin defines the webhook plugin
type WebhookPlugin struct {
	*mux.Router

	Session    *Session
	Permission *project.BCSProjectPerm
	option     *proxy.GitOpsOptions
}

// Init initialize webhook plugin
func (plugin *WebhookPlugin) Init() error {
	plugin.Permission = project.NewBCSProjectPermClient(plugin.option.IAMClient)
	plugin.Path("").Methods("POST").HandlerFunc(plugin.executeWebhook)

	blog.Infof("argocd webhook plugin init successfully")
	return nil
}

// TODO: 需增加权限校验
func (plugin *WebhookPlugin) executeWebhook(w http.ResponseWriter, r *http.Request) {
	user, err := proxy.GetJWTInfo(r, plugin.option.JWTDecoder)
	if err != nil {
		blog.Errorf("request %s get jwt token failure, %s", r.URL.Path, err.Error())
		http.Error(w,
			fmt.Sprintf("Bad Request: %s", err.Error()),
			http.StatusBadRequest,
		)
		return
	}
	blog.Infof("user %s request webhook", user.GetUser())
	plugin.Session.ServeHTTP(w, r)
}
