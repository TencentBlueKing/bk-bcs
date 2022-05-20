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
	"encoding/json"
	"fmt"

	v1 "k8s.io/api/admission/v1"
	"k8s.io/klog/v2"

	gssv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamestatefulset-operator/pkg/apis/tkex/v1alpha1"
)

// handleWorkload handles admission requests when deleting workload
func (whsvr *WebhookServer) handleWorkload(ar v1.AdmissionReview) *v1.AdmissionResponse {
	klog.V(2).Info("handlinging gamestatefulset")

	reviewResponse := v1.AdmissionResponse{}
	reviewResponse.Allowed = true

	if ar.Request.Operation != v1.Delete || ar.Request.SubResource != "" {
		return &reviewResponse
	}
	if len(ar.Request.OldObject.Raw) == 0 {
		klog.Warningf("Skip to validate GameStatefulSet %s deletion for no old object, "+
			"maybe because of Kubernetes version < 1.16",
			ar.Request.Name)
		return &reviewResponse
	}

	raw := ar.Request.OldObject.Raw
	var gss gssv1alpha1.GameStatefulSet

	err := json.Unmarshal(raw, &gss)
	if err != nil {
		klog.Error(err)
		return toV1AdmissionResponse(err)
	}

	if err := whsvr.validateWorkloadDeletion(gss.Labels, gss.Spec.Replicas); err != nil {
		return toV1AdmissionResponse(err)
	}

	return &reviewResponse

}

func (whsvr *WebhookServer) validateWorkloadDeletion(labels map[string]string, replicas *int32) error {
	switch val := labels[DeletionAllowKey]; val {
	case DeletionAllowTypeAlways:
		return nil
	case DeletionAllowTypeCascading:
		if replicas != nil && *replicas > 0 {
			return fmt.Errorf("forbidden by ResourceDeletionAllow for %s=%s and replicas %d>0",
				DeletionAllowKey, DeletionAllowTypeCascading, *replicas)
		}

	default:
		return fmt.Errorf("forbidden by ResourceDeletionAllow, set labels %s=%s to allow operation",
			DeletionAllowKey, DeletionAllowTypeAlways)
	}
	return nil
}
