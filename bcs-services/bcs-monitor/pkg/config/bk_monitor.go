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
	"time"

	"github.com/pkg/errors"
)

// BKMonitorConf :
type BKMonitorConf struct {
	URL                  string    `yaml:"url"`          // unify-query 访问地址
	EnableGrey           bool      `yaml:"enable_grey"`  // 是否使用灰度
	MetadataURL          string    `yaml:"metadata_url"` // 元数据地址
	AgentEnableAfter     string    `yaml:"agent_enable_after"`
	AgentEnableAfterTime time.Time `yaml:"-"`
}

func (m *BKMonitorConf) init() error {
	// 默认全部开启agent
	if m.AgentEnableAfter == "" {
		m.AgentEnableAfterTime = time.Time{}
		return nil
	}

	t, err := time.Parse("2006-01-02", m.AgentEnableAfter)
	if err != nil {
		return errors.Wrap(err, "agent_enable_after")
	}
	m.AgentEnableAfterTime = t

	return nil
}

// defaultBKMonitorConf 默认配置
func defaultBKMonitorConf() *BKMonitorConf {
	c := &BKMonitorConf{
		URL:              "",
		MetadataURL:      "",
		AgentEnableAfter: "",
	}
	return c
}
