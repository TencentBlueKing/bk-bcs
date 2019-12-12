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
	"strconv"
	"strings"

	"bk-bcs/bcs-common/common/blog"
	commtypes "bk-bcs/bcs-common/common/types"
	bcsv2 "bk-bcs/bcs-services/bcs-log-webhook-server/pkg/apis/bk-bcs/v2"
	mapset "github.com/deckarep/golang-set"
	"k8s.io/apimachinery/pkg/labels"
)

type Meta struct {
	commtypes.TypeMeta   `json:",inline"`
	commtypes.ObjectMeta `json:"metadata"`
}

//MesosLogInject inject bcs log config to container and respond to mesos
func (whSvr *WebhookServer) MesosLogInject(w http.ResponseWriter, r *http.Request) {
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
		}

		injectedApplication, err = whSvr.mesosApplicationInject(application)

		resp, err := json.Marshal(injectedApplication)
		if err != nil {
			blog.Errorf("Could not encode response: %v", err)
			http.Error(w, fmt.Sprintf("could encode response: %v", err), http.StatusInternalServerError)
		}
		if _, err := w.Write(resp); err != nil {
			blog.Errorf("Could not write response: %v", err)
			http.Error(w, fmt.Sprintf("could write response: %v", err), http.StatusInternalServerError)
		}
	} else if meta.Kind == commtypes.BcsDataType_DEPLOYMENT { // action to mesos deployment resource
		var deployment, injectedDeployment *commtypes.BcsDeployment
		err := json.Unmarshal(body, &deployment)
		if err != nil {
			blog.Errorf("Could not decode bodyto deployment: %s", err.Error())
			message := fmt.Errorf("could not decode body to deployment: %s", err.Error())
			http.Error(w, message.Error(), http.StatusInternalServerError)
		}

		injectedDeployment, err = whSvr.mesosDeploymentInject(deployment)
		resp, err := json.Marshal(injectedDeployment)
		if err != nil {
			blog.Errorf("Could not encode response: %v", err)
			http.Error(w, fmt.Sprintf("could encode response: %v", err), http.StatusInternalServerError)
		}
		if _, err := w.Write(resp); err != nil {
			blog.Errorf("Could not write response: %v", err)
			http.Error(w, fmt.Sprintf("could write response: %v", err), http.StatusInternalServerError)
		}
	}
}

// mesosApplicationInject inject bcs log config to containers for applications
func (whSvr *WebhookServer) mesosApplicationInject(application *commtypes.ReplicaController) (*commtypes.ReplicaController, error) {
	if !mesosInjectRequired(ignoredNamespaces, &application.ObjectMeta) {
		blog.Infof("Skipping %s/%s due to policy check", application.ObjectMeta.NameSpace, application.ObjectMeta.Name)
		return application, nil
	}

	// get all BcsLogConfig
	bcsLogConfs, err := whSvr.BcsLogConfigLister.List(labels.Everything())
	if err != nil {
		blog.Errorf("list bcslogconfig error %s", err.Error())
		return nil, err
	}

	//handle bcs-system modules' log inject
	namespaceSet := mapset.NewSet()
	for _, namespace := range ignoredNamespaces {
		namespaceSet.Add(namespace)
	}
	if namespaceSet.Contains(application.ObjectMeta.NameSpace) {
		matchedLogConf := findBcsSystemConfigType(bcsLogConfs)
		if matchedLogConf != nil {
			injected := whSvr.injectMesosContainers(application.ObjectMeta.NameSpace, application.ReplicaControllerSpec.Template, matchedLogConf)
			application.ReplicaControllerSpec.Template = injected
		}
		return application, nil
	}

	// handle business modules log inject
	var injectedContainers []commtypes.Container
	for _, container := range application.ReplicaControllerSpec.Template.PodSpec.Containers {
		matchedLogConf := findMatchedConfigType(container.Name, bcsLogConfs)
		if matchedLogConf != nil {
			injectedContainer := whSvr.injectMesosContainer(application.ObjectMeta.NameSpace, container, matchedLogConf)
			injectedContainers = append(injectedContainers, injectedContainer)
		} else {
			injectedContainers = append(injectedContainers, container)
		}
	}
	application.ReplicaControllerSpec.Template.PodSpec.Containers = injectedContainers

	for _, ct := range application.ReplicaControllerSpec.Template.PodSpec.Containers {
		blog.Infof("%v", ct.Env)
	}

	return application, nil
}

