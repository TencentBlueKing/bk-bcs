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
	"context"
	"net/http"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	monitoring "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/metric"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy"
	mw "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy/argocd/middleware"
)

// MetricPlugin defines the metric plugin to proxy all the metrics
type MetricPlugin struct {
	*mux.Router
	middleware mw.MiddlewareInterface

	monitorClient *monitoring.Clientset
	k8sClient     *kubernetes.Clientset
}

// Init will init the metric proxy
func (plugin *MetricPlugin) Init() error {
	plugin.Path("/{namespace}/{servicemonitor}").Methods("GET").
		Handler(plugin.middleware.HttpWrapper(plugin.metric))
	if err := plugin.inClusterClient(); err != nil {
		return errors.Wrapf(err, "init metric proxy plugin failed")
	}
	blog.Infof("metric proxy plugin init successfully")
	return nil
}

func (plugin *MetricPlugin) inClusterClient() error {
	config, err := rest.InClusterConfig()
	if err != nil {
		return errors.Wrapf(err, "get k8s incluster config failed")
	}
	plugin.monitorClient, err = monitoring.NewForConfig(config)
	if err != nil {
		return errors.Wrapf(err, "create prometheus client failed")
	}
	plugin.k8sClient, err = kubernetes.NewForConfig(config)
	if err != nil {
		return errors.Wrapf(err, "create k8s lient failed")
	}
	return nil
}

func (plugin *MetricPlugin) metric(r *http.Request) (*http.Request, *mw.HttpResponse) {
	namespace, smName, resp := plugin.parseParam(r.Context(), r)
	if resp != nil {
		return r, resp
	}
	query := metric.ServiceMonitorQuery{
		Rewrite:       true,
		K8sClient:     plugin.k8sClient,
		MonitorClient: plugin.monitorClient,
	}
	result, err := query.Do(r.Context(), namespace, smName)
	if err != nil {
		return r, mw.ReturnErrorResponse(http.StatusInternalServerError, err)
	}
	return r, mw.ReturnDirectResponse(strings.Join(result, "\n"))
}

func (plugin *MetricPlugin) parseParam(ctx context.Context, r *http.Request) (string, string, *mw.HttpResponse) {
	var namespace, smName string
	user := mw.User(ctx)
	if user.ClientID != proxy.AdminClientUser && user.ClientID != proxy.AdminGitOpsUser {
		return namespace, smName, mw.ReturnErrorResponse(http.StatusUnauthorized, errors.Errorf("not authorized"))
	}
	namespace = mux.Vars(r)["namespace"]
	if namespace == "" {
		return namespace, smName,
			mw.ReturnErrorResponse(http.StatusBadRequest, errors.Errorf("namespace cannot be empty"))
	}
	smName = mux.Vars(r)["servicemonitor"]
	if smName == "" {
		return namespace, smName,
			mw.ReturnErrorResponse(http.StatusBadRequest, errors.Errorf("service monitor cannot be empty"))
	}
	return namespace, smName, nil
}
