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

package handler

import (
	"fmt"
	"net/http"
	"path"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/gorilla/mux"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/auth"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/repo"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/utils/contextx"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/utils/httpx"
)

// NewSwaggerRouter swagger handler
func NewSwaggerRouter(opt *options.HelmManagerOptions) *mux.Router {
	r := mux.NewRouter()
	// swagger doc
	if len(opt.Swagger.Dir) != 0 {
		blog.Info("swagger doc is enabled")
		r.HandleFunc("/{uri:.*}", func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, path.Join(opt.Swagger.Dir, strings.TrimPrefix(r.URL.Path, "/helmmanager/swagger/")))
		})
	}
	return r
}

// NewAPIRouter http handler
func NewAPIRouter(hm *HelmManager) *mux.Router {
	r := mux.NewRouter()
	// add middleware
	r.Use(httpx.LoggingMiddleware)
	r.Use(httpx.AuthenticationMiddleware)
	r.Use(httpx.ParseProjectIDMiddleware)
	r.Use(httpx.AuthorizationMiddleware)
	r.Use(httpx.CheckUserResourceTenantMiddleware)
	r.Use(httpx.AuditMiddleware)

	// chart upload
	r.Methods("POST").Path("/helmmanager/api/v1/projects/{projectCode}/repos/{repoName}/charts/upload").
		HandlerFunc(UploadChartHandler(hm))
	return r
}

// UploadChartHandler upload chart handler
func UploadChartHandler(hm *HelmManager) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		ctx := r.Context()
		projectCode := contextx.GetProjectCodeFromCtx(ctx)
		repoName := vars["repoName"]
		version := r.URL.Query().Get("version")
		force := r.URL.Query().Get("force")

		f, _, err := r.FormFile("chart")
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to get file 'attachment': %s", err.Error()), http.StatusBadRequest)
			return
		}
		defer f.Close()

		// 获取仓库上传地址和账号密码
		repository, err := hm.model.GetProjectRepository(ctx, projectCode, repoName)
		if err != nil {
			httpx.ResponseSystemError(w, r, err)
			return
		}

		// check repo auth
		if (repository.Personal && repository.CreateBy != auth.GetUserFromCtx(ctx)) ||
			repoName == common.PublicRepoName {
			httpx.ResponseAuthError(w, r, fmt.Errorf("you have no permission to upload chart to this repo"))
			return
		}

		err = hm.platform.
			User(repo.User{
				Name:     repository.Username,
				Password: repository.Password,
			}).
			Project(repository.GetRepoProjectID()).
			Repository(
				repo.GetRepositoryType(repository.Type),
				repository.GetRepoName(),
			).
			UploadChart(ctx, repo.UploadOption{
				Content: f,
				Version: version,
				Force:   force == "true",
			})
		if err != nil {
			httpx.ResponseSystemError(w, r, err)
			return
		}

		httpx.ResponseOK(w, r, nil)
	}
}
