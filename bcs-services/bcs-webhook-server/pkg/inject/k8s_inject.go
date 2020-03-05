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
	"bk-bcs/bcs-services/bcs-webhook-server/pkg/inject/common"
	"bk-bcs/bcs-services/bcs-webhook-server/pkg/inject/k8s"
	"k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

var (
	runtimeScheme = runtime.NewScheme()
	codecs        = serializer.NewCodecFactory(runtimeScheme)
	deserializer  = codecs.UniversalDeserializer()

	// (https://github.com/kubernetes/kubernetes/issues/57982)
	defaulter = runtime.ObjectDefaulter(runtimeScheme)
)

func toAdmissionResponse(err error) *v1beta1.AdmissionResponse {
	return &v1beta1.AdmissionResponse{Result: &metav1.Status{Message: err.Error()}}
}

func (whSvr *WebhookServer) K8sInject(w http.ResponseWriter, r *http.Request) {
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
		return
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

	if !injectRequired(common.IgnoredNamespaces, &pod.ObjectMeta) {
		blog.Infof("Skipping %s/%s due to policy check", pod.ObjectMeta.Namespace, podName)
		return &v1beta1.AdmissionResponse{
			Allowed: true,
		}
	}

	patchBytes, err := whSvr.createPatch(&pod)
	if err != nil {
		blog.Errorf("AdmissionResponse: err=%s\n", err.Error())
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

	var patch []k8s.PatchOperation

	if whSvr.Injects.LogConfEnv {
		logConfInjectPatch, err := whSvr.K8sLogConfInject.InjectContent(pod)
		if err != nil {
			return nil, fmt.Errorf("failed to inject bcs log conf: %s", err.Error())
		}
		patch = append(patch, logConfInjectPatch...)
	}

	if whSvr.Injects.DbPriv.DbPrivInject {
		dbPrivConfInjectPatch, err := whSvr.K8sDbPrivConfInject.InjectContent(pod)
		if err != nil {
			return nil, fmt.Errorf("failed to inject db privilege conf: %s", err.Error())
		}
		patch = append(patch, dbPrivConfInjectPatch...)
	}

	if whSvr.Injects.Bscp.BscpInject {
		bscpInjectPatch, err := whSvr.K8sBscpInject.InjectContent(pod)
		if err != nil {
			return nil, fmt.Errorf("failed to inject bscp sidecar, err %s", err.Error())
		}
		patch = append(patch, bscpInjectPatch...)
	}

	return json.Marshal(patch)
}
