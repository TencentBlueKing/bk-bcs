/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	commtypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webhook-server/internal/metrics"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webhook-server/internal/types"
)

// Meta meta for mesos object
type Meta struct {
	commtypes.TypeMeta   `json:",inline"`
	commtypes.ObjectMeta `json:"metadata"`
}

// MesosHook do mesos hook
func (ws *WebhookServer) MesosHook(w http.ResponseWriter, r *http.Request) {
	var (
		handler = "MesosHook"
		method  = "POST"
		started = time.Now()
	)

	blog.Infof("received inject request")
	if ws.EngineType == "kubernetes" {
		blog.Warnf("this webhook server only supports kubernetes log config inject")
		http.Error(w, "only support kubernetes log config inject", http.StatusBadRequest)
		metrics.ReportBcsWebhookServerAPIMetrics(handler, method, strconv.Itoa(http.StatusBadRequest), started)
		return
	}

	var body []byte
	if r.Body != nil {
		if data, err := ioutil.ReadAll(r.Body); err == nil {
			body = data
		}
	}
	if len(body) == 0 {
		blog.Errorf("no body found")
		http.Error(w, "no body found", http.StatusBadRequest)
		metrics.ReportBcsWebhookServerAPIMetrics(handler, method, strconv.Itoa(http.StatusBadRequest), started)
		return
	}

	// verify the content type is accurate
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		blog.Errorf("contentType=%s, expect application/json", contentType)
		http.Error(w, "invalid Content-Type, want `application/json`", http.StatusUnsupportedMediaType)
		metrics.ReportBcsWebhookServerAPIMetrics(handler, method, strconv.Itoa(http.StatusUnsupportedMediaType), started)
		return
	}

	var meta *Meta
	err := json.Unmarshal(body, &meta)
	if err != nil {
		blog.Errorf("Could not decode body to meta: %s", err.Error())
		message := fmt.Errorf("could not decode body to meta: %s", err.Error())
		http.Error(w, message.Error(), http.StatusInternalServerError)
		metrics.ReportBcsWebhookServerAPIMetrics(handler, method, strconv.Itoa(http.StatusInternalServerError), started)
		return
	}

	if meta.Kind == commtypes.BcsDataType_APP {
		var application, injectedApplication *commtypes.ReplicaController
		err := json.Unmarshal(body, &application)
		if err != nil {
			blog.Errorf("Could not decode bodyto application: %s", err.Error())
			message := fmt.Errorf("could not decode body to application: %s", err.Error())
			http.Error(w, message.Error(), http.StatusInternalServerError)
			metrics.ReportBcsWebhookServerAPIMetrics(handler, method, strconv.Itoa(http.StatusInternalServerError), started)
			return
		}
		injectedApplication, err = ws.doAppHook(application)
		if err != nil {
			blog.Errorf("do application hook failed, err %s", err.Error())
			message := fmt.Errorf("do application hook failed, err %s", err.Error())
			http.Error(w, message.Error(), http.StatusInternalServerError)
			metrics.ReportBcsWebhookServerAPIMetrics(handler, method, strconv.Itoa(http.StatusInternalServerError), started)
			return
		}
		resp, err := json.Marshal(injectedApplication)
		if err != nil {
			blog.Errorf("Could not encode response: %v", err)
			http.Error(w, fmt.Sprintf("could not encode response: %v", err), http.StatusInternalServerError)
			metrics.ReportBcsWebhookServerAPIMetrics(handler, method, strconv.Itoa(http.StatusInternalServerError), started)
			return
		}
		if _, err := w.Write(resp); err != nil {
			blog.Errorf("Could not write response: %v", err)
			http.Error(w, fmt.Sprintf("could write response: %v", err), http.StatusInternalServerError)
			metrics.ReportBcsWebhookServerAPIMetrics(handler, method, strconv.Itoa(http.StatusInternalServerError), started)
			return
		}

	} else if meta.Kind == commtypes.BcsDataType_DEPLOYMENT {
		var deployment, injectedDeployment *commtypes.BcsDeployment
		err := json.Unmarshal(body, &deployment)
		if err != nil {
			blog.Errorf("Could not decode bodyto deployment: %s", err.Error())
			message := fmt.Errorf("could not decode body to deployment: %s", err.Error())
			http.Error(w, message.Error(), http.StatusInternalServerError)
			metrics.ReportBcsWebhookServerAPIMetrics(handler, method, strconv.Itoa(http.StatusInternalServerError), started)
			return
		}

		injectedDeployment, err = ws.doDepHook(deployment)
		if err != nil {
			blog.Errorf("failed to inject to deployment: %s\n", err.Error())
			http.Error(w, fmt.Sprintf("failed to inject to deployment: %s",
				err.Error()), http.StatusInternalServerError)
			metrics.ReportBcsWebhookServerAPIMetrics(handler, method, strconv.Itoa(http.StatusInternalServerError), started)
			return
		}

		resp, err := json.Marshal(injectedDeployment)
		if err != nil {
			blog.Errorf("Could not encode response: %v", err)
			http.Error(w, fmt.Sprintf("could encode response: %v", err), http.StatusInternalServerError)
			metrics.ReportBcsWebhookServerAPIMetrics(handler, method, strconv.Itoa(http.StatusInternalServerError), started)
			return
		}
		if _, err := w.Write(resp); err != nil {
			blog.Errorf("Could not write response: %v", err)
			http.Error(w, fmt.Sprintf("could write response: %v", err), http.StatusInternalServerError)
			metrics.ReportBcsWebhookServerAPIMetrics(handler, method, strconv.Itoa(http.StatusInternalServerError), started)
			return
		}
	}
	metrics.ReportBcsWebhookServerAPIMetrics(handler, method, strconv.Itoa(http.StatusOK), started)
}

