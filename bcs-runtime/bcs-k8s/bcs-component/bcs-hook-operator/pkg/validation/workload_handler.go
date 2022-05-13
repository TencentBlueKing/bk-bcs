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

	hookv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/bcs-hook/apis/tkex/v1alpha1"
)

// handleWorkload handles admission requests when deleting workload
func (whsvr *WebhookServer) handleWorkload(ar v1.AdmissionReview) *v1.AdmissionResponse {
	klog.V(2).Info("handlinging hooktemplate")

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
	var hooktem hookv1alpha1.HookTemplate

	err := json.Unmarshal(raw, &hooktem)
	if err != nil {
		klog.Error(err)
		return toV1AdmissionResponse(err)
	}

	if err := whsvr.validateHookTemplateDeletion(hooktem.Labels); err != nil {
		return toV1AdmissionResponse(err)
	}

	return &reviewResponse
}

func (whsvr *WebhookServer) validateHookTemplateDeletion(labels map[string]string) error {
	switch val := labels[DeletionAllowKey]; val {
	case DeletionAllowTypeAlways:
		return nil
	default:
		return fmt.Errorf("forbidden by ResourceDeletionAllow, set labels %s=%s to allow operation",
			DeletionAllowKey, DeletionAllowTypeAlways)
	}
}
