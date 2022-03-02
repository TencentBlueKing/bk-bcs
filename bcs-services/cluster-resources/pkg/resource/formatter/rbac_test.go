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

package formatter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var lightSAManifest = map[string]interface{}{
	"apiVersion": "v1",
	"kind":       "ServiceAccount",
	"metadata": map[string]interface{}{
		"creationTimestamp": "2022-01-01T10:00:00Z",
		"name":              "default",
		"namespace":         "default",
	},
	"secrets": []interface{}{
		map[string]interface{}{
			"name": "default-token-abc",
		},
	},
}

func TestFormatSA(t *testing.T) {
	ret := FormatSA(lightSAManifest)
	assert.Equal(t, "2022-01-01 10:00:00", ret["createTime"])
	assert.Equal(t, 1, ret["secrets"])
}
