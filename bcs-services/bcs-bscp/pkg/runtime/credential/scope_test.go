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

package credential

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCredentialScopeValidation(t *testing.T) {
	testCases := []struct {
		scope     string
		expectErr bool
	}{
		{"test-*/*", false},
		{"mysql/*", false},
		{"nginx/*.conf", false},
		{"nginx/*/*.conf", false},
		{"nginx/*abc*/*.conf", false},
		{"mysql/*/*", false},
		{"", true},
		{"invalid_scope", true},
		{"invalid//scope", true},
		{"invalid_scope/", true},
		{"/invalid_scope", true},
	}

	for _, tc := range testCases {
		cs := Scope(tc.scope)
		err := cs.Validate()
		if tc.expectErr {
			assert.Error(t, err, fmt.Sprintf("scope: %s", tc.scope))
		} else {
			assert.NoError(t, err, fmt.Sprintf("scope: %s", tc.scope))
		}
	}
}

func TestCredentialScopeMatchApp(t *testing.T) {
	testCases := []struct {
		scope       string
		appName     string
		expectMatch bool
	}{
		{"*/*", "mysql", true},
		{"mysql/*", "mysql", true},
		{"test-*/*", "test-123", true},
		{"*-test-*/*", "abc-test-123", true},
		{"test-*/*", "prod-123", false},
		{"mysql/*", "nginx", false},
	}

	for _, tc := range testCases {
		cs := Scope(tc.scope)
		match, err := cs.MatchApp(tc.appName)
		if err != nil {
			t.Error(err)
		}
		assert.Equal(t, tc.expectMatch, match, fmt.Sprintf("scope: %s, appName: %s", tc.scope, tc.appName))
	}
}

func TestCredentialScopeMatchConfigItem(t *testing.T) {
	testCases := []struct {
		scope       string
		path        string
		name        string
		expectMatch bool
	}{
		{"mysql/**", "/", "mysql.conf", true},
		{"mysql/**", "", "mysql.conf", true},
		{"mysql/**", "/mysql/test", "mysql.conf", true},
		{"mysql/**", "mysql/test", "mysql.conf", true},
		{"mysql/*/*.conf", "/mysql", "nginx.conf", true},
		{"mysql/**/*.conf", "/mysql/test", "nginx.conf", true},
		{"nginx/*.conf", "nginx", "nginx.ini", false},
		{"mysql/*.conf", "/nginx/test", "nginx.conf", false},
		{"mysql/*/*.conf", "/mysql/test", "nginx.conf", false},
	}

	for _, tc := range testCases {
		cs := Scope(tc.scope)
		match, err := cs.MatchConfigItem(tc.path, tc.name)
		if err != nil {
			t.Error(err)
		}
		assert.Equal(t, tc.expectMatch, match, fmt.Sprintf("scope: %s, path: %s, name: %s", tc.scope, tc.path, tc.name))
	}
}
