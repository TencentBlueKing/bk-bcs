/*
Tencent is pleased to support the open source community by making Blueking Container Service available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package hpm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestVersionCmpFunc test version compare function
func TestVersionCmpFunc(t *testing.T) {
	assert := assert.New(t)

	assert.Equal(0, VersionCmp("v1.0.0-000000000000", "v1.0.0-000000000000"))
	assert.Equal(0, VersionCmp("v0.1.0-000000000000", "v0.1.0-000000000000"))
	assert.Equal(0, VersionCmp("v0.0.1-000000000000", "v0.0.1-000000000000"))
	assert.Equal(0, VersionCmp("v0.0.0-100000000000", "v0.0.0-100000000000"))

	assert.Equal(1, VersionCmp("v1.0.0-000000000000", "v0.0.0-000000000000"))
	assert.Equal(1, VersionCmp("v0.1.0-000000000000", "v0.0.0-000000000000"))
	assert.Equal(1, VersionCmp("v0.0.1-000000000000", "v0.0.0-000000000000"))
	assert.Equal(1, VersionCmp("v0.0.0-100000000000", "v0.0.0-000000000000"))
	assert.Equal(1, VersionCmp("v0.0.0.1-000000000000", "v0.0.0-000000000000"))

	assert.Equal(-1, VersionCmp("v0.0.0-000000000000", "v1.0.0-000000000000"))
	assert.Equal(-1, VersionCmp("v0.0.0-000000000000", "v0.1.0-000000000000"))
	assert.Equal(-1, VersionCmp("v0.0.0-000000000000", "v0.0.1-000000000000"))
	assert.Equal(-1, VersionCmp("v0.0.0-000000000000", "v0.0.0-100000000000"))

	assert.Equal(1, VersionCmp("v0.0.0.0-000000000000","v0.0.0-00000000000"))
	assert.Equal(-1, VersionCmp("v0.0.0-000000000000","v0.0.0.0-00000000000"))

	assert.Equal(1, VersionCmp("v0.1.0-000000000000","v0.0.0.1-00000000000"))
	assert.Equal(-1, VersionCmp("v0.0.0.1-000000000000","v0.1.0-00000000000"))
}
