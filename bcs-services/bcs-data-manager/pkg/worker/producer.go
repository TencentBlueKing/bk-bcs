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
 */

package worker

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/codec"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/msgqueue"
	"github.com/panjf2000/ants/v2"
	"github.com/robfig/cron/v3"
	"go-micro.dev/v4/broker"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/bcsmonitor"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/datajob"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/kafka"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/prom"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/utils"
)

// Producer produce data job
type Producer struct {
	msgQueue        msgqueue.MessageQueue
	cron            *cron.Cron
	k8sStorageCli   bcsapi.Storage
	mesosStorageCli bcsapi.Storage
	ctx             context.Context
	cancel          context.CancelFunc
	resourceGetter  common.GetterInterface
	concurrency     int
	bcsMonitorCli   bcsmonitor.ClientInterface // nolint
	needSendKafka   bool
	kafkaConn       kafka.KafkaInterface
}

// NewProducer new producer
func NewProducer(rootCtx context.Context, msgQueue msgqueue.MessageQueue, cron *cron.Cron,
	k8sStorageCli, mesosStorageCli bcsapi.Storage,
	getter common.GetterInterface, concurrency int, needSendKafka bool) *Producer {
	ctx, cancel := context.WithCancel(rootCtx)
	return &Producer{
		msgQueue:        msgQueue,
		cron:            cron,
		k8sStorageCli:   k8sStorageCli,
		mesosStorageCli: mesosStorageCli,
		ctx:             ctx,
		cancel:          cancel,
		resourceGetter:  getter,
		concurrency:     concurrency,
		needSendKafka:   needSendKafka,
	}
}

// ImportKafkaConn import kafka conn
func (p *Producer) ImportKafkaConn(conn kafka.KafkaInterface) {
	p.kafkaConn = conn
}

// Stop stop producer
func (p *Producer) Stop() {
	p.cron.Stop()
	if p.kafkaConn != nil {
		if err := p.kafkaConn.Stop(); err != nil {
			blog.Errorf("stop kafka conn err:%s", err.Error())
		}
	}
}

// Run run producer
func (p *Producer) Run() {
	defer func() {
		if r := recover(); r != nil {
			blog.Errorf("internal error: %v", p)
		}
	}()
	p.cron.Start()
}

// InitCronList get all cron func
func (p *Producer) InitCronList() error {
	minSpec := "0-59/1 * * * * "
	if _, err := p.cron.AddFunc(minSpec, func() {
		p.WorkloadProducer(types.DimensionMinute)
	}); err != nil {
		return err
	}

	tenMinSpec := "0-59/10 * * * * "
	if _, err := p.cron.AddFunc(tenMinSpec, func() {
		p.NamespaceProducer(types.DimensionMinute)
	}); err != nil {
		return err
	}
	if _, err := p.cron.AddFunc(tenMinSpec, func() {
		p.ClusterProducer(types.DimensionMinute)
	}); err != nil {
		return err
	}

	halfHourSpec := "0-59/30 * * * * "
	if _, err := p.cron.AddFunc(halfHourSpec, func() {
		p.PodAutoscalerProducer(types.DimensionMinute)
	}); err != nil {
		return err
	}

	hourSpec := "10 * * * * "
	if _, err := p.cron.AddFunc(hourSpec, func() {
		p.WorkloadProducer(types.DimensionHour)
	}); err != nil {
		return err
	}
	if _, err := p.cron.AddFunc(hourSpec, func() {
		p.PodAutoscalerProducer(types.DimensionHour)
	}); err != nil {
		return err
	}
	if _, err := p.cron.AddFunc(hourSpec, func() {
		p.NamespaceProducer(types.DimensionHour)
	}); err != nil {
		return err
	}
	if _, err := p.cron.AddFunc(hourSpec, func() {
		p.ClusterProducer(types.DimensionHour)
	}); err != nil {
		return err
	}

	daySpec := "30 0 * * *"
	if _, err := p.cron.AddFunc(daySpec, func() {
		p.WorkloadProducer(types.DimensionDay)
	}); err != nil {
		return err
	}
	if _, err := p.cron.AddFunc(daySpec, func() {
		p.PodAutoscalerProducer(types.DimensionDay)
	}); err != nil {
		return err
	}
	if _, err := p.cron.AddFunc(daySpec, func() {
		p.NamespaceProducer(types.DimensionDay)
	}); err != nil {
		return err
	}
	if _, err := p.cron.AddFunc(daySpec, func() {
		p.ClusterProducer(types.DimensionDay)
	}); err != nil {
		return err
	}
	if _, err := p.cron.AddFunc(daySpec, func() {
		p.ProjectProducer(types.DimensionDay)
	}); err != nil {
		return err
	}
	if _, err := p.cron.AddFunc(daySpec, func() {
		p.PublicProducer(types.DimensionDay)
	}); err != nil {
		return err
	}
	blog.Infof("init cron list")
	return nil
}

