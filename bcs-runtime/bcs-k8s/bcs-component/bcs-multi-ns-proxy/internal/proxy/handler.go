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
	"k8s.io/client-go/tools/clientcmd"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-multi-ns-proxy/pkg/filewatcher"
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

// OnEvent implements fielwatcher.Handler
func (h *Handler) OnEvent(e filewatcher.Event) error {
	switch e.Type {
	case filewatcher.EventAdd, filewatcher.EventUpdate:
		config, err := clientcmd.RESTConfigFromKubeConfig([]byte(e.Content))
		if err != nil {
			return fmt.Errorf("build config from %s: %s failed, err %s", e.Filename, e.Content, err.Error())
		}
		proxyHandler, err := NewProxyHandlerFromConfig(config)
		if err != nil {
			return fmt.Errorf("build proxy handler from config %s failed, err %s", config.String(), err.Error())
		}
		h.handlerMapLock.Lock()
		h.handlerMap[e.Filename] = proxyHandler
		h.handlerMapLock.Unlock()
		zap.L().Info("ns updated successfully", zap.String("ns", e.Filename))
		return nil

	case filewatcher.EventDelete:
		if _, ok := h.handlerMap[e.Filename]; ok {
			h.handlerMapLock.Lock()
			delete(h.handlerMap, e.Filename)
			h.handlerMapLock.Unlock()
		}
		zap.L().Info("ns deleted", zap.String("ns", e.Filename))
		return nil

	default:
		return fmt.Errorf("no support event %v", e)
	}
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
