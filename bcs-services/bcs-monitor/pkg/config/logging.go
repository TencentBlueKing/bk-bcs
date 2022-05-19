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

package config

import (
	"github.com/TencentBlueKing/bkmonitor-kits/logger"
)

// LogConf : config for logging
type LogConf struct {
	logger.Options `yaml:",inline"`
}

// Init : init default logging config
func (c *LogConf) init() error {
	logger.SetOptions(c.Options)
	return nil
}

// SetByCmd 命令配置, 优先级最高
func (c *LogConf) SetByCmd(level string) error {
	if level == "" {
		return nil
	}

	c.Level = level
	return c.init()
}

// defaultLogConf 默认配置
func defaultLogConf() *LogConf {
	opt := logger.Options{
		Stdout: true,
		Level:  "info",
		Format: "logfmt",
	}
	c := &LogConf{
		Options: opt,
	}
	return c
}
