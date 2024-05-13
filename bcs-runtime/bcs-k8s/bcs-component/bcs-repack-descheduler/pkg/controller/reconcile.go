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

package controller

import (
	"context"
	"reflect"
	"time"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-repack-descheduler/internal/utils"
	deschedulev1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-repack-descheduler/pkg/apis/tkex/v1alpha1"
)

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (m *ControllerManager) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	blog.Infof("reconcile received: %s/%s", req.Namespace, req.Name)
	var policy = new(deschedulev1alpha1.DeschedulePolicy)
	if err := m.client.Get(ctx, req.NamespacedName, policy); err != nil {
		if k8serrors.IsNotFound(err) {
			blog.Warnf("policy '%s' is deleted, skipped", req.NamespacedName.String())
			m.migrator.DeleteMigrateJob(policy)
			return ctrl.Result{}, nil
		}
	}
	if policy.DeletionTimestamp != nil {
		m.migrator.DeleteMigrateJob(policy)
		return ctrl.Result{}, nil
	}
	if !policy.Spec.Converge.Disabled {
		if err := m.migrator.CreateMigrateJob(policy); err != nil {
			blog.Errorf("policy '%s' create migrate job failed: %s", req.NamespacedName, err.Error())
			return ctrl.Result{
				Requeue:      true,
				RequeueAfter: 30 * time.Second,
			}, nil
		}
	} else {
		blog.Warnf("policy '%s' is disabled, no need create migrate job", req.NamespacedName)
		m.migrator.DeleteMigrateJob(policy)
	}
	m.migrator.SendCalculateJob(policy)
	return ctrl.Result{Requeue: true, RequeueAfter: 180 * time.Second}, nil
}

// predicate filter policy
func (m *ControllerManager) predicate() predicate.Predicate {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			newDP, okNew := e.ObjectNew.(*deschedulev1alpha1.DeschedulePolicy)
			oldDP, okOld := e.ObjectOld.(*deschedulev1alpha1.DeschedulePolicy)
			if !okNew || !okOld {
				return true
			}
			if newDP.ObjectMeta.DeletionTimestamp != nil {
				return true
			}
			if !reflect.DeepEqual(newDP.Spec, oldDP.Spec) {
				blog.Infof("policy '%s/%s' spec changed, newSpec: %s, oldSpec: %s",
					newDP.GetNamespace(), newDP.GetName(),
					utils.ToJsonString(newDP.Spec), utils.ToJsonString(oldDP.Spec))
				return true
			}
			return false
		},
	}
}
