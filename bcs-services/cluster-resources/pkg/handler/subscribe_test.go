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

package handler

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/envs"
	res "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource"
	clusterRes "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/proto/cluster-resources"
)

func TestValidateSubscribeParams(t *testing.T) {
	// 在集群中初始化 CRD
	err := getOrCreateCRD()
	assert.Nil(t, err)

	req := clusterRes.SubscribeReq{
		ProjectID:       envs.TestProjectID,
		ClusterID:       envs.TestClusterID,
		ResourceVersion: "0",
	}
	// 检查命名空间域原生资源
	req.Kind = res.Deploy
	// 需要指定命名空间
	err = validateSubscribeParams(&req)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "Namespace")

	req.Namespace = envs.TestNamespace
	assert.Nil(t, validateSubscribeParams(&req))

	// 检查集群域原生资源
	req.Kind = res.PV
	assert.Nil(t, validateSubscribeParams(&req))

	// 检查命名空间域自定义资源
	req.Kind = "CronTab"
	// 没有指定 CRDName，ApiVersion
	err = validateSubscribeParams(&req)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "ApiVersion & CRDName")

	req.CRDName = "xxx.stable.example.com"
	req.ApiVersion = "stable.example.com/v1"
	// crd 在集群中不存在
	err = validateSubscribeParams(&req)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "not found")

	req.CRDName = "crontabs.stable.example.com"
	req.ApiVersion = "stable.example.com/v1"
	assert.Nil(t, validateSubscribeParams(&req))

	req.Kind = "ACObjKind"
	// kind 与 crd 中定义不一致
	err = validateSubscribeParams(&req)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "Kind")

	req.Kind = "CronTab"
	req.Namespace = ""
	// 命名空间域自定义资源需要指定命名空间
	err = validateSubscribeParams(&req)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "Namespace")
}

// Subscribe Handler 单元测试
func TestSubscribeHandler(t *testing.T) {
	crh := NewClusterResourcesHandler()
	req := clusterRes.SubscribeReq{
		ProjectID:       envs.TestProjectID,
		ClusterID:       envs.TestClusterID,
		ResourceVersion: "0",
		Kind:            "Deployment",
		Namespace:       envs.TestNamespace,
	}

	err := crh.Subscribe(context.TODO(), &req, &mockSubscribeStream{})
	// err != nil because force break websocket loop
	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), "force break websocket loop")
}
