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

var lightPVManifest = map[string]interface{}{
	"metadata": map[string]interface{}{
		"creationTimestamp": "2022-01-01T10:00:00Z",
		"name":              "task-pv",
	},
	"spec": map[string]interface{}{
		"accessModes": []interface{}{
			"ReadWriteOnce",
			"ReadOnlyMany",
			"ReadWriteMany",
		},
		"claimRef": map[string]interface{}{
			"name":      "task-pv-claim",
			"namespace": "default",
		},
	},
}

func TestParseShortAccessModes(t *testing.T) {
	assert.Equal(t, []string{"RWO", "ROX", "RWX"}, parseShortAccessModes(lightPVManifest))
}

func TestFormatPV(t *testing.T) {
	ret := FormatPV(lightPVManifest)
	assert.Equal(t, []string{"RWO", "ROX", "RWX"}, ret["accessModes"])
	assert.Equal(t, "default/task-pv-claim", ret["claim"])
}

var lightPVCManifest = map[string]interface{}{
	"metadata": map[string]interface{}{
		"creationTimestamp": "2022-01-01T10:00:00Z",
		"name":              "task-pv-claim",
		"namespace":         "default",
	},
	"spec": map[string]interface{}{
		"accessModes": []interface{}{
			"ReadWriteOnce",
		},
	},
}

func TestFormatPVC(t *testing.T) {
	assert.Equal(t, []string{"RWO"}, FormatPVC(lightPVCManifest)["accessModes"])
}

var lightSCManifest = map[string]interface{}{
	"metadata": map[string]interface{}{
		"creationTimestamp": "2022-01-01T10:00:00Z",
		"name":              "storage-class-test",
		"namespace":         "default",
		"annotations": map[string]interface{}{
			"storageclass.kubernetes.io/is-default-class": "true",
		},
	},
	"provisioner": "k8s.io/minikube-hostpath",
}

func TestFormatSC(t *testing.T) {
	ret := FormatSC(lightSCManifest)
	assert.Equal(t, true, ret["isDefault"])
}
