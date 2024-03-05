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

package webhookserver

import (
	v1 "k8s.io/api/admission/v1"
	"k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// admitv1beta1Func handles a v1beta1 admission
type admitv1beta1Func func(v1beta1.AdmissionReview) *v1beta1.AdmissionResponse

// admitv1beta1Func handles a v1 admission
type admitv1Func func(v1.AdmissionReview) *v1.AdmissionResponse

// admitHandler is a handler, for both validators and mutators, that supports multiple admission review versions
type admitHandler struct {
	v1beta1 admitv1beta1Func
	v1      admitv1Func
}

func newDelegateToV1AdmitHandler(f admitv1Func) admitHandler {
	return admitHandler{
		v1beta1: delegateV1beta1AdmitToV1(f),
		v1:      f,
	}
}

func delegateV1beta1AdmitToV1(f admitv1Func) admitv1beta1Func {
	return func(review v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
		in := v1.AdmissionReview{Request: convertAdmissionRequestToV1(review.Request)}
		out := f(in)
		return convertAdmissionResponseToV1beta1(out)
	}
}

func convertAdmissionRequestToV1(r *v1beta1.AdmissionRequest) *v1.AdmissionRequest {
	return &v1.AdmissionRequest{
		Kind:               r.Kind,
		Namespace:          r.Namespace,
		Name:               r.Name,
		Object:             r.Object,
		Resource:           r.Resource,
		Operation:          v1.Operation(r.Operation),
		UID:                r.UID,
		DryRun:             r.DryRun,
		OldObject:          r.OldObject,
		Options:            r.Options,
		RequestKind:        r.RequestKind,
		RequestResource:    r.RequestResource,
		RequestSubResource: r.RequestSubResource,
		SubResource:        r.SubResource,
		UserInfo:           r.UserInfo,
	}
}

func convertAdmissionResponseToV1beta1(r *v1.AdmissionResponse) *v1beta1.AdmissionResponse {
	var pt *v1beta1.PatchType
	if r.PatchType != nil {
		t := v1beta1.PatchType(*r.PatchType)
		pt = &t
	}
	return &v1beta1.AdmissionResponse{
		UID:              r.UID,
		Allowed:          r.Allowed,
		AuditAnnotations: r.AuditAnnotations,
		Patch:            r.Patch,
		PatchType:        pt,
		Result:           r.Result,
	}
}

func toV1AdmissionResponse(err error) *v1.AdmissionResponse {
	return &v1.AdmissionResponse{
		Result: &metav1.Status{
			Message: err.Error(),
		},
	}
}
