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

package envinronment

import (
	"io/ioutil"
	"net/http"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-argocd-manager/bcs-argocd-server/internal/options"
)

// NewProxyHandler return a new proxyHandler instance
func NewProxyHandler(opt *options.ArgocdServerOptions) *proxyHandler {
	cls := make(map[string]*cluster)
	for _, e := range opt.Environments {
		for _, c := range e.Clusters {
			cls[c.ClusterID] = &cluster{
				clusterID: c.ClusterID,
				project:   c.Project,
				apiServer: e.APIServer,
				token:     e.Token,
			}
		}
	}
	return &proxyHandler{
		clusters: cls,
	}
}

type proxyHandler struct {
	clusters map[string]*cluster
}

// ServeHTTP implements http.Handler
func (ph *proxyHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	body, _ := ioutil.ReadAll(req.Body)

	blog.Infof("headers get %v", req.Header)
	blog.Infof("bodys get %s", string(body))

	rw.WriteHeader(400)
}

type cluster struct {
	clusterID string
	project   string
	apiServer string
	token     string
}
