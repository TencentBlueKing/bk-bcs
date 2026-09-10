/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"
)

func TestMergeManifest(t *testing.T) {
	old := &unstructured.Unstructured{Object: map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "Service",
		"metadata": map[string]interface{}{
			"name":            "svc-demo",
			"namespace":       "default",
			"uid":             "uid-123",
			"resourceVersion": "100",
			"annotations": map[string]interface{}{
				"keep-key": "keep",
			},
		},
		"spec": map[string]interface{}{
			"type":      "ClusterIP",
			"clusterIP": "10.0.0.8",
			"ports": []interface{}{
				map[string]interface{}{"port": int64(80), "name": "http"},
			},
		},
		"status": map[string]interface{}{
			"loadBalancer": map[string]interface{}{},
		},
	}}
	manifest := map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "Service",
		"metadata": map[string]interface{}{
			"name":      "svc-demo",
			"namespace": "default",
			"annotations": map[string]interface{}{
				"new-key": "overlay",
			},
		},
		"spec": map[string]interface{}{
			"type": "NodePort",
			"ports": []interface{}{
				map[string]interface{}{"port": int64(8080), "name": "http"},
			},
		},
	}

	merged, err := mergeManifest(old, manifest)
	assert.NoError(t, err)
	assert.Equal(t, "100", merged.GetResourceVersion())
	assert.Equal(t, "uid-123", string(merged.GetUID()))
	assert.Equal(t, "keep", mapx.GetStr(merged.Object, "metadata.annotations.keep-key"))
	assert.Equal(t, "overlay", mapx.GetStr(merged.Object, "metadata.annotations.new-key"))
	assert.Equal(t, "NodePort", mapx.GetStr(merged.Object, "spec.type"))
	assert.Equal(t, "10.0.0.8", mapx.GetStr(merged.Object, "spec.clusterIP"))
	assert.NotEmpty(t, mapx.GetMap(merged.Object, "status"))
	ports := mapx.GetList(merged.Object, "spec.ports")
	assert.Len(t, ports, 1)
	assert.Equal(t, int64(8080), mapx.GetInt64(ports[0].(map[string]interface{}), "port"))
}
