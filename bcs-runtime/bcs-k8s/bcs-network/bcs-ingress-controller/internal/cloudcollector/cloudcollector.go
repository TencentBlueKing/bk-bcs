/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cloudcollector

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/cloud"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	statusUpdateInterval = 30 * time.Second
)

// CloudCollector prometheus metric collector for cloud loadbalance
type CloudCollector struct {
	banckendMetric *prometheus.Desc
	cloudClient    cloud.LoadBalance
	k8sClient      client.Client
	mutex          sync.Mutex
	cache          StatusCache
}

// NewCloudCollector create cloud collector
func NewCloudCollector(cloudClient cloud.LoadBalance, k8sClient client.Client) *CloudCollector {
	return &CloudCollector{
		banckendMetric: newBackendHealthMetric(
			namespaceForCloudBalance, "backend_status", "status for backend health", nil),
		cloudClient: cloudClient,
		k8sClient:   k8sClient,
		cache:       NewStatusCache(),
	}
}

// Describe sends all possible descriptors of metrics to channel.
func (cc *CloudCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- cc.banckendMetric
}

// Collect fetches status from cloud
func (cc *CloudCollector) Collect(ch chan<- prometheus.Metric) {
	cc.mutex.Lock()
	defer cc.mutex.Unlock()

	//get status cache
	totalStatusMap := cc.cache.Get()
	for lbid, lbStatusListt := range totalStatusMap {
		for _, lbStatus := range lbStatusListt {
			var statuNum int
			switch lbStatus.Status {
			case cloud.BackendHealthStatusHealthy:
				statuNum = 1
			case cloud.BackendHealthStatusUnhealthy:
				statuNum = 0
			default:
				statuNum = 2
			}
			ch <- prometheus.MustNewConstMetric(cc.banckendMetric,
				prometheus.GaugeValue, float64(statuNum),
				[]string{lbid, lbStatus.ListenerID, strconv.Itoa(lbStatus.ListenerPort), lbStatus.Protocol,
					lbStatus.Host, lbStatus.Path, lbStatus.IP, strconv.Itoa(lbStatus.Port)}...)
		}
	}

}

// map key is region string, map value is lb id list
func (cc *CloudCollector) getLbMap(
	ingressList *networkextensionv1.IngressList,
	poolList *networkextensionv1.PortPoolList) map[string]map[string]map[string]struct{} {
	retMap := make(map[string]map[string]map[string]struct{})
	if ingressList != nil {
		for _, ingress := range ingressList.Items {
			for _, lbObj := range ingress.Status.Loadbalancers {
				mergeLbMap(retMap, lbObj.Region, ingress.GetNamespace(), lbObj.LoadbalancerID)
			}
		}
	}
	if poolList != nil {
		for _, pool := range poolList.Items {
			for _, itemStatus := range pool.Status.PoolItemStatuses {
				for _, lbObj := range itemStatus.PoolItemLoadBalancers {
					mergeLbMap(retMap, lbObj.Region, pool.GetNamespace(), lbObj.LoadbalancerID)
				}
			}
		}
	}
	return retMap
}

func (cc *CloudCollector) Start() {
	tiker := time.NewTicker(statusUpdateInterval)
	for {
		select {
		case <-tiker.C:
			cc.update()
		}
	}
}

func (cc *CloudCollector) update() {
	//get status data
	ingressList := &networkextensionv1.IngressList{}
	if err := cc.k8sClient.List(context.Background(), ingressList); err != nil {
		blog.Errorf("list ext ingresses failed when collect metrics, err %s", err.Error())
		return
	}
	poolList := &networkextensionv1.PortPoolList{}
	if err := cc.k8sClient.List(context.Background(), poolList); err != nil {
		blog.Errorf("list port pool failed when collect metrics, err %s", err.Error())
	}
	lbMap := cc.getLbMap(ingressList, poolList)

	totalStatusMap := make(map[string][]*cloud.BackendHealthStatus)
	for region, nsMap := range lbMap {
		for ns, idMap := range nsMap {
			lbIDs := getMapKeys(idMap)
			statusMap, err := cc.cloudClient.DescribeBackendStatus(region, ns, lbIDs)
			if err != nil {
				fetchBackendStatusMetric.Set(0)
				blog.Errorf("describe backend status of region %s lbids %v", region, lbIDs)
				return
			}
			fetchBackendStatusMetric.Set(0)
			for k, v := range statusMap {
				totalStatusMap[k] = v
			}
		}
	}
	//update status to cache
	cc.cache.UpdateCache(totalStatusMap)
}
