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

// Package hwcheck xxx
package hwcheck

import (
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-reporter/internal/util"
)

// Options bcs log options
type Options struct {
	Interval          int             `json:"interval" yaml:"interval"`
	LogFileConfigList []LogFileConfig `json:"logFileConfigs" yaml:"logFileConfigs"`
}

// LogFileConfig xxx
type LogFileConfig struct {
	Path        string   `json:"path" yaml:"path"`
	KeyWordList []string `json:"keyWordList" yaml:"keyWordList"`
	Rule        string   `json:"rule" yaml:"rule"`
	logFile     *util.LogFile
}

// Validate validate options
func (o *Options) Validate() error {
	return nil
}