// PublicProducer is the function to produce public data job and send to message queue
func (p *Producer) PublicProducer(dimension string) {
	opts := types.JobCommonOpts{
		Dimension:   dimension,
		ObjectType:  types.PublicType,
		CurrentTime: utils.FormatTime(time.Now(), dimension),
	}
	err := p.SendJob(opts)
	if err != nil {
		blog.Errorf("send public job to msg queue error, opts: %v, err: %v", opts, err)
		return
	}
	blog.Infof("[producer] send public job success")
}

// ProjectProducer is the function to produce project data job and send to message queue
// It takes in the dimension parameter.
func (p *Producer) ProjectProducer(dimension string) {
	startTime := time.Now()
	var err error
	defer func() {
		prom.ReportProduceJobLatencyMetric(types.ProjectType, dimension, err, startTime)
	}()
	jobTime := utils.FormatTime(time.Now(), dimension)
	projectList, err := p.resourceGetter.GetProjectIDList(p.ctx)
	if err != nil || projectList == nil {
		blog.Errorf("get projectIDList error: %v", err)
		return
	}
	for _, project := range projectList {
		opts := types.JobCommonOpts{
			ProjectID:   project.ProjectID,
			ProjectCode: project.ProjectCode,
			BusinessID:  project.BusinessID,
			CurrentTime: jobTime,
			Dimension:   dimension,
			ObjectType:  types.ProjectType,
			Label:       project.Label,
		}
		err := p.SendJob(opts)
		if err != nil {
			blog.Errorf("send project job to msg queue error, opts: %v, err: %v", opts, err)
			return
		}
	}
	blog.Infof("[producer] send project job success, count: %d, jobTime:%v, startTime:%v, currentTime:%v, cost:%v",
		len(projectList), jobTime, startTime, time.Now(), time.Since(startTime))
}

// ClusterProducer is the function to produce cluster data job and send to message queue
// It takes in the dimension parameter.
func (p *Producer) ClusterProducer(dimension string) {
	startTime := time.Now()
	jobTime := utils.FormatTime(time.Now(), dimension)
	var err error
	defer func() {
		prom.ReportProduceJobLatencyMetric(types.ClusterType, dimension, err, startTime)
	}()

	clusterList, err := p.resourceGetter.GetClusterIDList(p.ctx)
	if err != nil || clusterList == nil {
		blog.Errorf("get clusterList error: %v", err)
		return
	}
	for _, cluster := range clusterList {
		opts := types.JobCommonOpts{
			ProjectID:   cluster.ProjectID,
			ProjectCode: cluster.ProjectCode,
			BusinessID:  cluster.BusinessID,
			ClusterID:   cluster.ClusterID,
			ClusterType: cluster.ClusterType,
			CurrentTime: jobTime,
			Dimension:   dimension,
			ObjectType:  types.ClusterType,
			Label:       cluster.Label,
			IsBKMonitor: cluster.IsBKMonitor,
		}
		err := p.SendJob(opts)
		if err != nil {
			blog.Errorf("send cluster job to msg queue error, opts: %v, err: %v", opts, err)
			return
		}
	}
	blog.Infof("[producer] send cluster job success, count: %d, jobTime:%v, startTime:%v, currentTime:%v, cost:%v",
		len(clusterList), jobTime, startTime, time.Now(), time.Since(startTime))
}

