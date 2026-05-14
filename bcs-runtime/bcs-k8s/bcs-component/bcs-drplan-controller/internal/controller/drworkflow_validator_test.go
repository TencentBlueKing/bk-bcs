/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2023 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package controller

import (
	"testing"

	drv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-drplan-controller/api/v1alpha1"
)

func TestKubernetesResourceActionValidatorApplyAllowsMissingRollback(t *testing.T) {
	validator := &KubernetesResourceActionValidator{}
	action := &drv1alpha1.Action{
		Name: "apply-manifest",
		Type: drv1alpha1.ActionTypeKubernetesResource,
		Resource: &drv1alpha1.KubernetesResourceAction{
			Operation: drv1alpha1.OperationApply,
			Manifest:  "apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: demo\n",
		},
	}

	if errors := validator.Validate(action, 0); len(errors) != 0 {
		t.Fatalf("expected Apply without rollback to be valid, got %v", errors)
	}
}

func TestKubernetesResourceActionValidatorPatchRequiresRollback(t *testing.T) {
	validator := &KubernetesResourceActionValidator{}
	action := &drv1alpha1.Action{
		Name: "patch-manifest",
		Type: drv1alpha1.ActionTypeKubernetesResource,
		Resource: &drv1alpha1.KubernetesResourceAction{
			Operation: drv1alpha1.OperationPatch,
			Manifest:  "apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: demo\n",
		},
	}

	if errors := validator.Validate(action, 0); len(errors) == 0 {
		t.Fatal("expected Patch without rollback to be invalid")
	}
}
