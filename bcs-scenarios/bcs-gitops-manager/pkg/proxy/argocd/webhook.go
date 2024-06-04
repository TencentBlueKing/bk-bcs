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

package argocd

import (
	"net/http"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/gorilla/mux"
	"gopkg.in/go-playground/webhooks.v5/github"
	"gopkg.in/go-playground/webhooks.v5/gitlab"

	mw "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy/argocd/middleware"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/store"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/utils"
)

// WebhookPlugin defines the webhook plugin
type WebhookPlugin struct {
	*mux.Router
	middleware    mw.MiddlewareInterface
	appsetWebhook string
	storage       store.Store

	github *github.Webhook
	gitlab *gitlab.Webhook
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
	reqCopy, err := utils.DeepCopyHttpRequest(r, plugin.appsetWebhook)
	if err != nil {
		blog.Errorf("RequestID[%s] copy webhook request failed: %s", mw.RequestID(r.Context()), err.Error())
	} else {
		go plugin.forwardToApplicationSet(reqCopy, requestID)
	}
	blog.Infof("RequestID[%s], user %s request webhook", requestID, user.GetUser())
	return r, mw.ReturnArgoReverse()
}

// forwardToApplicationSet this will check the webhook repoURL whether matched appset's git generator
// it'll refresh appset if matched.
func (plugin *WebhookPlugin) forwardToApplicationSet(r *http.Request, requestID string) {
	var payloadIf interface{}
	var err error
	switch {
	case r.Header.Get("X-GitHub-Event") != "":
		payloadIf, err = plugin.github.Parse(r, github.PushEvent, github.PingEvent)
		if err != nil {
			blog.Errorf("RequestID[%s] github event parse failed: %s", requestID, err.Error())
			return
		}
	case r.Header.Get("X-Gitlab-Event") != "":
		payloadIf, err = plugin.gitlab.Parse(r, gitlab.PushEvents, gitlab.TagEvents, gitlab.SystemHookEvents)
		if err != nil {
			blog.Errorf("RequestID[%s] gitlab event parse failed: %s", requestID, err.Error())
			return
		}
	default:
		blog.Errorf("RequestID[%s] ignore unknown event", requestID)
		return
	}
	var repoURL string
	switch payload := payloadIf.(type) {
	case github.PushPayload:
		repoURL = payload.Repository.HTMLURL
	case gitlab.PushEventPayload:
		repoURL = payload.Project.WebURL
	default:
		blog.Errorf("RequestID[%s] ignore unknown webhook payload type", requestID)
		return
	}
	blog.Infof("RequestID[%s] repo url: %s", requestID, repoURL)

	appSets := plugin.storage.AllApplicationSets()
	refreshAppSets := make([]*v1alpha1.ApplicationSet, 0)
	for _, k8sAppSet := range appSets {
		for i := range k8sAppSet.Spec.Generators {
			gitGenerator := k8sAppSet.Spec.Generators[i].Git
			if gitGenerator == nil {
				continue
			}
			if utils.CheckGitRepoSimilar(gitGenerator.RepoURL, repoURL) {
				refreshAppSets = append(refreshAppSets, k8sAppSet)
				break
			}
		}
	}
	for _, k8sAppSet := range refreshAppSets {
		if err = plugin.storage.RefreshApplicationSet(k8sAppSet.Namespace, k8sAppSet.Name); err != nil {
			blog.Errorf("RequestID[%s] refresh appset '%s' failed with webhook: %s",
				requestID, k8sAppSet.Name, err.Error())
		} else {
			blog.Infof("RequestID[%s] refresh appset '%s' success", requestID, k8sAppSet.Name)
		}
	}
}
