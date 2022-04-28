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
	"strings"
	"time"

	"github.com/prometheus/prometheus/pkg/labels"
)

// CredentialType 凭证类型
type CredentialType string

// Credential 支持的类型, 配置 Unmarshal 会通过, 但是不会有匹配
const (
	CredentialBasicAuth   CredentialType = "basic_auth"   // Basic验证
	CredentialBearerToken CredentialType = "bearer_token" // Token验证
	CredentialAppCode     CredentialType = "app_code"     // 蓝鲸App
	CredentialManager     CredentialType = "manager"      // 管理员
)

// ScopeType 类型, 配置 Unmarshal 会通过, 但是不会有匹配
type ScopeType string

const (
	ScopeProjectId   ScopeType = "project_id"   // 项目Id
	ScopeProjectCode ScopeType = "project_code" // 项目Code
	ScopeClusterId   ScopeType = "cluster_id"   // 集群Id
)

// Scope 权限控制，格式如cluster_id: "RE_BCS-K8S-40000", 多个取且关系
type Scope map[ScopeType]string

// LabelMatchers :
type LabelMatchers []*labels.Matcher

// Matchs : 多个是且的关系, 注意, 只匹配单个
func (m *LabelMatchers) Matches(name ScopeType, value string) bool {
	for _, matcher := range *m {
		if matcher.Name != string(name) {
			return false
		}

		if !matcher.Matches(value) {
			return false
		}
	}
	return true
}

// Credential 鉴权
type Credential struct {
	ProjectId      string              `yaml:"project_id"`
	CredentialType CredentialType      `yaml:"credential_type"`
	Credential     string              `yaml:"credential"`
	CredentialList []string            `yaml:"credential_list"`
	Scopes         []Scope             `yaml:"scopes"` // 多个取或关系
	ExpireTime     time.Time           `yaml:"expire_time"`
	Enabled        bool                `yaml:"enabled"`
	Operator       string              `yaml:"operator"`
	Comment        string              `yaml:"comment"`
	credentialKeys map[string]struct{} `yaml:"-"`
	scopeMatcher   []*LabelMatchers    `yaml:"-"`
}

func (c *Credential) InitCred() error {
	if len(c.Scopes) == 0 {
		return errors.New("scopes is required")
	}

	// 初始化凭证, 可为多个
	c.credentialKeys = make(map[string]struct{})
	// 不能为空值
	if c.Credential != "" {
		c.credentialKeys[c.Credential] = struct{}{}
	}
	for _, v := range c.CredentialList {
		if v == "" {
			continue
		}
		c.credentialKeys[v] = struct{}{}
	}

	c.scopeMatcher = make([]*LabelMatchers, 0, len(c.Scopes))
	for _, scope := range c.Scopes {
		matchers := LabelMatchers{}
		for k, v := range scope {
			m, err := NewScopeMatcher(k, v)
			if err != nil {
				return err
			}
			matchers = append(matchers, m)
		}
		c.scopeMatcher = append(c.scopeMatcher, &matchers)
	}

	return nil
}

func (c *Credential) Matches(credType CredentialType, cred string, scopeType ScopeType, scopeValue string) bool {
	// 类型必须优先匹配
	if c.CredentialType != credType {
		return false
	}

	if !c.Enabled {
		return false
	}
	if !c.ExpireTime.IsZero() && c.ExpireTime.Before(time.Now()) {
		return false
	}

	// 必须在凭证中
	if _, ok := c.credentialKeys[cred]; !ok {
		return false
	}

	// 多个是或的关系
	for _, matcher := range c.scopeMatcher {
		if matcher.Matches(scopeType, scopeValue) {
			return true
		}
	}
	return false
}

// NewScopeMatcher : match 4种类型实现
func NewScopeMatcher(name ScopeType, value string) (*labels.Matcher, error) {
	var (
		typeStr   string
		matchType labels.MatchType
	)

	// 只分割第一个_字符
	values := strings.SplitN(value, "_", 2)
	if len(values) >= 2 {
		typeStr = values[0]
		value = values[1]
	}

	switch typeStr {
	case "NEQ":
		matchType = labels.MatchNotEqual
	case "NRE":
		matchType = labels.MatchNotRegexp
	case "EQ":
		matchType = labels.MatchEqual
	case "RE":
		matchType = labels.MatchRegexp
	default:
		matchType = labels.MatchEqual
	}
	return labels.NewMatcher(matchType, string(name), value)
}
