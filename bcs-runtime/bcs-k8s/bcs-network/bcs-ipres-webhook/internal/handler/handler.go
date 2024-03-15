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

package handler

import (
	"encoding/json"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"k8s.io/api/admission/v1beta1"
	k8scorev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

const (
	PatchOpReplace                 = "replace"
	ContainerResourcePatchPath     = "/spec/containers/0/resources"
	InitContainerResourcePatchPath = "/spec/initContainers/0/resources"
)

// HookHandler handler for webhook
type HookHandler struct {
	podAnnotationKey   string
	podAnnotationValue string
	resourceName       string
}

// NewHookHandler create new hook handler
func NewHookHandler(annotationKey, annotationValue, resourceName string) *HookHandler {
	return &HookHandler{
		podAnnotationKey:   annotationKey,
		podAnnotationValue: annotationValue,
		resourceName:       resourceName,
	}
}

// HandleValidatingWebhook handle validating webhook
func (hh *HookHandler) HandleValidatingWebhook(ar v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
	return &v1beta1.AdmissionResponse{Allowed: true}
}

// HandleMutatingWebhook handle mutating wehbook
func (hh *HookHandler) HandleMutatingWebhook(ar v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
	req := ar.Request
	if req.Operation != v1beta1.Create {
		blog.Warnf("operation is not create, ignore")
		return &v1beta1.AdmissionResponse{Allowed: true}
	}
	// only hook create operation of pod
	if req.Kind.Kind != "Pod" {
		blog.Warnf("kind %s is not Pod", req.Kind.Kind)
		return errResponse(fmt.Errorf("kind %s is not Pod", req.Kind.Kind))
	}
	pod := &k8scorev1.Pod{}
	if err := json.Unmarshal(req.Object.Raw, pod); err != nil {
		blog.Warnf("decode %s to pod failed, err %s", string(req.Object.Raw), err.Error)
		return errResponse(fmt.Errorf("decode %s to pod failed, err %s", string(req.Object.Raw), err.Error()))
	}

	if pod.OwnerReferences != nil && len(pod.OwnerReferences) > 0 {
		blog.Infof("hook pod of %s/%s in ns %s",
			pod.OwnerReferences[0].Kind, pod.OwnerReferences[0].Name, ar.Request.Namespace)
	} else {
		blog.Infof("hook pod %s/%s", req.Name, req.Namespace)
	}

	if networkValue, ok := pod.Annotations[hh.podAnnotationKey]; ok {
		if networkValue != hh.podAnnotationValue {
			return &v1beta1.AdmissionResponse{Allowed: true}
		}
	} else {
		return &v1beta1.AdmissionResponse{Allowed: true}
	}

	// inject extended resource
	var patches []PatchOperation
	var patch PatchOperation
	var err error
	blog.Infof("do inject")
	if len(pod.Spec.Containers) > 0 {
		patch, err = hh.generatePatchData(pod.Spec.Containers[0].Resources, ContainerResourcePatchPath)
	} else if len(pod.Spec.InitContainers) > 0 {
		patch, err = hh.generatePatchData(pod.Spec.InitContainers[0].Resources, InitContainerResourcePatchPath)
	} else {
		blog.Infof("pod %s/%s has no containers or init containers", pod.GetName(), pod.GetNamespace())
		return &v1beta1.AdmissionResponse{Allowed: true}
	}
	if err != nil {
		blog.Warnf("generate patch data failed, err %s", err.Error())
		return errResponse(fmt.Errorf("generate patch data failed, err %s", err.Error()))
	}
	patches = append(patches, patch)
	patchesBytes, err := json.Marshal(patches)
	if err != nil {
		blog.Warnf("encoding patches faile, err %s", err.Error())
		return errResponse(fmt.Errorf("encoding patches faile, err %s", err.Error()))
	}
	return &v1beta1.AdmissionResponse{
		Allowed: true,
		Patch:   patchesBytes,
		PatchType: func() *v1beta1.PatchType {
			pt := v1beta1.PatchTypeJSONPatch
			return &pt
		}(),
	}
}

func (hh *HookHandler) generatePatchData(res k8scorev1.ResourceRequirements, path string) (PatchOperation, error) {
	if res.Limits == nil {
		res.Limits = make(k8scorev1.ResourceList)
	}
	res.Limits[k8scorev1.ResourceName(hh.resourceName)] = *resource.NewQuantity(1, resource.DecimalSI)
	return PatchOperation{
		Op:    PatchOpReplace,
		Path:  path,
		Value: res,
	}, nil
}
