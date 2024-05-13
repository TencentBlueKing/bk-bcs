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

// Package fake xxx
package fake

import (
	"encoding/json"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-webhook-server/internal/pluginutil"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-webhook-server/internal/types"
)

// Hooker fake hooker
type Hooker struct {
}

// Init implements plugin interface
func (h *Hooker) Init(configFilePath string) error {
	return nil
}

// Handle implements plugin interface
func (h *Hooker) Handle(ar v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
	req := ar.Request

	if req.Operation != v1beta1.Create {
		return &v1beta1.AdmissionResponse{Allowed: true}
	}

	// when the kind is not Pod, ignore hook
	if req.Kind.Kind != "Pod" {
		return &v1beta1.AdmissionResponse{Allowed: true}
	}
	pod := &corev1.Pod{}
	if err := json.Unmarshal(req.Object.Raw, pod); err != nil {
		blog.Errorf("cannot decode raw object %s to pod, err %s", string(req.Object.Raw), err.Error())
		return pluginutil.ToAdmissionResponse(err)
	}

	initContainers := append(pod.Spec.InitContainers, corev1.Container{ // nolint
		Image: "fakeimgae",
	})
	var patch []types.PatchOperation
	patch = append(patch, types.PatchOperation{
		Op:    "replace",
		Path:  "/spec/initContainers",
		Value: initContainers,
	})
	patchesBytes, _ := json.Marshal(patch)
	reviewResponse := &v1beta1.AdmissionResponse{
		Allowed: true,
		Patch:   patchesBytes,
		PatchType: func() *v1beta1.PatchType {
			pt := v1beta1.PatchTypeJSONPatch
			return &pt
		}(),
	}
	return reviewResponse
}

// AnnotationKey implements plugin interface
func (h *Hooker) AnnotationKey() string {
	return ""
}

// Close implements plugin interface
func (h *Hooker) Close() error {
	return nil
}
