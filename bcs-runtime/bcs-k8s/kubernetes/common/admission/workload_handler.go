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

package admission

import (
	"encoding/json"
	"fmt"

	v1 "k8s.io/api/admission/v1"
	"k8s.io/klog/v2"

	gdv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamedeployment-operator/pkg/apis/tkex/v1alpha1"
	gssv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamestatefulset-operator/pkg/apis/tkex/v1alpha1"
	hookv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/bcs-hook/apis/tkex/v1alpha1"
)

// handleWorkload handles admission requests when deleting workload
func (whsvr *WebhookServer) handleWorkload(ar v1.AdmissionReview) *v1.AdmissionResponse {
	klog.V(2).Infof("handlinging workload: kind=%v, namespace=%v, name=%v",
		ar.Request.Kind.Kind, ar.Request.Namespace, ar.Request.Name)

	reviewResponse := v1.AdmissionResponse{}
	reviewResponse.Allowed = true

	if ar.Request.Operation != v1.Delete || ar.Request.SubResource != "" {
		return &reviewResponse
	}
	if len(ar.Request.OldObject.Raw) == 0 {
		klog.Warningf("Skip to validate workload deletion for no old object, "+
			"maybe because of Kubernetes version < 1.16",
			ar.Request.Name)
		return &reviewResponse
	}

	var err error
	switch ar.Request.Kind.Kind {
	case HookTemplateKind:
		err = whsvr.validateHookTemplateDeletion(ar.Request.OldObject.Raw)
	case gssv1alpha1.Kind:
		err = whsvr.validateGameStatefulSetDeletion(ar.Request.OldObject.Raw)
	case gdv1alpha1.Kind:
		err = whsvr.validateGameDeploymentDeletion(ar.Request.OldObject.Raw)
	default:
		err = nil
	}

	if err != nil {
		return toV1AdmissionResponse(err)
	}

	return &reviewResponse
}

func (whsvr *WebhookServer) validateHookTemplateDeletion(raw []byte) error {
	var hooktem hookv1alpha1.HookTemplate
	err := json.Unmarshal(raw, &hooktem)
	if err != nil {
		klog.Error(err)
		return err
	}

	switch val := hooktem.Labels[DeletionAllowKey]; val {
	case DeletionAllowTypeAlways:
		return nil
	default:
		return fmt.Errorf("forbidden by ResourceDeletionAllow, set labels %s=%s to allow operation",
			DeletionAllowKey, DeletionAllowTypeAlways)
	}
}

func (whsvr *WebhookServer) validateGameStatefulSetDeletion(raw []byte) error {
	var gss gssv1alpha1.GameStatefulSet
	err := json.Unmarshal(raw, &gss)
	if err != nil {
		klog.Error(err)
		return err
	}

	switch val := gss.Labels[DeletionAllowKey]; val {
	case DeletionAllowTypeAlways:
		return nil
	case DeletionAllowTypeCascading:
		if gss.Spec.Replicas != nil && *gss.Spec.Replicas > 0 {
			return fmt.Errorf("forbidden by ResourceDeletionAllow for %s=%s and replicas %d>0",
				DeletionAllowKey, DeletionAllowTypeCascading, *gss.Spec.Replicas)
		}

	default:
		return fmt.Errorf("forbidden by ResourceDeletionAllow, set labels %s=%s to allow operation",
			DeletionAllowKey, DeletionAllowTypeAlways)
	}
	return nil
}

func (whsvr *WebhookServer) validateGameDeploymentDeletion(raw []byte) error {
	var gd gdv1alpha1.GameDeployment
	err := json.Unmarshal(raw, &gd)
	if err != nil {
		klog.Error(err)
		return err
	}

	switch val := gd.Labels[DeletionAllowKey]; val {
	case DeletionAllowTypeAlways:
		return nil
	case DeletionAllowTypeCascading:
		if gd.Spec.Replicas != nil && *gd.Spec.Replicas > 0 {
			return fmt.Errorf("forbidden by ResourceDeletionAllow for %s=%s and replicas %d>0",
				DeletionAllowKey, DeletionAllowTypeCascading, *gd.Spec.Replicas)
		}

	default:
		return fmt.Errorf("forbidden by ResourceDeletionAllow, set labels %s=%s to allow operation",
			DeletionAllowKey, DeletionAllowTypeAlways)
	}
	return nil
}
