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

package resource

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const testClusterID = "BCS-K8S-T99999"

func TestGenCacheKey(t *testing.T) {
	k := genCacheKey(testClusterID, "v1")
	assert.Equal(t, "BCS-K8S-T99999:v1:serverresources", k.Key())

	k = genCacheKey(testClusterID, "networking.k8s.io/v1")
	assert.Equal(t, "BCS-K8S-T99999:networking.k8s.io/v1:serverresources", k.Key())

	k = genCacheKey(testClusterID, "")
	assert.Equal(t, "BCS-K8S-T99999:all:servergroups", k.Key())
}

func TestFilterResByKind(t *testing.T) {
	allRes := []*metav1.APIResourceList{{
		GroupVersion: "v1",
		APIResources: []metav1.APIResource{{Kind: Po}},
	}, {
		GroupVersion: "apps/v1",
		APIResources: []metav1.APIResource{{Kind: Deploy}},
	}}

	// groupVersion 特殊情况（只有 version，没有 group）
	res, err := filterResByKind(Po, testClusterID, "", allRes)
	assert.Nil(t, err)
	assert.Equal(t, "", res.Group)
	assert.Equal(t, "v1", res.Version)

	// 普通情况
	res, err = filterResByKind(Deploy, testClusterID, "", allRes)
	assert.Nil(t, err)
	assert.Equal(t, "apps", res.Group)
	assert.Equal(t, "v1", res.Version)

	// 找不到的情况
	_, err = filterResByKind("NotExistsKind", testClusterID, "", allRes)
	assert.NotNil(t, err)
}

// helpers func
func getResByDiscovery(t *testing.T, rcc *RedisCacheClient) {
	t.Helper()

	// preferred deployment
	res, err := rcc.getPreferredResource(Deploy)
	assert.Nil(t, err)
	assert.Equal(t, "deployments", res.Resource)

	// not exists kind
	_, err = rcc.getPreferredResource("NotExistsKind")
	assert.NotNil(t, err)

	// v1 pod
	res, err = rcc.getResWithGroupVersion(Po, "v1")
	assert.Nil(t, err)
	assert.Equal(t, "", res.Group)
	assert.Equal(t, "v1", res.Version)

	// v3 deployment (not exists)
	_, err = rcc.getResWithGroupVersion(Deploy, "v3")
	assert.NotNil(t, err)
}

func TestRedisCacheClient(t *testing.T) {
	rcc, _ := NewRedisCacheClient4Conf(context.TODO(), NewClusterConfig(testClusterID))

	// 检查确保 Redis 中对应键不存在
	srV1Key := genCacheKey(testClusterID, "v1")
	srNetV1Key := genCacheKey(testClusterID, "networking.k8s.io/v1")
	sgKey := genCacheKey(testClusterID, "")
	assert.False(t, rcc.rdsCache.Exists(srV1Key))
	assert.False(t, rcc.rdsCache.Exists(srNetV1Key))
	assert.False(t, rcc.rdsCache.Exists(sgKey))

	// 第一次取，会写 Redis 缓存
	getResByDiscovery(t, rcc)

	assert.True(t, rcc.rdsCache.Exists(srV1Key))
	assert.True(t, rcc.rdsCache.Exists(srNetV1Key))
	assert.True(t, rcc.rdsCache.Exists(sgKey))

	// 强制缓存失效
	assert.True(t, rcc.Fresh())
	rcc.Invalidate()
	assert.False(t, rcc.Fresh())

	// 第二次取，会再写 Redis 缓存
	getResByDiscovery(t, rcc)
	assert.True(t, rcc.Fresh())

	// 清理缓存内容
	assert.Nil(t, rcc.ClearCache())

	// rcc 其他方法测试
	_ = rcc.RESTClient()

	_, err := rcc.ServerResources()
	assert.Nil(t, err)

	_, _, err = rcc.ServerGroupsAndResources()
	assert.Nil(t, err)

	_, err = rcc.ServerPreferredNamespacedResources()
	assert.Nil(t, err)

	_, err = rcc.ServerVersion()
	assert.Nil(t, err)

	_, err = rcc.OpenAPISchema()
	assert.Nil(t, err)
}

func TestGetGroupVersionResource(t *testing.T) {
	clusterConf := NewClusterConfig(testClusterID)

	ret, err := GetGroupVersionResource(context.TODO(), clusterConf, Deploy, "")
	assert.Nil(t, err)
	assert.Equal(t, ret.Resource, "deployments")

	ret, err = GetGroupVersionResource(context.TODO(), clusterConf, Po, "v1")
	assert.Nil(t, err)
	assert.Equal(t, ret.Resource, "pods")

	_, err = GetGroupVersionResource(context.TODO(), clusterConf, "NotExistsKind", "")
	assert.NotNil(t, err)

	_, err = GetGroupVersionResource(context.TODO(), clusterConf, "NotExistsKind", "v1")
	assert.NotNil(t, err)
}
