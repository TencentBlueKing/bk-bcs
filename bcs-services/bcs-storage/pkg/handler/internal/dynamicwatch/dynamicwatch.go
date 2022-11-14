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

package dynamicwatch

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/pkg/constants"
	storage "github.com/Tencent/bk-bcs/bcs-services/bcs-storage/pkg/proto"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/lib"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/v1http/dynamicwatch"
)

// HandlerWatchDynamic WatchDynamic 业务方法
func HandlerWatchDynamic(ctx context.Context, req *storage.WatchDynamicRequest) (chan *lib.Event, error) {
	cond := req.Option.Cond.AsMap()
	cond[constants.ClusterIDTag] = req.ClusterId
	watchOption := &lib.StoreWatchOption{
		Cond:      cond,
		SelfOnly:  req.Option.SelfOnly,
		MaxEvents: uint(req.Option.MaxEvents),
		Timeout:   req.Option.Timeout.AsDuration(),
		MustDiff:  req.Option.MustDiff,
	}
	return dynamicwatch.GetStore().Watch(ctx, req.ResourceType, watchOption)
}

// HandlerWatchContainer WatchContainer 业务方法
func HandlerWatchContainer(ctx context.Context, req *storage.WatchContainerRequest) (chan *lib.Event, error) {
	cond := req.Option.Cond.AsMap()
	cond[constants.ClusterIDTag] = req.ClusterId
	watchOption := &lib.StoreWatchOption{
		Cond:      cond,
		SelfOnly:  req.Option.SelfOnly,
		MaxEvents: uint(req.Option.MaxEvents),
		Timeout:   req.Option.Timeout.AsDuration(),
		MustDiff:  req.Option.MustDiff,
	}
	return dynamicwatch.GetStore().Watch(ctx, constants.ContainerInfo, watchOption)
}