func (ws *WebhookServer) doAppHook(application *commtypes.ReplicaController) (*commtypes.ReplicaController, error) {
	plugins := ws.PluginMgr.GetMesosPlugins()
	pluginNames := ws.PluginMgr.GetMesosPluginNames()
	var err error
	patchedApplication := application

	// check if object in ignore namespaces should be hooked
	if types.IsIgnoredNamespace(application.GetNamespace()) {
		tmpAnnotation := application.GetAnnotations()
		if tmpAnnotation == nil {
			return patchedApplication, nil
		}
		value, ok := tmpAnnotation[types.BcsWebhookAnnotationInjectKey]
		if !ok {
			return patchedApplication, nil
		}
		switch value {
		default:
			return patchedApplication, nil
		case "y", "yes", "true", "on":
			// do nothing, let it go
		}
	}

	for index, p := range plugins {
		patchedApplication, err = p.InjectApplicationContent(patchedApplication)
		if err != nil {
			return nil, fmt.Errorf("plugin %s inject appliction failed, err %s", pluginNames[index], err)
		}
	}
	return patchedApplication, nil
}

func (ws *WebhookServer) doDepHook(deployment *commtypes.BcsDeployment) (*commtypes.BcsDeployment, error) {
	plugins := ws.PluginMgr.GetMesosPlugins()
	pluginNames := ws.PluginMgr.GetMesosPluginNames()
	var err error
	patchedDeployment := deployment

	// check if object in ignore namespaces should be hooked
	if types.IsIgnoredNamespace(deployment.GetNamespace()) {
		tmpAnnotation := deployment.GetAnnotations()
		if tmpAnnotation == nil {
			return patchedDeployment, nil
		}
		value, ok := tmpAnnotation[types.BcsWebhookAnnotationInjectKey]
		if !ok {
			return patchedDeployment, nil
		}
		switch value {
		default:
			return patchedDeployment, nil
		case "y", "yes", "true", "on":
			// do nothing, let it go
		}
	}

	for index, p := range plugins {
		patchedDeployment, err = p.InjectDeployContent(patchedDeployment)
		if err != nil {
			return nil, fmt.Errorf("plugin %s inject deployment failed, err %s", pluginNames[index], err)
		}
	}
	return patchedDeployment, nil
}
