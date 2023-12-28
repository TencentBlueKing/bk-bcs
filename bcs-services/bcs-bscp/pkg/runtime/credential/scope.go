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

// Package credential provides credential scope related operations.
package credential

import (
	"fmt"
	"strings"

	"github.com/gobwas/glob"

	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/tools"
)

// Scope defines the credential scope expression.
type Scope string

// Validate validate a credential scope is valid or not.
func (cs Scope) Validate() error {
	strs := strings.Split(string(cs), "/")
	if len(strs) < 2 {
		return fmt.Errorf("invalid credential scope %s", string(cs))
	}
	for _, str := range strs {
		if len(str) == 0 {
			return fmt.Errorf("invalid credential scope %s", string(cs))
		}
	}
	return nil
}

// Split 拆分成 app / scope 格式
func (cs Scope) Split() (app string, scope string, err error) {
	index := strings.Index(string(cs), "/")
	if index == -1 {
		return "", "", fmt.Errorf("invalid credential scope %s", cs)
	}

	app = string(cs[:index])
	scope = string(cs[index:])
	return
}

// MatchApp checks if the credential scope matches the app name.
func (cs Scope) MatchApp(name string) (bool, error) {
	if err := cs.Validate(); err != nil {
		return false, err
	}

	appPattern := strings.Split(string(cs), "/")[0]
	return glob.MustCompile(appPattern).Match(name), nil
}

// MatchConfigItem checks if the credential scope matches the config item.
func (cs Scope) MatchConfigItem(path, name string) (bool, error) {
	if err := cs.Validate(); err != nil {
		return false, err
	}
	configItemPattern := strings.SplitN(string(cs), "/", 2)[1]
	return tools.MatchConfigItem(configItemPattern, path, name), nil
}

// New 通过 app scope 组装
func New(app string, scope string) (Scope, error) {
	if len(app) == 0 {
		return "", fmt.Errorf("app is required")
	}

	if len(scope) == 0 {
		return "", fmt.Errorf("scope is required")
	}

	return Scope(app + scope), nil
}
