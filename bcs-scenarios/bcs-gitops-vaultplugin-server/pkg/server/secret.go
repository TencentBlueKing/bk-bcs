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

package server

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-vaultplugin-server/pkg/secret"
)

// POST,PUT /secrets/{project}/{path}, create new secret && update(create new version) new secret
func (s *Server) routerSaveOrUpdateSecret(w http.ResponseWriter, r *http.Request) {
	project := mux.Vars(r)["project"]
	rawPath := mux.Vars(r)["path"]
	path, err := url.PathUnescape(rawPath)
	if err != nil {
		s.responseError(r, w, http.StatusBadRequest, errors.Wrapf(err, "encode purl path failed"))
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		s.responseError(r, w, http.StatusBadRequest, errors.Wrapf(err, "read request body failed"))
		return
	}

	opt := &secret.SecretRequest{}
	if err = json.Unmarshal(body, opt); err != nil {
		s.responseError(r, w, http.StatusBadRequest,
			errors.Wrapf(err, "unmarshal request body failed: %s", string(body)))
		return
	}
	opt.Project = project
	opt.Path = path

	switch r.Method {
	case http.MethodPut:
		if err = s.secretManager.UpdateSecret(r.Context(), opt); err != nil {
			s.responseError(r, w, http.StatusInternalServerError,
				errors.Wrapf(err, "update secret for project '%s/%s' failed", opt.Project, path))
			return
		}
		s.responseSuccess(w, nil)
		return
	case http.MethodPost:
		if err = s.secretManager.CreateSecret(r.Context(), opt); err != nil {
			s.responseError(r, w, http.StatusInternalServerError,
				errors.Wrapf(err, "create secret for project '%s/%s' failed", opt.Project, path))
			return
		}
		s.responseSuccess(w, nil)
		return
	}
	return
}

func (s *Server) routerDeleteSecret(w http.ResponseWriter, r *http.Request) {
	project := mux.Vars(r)["project"]
	rawPath := mux.Vars(r)["path"]
	path, err := url.PathUnescape(rawPath)
	if err != nil {
		s.responseError(r, w, http.StatusBadRequest, errors.Wrapf(err, "encode url path failed"))
		return
	}

	opt := &secret.SecretRequest{
		Project: project,
		Path:    path,
	}
	err = s.secretManager.DeleteSecret(r.Context(), opt)
	if err != nil {
		s.responseError(r, w, http.StatusInternalServerError,
			errors.Wrapf(err, "delete secret '%s/%s' failed", opt.Project, path))
		return
	}
	s.responseSuccess(w, nil)
	return
}

func (s *Server) routerGetSecret(w http.ResponseWriter, r *http.Request) {
	project := mux.Vars(r)["project"]
	rawPath := mux.Vars(r)["path"]
	path, err := url.PathUnescape(rawPath)
	if err != nil {
		s.responseError(r, w, http.StatusBadRequest, errors.Wrapf(err, "encode url path failed"))
		return
	}

	opt := &secret.SecretRequest{
		Project: project,
		Path:    path,
	}
	version := r.URL.Query().Get("version")
	v, err := strconv.Atoi(version)
	if err != nil {
		s.responseError(r, w, http.StatusBadRequest,
			errors.Wrapf(err, "version not int format for '%s/%s'", opt.Project, path))
		return
	}

	sec, err := s.secretManager.GetSecretWithVersion(r.Context(), opt, v)
	if err != nil {
		s.responseError(r, w, http.StatusInternalServerError,
			errors.Wrapf(err, "get secret with version %d failed for '%s/%s'", v, opt.Project, opt.Path))
		return
	}
	s.responseSuccess(w, sec)
	return
}

func (s *Server) routerListSecret(w http.ResponseWriter, r *http.Request) {
	project := mux.Vars(r)["project"]
	path := r.URL.Query().Get("path")

	opt := &secret.SecretRequest{
		Project: project,
		Path:    path,
	}
	secretList, err := s.secretManager.ListSecret(r.Context(), opt)
	if err != nil {
		s.responseError(r, w, http.StatusInternalServerError,
			errors.Wrapf(err, "list secrets for project '%s/%s' failed", opt.Project, path))
		return
	}
	s.responseSuccess(w, secretList)
	return
}

