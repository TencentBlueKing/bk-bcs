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

// Package argocd xxx
package argocd

import (
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/jwt"
	bcsapi "github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapiv4"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/cmd/manager/options"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/internal/dao"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy/argocd/analyze"
	mw "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy/argocd/middleware"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/store"
)

// NewGitOpsProxy create proxy instance
func NewGitOpsProxy() proxy.GitOpsProxy {
	return &ArgocdProxy{
		option: options.GlobalOptions(),
		Router: mux.NewRouter(),
	}
}

// ArgocdProxy simple revese proxy for argocd according kubernetes service.
// gitops proxy implements http.Handler interface.
// nolint
type ArgocdProxy struct {
	*mux.Router // for http handler implementation

	option *options.Options
	// JWTClient for authentication
	JWTDecoder *jwt.JWTClient
	// IAMClient is basic client
	IAMClient iam.PermClient
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
// nolint
func (ops *ArgocdProxy) initArgoPathHandler() error {
	middleware := mw.NewMiddlewareHandler()
	if err := middleware.Init(); err != nil {
		return errors.Wrapf(err, "middleware init failed")
	}
	bcsStorage := bcsapi.NewStorage(&bcsapi.Config{
		Hosts:     []string{ops.option.APIGatewayForCluster},
		AuthToken: ops.option.APIGatewayToken,
		// just a fake tls config
		TLSConfig: &tls.Config{},
		Gateway:   true,
	})

	projectPlugin := &ProjectPlugin{
		Router:         ops.PathPrefix(common.GitOpsProxyURL + "/api/v1/projects").Subrouter(),
		middleware:     middleware,
		analysisClient: analyze.NewCollectApplication(),
	}
	clusterPlugin := &ClusterPlugin{
		Router:     ops.PathPrefix(common.GitOpsProxyURL + "/api/v1/clusters").Subrouter(),
		middleware: middleware,
	}
	repositoryRouter := ops.PathPrefix(common.GitOpsProxyURL + "/api/v1/repositories").Subrouter()
	repositoryRouter.UseEncodedPath()
	repositoryPlugin := &RepositoryPlugin{
		Router:     repositoryRouter,
		middleware: middleware,
	}
	appPlugin := &AppPlugin{
		storage:    store.GlobalStore(),
		db:         dao.GlobalDB(),
		Router:     ops.PathPrefix(common.GitOpsProxyURL + "/api/v1/applications").Subrouter(),
		middleware: middleware,
		appCollect: analyze.NewCollectApplication(),
		bcsStorage: bcsStorage,
	}
	appsetPlugin := &ApplicationSetPlugin{
		storage:    store.GlobalStore(),
		Router:     ops.PathPrefix(common.GitOpsProxyURL + "/api/v1/applicationsets").Subrouter(),
		middleware: middleware,
	}
	secretPlugin := &SecretPlugin{
		Router:     ops.PathPrefix(common.GitOpsProxyURL + "/api/v1/secrets").Subrouter(),
		middleware: middleware,
	}
	streamPlugin := &StreamPlugin{
		Router:     ops.PathPrefix(common.GitOpsProxyURL + "/api/v1/stream/applications").Subrouter(),
		appHandler: appPlugin,
		middleware: middleware,
		storage:    store.GlobalStore(),
		bcsStorage: bcsStorage,
	}
	webhookPlugin := &WebhookPlugin{
		Router:        ops.PathPrefix(common.GitOpsProxyURL + "/api/webhook").Subrouter(),
		middleware:    middleware,
		appsetWebhook: ops.option.GitOps.AppsetControllerWebhook,
	}
	// grpc access handler
	grpcPlugin := &GrpcPlugin{
		Router: ops.NewRoute().Path(common.GitOpsProxyURL +
			"/{package:[a-z]+}.{service:[A-Z][a-zA-Z]+}/{method:[A-Z][a-zA-Z]+}").Subrouter(),
		middleware: middleware,
	}
	metricPlugin := &MetricPlugin{
		Router:     ops.PathPrefix(common.GitOpsProxyURL + "/api/metric").Subrouter(),
		middleware: middleware,
	}
	analysisPlugin := &AnalysisPlugin{
		Router:         ops.PathPrefix(common.GitOpsProxyURL + "/api/analysis").Subrouter(),
		middleware:     middleware,
		analysisClient: analyze.NewAnalysisOverview(),
		store:          store.GlobalStore(),
	}
	monitorPlugin := &MonitorPlugin{
		Router:     ops.PathPrefix(common.GitOpsProxyURL + "/api/v1/monitor").Subrouter(),
		middleware: middleware,
	}
	initializer := []func() error{
		projectPlugin.Init, clusterPlugin.Init, repositoryPlugin.Init,
		appPlugin.Init, streamPlugin.Init, webhookPlugin.Init, grpcPlugin.Init,
		secretPlugin.Init, metricPlugin.Init, appsetPlugin.Init, analysisPlugin.Init,
		monitorPlugin.Init,
	}

	// access deny URL, keep in mind that there are paths need to proxy
	ops.PathPrefix(common.GitOpsProxyURL + "/api/v1/account").HandlerFunc(http.NotFound)
	ops.PathPrefix(common.GitOpsProxyURL + "/api/v1/gpgkeys").HandlerFunc(http.NotFound)
	ops.PathPrefix(common.GitOpsProxyURL + "/api/v1/repocreds").HandlerFunc(http.NotFound)
	ops.PathPrefix(common.GitOpsProxyURL + "/api/v1/settings").HandlerFunc(http.NotFound)
	ops.PathPrefix(common.GitOpsProxyURL + "/api/v1/certificates").HandlerFunc(http.NotFound)
	for _, init := range initializer {
		if err := init(); err != nil {
			return err
		}
	}
	blog.Infof("Argocd proxy init successfully")
	return nil
}
