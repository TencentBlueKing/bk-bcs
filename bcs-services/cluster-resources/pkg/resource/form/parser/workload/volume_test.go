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

package workload

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/form/model"
)

var lightManifest4VolumeTest = map[string]interface{}{
	"apiVersion": "apps/v1",
	"kind":       "Deployment",
	"spec": map[string]interface{}{
		"template": map[string]interface{}{
			"spec": map[string]interface{}{
				"volumes": []interface{}{
					map[string]interface{}{
						"name": "nfs",
						"nfs": map[string]interface{}{
							"path":   "/data",
							"server": "1.1.1.1",
						},
					},
					map[string]interface{}{
						"name": "pvc",
						"persistentVolumeClaim": map[string]interface{}{
							"claimName": "pvc-123456",
						},
					},
					map[string]interface{}{
						"name":     "emptydir",
						"emptyDir": map[string]interface{}{},
					},
					map[string]interface{}{
						"name": "hostpath",
						"hostPath": map[string]interface{}{
							"path": "/tmp/hostP.log",
							"type": "FileOrCreate",
						},
					},
					map[string]interface{}{
						"name": "cm",
						"configMap": map[string]interface{}{
							"defaultMode": int64(420),
							"items": []interface{}{
								map[string]interface{}{
									"key":  "ca.crt",
									"path": "ca.crt",
								},
							},
							"name": "kube-root-ca.crt",
						},
					},
					map[string]interface{}{
						"name": "secret",
						"secret": map[string]interface{}{
							"defaultMode": int64(420),
							"secretName":  "ssh-auth-test",
						},
					},
				},
			},
		},
	},
}

var exceptedVolume = model.WorkloadVolume{
	PVC: []model.PVCVolume{
		{
			Name:     "pvc",
			PVCName:  "pvc-123456",
			ReadOnly: false,
		},
	},
	HostPath: []model.HostPathVolume{
		{
			Name: "hostpath",
			Path: "/tmp/hostP.log",
			Type: "FileOrCreate",
		},
	},
	ConfigMap: []model.CMVolume{
		{
			Name:        "cm",
			DefaultMode: int64(420),
			CMName:      "kube-root-ca.crt",
			Items: []model.KeyToPath{
				{
					Key:  "ca.crt",
					Path: "ca.crt",
				},
			},
		},
	},
	Secret: []model.SecretVolume{
		{
			Name:        "secret",
			DefaultMode: int64(420),
			SecretName:  "ssh-auth-test",
			Items:       []model.KeyToPath{},
		},
	},
	EmptyDir: []model.EmptyDirVolume{
		{
			Name: "emptydir",
		},
	},
	NFS: []model.NFSVolume{
		{
			Name:     "nfs",
			Path:     "/data",
			Server:   "1.1.1.1",
			ReadOnly: false,
		},
	},
}

func TestParseWorkloadVolume(t *testing.T) {
	actualVolume := model.WorkloadVolume{}
	ParseWorkloadVolume(lightManifest4VolumeTest, &actualVolume)
	assert.Equal(t, exceptedVolume, actualVolume)
}
