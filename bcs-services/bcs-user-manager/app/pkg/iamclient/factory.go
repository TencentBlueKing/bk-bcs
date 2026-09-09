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

// NewFactory 按租户缓存 IAM client，进程内只注册一次 prometheus 指标。
func NewFactory(baseOpt iam.Options) func(tenantID string) iam.PermMigrateClient {
	var clients sync.Map
	return func(tenantID string) iam.PermMigrateClient {
		if tenantID == "" {
			tenantID = iam.DefaultTenantId
		}
		if cached, ok := clients.Load(tenantID); ok {
			return cached.(iam.PermMigrateClient)
		}

		opt := baseOpt
		opt.TenantId = tenantID
		// NewIamMigrateClient 在 Metric=true 时会 MustRegister，只允许第一次创建带上指标。
		opt.Metric = false
		metricRegisterOnce.Do(func() {
			opt.Metric = baseOpt.Metric
		})

		iamCli, err := iam.NewIamMigrateClient(&opt)
		if err != nil {
			panic(err)
		}
		actual, _ := clients.LoadOrStore(tenantID, iamCli)
		return actual.(iam.PermMigrateClient)
	}
}
