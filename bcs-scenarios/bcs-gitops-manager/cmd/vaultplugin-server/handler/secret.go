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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/cmd/vaultplugin-server/secret"
	"github.com/gorilla/mux"
)

const (
	ErrHttpCode     = 1
	SuccessHttpCode = 0
)

// SecretResponse secret response for vaultplugin-server
type SecretResponse struct {
	// code 为 0 表示请求是正常的
	// code 为 1 表示请求不正常的
	Code int `json:"code"`

	// 如果 code 为 0，这里 message 为空
	// 如果 code 不为 0，这里 message 应该是出错内容
	Message string `json:"message"`

	// code 为 0，如果有需要返回对象，这里返回对应对象
	// code 为 0，不需要返回对象这里就可以为空（比如删除的场景）
	// code 不为 0，这个字段一般就不设置
	Data interface{} `json:"data"`
}

// Response response function for secret
func (r *SecretResponse) Response(w http.ResponseWriter, code int) {
	r.Code = code
	b, err := json.Marshal(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}

// POST,PUT /secrets/{project}/{path}, create new secret && update(create new version) new secret
func (v1 *V1VaultPluginHandler) createPutSecretHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	resp := &SecretResponse{}

	project := mux.Vars(r)["project"]
	rawPath := mux.Vars(r)["path"]
	path, err := url.PathUnescape(rawPath)
	if err != nil {
		resp.Message = "[createPutSecret] encode url path error."
		blog.Errorf("[createPutSecret] encode url path error, error: %v", err)
		resp.Response(w, ErrHttpCode)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		resp.Message = "[createPutSecret] read body failed"
		blog.Errorf("[createPutSecret] read body failed, error: %v", err)
		resp.Response(w, ErrHttpCode)
		return
	}

	opt := &secret.SecretRequest{}
	if err = json.Unmarshal(body, opt); err != nil {
		resp.Message = "[createPutSecret] body json unmarshal failed"
		blog.Errorf("[createPutSecret] body json unmarshal failed, body: %s, error: %v", string(body), err)
		resp.Response(w, ErrHttpCode)
		return
	}
	opt.Project = project
	opt.Path = path

	switch r.Method {
	case http.MethodPut:
		err = v1.Opts.Secret.UpdateSecret(ctx, opt)
		if err != nil {
			resp.Message = fmt.Sprintf("put Secrets by project '%s' and path '%s' failed, error: %v", opt.Project, path, err)
			blog.Errorf("put Secrets by project '%s' and path '%s' failed, error: %v", opt.Project, path, err)
			resp.Response(w, ErrHttpCode)
			return
		}
		resp.Response(w, SuccessHttpCode)
		return
	case http.MethodPost:
		err = v1.Opts.Secret.CreateSecret(ctx, opt)
		if err != nil {
			resp.Message = fmt.Sprintf("create Secrets by project '%s' from secret failed, error: %v", opt.Project, err)
			blog.Errorf("create Secrets by project '%s' from secret failed, error: %v", opt.Project, err)
			resp.Response(w, ErrHttpCode)
			return
		}
		resp.Response(w, SuccessHttpCode)
		return
	}

	resp.Message = fmt.Sprintf("http method %s not in [POST, PUT]", r.Method)
	resp.Response(w, ErrHttpCode)
	return
}

// DELETE /secrets/{project}/{path}
func (v1 *V1VaultPluginHandler) deleteSecretHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	resp := &SecretResponse{}

	project := mux.Vars(r)["project"]
	rawPath := mux.Vars(r)["path"]
	path, err := url.PathUnescape(rawPath)
	if err != nil {
		resp.Message = "[deleteSecret] encode url path error."
		blog.Errorf("[deleteSecret] encode url path error, error: %v", err)
		resp.Response(w, ErrHttpCode)
		return
	}

	opt := &secret.SecretRequest{
		Project: project,
		Path:    path,
	}
	err = v1.Opts.Secret.DeleteSecret(ctx, opt)
	if err != nil {
		resp.Message = fmt.Sprintf("delete secret by project '%s' and path '%s' failed, error: %v", opt.Project, path, err)
		blog.Errorf("delete Secrets by project '%s' and path '%s' failed, error: %v", opt.Project, path, err)
		resp.Response(w, ErrHttpCode)
		return
	}
	resp.Response(w, SuccessHttpCode)
	return
}

// Get /api/v1/secrets/{project}/{path}?version={version}
func (v1 *V1VaultPluginHandler) getSecretHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	resp := &SecretResponse{}

	project := mux.Vars(r)["project"]
	rawPath := mux.Vars(r)["path"]
	path, err := url.PathUnescape(rawPath)
	if err != nil {
		resp.Message = "[getSecret] encode url path error."
		blog.Errorf("[getSecret] encode url path error, error: %v", err)
		resp.Response(w, ErrHttpCode)
		return
	}

	opt := &secret.SecretRequest{
		Project: project,
		Path:    path,
	}
	version := r.URL.Query().Get("version")
	v, err := strconv.Atoi(version)
	if err != nil {
		resp.Message = fmt.Sprintf("version not int format by project '%s' and path '%s', error: %v", opt.Project, path, err)
		blog.Errorf("version not int format by project '%s' and path '%s', error: %v", opt.Project, path, err)
		resp.Response(w, ErrHttpCode)
		return
	}

	sec, err := v1.Opts.Secret.GetSecretWithVersion(ctx, opt, v)
	if err != nil {
		resp.Message = fmt.Sprintf("get secret by project '%s' and path '%s' failed, error: %v", opt.Project, path, err)
		blog.Errorf("get Secrets by project '%s' and path '%s' failed, error: %v", opt.Project, path, err)
		resp.Response(w, ErrHttpCode)
		return
	}

	resp.Data = sec
	resp.Response(w, SuccessHttpCode)
	return
}

