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
	"time"

	"github.com/gorilla/mux"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	mw "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy/argocd/middleware"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/utils"
)

// WebhookPlugin defines the webhook plugin
type WebhookPlugin struct {
	*mux.Router
	middleware    mw.MiddlewareInterface
	appsetWebhook string
}

// Init initialize webhook plugin
func (plugin *WebhookPlugin) Init() error {
	plugin.Path("").Methods("POST").
		Handler(plugin.middleware.HttpWrapper(plugin.executeWebhook))

	blog.Infof("argocd webhook plugin init successfully")
	return nil
}

func (plugin *WebhookPlugin) executeWebhook(r *http.Request) (*http.Request, *mw.HttpResponse) {
	user := mw.User(r.Context())
	requestID := mw.RequestID(r.Context())
	blog.Infof("RequestID[%s], user %s request webhook", requestID, user.GetUser())
	reqCopy, err := utils.DeepCopyHttpRequest(r, plugin.appsetWebhook)
	if err != nil {
		blog.Errorf("RequestID[%s] copy webhook request failed: %s", mw.RequestID(r.Context()), err.Error())
	} else {
		go plugin.forwardToApplicationSet(reqCopy, requestID)
	}
	return r, mw.ReturnArgoReverse()
}

func (plugin *WebhookPlugin) forwardToApplicationSet(r *http.Request, requestID string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	r.WithContext(ctx)
	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		blog.Errorf("RequestID[%s] webhook forward to appset controller failed: %s", requestID, err.Error())
		return
	}
	if resp.StatusCode != http.StatusOK {
		blog.Errorf("RequestID[%s] webhook forward to appset controller resp code %d: %s",
			requestID, resp.StatusCode)
		return
	}
	blog.Infof("RequestID[%s] webhook forward to appset controller success", requestID)
}