// NamespaceProducer is the function to produce namespace data job and send to message queue
func (p *Producer) NamespaceProducer(dimension string) {
	startTime := time.Now()
	jobTime := utils.FormatTime(time.Now(), dimension)
	var err error
	defer func() {
		prom.ReportProduceJobLatencyMetric(types.NamespaceType, dimension, err, startTime)
	}()

	namespaceList, err := p.resourceGetter.GetNamespaceList(p.ctx, p.k8sStorageCli, p.mesosStorageCli)
	if err != nil || namespaceList == nil {
		blog.Errorf("get namespace list error: %v", err)
		return
	}
	for _, namespace := range namespaceList {
		opts := types.JobCommonOpts{
			ClusterID:   namespace.ClusterID,
			ProjectID:   namespace.ProjectID,
			ProjectCode: namespace.ProjectCode,
			BusinessID:  namespace.BusinessID,
			ClusterType: namespace.ClusterType,
			Namespace:   namespace.Name,
			CurrentTime: jobTime,
			Dimension:   dimension,
			ObjectType:  types.NamespaceType,
			Label:       namespace.Label,
			IsBKMonitor: namespace.IsBKMonitor,
		}
		err := p.SendJob(opts)
		if err != nil {
			blog.Errorf("send namespace job to msg queue error, opts: %v, err: %v", opts, err)
			return
		}
	}
	blog.Infof("[producer] send all namespace job, count:%d, jobTime:%v, startTime:%v, "+
		"currentTime:%v, cost:%v", len(namespaceList), jobTime, startTime, time.Now(), time.Now().Sub(startTime)) // nolint
}

// WorkloadProducer is a function that produces workload data jobs and sends them to the message queue.
// It takes in the dimension parameter.
// It retrieves the list of clusters from the Cluster Manager and
// creates a channel to count the number of workloads sent to the Kafka topic.
// It creates a pool of goroutines to retrieve the workload data for each cluster in parallel.
// It logs the number of workloads sent to the Kafka topic and the time it took to send them.
func (p *Producer) WorkloadProducer(dimension string) {
	startTime := time.Now()
	jobTime := utils.FormatTime(time.Now(), dimension)
	var err error
	defer func() {
		prom.ReportProduceJobLatencyMetric(types.WorkloadType, dimension, err, startTime)
	}()

	clusterList, err := p.resourceGetter.GetClusterIDList(p.ctx)
	if err != nil || clusterList == nil {
		blog.Errorf("get clusterList error: %v", err)
		return
	}
	var totalWorkload int
	countCh := make(chan int, 200)
	go func() {
		for count := range countCh {
			totalWorkload += count
		}
	}()

	wg := sync.WaitGroup{}
	pool, err := ants.NewPool(p.concurrency)
	if err != nil {
		blog.Errorf("[producer] init new pool err:%v", err)
		return
	}
	// Retrieve the workload data for each cluster in parallel.
	blog.Infof("[producer] concurrency:%d", p.concurrency)
	defer pool.Release()
	for key := range clusterList {
		wg.Add(1)
		clusterMeta := clusterList[key]
		err := pool.Submit(func() {
			p.getSingleClusterWorkloadList(jobTime, dimension, countCh, clusterMeta)
			wg.Done()
		})
		if err != nil {
			blog.Errorf("submit task to ch pool err:%v", err)
		}
	}
	wg.Wait()
	time.Sleep(100 * time.Microsecond)
	close(countCh)
	// Log the number of workloads sent to the Kafka topic and the time it took to send them.
	blog.Infof("[producer] send all workload job, count:%d, jobTime:%v, startTime:%v, "+
		"currentTime:%v, cost:%v", totalWorkload, jobTime, startTime, time.Now(), time.Since(startTime))
}

