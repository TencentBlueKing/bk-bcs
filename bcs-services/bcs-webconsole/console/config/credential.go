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

// 权限控制
const (
	CredentialBasicAuth   = "basic_auth"
	CredentialBearerToken = "bearer_token"
	CredentialAppCode     = "app_code"
	CredentialUsername    = "username"
)

// Scope 权限控制，格式如cluster_id: "RE_BCS-K8S-40000", 多个取且关系
type Scope map[string]string

// Credential 鉴权
type Credential struct {
	ProjectId      string    `yaml:"project_id"`
	CredentialType string    `yaml:"credential_type"`
	Credential     string    `yaml:"credential"`
	Scopes         []Scope   `yaml:"scopes"` // 多个取或关系
	ExpireTime     time.Time `yaml:"expire_time"`
	Enabled        bool      `yaml:"enabled"`
	Operator       string    `yaml:"operator"`
	Comment        string    `yaml:"comment"`
}

func (c *Credential) IsValid(projectId, clusterId string) bool {
	if !c.Enabled {
		return false
	}
	if !c.ExpireTime.IsZero() && c.ExpireTime.Before(time.Now()) {
		return false
	}
	return true
}
