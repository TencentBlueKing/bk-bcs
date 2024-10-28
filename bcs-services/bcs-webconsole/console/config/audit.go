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

package config

import (
	"os"
	"path/filepath"
)

// AuditConf 终端session记录配置
type AuditConf struct {
	Enabled       bool   `yaml:"enabled"`
	DataDir       string `yaml:"data_dir"` // 格式如 ./data
	RetentionDays int    `yaml:"retention_days"`
}

// Init : AuditConf init
func (t *AuditConf) Init() {
	t.RetentionDays = 0
}

func (t *AuditConf) defaultPath() string {
	// NOCC:gas/error(设计如此)
	pwd, _ := os.Getwd()
	return filepath.Join(pwd, "data")
}