// getSingleClusterWorkloadList get single cluster workload list
func (p *Producer) getSingleClusterWorkloadList(jobTime time.Time, dimension string, countCh chan int,
	clusterMeta *types.ClusterMeta) {
	workloadList := make([]*types.WorkloadMeta, 0)
	var err error
	defer func() {
		countCh <- len(workloadList)
	}()
	switch clusterMeta.ClusterType {
	case types.Kubernetes:
		namespaceList, err := p.resourceGetter.GetNamespaceListByCluster(p.ctx, clusterMeta, p.k8sStorageCli, // nolint
			p.mesosStorageCli)
		if err != nil {
			blog.Errorf("get workload list error: %v", err)
			return
		}
		if workloadList, err = p.resourceGetter.GetK8sWorkloadList(namespaceList, p.k8sStorageCli); err != nil {
			blog.Errorf("get workload list error: %v", err)
			return
		}
	case types.Mesos:
		if workloadList, err = p.resourceGetter.GetMesosWorkloadList(clusterMeta, p.mesosStorageCli); err != nil {
			blog.Errorf("get workload list error: %v", err)
			return
		}
	}
	for _, workload := range workloadList {
		opts := types.JobCommonOpts{
			ProjectID:    workload.ProjectID,
			ProjectCode:  workload.ProjectCode,
			BusinessID:   workload.BusinessID,
			ClusterID:    workload.ClusterID,
			ClusterType:  workload.ClusterType,
			Namespace:    workload.Namespace,
			WorkloadType: workload.ResourceType,
			WorkloadName: workload.Name,
			CurrentTime:  jobTime,
			Dimension:    dimension,
			ObjectType:   types.WorkloadType,
			Label:        workload.Label,
			IsBKMonitor:  workload.IsBKMonitor,
		}
		if err = p.SendJob(opts); err != nil {
			blog.Errorf("send workload job to msg queue error, opts: %v, err: %v", opts, err)
			return
		}
	}
	blog.Infof("[producer] send cluster[%s] workload job success, count: %d", clusterMeta.ClusterID,
		len(workloadList))
}

// PodAutoscalerProducer is the function to produce podAutoscaler data job and send to message queue
func (p *Producer) PodAutoscalerProducer(dimension string) {
	startTime := time.Now()
	jobTime := utils.FormatTime(time.Now(), dimension)
	var err error
	defer func() {
		prom.ReportProduceJobLatencyMetric(types.PodAutoscalerType, dimension, err, startTime)
	}()

	clusterList, err := p.resourceGetter.GetClusterIDList(p.ctx)
	if err != nil || clusterList == nil {
		blog.Errorf("get clusterList error: %v", err)
		return
	}
	var totalPodAutoscaler int
	countCh := make(chan int, 200)
	go func() {
		for count := range countCh {
			totalPodAutoscaler += count
		}
	}()

	wg := sync.WaitGroup{}
	pool, err := ants.NewPool(p.concurrency)
	if err != nil {
		blog.Errorf("[producer] init new pool err:%v", err)
		return
	}
	blog.Infof("[producer] concurrency:%d", p.concurrency)
	defer pool.Release()
	for key := range clusterList {
		wg.Add(1)
		clusterMeta := clusterList[key]
		err := pool.Submit(func() {
			p.getSingleClusterAutoscalerList(jobTime, dimension, countCh, clusterMeta)
			wg.Done()
		})
		if err != nil {
			blog.Errorf("submit task to ch pool err:%v", err)
		}
	}
	wg.Wait()
	time.Sleep(100 * time.Microsecond)
	close(countCh)
	blog.Infof("[producer] send all podAutoscaler job, count:%d, jobTime:%v, startTime:%v, "+
		"currentTime:%v, cost:%v", totalPodAutoscaler, jobTime, startTime, time.Now(), time.Since(startTime))
}

