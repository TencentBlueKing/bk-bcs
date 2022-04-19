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

package config

const (
	// ServiceDomain 域名，用于注册到APISIX
	ServiceDomain = "project.bkbcs.tencent.com"
	// DefaultConfigPath 配置路径
	DefaultConfigPath = "etc/project.yaml"
	// MicroMetaKeyHTTPPort 初始化micro服务需要的httpport
	MicroMetaKeyHTTPPort = "httpport"

	// TimeLayout time layout
	TimeLayout = "2006-01-02 15:04:05"

	// MaxMsgSize grpc限制的message的最大值
	MaxMsgSize int = 50 * 1024 * 1024
)
