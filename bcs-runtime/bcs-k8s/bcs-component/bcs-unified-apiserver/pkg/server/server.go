/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package server

import (
	"net/http"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-unified-apiserver/pkg/cluster"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	v1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/json"
)

// Handler handler for http request
type Handler struct{}

func NewHandler() (*Handler, error) {
	return &Handler{}, nil
}

// ServeHTTP serves http request
func (h *Handler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	zap.L().Info("receive request", zap.String("client", req.RemoteAddr), zap.String("method", req.Method), zap.String("path", req.URL.Path))

	vars := mux.Vars(req)
	clusterId := vars["clusterId"]
	uri := vars["uri"]

	reqInfo, err := getNamespaceFromRequest(req)
	if err != nil {
		rw.Header().Set("Content-Type", "application/json; charset=utf-8")
		rw.Header().Set("Cache-Control", "no-cache, no-store")
		result := apierrors.NewNotFound(v1.Resource("secrets"), req.URL.Path)
		json.NewEncoder(rw).Encode(result)
		rw.WriteHeader(int(result.ErrStatus.Code))
		return
	}

	handler, err := cluster.ClusterFactory(clusterId, reqInfo, uri)
	if err != nil {
		rw.Header().Set("Content-Type", "application/json; charset=utf-8")
		rw.Header().Set("Cache-Control", "no-cache, no-store")
		result := apierrors.NewNotFound(v1.Resource("secrets"), req.URL.Path)
		json.NewEncoder(rw).Encode(result)
		rw.WriteHeader(int(result.ErrStatus.Code))
		return
	}

	// Delete the original auth header so that the original user token won't be passed to the rev-proxy request and
	// damage the real cluster authentication process.
	delete(req.Header, "Authorization")

	handler.ServeHTTP(rw, req)
}
