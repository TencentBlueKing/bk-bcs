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

package resources

import (
	"fmt"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"sync"

	commtypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-hpacontroller/hpacontroller/config"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-hpacontroller/hpacontroller/metrics"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-hpacontroller/hpacontroller/reflector"
)

type resourceMetrics struct {
	sync.RWMutex
	//hpa controller config
	config *config.Config

	// Reflector watches a specified resource and causes all changes to be reflected in the given store
	store reflector.Reflector

	//hpa autoscaler work queue, key = BcsAutoscaler.GetUuid()
	workQueue map[string]*resourcesCollector
}

func NewResourceMetrics(conf *config.Config, store reflector.Reflector) metrics.MetricsController {
	resources := &resourceMetrics{
		config:    conf,
		store:     store,
		workQueue: make(map[string]*resourcesCollector),
	}

	return resources
}

//start to collect scaler metrics
func (resources *resourceMetrics) StartScalerMetrics(scaler *commtypes.BcsAutoscaler) {
	resources.Lock()
	defer resources.Unlock()

	_, ok := resources.workQueue[scaler.GetUuid()]
	if ok {
		return
	}

	//start collector scaler target ref resources metrics
	resources.workQueue[scaler.GetUuid()] = newResourcesCollector(resources, scaler)
	resources.workQueue[scaler.GetUuid()].start()
	blog.Infof("start collector scaler %s resources metrics", scaler.GetUuid())
}

//stop to collect scaler metrics
func (resources *resourceMetrics) StopScalerMetrics(scaler *commtypes.BcsAutoscaler) {
	resources.Lock()
	defer resources.Unlock()

	collector, ok := resources.workQueue[scaler.GetUuid()]
	if !ok {
		return
	}
	collector.stop()
	blog.Infof("stop collector scaler %s resources metrics", scaler.GetUuid())
}

// GetResourceMetric gets the given resource metric (and an associated oldest timestamp)
// for all taskgroup matching the specified scaler uuid
func (resources *resourceMetrics) GetResourceMetric(resourceName, uuid string) (metrics.TaskgroupMetricsInfo, error) {
	resources.RLock()
	defer resources.RUnlock()

	collector, ok := resources.workQueue[uuid]
	if !ok {
		return nil, fmt.Errorf("sacler %s not found", uuid)
	}

	switch resourceName {
	case metrics.TaskgroupResourcesCpuMetricsName:
		return collector.getCpuMetricsInfo(), nil

	case metrics.TaskgroupResourcesMemoryMetricsName:
		return collector.getMemoryMetricsInfo(), nil

	}

	return nil, fmt.Errorf("resource name %s is invalid", resourceName)
}
