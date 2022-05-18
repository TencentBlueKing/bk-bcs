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

// StoreGWConf
type StoreGWConf struct {
	HTTP    *EndpointConfig `yaml:"http" mapstructure:"http"`
	GRPC    *EndpointConfig `yaml:"grpc" mapstructure:"grpc"`
	DataDir string          `yaml:"data_dir" mapstructure:"data_dir"`
}

// Init
func (s *StoreGWConf) Init() error {
	s.DataDir = "./data/store"

	s.HTTP = &EndpointConfig{
		Address:     "127.0.0.1:10212",
		GracePeriod: time.Minute * 2,
	}

	s.GRPC = &EndpointConfig{
		Address:     "127.0.0.1:10213",
		GracePeriod: time.Minute * 2,
	}

	return nil
}
