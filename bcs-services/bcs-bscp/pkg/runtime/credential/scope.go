/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "as IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package credential

import (
	"fmt"
	"strings"

	"bscp.io/pkg/tools"
	"github.com/gobwas/glob"
)

type CredentialScope string

// Validate validate a credential scope is valid or not.
func (cs CredentialScope) Validate() error {
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

// MatchApp checks if the credential scope matches the app name.
func (cs CredentialScope) MatchApp(name string) (bool, error) {
	if err := cs.Validate(); err != nil {
		return false, err
	}

	appPattern := strings.Split(string(cs), "/")[0]
	return glob.MustCompile(appPattern).Match(name), nil
}

// MatchConfigItem checks if the credential scope matches the config item.
func (cs CredentialScope) MatchConfigItem(path, name string) (bool, error) {
	if err := cs.Validate(); err != nil {
		return false, err
	}
	configItemPattern := strings.SplitN(string(cs), "/", 2)[1]
	return tools.MatchConfigItem(configItemPattern, path, name), nil
}
