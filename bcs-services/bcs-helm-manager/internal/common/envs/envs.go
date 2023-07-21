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

// Package envs xxx
package envs

import (
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/utils/envx"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/utils/stringx"
)

// 以下变量值可通过环境变量指定（仅用于单元测试）
var (
	// TestProjectCode 单测指定的项目 Code
	TestProjectCode = envx.GetEnv("TEST_PROJECT_CODE", "blueking")
	// TestProjectID 单测指定的项目 ID
	TestProjectID = envx.GetEnv("TEST_PROJECT_ID", stringx.Rand(32, ""))
	// TestSharedClusterID 单测指定的共享集群 ID
	TestSharedClusterID = envx.GetEnv("TEST_SHARED_CLUSTER_ID", "BCS-K8S-S99999")
)
