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

// Package ctxkey xxx
package ctxkey

import "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/types"

const (
	// RequestIDKey xxx
	RequestIDKey = types.ContextKey("requestID")
	// UsernameKey xxx
	UsernameKey = types.ContextKey("username")
	// ProjKey xxx
	ProjKey = types.ContextKey("project")
	// ClusterKey xxx
	ClusterKey = types.ContextKey("cluster")
	// LangKey 语言版本
	LangKey = types.ContextKey("lang")
)
