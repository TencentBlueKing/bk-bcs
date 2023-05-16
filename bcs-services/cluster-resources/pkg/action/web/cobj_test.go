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
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	resCsts "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/constants"
)

var gdeployList4Test = []interface{}{
	// 不允许删除的情况
	map[string]interface{}{
		"metadata": map[string]interface{}{
			"labels": map[string]interface{}{
				resCsts.DeletionProtectLabelKey: resCsts.DeletionProtectPolicyNotAllow,
			},
			"uid": "0001",
		},
		"spec": map[string]interface{}{
			"replicas": int64(1),
		},
	},
	map[string]interface{}{
		"metadata": map[string]interface{}{
			"labels": map[string]interface{}{},
			"uid":    "0002",
		},
	},
	map[string]interface{}{
		"metadata": map[string]interface{}{
			"annotations": map[string]interface{}{
				resCsts.EditModeAnnoKey: resCsts.EditModeForm,
			},
			"uid": "0003",
		},
	},
	map[string]interface{}{
		"metadata": map[string]interface{}{
			"uid": "0004",
		},
	},
	// 级联删除保护的情况
	map[string]interface{}{
		"metadata": map[string]interface{}{
			"labels": map[string]interface{}{
				resCsts.DeletionProtectLabelKey: resCsts.DeletionProtectPolicyCascading,
			},
			"uid": "0005",
		},
		"spec": map[string]interface{}{
			"replicas": int64(1),
		},
	},
	map[string]interface{}{
		"metadata": map[string]interface{}{
			"labels": map[string]interface{}{
				resCsts.DeletionProtectLabelKey: resCsts.DeletionProtectPolicyCascading,
			},
			"annotations": map[string]interface{}{
				resCsts.EditModeAnnoKey: resCsts.EditModeForm,
			},
			"uid": "0006",
		},
		"spec": map[string]interface{}{
			"replicas": int64(1),
		},
	},
	map[string]interface{}{
		"metadata": map[string]interface{}{
			"labels": map[string]interface{}{
				resCsts.DeletionProtectLabelKey: resCsts.DeletionProtectPolicyCascading,
			},
			"uid": "0007",
		},
		"spec": map[string]interface{}{
			"replicas": int64(0),
		},
	},
	// 总是允许删除的情况
	map[string]interface{}{
		"metadata": map[string]interface{}{
			"labels": map[string]interface{}{
				resCsts.DeletionProtectLabelKey: resCsts.DeletionProtectPolicyAlways,
			},
			"uid": "0008",
		},
	},
}

func TestGenCObjListWebAnnoFuncs(t *testing.T) {
	notAllowDeleteYamlTips := fmt.Sprintf(
		"当前实例已添加删除保护功能，若确认要删除，请修改实例标签字段 %s: %s",
		resCsts.DeletionProtectLabelKey, resCsts.DeletionProtectPolicyAlways,
	)
	notAllowDeleteFormTips := "当前实例已添加删除保护功能，若确认要删除，请修改实例配置信息->删除保护策略->总是允许删除"
	cascadingDeleteYamlTips := notAllowDeleteYamlTips + "或确保实例数量为 0"
	cascadingDeleteFormTips := notAllowDeleteFormTips + "或确保实例数量为 0"

	webAnno := NewAnnos(genResListDeleteProtectAnnoFuncs(context.TODO(), gdeployList4Test, resCsts.GDeploy)...)
	objPerm := webAnno.Perms.Items["0001"]
	assert.False(t, objPerm[DeleteBtn].Clickable)
	assert.Equal(t, notAllowDeleteYamlTips, objPerm[DeleteBtn].Tip)

	objPerm = webAnno.Perms.Items["0002"]
	assert.False(t, objPerm[DeleteBtn].Clickable)
	assert.Equal(t, notAllowDeleteYamlTips, objPerm[DeleteBtn].Tip)

	objPerm = webAnno.Perms.Items["0003"]
	assert.False(t, objPerm[DeleteBtn].Clickable)
	assert.Equal(t, notAllowDeleteFormTips, objPerm[DeleteBtn].Tip)

	objPerm = webAnno.Perms.Items["0004"]
	assert.False(t, objPerm[DeleteBtn].Clickable)
	assert.Equal(t, notAllowDeleteYamlTips, objPerm[DeleteBtn].Tip)

	objPerm = webAnno.Perms.Items["0005"]
	assert.False(t, objPerm[DeleteBtn].Clickable)
	assert.Equal(t, cascadingDeleteYamlTips, objPerm[DeleteBtn].Tip)

	objPerm = webAnno.Perms.Items["0006"]
	assert.False(t, objPerm[DeleteBtn].Clickable)
	assert.Equal(t, cascadingDeleteFormTips, objPerm[DeleteBtn].Tip)

	objPerm = webAnno.Perms.Items["0007"]
	assert.True(t, objPerm[DeleteBtn].Clickable)
	assert.Equal(t, "", objPerm[DeleteBtn].Tip)

	objPerm = webAnno.Perms.Items["0008"]
	assert.True(t, objPerm[DeleteBtn].Clickable)
	assert.Equal(t, "", objPerm[DeleteBtn].Tip)
}
