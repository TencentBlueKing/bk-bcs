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

package podmanager

import "time"

const (
	// heartbeat:{run_env}
	webConsolePodHeartbeatKey     = "bcs::webconsole::heartbeat::pods:%s"
	webConsoleClusterHeartbeatKey = "bcs::webconsole::heartbeat::clusters:%s"
	// Namespace ..
	Namespace = "bcs-webconsole"
	// KubectlContainerName ..
	KubectlContainerName = "kubectl"
	// CleanUserPodInterval pod清理时间间隔
	CleanUserPodInterval = time.Second * 60

	// UserPodExpireTime 清理POD，4个小时
	UserPodExpireTime = time.Hour * 4

	// UserCtxExpireTime Context 过期时间, 12个小时
	UserCtxExpireTime    = 3600 * 12
	clusterExpireSeconds = 3600 * 24 * 7
)
