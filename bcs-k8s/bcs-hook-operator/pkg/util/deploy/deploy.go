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
 *
 */

package deploy

const (
	//DefaultRevisionHistoryLimit limit version for rollback, maybe setting in command line args?
	DefaultRevisionHistoryLimit = 10
	HookRunController           = "HookRun"

	DeployMode_OPERATOR  = "operator"
	DeployMode_DAEMONSET = "daemonset"
)

var DeployMode string

func SetDeployMode(deployMode string) {
	DeployMode = deployMode
}

// GetOperatorName for operator register in kubernetes
func GetOperatorName() string {
	if DeployMode == DeployMode_DAEMONSET {
		return "bcs-hook-daemonset"
	}else {
		return "bcs-hook-operator"
	}
}
