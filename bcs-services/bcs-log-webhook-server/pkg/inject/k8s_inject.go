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
	bcsv2 "bk-bcs/bcs-services/bcs-log-webhook-server/pkg/apis/bk-bcs/v2"
	mapset "github.com/deckarep/golang-set"
	"k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

const (
	StandardConfigType               = "standard"
	BcsSystemConfigType              = "bcs-system"
	DataIdEnvKey                     = "io_tencent_bcs_app_dataid"
	AppIdEnvKey                      = "io_tencent_bcs_app_appid"
	StdoutEnvKey                     = "io_tencent_bcs_app_stdout"
	LogPathEnvKey                    = "io_tencent_bcs_app_logpath"
	ClusterIdEnvKey                  = "io_tencent_bcs_app_cluster"
	NamespaceEnvKey                  = "io_tencent_bcs_app_namespace"
	BcsLogWebhookAnnotationInjectKey = "sidecar.log.conf.bkbcs.tencent.com"
)

var (
	runtimeScheme = runtime.NewScheme()
	codecs        = serializer.NewCodecFactory(runtimeScheme)
	deserializer  = codecs.UniversalDeserializer()

	// (https://github.com/kubernetes/kubernetes/issues/57982)
	defaulter = runtime.ObjectDefaulter(runtimeScheme)
)

type patchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

func toAdmissionResponse(err error) *v1beta1.AdmissionResponse {
	return &v1beta1.AdmissionResponse{Result: &metav1.Status{Message: err.Error()}}
}

