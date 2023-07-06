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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/common"
)

// AppPlugin for internal project authorization
type AppPlugin struct {
	*mux.Router
	middleware MiddlewareInterface
}

// all argocd application URL:
// * required Project Edit projectPermission
// POST：  /api/v1/applications，创建
// DELETE：/api/v1/applications/{name}，指定删除
// PUT：   /api/v1/applications/{name}，指定更新
// PATCH： /api/v1/applications/{name}，指定字段更新
// DELETE：/api/v1/applications/{name}/operation，终止当前操作
// POST：  /api/v1/applications/{name}/resource，patch资源
// DELETE：/api/v1/applications/{name}/resource，删除资源
// POST：  /api/v1/applications/{name}/resource/actions，run resource action
// POST：  /api/v1/applications/{name}/rollback，回退到上个版本
// PUT：   /api/v1/applications/{name}/spec，更新spec
// POST：  /api/v1/applications/{name}/sync，发起app同步
//
// * required Project View projectPermission
// GET：/api/v1/applications?projects={projects}，获取列表，强制启用projects参数
//
// path prefix format: /api/v1/applications/{name}
// GET：/api/v1/applications/{name}，获取具体信息
// GET：/api/v1/applications/{name}/managed-resources，返回管理资源
// GET：/api/v1/applications/{name}/resource-tree，返回资源树
// GET：/api/v1/applications/{name}/events
// GET：/api/v1/applications/{name}/logs，日志，建议直接访问集群接口
// GET：/api/v1/applications/{name}/manifests
// GET：/api/v1/applications/{name}/pods/{podName}/logs，获取Pod日志，建议直接访问集群接口
// GET：/api/v1/applications/{name}/resource，获取资源
// GET：/api/v1/applications/{name}/resource/actions，获取actions
// GET：/api/v1/applications/{name}/revisions/{revision}/metadata，获取指定版本的meta数据
// GET：/api/v1/applications/{name}/syncwindws，获取syncwindows
//

// Init all project sub path handler
// project plugin is a subRouter, all path registed is relative
func (plugin *AppPlugin) Init() error {
	// POST /api/v1/applications, create new application
	plugin.Path("").Methods("POST").
		Handler(plugin.middleware.HttpWrapper(plugin.createApplicationHandler))
	// force check query, GET /api/v1/applications?projects={projects}
	plugin.Path("").Methods("GET").Queries("projects", "{projects}").
		Handler(plugin.middleware.HttpWrapper(plugin.listApplicationsHandler))

	// Put,Patch,Delete with preifx /api/v1/applications/{name}
	plugin.PathPrefix("/{name}").Methods("PUT", "POST", "DELETE", "PATCH").
		Handler(plugin.middleware.HttpWrapper(plugin.applicationEditHandler))

	// GET with prefix /api/v1/applications/{name}
	plugin.PathPrefix("/{name}").Methods("GET").
		Handler(plugin.middleware.HttpWrapper(plugin.applicationViewsHandler))

	// todo(DeveloperJim): GET /api/v1/stream/applications?project={project}
	// todo(DeveloperJim): GET /api/v1/stream/applications/{name}/resource-tree
	blog.Infof("argocd application plugin init successfully")
	return nil
}

// POST /api/v1/applications, create new application
// validate project detail from request
func (plugin *AppPlugin) createApplicationHandler(ctx context.Context, r *http.Request) *httpResponse {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return &httpResponse{
			statusCode: http.StatusBadRequest,
			err:        errors.Wrapf(err, "read body failed"),
		}
	}
	app := &v1alpha1.Application{}
	if err = json.Unmarshal(body, app); err != nil {
		return &httpResponse{
			statusCode: http.StatusBadRequest,
			err:        errors.Wrapf(err, "unmarshal body failed"),
		}
	}
	if app.Spec.Project == "" || app.Spec.Project == "default" {
		return &httpResponse{
			statusCode: http.StatusBadRequest,
			err:        errors.Errorf("project information lost"),
		}
	}
	argoProject, statusCode, err := plugin.middleware.CheckProjectPermission(ctx, app.Spec.Project, iam.ProjectEdit)
	if statusCode != http.StatusOK {
		return &httpResponse{
			statusCode: statusCode,
			err:        errors.Wrapf(err, "check project '%s' edit permission failed", app.Spec.Project),
		}
	}

	// setting control annotations
	if app.Annotations == nil {
		app.Annotations = make(map[string]string)
	}
	app.Annotations[common.ProjectIDKey] = common.GetBCSProjectID(argoProject.Annotations)
	app.Annotations[common.ProjectBusinessIDKey] = argoProject.Annotations[common.ProjectBusinessIDKey]
	updatedBody, err := json.Marshal(app)
	if err != nil {
		return &httpResponse{
			statusCode: http.StatusBadRequest,
			err:        errors.Wrapf(err, "json marshal application failed: %v", app),
		}
	}
	r.Body = ioutil.NopCloser(bytes.NewBuffer(updatedBody))
	length := len(updatedBody)
	r.Header.Set("Content-Length", strconv.Itoa(length))
	r.ContentLength = int64(length)
	return nil
}

// GET /api/v1/applications?projects={projects}
func (plugin *AppPlugin) listApplicationsHandler(ctx context.Context, r *http.Request) *httpResponse {
	projectName := r.URL.Query().Get("projects")
	_, statusCode, err := plugin.middleware.CheckProjectPermission(ctx, projectName, iam.ProjectView)
	if statusCode != http.StatusOK {
		return &httpResponse{
			statusCode: statusCode,
			err:        err,
		}
	}
	appList, err := plugin.middleware.ListApplications(ctx, []string{projectName})
	if err != nil {
		return &httpResponse{
			statusCode: http.StatusInternalServerError,
			err:        errors.Wrapf(err, "list applications by project '%s' from storage failed", projectName),
		}
	}
	return &httpResponse{
		obj: appList,
	}
}

// Put,Patch,Delete with preifx /api/v1/applications/{name}
func (plugin *AppPlugin) applicationEditHandler(ctx context.Context, r *http.Request) *httpResponse {
	appName := mux.Vars(r)["name"]
	if appName == "" {
		return &httpResponse{
			statusCode: http.StatusBadRequest,
			err:        fmt.Errorf("request application name cannot be empty"),
		}
	}
	_, statusCode, err := plugin.middleware.CheckApplicationPermission(ctx, appName, iam.ProjectEdit)
	if statusCode != http.StatusOK {
		return &httpResponse{
			statusCode: statusCode,
			err:        err,
		}
	}
	return nil
}

// GET with prefix /api/v1/applications/{name}
func (plugin *AppPlugin) applicationViewsHandler(ctx context.Context, r *http.Request) *httpResponse {
	appName := mux.Vars(r)["name"]
	if appName == "" {
		return &httpResponse{
			statusCode: http.StatusBadRequest,
			err:        fmt.Errorf("request application name cannot be empty"),
		}
	}
	_, statusCode, err := plugin.middleware.CheckApplicationPermission(ctx, appName, iam.ProjectView)
	if statusCode != http.StatusOK {
		return &httpResponse{
			statusCode: statusCode,
			err:        err,
		}
	}
	return nil
}
