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

package namespacecontroller

import (
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
	k8scorev1 "k8s.io/api/core/v1"
)

func checkNodeBindLabel(ns *k8scorev1.Namespace) bool {
	if ns == nil || ns.Labels == nil {
		return false
	}
	value, ok := ns.Labels[networkextensionv1.NodePortBindingConfigMapNsLabel]
	if ok && value == networkextensionv1.NodePortBindingConfigMapNsLabelValue {
		return true
	}
	return false
}
