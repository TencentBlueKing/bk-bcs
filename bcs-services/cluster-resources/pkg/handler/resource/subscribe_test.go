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

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/envs"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/handler"
	log "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/logging"
	res "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource"
	cli "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/client"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/example"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"
	clusterRes "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/proto/cluster-resources"
)

func TestValidateSubscribeParams(t *testing.T) {
	// 在集群中初始化 CRD
	err := handler.GetOrCreateCRD()
	assert.Nil(t, err)

	req := clusterRes.SubscribeReq{
		ProjectID:       envs.TestProjectID,
		ClusterID:       envs.TestClusterID,
		ResourceVersion: "0",
	}
	// 检查命名空间域原生资源
	req.Kind = res.Deploy
	ctx := context.TODO()
	// 需要指定命名空间
	err = validateSubscribeParams(ctx, &req)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "Namespace")

	req.Namespace = envs.TestNamespace
	assert.Nil(t, validateSubscribeParams(ctx, &req))

	// 检查集群域原生资源
	req.Kind = res.PV
	assert.Nil(t, validateSubscribeParams(ctx, &req))

	// 检查命名空间域自定义资源
	req.Kind = "CronTab"
	// 没有指定 CRDName，ApiVersion
	err = validateSubscribeParams(ctx, &req)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "ApiVersion & CRDName")

	req.CRDName = "xxx.stable.example.com"
	req.ApiVersion = "stable.example.com/v1"
	// crd 在集群中不存在
	err = validateSubscribeParams(ctx, &req)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "not found")

	req.CRDName = "crontabs.stable.example.com"
	req.ApiVersion = "stable.example.com/v1"
	assert.Nil(t, validateSubscribeParams(ctx, &req))

	req.Kind = "ACObjKind"
	// kind 与 crd 中定义不一致
	err = validateSubscribeParams(ctx, &req)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "Kind")

	req.Kind = "CronTab"
	req.Namespace = ""
	// 命名空间域自定义资源需要指定命名空间
	err = validateSubscribeParams(ctx, &req)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "Namespace")
}

func TestSubscribe(t *testing.T) {
	h := New()
	req := clusterRes.SubscribeReq{
		ProjectID:       envs.TestProjectID,
		ClusterID:       envs.TestClusterID,
		Kind:            res.Po,
		ResourceVersion: "0",
		Namespace:       envs.TestNamespace,
	}

	ctx := context.TODO()
	log.Info(ctx, "start test subscribe pod's event; loop will never break if event is empty!")
	err := h.Subscribe(ctx, &req, &mockSubscribeStream{})
	// err != nil because force break websocket loop
	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), "force break websocket loop")
}

func TestSubscribeDisabledKind(t *testing.T) {
	h := New()
	req := clusterRes.SubscribeReq{
		ProjectID:       envs.TestProjectID,
		ClusterID:       envs.TestSharedClusterID,
		ResourceVersion: "0",
	}

	for _, kind := range []string{res.PV, res.SC} {
		req.Kind = kind
		err := h.Subscribe(context.TODO(), &req, &mockSubscribeStream{})
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "只有指定的数类资源可以执行订阅功能")
	}
}

func TestSubscribeCMInSharedCluster(t *testing.T) {
	err := handler.GetOrCreateNS(envs.TestSharedClusterNS)
	assert.Nil(t, err)

	h := New()
	req := clusterRes.SubscribeReq{
		ProjectID:       envs.TestProjectID,
		ClusterID:       envs.TestSharedClusterID,
		ResourceVersion: "0",
		Kind:            res.CM,
		Namespace:       envs.TestNamespace,
	}

	err = h.Subscribe(context.TODO(), &req, &mockSubscribeStream{})
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "命名空间不属于指定项目")

	// 在共享集群项目命名空间中创建 configmap 确保存在事件
	cmManifest, _ := example.LoadDemoManifest("config/simple_configmap")
	_ = mapx.SetItems(cmManifest, "metadata.namespace", envs.TestSharedClusterNS)
	clusterConf := res.NewClusterConfig(envs.TestSharedClusterID)
	cmRes, err := res.GetGroupVersionResource(context.TODO(), clusterConf, res.CM, "")
	assert.Nil(t, err)
	_, err = cli.NewResClient(clusterConf, cmRes).Create(context.TODO(), cmManifest, true, metav1.CreateOptions{})
	assert.Nil(t, err)

	// 验证查询到事件后退出
	req.Namespace = envs.TestSharedClusterNS
	err = h.Subscribe(context.TODO(), &req, &mockSubscribeStream{})
	// err != nil because force break websocket loop
	assert.NotNil(t, err)
	assert.Equal(t, err.Error(), "force break websocket loop")
}
