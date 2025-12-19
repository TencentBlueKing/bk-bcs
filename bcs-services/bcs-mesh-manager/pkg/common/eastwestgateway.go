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

// Package common 提供eastwest gateway相关的YAML模板
package common

import (
	"fmt"
	"strings"
)

// GatewayYAML Gateway的YAML模板
const GatewayYAML = `
apiVersion: networking.istio.io/v1alpha3
kind: Gateway
metadata: 
  name: istiod-gateway
  namespace: istio-system
  labels:
    created-by: bcs-mesh-manager
spec:
  selector:
    istio: bcs-istio-eastwestgateway
  servers:
    - port:
        name: tls-istiod
        number: 15012
        protocol: tls
      tls:
        mode: PASSTHROUGH        
      hosts:
        - "*"
    - port:
        name: tls-istiodwebhook
        number: 15017
        protocol: tls
      tls:
        mode: PASSTHROUGH          
      hosts:
        - "*"
`

// VirtualServiceYAML VirtualService的YAML模板
const VirtualServiceYAML = `
apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: istiod-vs
  namespace: istio-system
  labels:
    created-by: bcs-mesh-manager
spec:
  hosts:
  - "*"
  gateways:
  - istiod-gateway
  tls:
  - match:
    - port: 15012
      sniHosts:
      - "*"
    route:
    - destination:
        host: %s
        port:
          number: 15012
  - match:
    - port: 15017
      sniHosts:
      - "*"
    route:
    - destination:
        host: %s
        port:
          number: 443
`

// GetGatewayYAML 获取Gateway的YAML模板
func GetGatewayYAML() string {
	return GatewayYAML
}

// GetVirtualServiceYAML 获取VirtualService的YAML模板
func GetVirtualServiceYAML(revision string) string {
	istiodSvc := GenerateIstiodServiceFQDN(revision)
	return fmt.Sprintf(VirtualServiceYAML,
		istiodSvc,
		istiodSvc,
	)
}

// GetRevision 获取revision
func GetRevision(chartVersion string) string {
	revision := strings.ReplaceAll(chartVersion, ".", "-") // "1.18-bcs.2" -> "1-18-bcs-2"
	return revision
}

// GenerateIstiodServiceFQDN 生成istiod服务的完全限定域名
// revision: istio版本标识，如 "1-18-bcs-2"
func GenerateIstiodServiceFQDN(revision string) string {
	return fmt.Sprintf("istiod-%s.%s.%s", revision, IstioNamespace, KubernetesServiceSuffix)
}

// 东西向网关资源名称常量
const (
	// GatewayName Gateway资源名称
	GatewayName = "istiod-gateway"
	// VirtualServiceName VirtualService资源名称
	VirtualServiceName = "istiod-vs"
)

// 东西向网关资源类型常量
const (
	// GatewayKind Gateway资源类型
	GatewayKind = "Gateway"
	// VirtualServiceKind VirtualService资源类型
	VirtualServiceKind = "VirtualService"
)

// KubernetesServiceSuffix Kubernetes服务DNS后缀
const KubernetesServiceSuffix = "svc.cluster.local"
