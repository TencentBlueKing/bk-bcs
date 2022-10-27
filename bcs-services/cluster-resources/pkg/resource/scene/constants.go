/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * 	http://opensource.org/licenses/MIT
 *
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package scene

// 场景常量一般搭配 selectItems API 使用，用于区分场景对 selectItems 做禁用，屏蔽等操作
// 目前默认列出 workloads，bcs-crd，network 三类，可按需要添加

const (
	// IngTargetSVC 作为 ingress targetService 的 Service
	IngTargetSVC = "ing-target-svc"

	// IngTLSCert 作为 ingress tls 证书的 Secret
	IngTLSCert = "ing-tls-cert"

	// WorkloadImagePullSecret 作为工作负载镜像拉取凭证的 Secret
	WorkloadImagePullSecret = "workload-image-pull-secrets"
)
