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
 *
 */

package watchk8smesos

import (
	"context"
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/pkg/constants"
	storage "github.com/Tencent/bk-bcs/bcs-services/bcs-storage/pkg/proto"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/v1http/watchk8smesos"
	sto "github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/types"
)

type general struct {
	env      string
	data     map[string]interface{}
	ctx      context.Context
	resource *storage.WatchResource
}

func (g *general) get() (interface{}, error) {
	resType := types.ObjectType(g.resource.ResourceType)
	key := types.ObjectKey{
		ClusterID: g.resource.ClusterId,
		Namespace: g.resource.Namespace,
		Name:      g.resource.ResourceName,
	}

	opt := &sto.GetOptions{Env: g.env}

	return watchk8smesos.GetData(g.ctx, resType, key, opt)
}

func (g *general) put() error {
	resType := types.ObjectType(g.resource.ResourceName)
	data := g.data
	data[constants.UpdateTimeTag] = time.Now()

	newObj := &types.RawObject{
		Meta: types.Meta{
			Type:      resType,
			ClusterID: g.resource.ClusterId,
			Namespace: g.resource.Namespace,
			Name:      g.resource.ResourceName,
		},
		Data: data,
	}

	opt := &sto.UpdateOptions{
		Env:             g.env,
		CreateNotExists: true,
	}

	return watchk8smesos.PutData(g.ctx, newObj, opt)
}

func (g *general) remove() error {
	resType := g.resource.ResourceType
	clusterID := g.resource.ClusterId
	ns := g.resource.Namespace
	name := g.resource.ResourceName

	newObj := &types.RawObject{
		Meta: types.Meta{
			Type:      types.ObjectType(resType),
			ClusterID: clusterID,
			Namespace: ns,
			Name:      name,
		},
	}

	opt := &sto.DeleteOptions{
		Env:            g.env,
		IgnoreNotFound: true,
	}

	return watchk8smesos.RemoveDta(g.ctx, newObj, opt)
}

func (g *general) list() ([]string, error) {
	resType := types.ObjectType(g.resource.ResourceType)
	clusterID := g.resource.ClusterId
	ns := g.resource.Namespace

	opt := &sto.ListOptions{
		Env:       g.env,
		Cluster:   clusterID,
		Namespace: ns,
	}

	return watchk8smesos.GetList(g.ctx, resType, opt)
}

// HandlerK8SGetWatchResource 查询watch资源
func HandlerK8SGetWatchResource(ctx context.Context, req *storage.K8SGetWatchResourceRequest) (interface{}, error) {
	r := &storage.WatchResource{
		ClusterId:    req.ClusterId,
		Namespace:    req.Namespace,
		ResourceType: req.ResourceType,
		ResourceName: req.ResourceName,
	}
	g := &general{
		ctx:      ctx,
		resource: r,
		env:      watchk8smesos.K8sEnv,
	}

	return g.get()
}

// HandlerK8SPutWatchResource 修改watch资源
func HandlerK8SPutWatchResource(ctx context.Context, req *storage.K8SPutWatchResourceRequest) error {
	r := &storage.WatchResource{
		ClusterId:    req.ClusterId,
		Namespace:    req.Namespace,
		ResourceType: req.ResourceType,
		ResourceName: req.ResourceName,
	}
	g := &general{
		ctx:      ctx,
		data:     req.Data.AsMap(),
		resource: r,
		env:      watchk8smesos.K8sEnv,
	}

	return g.put()
}

// HandlerK8SDeleteWatchResource 删除watch资源
func HandlerK8SDeleteWatchResource(ctx context.Context, req *storage.K8SDeleteWatchResourceRequest) error {
	r := &storage.WatchResource{
		ClusterId:    req.ClusterId,
		Namespace:    req.Namespace,
		ResourceType: req.ResourceType,
		ResourceName: req.ResourceName,
	}
	g := &general{
		ctx:      ctx,
		resource: r,
		env:      watchk8smesos.K8sEnv,
	}

	return g.remove()
}

// HandlerK8SListWatchResource 批量查询watch资源
func HandlerK8SListWatchResource(ctx context.Context, req *storage.K8SListWatchResourceRequest) ([]string, error) {
	r := &storage.WatchResource{
		ClusterId:    req.ClusterId,
		Namespace:    req.Namespace,
		ResourceType: req.ResourceType,
	}
	g := &general{
		ctx:      ctx,
		resource: r,
		env:      watchk8smesos.K8sEnv,
	}

	return g.list()
}
