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

// 权限控制
const (
	CredentialBasicAuth   = "basic_auth"
	CredentialBearerToken = "bearer_token"
	CredentialAppCode     = "app_code"
	CredentialUsername    = "username"
)

// Scope 权限控制，格式如cluster_id: "RE_BCS-K8S-40000", 多个取且关系
type Scope map[string]string

// LabelMatchers :
type LabelMatchers []*labels.Matcher

// Matchs : 多个是且的关系
func (m *LabelMatchers) Matches(s string) bool {
	for _, matcher := range *m {
		if !matcher.Matches(s) {
			return false
		}
	}
	return true
}

// Credential 鉴权
type Credential struct {
	ProjectId      string           `yaml:"project_id"`
	CredentialType string           `yaml:"credential_type"`
	Credential     string           `yaml:"credential"`
	Scopes         []Scope          `yaml:"scopes"` // 多个取或关系
	scopeMatcher   []*LabelMatchers `yaml:"-"`
	ExpireTime     time.Time        `yaml:"expire_time"`
	Enabled        bool             `yaml:"enabled"`
	Operator       string           `yaml:"operator"`
	Comment        string           `yaml:"comment"`
}

func (c *Credential) InitMatcher() error {
	if len(c.Scopes) == 0 {
		return errors.New("scopes is required")
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

func (c *Credential) Matches(appCode, projectCode string) bool {
	if !c.Enabled {
		return false
	}
	if !c.ExpireTime.IsZero() && c.ExpireTime.Before(time.Now()) {
		return false
	}

	// 多个是或的关系
	for _, matcher := range c.scopeMatcher {
		if matcher.Matches(projectCode) {
			return true
		}
	}
	return false
}

// NewScopeMatcher : match 4种类型实现
func NewScopeMatcher(name, value string) (*labels.Matcher, error) {
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
	return labels.NewMatcher(matchType, name, value)

}
