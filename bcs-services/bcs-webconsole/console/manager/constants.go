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
package manager

const (
	webConsoleHeartbeatKey = "bcs::web_console::heartbeat"
	Namespace              = "web-console"

	// DefaultCols DefaultRows 1080p页面测试得来
	DefaultCols = 211
	DefaultRows = 25

	// WebsocketPingInterval ping/pong时间间隔
	WebsocketPingInterval = 10
	// CleanUserPodInterval pod清理时间间隔
	CleanUserPodInterval = 60
	// LockShift 锁偏差时间常量
	LockShift = -2

	// TickTimeout 链接自动断开时间, 30分钟
	TickTimeout = 60 * 30
	// LoginTimeout 自动登出时间
	LoginTimeout = 60 * 60 * 24
	// UserPodExpireTime 清理POD，4个小时
	UserPodExpireTime = 3600 * 4
	// UserCtxExpireTime Context 过期时间, 12个小时
	UserCtxExpireTime = 3600 * 12

	//InterNel 用户自己集群
	InterNel = "internel"
	//EXTERNAL 平台集群
	EXTERNAL = "external"
)
