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

package example

import (
	"testing"

	"github.com/stretchr/testify/assert"

	res "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"
)

func TestLoadResConf(t *testing.T) {
	conf, _ := LoadResConf(res.Deploy)
	assert.Equal(t, "Deployment", conf["kind"])
	assert.Equal(t, "workload", conf["class"])

	for _, kind := range HasDemoManifestResKinds {
		conf, _ = LoadResConf(kind)
		assert.Equal(t, kind, conf["kind"])
	}

	_, err := LoadResConf(res.CRD)
	assert.NotNil(t, err)
}

func TestLoadResRefs(t *testing.T) {
	refs, _ := LoadResRefs(res.Deploy)
	assert.True(t, len(refs) > 0)

	refs, _ = LoadResRefs(res.Secret)
	assert.True(t, len(refs) > 0)

	_, err := LoadResRefs(res.CRD)
	assert.NotNil(t, err)
}

func TestLoadDemoManifest(t *testing.T) {
	manifest, _ := LoadDemoManifest("workload/simple_deployment", "")
	assert.Equal(t, "Deployment", manifest["kind"])

	manifest, _ = LoadDemoManifest("storage/simple_persistent_volume", "")
	assert.Equal(t, "PersistentVolume", manifest["kind"])

	// 指定命名空间不生效的
	manifest, _ = LoadDemoManifest("storage/simple_storage_class", "custom-namespace")
	_, err := mapx.GetItems(manifest, "metadata.namespace")
	assert.NotNil(t, err)

	// 指定命名空间生效的
	manifest, _ = LoadDemoManifest("config/simple_secret", "custom-namespace")
	namespace, _ := mapx.GetItems(manifest, "metadata.namespace")
	assert.Equal(t, "custom-namespace", namespace)

	_, err = LoadDemoManifest("custom_resource/custom_object", "")
	assert.NotNil(t, err)
}
