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

package clbingress

import (
	ingressv1 "github.com/Tencent/bk-bcs/bcs-k8s/kubedeprecated/apis/clb/v1"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-clb-controller/pkg/model"
)

// IngressRegistry interface for clb ingress rule discovery
type Registry interface {
	AddIngressHandler(handler model.EventHandler)
	ListIngresses() ([]*ingressv1.ClbIngress, error)
	GetIngress(name string) (*ingressv1.ClbIngress, error)
	SetIngress(*ingressv1.ClbIngress) error
}
