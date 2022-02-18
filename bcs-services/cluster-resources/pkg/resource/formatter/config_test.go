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

var lightCMManifest = map[string]interface{}{
	"apiVersion": "v1",
	"kind":       "ConfigMap",
	"metadata": map[string]interface{}{
		"name":              "configmap-alpha",
		"namespace":         "default",
		"creationTimestamp": "2022-01-01T10:00:00Z",
	},
	"data": map[string]interface{}{
		"special.how":  "very",
		"special.type": "charm",
	},
}

func TestFormatConfigRes(t *testing.T) {
	ret := FormatConfigRes(lightCMManifest)
	assert.Equal(t, 2, len(ret["data"].([]string)))
}
