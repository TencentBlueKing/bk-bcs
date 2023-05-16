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
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/ctxkey"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/i18n"
	resCsts "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/constants"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"
)

func TestLoadResConf(t *testing.T) {
	ctx := context.TODO()
	conf, _ := LoadResConf(ctx, resCsts.Deploy)
	assert.Equal(t, "Deployment", conf["kind"])
	assert.Equal(t, "workload", conf["class"])

	for _, kind := range HasDemoManifestResKinds {
		conf, _ = LoadResConf(ctx, kind)
		assert.Equal(t, kind, conf["kind"])
	}

	ctx = context.WithValue(ctx, ctxkey.LangKey, i18n.EN)
	for _, kind := range HasDemoManifestResKinds {
		conf, _ = LoadResConf(ctx, kind)
		assert.Equal(t, kind, conf["kind"])
	}

	_, err := LoadResConf(ctx, resCsts.CRD)
	assert.NotNil(t, err)
}

func TestLoadResRefs(t *testing.T) {
	ctx := context.TODO()
	refs, _ := LoadResRefs(ctx, resCsts.Deploy)
	assert.True(t, len(refs) > 0)

	refs, _ = LoadResRefs(ctx, resCsts.Secret)
	assert.True(t, len(refs) > 0)

	ctx = context.WithValue(ctx, ctxkey.LangKey, i18n.EN)
	refs, _ = LoadResRefs(ctx, resCsts.Secret)
	assert.True(t, len(refs) > 0)

	_, err := LoadResRefs(ctx, resCsts.CRD)
	assert.NotNil(t, err)
}

func TestLoadDemoManifest(t *testing.T) {
	ctx := context.TODO()

	manifest, _ := LoadDemoManifest(ctx, "workload/simple_deployment", "", "", resCsts.Deploy)
	assert.Equal(t, "Deployment", manifest["kind"])

	manifest, _ = LoadDemoManifest(ctx, "storage/simple_persistent_volume", "", "", resCsts.PV)
	assert.Equal(t, "PersistentVolume", manifest["kind"])

	// 指定命名空间不生效的
	manifest, _ = LoadDemoManifest(ctx, "storage/simple_storage_class", "", "custom-namespace", resCsts.SC)
	_, err := mapx.GetItems(manifest, "metadata.namespace")
	assert.NotNil(t, err)

	// 指定命名空间生效的
	manifest, _ = LoadDemoManifest(ctx, "config/simple_secret", "", "custom-namespace", resCsts.Secret)
	namespace, _ := mapx.GetItems(manifest, "metadata.namespace")
	assert.Equal(t, "custom-namespace", namespace)

	_, err = LoadDemoManifest(ctx, "custom_resource/custom_object", "", "", resCsts.CObj)
	assert.NotNil(t, err)
}
