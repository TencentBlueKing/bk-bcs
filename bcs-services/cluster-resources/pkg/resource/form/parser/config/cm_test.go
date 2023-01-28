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

package config

import (
	"testing"

	"github.com/stretchr/testify/assert"

	resCsts "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/constants"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/form/model"
)

var lightCMManifest = map[string]interface{}{
	"apiVersion": "v1",
	"kind":       "ConfigMap",
	"metadata": map[string]interface{}{
		"annotations": map[string]interface{}{
			resCsts.EditModeAnnoKey: "form",
		},
		"name":      "configmap-test",
		"namespace": "default",
	},
	"immutable": true,
	"data": map[string]interface{}{
		"key1": "value1",
	},
}

var exceptedCMData = model.CMData{
	Immutable: true,
	Items: []model.OpaqueData{
		{
			Key:   "key1",
			Value: "value1",
		},
	},
}

func TestParseCMData(t *testing.T) {
	actualCMData := model.CMData{}
	ParseCMData(lightCMManifest, &actualCMData)
	assert.Equal(t, exceptedCMData, actualCMData)
}
