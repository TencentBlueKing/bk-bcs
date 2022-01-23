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
	"fmt"
	"io/ioutil"
)

// Config is the configuration for the federation apiserver.
type Config struct {
	BcsStorageAddress   string         `json:"bcs_storage_address" yaml:"bcs_storage_address"`
	BcsStorageToken     string         `json:"bcs_storage_token" yaml:"bcs_storage_token"`
	BcsStorageURLPrefix string         `json:"bcs_storage_url_prefix" yaml:"bcs_storage_url_prefix"`
	MemberCluster       string         `json:"member_cluster" yaml:"member_cluster"`
	APIResources        []*APIResource `json:"api_resources" yaml:"api_resources"`
}

// APIResource 待聚合的资源
type APIResource struct {
	Group   string `json:"group" yaml:"group"`
	Version string `json:"version" yaml:"version"`
	Kind    string `json:"kind" yaml:"kind"`
}

// ParseConfig 解析配置文件
func ParseConfig(filepath string) (*Config, error) {
	if filepath == "" {
		return nil, fmt.Errorf("配置文件不能为空")
	}
	configData, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}
	var c Config
	err = json.Unmarshal(configData, &c)
	return &c, err
}
