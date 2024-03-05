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

package portbindingcontroller

import (
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
)

const (
	// ReasonPortBindingCreatSuccess event reason for create port binding success
	ReasonPortBindingCreatSuccess = "CreateSuccess"
	// MsgPortBindingCreateSuccess msg create port binding success
	MsgPortBindingCreateSuccess = "PortBinding create success"
	// ReasonPortBindingNotReady reason port binding status is not ready
	ReasonPortBindingNotReady = "NotReady"
	// MsgPortBindingNotReady msg wait port binding status to be ready
	MsgPortBindingNotReady = "Waiting item status to be ready, unready num %d"
	// ReasonPortBindingReady reason port binding status be ready
	ReasonPortBindingReady = "Ready"
	// MsgPortBindingReady msg port binding status be ready
	MsgPortBindingReady = "PortBinding status is ready"
	// ReasonPortBindingUpdatePodSuccess reason port binding update related pod success
	ReasonPortBindingUpdatePodSuccess = "UpdatePodSuccess"
	// MsgPortBindingUpdatePodSuccess msg port binding update related pod success
	MsgPortBindingUpdatePodSuccess = "Update related pod success"
	// ReasonPortBindingEnsureFailed reason port binding ensure failed
	ReasonPortBindingEnsureFailed = "EnsureFailed"
	// MsgPortBindingEnsureFailed msg port binding ensure failed
	MsgPortBindingEnsureFailed = "PortBinding ensure failed, err: %s"
	// ReasonPortBindingCleanSuccess reason port binding clean success
	ReasonPortBindingCleanSuccess = "CleanSuccess"
	// MsgPortBindingCleanSuccess msg port binding clean success
	MsgPortBindingCleanSuccess = "PortBinding items clean success"
	// ReasonPortBindingCleaning reason port binding waiting item status to be cleaned
	ReasonPortBindingCleaning = "Cleaning"
	// MsgPortBindingCleaning msg port binding waiting item status to be cleaned
	MsgPortBindingCleaning = "PortBinding is cleaning items, not cleaned num %d"
	// ReasonPortBindingCleanFailed reason port binding clean failed
	ReasonPortBindingCleanFailed = "CleanFailed"
	// MsgPortBindingCleanFailed msg port binding clean failed
	MsgPortBindingCleanFailed = "PortBinding clean failed, err: %s"
)

func (pbr *PortBindingReconciler) recordEvent(portBinding *networkextensionv1.PortBinding, eType, reason, msg string) {
	if pbr.eventer == nil {
		return
	}
	pbr.eventer.Event(portBinding, eType, reason, msg)
}

func (pbh *portBindingHandler) recordEvent(portBinding *networkextensionv1.PortBinding, eType, reason, msg string) {
	if pbh.eventer == nil {
		return
	}
	pbh.eventer.Event(portBinding, eType, reason, msg)
}
