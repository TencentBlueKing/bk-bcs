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

// Package thirdparty provides a client for interacting with bcs-thirdparty-service.
package thirdparty

const (
	// BcsHeaderClientKey client key in header
	BcsHeaderClientKey = "X-Bcs-Client"
	// BcsHeaderUsernameKey username key in header
	BcsHeaderUsernameKey = "X-Bcs-Username"
	// HeaderAuthorizationKey authorization key in header
	HeaderAuthorizationKey = "Authorization"
	// InnerModuleName full module name
	InnerModuleName = "bcs-push-manager"
	// ModuleThirdpartyServiceManager helm manager discovery name
	ModuleThirdpartyServiceManager = "bcsthirdpartyservice.bkbcs.tencent.com"
)
