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

// Package envs xxx
package envs

import (
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/common/constant"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/stringx"
)

// 以下变量值可通过环境变量指定
var (
	// LocalIP 服务运行 Pod IP，用于向etcd注册服务
	LocalIP   = stringx.GetEnv("localIp", "")
	LocalIPV6 = stringx.GetEnv("localIpv6", "")
	// MongoPwd mongo password
	MongoPwd = stringx.GetEnv("mongoPwd", "")

	// BCSGatewayToken bcs gateway token
	BCSGatewayToken    = stringx.GetEnv("gatewayToken", "")
	BCSNamespacePrefix = stringx.GetEnv("BCS_NAMESPACE_PREFIX", "bcs")

	// AnnotationKeyProjectCode shared cluster project code annotation key
	AnnotationKeyProjectCode = stringx.GetEnv("annotationKeyProjectCode", constant.AnnotationKeyProjectCode)
)
