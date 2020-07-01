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

package watch

import (
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-metricservice/app/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-metricservice/pkg/storage"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-metricservice/pkg/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-metricservice/pkg/zk"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/operator"

	"context"
	"sync"
)

type ClusterWatcher struct {
	pCtx   context.Context
	ctx    context.Context
	cancel context.CancelFunc

	outPutEvent chan *MetricEvent

	clusterId     string
	clusterType   types.ClusterType
	metricEvent   chan *operator.Event
	dynamicEvent  chan *operator.Event
	endpointEvent chan *operator.Event

	metric     map[string]*types.Metric
	metricLock sync.RWMutex

	storage storage.Storage
	zk      zk.Zk
	config  *config.Config
}

func NewClusterWatcher(pCtx context.Context, clusterId string, storage storage.Storage, zk zk.Zk, config *config.Config) *ClusterWatcher {
	return &ClusterWatcher{pCtx: pCtx, clusterId: clusterId, storage: storage, zk: zk, config: config}
}

func (cw *ClusterWatcher) Start(event chan *MetricEvent) {
	blog.Infof("start watching cluster %s", cw.clusterId)
	cw.ctx, cw.cancel = context.WithCancel(cw.pCtx)
	cw.outPutEvent = event
	cw.metricEvent = make(chan *operator.Event, 100)
	cw.dynamicEvent = make(chan *operator.Event, 100)
	cw.endpointEvent = make(chan *operator.Event, 100)
	cw.metric = make(map[string]*types.Metric)
	cw.metricLock = sync.RWMutex{}
	go cw.metricManager()

	// if is Component Metric, then launch the endpoint manager for watching endpoint changes,
	// otherwise launch the dynamic manager for watching dynamic changes
	switch cw.clusterId {
	case types.BcsComponentsClusterId:
		go cw.endpointManger()
	default:
		go cw.dynamicManager()
	}
}

func (cw *ClusterWatcher) Stop() {
	if cw.cancel != nil {
		blog.Infof("end watching cluster %s", cw.clusterId)
		cw.cancel()
	}
}

func (cw *ClusterWatcher) getClusterType() types.ClusterType {
	if cw.clusterType != types.ClusterUnknown {
		return cw.clusterType
	}

	cw.metricLock.RLock()
	defer cw.metricLock.RUnlock()

	var t string
	for _, metric := range cw.metric {
		t = metric.ClusterType
		break
	}
	cw.clusterType = types.GetClusterType(t)
	return cw.clusterType
}
