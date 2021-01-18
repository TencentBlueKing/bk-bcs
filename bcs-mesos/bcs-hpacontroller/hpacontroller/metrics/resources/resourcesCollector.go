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
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	commtypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	schedtypes "github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/schetypes"
	containertypes "github.com/Tencent/bk-bcs/bcs-mesos/bcs-container-executor/container"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-hpacontroller/hpacontroller/metrics"

	cadClient "github.com/google/cadvisor/client"
	cadvisorV1 "github.com/google/cadvisor/info/v1"
)

type taskgroupStats struct {
	id        string                //taskgroup id
	taskgroup *schedtypes.TaskGroup //taskgroup info
	//key = contianerId, value = *cadvisorV1.ContainerStats
	containerStats map[string]*cadvisorV1.ContainerStats
	//cadvisor client
	cadvisorClient *cadClient.Client
	//colletor numbers
	number int
	//collector time stamp
	timeStamp time.Time
}

type resourcesCollector struct {
	sync.RWMutex
	controller *resourceMetrics

	scaler *commtypes.BcsAutoscaler

	//key = taskgroupid
	//value = TaskgroupMetricsInfo
	cpuMetricsInfo    metrics.TaskgroupMetricsInfo
	memoryMetricsInfo metrics.TaskgroupMetricsInfo

	//key = taskgroupid
	//value = taskgroupStats
	taskgroupStats map[string]*taskgroupStats

	// the latest value of metrics as an average aggregated across window minutes
	// seconds, default 30s
	collectMetricsWindow int

	validMetricsWindow int

	//container resources cadvisor port
	cadvisorPort int

	ctx    context.Context
	cancel context.CancelFunc
}

func newResourcesCollector(controller *resourceMetrics, scaler *commtypes.BcsAutoscaler) *resourcesCollector {
	collector := &resourcesCollector{
		controller:           controller,
		scaler:               scaler,
		cadvisorPort:         controller.config.CadvisorPort,
		collectMetricsWindow: controller.config.CollectMetricsWindow,
		validMetricsWindow:   controller.config.CollectMetricsWindow + 30,
		taskgroupStats:       make(map[string]*taskgroupStats),
		cpuMetricsInfo:       make(metrics.TaskgroupMetricsInfo),
		memoryMetricsInfo:    make(metrics.TaskgroupMetricsInfo),
	}

	collector.ctx, collector.cancel = context.WithCancel(context.Background())

	return collector
}

func (collector *resourcesCollector) start() {
	//start ticker collector metrics
	go collector.tickerCollectorMetrics()
}

func (collector *resourcesCollector) stop() {
	collector.cancel()
}

func (collector *resourcesCollector) getCpuMetricsInfo() metrics.TaskgroupMetricsInfo {
	collector.RLock()
	metric := collector.cpuMetricsInfo
	collector.RUnlock()

	return metric
}

func (collector *resourcesCollector) getMemoryMetricsInfo() metrics.TaskgroupMetricsInfo {
	collector.RLock()
	metric := collector.memoryMetricsInfo
	collector.RUnlock()

	return metric
}

func (collector *resourcesCollector) tickerCollectorMetrics() {
	ticker := time.NewTicker(time.Second * time.Duration(collector.collectMetricsWindow))
	defer ticker.Stop()

	for {
		select {
		case <-collector.ctx.Done():
			blog.Infof("stop ticker collector scaler %s resources metrics", collector.scaler.GetUuid())
			return

		case <-ticker.C:
			blog.Infof("ticker collector scaler %s resources metrics", collector.scaler.GetUuid())
			collector.collectorMetrics()
		}
	}
}

