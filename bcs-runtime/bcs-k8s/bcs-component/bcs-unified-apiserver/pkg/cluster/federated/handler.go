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

package federated

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-unified-apiserver/pkg/clientutil"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/util/proxy"
)

type Handler struct {
	clusterId    string
	proxyHandler *proxy.UpgradeAwareHandler
}

// NewHandler create handler
func NewHandler(clusterId string) (*Handler, error) {
	kubeConf, err := clientutil.GetKubeConfByClusterId(clusterId)
	if err != nil {
		return nil, fmt.Errorf("build proxy handler from config %s failed, err %s", kubeConf.String(), err.Error())
	}

	proxyHandler, err := NewProxyHandlerFromConfig(kubeConf)
	if err != nil {
		return nil, fmt.Errorf("build proxy handler from config %s failed, err %s", kubeConf.String(), err.Error())
	}

	return &Handler{
		clusterId:    clusterId,
		proxyHandler: proxyHandler,
	}, nil
}

// ServeHTTP serves http request
func (h *Handler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	zap.L().Info("receive request", zap.String("client", req.RemoteAddr), zap.String("method", req.Method), zap.String("path", req.URL.Path))

	vars := mux.Vars(req)
	fmt.Println("vars", vars)

	namespace, ok := vars["namespace"]
	if ok {
		stor, err := NewPodStor([]string{})
		listOptions, err := clientutil.GetListOptionsFromQueryParam(req.URL.Query())
		if err != nil {
			fmt.Println(err)
		}
		acceptHeader := req.Header.Get("Accept")
		if strings.Contains(acceptHeader, "as=Table") {
			result, err := stor.ListAsTable(req.Context(), namespace, listOptions, acceptHeader)
			if err == nil {
				rw.Header().Set("Content-Type", "application/json; charset=utf-8")
				rw.Header().Set("Cache-Control", "no-cache, no-store")
				if err := json.NewEncoder(rw).Encode(result); err != nil {
					rw.WriteHeader(http.StatusOK)
				} else {
					rw.WriteHeader(http.StatusInternalServerError)
				}
				return
			}
		} else {
			result, err := stor.List(req.Context(), namespace, listOptions)
			if err == nil {
				rw.Header().Set("Content-Type", "application/json; charset=utf-8")
				rw.Header().Set("Cache-Control", "no-cache, no-store")
				if err := json.NewEncoder(rw).Encode(result); err != nil {
					rw.WriteHeader(http.StatusOK)
				} else {
					rw.WriteHeader(http.StatusInternalServerError)
				}
				return
			}
		}
	}

	// Delete the original auth header so that the original user token won't be passed to the rev-proxy request and
	// damage the real cluster authentication process.
	delete(req.Header, "Authorization")

	h.proxyHandler.ServeHTTP(rw, req)
}
