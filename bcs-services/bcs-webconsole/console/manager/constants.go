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

// Package manager xxx
package manager

const (
	// LongDateTimeLayout xxx
	LongDateTimeLayout = "20060102150405"

	// WebsocketPingInterval ping/pong时间间隔
	WebsocketPingInterval = 10

	// LoginTimeout 自动登出时间
	LoginTimeout = 60 * 60 * 24

	// InputLineBreaker 输入分行标识
	InputLineBreaker = "\r"
	// OutputLineBreaker 输出分行标识
	OutputLineBreaker = "\r\n"

	// AnsiEscape bash 颜色标识
	AnsiEscape = "r\"\\x1B\\[[0-?]*[ -/]*[@-~]\""

	// StdinChannel xxx
	// same as kubernete client
	// https://github.com/kubernetes-client/python/blob/master/kubernetes/base/stream/ws_client.py#L3
	StdinChannel = "0"
	// StdoutChannel xxx
	StdoutChannel = "1"
	// StderrChannel xxx
	StderrChannel = "2"
	// ErrorChannel xxx
	ErrorChannel = "3"
	// ResizeChannel xxx
	ResizeChannel = "4"

	helloBcsMessage = "Welcome to the BCS Console"

	// ShellSH sh type
	ShellSH = "sh"
	// ShellBash bash type
	ShellBash = "bash"
)

var (
	// ShCommand sh 命令
	ShCommand = []string{
		"/bin/sh",
		"-c",
		"export TERM=xterm-256color; export PS1=\"\\u:\\W$ \"; exec /bin/sh",
	}
	// BashCommand bash 命令
	BashCommand = []string{
		"/bin/bash",
		"-c",
		"export TERM=xterm-256color; export PS1=\"\\u:\\W$ \"; exec /bin/bash",
	}
)
