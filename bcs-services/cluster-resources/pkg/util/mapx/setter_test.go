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

package mapx_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"
)

// SetItems 成功的情况
func TestSetItemsSuccessCase(t *testing.T) {
	// depth 1，val type int
	err := mapx.SetItems(deploySpec, "intKey4SetItem", 5)
	assert.Nil(t, err)
	ret, _ := mapx.GetItems(deploySpec, []string{"intKey4SetItem"})
	assert.Equal(t, 5, ret)

	// depth 2, val type string
	err = mapx.SetItems(deploySpec, "strategy.type", "Rolling")
	assert.Nil(t, err)
	ret, _ = mapx.GetItems(deploySpec, []string{"strategy", "type"})
	assert.Equal(t, "Rolling", ret)

	// depth 3, val type string
	err = mapx.SetItems(deploySpec, []string{"template", "spec", "restartPolicy"}, "Never")
	assert.Nil(t, err)
	ret, _ = mapx.GetItems(deploySpec, []string{"template", "spec", "restartPolicy"})
	assert.Equal(t, "Never", ret)

	// key noy exists
	err = mapx.SetItems(deploySpec, []string{"selector", "testKey"}, "testVal")
	assert.Nil(t, err)
	ret, _ = mapx.GetItems(deploySpec, "selector.testKey")
	assert.Equal(t, "testVal", ret)
}

// SetItems 失败的情况
func TestSetItemsFailCase(t *testing.T) {
	// not paths error
	err := mapx.SetItems(deploySpec, []string{}, 1)
	assert.NotNil(t, err)

	// not map[string]interface{} type error
	err = mapx.SetItems(deploySpec, []string{"replicas", "testKey"}, 1)
	assert.NotNil(t, err)

	// key not exist
	err = mapx.SetItems(deploySpec, []string{"templateKey", "spec"}, 1)
	assert.NotNil(t, err)

	err = mapx.SetItems(deploySpec, "templateKey.spec", 123)
	assert.NotNil(t, err)

	// paths type error
	err = mapx.SetItems(deploySpec, []int{123, 456}, 1)
	assert.NotNil(t, err)

	err = mapx.SetItems(deploySpec, 123, 1)
	assert.NotNil(t, err)
}
