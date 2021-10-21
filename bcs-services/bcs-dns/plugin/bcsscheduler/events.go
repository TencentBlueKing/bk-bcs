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

package bcsscheduler

import (
	"log"
	"reflect"
	"strings"
	"time"

	bcstypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-dns/plugin/bcsscheduler/metrics"
	"github.com/coredns/coredns/plugin/etcd/msg"
)

func (bcs *BcsScheduler) svcOnAdd(obj interface{}) {
	if bcs.storage == nil || !bcs.registery.IsMaster() {
		return
	}
}

func (bcs *BcsScheduler) svcOnUpdate(old, cur interface{}) {
	if bcs.storage == nil || !bcs.registery.IsMaster() {
		return
	}
	//Service update, cluster ip list may be changed
	//update remote storage
}

func (bcs *BcsScheduler) svcOnDelete(obj interface{}) {
	if bcs.storage == nil || !bcs.registery.IsMaster() {
		return
	}
}

/*
 * EventFuncs for BcsEndpoint event funcs
 */

func (bcs *BcsScheduler) endpointOnAdd(obj interface{}) {
	if bcs.storage == nil || !bcs.registery.IsMaster() {
		return
	}

	start := time.Now()

	endpoint, ok := obj.(*bcstypes.BcsEndpoint)
	if !ok {
		log.Printf("[ERROR] scheduler endpoint ADD event get error data type")
		return
	}
	domain := endpoint.GetName() + "." + endpoint.GetNamespace() + ".svc." + bcs.PrimaryZone()
	svc := bcs.endpoint2Message(endpoint)
	if err := bcs.storage.AddService(domain, svc); err != nil {
		//todo(developer): DELETE error, try again?
		log.Printf("[ERROR] scheduler ADD endpoint %s to storage failed, %s", domain, err.Error())
		return
	}
	log.Printf("[WARN] scheduler ADD domain %s in sotrage success.", domain)
	metrics.StorageOperatorTotal.WithLabelValues(metrics.AddOperation).Inc()
	metrics.StorageOperatorLatency.WithLabelValues(metrics.AddOperation).Observe(time.Since(start).Seconds())
}

func (bcs *BcsScheduler) endpointOnUpdate(old, cur interface{}) {
	if bcs.storage == nil || !bcs.registery.IsMaster() {
		return
	}

	start := time.Now()

	if reflect.DeepEqual(old, cur) {
		return
	}
	oldEndpoint, ok := old.(*bcstypes.BcsEndpoint)
	if !ok {
		log.Printf("[ERROR] scheduler endpoint UPDATE get error data type in oldData.")
		return
	}
	curEndpoint, ok := cur.(*bcstypes.BcsEndpoint)
	if !ok {
		log.Printf("[ERROR] scheduler endpoint UPDATE get error data type in curData")
		return
	}
	domain := curEndpoint.GetName() + "." + curEndpoint.GetNamespace() + ".svc." + bcs.PrimaryZone()
	//change to endpoints
	oldSvc := bcs.endpoint2Message(oldEndpoint)
	curSvc := bcs.endpoint2Message(curEndpoint)
	if err := bcs.storage.UpdateService(domain, oldSvc, curSvc); err != nil {
		//todo(developer): update error, try again?
		log.Printf("[ERROR] scheduler update endpoint [%s] to storage failed, %s", domain, err.Error())
		return
	}
	log.Printf("[WARN] scheduler update %s to storage success.", domain)

	metrics.StorageOperatorTotal.WithLabelValues(metrics.UpdateOperation).Inc()
	metrics.StorageOperatorLatency.WithLabelValues(metrics.UpdateOperation).Observe(time.Since(start).Seconds())
}

func (bcs *BcsScheduler) endpointOnDelete(obj interface{}) {
	if bcs.storage == nil || !bcs.registery.IsMaster() {
		return
	}

	start := time.Now()

	endpoint, ok := obj.(*bcstypes.BcsEndpoint)
	if !ok {
		log.Printf("[ERROR] scheduler endpoint DELETE event get error data type")
		return
	}
	domain := endpoint.GetName() + "." + endpoint.GetNamespace() + ".svc." + bcs.PrimaryZone()
	svc := bcs.endpoint2Message(endpoint)
	if err := bcs.storage.DeleteService(domain, svc); err != nil {
		//todo(developer): DELETE error, try again?
		log.Printf("[ERROR] scheduler delete endpoint %s to storage failed, %s", domain, err.Error())
		return
	}
	log.Printf("[WARN] scheduler delete domain %s in sotrage success.", domain)
	metrics.StorageOperatorTotal.WithLabelValues(metrics.DeleteOperation).Inc()
	metrics.StorageOperatorLatency.WithLabelValues(metrics.DeleteOperation).Observe(time.Since(start).Seconds())
}

//endpoint2Message create etcd message with BcsEndpoint
//now only create etcd service for TypeA
//todo(developer): PTR/SRV message will support future
func (bcs *BcsScheduler) endpoint2Message(endpoint *bcstypes.BcsEndpoint) (svcList []msg.Service) {
	if endpoint == nil {
		return svcList
	}
	if endpoint.Endpoints == nil || len(endpoint.Endpoints) == 0 {
		return svcList
	}
	//check cluster ip address list
	var clusterIPs []string
	service := bcs.svcCache.GetServiceByEndpoint(endpoint)
	if service == nil {
		//todo(developer): lost bcs service info in cache. how to fix
		log.Printf("[ERROR] %s/%s endpoint lost bcs service in cache", endpoint.GetNamespace(), endpoint.GetName())
	} else {
		if service.Spec.ClusterIP != nil && len(service.Spec.ClusterIP) != 0 {
			clusterIPs = service.Spec.ClusterIP
		}
	}
	if len(clusterIPs) != 0 {
		//get cluster ip address, push to cluster storage
		//no matter what network mode is
		for _, ip := range clusterIPs {
			svc := createDefaultMessage(ip)
			svcList = append(svcList, svc)
		}
		return svcList
	}
	//no cluster ip, construct etcd service with endpoints
	for _, ep := range endpoint.Endpoints {
		mode := strings.ToLower(ep.NetworkMode)
		if mode == "none" {
			//filter none network mode
			continue
		}
		var svc msg.Service
		if mode == "bridge" || mode == "host" {
			//in mesos: no solution for `bridge`, `host` interconnection.
			//handle it with node ip address
			svc = createDefaultMessage(ep.NodeIP)
		} else {
			//importance: `default` mode is bridge in k8s.
			//user mode also setting to container ip address
			svc = createDefaultMessage(ep.ContainerIP)
		}
		svcList = append(svcList, svc)
	}
	return svcList
	//todo(developer): SRV SUPPORT
}

//createDefaultMessage default etcd service message
func createDefaultMessage(ip string) msg.Service {
	return msg.Service{
		Host:     ip,
		Port:     0,
		Priority: 10,
		Weight:   10,
		TTL:      5,
	}
}
