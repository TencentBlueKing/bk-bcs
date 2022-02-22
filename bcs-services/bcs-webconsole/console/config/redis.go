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

type RedisConf struct {
	Host           string `yaml:"host"`
	Port           int    `yaml:"port"`
	Password       string `yaml:"password"`
	DB             int    `yaml:"db"`
	MaxPoolSize    int    `yaml:"max_pool_size"`
	MaxConnTimeout int    `yaml:"max_conn_timeout"`
	IdleTimeout    int    `yaml:"idle_timeout"`
	ReadTimeout    int    `yaml:"read_timeout"`
	WriteTimeout   int    `yaml:"write_timeout"`
}

func (c *RedisConf) Init() error {
	// only for development
	c.Host = "127.0.0.1"
	c.Port = 6379
	c.Password = ""
	c.DB = 0

	c.MaxPoolSize = 100
	c.MaxConnTimeout = 6
	c.IdleTimeout = 600
	c.ReadTimeout = 10
	c.WriteTimeout = 10
	return nil
}
