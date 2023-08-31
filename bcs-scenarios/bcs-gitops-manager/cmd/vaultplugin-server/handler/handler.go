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
	"net/http"

	"github.com/gorilla/mux"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/cmd/vaultplugin-server/secret"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/store"
)

// Handler interface for vaultplugin-server's api
type Handler interface {
	Init()
	Router() *mux.Router
}

// Options handler's options
type Options struct {
	Secret      secret.SecretManagerWithVersion
	GitopsStore store.Store
}

// V1VaultPluginHandler versions v1 for vaultplugin api
type V1VaultPluginHandler struct {
	router *mux.Router
	Opts   Options
}

// NewV1VaultPluginHandler new handler for V1VaultPluginHandler
func NewV1VaultPluginHandler(router *mux.Router, opts Options) *V1VaultPluginHandler {
	return &V1VaultPluginHandler{
		router: router,
		Opts:   opts,
	}
}

// Init initializes the plugin handler
func (h *V1VaultPluginHandler) Init() {
	h.router.HandleFunc("/health", h.healthy).Methods(http.MethodGet)
	h.router.HandleFunc("/decryption/manifest", h.descryptionManifest).Methods(http.MethodPost)

	h.router.HandleFunc("/secrets/{project}/{path}", h.createPutSecretHandler).Methods(http.MethodPost, http.MethodPut)
	h.router.HandleFunc("/secrets/{project}/{path}", h.deleteSecretHandler).Methods(http.MethodDelete)
	h.router.HandleFunc("/secrets/{project}/{path}", h.getSecretHandler).Methods(http.MethodGet).Queries("version", "{version}")

	// 这里因为path可能为空，无法在url中定义空字段，所以用parameter来做path
	h.router.HandleFunc("/secrets/{project}/list", h.listSecretHandler).Methods(http.MethodGet).Queries("path", "{path}")
	h.router.HandleFunc("/secrets/{project}/{path}/metadata", h.getMetadataHandler).Methods(http.MethodGet)
	h.router.HandleFunc("/secrets/{project}/{path}/version", h.getVersionHandler).Methods(http.MethodGet)
	h.router.HandleFunc("/secrets/{project}/{path}/rollback", h.rollbackHandler).Methods(http.MethodPost).Queries("version", "{version}")
	h.router.HandleFunc("/secrets/init", h.initHandler).Methods(http.MethodPost).Queries("project", "{project}")
	h.router.HandleFunc("/secrets/annotation", h.getSecretAnnotationHandler).Methods(http.MethodGet).Queries("project", "{project}")
}

// Router v1 vault routes
func (h *V1VaultPluginHandler) Router() *mux.Router {
	return h.router
}
