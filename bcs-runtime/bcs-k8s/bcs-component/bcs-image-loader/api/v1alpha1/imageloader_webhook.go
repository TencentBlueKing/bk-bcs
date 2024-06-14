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
 *
 */

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var imageloaderlog = logf.Log.WithName("imageloader-resource")

// SetupWebhookWithManager will setup the manager to manage the webhooks
func (r *ImageLoader) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// NOCC:tosa/linelength(设计如此)
// +kubebuilder:webhook:path=/validate-imageloader,mutating=false,failurePolicy=fail,sideEffects=None,groups=tkex.tencent.com,resources=imageloaders,verbs=create;update,versions=v1alpha1,name=vimageloader.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &ImageLoader{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *ImageLoader) ValidateCreate() (admission.Warnings, error) {
	imageloaderlog.Info("validate create", "name", r.Name)

	return nil, nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *ImageLoader) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	imageloaderlog.Info("validate update", "name", r.Name)

	return nil, nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *ImageLoader) ValidateDelete() (admission.Warnings, error) {
	imageloaderlog.Info("validate delete", "name", r.Name)

	return nil, nil
}