func (collector *resourcesCollector) collectorMetrics() {
	//sync the latest taskgroup to collector.taskgroupStats queue
	blog.Infof("sync scaler %s target ref taskgroups start...", collector.scaler.GetUuid())
	err := collector.syncTargetRefTaskgroups()
	if err != nil {
		blog.Errorf("sync scaler %s target ref taskgroups error %s", collector.scaler.GetUuid(), err.Error())
		return
	}
	blog.Infof("sync scaler %s target ref taskgroups done", collector.scaler.GetUuid())

	for _, stats := range collector.taskgroupStats {
		collector.Lock()
		collector.memoryMetricsInfo[stats.taskgroup.ID] = metrics.TaskgroupMetric{}
		collector.cpuMetricsInfo[stats.taskgroup.ID] = metrics.TaskgroupMetric{}
		collector.Unlock()

		oldContainerStats := stats.containerStats
		oldTimestamp := stats.timeStamp
		//collector from cadvisor service
		success := true
		newContainerStats := make(map[string]*cadvisorV1.ContainerStats)
		for containerId := range oldContainerStats {
			containerInfo, err := stats.cadvisorClient.DockerContainer(containerId, &cadvisorV1.ContainerInfoRequest{NumStats: 1})
			if err != nil {
				blog.Errorf("scaler %s taskgroup %s docker stats error %s, and continue collector metrics",
					collector.scaler.GetUuid(), stats.id, err.Error())
				success = false
				break
			}

			newContainerStats[containerId] = containerInfo.Stats[0]
		}
		//if get cadvisor container stats failed, then continue
		if !success {
			continue
		}

		//update current container stats
		stats.containerStats = newContainerStats
		stats.timeStamp = time.Now()
		stats.number++
		//if old container stats ==nil or stats timestamp is too old, and don't need to compute metrics
		for containerid, cStats := range oldContainerStats {
			if cStats == nil {
				blog.Errorf("scaler %s taskgroup %s container %s stats is nil, and continue collector metrics",
					collector.scaler.GetUuid(), stats.id, containerid)
				success = false
				break
			}

			if (time.Now().Unix() - cStats.Timestamp.Unix()) > int64(collector.validMetricsWindow) {
				blog.Errorf("scaler %s taskgroup %s container %s stats timestamp %s is invalid, and continue collector metrics",
					collector.scaler.GetUuid(), stats.id, containerid, cStats.Timestamp.Format("2006-01-02 15:04:05"))
				success = false
				break
			}
		}
		//if old cadvisor container stats == nil, then continue
		if !success {
			continue
		}

		duration := float32(stats.timeStamp.Unix() - oldTimestamp.Unix())
		cpuMetric := collector.computeCpuMetricsInfo(oldContainerStats, stats, duration)
		memMetric := collector.computeMemoryMetricsInfo(stats, duration)
		blog.Infof("scaler %s compute taskgroup %s cpu metrics %.2f", collector.scaler.GetUuid(), stats.taskgroup.ID, cpuMetric.Value)
		blog.Infof("scaler %s compute taskgroup %s memory metrics %.2f", collector.scaler.GetUuid(), stats.taskgroup.ID, memMetric.Value)

		collector.Lock()
		collector.memoryMetricsInfo[stats.taskgroup.ID] = memMetric
		collector.cpuMetricsInfo[stats.taskgroup.ID] = cpuMetric
		collector.Unlock()
	}
}

func (collector *resourcesCollector) computeCpuMetricsInfo(oldStats map[string]*cadvisorV1.ContainerStats,
	taskgroupStats *taskgroupStats, duration float32) metrics.TaskgroupMetric {

	var total float32
	for cid, stats := range oldStats {
		newCpu := taskgroupStats.containerStats[cid].Cpu.Usage
		containerInfo := collector.getContainerByContainerid(taskgroupStats.taskgroup, cid)
		if containerInfo == nil {
			blog.Errorf("taskgroup %s not found container %s", taskgroupStats.taskgroup.ID, cid)
			continue
		}
		//cpu usage, cpu_usage*100 = docker stats cpu usage
		usage := float32(newCpu.Total-stats.Cpu.Usage.Total) / duration / 1000000000
		total += usage

		blog.V(3).Infof("taskgroup %s container %s cpu usage %f, and launch cpu resource %f",
			taskgroupStats.taskgroup.ID, cid, usage, taskgroupStats.taskgroup.LaunchResource.Cpus)

	}

	metric := metrics.TaskgroupMetric{
		Timestamp: time.Now(),
		Window:    int(duration),
		Value:     total / float32(taskgroupStats.taskgroup.LaunchResource.Cpus) * 100,
	}

	return metric
}

func (collector *resourcesCollector) computeMemoryMetricsInfo(taskgroupStats *taskgroupStats, duration float32) metrics.TaskgroupMetric {

	var total float32 //MB
	for cid, stats := range taskgroupStats.containerStats {

		//Bytes to MB
		usage := stats.Memory.Usage / 1024 / 1024
		total += float32(usage)

		blog.V(3).Infof("taskgroup %s container %s memory usage %dMB, and launch memory resource %f",
			taskgroupStats.taskgroup.ID, cid, usage, taskgroupStats.taskgroup.LaunchResource.Mem)
	}

	metric := metrics.TaskgroupMetric{
		Timestamp: time.Now(),
		Window:    int(duration),
		Value:     total / float32(taskgroupStats.taskgroup.LaunchResource.Mem) * 100,
	}

	return metric
}

