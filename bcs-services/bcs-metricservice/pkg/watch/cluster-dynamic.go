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
	"fmt"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/codec"
	btypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-metricservice/pkg/driver"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-metricservice/pkg/storage"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-metricservice/pkg/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/operator"
	"time"
)

func (cw *ClusterWatcher) dynamicManager() {
	syncTick := time.NewTicker(30 * time.Second)
	defer syncTick.Stop()
	ctx, cancel := context.WithCancel(cw.ctx)
	cw.syncDynamic()
	go cw.watchDynamic(ctx)

	for {
		select {
		case <-cw.ctx.Done():
			blog.Warnf("metric watching cluster manager shut down: %s", cw.clusterId)
			cancel()
			return
		case <-syncTick.C:
			cw.syncDynamic()
		case event := <-cw.dynamicEvent:
			if event == operator.EventWatchBreak {
				cancel()

				wait := time.Duration(5)
				blog.Infof("dynamic-cluster-sync(%s) will reconnect in %d seconds", cw.clusterId, wait)
				// after watcher connection break, sleep 5 seconds and retry connecting
				time.Sleep(wait * time.Second)
				ctx, cancel = context.WithCancel(cw.ctx)
				go cw.watchDynamic(ctx)
				continue
			}

			blog.V(3).Infof("received raw dynamic event(%s): %v", cw.clusterId, event)
			if err := cw.handleWatchDynamicMetric(event); err != nil {
				blog.Errorf("handle dynamic metric failed: %v", err)
			}
		}
	}
}

func (cw *ClusterWatcher) watchDynamic(ctx context.Context) {
	blog.V(3).Infof("enter dynamic-watch(%s) goroutine", cw.clusterId)
	watcher, err := cw.storage.GetDynamicWatcher(&storage.Param{ClusterID: cw.clusterId})
	if err != nil {
		blog.Errorf("dynamic-cluster-watcher(%s) brings up failed: %v", cw.clusterId, err)
		cw.dynamicEvent <- operator.EventWatchBreak
		return
	}

	// close the watcher if context done
	go func() {
		select {
		case <-ctx.Done():
			blog.Infof("dynamic-cluster-watcher(%s) shut down ", cw.clusterId)
			watcher.Close()
		}
	}()

	blog.Infof("dynamic-cluster-watcher(%s) brings up", cw.clusterId)
	for {
		event, err := watcher.Next()
		if err != nil {
			blog.Errorf("dynamic-cluster-watcher(%s) get event failed: %v", cw.clusterId, err)
		}
		cw.dynamicEvent <- event
		if event == operator.EventWatchBreak {
			blog.Errorf("dynamic-cluster-watcher(%s) connection break", cw.clusterId)
			return
		}
	}
}

func (cw *ClusterWatcher) syncDynamic() {
	t := cw.getClusterType()
	if t == types.ClusterUnknown {
		blog.Warnf("cluster type of %s is unknown yet", cw.clusterId)
		return
	}

	r, err := cw.storage.QueryDynamic(&storage.Param{ClusterID: cw.clusterId, ClusterType: t, Type: t.GetContainerTypeName()})
	if err != nil {
		blog.Errorf("dynamic-cluster-sync(%s) get dynamic failed: %v", cw.clusterId, err)
		return
	}

	cw.metricLock.RLock()
	defer cw.metricLock.RUnlock()

	for _, metric := range cw.metric {
		if types.GetClusterType(metric.ClusterType) == types.BcsComponents {
			continue
		}
		ipMeta, err := driver.GetIPMetaFromDynamic(r, metric)
		if err != nil {
			blog.Errorf("get IPMeta(%s) failed, metric(namespace %s, name %s): %v", cw.clusterId, metric.Namespace, metric.Name, err)
			continue
		}

		blog.V(3).Infof("get ipMeta clusterId(%s) namespace(%s) metricName(%s): %v", metric.ClusterID, metric.Namespace, metric.Name, ipMeta)
		cw.outPutEvent <- &MetricEvent{
			Metric: metric,
			Type:   EventDynamicUpd,
			Meta:   ipMeta,
		}
	}
	blog.Infof("dynamic-cluster-sync(%s): %d", cw.clusterId, len(cw.metric))
}

