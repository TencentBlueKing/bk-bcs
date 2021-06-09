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

package types

import (
	"github.com/Tencent/bk-bcs/bcs-services/bcs-logbeat-sidecar/metric"
)

// Yaml is the structure viewed of log config file, contains metric info
type Yaml struct {
	Local           []Local                 `yaml:"local"`
	Metric          *metric.LogFileInfoType `yaml:"-"`
	BCSLogConfigKey string                  `yaml:"-"`
}

// Local is a single log collection task with single dataid
type Local struct {
	DataID       int               `yaml:"dataid"`
	OutputFormat string            `yaml:"output_format"`
	Paths        []string          `yaml:"paths"`
	ToJSON       bool              `yaml:"to_json"`
	Package      *bool             `yaml:"package,omitempty"`
	ExtMeta      map[string]string `yaml:"ext_meta"`

	//stdout dataid
	StdoutDataid string `yaml:"-"`
	//nonstandard log dataid
	NonstandardDataid string `yaml:"-"`
	//nonstandard paths
	NonstandardPaths []string `yaml:"-"`
	//host paths
	HostPaths []string `yaml:"-"`
	//log tags
	LogTags map[string]string `yaml:"-"`
}
