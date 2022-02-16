/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package validation

import (
	"context"
	"fmt"

	v1 "k8s.io/api/admission/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"

	gdv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamedeployment-operator/pkg/apis/tkex/v1alpha1"
)

// handleCRD handles admission requests when deleting CRD
func (whsvr *WebhookServer) handleCRD(ar v1.AdmissionReview) *v1.AdmissionResponse {
	klog.V(2).Info("handling crd")

	resource := "customresourcedefinitions"
	v1beta1GVR := metav1.GroupVersionResource{Group: apiextensionsv1beta1.GroupName,
		Version: "v1beta1", Resource: resource}
	v1GVR := metav1.GroupVersionResource{Group: apiextensionsv1.GroupName,
		Version: "v1", Resource: resource}

	reviewResponse := v1.AdmissionResponse{}
	reviewResponse.Allowed = true

	if ar.Request.Operation != v1.Delete || ar.Request.SubResource != "" {
		return &reviewResponse
	}
	if len(ar.Request.OldObject.Raw) == 0 {
		klog.Warningf("Skip to validate CRD %s deletion for no old object, maybe because of Kubernetes version < 1.16",
			ar.Request.Name)
		return &reviewResponse
	}

	raw := ar.Request.OldObject.Raw
	var labels map[string]string

	switch ar.Request.Resource {
	case v1beta1GVR:
		crd := apiextensionsv1beta1.CustomResourceDefinition{}
		deserializer := codecs.UniversalDeserializer()
		if _, _, err := deserializer.Decode(raw, nil, &crd); err != nil {
			klog.Error(err)
			return toV1AdmissionResponse(err)
		}
		if crd.Spec.Names.Kind != gdv1alpha1.Kind {
			return &reviewResponse
		}
		labels = crd.Labels
	case v1GVR:
		crd := apiextensionsv1.CustomResourceDefinition{}
		deserializer := codecs.UniversalDeserializer()
		if _, _, err := deserializer.Decode(raw, nil, &crd); err != nil {
			klog.Error(err)
			return toV1AdmissionResponse(err)
		}
		if crd.Spec.Names.Kind != gdv1alpha1.Kind {
			return &reviewResponse
		}
		labels = crd.Labels
	default:
		err := fmt.Errorf("expect resource to be one of [%v, %v] but got %v", v1beta1GVR, v1GVR, ar.Request.Resource)
		klog.Error(err)
		return toV1AdmissionResponse(err)
	}

	if err := whsvr.validateCRDDeletion(labels); err != nil {
		return toV1AdmissionResponse(err)
	}

	return &reviewResponse

}

func (whsvr *WebhookServer) validateCRDDeletion(labels map[string]string) error {
	switch val := labels[DeletionAllowKey]; val {
	case DeletionAllowTypeAlways:
		return nil
	case DeletionAllowTypeCascading:
		gdList, err := whsvr.gdClient.TkexV1alpha1().GameDeployments("").List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return fmt.Errorf("failed to list CRs of GameDeployments: %v", err)
		}
		var activeCnt int
		for i := range gdList.Items {
			if gdList.Items[i].GetDeletionTimestamp() == nil {
				activeCnt++
			}
		}
		if activeCnt > 0 {
			return fmt.Errorf("forbidden by ResourceDeletionAllow for %s=%s and active CRs %d>0",
				DeletionAllowKey, DeletionAllowTypeCascading, activeCnt)
		}

	default:
		return fmt.Errorf("forbidden by ResourceDeletionAllow, set labels %s=%s to allow operation",
			DeletionAllowKey, DeletionAllowTypeAlways)
	}
	return nil
}
