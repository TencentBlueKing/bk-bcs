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

package mapx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var detail = map[interface{}]interface{}{
	"apiVersion": "v1",
	"kind":       "Deployment",
	"metadata": map[interface{}]interface{}{
		"test": "test",
	},
	"spec": map[interface{}]interface{}{
		"template": map[interface{}]interface{}{
			"metadata": map[interface{}]interface{}{
				"labels": map[interface{}]interface{}{
					"test": "test",
				},
			},
		},
	},
}

func TestGetItemsSuccessCase(t *testing.T) {
	// 1层
	val, err := GetItems(detail, []string{"kind"})
	assert.Nil(t, err)
	assert.Equal(t, "Deployment", val)

	// 2层
	val, err = GetItems(detail, []string{"metadata", "test"})
	assert.Nil(t, err)
	assert.Equal(t, "test", val)

	// 多层
	val, err = GetItems(detail, []string{"spec", "template", "metadata", "labels", "test"})
	assert.Nil(t, err)
	assert.Equal(t, "test", val)
}

func TestGetItemsErrorCase(t *testing.T) {
	_, err := GetItems(detail, []string{"kind1"})
	assert.NotNil(t, err)

	_, err = GetItems(detail, []string{})
	assert.NotNil(t, err)
}

func TestSetItemsSuccessCase(t *testing.T) {
	// 1层
	err := SetItems(detail, []string{"apiVersion"}, "v2")
	assert.Nil(t, err)
	realVal, _ := GetItems(detail, []string{"apiVersion"})
	assert.Equal(t, "v2", realVal)

	// 2层
	err = SetItems(detail, []string{"metadata", "test"}, "test1")
	assert.Nil(t, err)
	realVal, _ = GetItems(detail, []string{"metadata", "test"})
	assert.Equal(t, "test1", realVal)

	// 多层
	err = SetItems(detail, []string{"spec", "template", "metadata", "labels", "test"}, "test1")
	assert.Nil(t, err)
	realVal, _ = GetItems(detail, []string{"spec", "template", "metadata", "labels", "test"})
	assert.Equal(t, "test1", realVal)

	// key 不存在
	err = SetItems(detail, []string{"spec", "template", "metadata", "labels", "test1"}, "test1")
	assert.Nil(t, err)
	realVal, _ = GetItems(detail, []string{"spec", "template", "metadata", "labels", "test1"})
	assert.Equal(t, "test1", realVal)
}

func TestSetItemsErrorCase(t *testing.T) {
	// 路径类型不正确
	err := SetItems(detail, []int{12}, "test")
	assert.NotNil(t, err)

	// 路径为空
	err = SetItems(detail, []string{}, "test")
	assert.NotNil(t, err)

	// 路径不正确
	err = SetItems(detail, []string{"spec1", "template", "test"}, "test")
	assert.NotNil(t, err)
}
