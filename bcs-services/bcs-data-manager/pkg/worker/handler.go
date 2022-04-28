/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 *  Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 *  Licensed under the MIT License (the "License"); you may not use this file except
 *  in compliance with the License. You may obtain a copy of the License at
 *  http://opensource.org/licenses/MIT
 *  Unless required by applicable law or agreed to in writing, software distributed under
 *  the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 *  either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/msgqueue"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/bcsmonitor"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/cmanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/datajob"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/store"
)

// DataJobHandler handler for dataJob
type DataJobHandler struct {
	unSub         func()
	stopCtx       context.Context
	stopCancel    context.CancelFunc
	jobListCh     chan msgqueue.HandlerData
	filters       []msgqueue.Filter
	clients       HandleClients
	policyFactory datajob.PolicyFactoryInterface
	concurrency   int64
}

// HandlerOptions for DataJobHandler
type HandlerOptions struct {
	ChanQueueNum int64
}

// HandleClients handleClients type
type HandleClients struct {
	Store            store.Server
	BcsMonitorClient bcsmonitor.ClientInterface
	K8sStorageCli    bcsapi.Storage
	MesosStorageCli  bcsapi.Storage
	CmCli            *cmanager.ClusterManagerClient
}

// NewDataJobHandler create dataJob handler object
func NewDataJobHandler(opts HandlerOptions, client HandleClients, concurrency int64) *DataJobHandler {
	ctx, cancel := context.WithCancel(context.Background())
	factory := datajob.NewPolicyFactory(client.Store)
	factory.Init()
	return &DataJobHandler{
		stopCtx:       ctx,
		stopCancel:    cancel,
		jobListCh:     make(chan msgqueue.HandlerData, opts.ChanQueueNum),
		clients:       client,
		policyFactory: factory,
		concurrency:   concurrency,
	}
}

// Consume consume data
func (h *DataJobHandler) Consume(sub msgqueue.MessageQueue) error {
	unSub, err := sub.Subscribe(msgqueue.HandlerWrap("data-job-handler", h.HandleQueue), h.filters, common.DataJobQueue)
	if err != nil {
		blog.Errorf("subscribe err :%v", err)
		return fmt.Errorf("subscribe err :%v", err)
	}
	blog.Infof("subscribe success")
	h.unSub = func() {
		unSub.Unsubscribe()
	}
	go h.handleJob()
	return nil
}

// Stop stop handler
func (h *DataJobHandler) Stop() error {
	h.unSub()
	h.stopCancel()
	close(h.jobListCh)
	time.Sleep(time.Second * 3)

	return nil
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
	h.jobListCh <- *dataJobHandlerData
	return nil
}

func (h *DataJobHandler) handleJob() {
	chPool := make(chan struct{}, h.concurrency)
	var handleJobCount int64
	for job := range h.jobListCh {
		select {
		case <-h.stopCtx.Done():
			blog.Info("handleJob has been stopped")
			return
		default:
			handleJobCount++
		}
		chPool <- struct{}{}
		go func(job msgqueue.HandlerData) {
			h.handleOneJob(job)
			<-chPool
		}(job)
		if handleJobCount%1000 == 0 {
			blog.Infof("jobListChan length:%d", len(h.jobListCh))
			blog.Infof("handle job count: %d", handleJobCount)
		}
	}
}

func (h *DataJobHandler) handleOneJob(job msgqueue.HandlerData) {
	dataJob := &datajob.DataJob{}
	err := json.Unmarshal(job.Body, dataJob)
	if err != nil {
		blog.Errorf("unmarshal job error: %v", err)
		return
	}
	policy := h.policyFactory.GetPolicy(dataJob.Opts.ObjectType, dataJob.Opts.Dimension)
	dataJob.SetPolicy(policy)
	cmConn, err := h.clients.CmCli.GetClusterManagerConn()
	if err != nil {
		blog.Errorf("get cm conn error:%v", err)
		return
	}
	defer cmConn.Close()
	cliWithHeader := h.clients.CmCli.NewGrpcClientWithHeader(context.Background(), cmConn)
	client := common.NewClients(h.clients.BcsMonitorClient, h.clients.K8sStorageCli,
		h.clients.MesosStorageCli, cliWithHeader)
	dataJob.SetClient(client)
	dataJob.DoPolicy(h.stopCtx)
}
