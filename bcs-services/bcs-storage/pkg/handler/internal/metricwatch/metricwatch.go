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

package metricwatch

import (
	"context"

	storage "github.com/Tencent/bk-bcs/bcs-services/bcs-storage/pkg/proto"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/lib"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/v1http/metricwatch"
)

// HandlerWatch Watch 业务方法
func HandlerWatch(ctx context.Context, req *storage.WatchMetricRequest) (chan *lib.Event, error) {
	store := metricwatch.GetStore()
	watchOption := &lib.StoreWatchOption{
		Cond:      req.Option.Cond.AsMap(),
		SelfOnly:  req.Option.SelfOnly,
		MaxEvents: uint(req.Option.MaxEvents),
		Timeout:   req.Option.Timeout.AsDuration(),
		MustDiff:  req.Option.MustDiff,
	}
	return store.Watch(ctx, req.ClusterId, watchOption)
}