func (s *Server) routerGetMetadata(w http.ResponseWriter, r *http.Request) {
	project := mux.Vars(r)["project"]
	rawPath := mux.Vars(r)["path"]
	path, err := url.PathUnescape(rawPath)
	if err != nil {
		s.responseError(r, w, http.StatusBadRequest, errors.Wrapf(err, "encode url path failed"))
		return
	}

	opt := &secret.SecretRequest{
		Project: project,
		Path:    path,
	}
	metadata, err := s.secretManager.GetMetadata(r.Context(), opt)
	if err != nil {
		s.responseError(r, w, http.StatusInternalServerError, errors.Wrapf(err,
			"get secret metadata for '%s/%s' failed", opt.Project, path))
		return
	}
	s.responseSuccess(w, metadata)
	return
}

func (s *Server) routerGetVersion(w http.ResponseWriter, r *http.Request) {
	project := mux.Vars(r)["project"]
	rawPath := mux.Vars(r)["path"]

	path, err := url.PathUnescape(rawPath)
	if err != nil {
		s.responseError(r, w, http.StatusBadRequest, errors.Wrapf(err, "encode url path failed"))
		return
	}

	opt := &secret.SecretRequest{
		Project: project,
		Path:    path,
	}
	version, err := s.secretManager.GetVersionsAsList(r.Context(), opt)
	if err != nil {
		s.responseError(r, w, http.StatusInternalServerError, errors.Wrapf(err,
			"get secret version failed for '%s/%s'", opt.Project, path))
		return
	}
	s.responseSuccess(w, version)
	return
}

func (s *Server) routerRollback(w http.ResponseWriter, r *http.Request) {
	project := mux.Vars(r)["project"]
	rawPath := mux.Vars(r)["path"]
	path, err := url.PathUnescape(rawPath)
	if err != nil {
		s.responseError(r, w, http.StatusBadRequest, errors.Wrapf(err, "encode url path failed"))
		return
	}

	opt := &secret.SecretRequest{
		Project: project,
		Path:    path,
	}
	version := r.URL.Query().Get("version")
	v, err := strconv.Atoi(version)
	if err != nil {
		s.responseError(r, w, http.StatusBadRequest,
			errors.Wrapf(err, "version %d not int format for '%s/%s'", version, opt.Project, path))
		return
	}

	err = s.secretManager.Rollback(r.Context(), opt, v)
	if err != nil {
		s.responseError(r, w, http.StatusInternalServerError,
			errors.Wrapf(err, "rollback secret for '%s/%s' failed", opt.Project, path))
		return
	}
	s.responseSuccess(w, nil)
	return
}

func (s *Server) routerInitProject(w http.ResponseWriter, r *http.Request) {
	project := mux.Vars(r)["project"]

	err := s.secretManager.InitProject(project)
	if err != nil {
		blog.Errorf("init project '%s' failed: %s", project, err.Error())
		if errs := s.secretManager.ReverseInitProject(project); len(errs) != 0 {
			for i := range errs {
				blog.Errorf("project '%s' reserve init failed: %s", project, errs[i].Error())
			}
		} else {
			blog.Warnf("project '%s' reverse init complete", project)
		}
		s.responseError(r, w, http.StatusInternalServerError,
			errors.Wrapf(err, "init secret for project %s failed", project))
		return
	}
	blog.Infof("init project '%s' success", project)
	s.responseSuccess(w, nil)
	return
}

func (s *Server) routerGetSecretAnnotation(w http.ResponseWriter, r *http.Request) {
	project := mux.Vars(r)["project"]

	anno := s.secretManager.GetSecretAnnotation(project)
	s.responseSuccess(w, anno)
	return
}
