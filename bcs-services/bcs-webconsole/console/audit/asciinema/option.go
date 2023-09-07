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

package asciinema

import "time"

// Config asciinema文件Header信息
type Config struct {
	Title     string
	EnvShell  string
	EnvTerm   string
	Width     uint16
	Height    uint16
	Timestamp time.Time
}

// Option 可选配置
type Option func(options *Config)

// WithWidth 配置回放文件头部信息:终端宽度
func WithWidth(width uint16) Option {
	return func(options *Config) {
		options.Width = width
	}
}

// WithHeight 配置回放文件头部信息:终端高度
func WithHeight(height uint16) Option {
	return func(options *Config) {
		options.Height = height
	}
}

// WithTimestamp 配置回放文件头部信息:时间戳
func WithTimestamp(timestamp time.Time) Option {
	return func(options *Config) {
		options.Timestamp = timestamp
	}
}

// WithTitle 配置title信息
func WithTitle(title string) Option {
	return func(options *Config) {
		options.Title = title
	}
}

// WithEnvShell shell信息
func WithEnvShell(shell string) Option {
	return func(options *Config) {
		options.EnvShell = shell
	}
}

// WithEnvTerm 终端信息
func WithEnvTerm(term string) Option {
	return func(options *Config) {
		options.EnvTerm = term
	}
}
