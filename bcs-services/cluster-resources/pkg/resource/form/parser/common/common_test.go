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

package common

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/form/model"
)

var lightDeployManifest = map[string]interface{}{
	"apiVersion": "apps/v1",
	"kind":       "Deployment",
	"metadata": map[string]interface{}{
		"name":      "deployment-test-12345",
		"namespace": "default",
		"labels": map[string]interface{}{
			"app": "busybox",
		},
		"annotations": map[string]interface{}{
			"testKey": "testVal",
		},
	},
}

func TestParseAPIVersionKind(t *testing.T) {
	apiVersion, kind := ParseAPIVersionKind(lightDeployManifest)
	assert.Equal(t, "apps/v1", apiVersion)
	assert.Equal(t, "Deployment", kind)
}

func TestParseMetadata(t *testing.T) {
	expectedMetadata := model.Metadata{
		Name:      "deployment-test-12345",
		Namespace: "default",
		Labels: []model.Label{
			{Key: "app", Value: "busybox"},
		},
		Annotations: []model.Annotation{
			{Key: "testKey", Value: "testVal"},
		},
	}
	actualMetadata := model.Metadata{}
	ParseMetadata(lightDeployManifest, &actualMetadata)
	assert.Equal(t, expectedMetadata, actualMetadata)
}
