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

	"github.com/Tencent/bk-bcs/bcs-k8s/bcs-public-cluster-webhook/pkg/check"

	k8sunstruct "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-k8s/bcs-public-cluster-webhook/pkg/metrics"
	"github.com/Tencent/bk-bcs/bcs-k8s/bcs-public-cluster-webhook/pkg/util"
	"k8s.io/api/admission/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

var (
	runtimeScheme = runtime.NewScheme()
	codecs        = serializer.NewCodecFactory(runtimeScheme)
	deserializer  = codecs.UniversalDeserializer()
)

// Validate validate k8s resource
func (ws *WebhookServer) Validate(w http.ResponseWriter, r *http.Request) {
	var (
		handler = "validate"
		method  = "POST"
		started = time.Now()
	)

	var body []byte
	if r.Body != nil {
		if data, err := ioutil.ReadAll(r.Body); err == nil {
			body = data
		}
	}
	if len(body) == 0 {
		blog.Errorf("no body found")
		http.Error(w, "no body found", http.StatusBadRequest)
		metrics.ReportWebhookAPIMetrics(handler, method, strconv.Itoa(http.StatusBadRequest), started)
		return
	}

	// verify the content type is accurate
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		blog.Errorf("contentType=%s, expect application/json", contentType)
		http.Error(w, "invalid Content-Type, want `application/json`", http.StatusUnsupportedMediaType)
		metrics.ReportWebhookAPIMetrics(handler, method, strconv.Itoa(http.StatusUnsupportedMediaType), started)
		return
	}

	var reviewResponse *v1beta1.AdmissionResponse
	ar := v1beta1.AdmissionReview{}
	if _, _, err := deserializer.Decode(body, nil, &ar); err != nil {
		blog.Errorf("Could not decode body: %s", err.Error())
		reviewResponse = util.ToAdmissionResponse(err)
	} else {
		reviewResponse = ws.doValidate(ar)
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
		metrics.ReportWebhookAPIMetrics(handler, method, strconv.Itoa(http.StatusInternalServerError), started)
		return
	}
	if _, err := w.Write(resp); err != nil {
		blog.Errorf("Could not write response: %v", err)
		http.Error(w, fmt.Sprintf("could write response: %v", err), http.StatusInternalServerError)
		metrics.ReportWebhookAPIMetrics(handler, method, strconv.Itoa(http.StatusInternalServerError), started)
		return
	}

	metrics.ReportWebhookAPIMetrics(handler, method, strconv.Itoa(http.StatusOK), started)
}

func (ws *WebhookServer) doValidate(ar v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
	req := ar.Request

	runtimeObj := req.Object
	if req.Operation == v1beta1.Delete {
		runtimeObj = req.OldObject
	}
	// decode object bytes
	tmpMapIf := make(map[string]interface{})
	if err := json.Unmarshal(runtimeObj.Raw, &tmpMapIf); err != nil {
		blog.Errorf("decode %s to map[string]interface failed, err %s", string(runtimeObj.Raw), err.Error())
		return util.ToAdmissionResponse(
			fmt.Errorf("decode data to map[string]interface failed failed, err %s", err.Error()))
	}
	tmpUnstruct := &k8sunstruct.Unstructured{}
	tmpUnstruct.SetUnstructuredContent(tmpMapIf)
	// for k8s 1.12.6, the GroupVersionKind may lost in runtimeObj
	tmpUnstruct.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   req.Kind.Group,
		Version: req.Kind.Version,
		Kind:    req.Kind.Kind,
	})
	tmpUnstructNs := tmpUnstruct.GetNamespace()
	tmpUnstructName := tmpUnstruct.GetName()
	tmpUnstructKind := tmpUnstruct.GetKind()
	// Deal with potential empty fields, e.g., when the pod is created by a deployment
	if len(tmpUnstructNs) == 0 {
		tmpUnstructNs = req.Namespace
	}
	if len(tmpUnstructName) == 0 {
		tmpUnstructName = "Unknown"
	}

	//检查k8s资源的某些字段是否在黑名单中
	reqCheck := &check.RequestCheck{
		Kind:      tmpUnstructKind,
		Namespace: tmpUnstructNs,
		Name:      tmpUnstructName,
		Object:    runtimeObj.Raw,
	}
	blog.Infof("开始校验资源是否在黑名单中 kind:%s, namespace:%s, name:%s", tmpUnstructKind, tmpUnstructNs, tmpUnstructName)
	resCheck, err := ws.BlackList.Check(reqCheck)
	if err != nil {
		blog.Errorf("check resource in blacklist error, err=%s", err)
		return util.ToAdmissionResponse(
			fmt.Errorf("check resource in blacklist error, err %s", err.Error()))
	}
	if !resCheck.Allowed {
		blog.Infof("资源 kind:%s, namespace:%s, name:%s 不允许创建/更新, 原因：%s, 资源对象: %s", tmpUnstructNs, tmpUnstructKind, tmpUnstructName, resCheck.Message, string(runtimeObj.Raw))
		return util.ToAdmissionResponse(
			fmt.Errorf("%s", resCheck.Message))
	}

	return &v1beta1.AdmissionResponse{Allowed: true}
}
