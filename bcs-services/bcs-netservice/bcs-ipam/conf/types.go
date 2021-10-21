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

package conf

import (
	"encoding/json"
	types "github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/netservice"
	"io/ioutil"
)

//BcsConfig config item for ipam
type BcsConfig struct {
	ZkHost   string         `json:"zkHost"`
	TLS      *types.SSLInfo `json:"tls,omitempty"`
	Interval int            `json:"interval,omitempty"`
}

//LoadConfigFromFile load config item from file
func LoadConfigFromFile(f string) (*BcsConfig, error) {
	data, err := ioutil.ReadFile(f)
	if err != nil {
		return nil, err
	}
	config := &BcsConfig{}
	if err := json.Unmarshal(data, config); err != nil {
		return nil, err
	}
	return config, nil
}
