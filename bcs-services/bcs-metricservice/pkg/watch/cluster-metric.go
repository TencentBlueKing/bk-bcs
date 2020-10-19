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
	"github.com/Tencent/bk-bcs/bcs-common/common/codec"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-metricservice/pkg/driver"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-metricservice/pkg/storage"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-metricservice/pkg/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/operator"
	"time"
)

func (cw *ClusterWatcher) metricManager() {
	syncTick := time.NewTicker(30 * time.Second)
	defer syncTick.Stop()
	ctx, cancel := context.WithCancel(cw.ctx)
	cw.syncMetric()
	go cw.watchMetric(ctx)

	for {
		select {
		case <-cw.ctx.Done():
			blog.Warnf("metric watching cluster manager shut down: %s", cw.clusterId)
			cancel()
			return
		case <-syncTick.C:
			cw.syncMetric()
		case event := <-cw.metricEvent:
			if event == operator.EventWatchBreak {
				cancel()

				wait := time.Duration(5)
				blog.Infof("metric-cluster-sync(%s) will reconnect in %d seconds", cw.clusterId, wait)
				// after watcher connection break, sleep 5 seconds and retry connecting
				time.Sleep(wait * time.Second)
				ctx, cancel = context.WithCancel(cw.ctx)
				go cw.watchMetric(ctx)
				continue
			}

			blog.V(3).Infof("received raw metric event: %v", event)
			metricEvent, err := convertStorageMetricEvent2MetricEvent(event)
			if err != nil {
				blog.Errorf("convert storage event to metric event failed: %v", err)
				continue
			}

			cw.handleWatchMetricEvent(metricEvent)
		}
	}
}

func (cw *ClusterWatcher) syncMetric() {
	r, err := cw.storage.QueryMetric(&storage.Param{ClusterID: cw.clusterId, Type: types.ResourceMetricType})
	if err != nil {
		blog.Errorf("metric-cluster-sync(%s) get metric failed: %v", cw.clusterId, err)
		return
	}

	events, err := convertStorageMetricData2MetricEvent(r)
	if err != nil {
		blog.Errorf("metric-cluster-sync(%s) convert to event failed: %v", cw.clusterId, err)
		return
	}

	blog.Infof("metric-cluster-sync(%s): %d", cw.clusterId, len(events))
	cw.handleSyncMetricEvent(events)
}

func (cw *ClusterWatcher) watchMetric(ctx context.Context) {
	blog.V(3).Infof("enter metric-watch(%s) goroutine", cw.clusterId)
	watcher, err := cw.storage.GetMetricWatcher(&storage.Param{ClusterID: cw.clusterId, Type: types.ResourceMetricType})
	if err != nil {
		blog.Errorf("metric-cluster-watcher(%s) brings up failed: %v", cw.clusterId, err)
		cw.metricEvent <- operator.EventWatchBreak
		return
	}

	// close the watcher if context done
	go func() {
		select {
		case <-ctx.Done():
			blog.Infof("metric-cluster-watcher(%s) shut down ", cw.clusterId)
			watcher.Close()
		}
	}()

	blog.Infof("metric-cluster-watcher(%s) brings up", cw.clusterId)
	for {
		event, err := watcher.Next()
		if err != nil {
			blog.Errorf("metric-cluster-watcher(%s) get event failed: %v", cw.clusterId, err)
		}
		cw.metricEvent <- event
		if event == operator.EventWatchBreak {
			blog.Errorf("metric-cluster-watcher(%s) connection break", cw.clusterId)
			return
		}
	}
}

func (cw *ClusterWatcher) handleWatchMetricEvent(event *MetricEvent) {
	switch event.Type {
	case EventMetricUpd:
		cw.metricLock.Lock()
		cw.metric[event.ID] = event.Metric
		cw.metricLock.Unlock()

		event.First, _ = cw.isApplicationSettingsNotExist(event)
	case EventMetricDel:
		cw.metricLock.Lock()
		if cw.metric[event.ID] == nil {
			cw.metricLock.Unlock()
			return
		}
		event.Metric = cw.metric[event.ID]
		delete(cw.metric, event.ID)
		cw.metricLock.Unlock()

		event.Last = cw.isLastMetricInNamespace(event.Metric)
	default:
		return
	}
	cw.outPutEvent <- event
}

