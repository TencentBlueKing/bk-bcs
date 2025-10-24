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
	"encoding/json"
	"os"
)

// Parse parse all config
func Parse(configPath string) (*SyncConfig, error) {
	// 1. 打开 JSON 文件
	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// 2. 创建一个结构体实例
	var syncConf SyncConfig

	// 3. 解码 JSON 到 struct
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&syncConf); err != nil {
		return nil, err
	}
	syncConf.defaultSyncConfig()
	return &syncConf, nil
}
