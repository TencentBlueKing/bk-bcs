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

// BKLogConf bk log config
type BKLogConf struct {
	APIServer        string `yaml:"api_server"`        // openapi 地址
	Entrypoint       string `yaml:"entrypoint"`        // bk-log 页面 host 地址，e.g: https://bklog.example.com
	BKBaseEntrypoint string `yaml:"bkbase_entrypoint"` // bkbase 页面 host 地址，e.g: https://bkbase.example.com
}

// BKBaseConf bk base config
type BKBaseConf struct {
	APIServer        string `yaml:"api_server"`         // openapi 地址
	AuditChannelName string `yaml:"audit_channel_name"` // 审计通道名称
}
