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
	"errors"
	"time"
	_ "time/tzdata" // tzdata TODO

	"github.com/Tencent/bk-bcs/bcs-ui/pkg/i18n"
)

const (
	// DevEnv TODO
	DevEnv = "dev"
	// ProdEnv TODO
	ProdEnv = "prod"
	// LocalEnv 本地开发, 和前端区别
	LocalEnv = "local"
)

// BaseConf :
type BaseConf struct {
	TimeZone     string         `yaml:"time_zone"`
	LanguageCode string         `yaml:"language_code"`
	RunEnv       string         `yaml:"run_env"` // 前端依赖, 必须是 dev / prod
	Region       string         `yaml:"region"`
	Location     *time.Location `yaml:"-"`
	Domain       string         `yaml:"domain"`
}

// Init :
func (c *BaseConf) Init() error {
	var err error
	c.TimeZone = "Asia/Shanghai"
	c.LanguageCode = "en-us"
	c.RunEnv = DevEnv
	c.Region = "ce"
	c.Location, err = time.LoadLocation(c.TimeZone)
	if err != nil {
		return err
	}
	c.Domain = ""
	return nil
}

// InitBaseConf init base config
func (c *BaseConf) InitBaseConf() error {
	// if the configuration is incorrect, panic
	if !i18n.IsAvailableLanguage(c.LanguageCode) {
		return errors.New("invalid language configuration")
	}
	return nil
}