func (whSvr *WebhookServer) K8sLogInject(w http.ResponseWriter, r *http.Request) {
	if whSvr.EngineType == "mesos" {
		blog.Warnf("this webhook server only supports mesos log config inject")
		http.Error(w, "only support mesos log config inject", http.StatusBadRequest)
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

	var reviewResponse *v1beta1.AdmissionResponse
	ar := v1beta1.AdmissionReview{}
	if _, _, err := deserializer.Decode(body, nil, &ar); err != nil {
		blog.Errorf("Could not decode body: %s", err.Error())
		reviewResponse = toAdmissionResponse(err)
	} else {
		reviewResponse = whSvr.k8sInject(&ar)
	}

	response := v1beta1.AdmissionReview{}
	if reviewResponse != nil {
		response.Response = reviewResponse
		if ar.Request != nil {
			response.Response.UID = ar.Request.UID
		}
	}

	resp, err := json.Marshal(response)
	if err != nil {
		blog.Errorf("Could not encode response: %v", err)
		http.Error(w, fmt.Sprintf("could encode response: %v", err), http.StatusInternalServerError)
	}
	if _, err := w.Write(resp); err != nil {
		blog.Errorf("Could not write response: %v", err)
		http.Error(w, fmt.Sprintf("could write response: %v", err), http.StatusInternalServerError)
	}
}

func (whSvr *WebhookServer) k8sInject(ar *v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
	req := ar.Request
	var pod corev1.Pod
	if err := json.Unmarshal(req.Object.Raw, &pod); err != nil {
		blog.Errorf("Could not unmarshal raw object: %v %s", err,
			string(req.Object.Raw))
		return toAdmissionResponse(err)
	}

	// Deal with potential empty fields, e.g., when the pod is created by a deployment
	podName := potentialPodName(&pod.ObjectMeta)
	if pod.ObjectMeta.Namespace == "" {
		pod.ObjectMeta.Namespace = req.Namespace
	}

	blog.Infof("AdmissionReview for Kind=%v Namespace=%v Name=%v (%v) UID=%v Rfc6902PatchOperation=%v UserInfo=%v",
		req.Kind, req.Namespace, req.Name, podName, req.UID, req.Operation, req.UserInfo)
	//blog.Infof("Object: %v", string(req.Object.Raw))
	//blog.Infof("OldObject: %v", string(req.OldObject.Raw))

	if !injectRequired(ignoredNamespaces, &pod.ObjectMeta) {
		blog.Infof("Skipping %s/%s due to policy check", pod.ObjectMeta.Namespace, podName)
		return &v1beta1.AdmissionResponse{
			Allowed: true,
		}
	}

	patchBytes, err := whSvr.createPatch(&pod)
	if err != nil {
		blog.Infof("AdmissionResponse: err=%v\n", err)

		return toAdmissionResponse(err)
	}

	blog.Infof("AdmissionResponse: patch=%v\n", string(patchBytes))

	reviewResponse := v1beta1.AdmissionResponse{
		Allowed: true,
		Patch:   patchBytes,
		PatchType: func() *v1beta1.PatchType {
			pt := v1beta1.PatchTypeJSONPatch
			return &pt
		}(),
	}
	return &reviewResponse
}

func injectRequired(ignored []string, metadata *metav1.ObjectMeta) bool {

	annotations := metadata.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}

	for _, namespace := range ignored {
		if metadata.Namespace == namespace {
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

func potentialPodName(metadata *metav1.ObjectMeta) string {
	if metadata.Name != "" {
		return metadata.Name
	}
	if metadata.GenerateName != "" {
		return metadata.GenerateName + "***** (actual name not yet known)"
	}
	return ""
}

func (whSvr *WebhookServer) createPatch(pod *corev1.Pod) ([]byte, error) {

	var patch []patchOperation

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
	if namespaceSet.Contains(pod.ObjectMeta.Namespace) {
		matchedLogConf := findBcsSystemConfigType(bcsLogConfs)
		if matchedLogConf != nil {
			for i, container := range pod.Spec.Containers {
				patchedContainer := whSvr.injectK8sContainer(pod.Namespace, &container, matchedLogConf)
				patch = append(patch, replaceContainer(i, *patchedContainer))
			}
		}
		return json.Marshal(patch)
	}

	for i, container := range pod.Spec.Containers {
		matchedLogConf := findMatchedConfigType(container.Name, bcsLogConfs)
		if matchedLogConf != nil {
			patchedContainer := whSvr.injectK8sContainer(pod.Namespace, &container, matchedLogConf)
			patch = append(patch, replaceContainer(i, *patchedContainer))
		}
	}

	return json.Marshal(patch)
}

func replaceContainer(index int, patchedContainer corev1.Container) (patch patchOperation) {
	patch = patchOperation{
		Op:    "replace",
		Path:  fmt.Sprintf("/spec/containers/%v", index),
		Value: patchedContainer,
	}
	return patch
}

func (whSvr *WebhookServer) injectK8sContainer(namespace string, container *corev1.Container, logConf *bcsv2.BcsLogConfig) *corev1.Container {

	patchedContainer := container.DeepCopy()

	var envs []corev1.EnvVar
	dataIdEnv := corev1.EnvVar{
		Name:  DataIdEnvKey,
		Value: logConf.Spec.DataId,
	}
	envs = append(envs, dataIdEnv)

	appIdEnv := corev1.EnvVar{
		Name:  AppIdEnvKey,
		Value: logConf.Spec.AppId,
	}
	envs = append(envs, appIdEnv)

	stdoutEnv := corev1.EnvVar{
		Name:  StdoutEnvKey,
		Value: strconv.FormatBool(logConf.Spec.Stdout),
	}
	envs = append(envs, stdoutEnv)

	logPathEnv := corev1.EnvVar{
		Name:  LogPathEnvKey,
		Value: logConf.Spec.LogPath,
	}
	envs = append(envs, logPathEnv)

	clusterIdEnv := corev1.EnvVar{
		Name:  ClusterIdEnvKey,
		Value: logConf.Spec.ClusterId,
	}
	envs = append(envs, clusterIdEnv)

	namespaceEnv := corev1.EnvVar{
		Name:  NamespaceEnvKey,
		Value: namespace,
	}
	envs = append(envs, namespaceEnv)

	patchedContainer.Env = envs

	return patchedContainer
}
