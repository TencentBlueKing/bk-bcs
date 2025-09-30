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
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/msgqueue"
	"github.com/smallnest/chanx"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/bcsmonitor"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/datajob"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/prom"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/types"
)

// DataJobHandler handler for dataJob
type DataJobHandler struct {
	unSub                  func()
	stopCtx                context.Context
	stopCancel             context.CancelFunc
	jobChanList            chan chanx.UnboundedChan
	chanMap                sync.Map
	filters                []msgqueue.Filter
	clients                HandleClients
	policyFactory          datajob.PolicyFactoryInterface
	concurrency            int64
	ignoreBkMonitorCluster bool
}

// HandlerOptions for DataJobHandler
type HandlerOptions struct {
	ChanQueueNum           int64
	ignoreBkMonitorCluster bool
}

// HandleClients handleClients type
type HandleClients struct {
	Store            store.Server
	BcsMonitorClient bcsmonitor.ClientInterface
	K8sStorageCli    bcsapi.Storage
	MesosStorageCli  bcsapi.Storage
}

// NewDataJobHandler create dataJob handler object
func NewDataJobHandler(opts HandlerOptions, client HandleClients, concurrency int64) *DataJobHandler {
	ctx, cancel := context.WithCancel(context.Background())
	factory := datajob.NewPolicyFactory(client.Store)
	factory.Init()
	return &DataJobHandler{
		stopCtx:                ctx,
		stopCancel:             cancel,
		jobChanList:            make(chan chanx.UnboundedChan, opts.ChanQueueNum),
		chanMap:                sync.Map{},
		clients:                client,
		policyFactory:          factory,
		concurrency:            concurrency,
		ignoreBkMonitorCluster: opts.ignoreBkMonitorCluster,
	}
}

// Consume consume data
func (h *DataJobHandler) Consume(sub msgqueue.MessageQueue) error {
	unSub, err := sub.Subscribe(
		msgqueue.NewHandlerWrapper("data-job-handler", h.HandleQueue),
		h.filters,
		types.DataJobQueue,
	)
	if err != nil {
		blog.Errorf("subscribe err :%v", err)
		return fmt.Errorf("subscribe err :%v", err)
	}
	blog.Infof("subscribe success")
	h.unSub = func() {
		_ = unSub.Unsubscribe()
	}
	go h.handleJob()
	return nil
}

// Stop stop handler
func (h *DataJobHandler) Stop() error {
	h.unSub()
	close(h.jobChanList)
	h.chanMap.Range(func(key, value interface{}) bool {
		unboundedChan, ok := value.(chanx.UnboundedChan)
		if !ok {
			return true
		}
		close(unboundedChan.In)
		return true
	})
	return nil
}

// Done waiting for all job in channel finished
func (h *DataJobHandler) Done() {
	for {
		handleEnd := true
		h.chanMap.Range(func(key, value interface{}) bool {
			unboundedChan, ok := value.(chanx.UnboundedChan)
			if !ok {
				return true
			}
			if unboundedChan.Len() != 0 {
				handleEnd = false
				return false
			}
			return true
		})
		if handleEnd {
			break
		}
	}
	h.stopCancel()
}

