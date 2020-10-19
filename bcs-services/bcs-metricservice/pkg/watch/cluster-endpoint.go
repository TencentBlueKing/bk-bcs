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
	"fmt"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	btypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-metricservice/pkg/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-metricservice/pkg/zk"
	"time"
)

func (cw *ClusterWatcher) endpointManger() {
	syncTick := time.NewTicker(30 * time.Second)
	defer syncTick.Stop()
	endpoints := cw.config.EndpointWatchPath

	event := cw.zk.Watch(endpoints)
	cw.syncEndpoint(endpoints)

	for {
		select {
		case <-cw.ctx.Done():
			blog.Warnf("endpoint watching cluster manager shut down")
			cw.zk.Close()
			return
		case <-syncTick.C:
			cw.syncEndpoint(endpoints)
		case myEvent := <-event:
			blog.V(3).Infof("received raw endpoint event(%s): %v", cw.clusterId, myEvent)
			cw.handleEndpointEvent(myEvent)
		}
	}
}

func (cw *ClusterWatcher) syncEndpoint(paths []string) {
	endpoints := cw.zk.List(paths)
	ipMeta := make(map[string]btypes.ObjectMeta)

	for _, endpoint := range endpoints {
		key := fmt.Sprintf("%s:%d", endpoint.IP, endpoint.Port)
		ipMeta[key] = btypes.ObjectMeta{
			Name:        endpoint.Name,
			NameSpace:   endpoint.Path,
			Annotations: map[string]string{types.BcsComponentsSchemeKey: endpoint.Scheme},
		}
	}

	cw.metricLock.RLock()
	defer cw.metricLock.RUnlock()
	for _, metric := range cw.metric {
		if types.GetClusterType(metric.ClusterType) != types.BcsComponents {
			continue
		}
		cw.outPutEvent <- &MetricEvent{
			Metric: metric,
			Type:   EventDynamicUpd,
			Meta:   ipMeta,
		}

		// should be only one metric in cluster BcsComponents
		break
	}
}

func (cw *ClusterWatcher) handleEndpointEvent(event *zk.Event) {
	cw.metricLock.RLock()
	defer cw.metricLock.RUnlock()

	for _, metric := range cw.metric {
		if types.GetClusterType(metric.ClusterType) != types.BcsComponents {
			continue
		}

		collectors, err := cw.queryCollectorSettings(metric)
		if err != nil {
			blog.Errorf("handle endpoint event, check collector failed: %v", err)
			continue
		}
		if len(collectors) == 0 {
			blog.Warnf("handle endpoint event there is not collector(%s): name(%s) namespace(%s)", cw.clusterId, metric.Namespace, metric.Name)
			continue
		}
		collector := collectors[0]
		ipMeta := make(map[string]btypes.ObjectMeta)
		for _, c := range collector.Data.Cfg {
			key := fmt.Sprintf("%s:%d", c.IP, c.Port)
			if _, ok := ipMeta[key]; !ok {
				ipMeta[key] = c.Meta
			}
		}

		for _, endpoint := range event.Endpoints {
			key := fmt.Sprintf("%s:%d", endpoint.IP, endpoint.Port)

			switch event.Type {
			case zk.EventNodeUp:
				ipMeta[key] = btypes.ObjectMeta{
					Name:        endpoint.Name,
					NameSpace:   endpoint.Path,
					Annotations: map[string]string{types.BcsComponentsSchemeKey: endpoint.Scheme}}
			case zk.EventNodeDown:
				delete(ipMeta, key)
			default:
				continue
			}
		}

		cw.outPutEvent <- &MetricEvent{
			Metric: metric,
			Type:   EventDynamicUpd,
			Meta:   ipMeta,
		}

		// should be only one metric in cluster BcsComponents
		break
	}
}
