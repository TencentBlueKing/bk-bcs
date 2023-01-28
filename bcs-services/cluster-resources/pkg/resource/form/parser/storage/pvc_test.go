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

package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"

	resCsts "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/constants"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/form/model"
)

var lightPVCManifest = map[string]interface{}{
	"apiVersion": "v1",
	"kind":       "PersistentVolumeClaim",
	"metadata": map[string]interface{}{
		"name": "pvc-test-o8uxj7sm",
	},
	"spec": map[string]interface{}{
		"accessModes": []interface{}{
			"ReadOnlyMany",
			"ReadWriteMany",
		},
		"resources": map[string]interface{}{
			"requests": map[string]interface{}{
				"storage": "5Gi",
			},
		},
		"storageClassName": "local-path",
		"volumeName":       "task-pv-volume",
	},
}

var exceptedPVCSpec = model.PVCSpec{
	ClaimType:   resCsts.PVCTypeUseExistPV,
	PVName:      "task-pv-volume",
	SCName:      "local-path",
	StorageSize: 5,
	AccessModes: []string{"ReadOnlyMany", "ReadWriteMany"},
}

func TestParsePVCSpec(t *testing.T) {
	actualPVCSpec := model.PVCSpec{}
	ParsePVCSpec(lightPVCManifest, &actualPVCSpec)
	assert.Equal(t, exceptedPVCSpec, actualPVCSpec)
}
