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
 *
 */

package inject

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"bk-bcs/bcs-common/common/blog"
	commtypes "bk-bcs/bcs-common/common/types"
	"bk-bcs/bcs-services/bcs-webhook-server/pkg/inject/common"
)

type Meta struct {
	commtypes.TypeMeta   `json:",inline"`
	commtypes.ObjectMeta `json:"metadata"`
}

//MesosLogInject inject bcs log config to container and respond to mesos
func (whSvr *WebhookServer) MesosInject(w http.ResponseWriter, r *http.Request) {
	blog.Infof("received inject request")
	if whSvr.EngineType == "kubernetes" {
		blog.Warnf("this webhook server only supports kubernetes log config inject")
		http.Error(w, "only support kubernetes log config inject", http.StatusBadRequest)
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
		return
	}

	// verify the content type is accurate
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		blog.Errorf("contentType=%s, expect application/json", contentType)
		http.Error(w, "invalid Content-Type, want `application/json`", http.StatusUnsupportedMediaType)
		return
	}

	var meta *Meta
	err := json.Unmarshal(body, &meta)
	if err != nil {
		blog.Errorf("Could not decode body to meta: %s", err.Error())
		message := fmt.Errorf("could not decode body to meta: %s", err.Error())
		http.Error(w, message.Error(), http.StatusInternalServerError)
		return
	}

	blog.Info(string(meta.Kind))

	// action to mesos application resource
	if meta.Kind == commtypes.BcsDataType_APP {
		var application, injectedApplication *commtypes.ReplicaController
		err := json.Unmarshal(body, &application)
		if err != nil {
			blog.Errorf("Could not decode bodyto application: %s", err.Error())
			message := fmt.Errorf("could not decode body to application: %s", err.Error())
			http.Error(w, message.Error(), http.StatusInternalServerError)
			return
		}

		injectedApplication, err = whSvr.mesosApplicationInject(application)
		if err != nil {
			blog.Errorf("failed to inject to application: %s\n", err.Error())
			http.Error(w, fmt.Sprintf("failed to inject to application: %s", err.Error()), http.StatusInternalServerError)
			return
		}

		resp, err := json.Marshal(injectedApplication)
		if err != nil {
			blog.Errorf("Could not encode response: %v", err)
			http.Error(w, fmt.Sprintf("could not encode response: %v", err), http.StatusInternalServerError)
			return
		}
		if _, err := w.Write(resp); err != nil {
			blog.Errorf("Could not write response: %v", err)
			http.Error(w, fmt.Sprintf("could write response: %v", err), http.StatusInternalServerError)
			return
		}
	} else if meta.Kind == commtypes.BcsDataType_DEPLOYMENT { // action to mesos deployment resource
		var deployment, injectedDeployment *commtypes.BcsDeployment
		err := json.Unmarshal(body, &deployment)
		if err != nil {
			blog.Errorf("Could not decode bodyto deployment: %s", err.Error())
			message := fmt.Errorf("could not decode body to deployment: %s", err.Error())
			http.Error(w, message.Error(), http.StatusInternalServerError)
			return
		}

		injectedDeployment, err = whSvr.mesosDeploymentInject(deployment)
		if err != nil {
			blog.Errorf("failed to inject to deployment: %s\n", err.Error())
			http.Error(w, fmt.Sprintf("failed to inject to deployment: %s", err.Error()), http.StatusInternalServerError)
			return
		}

		resp, err := json.Marshal(injectedDeployment)
		if err != nil {
			blog.Errorf("Could not encode response: %v", err)
			http.Error(w, fmt.Sprintf("could encode response: %v", err), http.StatusInternalServerError)
			return
		}
		if _, err := w.Write(resp); err != nil {
			blog.Errorf("Could not write response: %v", err)
			http.Error(w, fmt.Sprintf("could write response: %v", err), http.StatusInternalServerError)
			return
		}
	}
}

// mesosApplicationInject inject bcs log config to containers for applications
func (whSvr *WebhookServer) mesosApplicationInject(application *commtypes.ReplicaController) (*commtypes.ReplicaController, error) {
	if !mesosInjectRequired(common.IgnoredNamespaces, &application.ObjectMeta) {
		blog.Infof("Skipping %s/%s due to policy check", application.ObjectMeta.NameSpace, application.ObjectMeta.Name)
		return application, nil
	}

	var patchedApplication *commtypes.ReplicaController
	var err error

	if whSvr.Injects.LogConfEnv {
		patchedApplication, err = whSvr.MesosLogConfInject.InjectApplicationContent(application)
		if err != nil {
			return nil, fmt.Errorf("failed to inject bcs log conf to application: %s", err.Error())
		}
	}

	if whSvr.Injects.Bscp.BscpInject {
		patchedApplication, err = whSvr.MesosBscpInject.InjectApplicationContent(application)
		if err != nil {
			return nil, fmt.Errorf("failed to inject bscp sidecar to application %s/%s, err %s",
				application.GetNamespace(), application.GetName(), err.Error())
		}
	}

	return patchedApplication, nil
}

// mesosDeploymentInject inject bcs log config to containers for deployment
func (whSvr *WebhookServer) mesosDeploymentInject(deployment *commtypes.BcsDeployment) (*commtypes.BcsDeployment, error) {
	if !mesosInjectRequired(common.IgnoredNamespaces, &deployment.ObjectMeta) {
		blog.Infof("Skipping %s/%s due to policy check", deployment.ObjectMeta.NameSpace, deployment.ObjectMeta.Name)
		return deployment, nil
	}

	var patchedDeploy *commtypes.BcsDeployment
	var err error

	if whSvr.Injects.LogConfEnv {
		patchedDeploy, err = whSvr.MesosLogConfInject.InjectDeployContent(deployment)
		if err != nil {
			return nil, fmt.Errorf("failed to inject bcs log conf to deployment: %s", err.Error())
		}
	}

	return patchedDeploy, nil
}

// mesosInjectRequired validates whether an application or deployment should be injected
func mesosInjectRequired(ignored []string, metadata *commtypes.ObjectMeta) bool {

	annotations := metadata.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}

	for _, namespace := range ignored {
		if metadata.NameSpace == namespace {
			switch strings.ToLower(annotations[common.BcsWebhookAnnotationInjectKey]) {
			default:
				return false
			case "y", "yes", "true", "on":
				return true
			}
		}
	}
	return true
}
