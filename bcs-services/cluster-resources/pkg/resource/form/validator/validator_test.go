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

package validator

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	resCsts "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/constants"
)

func TestFormDataValidator(t *testing.T) {
	ctx := context.TODO()

	formData := map[string]interface{}{
		"metadata": map[string]interface{}{
			"labels": []interface{}{
				map[string]interface{}{
					"key":   "testLabelKey1",
					"value": "testLabelValue1",
				},
				map[string]interface{}{
					"key":   "testLabelKey2",
					"value": "testLabelValue2",
				},
			},
			"annotations": []interface{}{
				map[string]interface{}{
					"key":   "testAnnoKey1",
					"value": "testAnnoValue1",
				},
				map[string]interface{}{
					"key":   "testAnnoKey2",
					"value": "testAnnoValue2",
				},
			},
		},
	}
	assert.Nil(t, New(ctx, formData, "apps/v1", resCsts.Deploy).Validate())

	// 资源类型不支持表单化
	err := New(ctx, formData, "v1", resCsts.SA).Validate()
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "不支持表单化")

	// 指定 APIVersion 不支持表单化
	err = New(ctx, formData, "apps/v2", resCsts.Deploy).Validate()
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "请改用 Yaml 模式而非表单化")

	// 标签有重复键
	formData = map[string]interface{}{
		"metadata": map[string]interface{}{
			"labels": []interface{}{
				map[string]interface{}{
					"key":   "testLabelKey1",
					"value": "testLabelValue1",
				},
				map[string]interface{}{
					"key":   "testLabelKey1",
					"value": "testLabelValue2",
				},
			},
		},
	}
	err = New(ctx, formData, "apps/v1", resCsts.Deploy).Validate()
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "标签有重复的键")

	// 注解有重复键
	formData = map[string]interface{}{
		"metadata": map[string]interface{}{
			"annotations": []interface{}{
				map[string]interface{}{
					"key":   "testAnnoKey1",
					"value": "testAnnoValue1",
				},
				map[string]interface{}{
					"key":   "testAnnoKey1",
					"value": "testAnnoValue2",
				},
			},
		},
	}
	err = New(ctx, formData, "apps/v1", resCsts.Deploy).Validate()
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "注解有重复的键")
}
