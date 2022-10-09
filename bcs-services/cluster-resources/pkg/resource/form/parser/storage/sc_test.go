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

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/form/model"
)

var lightSCManifest = map[string]interface{}{
	"apiVersion": "storage.k8s.io/v1",
	"kind":       "StorageClass",
	"metadata": map[string]interface{}{
		"annotations": map[string]interface{}{
			"storageclass.kubernetes.io/is-default-class": "true",
		},
		"labels": map[string]interface{}{
			"addonmanager.kubernetes.io/mode": "EnsureExists",
		},
		"name": "standard",
	},
	"provisioner":       "k8s.io/minikube-hostpath",
	"reclaimPolicy":     "Delete",
	"volumeBindingMode": "Immediate",
	"parameters": map[string]interface{}{
		"type": "io1",
	},
	"mountOptions": []interface{}{
		"ro",
		"soft",
	},
}

var exceptedSCSpec = model.SCSpec{
	SetAsDefault:      true,
	Provisioner:       "k8s.io/minikube-hostpath",
	VolumeBindingMode: "Immediate",
	ReclaimPolicy:     "Delete",
	Params: []model.SCParam{
		{"type", "io1"},
	},
	MountOpts: []string{"ro", "soft"},
}

func TestParseSCSpec(t *testing.T) {
	actualSCSpec := model.SCSpec{}
	ParseSCSpec(lightSCManifest, &actualSCSpec)
	assert.Equal(t, exceptedSCSpec, actualSCSpec)
}
