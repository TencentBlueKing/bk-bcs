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
	"github.com/gorilla/mux"
	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/internal/dao"
	mw "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy/argocd/middleware"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy/argocd/middleware/ctxutils"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy/argocd/permitcheck"
)

// SecretPlugin for internal project authorization
type SecretPlugin struct {
	*mux.Router
	middleware    mw.MiddlewareInterface
	permitChecker permitcheck.PermissionInterface
}

// all argocd secret URL:
// * required Project Edit projectPermission
// NOTE 因为密钥是敏感数据，所以get权限需要使用项目可编辑权限，后续可能需要可配置
// POST：   /api/v1/secrets/{project}/{path}，创建
// PUT：    /api/v1/secrets/{project}/{path}，指定更新(创建新版本)
// DELETE： /api/v1/secrets/{project}/{path}，完全销毁secret路径
// GET：/api/v1/secrets/{project}/{path}?version={version}，获取版本密钥secret信息
// GET：/api/v1/secrets/{project}/{path}/metadata，返回secret的metadata，包括版本信息以及创建修改时间等
// GET：/api/v1/secrets/{project}/{path}/version，返回secret的版本信息
// POST：/api/v1/secrets/{project}/{path}/rollback，secret版本回滚，会根据某个版本创建新版本

// * required Project View projectPermission
// GET：/api/v1/secrets/list?project={project}&path={path}，返回列表
//

// Init all project sub path handler
// project plugin is a subRouter, all path registered is relative
func (plugin *SecretPlugin) Init() error {
	// Create or Update(create new version) with preifx /api/v1/secrets/{project}/{path}
	plugin.Path("/{project}/{path}").Methods("POST", "PUT").
		Handler(plugin.middleware.HttpWrapper(plugin.createPutSecretHandler))
	// Delete with preifx /api/v1/secrets/{project}/{path}
	plugin.Path("/{project}/{path}").Methods("DELETE").
		Handler(plugin.middleware.HttpWrapper(plugin.deleteSecretHandler))
	// Get with preifx /api/v1/secrets/{project}/{path}?version={version}
	plugin.Path("/{project}/{path}").Methods("GET").
		Handler(plugin.middleware.HttpWrapper(plugin.getSecretHandler)).Queries("version", "{version}")

	// force check query, GET /api/v1/secrets/{project}/{path}/list
	// 这里因为path可能为空，无法在url中定义空字段，所以用parameter来做path
	plugin.Path("/{project}/list").Methods("GET").Queries("path", "{path}").
		Handler(plugin.middleware.HttpWrapper(plugin.listSecretHandler))
	// force check query, GET /api/v1/secrets/{project}/{path}/metadata
	plugin.Path("/{project}/{path}/metadata").Methods("GET").
		Handler(plugin.middleware.HttpWrapper(plugin.getMetadataHandler))
	// force check query, GET /api/v1/secrets/{project}/{path}/version
	plugin.Path("/{project}/{path}/version").Methods("GET").
		Handler(plugin.middleware.HttpWrapper(plugin.getVersionHandler))
	// force check query, POST /api/v1/secrets/{project}/{path}/rollback?version={version}
	plugin.Path("/{project}/{path}/rollback").Methods("POST").Queries("version", "{version}").
		Handler(plugin.middleware.HttpWrapper(plugin.rollbackHandler))

	blog.Infof("secret plugin init successfully")
	return nil
}

// POST,PUT /api/v1/secrets, create new secret && update(create new version) new secret
// validate project detail from request
func (plugin *SecretPlugin) createPutSecretHandler(r *http.Request) (*http.Request, *mw.HttpResponse) {
	project := mux.Vars(r)["project"]
	secretName := mux.Vars(r)["path"]
	_, statusCode, err := plugin.permitChecker.CheckProjectPermission(r.Context(), project,
		permitcheck.ProjectEditRSAction)
	if statusCode != http.StatusOK {
		return r, mw.ReturnErrorResponse(statusCode,
			errors.Wrapf(err, "check project '%s' edit permission failed", project))
	}
	var action ctxutils.AuditAction
	switch r.Method {
	case http.MethodPost:
		action = ctxutils.SecretCreate
	case http.MethodPut:
		action = ctxutils.SecretUpdate
	}
	r = plugin.setSecretAudit(r, project, secretName, action, ctxutils.EmptyData)
	return r, mw.ReturnSecretReverse()
}

