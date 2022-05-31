/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package stringx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestYaml2Json(t *testing.T) {
	y := `
test:
  test: 1
  test1: "a"
`
	j, err := Yaml2Json(y)
	assert.Nil(t, err)
	assert.Equal(t, map[interface{}]interface{}{"test": map[interface{}]interface{}{"test": 1, "test1": "a"}}, j)
}

func TestJson2Yaml(t *testing.T) {
	j := map[interface{}]interface{}{"test": map[interface{}]interface{}{"test": 1, "test1": "a"}}
	y, err := Json2Yaml(j)
	expectedYaml := `test:
  test: 1
  test1: a
`
	assert.Nil(t, err)
	assert.Equal(t, expectedYaml, string(y))
}
