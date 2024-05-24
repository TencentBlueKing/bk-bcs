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
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
	bcsapi "github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapiv4"
	traceconst "github.com/Tencent/bk-bcs/bcs-common/pkg/otel/trace/constants"
	iamnamespace "github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth-v4/namespace"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/cmd/manager/options"
	mw "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy/argocd/middleware"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy/argocd/resources"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/store"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/utils/jsonq"
)

// * addinational stream path, application wrapper
// GET：/api/v1/stream/applications?projects={projects}，获取event事件流，强制启用projects
// GET：/api/v1/stream/applications/{name}/resource-tree，指定资源树事件流

// StreamPlugin for internal streaming
type StreamPlugin struct {
	*mux.Router

	appHandler *AppPlugin
	middleware mw.MiddlewareInterface

	storage    store.Store
	bcsStorage bcsapi.Storage
}

// Init all project sub path handler
// project plugin is a subRouter, all path registered is relative
func (plugin *StreamPlugin) Init() error {
	// done(DeveloperJim): GET /api/v1/stream/applications?projects={projects}
	plugin.Path("").Methods("GET").Handler(plugin.middleware.HttpWrapper(plugin.projectViewHandler))

	// done(DeveloperJim): GET /api/v1/stream/applications/{name}/resource-tree
	plugin.Path("/{name}/resource-tree").Methods("GET").
		Handler(plugin.middleware.HttpWrapper(plugin.applicationViewsHandler))

	// 自定义接口
	plugin.Path("/{name}/pod-resources").Methods("GET").
		Handler(plugin.extraInfo())
	blog.Infof("argocd stream applications plugin init successfully")
	return nil
}

func (plugin *StreamPlugin) projectViewHandler(r *http.Request) (*http.Request, *mw.HttpResponse) {
	projects := r.URL.Query()["projects"]
	if len(projects) == 0 {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest,
			fmt.Errorf("query param 'projects' cannot be empty"))
	}
	for i := range projects {
		projectName := projects[i]
		_, statusCode, err := plugin.middleware.CheckProjectPermission(r.Context(), projectName, iam.ProjectView)
		if statusCode != http.StatusOK {
			return r, mw.ReturnErrorResponse(statusCode,
				errors.Wrapf(err, "check project '%s' permission failed", projectName))
		}
	}
	// only return argo stream reverse when fields query not empty
	fields := r.URL.Query().Get("fields")
	if fields == "" {
		return r, mw.ReturnArgoReverse()
	}
	return r, mw.ReturnArgoStreamReverse()
}

// GET with prefix /api/v1/applications/{name}
func (plugin *StreamPlugin) applicationViewsHandler(r *http.Request) (*http.Request, *mw.HttpResponse) {
	appName := mux.Vars(r)["name"]
	if appName == "" {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest,
			fmt.Errorf("request application name cannot be empty"))
	}
	_, statusCode, err := plugin.middleware.CheckApplicationPermission(r.Context(), appName,
		iamnamespace.NameSpaceScopedView)
	if statusCode != http.StatusOK {
		return r, mw.ReturnErrorResponse(statusCode, err)
	}
	// only return argo stream reverse when fields query not empty
	fields := r.URL.Query().Get("fields")
	if fields == "" {
		return r, mw.ReturnArgoReverse()
	}
	return r, mw.ReturnArgoStreamReverse()
}

type streamPodResources struct {
	plugin *StreamPlugin
	op     *options.Options
}

func (plugin *StreamPlugin) extraInfo() http.Handler {
	return &streamPodResources{
		plugin: plugin,
		op:     options.GlobalOptions(),
	}
}

const (
	streamPodResourcesDuration = 10
	streamPodResourcesTimeout  = 120
)

// ServeHTTP 用来处理 application 的 pod-resource event-stream 接口实现
func (s *streamPodResources) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	ctx, requestID := mw.SetContext(rw, req, s.op.JWTDecoder)
	if ctx == nil {
		return
	}
	req = req.WithContext(ctx)
	appName := mux.Vars(req)["name"]
	if appName == "" {
		http.Error(rw, "request application name cannot be empty", http.StatusBadRequest)
		return
	}
	argoApp, statusCode, err := s.plugin.middleware.CheckApplicationPermission(
		req.Context(), appName, iamnamespace.NameSpaceScopedView)
	if statusCode != http.StatusOK {
		http.Error(rw, err.Error(), statusCode)
		return
	}
	rw.Header().Set("Content-Type", "text/event-stream")
	rw.Header().Set("Cache-Control", "no-cache")
	rw.Header().Set("Connection", "keep-alive")
	rw.Header().Set("Access-Control-Allow-Origin", "*")
	rw.Header().Set(traceconst.RequestIDHeaderKey, requestID)

	fields := req.URL.Query().Get("fields")
	var fieldPath []string
	if fields != "" {
		fieldPath = strings.Split(fields, ",")
	}
	podQuery := &resources.PodQuery{
		Storage:    s.plugin.storage,
		BCSStorage: s.plugin.bcsStorage,
	}
	var queryStatus int
	if queryStatus, err = s.queryPodResources(rw, req, podQuery, argoApp, fieldPath); err != nil {
		http.Error(rw, err.Error(), queryStatus)
		return
	}

	ticker := time.NewTicker(streamPodResourcesDuration * time.Second)
	defer ticker.Stop()
	timeout := time.After(streamPodResourcesTimeout * time.Second)
	for {
		select {
		case <-ticker.C:
			if queryStatus, err = s.queryPodResources(rw, req, podQuery, argoApp, fieldPath); err != nil {
				http.Error(rw, err.Error(), queryStatus)
				return
			}
		case <-timeout:
			rw.WriteHeader(http.StatusOK)
			return
		case <-req.Context().Done():
			rw.WriteHeader(http.StatusOK)
			return
		}
	}
}

func (s *streamPodResources) queryPodResources(rw http.ResponseWriter, req *http.Request,
	podQuery *resources.PodQuery, argoApp *v1alpha1.Application, fieldPath []string) (int, error) {
	var pods []corev1.Pod
	pods, err := podQuery.Query(req.Context(), argoApp)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	var result []byte
	if len(fieldPath) == 0 {
		result, _ = json.Marshal(pods)
	} else {
		result, err = jsonq.ReserveField(pods, fieldPath)
		if err != nil {
			return http.StatusInternalServerError, fmt.Errorf("parse by query 'fields' failed: %s", err.Error())
		}
	}
	// nolint
	_, _ = rw.Write([]byte(fmt.Sprintf("data: %s\n\n", result)))
	rw.(http.Flusher).Flush()
	return http.StatusOK, nil
}