// mesosDeploymentInject inject bcs log config to containers for deployment
func (whSvr *WebhookServer) mesosDeploymentInject(deployment *commtypes.BcsDeployment) (*commtypes.BcsDeployment, error) {
	if !mesosInjectRequired(ignoredNamespaces, &deployment.ObjectMeta) {
		blog.Infof("Skipping %s/%s due to policy check", deployment.ObjectMeta.NameSpace, deployment.ObjectMeta.Name)
		return deployment, nil
	}

	// get all BcsLogConfig
	bcsLogConfs, err := whSvr.BcsLogConfigLister.List(labels.Everything())
	if err != nil {
		blog.Errorf("list bcslogconfig error %s", err.Error())
		return nil, err
	}

	//handle bcs-system modules' log inject
	namespaceSet := mapset.NewSet()
	for _, namespace := range ignoredNamespaces {
		namespaceSet.Add(namespace)
	}
	if namespaceSet.Contains(deployment.ObjectMeta.NameSpace) {
		matchedLogConf := findBcsSystemConfigType(bcsLogConfs)
		if matchedLogConf != nil {
			injected := whSvr.injectMesosContainers(deployment.ObjectMeta.NameSpace, deployment.Spec.Template, matchedLogConf)
			deployment.Spec.Template = injected
		}
		return deployment, nil
	}

	// handle business modules log inject
	var injectedContainers []commtypes.Container
	for _, container := range deployment.Spec.Template.PodSpec.Containers {
		matchedLogConf := findMatchedConfigType(container.Name, bcsLogConfs)
		if matchedLogConf != nil {
			injectedContainer := whSvr.injectMesosContainer(deployment.ObjectMeta.NameSpace, container, matchedLogConf)
			injectedContainers = append(injectedContainers, injectedContainer)
		} else {
			injectedContainers = append(injectedContainers, container)
		}
	}
	deployment.Spec.Template.PodSpec.Containers = injectedContainers
	return deployment, nil
}

// injectMesosContainer injects bcs log config to an container
func (whSvr *WebhookServer) injectMesosContainer(namespace string, container commtypes.Container, logConf *bcsv2.BcsLogConfig) commtypes.Container {
	var envs []commtypes.EnvVar
	dataIdEnv := commtypes.EnvVar{
		Name:  DataIdEnvKey,
		Value: logConf.Spec.DataId,
	}
	envs = append(envs, dataIdEnv)

	appIdEnv := commtypes.EnvVar{
		Name:  AppIdEnvKey,
		Value: logConf.Spec.AppId,
	}
	envs = append(envs, appIdEnv)

	stdoutEnv := commtypes.EnvVar{
		Name:  StdoutEnvKey,
		Value: strconv.FormatBool(logConf.Spec.Stdout),
	}
	envs = append(envs, stdoutEnv)

	logPathEnv := commtypes.EnvVar{
		Name:  LogPathEnvKey,
		Value: logConf.Spec.LogPath,
	}
	envs = append(envs, logPathEnv)

	clusterIdEnv := commtypes.EnvVar{
		Name:  ClusterIdEnvKey,
		Value: logConf.Spec.ClusterId,
	}
	envs = append(envs, clusterIdEnv)

	namespaceEnv := commtypes.EnvVar{
		Name:  NamespaceEnvKey,
		Value: namespace,
	}
	envs = append(envs, namespaceEnv)

	container.Env = envs

	blog.Infof("%v", container.Env)
	return container
}

// injectMesosContainers injects bcs log config to all containers
func (whSvr *WebhookServer) injectMesosContainers(namespace string, podTemplate *commtypes.PodTemplateSpec, logConf *bcsv2.BcsLogConfig) *commtypes.PodTemplateSpec {

	var injectedContainers []commtypes.Container
	for _, container := range podTemplate.PodSpec.Containers {
		injectedContainer := whSvr.injectMesosContainer(namespace, container, logConf)
		injectedContainers = append(injectedContainers, injectedContainer)
	}

	podTemplate.PodSpec.Containers = injectedContainers
	return podTemplate
}

// mesosInjectRequired validates whether an application or deployment should be injected
func mesosInjectRequired(ignored []string, metadata *commtypes.ObjectMeta) bool {

	annotations := metadata.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}

	for _, namespace := range ignored {
		if metadata.NameSpace == namespace {
			switch strings.ToLower(annotations[BcsLogWebhookAnnotationInjectKey]) {
			default:
				return false
			case "y", "yes", "true", "on":
				return true
			}
		}
	}
	return true
}