// GET /api/v1/secrets/{project}/list？path=${path}
func (v1 *V1VaultPluginHandler) listSecretHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	resp := &SecretResponse{}

	project := mux.Vars(r)["project"]
	path := r.URL.Query().Get("path")

	opt := &secret.SecretRequest{
		Project: project,
		Path:    path,
	}
	secretList, err := v1.Opts.Secret.ListSecret(ctx, opt)
	if err != nil {
		resp.Message = fmt.Sprintf("list secret by project '%s' and path '%s' failed, error: %v", opt.Project, path, err)
		blog.Errorf("list Secrets by project '%s' and path '%s' failed, error: %v", opt.Project, path, err)
		resp.Response(w, ErrHttpCode)
		return
	}

	resp.Data = secretList
	resp.Response(w, SuccessHttpCode)
	return
}

// GET /api/v1/secrets/{project}/{path}/metadata
func (v1 *V1VaultPluginHandler) getMetadataHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	resp := &SecretResponse{}

	project := mux.Vars(r)["project"]
	rawPath := mux.Vars(r)["path"]
	path, err := url.PathUnescape(rawPath)
	if err != nil {
		resp.Message = "[getMetadata] encode url path error."
		blog.Errorf("[getMetadata] encode url path error, error: %v", err)
		resp.Response(w, ErrHttpCode)
		return
	}

	opt := &secret.SecretRequest{
		Project: project,
		Path:    path,
	}
	metadata, err := v1.Opts.Secret.GetMetadata(ctx, opt)
	if err != nil {
		resp.Message = fmt.Sprintf("get secret metadata by project '%s' and path '%s' failed, error: %v", opt.Project, path, err)
		blog.Errorf("get secret metadata by project '%s' and path '%s' failed, error: %v", opt.Project, path, err)
		resp.Response(w, ErrHttpCode)
		return
	}

	resp.Data = metadata
	resp.Response(w, SuccessHttpCode)
	return
}

// GET /api/v1/secrets/{project}/{path}/version
func (v1 *V1VaultPluginHandler) getVersionHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	resp := &SecretResponse{}
	project := mux.Vars(r)["project"]
	rawPath := mux.Vars(r)["path"]

	path, err := url.PathUnescape(rawPath)
	if err != nil {
		resp.Message = "[getVersion] encode url path error."
		blog.Errorf("[getVersion] encode url path error, error: %v", err)
		resp.Response(w, ErrHttpCode)
		return
	}

	opt := &secret.SecretRequest{
		Project: project,
		Path:    path,
	}
	version, err := v1.Opts.Secret.GetVersionsAsList(ctx, opt)
	if err != nil {
		resp.Message = fmt.Sprintf("get secret metadata by project '%s' and path '%s' failed, error: %v", opt.Project, path, err)
		blog.Errorf("get secret metadata by project '%s' and path '%s' failed, error: %v", opt.Project, path, err)
		resp.Response(w, ErrHttpCode)
		return
	}

	resp.Data = version
	resp.Response(w, SuccessHttpCode)

	return
}

// POST /api/v1/secrets/{project}/{path}/rollback?version={version}
func (v1 *V1VaultPluginHandler) rollbackHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	resp := &SecretResponse{}

	project := mux.Vars(r)["project"]
	rawPath := mux.Vars(r)["path"]
	path, err := url.PathUnescape(rawPath)
	if err != nil {
		resp.Message = "[rollback] encode url path error."
		blog.Errorf("[rollback] encode url path error, error: %v", err)
		resp.Response(w, ErrHttpCode)
		return
	}

	opt := &secret.SecretRequest{
		Project: project,
		Path:    path,
	}

	version := r.URL.Query().Get("version")
	v, err := strconv.Atoi(version)
	if err != nil {
		resp.Message = fmt.Sprintf("version not int format by project '%s' and path '%s', error: %v", opt.Project, path, err)
		blog.Errorf("version not int format by project '%s' and path '%s', error: %v", opt.Project, path, err)
		resp.Response(w, ErrHttpCode)
		return
	}

	err = v1.Opts.Secret.Rollback(ctx, opt, v)
	if err != nil {
		resp.Message = fmt.Sprintf("rollback secret by project '%s' and path '%s' failed, error: %v", opt.Project, path, err)
		blog.Errorf("rollback secret by project '%s' and path '%s' failed, error: %v", opt.Project, path, err)
		resp.Response(w, ErrHttpCode)
		return
	}

	resp.Response(w, SuccessHttpCode)
	return
}

// POST /api/v1/secrets/init？project={project}
func (v1 *V1VaultPluginHandler) initHandler(w http.ResponseWriter, r *http.Request) {
	resp := &SecretResponse{}
	project := mux.Vars(r)["project"]

	err := v1.Opts.Secret.InitProject(project)
	if err != nil {
		resp.Message = fmt.Sprintf("initProject secret by project '%s' failed, error: %v", project, err)
		blog.Errorf("initProject secret by project '%s' failed, error: %v", project, err)
		resp.Response(w, ErrHttpCode)
		return
	}

	resp.Response(w, SuccessHttpCode)
	return
}

// GET /api/v1/secrets/annotation？project={project}
func (v1 *V1VaultPluginHandler) getSecretAnnotationHandler(w http.ResponseWriter, r *http.Request) {
	resp := &SecretResponse{}
	project := mux.Vars(r)["project"]

	anno := v1.Opts.Secret.GetSecretAnnotation(project)

	resp.Data = anno
	resp.Response(w, SuccessHttpCode)
	return
}
