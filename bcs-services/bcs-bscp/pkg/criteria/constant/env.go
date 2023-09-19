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

// Package constant is for constant
package constant

const (
	// EnvHostIP container env in bind the env variable of the host IP
	EnvHostIP = "ENV_BK_BSCP_HOST_IP"
	// EnvMaxDownloadFileGoroutines is the env of sidecar maximum combined weight for concurrent download access.
	// the minimum value is 1 and the maximum value is 15.
	EnvMaxDownloadFileGoroutines = "ENV_BK_BSCP_MAX_DOWNLOAD_FILE_GOROUTINES"
	// EnvSuitTestSidecarWorkspace is the env of sidecar workspace, only sidecar suite use.
	EnvSuitTestSidecarWorkspace = "ENV_BSCP_TEST_SIDECAR_WORKSPACE"
)
