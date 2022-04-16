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
	"strings"

	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/util/proxy"
	"k8s.io/client-go/rest"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-unified-apiserver/pkg/config"
)

// Handler handler for http request
type Handler struct {
	memberId     string
	proxyHandler *proxy.UpgradeAwareHandler
}

// NewHandler create handler
func NewHandler(memberId string) (*Handler, error) {
	kubeConf, err := GetKubeConfByClusterId(memberId)
	if err != nil {
		return nil, fmt.Errorf("build proxy handler from config %s failed, err %s", kubeConf.String(), err.Error())
	}

	proxyHandler, err := NewProxyHandlerFromConfig(kubeConf)
	if err != nil {
		return nil, fmt.Errorf("build proxy handler from config %s failed, err %s", kubeConf.String(), err.Error())
	}

	return &Handler{
		memberId:     memberId,
		proxyHandler: proxyHandler,
	}, nil
}

// GetEnvByClusterId 获取集群所属环境, 目前通过集群ID前缀判断
func GetEnvByClusterId(clusterId string) config.BCSClusterEnv {
	if strings.HasPrefix(clusterId, "BCS-K8S-1") {
		return config.UatCluster
	}
	if strings.HasPrefix(clusterId, "BCS-K8S-2") {
		return config.DebugCLuster
	}
	if strings.HasPrefix(clusterId, "BCS-K8S-4") {
		return config.ProdEnv
	}
	return config.ProdEnv
}

// GetK8SClientByClusterId 通过集群 ID 获取 k8s client 对象
func GetKubeConfByClusterId(clusterId string) (*rest.Config, error) {
	bcsConf := GetBCSConfByClusterId(clusterId)
	host := fmt.Sprintf("%s/clusters/%s", bcsConf.Host, clusterId)
	config := &rest.Config{
		Host:        host,
		BearerToken: bcsConf.Token,
	}

	return config, nil
}

// GetBCSConfByClusterId 通过集群ID, 获取不同admin token 信息
func GetBCSConfByClusterId(clusterId string) *config.BCSConf {
	env := GetEnvByClusterId(clusterId)
	conf, ok := config.G.BCSEnvMap[env]
	if ok {
		return conf
	}
	// 默认返回bcs配置
	return config.G.BCS
}

// ServeHTTP serves http request
func (h *Handler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	zap.L().Info("receive request", zap.String("client", req.RemoteAddr),
		zap.String("method", req.Method), zap.String("path", req.URL.Path))

	// Delete the original auth header so that the original user token won't be passed to the rev-proxy request and
	// damage the real cluster authentication process.
	delete(req.Header, "Authorization")

	h.proxyHandler.ServeHTTP(rw, req)
}
