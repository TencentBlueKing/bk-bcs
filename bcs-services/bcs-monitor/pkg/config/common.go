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

import "time"

const (
	// tsdb 块最小时间, 默认 2 个小时
	MinBlockDuration = time.Hour * 2
	// tsdb 块最大时间, 默认 2 个小时, 最大/最小时间需要一致
	MaxBlockDuration = time.Hour * 2
	// 数据滚动时间, 默认 2 天
	RetentionDuration = time.Hour * 24 * 2
)

// EndpointConfig
type EndpointConfig struct {
	Address     string        `yaml:"address" mapstructure:"address"`
	GracePeriod time.Duration `yaml:"grace_period" mapstructure:"grace_period"`
}

// TSDBConfig
type TSDBConfig struct {
	MinBlockDuration time.Duration `yaml:"min-block-duration" mapstructure:"min-block-duration"`
	MaxBlockDuration time.Duration `yaml:"max-block-duration" mapstructure:"max-block-duration"`
	Retention        time.Duration `yaml:"retention" mapstructure:"retention"`
}

// Init
func (c *TSDBConfig) Init() error {
	c.MinBlockDuration = MinBlockDuration
	c.MaxBlockDuration = MaxBlockDuration
	c.Retention = RetentionDuration
	return nil
}