// getSingleClusterAutoscalerList get cluster pod autoscaler data list
// It takes in the jobTime, dimension, countCh, and clusterMeta parameters.
// It retrieves the namespace list for the cluster and then retrieves the pod autoscaler data for each namespace.
// It generates and sends an autoscaler job for each pod autoscaler.
// It logs the number of pod autoscalers sent to the Kafka topic.
func (p *Producer) getSingleClusterAutoscalerList(jobTime time.Time, dimension string, countCh chan int,
	clusterMeta *types.ClusterMeta) {
	hpaList := make([]*types.PodAutoscalerMeta, 0)
	gpaList := make([]*types.PodAutoscalerMeta, 0)
	defer func() {
		countCh <- len(hpaList)
		countCh <- len(gpaList)
	}()
	switch clusterMeta.ClusterType {
	case types.Kubernetes:
		// Get the namespace list for the cluster.
		namespaceList, err := p.resourceGetter.GetNamespaceListByCluster(p.ctx, clusterMeta,
			p.k8sStorageCli, p.mesosStorageCli)
		if err != nil {
			blog.Errorf("get workload list error: %v", err)
			return
		}
		// Get the pod autoscaler data for each namespace.
		if hpaList, err = p.resourceGetter.GetPodAutoscalerList(types.HPAType, namespaceList,
			p.k8sStorageCli); err != nil {
			blog.Errorf("get hpa list error: %v", err)
			return
		}
		if gpaList, err = p.resourceGetter.GetPodAutoscalerList(types.GPAType, namespaceList,
			p.k8sStorageCli); err != nil {
			blog.Errorf("get gpa list error: %v", err)
			return
		}
	case types.Mesos:
		return
	}
	// Generate and send an autoscaler job for each pod autoscaler.
	p.genAndSendAutoscalerJob(types.HPAType, dimension, jobTime, hpaList)
	p.genAndSendAutoscalerJob(types.GPAType, dimension, jobTime, gpaList)
	blog.Infof("[producer] send cluster[%s] podAutoscaler job success, count: %d", clusterMeta.ClusterID,
		len(hpaList)+len(gpaList))
}

// SendJob is the function to send data job to msg queue
// if it needs to send kafka, send to particular kafka topic
func (p *Producer) SendJob(opts types.JobCommonOpts) error {
	var err error
	defer func() {
		prom.ReportProduceJobTotalMetric(opts.ObjectType, opts.Dimension, err)
	}()
	dataJob := datajob.DataJob{Opts: opts}
	msg := &broker.Message{Header: map[string]string{
		"resourceType": types.DataJobQueue,
		"clusterId":    "dataManager",
	}}
	err = codec.EncJson(dataJob, &msg.Body)
	if err != nil {
		blog.Errorf("transfer dataJob to msg body error, dataJob: %v, error: %v", dataJob, err)
		return err
	}
	err = p.msgQueue.Publish(msg)
	if err != nil {
		blog.Errorf("send message error: %v", err)
		return err
	}
	if p.needSendKafka {
		var kafkaMsg []byte
		opts.Timestamp = opts.CurrentTime.Unix()
		err = codec.EncJson(opts, &kafkaMsg)
		if err != nil {
			blog.Errorf("transfer opts to []byte error, opts: %v, error: %s", opts, err.Error())
			return err
		}
		err := p.kafkaConn.PublishWithTopic(p.ctx, opts.ObjectType, 0, kafkaMsg)
		if err != nil {
			return fmt.Errorf("send message to kafka err:%s", err.Error())
		}
	}
	return nil
}

// genAndSendAutoscalerJob generate pod autoscaler calculate job
func (p *Producer) genAndSendAutoscalerJob(autoscalerType, dimension string, jobTime time.Time,
	list []*types.PodAutoscalerMeta) {
	for _, autoscaler := range list {
		opts := types.JobCommonOpts{
			ProjectID:         autoscaler.ProjectID,
			ProjectCode:       autoscaler.ProjectCode,
			BusinessID:        autoscaler.BusinessID,
			ClusterID:         autoscaler.ClusterID,
			ClusterType:       autoscaler.ClusterType,
			Namespace:         autoscaler.Namespace,
			WorkloadType:      autoscaler.TargetResourceType,
			WorkloadName:      autoscaler.TargetWorkloadName,
			PodAutoscalerName: autoscaler.PodAutoscaler,
			PodAutoscalerType: autoscalerType,
			CurrentTime:       jobTime,
			Dimension:         dimension,
			ObjectType:        types.PodAutoscalerType,
			Label:             autoscaler.Label,
			IsBKMonitor:       autoscaler.IsBKMonitor,
		}
		if err := p.SendJob(opts); err != nil {
			blog.Errorf("send gpa job to msg queue error, opts: %v, err: %v", opts, err)
			return
		}
	}
}
