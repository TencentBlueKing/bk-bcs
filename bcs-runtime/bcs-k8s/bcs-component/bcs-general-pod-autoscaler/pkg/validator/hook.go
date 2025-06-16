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

// Package webhook xxx
package webhook

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	v1 "k8s.io/api/admission/v1"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/klog/v2"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-general-pod-autoscaler/pkg/apis/autoscaling/v1alpha1"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-general-pod-autoscaler/pkg/validation"
)

var (
	runtimeScheme = runtime.NewScheme()
	codecs        = serializer.NewCodecFactory(runtimeScheme)
	deserializer  = codecs.UniversalDeserializer()
)

// WebhookServer is the server of webhook
// nolint
type WebhookServer struct {
	*http.Server
}

func init() {
	_ = corev1.AddToScheme(runtimeScheme)
	_ = admissionregistrationv1.AddToScheme(runtimeScheme)
	runtimeScheme.AddKnownTypes(v1alpha1.SchemeGroupVersion)
}

// NewWebhookServer new web hook server
func NewWebhookServer() *WebhookServer {
	return &WebhookServer{}
}

// validate validates gpa spec
func (whsvr *WebhookServer) validate(ar *v1.AdmissionReview) *v1.AdmissionResponse {
	req := ar.Request

	klog.Infof("AdmissionReview for Kind=%v, Namespace=%v Name=%v UID=%v Operation=%v UserInfo=%v",
		req.Kind, req.Namespace, req.Name, req.UID, req.Operation, req.UserInfo)
	var err error
	var causes []metav1.StatusCause
	switch req.Kind.Kind {
	case "GeneralPodAutoscaler":
		causes, err = forGPA(req)

	default:
		return &v1.AdmissionResponse{
			Allowed: false,
		}
	}

	result := metav1.Status{
		Details: &metav1.StatusDetails{
			Name:  ar.Request.Name,
			Group: ar.Request.Kind.Group,
			Kind:  ar.Request.Kind.Kind,
			UID:   ar.Request.UID,
		},
	}
	if err != nil {
		klog.Error(err)
		result.Code = 400
		result.Message = err.Error()
		result.Details.Causes = causes
		return &v1.AdmissionResponse{
			Allowed: false,
			Result:  &result,
		}
	}
	return &v1.AdmissionResponse{
		Allowed: true,
	}
}

// Serve method for webhook server
func (whsvr *WebhookServer) Serve(w http.ResponseWriter, r *http.Request) {
	var body []byte
	if r.Body != nil {
		if data, err := ioutil.ReadAll(r.Body); err == nil {
			body = data
		}
	}
	klog.V(6).Infof("Receive request: %+v", *r)
	if len(body) == 0 {
		klog.Error("empty body")
		http.Error(w, "empty body", http.StatusBadRequest)
		return
	}

	// verify the content type is accurate
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		klog.Errorf("Content-Type=%s, expect application/json", contentType)
		http.Error(w, "invalid Content-Type, expect `application/json`", http.StatusUnsupportedMediaType)
		return
	}

	var admissionResponse *v1.AdmissionResponse
	ar := v1.AdmissionReview{}
	if _, _, err := deserializer.Decode(body, nil, &ar); err != nil {
		klog.Errorf("Can't decode body: %v", err)
		admissionResponse = &v1.AdmissionResponse{
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	} else {
		if r.URL.Path == "/validate" {
			admissionResponse = whsvr.validate(&ar)
		}
	}

	admissionReview := v1.AdmissionReview{}
	if admissionResponse != nil {
		admissionReview.Response = admissionResponse
		if ar.Request != nil {
			admissionReview.Response.UID = ar.Request.UID
		}
	}

	resp, err := json.Marshal(admissionReview)
	if err != nil {
		klog.Errorf("Can't encode response: %v", err)
		http.Error(w, fmt.Sprintf("could not encode response: %v", err), http.StatusInternalServerError)
	}
	if _, err := w.Write(resp); err != nil {
		klog.Errorf("Can't write response: %v", err)
		http.Error(w, fmt.Sprintf("could not write response: %v", err), http.StatusInternalServerError)
	}
}

// forGPA TODO
// nolint
func forGPA(req *v1.AdmissionRequest) ([]metav1.StatusCause, error) {
	var errs field.ErrorList
	causes := make([]metav1.StatusCause, 0)
	defer func() {
		if len(errs) == 0 {
			return
		}
		for i := range errs {
			err := errs[i]
			causes = append(causes, metav1.StatusCause{
				Type:    metav1.CauseType(err.Type),
				Message: err.ErrorBody(),
				Field:   err.Field,
			})
		}
	}()
	var gpa, oldGPA v1alpha1.GeneralPodAutoscaler
	if err := json.Unmarshal(req.Object.Raw, &gpa); err != nil {
		klog.Errorf("Could not unmarshal raw object: %v", err)
		return nil, err
	}
	if req.Operation == v1.Create {
		// validate
		errs = validation.ValidateHorizontalPodAutoscaler(&gpa)
		if len(errs) > 0 {
			return causes, errs.ToAggregate()
		}
	}
	if req.Operation == v1.Update {
		if err := json.Unmarshal(req.OldObject.Raw, &oldGPA); err != nil {
			klog.Errorf("Could not unmarshal old raw object: %v", err)
			return nil, err
		}
		// validate
		errs = validation.ValidateHorizontalPodAutoscalerUpdate(&gpa, &oldGPA)
		if len(errs) > 0 {
			return causes, errs.ToAggregate()
		}
	}
	return nil, nil
}
