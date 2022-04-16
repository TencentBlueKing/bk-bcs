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

package proxy

import (
	"fmt"
	"net/http"
	"sync"

	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/util/proxy"
)

// Handler handler for http request
type Handler struct {
	handlerMapLock sync.Mutex
	defaultNs      string
	handlerMap     map[string]*proxy.UpgradeAwareHandler
}

// NewHandler create handler
func NewHandler(defaultNs string) (*Handler, error) {
	return &Handler{
		handlerMap: make(map[string]*proxy.UpgradeAwareHandler),
		defaultNs:  defaultNs,
	}, nil
}

// ServeHTTP serves http request
func (h *Handler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	zap.L().Info("receive request", zap.String("client", req.RemoteAddr),
		zap.String("method", req.Method), zap.String("path", req.URL.Path))

	// Delete the original auth header so that the original user token won't be passed to the rev-proxy request and
	// damage the real cluster authentication process.
	delete(req.Header, "Authorization")

	ns, err := getNamespaceFromRequest(req)
	if err != nil {
		zap.L().Error("get ns from request failed", zap.Error(err),
			zap.String("client", req.RemoteAddr), zap.String("path", req.URL.Path))
		h.backtoDefaultHandler(rw, req)
		return
	}
	if len(ns) == 0 {
		h.backtoDefaultHandler(rw, req)
		return
	}

	h.handlerMapLock.Lock()
	handler, ok := h.handlerMap[ns]
	h.handlerMapLock.Unlock()
	if !ok {
		http.Error(rw, fmt.Sprintf("no credential for ns %s", ns), http.StatusNotFound)
		return
	}

	handler.ServeHTTP(rw, req)
}

func (h *Handler) backtoDefaultHandler(rw http.ResponseWriter, req *http.Request) {
	h.handlerMapLock.Lock()
	defaultHandler, ok := h.handlerMap[h.defaultNs]
	h.handlerMapLock.Unlock()
	if !ok {
		http.Error(rw, "no credential for default kubeconfig", http.StatusInternalServerError)
		return
	}
	defaultHandler.ServeHTTP(rw, req)
}