// Delete with preifx /api/v1/secrets/{project}/{path}
func (plugin *SecretPlugin) deleteSecretHandler(r *http.Request) (*http.Request, *mw.HttpResponse) {
	projectName := mux.Vars(r)["project"]
	secretName := mux.Vars(r)["path"]
	_, statusCode, err := plugin.permitChecker.CheckProjectPermission(r.Context(), projectName,
		permitcheck.ProjectEditRSAction)
	if statusCode != http.StatusOK {
		return r, mw.ReturnErrorResponse(statusCode, errors.Wrapf(err, "check project permission failed"))
	}
	r = plugin.setSecretAudit(r, projectName, secretName, ctxutils.SecretDelete, ctxutils.EmptyData)
	return r, mw.ReturnSecretReverse()
}

// Get with preifx /api/v1/secrets/{project}/{path}?version={version}
func (plugin *SecretPlugin) getSecretHandler(r *http.Request) (*http.Request, *mw.HttpResponse) {
	projectName := mux.Vars(r)["project"]
	secretName := mux.Vars(r)["path"]
	_, statusCode, err := plugin.permitChecker.CheckProjectPermission(r.Context(), projectName,
		permitcheck.ProjectEditRSAction)
	if statusCode != http.StatusOK {
		return r, mw.ReturnErrorResponse(statusCode, errors.Wrapf(err, "check project permission failed"))
	}
	r = plugin.setSecretAudit(r, projectName, secretName, ctxutils.SecretView, ctxutils.EmptyData)
	return r, mw.ReturnSecretReverse()
}

// GET /api/v1/secrets/{project}/list？path=${path}
func (plugin *SecretPlugin) listSecretHandler(r *http.Request) (*http.Request, *mw.HttpResponse) {
	projectName := mux.Vars(r)["project"]

	_, statusCode, err := plugin.permitChecker.CheckProjectPermission(r.Context(), projectName,
		permitcheck.ProjectViewRSAction)
	if statusCode != http.StatusOK {
		return r, mw.ReturnErrorResponse(statusCode, errors.Wrapf(err, "check project permission failed"))
	}

	return r, mw.ReturnSecretReverse()
}

// GET /api/v1/secrets/{project}/{path}/metadata
func (plugin *SecretPlugin) getMetadataHandler(r *http.Request) (*http.Request, *mw.HttpResponse) {
	projectName := mux.Vars(r)["project"]

	_, statusCode, err := plugin.permitChecker.CheckProjectPermission(r.Context(), projectName,
		permitcheck.ProjectEditRSAction)
	if statusCode != http.StatusOK {
		return r, mw.ReturnErrorResponse(statusCode, errors.Wrapf(err, "check project permission failed"))
	}
	return r, mw.ReturnSecretReverse()
}

// GET /api/v1/secrets/{project}/{path}/version
func (plugin *SecretPlugin) getVersionHandler(r *http.Request) (*http.Request, *mw.HttpResponse) {
	projectName := mux.Vars(r)["project"]

	_, statusCode, err := plugin.permitChecker.CheckProjectPermission(r.Context(), projectName,
		permitcheck.ProjectEditRSAction)
	if statusCode != http.StatusOK {
		return r, mw.ReturnErrorResponse(statusCode, errors.Wrapf(err, "check project permission failed"))
	}

	return r, mw.ReturnSecretReverse()
}

// POST /api/v1/secrets/{project}/{path}/rollback
func (plugin *SecretPlugin) rollbackHandler(r *http.Request) (*http.Request, *mw.HttpResponse) {
	projectName := mux.Vars(r)["project"]
	secretName := mux.Vars(r)["path"]
	_, statusCode, err := plugin.permitChecker.CheckProjectPermission(r.Context(), projectName,
		permitcheck.ProjectEditRSAction)
	if statusCode != http.StatusOK {
		return r, mw.ReturnErrorResponse(statusCode, errors.Wrapf(err, "check project permission failed"))
	}
	r = plugin.setSecretAudit(r, projectName, secretName, ctxutils.SecretRollback, ctxutils.EmptyData)
	return r, mw.ReturnSecretReverse()
}

func (plugin *SecretPlugin) setSecretAudit(r *http.Request, project, secretName string,
	action ctxutils.AuditAction, data string) *http.Request {
	return ctxutils.SetAuditMessage(r, &dao.UserAudit{
		Project:      project,
		Action:       string(action),
		ResourceType: string(ctxutils.SecretResource),
		ResourceName: secretName,
		ResourceData: data,
		RequestType:  string(ctxutils.HTTPRequest),
	})
}
