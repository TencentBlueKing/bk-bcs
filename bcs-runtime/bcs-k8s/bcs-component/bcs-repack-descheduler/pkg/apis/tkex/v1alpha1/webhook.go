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

package v1alpha1

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// DeschedulePolicyAnnotator defines the annotator for webhook
type DeschedulePolicyAnnotator struct{}

var (
	validateCreate func(policy *DeschedulePolicy) error
	validateUpdate func(policy *DeschedulePolicy) error
)

// Path: /mutate-tkex-tencent-com-v1alpha1-deschedulepolicy
var _ webhook.CustomDefaulter = &DeschedulePolicyAnnotator{}

// Path: /validate-tkex-tencent-com-v1alpha1-deschedulepolicy
var _ webhook.CustomValidator = &DeschedulePolicyAnnotator{}

// SetupWebhookWithManager will set the webhook with controller-runtime manager
func (in *DeschedulePolicyAnnotator) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(&DeschedulePolicy{}).
		WithDefaulter(in).
		WithValidator(in).
		Complete()
}

// RegisterValidateCreate register the validate create function
func (in *DeschedulePolicyAnnotator) RegisterValidateCreate(f func(policy *DeschedulePolicy) error) {
	validateCreate = f
}

// RegisterValidateUpdate register the validate update function
func (in *DeschedulePolicyAnnotator) RegisterValidateUpdate(f func(policy *DeschedulePolicy) error) {
	validateUpdate = f
}

// Default implement function
func (in *DeschedulePolicyAnnotator) Default(ctx context.Context, obj runtime.Object) error {
	return nil
}

// ValidateCreate handle create action
func (in *DeschedulePolicyAnnotator) ValidateCreate(ctx context.Context, obj runtime.Object) error {
	if err := validateCreate(obj.(*DeschedulePolicy)); err != nil {
		return err
	}
	return nil
}

// ValidateUpdate handle update action
func (in *DeschedulePolicyAnnotator) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) error {
	if err := validateUpdate(newObj.(*DeschedulePolicy)); err != nil {
		return err
	}
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
// No need handle delete action.
func (in *DeschedulePolicyAnnotator) ValidateDelete(ctx context.Context, obj runtime.Object) error {
	blog.Infof("[Webhook] ValidateDelete received: %s", obj.(*DeschedulePolicy).Name)
	return nil
}
