/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * 	http://opensource.org/licenses/MIT
 *
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package renderer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToYaml(t *testing.T) {
	assert.Equal(t, "foo: bar\nkey: val", toYaml(map[string]interface{}{"foo": "bar", "key": "val"}))
	assert.Equal(t, "- foo\n- bar", toYaml([]string{"foo", "bar"}))
}

func TestGenDockerConfigJson(t *testing.T) {
	assert.Equal(
		t, "{\"auths\":{\"docker.io\":{\"password\":\"pw4321\",\"username\":\"admin0\"}}}",
		genDockerConfigJSON("docker.io", "admin0", "pw4321"),
	)

	assert.Equal(
		t, "{\"auths\":{\"query.io\":{\"password\":\"pw1234\",\"username\":\"admin1\"}}}",
		genDockerConfigJSON("query.io", "admin1", "pw1234"),
	)
}