func (cw *ClusterWatcher) getDynamicIPMeta(metric *types.Metric) (ipMeta map[string]btypes.ObjectMeta, err error) {
	t := cw.getClusterType()
	if t == types.ClusterUnknown {
		blog.Warnf("cluster type of %s is unknown yet", cw.clusterId)
		err = fmt.Errorf("cluster type unknown")
		return
	}

	r, err := cw.storage.QueryDynamic(&storage.Param{ClusterID: cw.clusterId, Namespace: metric.Namespace, ClusterType: t, Type: t.GetContainerTypeName()})
	if err != nil {
		blog.Errorf("get dynamic clusterId(%s) failed: %v", cw.clusterId, err)
		return
	}

	ipMeta, err = driver.GetIPMetaFromDynamic(r, metric)
	if err != nil {
		blog.Errorf("get dynamic ipMeta clusterId(%s) failed: %v", cw.clusterId, err)
		return
	}
	return
}

func (cw *ClusterWatcher) handleWatchDynamicMetric(storageEvent *operator.Event) (err error) {
	var b []byte
	if err = codec.EncJson([]interface{}{storageEvent.Value}, &b); err != nil {
		return
	}

	cw.metricLock.RLock()
	defer cw.metricLock.RUnlock()

	for _, metric := range cw.metric {
		if types.GetClusterType(metric.ClusterType) == types.BcsComponents {
			continue
		}

		var ipMeta map[string]btypes.ObjectMeta

		switch storageEvent.Type {
		case operator.Del:
			// if delete event occurred, query dynamic to fresh the ipMeta
			ipMeta, err = cw.getDynamicIPMeta(metric)
			if err != nil {
				continue
			}
		case operator.Add, operator.Chg:
			// different namespace can be passed
			if ns, ok := storageEvent.Value["namespace"].(string); !ok || ns != metric.Namespace {
				continue
			}

			ipMeta, err = driver.GetIPMetaFromDynamic(b, metric)
			if err != nil {
				blog.Errorf("get IPMeta(%s) failed, metric(namespace %s, name %s): %v", cw.clusterId, metric.Namespace, metric.Name, err)
				continue
			}

			collectors, err := cw.queryCollectorSettings(metric)
			if err != nil {
				blog.Errorf("handle dynamic metric, check collector failed: %v", err)
				continue
			}
			if len(collectors) == 0 {
				blog.Warnf("handle dynamic metric there is not collector(%s): name(%s) namespace(%s)", cw.clusterId, metric.Namespace, metric.Name)
				continue
			}
			collector := collectors[0]
			for _, c := range collector.Data.Cfg {
				if _, ok := ipMeta[c.IP]; !ok {
					ipMeta[c.IP] = c.Meta
				}
			}
		default:
			continue
		}

		cw.outPutEvent <- &MetricEvent{
			Metric: metric,
			Type:   EventDynamicUpd,
			Meta:   ipMeta,
		}
	}
	return
}

func convertStorageMetricEvent2MetricEvent(storageEvent *operator.Event) (metricEvent *MetricEvent, err error) {
	var b []byte
	if err = codec.EncJson(storageEvent.Value, &b); err != nil {
		return
	}

	var sMIf StorageMetricIf
	if err = codec.DecJson(b, &sMIf); err != nil {
		return
	}

	metricEvent = &MetricEvent{ID: sMIf.ID, Metric: sMIf.Data}
	switch storageEvent.Type {
	case operator.Add, operator.Chg:
		metricEvent.Type = EventMetricUpd
	case operator.Del:
		metricEvent.Type = EventMetricDel
	default:
		err = fmt.Errorf("unknown event type: %s", storageEvent.Type)
	}
	return
}

func convertStorageMetricData2MetricEvent(storageIf []byte) (metricEvent []*MetricEvent, err error) {
	var data []StorageMetricIf
	if err = codec.DecJson(storageIf, &data); err != nil {
		return
	}

	metricEvent = make([]*MetricEvent, 0)
	for _, v := range data {
		metricEvent = append(metricEvent, &MetricEvent{
			ID:     v.ID,
			Metric: v.Data,
			Type:   EventMetricUpd,
		})
	}
	return
}

type StorageMetricIf struct {
	ID   string        `json:"_id"`
	Data *types.Metric `json:"data"`
}
