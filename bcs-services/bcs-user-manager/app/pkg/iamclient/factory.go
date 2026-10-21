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

// Package iamclient caches IAM clients per tenant.
package iamclient

import (
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
)

var metricRegisterOnce sync.Once

// NewAuthFactory 按 enable_v4 写入 Version 后创建鉴权 client。
// enable_v4=true 时走 IAM V4（v4PermClient），否则走 V3 SDK。
func NewAuthFactory(baseOpt iam.Options, enableV4 bool, v4Host string) func(tenantID string) iam.PermClient {
	iam.ApplyV4Config(&baseOpt, enableV4, v4Host)
	return NewFactory(baseOpt)
}

// NewFactory 按租户缓存 IAM client，进程内只注册一次 prometheus 指标。
// 调用方应已通过 ApplyV4Config / NewAuthFactory 写好 Version。
func NewFactory(baseOpt iam.Options) func(tenantID string) iam.PermClient {
	var clients sync.Map
	return func(tenantID string) iam.PermClient {
		if tenantID == "" {
			tenantID = iam.DefaultTenantId
		}
		if cached, ok := clients.Load(tenantID); ok {
			return cached.(iam.PermClient)
		}

		opt := baseOpt
		opt.TenantId = tenantID
		// NewIamClient 在 Metric=true 时会 MustRegister，只允许第一次创建带上指标。
		opt.Metric = false
		metricRegisterOnce.Do(func() {
			opt.Metric = baseOpt.Metric
		})

		iamCli, err := iam.NewIamClient(&opt)
		if err != nil {
			panic(err)
		}
		actual, _ := clients.LoadOrStore(tenantID, iamCli)
		return actual.(iam.PermClient)
	}
}

// NewV3MigrateClient 始终走 V3 SDK，供权限模型 migration 与 V3 provider token。
func NewV3MigrateClient(baseOpt iam.Options) (iam.PermMigrateClient, error) {
	baseOpt.Version = ""
	baseOpt.Metric = false
	return iam.NewIamMigrateClient(&baseOpt)
}
