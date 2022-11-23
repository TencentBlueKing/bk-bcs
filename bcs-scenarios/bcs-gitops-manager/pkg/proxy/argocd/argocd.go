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

package argocd

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy"
)

// NewGitOpsProxy create proxy instance
func NewGitOpsProxy(opt *proxy.GitOpsOptions) proxy.GitOpsProxy {
	return &ArgocdProxy{
		option:  opt,
		Router:  mux.NewRouter(),
		session: &Session{option: opt},
	}
}

// ArgocdProxy simple revese proxy for argocd according kubernetes service.
// gitops proxy implements http.Handler interface.
type ArgocdProxy struct {
	*mux.Router // for http handler implementation

	option  *proxy.GitOpsOptions
	session *Session
}

// Init gitops essential session
func (ops *ArgocdProxy) Init() error {
	if ops.option == nil {
		return fmt.Errorf("lost GitOps options")
	}
	if err := ops.option.Validate(); err != nil {
		return err
	}
	ops.UseEncodedPath()
	if err := ops.initArgoPathHandler(); err != nil {
		return err
	}
	return nil
}

// Run ready to run
func (ops *ArgocdProxy) Run() {

}

// Stop proxy
func (ops *ArgocdProxy) Stop() {

}

// initArgoPathHandler
func (ops *ArgocdProxy) initArgoPathHandler() error {
	initializer := []func() error{}
	// project path
	projectRouter := ops.PathPrefix(common.GitOpsProxyURL + "/api/v1/projects").Subrouter()
	project := &ProjectPlugin{
		Router:  projectRouter,
		Session: ops.session,
		option:  ops.option,
	}
	initializer = append(initializer, project.Init)
	// cluster path
	clusterRouter := ops.PathPrefix(common.GitOpsProxyURL + "/api/v1/clusters").Subrouter()
	cluster := &ClusterPlugin{
		Router:  clusterRouter,
		Session: ops.session,
		option:  ops.option,
	}
	initializer = append(initializer, cluster.Init)
	// repository path
	repositoryRouter := ops.PathPrefix(common.GitOpsProxyURL + "/api/v1/repositories").Subrouter()
	repositoryRouter.UseEncodedPath()
	repository := &RepositoryPlugin{
		Router:  repositoryRouter,
		Session: ops.session,
		option:  ops.option,
	}
	initializer = append(initializer, repository.Init)
	// application path
	appRouter := ops.PathPrefix(common.GitOpsProxyURL + "/api/v1/applications").Subrouter()
	appPlugin := &AppPlugin{
		Router:  appRouter,
		Session: ops.session,
		option:  ops.option,
	}
	initializer = append(initializer, appPlugin.Init)

	// application stream path
	streamRouter := ops.PathPrefix(common.GitOpsProxyURL + "/api/v1/stream/applications").Subrouter()
	streamPlugin := &StreamPlugin{
		Router:     streamRouter,
		appHandler: appPlugin,
		Session:    ops.session,
		option:     ops.option,
	}
	initializer = append(initializer, streamPlugin.Init)

	// access deny URL, keep in mind that there are paths need to proxy
	ops.PathPrefix(common.GitOpsProxyURL + "/api/v1/account").HandlerFunc(http.NotFound)
	ops.PathPrefix(common.GitOpsProxyURL + "/api/v1/gpgkeys").HandlerFunc(http.NotFound)
	ops.PathPrefix(common.GitOpsProxyURL + "/api/v1/repocreds").HandlerFunc(http.NotFound)
	ops.PathPrefix(common.GitOpsProxyURL + "/api/v1/settings").HandlerFunc(http.NotFound)
	ops.PathPrefix(common.GitOpsProxyURL + "/api/v1/certificates").HandlerFunc(http.NotFound)

	// grpc access management
	grpcPlugin := &GrpcPlugin{
		Session: ops.session,
	}
	ops.HandleFunc(
		common.GitOpsProxyURL+"/{package:[a-z]+}.{service:[A-Z][a-zA-Z]+}/{method:[A-Z][a-zA-Z]+}",
		grpcPlugin.ServeHTTP,
	)

	for _, init := range initializer {
		if err := init(); err != nil {
			return err
		}
	}

	blog.Infof("Argocd proxy init successfully")
	return nil
}
