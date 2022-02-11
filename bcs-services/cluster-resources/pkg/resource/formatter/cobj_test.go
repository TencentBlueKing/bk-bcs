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

var lightCrdManifest1 = map[string]interface{}{
	"spec": map[string]interface{}{
		"group": "foo.example.com",
		"versions": []interface{}{
			map[string]interface{}{
				"name":   "v1",
				"served": false,
			},
		},
	},
}

var lightCrdManifest2 = map[string]interface{}{
	"spec": map[string]interface{}{
		"group": "foo.example.com",
		"versions": []interface{}{
			map[string]interface{}{
				"name":   "v1",
				"served": false,
			},
			map[string]interface{}{
				"name":   "v2",
				"served": true,
			},
		},
	},
}

var lightCrdManifest3 = map[string]interface{}{
	"spec": map[string]interface{}{
		"group":   "foo.example.com",
		"version": "v3",
	},
}

var lightCrdManifest4 = map[string]interface{}{
	"spec": map[string]interface{}{
		"group": "foo.example.com",
	},
}

func TestParseCObjAPIVersion(t *testing.T) {
	assert.Equal(t, "foo.example.com/v1", parseCObjAPIVersion(lightCrdManifest1))

	assert.Equal(t, "foo.example.com/v2", parseCObjAPIVersion(lightCrdManifest2))

	assert.Equal(t, "foo.example.com/v3", parseCObjAPIVersion(lightCrdManifest3))

	assert.Equal(t, "foo.example.com/v1alpha1", parseCObjAPIVersion(lightCrdManifest4))
}

var lightCrdManifest = map[string]interface{}{
	"metadata": map[string]interface{}{
		"name":              "crontabs.stable.example.com",
		"creationTimestamp": "2022-01-01T10:00:00Z",
	},
	"spec": map[string]interface{}{
		"group": "stable.example.com",
		"versions": []interface{}{
			map[string]interface{}{
				"name":   "v1",
				"served": true,
			},
		},
		"scope": "Namespaced",
		"names": map[string]interface{}{
			"kind": "CronTab",
		},
	},
}

func TestFormatCRD(t *testing.T) {
	ret := FormatCRD(lightCrdManifest)
	assert.Equal(t, "crontabs.stable.example.com", ret["name"])
	assert.Equal(t, "Namespaced", ret["scope"])
	assert.Equal(t, "CronTab", ret["kind"])
	assert.Equal(t, "stable.example.com/v1", ret["apiVersion"])
}

var lightCObjManifest = map[string]interface{}{
	"apiVersion": "stable.example.com/v1",
	"kind":       "CronTab",
	"metadata": map[string]interface{}{
		"creationTimestamp": "2022-01-01T10:00:00Z",
	},
	"spec": map[string]interface{}{
		"cronSpec": "* * * * */10",
		"image":    "my-awesome-cron-image",
	},
}

func TestFormatCObj(t *testing.T) {
	assert.Equal(t, "2022-01-01 10:00:00", FormatCRD(lightCObjManifest)["createTime"])
}
