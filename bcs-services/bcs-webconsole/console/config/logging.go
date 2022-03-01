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
	"path/filepath"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
)

// LogConf : config for logging
type LogConf struct {
	Level    string `yaml:"level"`
	File     string `yaml:"file"`
	Stderr   bool   `yaml:"stderr"`
	CmdFile  string `yaml:"-"`
	CmdLevel string `yaml:"-"`
}

// Init : init default logging config
func (c *LogConf) Init() error {
	// only for development
	c.Level = "info"
	c.File = ""
	c.Stderr = true
	c.CmdFile = ""
	c.CmdLevel = "info"

	return nil
}

// InitBlog 初始化 blog 模块, 注意只能初始化一次
func (c *LogConf) InitBlog() error {
	var blogLevel int32
	blogDir := ""

	// blog 只有 0 和 3 个等级
	switch c.Level {
	case "debug":
		blogLevel = 3
	default:
		blogLevel = 0
	}

	// 不会自动创建目录, 需要管理员手动创建
	if c.File != "" {
		logFile, err := filepath.Abs(c.File)
		if err != nil {
			return err
		}

		blogDir = filepath.Dir(logFile)
	}

	blogConf := conf.LogConfig{
		Verbosity:    blogLevel,
		AlsoToStdErr: c.Stderr,
		LogDir:       blogDir,
		LogMaxSize:   500,
		LogMaxNum:    10,
	}

	blog.InitLogs(blogConf)
	return nil
}
