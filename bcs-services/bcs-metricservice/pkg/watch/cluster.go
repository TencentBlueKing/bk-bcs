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
	"context"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	stowatch "github.com/Tencent/bk-bcs/bcs-common/common/storage/watch"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-metricservice/app/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-metricservice/pkg/storage"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-metricservice/pkg/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-metricservice/pkg/zk"
)

// ClusterWatcher cluster watcher
type ClusterWatcher struct {
	pCtx   context.Context
	ctx    context.Context
	cancel context.CancelFunc

	outPutEvent chan *MetricEvent

	clusterID     string
	clusterType   types.ClusterType
	metricEvent   chan *stowatch.Event
	dynamicEvent  chan *stowatch.Event
	endpointEvent chan *stowatch.Event

	metric     map[string]*types.Metric
	metricLock sync.RWMutex

	storage storage.Storage
	zk      zk.Zk
	config  *config.Config
}

// NewClusterWatcher create cluster watcher
func NewClusterWatcher(pCtx context.Context,
	clusterID string, storage storage.Storage, zk zk.Zk, config *config.Config) *ClusterWatcher {
	return &ClusterWatcher{pCtx: pCtx, clusterID: clusterID, storage: storage, zk: zk, config: config}
}

// Start start watcher
func (cw *ClusterWatcher) Start(event chan *MetricEvent) {
	blog.Infof("start watching cluster %s", cw.clusterID)
	cw.ctx, cw.cancel = context.WithCancel(cw.pCtx)
	cw.outPutEvent = event
	cw.metricEvent = make(chan *stowatch.Event, 100)
	cw.dynamicEvent = make(chan *stowatch.Event, 100)
	cw.endpointEvent = make(chan *stowatch.Event, 100)
	cw.metric = make(map[string]*types.Metric)
	cw.metricLock = sync.RWMutex{}
	go cw.metricManager()

	// if is Component Metric, then launch the endpoint manager for watching endpoint changes,
	// otherwise launch the dynamic manager for watching dynamic changes
	switch cw.clusterID {
	case types.BcsComponentsClusterId:
		go cw.endpointManger()
	default:
		go cw.dynamicManager()
	}
}

// Stop stop watcher
func (cw *ClusterWatcher) Stop() {
	if cw.cancel != nil {
		blog.Infof("end watching cluster %s", cw.clusterID)
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
