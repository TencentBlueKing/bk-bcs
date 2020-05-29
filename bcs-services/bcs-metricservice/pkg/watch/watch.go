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
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-metricservice/app/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-metricservice/pkg/storage"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-metricservice/pkg/zk"
	"sync"
	"time"
)

type Watcher struct {
	ctx    context.Context
	cancel context.CancelFunc

	event          chan *MetricEvent
	clusterLock    sync.RWMutex
	clusterWatcher map[string]*ClusterWatcher

	storage storage.Storage
	zk      zk.Zk
	config  *config.Config
}

func NewWatcher(storage storage.Storage, zk zk.Zk, config *config.Config) *Watcher {
	return &Watcher{storage: storage, zk: zk, config: config}
}

func (w *Watcher) Start() chan *MetricEvent {
	blog.Infof("start watching metric")
	w.ctx, w.cancel = context.WithCancel(context.Background())
	w.event = make(chan *MetricEvent, 1000)
	w.clusterLock = sync.RWMutex{}
	w.clusterWatcher = make(map[string]*ClusterWatcher)
	go w.manager()
	return w.event
}

func (w *Watcher) Stop() {
	if w.cancel != nil {
		blog.Infof("end watching metric")
		w.cancel()
	}
}

func (w *Watcher) manager() {
	blog.Infof("launch watcher manager")
	defer blog.Infof("shut down watcher manager")
	logTick := time.NewTicker(180 * time.Second)
	defer logTick.Stop()
	syncTick := time.NewTicker(3 * time.Second)
	defer syncTick.Stop()

	for {
		select {
		case <-w.ctx.Done():
			blog.Infof("end watching clusters")
			w.clusterLock.Lock()
			for _, v := range w.clusterWatcher {
				v.Stop()
			}
			w.clusterLock.Unlock()
			return
		case <-logTick.C:
			w.clusterLock.RLock()
			clusters := make([]string, 0)
			for k := range w.clusterWatcher {
				clusters = append(clusters, k)
			}
			blog.Infof("watching %d clusters: %v", len(clusters), clusters)
			w.clusterLock.RUnlock()
		case <-syncTick.C:
			if err := w.syncClusters(); err != nil {
				blog.Errorf("sync clusters failed: %v", err)
			}
		}
	}
}

func (w *Watcher) syncClusters() error {
	clusters, err := w.storage.GetClusters()
	if err != nil {
		return err
	}
	blog.V(3).Infof("sync clusters: %v", clusters)

	w.clusterLock.Lock()
	slot := make(map[string]bool, 0)
	for cluster := range w.clusterWatcher {
		slot[cluster] = true
	}
	for _, cluster := range clusters {
		slot[cluster] = false
		if w.clusterWatcher[cluster] == nil {
			cw := NewClusterWatcher(w.ctx, cluster, w.storage, w.zk, w.config)
			cw.Start(w.event)
			w.clusterWatcher[cluster] = cw
		}
	}
	for cluster, noExist := range slot {
		if noExist {
			w.clusterWatcher[cluster].Stop()
			delete(w.clusterWatcher, cluster)
		}
	}
	w.clusterLock.Unlock()
	return nil
}
