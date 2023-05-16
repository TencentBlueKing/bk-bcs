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

var lightPVManifestLocal = map[string]interface{}{
	"apiVersion": "v1",
	"kind":       "PersistentVolume",
	"metadata": map[string]interface{}{
		"name": "pv-test-8io98uwj",
	},
	"spec": map[string]interface{}{
		"capacity": map[string]interface{}{
			"storage": "3000Mi",
		},
		"accessModes": []interface{}{
			"ReadOnlyMany",
			"ReadWriteOnce",
		},
		"local": map[string]interface{}{
			"path": "/data0",
		},
		"storageClassName": "local-path",
	},
}

var lightPVManifestHostPath = map[string]interface{}{
	"apiVersion": "v1",
	"kind":       "PersistentVolume",
	"metadata": map[string]interface{}{
		"name": "pv-test-8io98uwk",
	},
	"spec": map[string]interface{}{
		"capacity": map[string]interface{}{
			"storage": "400Mi",
		},
		"accessModes": []interface{}{
			"ReadOnlyMany",
			"ReadWriteMany",
		},
		"hostPath": map[string]interface{}{
			"path": "/data1",
			"type": "DirectoryOrCreate",
		},
		"storageClassName": "local-path",
	},
}

var lightPVManifestNFS = map[string]interface{}{
	"apiVersion": "v1",
	"kind":       "PersistentVolume",
	"metadata": map[string]interface{}{
		"name": "pv-test-8io98uwl",
	},
	"spec": map[string]interface{}{
		"capacity": map[string]interface{}{
			"storage": "4Gi",
		},
		"accessModes": []interface{}{
			"ReadOnlyMany",
			"ReadWriteMany",
		},
		"nfs": map[string]interface{}{
			"path":     "/data",
			"server":   "127.0.0.1",
			"readOnly": false,
		},
		"storageClassName": "local-path",
	},
}

var exceptedPVSpecLocal = model.PVSpec{
	Type:        resCsts.PVTypeLocalVolume,
	SCName:      "local-path",
	StorageSize: 3,
	AccessModes: []string{"ReadOnlyMany", "ReadWriteOnce"},
	LocalPath:   "/data0",
}

var exceptedPVSpecHostPath = model.PVSpec{
	Type:         resCsts.PVTypeHostPath,
	SCName:       "local-path",
	StorageSize:  1,
	AccessModes:  []string{"ReadOnlyMany", "ReadWriteMany"},
	HostPath:     "/data1",
	HostPathType: "DirectoryOrCreate",
}

var exceptedPVSpecNFS = model.PVSpec{
	Type:        resCsts.PVTypeNFS,
	SCName:      "local-path",
	StorageSize: 4,
	AccessModes: []string{"ReadOnlyMany", "ReadWriteMany"},
	NFSPath:     "/data",
	NFSServer:   "127.0.0.1",
	NFSReadOnly: false,
}

func TestParsePVSpec(t *testing.T) {
	actualPVSpec := model.PVSpec{}
	ParsePVSpec(lightPVManifestLocal, &actualPVSpec)
	assert.Equal(t, exceptedPVSpecLocal, actualPVSpec)

	actualPVSpec = model.PVSpec{}
	ParsePVSpec(lightPVManifestHostPath, &actualPVSpec)
	assert.Equal(t, exceptedPVSpecHostPath, actualPVSpec)

	actualPVSpec = model.PVSpec{}
	ParsePVSpec(lightPVManifestNFS, &actualPVSpec)
	assert.Equal(t, exceptedPVSpecNFS, actualPVSpec)
}
