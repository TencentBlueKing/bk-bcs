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

package renderer

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/envs"
	res "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/example"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/form/parser/workload"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/stringx"
)

var deployManifest4RenderTest = map[string]interface{}{
	"apiVersion": "apps/v1",
	"kind":       "Deployment",
	"metadata": map[string]interface{}{
		"name":      "deployment-test-" + stringx.Rand(example.RandomSuffixLength, example.SuffixCharset),
		"namespace": envs.TestNamespace,
		"labels": map[string]interface{}{
			"app": "busybox",
		},
	},
	"spec": map[string]interface{}{
		"replicas": int64(2),
		"selector": map[string]interface{}{
			"matchLabels": map[string]interface{}{
				"app": "busybox",
			},
		},
		"template": map[string]interface{}{
			"metadata": map[string]interface{}{
				"labels": map[string]interface{}{
					"app": "busybox",
				},
			},
			"spec": map[string]interface{}{
				"containers": []interface{}{
					map[string]interface{}{
						"name":  "busybox",
						"image": "busybox:latest",
						"ports": []interface{}{
							map[string]interface{}{
								"containerPort": int64(80),
							},
						},
						"command": []interface{}{
							"/bin/sh",
							"-c",
						},
						"args": []interface{}{
							"echo hello",
						},
					},
				},
			},
		},
	},
}

func TestNewManifestRenderer(t *testing.T) {
	formData := workload.ParseDeploy(deployManifest4RenderTest)
	manifest, err := NewManifestRenderer(context.TODO(), formData, envs.TestClusterID, res.Deploy).Render()
	assert.Nil(t, err)

	assert.Equal(t, "busybox", mapx.Get(manifest, "metadata.labels.app", ""))
	assert.Equal(t, 2, mapx.Get(manifest, "spec.replicas", 0))
	assert.Equal(t, "busybox", mapx.Get(manifest, "spec.selector.matchLabels.app", ""))
}

func TestSchemaRenderer(t *testing.T) {
	for kind := range FormRenderSupportedResAPIVersion {
		_, err := NewSchemaRenderer(kind).Render()
		assert.Nil(t, err)
	}
}
