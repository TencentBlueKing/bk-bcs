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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"

	mw "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy/argocd/middleware"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy/argocd/permitcheck"
)

// PreCheckPlugin plugin
type PreCheckPlugin struct {
	*mux.Router
	permitChecker permitcheck.PermissionInterface
	middleware    mw.MiddlewareInterface
}

// GET:		/api/v1/monitor/{biz_id}, 查询bizID下对应的所有AppMonitor信息
// GET:		/api/v1/monitor/{biz_id}/list_scenario, 查询bizID下能安装的场景信息
// DELETE:	/api/v1/monitor/{biz_id}/{scenario}, 删除bizID下安装的对应场景监控
// POST:	/api/v1/monitor/{biz_id}/{scenario}, 创建/更新bizID下安装的对应场景监控

// Init all project sub path handler
// project plugin is a subRouter, all path registered is relative
func (plugin *PreCheckPlugin) Init() error {
	plugin.Path("/mr/info").Methods(http.MethodGet).
		Handler(plugin.middleware.HttpWrapper(plugin.preCheckViewHandler))
	plugin.Path("/record").Methods(http.MethodPost).
		Handler(plugin.middleware.HttpWrapper(plugin.preCheckCreateRecordHandler))
	plugin.Path("/task").Methods(http.MethodPut).
		Handler(plugin.middleware.HttpWrapper(plugin.preCheckCreateRecordHandler))
	plugin.Path("/task").Methods(http.MethodGet).
		Handler(plugin.middleware.HttpWrapper(plugin.preCheckGetTaskHandler))
	plugin.Path("/tasks").Methods(http.MethodGet).
		Handler(plugin.middleware.HttpWrapper(plugin.preCheckListTaskHandler))
	plugin.Path("/scan/report").Methods(http.MethodGet).
		Handler(plugin.middleware.HttpWrapper(plugin.preCheckGetScanReportHandler))
	plugin.Path("/scan/report").Methods(http.MethodPost).
		Handler(plugin.middleware.HttpWrapper(plugin.preCheckCreateScanReportHandler))
	blog.Infof("precheck plugin init successfully")
	return nil
}

// GET /api/v1/monitor/{biz_id}
func (plugin *PreCheckPlugin) preCheckViewHandler(r *http.Request) (*http.Request, *mw.HttpResponse) {
	// repo := r.PathValue("repository")
	queryValues := r.URL.Query()
	repo := queryValues.Get("repository")
	project := queryValues.Get("project")
	// mrIID := r.PathValue("mrIID")
	_, statusCode, err := plugin.permitChecker.CheckRepoPermission(r.Context(), project,
		repo, permitcheck.RepoViewRSAction)
	if statusCode != http.StatusOK {
		return r, mw.ReturnErrorResponse(statusCode, fmt.Errorf("check repo '%s' permission failed: %v", repo,
			err))
	}
	return r, mw.ReturnPreCheckReverse()
}

func (plugin *PreCheckPlugin) preCheckGetScanReportHandler(r *http.Request) (*http.Request, *mw.HttpResponse) {
	queryValues := r.URL.Query()
	app := queryValues.Get("app")
	_, statusCode, err := plugin.permitChecker.CheckApplicationPermission(r.Context(), app, permitcheck.AppViewRSAction)
	if statusCode != http.StatusOK {
		return r, mw.ReturnErrorResponse(statusCode, fmt.Errorf("check app '%s' permission failed: %v", app,
			err))
	}
	return r, mw.ReturnPreCheckReverse()
}

func (plugin *PreCheckPlugin) preCheckCreateScanReportHandler(r *http.Request) (*http.Request, *mw.HttpResponse) {
	req := make(map[string]interface{})
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest, errors.Wrapf(err, "read body failed"))
	}
	if err = json.Unmarshal(body, &req); err != nil {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest, errors.Wrapf(err, "unmarshal body failed"))
	}
	app := req["appName"]
	appStr, ok := app.(string)
	if !ok {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest,
			fmt.Errorf("assert app to string failed:%v", app))
	}
	_, statusCode, err := plugin.permitChecker.CheckApplicationPermission(r.Context(), appStr,
		permitcheck.AppUpdateRSAction)
	if statusCode != http.StatusOK {
		return r, mw.ReturnErrorResponse(statusCode, fmt.Errorf("check app '%s' permission failed: %v", app,
			err))
	}
	r.Body = io.NopCloser(bytes.NewBuffer(body))
	length := len(body)
	r.Header.Set("Content-Length", strconv.Itoa(length))
	r.ContentLength = int64(length)
	return r, mw.ReturnPreCheckReverse()
}

func (plugin *PreCheckPlugin) preCheckListTaskHandler(r *http.Request) (*http.Request, *mw.HttpResponse) {
	projects := r.URL.Query()["projects"]
	if len(projects) == 0 {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest,
			fmt.Errorf("query param 'projects' cannot be empty"))
	}
	for _, project := range projects {
		_, statusCode, err := plugin.permitChecker.CheckProjectPermission(r.Context(), project,
			permitcheck.ProjectViewRSAction)
		if statusCode != http.StatusOK {
			return r, mw.ReturnErrorResponse(statusCode, fmt.Errorf("check project '%s' permission failed: %v",
				project, err))
		}
	}
	return r, mw.ReturnPreCheckReverse()
}

func (plugin *PreCheckPlugin) preCheckGetTaskHandler(r *http.Request) (*http.Request, *mw.HttpResponse) {
	queryValues := r.URL.Query()
	project := queryValues.Get("project")

	_, statusCode, err := plugin.permitChecker.CheckProjectPermission(r.Context(), project,
		permitcheck.ProjectViewRSAction)
	if statusCode != http.StatusOK {
		return r, mw.ReturnErrorResponse(statusCode, fmt.Errorf("check project '%s' permission failed: %v",
			project, err))
	}

	return r, mw.ReturnPreCheckReverse()
}

func (plugin *PreCheckPlugin) preCheckCreateRecordHandler(r *http.Request) (*http.Request, *mw.HttpResponse) {
	req := make(map[string]interface{})
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest, errors.Wrapf(err, "read body failed"))
	}
	if err = json.Unmarshal(body, &req); err != nil {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest, errors.Wrapf(err, "unmarshal body failed"))
	}
	project := req["project"]
	projectStr, ok := project.(string)
	if !ok {
		return r, mw.ReturnErrorResponse(http.StatusBadRequest,
			fmt.Errorf("assert app to string failed:%v", project))
	}
	_, statusCode, err := plugin.permitChecker.CheckProjectPermission(r.Context(), projectStr,
		permitcheck.ProjectViewRSAction)
	if statusCode != http.StatusOK {
		return r, mw.ReturnErrorResponse(statusCode, fmt.Errorf("check project '%s' permission failed: %v",
			project, err))
	}
	r.Body = io.NopCloser(bytes.NewBuffer(body))
	length := len(body)
	r.Header.Set("Content-Length", strconv.Itoa(length))
	r.ContentLength = int64(length)
	return r, mw.ReturnPreCheckReverse()
}
