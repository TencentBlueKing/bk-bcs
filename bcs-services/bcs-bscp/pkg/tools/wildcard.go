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

package tools

import (
	"fmt"
	"strings"

	"github.com/gobwas/glob"
)

// MatchConfigItem checks if the scope string matches the config item.
func MatchConfigItem(scope, path, name string) (bool, error) {
	scope = strings.Trim(scope, "/")
	path = strings.Trim(path, "/")
	fullPath := strings.Trim(path+"/"+name, "/")
	g, err := glob.Compile(scope, '/')
	if err != nil {
		return false, err
	}
	return g.Match(fullPath), nil
}

// MatchAppConfigItem checks if the scope string matches the app and config item.
func MatchAppConfigItem(scope, app, path, name string) (bool, error) {
	arr := strings.SplitN(scope, "/", 2)
	if len(arr) != 2 {
		return false, fmt.Errorf("invalid scope %s for app %s, it can't be splited into 2 substrings by /", scope, app)
	}
	appPattern := arr[0]
	ciPattern := arr[1]
	g, err := glob.Compile(appPattern)
	if err != nil {
		return false, err
	}
	if !g.Match(app) {
		return false, nil
	}
	return MatchConfigItem(ciPattern, path, name)
}
