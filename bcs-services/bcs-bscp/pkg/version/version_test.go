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

package version

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCurrentVersion(t *testing.T) {
	VERSION = "v1.2.3-alpha1-dev+build1"
	ver, err := parseVersion(VERSION)

	assert.NoError(t, err, "parse version failed, err: %v", err)
	assert.Equal(t, uint32(1), ver[0], "invalid major version: %d", ver[0])
	assert.Equal(t, uint32(2), ver[1], "invalid minor version: %d", ver[1])
	assert.Equal(t, uint32(3), ver[2], "invalid patch version: %d", ver[2])
}

func TestIncorrectVersion(t *testing.T) {
	tests := []string{
		"1.2.3",
		"v1.2.3+",
		"v1.2-",
		"v1 ",
	}
	for _, version := range tests {
		t.Run(version, func(t *testing.T) {
			_, err := parseVersion(version)
			assert.Error(t, err, "expect parse version %s failed, but not", version)
		})
	}
}