// HandleQueue register queue for job callback
func (h *DataJobHandler) HandleQueue(ctx context.Context, data []byte) error {
	defer func() {
		if r := recover(); r != nil {
			blog.Errorf("data job handle panic: %v\n", r)
		}
	}()

	select {
	case <-ctx.Done():
		blog.Errorf("queue handler timeout ctx done.")
	case <-h.stopCtx.Done():
		blog.Errorf("data job handler has been closed.")
		return nil
	default:
	}
	dataJobHandlerData := &msgqueue.HandlerData{}
	err := json.Unmarshal(data, dataJobHandlerData)
	if err != nil {
		blog.Errorf("Unmarshal handler data failed: %v", err)
		return err
	}
	dataJob := &datajob.DataJob{}
	err = json.Unmarshal(dataJobHandlerData.Body, dataJob)
	if err != nil {
		blog.Errorf("unmarshal job error: %v", err)
		return fmt.Errorf("unmarshal job error: %v", err)
	}
	switch dataJob.Opts.ObjectType {
	case types.ProjectType, types.PublicType:
		if _, ok := h.chanMap.Load("public"); !ok {
			publicChan := chanx.NewUnboundedChan(100)
			h.jobChanList <- *publicChan
			h.chanMap.Store("public", *publicChan)
			blog.Infof("[handler] add public chan")
		}
		publicCh, _ := h.chanMap.Load("public")
		publicChan, ok := publicCh.(chanx.UnboundedChan)
		if !ok {
			blog.Errorf("trans publicChan to chanx.UnboundedChan error")
			return fmt.Errorf("trans publicChan to chanx.UnboundedChan error")
		}
		publicChan.In <- *dataJob
	default:
		if _, ok := h.chanMap.Load(dataJob.Opts.ClusterID); !ok {
			clusterChan := chanx.NewUnboundedChan(100)
			h.jobChanList <- *clusterChan
			h.chanMap.Store(dataJob.Opts.ClusterID, *clusterChan)
			blog.Infof("[handler] add cluster chan:%s", dataJob.Opts.ClusterID)
		}
		clusterCh, _ := h.chanMap.Load(dataJob.Opts.ClusterID)
		clusterChan, ok := clusterCh.(chanx.UnboundedChan)
		if !ok {
			blog.Errorf("trans clusterChan to chanx.UnboundedChan error")
			return fmt.Errorf("trans clusterChan to chanx.UnboundedChan error")
		}
		clusterChan.In <- *dataJob
	}
	return nil
}

func (h *DataJobHandler) handleJob() {
	var clusterChanCount int
	for clusterChan := range h.jobChanList {
		select {
		case <-h.stopCtx.Done():
			blog.Info("handleJob has been stopped")
			return
		default:
			clusterChanCount++
		}
		go h.handleOneChan(clusterChan)
		prom.ReportConsumeConcurrency(clusterChanCount)
	}
}

func (h *DataJobHandler) handleOneChan(jobChan chanx.UnboundedChan) {
	for job := range jobChan.Out {
		select {
		case <-h.stopCtx.Done():
			blog.Info("handleJob has been stopped")
			return
		default:

		}
		value, ok := job.(datajob.DataJob)
		if !ok {
			continue
		}
		prom.ReportWaitingJobCount(value.Opts.ClusterID, jobChan.Len())
		if value.Opts.ObjectType == types.NamespaceType && value.Opts.Dimension == types.DimensionHour {
			newValue := value
			newValue.Opts.Dimension = types.GetWorkloadRequestType
			newValue.Opts.ObjectType = types.WorkloadType
			blog.Infof("[handler] create one getWorkloadRequest type job")
			h.handleOneJob(newValue)
		}
		if value.Opts.IsBKMonitor && h.ignoreBkMonitorCluster {
			continue
		}
		h.handleOneJob(value)
	}
}

func (h *DataJobHandler) handleOneJob(job datajob.DataJob) {
	start := time.Now()
	var err error
	defer func() {
		prom.ReportConsumeJobMetric(job.Opts.ObjectType, job.Opts.Dimension, err, start)
		prom.ReportJobMetric(job.Opts.ObjectType, job.Opts.Dimension, err, job.Opts.CurrentTime)
	}()
	policy := h.policyFactory.GetPolicy(job.Opts.ObjectType, job.Opts.Dimension)
	job.SetPolicy(policy)
	client := types.NewClients(h.clients.BcsMonitorClient, h.clients.K8sStorageCli,
		h.clients.MesosStorageCli)
	job.SetClient(client)
	job.DoPolicy(h.stopCtx)
}
