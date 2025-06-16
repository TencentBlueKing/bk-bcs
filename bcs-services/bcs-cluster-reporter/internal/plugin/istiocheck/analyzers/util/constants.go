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

package util

import (
	"regexp"

	"istio.io/istio/pkg/config/constants"
)

const (
	DefaultClusterLocalDomain  = "svc." + constants.DefaultClusterLocalDomain
	ExportToNamespaceLocal     = "."
	ExportToAllNamespaces      = "*"
	IstioProxyName             = "istio-proxy"
	IstioOperator              = "istio-operator"
	MeshGateway                = "mesh"
	Wildcard                   = "*"
	MeshConfigName             = "istio"
	InjectionLabelName         = "istio-injection"
	InjectionLabelEnableValue  = "enabled"
	InjectionConfigMap         = "istio-sidecar-injector"
	InjectionConfigMapValue    = "values"
	InjectorWebhookConfigKey   = "sidecarInjectorWebhook"
	InjectorWebhookConfigValue = "enableNamespacesByDefault"
)

var fqdnPattern = regexp.MustCompile(`^(.+)\.(.+)\.svc\.cluster\.local$`)
