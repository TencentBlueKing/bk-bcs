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

package web

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

var cmList4Test = []interface{}{
	// 不可变更的的情况
	map[string]interface{}{
		"metadata": map[string]interface{}{
			"uid": "0001",
		},
		"immutable": true,
	},
	// 可变更的情况
	map[string]interface{}{
		"metadata": map[string]interface{}{
			"uid": "0002",
		},
		"immutable": false,
	},
	map[string]interface{}{
		"metadata": map[string]interface{}{
			"uid": "0003",
		},
	},
}

func TestGenResListImmutableAnnoFuncs(t *testing.T) {
	tips := "当前资源已设置为不可变更状态，无法编辑"

	webAnno := NewAnnos(genResListImmutableAnnoFuncs(context.TODO(), cmList4Test)...)
	objPerm := webAnno.Perms.Items["0001"]
	assert.False(t, objPerm[UpdateBtn].Clickable)
	assert.Equal(t, tips, objPerm[UpdateBtn].Tip)

	objPerm = webAnno.Perms.Items["0002"]
	assert.True(t, objPerm[UpdateBtn].Clickable)
	assert.Equal(t, "", objPerm[UpdateBtn].Tip)

	objPerm = webAnno.Perms.Items["0003"]
	assert.True(t, objPerm[UpdateBtn].Clickable)
	assert.Equal(t, "", objPerm[UpdateBtn].Tip)
}