func (collector *resourcesCollector) getContainerByContainerid(taskgroup *schedtypes.TaskGroup, containerId string) *schedtypes.Task {
	for _, task := range taskgroup.Taskgroup {
		if strings.Contains(task.StatusData, containerId) {
			return task
		}
	}

	return nil
}

//sync the latest taskgroup to collector.taskgroupStats queue
func (collector *resourcesCollector) syncTargetRefTaskgroups() error {
	var taskgroups []*schedtypes.TaskGroup
	var err error

	targetRef := collector.scaler.Spec.ScaleTargetRef
	switch targetRef.Kind {
	case commtypes.AutoscalerTargetRefDeployment:
		taskgroups, err = collector.controller.store.ListTaskgroupRefDeployment(targetRef.Namespace, targetRef.Name)

	case commtypes.AutoscalerTargetRefApplication:
		taskgroups, err = collector.controller.store.ListTaskgroupRefApplication(targetRef.Namespace, targetRef.Name)
	}

	if err != nil {
		blog.Errorf("sync scaler %s taskgroups error %s", collector.scaler.GetUuid(), err.Error())
		return err
	}

	collector.Lock()
	//key = taskgroupid
	currentQueue := make(map[string]struct{}, len(collector.taskgroupStats))
	for k := range collector.taskgroupStats {
		currentQueue[k] = struct{}{}
	}

	for _, taskgroup := range taskgroups {
		//if taskgroup status is not running, contine
		if taskgroup.Status != schedtypes.TASKGROUP_STATUS_RUNNING {
			blog.Warnf("scaler %s taskgroup %s status %s", collector.scaler.GetUuid(), taskgroup.ID, taskgroup.Status)
			continue
		}

		// if zk taskgroup exist, then delete currentQueue
		delete(currentQueue, taskgroup.ID)

		//if taskgroup is already in the workQueue, then continue
		_, ok := collector.taskgroupStats[taskgroup.ID]
		if ok {
			/*blog.V(3).Infof("ticker sync scaler %s taskgroup %s already exists",
			collector.scaler.GetUuid(), taskgroup.ID)*/
			continue
		}

		//add taskgroup into workqueue
		blog.Infof("add taskgroup %s into scaler %s workqueue", taskgroup.ID, collector.scaler.GetUuid())
		collector.taskgroupStats[taskgroup.ID] = &taskgroupStats{
			id:             taskgroup.ID,
			taskgroup:      taskgroup,
			containerStats: make(map[string]*cadvisorV1.ContainerStats),
		}

		marshalSuccess := true
		for _, task := range taskgroup.Taskgroup {
			var bcsInfo containertypes.BcsContainerInfo
			err := json.Unmarshal([]byte(task.StatusData), &bcsInfo)
			if err != nil {
				blog.Errorf("task %s Unmarshal StatusData error %s", task.ID, err.Error())
				marshalSuccess = false
				break

			}

			collector.taskgroupStats[taskgroup.ID].containerStats[bcsInfo.ID] = nil
		}
		if !marshalSuccess {
			continue
		}

		hostIp := taskgroup.Taskgroup[0].AgentIPAddress
		url := fmt.Sprintf("http://%s:%d/", hostIp, collector.cadvisorPort)
		collector.taskgroupStats[taskgroup.ID].cadvisorClient, _ = cadClient.NewClientWithTimeout(url, time.Second*5)

		//add TaskgroupMetric
		collector.cpuMetricsInfo[taskgroup.ID] = metrics.TaskgroupMetric{}
		collector.memoryMetricsInfo[taskgroup.ID] = metrics.TaskgroupMetric{}
	}

	//delete invalid taskgroup in workqueue
	for k := range currentQueue {
		//add taskgroup into workqueue
		blog.Infof("delete taskgroup %s into scaler %s workqueue", k, collector.scaler.GetUuid())
		delete(collector.taskgroupStats, k)
		delete(collector.cpuMetricsInfo, k)
		delete(collector.memoryMetricsInfo, k)
	}
	collector.Unlock()

	return nil
}
