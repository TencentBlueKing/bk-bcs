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
	"context"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/constant"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
	"github.com/pkg/errors"
	v1 "k8s.io/api/admission/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// DeletionAllowKey is a key in object labels and its value can be Always and Cascading.
	DeletionAllowKey = "io.tencent.bcs.dev/deletion-allow"
	// DeletionAllowTypeAlways indicates this object will always be allowed to be deleted.
	DeletionAllowTypeAlways = "Always"
	// DeletionAllowTypeCascading indicates this object will be forbidden to be deleted, if it
	// has active resources owned.
	DeletionAllowTypeCascading = "Cascading"
)

func (s *Server) getCRDLabelFromAR(ar v1.AdmissionReview) (map[string]string, error) {
	resource := "customresourcedefinitions"
	v1beta1GVR := metav1.GroupVersionResource{Group: apiextensionsv1beta1.GroupName,
		Version: "v1beta1", Resource: resource}
	v1GVR := metav1.GroupVersionResource{Group: apiextensionsv1.GroupName,
		Version: "v1", Resource: resource}

	raw := ar.Request.OldObject.Raw

	var label map[string]string
	switch ar.Request.Resource {
	case v1beta1GVR:
		crd := apiextensionsv1beta1.CustomResourceDefinition{}
		if _, _, err := deserializer.Decode(raw, nil, &crd); err != nil {
			return nil, errors.Wrapf(err, "could not decode body")
		}
		if !isValidKind(crd.Spec.Names.Kind) {
			return nil, nil
		}
		label = crd.Labels
	case v1GVR:
		crd := apiextensionsv1.CustomResourceDefinition{}
		if _, _, err := deserializer.Decode(raw, nil, &crd); err != nil {
			return nil, errors.Wrapf(err, "could not decode body")
		}
		if !isValidKind(crd.Spec.Names.Kind) {
			return nil, nil
		}
		label = crd.Labels
	default:
		return nil, fmt.Errorf("expect resource to be one of [%v, %v] but got %v", v1beta1GVR, v1GVR, ar.Request.Resource)
	}

	return label, nil
}

// validateCRDDeletion return nil if crd deletion is valid
func (s *Server) validateCRDDeletion(labels map[string]string) error {
	switch val := labels[DeletionAllowKey]; val {
	case DeletionAllowTypeAlways:
		return nil

	case DeletionAllowTypeCascading:
		ingressList := &networkextensionv1.IngressList{}
		ingressCnt := 0
		if err := s.k8sClient.List(context.Background(), ingressList); err != nil {
			return fmt.Errorf("failed to list CRs of ingress: %v", err)
		}
		for _, obj := range ingressList.Items {
			if obj.GetDeletionTimestamp() == nil {
				ingressCnt++
			}
		}

		portPoolList := &networkextensionv1.PortPoolList{}
		portPoolCnt := 0
		if err := s.k8sClient.List(context.Background(), portPoolList); err != nil {
			return fmt.Errorf("failed to list CRs of portPool: %v", err)
		}
		for _, obj := range portPoolList.Items {
			if obj.GetDeletionTimestamp() == nil {
				portPoolCnt++
			}
		}

		listenerList := &networkextensionv1.ListenerList{}
		listenerCnt := 0
		if err := s.k8sClient.List(context.Background(), listenerList); err != nil {
			return fmt.Errorf("failed to list CRs of listener: %v", err)
		}
		for _, obj := range listenerList.Items {
			if obj.GetDeletionTimestamp() == nil {
				listenerCnt++
			}
		}

		portBindingList := &networkextensionv1.PortBindingList{}
		portBindingCnt := 0
		if err := s.k8sClient.List(context.Background(), portBindingList); err != nil {
			return fmt.Errorf("failed to list CRs of portBinding: %v", err)
		}
		for _, obj := range portBindingList.Items {
			if obj.GetDeletionTimestamp() == nil {
				portBindingCnt++
			}
		}

		activeCnt := ingressCnt + portPoolCnt + listenerCnt + portBindingCnt
		if activeCnt > 0 {
			return fmt.Errorf("forbidden by ResourceDeletionAllow for %s=%s and active CRs %d>0( ingress=%d, "+
				"portpool=%d, listener=%d, portBinding=%d)", DeletionAllowKey, DeletionAllowTypeCascading, activeCnt,
				ingressCnt, portPoolCnt, listenerCnt, portBindingCnt)
		}

	default:
		return fmt.Errorf("forbidden by ResourceDeletionAllow, set labels %s=%s to allow operation",
			DeletionAllowKey, DeletionAllowTypeAlways)
	}
	return nil
}

// isValidKind return true if CRD's kind is valid
func isValidKind(kind string) bool {
	switch kind {
	case constant.KindListener,
		constant.KindIngress,
		constant.KindPortPool,
		constant.KindPortBinding:
		return true
	}

	return false
}
