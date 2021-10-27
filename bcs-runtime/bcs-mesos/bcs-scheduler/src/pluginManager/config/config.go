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
	"encoding/json"
	"io/ioutil"
	"os"

	typesplugin "github.com/Tencent/bk-bcs/bcs-common/common/plugin"
)

// PluginConfig configuration for loading plugin
type PluginConfig struct {
	Version      string                   `json:"version"`
	Name         string                   `json:"name"`
	Type         PluginType               `json:"type"`
	DefaultAtrrs []*typesplugin.Attribute `json:"defaultAttrs"`
	Timeout      int                      `json:"timeout"`
}

// PluginType type of plugin, dynamic lib/executable-file
type PluginType string

const (
	// DynamicPluginType dynamic lib
	DynamicPluginType PluginType = "dynamic-plugin"
	// ExecutablePluginType executable file
	ExecutablePluginType PluginType = "executable-plugin"
	// DefaultTimeout default timeout for plugin invocation
	DefaultTimeout int = 5
)

// NewConfig loading plugin config with specified file
func NewConfig(path string) (*PluginConfig, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	by, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	var conf *PluginConfig

	err = json.Unmarshal(by, &conf)
	if err != nil {
		return nil, err
	}

	if conf.Timeout <= 0 {
		conf.Timeout = DefaultTimeout
	}

	return conf, nil
}
