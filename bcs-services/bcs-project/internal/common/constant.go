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

package common

type ContextKey string

const (
	ServiceDomain        = "project.bkbcs.tencent.com"
	DefaultConfigPath    = "project.yaml"
	MicroMetaKeyHTTPPort = "httpport"

	// time layout
	TimeLayout = "2006-01-02 15:04:05"

	RequestIDKey ContextKey = "requestID"
	TraceIDKey   ContextKey = "string"
	MaxMsgSize   int        = 50 * 1024 * 1024
)
