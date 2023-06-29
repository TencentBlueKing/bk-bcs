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
	"context"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy"
)

// WebhookPlugin defines the webhook plugin
type WebhookPlugin struct {
	*mux.Router
	middleware MiddlewareInterface
}

// Init initialize webhook plugin
func (plugin *WebhookPlugin) Init() error {
	plugin.Path("").Methods("POST").
		Handler(plugin.middleware.HttpWrapper(plugin.executeWebhook))

	blog.Infof("argocd webhook plugin init successfully")
	return nil
}

func (plugin *WebhookPlugin) executeWebhook(ctx context.Context, r *http.Request) *httpResponse {
	user := ctx.Value(ctxKeyUser).(*proxy.UserInfo)
	blog.Infof("user %s request webhook", user.GetUser())
	return nil
}
