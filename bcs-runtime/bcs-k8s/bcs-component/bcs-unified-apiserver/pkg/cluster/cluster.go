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

package cluster

import (
	"errors"
	"net/http"

	"go.uber.org/zap"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-unified-apiserver/pkg/cluster/federated"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-unified-apiserver/pkg/cluster/isolated"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-unified-apiserver/pkg/cluster/shared"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-unified-apiserver/pkg/config"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-unified-apiserver/pkg/rest"
	"github.com/gorilla/mux"
)

// Handler handler for http request
type Handler struct{}

func NewHandler() (*Handler, error) {
	return &Handler{}, nil
}

func ClusterFactory(clusterId string) (rest.Handler, error) {
	cluster, ok := config.G.GetMember(clusterId)
	if !ok {
		return nil, errors.New("invalid cluster")
	}

	var (
		handle rest.Handler
		err    error
	)

	switch cluster.Kind {
	case string(config.IsolatedCLuster):
		handle, err = isolated.NewHandler(cluster.Member)
	case string(config.SharedCluster):
		handle, err = shared.NewHandler(cluster.Member)
	case string(config.FederatedCluter):
		handle, err = federated.NewHandler(cluster.Master, cluster.Members)
	default:
		return nil, errors.New("not valid cluster kind")
	}
	return handle, err
}

// ServeHTTP serves http request
func (h *Handler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	zap.L().Info("receive request", zap.String("client", req.RemoteAddr), zap.String("method", req.Method), zap.String("path", req.URL.Path))

	vars := mux.Vars(req)
	clusterId := vars["cluster_id"]
	uri := vars["uri"]
	// rewrite url to k8s api path
	req.URL.Path = "/" + uri

	reqInfo, err := rest.NewRequestContext(rw, req)
	if err != nil {
		rest.AbortWithError(rw, err)
		return
	}

	handler, err := ClusterFactory(clusterId)
	if err != nil {
		rest.AbortWithError(rw, err)
		return
	}

	// Delete the original auth header so that the original user token won't be passed to the rev-proxy request and
	// damage the real cluster authentication process.
	delete(req.Header, "Authorization")
	handler.Serve(reqInfo)
}