func (cw *ClusterWatcher) handleSyncMetricEvent(events []*MetricEvent) {
	slot := make(map[string]bool, 0)
	cw.metricLock.RLock()
	for k := range cw.metric {
		slot[k] = true
	}
	cw.metricLock.RUnlock()

	for _, event := range events {
		slot[event.ID] = false

		isCDiff, err := cw.isCollectorSettingsVersionDiff(event)
		if err != nil {
			continue
		}
		isANExist, err := cw.isApplicationSettingsNotExist(event)
		if err != nil {
			continue
		}
		event.First = isANExist

		if isCDiff || isANExist {
			cw.metricLock.Lock()
			cw.metric[event.ID] = event.Metric
			cw.metricLock.Unlock()

			cw.outPutEvent <- event
			continue
		}

		cw.metricLock.RLock()
		metric := cw.metric[event.ID]
		cw.metricLock.RUnlock()

		if metric != nil && metric.Version == event.Metric.Version {
			continue
		}

		cw.metricLock.Lock()
		cw.metric[event.ID] = event.Metric
		cw.metricLock.Unlock()
	}

	for k, v := range slot {
		if v {
			cw.metricLock.RLock()
			metric := cw.metric[k]
			cw.metricLock.RUnlock()
			cw.outPutEvent <- &MetricEvent{
				ID:     k,
				Metric: metric,
				Type:   EventMetricDel,
				Last:   cw.isLastMetricInNamespace(metric),
			}
			cw.metricLock.Lock()
			delete(cw.metric, k)
			cw.metricLock.Unlock()
		}
	}
}

func (cw *ClusterWatcher) isLastMetricInNamespace(metric *types.Metric) bool {
	cw.metricLock.RLock()
	defer cw.metricLock.RUnlock()
	for _, m := range cw.metric {
		if m.Namespace == metric.Namespace && m.Name != metric.Name {
			return false
		}
	}
	return true
}

func (cw *ClusterWatcher) queryCollectorSettings(metric *types.Metric) (data []*StorageCollectorIf, err error) {
	r, err := cw.storage.QueryMetric(&storage.Param{
		ClusterID: cw.clusterId,
		Type:      types.ResourceCollectorType,
		Name:      metric.Name,
		Namespace: metric.Namespace,
	})

	if err != nil {
		return
	}

	err = codec.DecJson(r, &data)
	return
}

func (cw *ClusterWatcher) isCollectorSettingsVersionDiff(event *MetricEvent) (b bool, err error) {
	var data []*StorageCollectorIf
	if data, err = cw.queryCollectorSettings(event.Metric); err != nil {
		blog.Errorf("metric-cluster-sync(%s) check collector failed: %v", cw.clusterId, err)
		return
	}

	if len(data) == 0 || data[0].Data.Version != event.Metric.Version {
		b = true
		return
	}
	return
}

func (cw *ClusterWatcher) isApplicationSettingsNotExist(event *MetricEvent) (b bool, err error) {
	r, err := cw.storage.QueryMetric(&storage.Param{
		ClusterID: cw.clusterId,
		Type:      types.ResourceApplicationType,
		Name:      driver.GetApplicationName(event.Metric),
		Namespace: event.Metric.Namespace,
	})

	if err != nil {
		blog.Errorf("metric-cluster-sync(%s) get application failed: %v", cw.clusterId, err)
		return
	}

	data := make([]*StorageApplicationIf, 0)
	if err = codec.DecJson(r, &data); err != nil {
		blog.Errorf("metric-cluster-sync(%s) decode application failed: %v | %s", cw.clusterId, err, string(r))
		return
	}

	b = len(data) == 0
	return
}

type StorageCollectorIf struct {
	Data *types.ApplicationCollectorCfg `json:"data"`
}

type StorageApplicationIf struct {
	Data string `json:"data"`
}
