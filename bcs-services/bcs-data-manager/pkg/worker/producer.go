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
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/codec"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/msgqueue"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/cmanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/datajob"
	"github.com/micro/go-micro/v2/broker"
	"github.com/robfig/cron/v3"
)

// Producer produce data job
type Producer struct {
	msgQueue       msgqueue.MessageQueue
	cron           *cron.Cron
	CMClient       *cmanager.ClusterManagerClient
	storageCli     bcsapi.Storage
	ctx            context.Context
	cancel         context.CancelFunc
	resourceGetter common.GetterInterface
}

// NewProducer new producer
func NewProducer(msgQueue msgqueue.MessageQueue, cron *cron.Cron, cmClient *cmanager.ClusterManagerClient,
	storeCli bcsapi.Storage, rootCtx context.Context,
	getter common.GetterInterface) *Producer {
	ctx, cancel := context.WithCancel(rootCtx)
	return &Producer{
		msgQueue:       msgQueue,
		cron:           cron,
		CMClient:       cmClient,
		storageCli:     storeCli,
		ctx:            ctx,
		cancel:         cancel,
		resourceGetter: getter,
	}
}

// Stop stop producer
func (p *Producer) Stop() {
	p.cron.Stop()
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
		p.WorkloadProducer(common.DimensionMinute)
	}); err != nil {
		return err
	}

	tenMinSpec := "0-59/10 * * * * "
	if _, err := p.cron.AddFunc(tenMinSpec, func() {
		p.NamespaceProducer(common.DimensionMinute)
	}); err != nil {
		return err
	}
	if _, err := p.cron.AddFunc(tenMinSpec, func() {
		p.ClusterProducer(common.DimensionMinute)
	}); err != nil {
		return err
	}

	hourSpec := "10 * * * * "
	if _, err := p.cron.AddFunc(hourSpec, func() {
		p.WorkloadProducer(common.DimensionHour)
	}); err != nil {
		return err
	}
	if _, err := p.cron.AddFunc(hourSpec, func() {
		p.NamespaceProducer(common.DimensionHour)
	}); err != nil {
		return err
	}
	if _, err := p.cron.AddFunc(hourSpec, func() {
		p.ClusterProducer(common.DimensionHour)
	}); err != nil {
		return err
	}

	daySpec := "30 0 * * *"
	if _, err := p.cron.AddFunc(daySpec, func() {
		p.WorkloadProducer(common.DimensionDay)
	}); err != nil {
		return err
	}
	if _, err := p.cron.AddFunc(daySpec, func() {
		p.NamespaceProducer(common.DimensionDay)
	}); err != nil {
		return err
	}
	if _, err := p.cron.AddFunc(daySpec, func() {
		p.ClusterProducer(common.DimensionDay)
	}); err != nil {
		return err
	}
	if _, err := p.cron.AddFunc(daySpec, func() {
		p.ProjectProducer(common.DimensionDay)
	}); err != nil {
		return err
	}
	if _, err := p.cron.AddFunc(daySpec, func() {
		p.PublicProducer(common.DimensionDay)
	}); err != nil {
		return err
	}
	blog.Infof("init cron list")
	return nil
}

// PublicProducer is the function to produce public data job and send to message queue
func (p *Producer) PublicProducer(dimension string) {
	opts := common.JobCommonOpts{
		Dimension:   dimension,
		ObjectType:  common.PublicType,
		CurrentTime: common.FormatTime(time.Now(), dimension),
	}
	err := p.SendJob(opts)
	if err != nil {
		blog.Errorf("send public job to msg queue error, opts: %v, err: %v", opts, err)
		return
	}
	blog.Infof("[producer] send public job success")
}

// ProjectProducer is the function to produce project data job and send to message queue
func (p *Producer) ProjectProducer(dimension string) {
	cmConn, err := p.CMClient.GetClusterManagerConn()
	if err != nil {
		blog.Errorf("get cm conn error:%v", err)
		return
	}
	defer cmConn.Close()
	cliWithHeader := p.CMClient.NewGrpcClientWithHeader(p.ctx, cmConn)
	projectList, err := p.resourceGetter.GetProjectIDList(cliWithHeader.Ctx, cliWithHeader.Cli)
	if err != nil || projectList == nil {
		blog.Errorf("get projectIDList error: %v", err)
		return
	}
	for _, project := range projectList {
		opts := common.JobCommonOpts{
			ProjectID:   project,
			CurrentTime: common.FormatTime(time.Now(), dimension),
			Dimension:   dimension,
			ObjectType:  common.ProjectType,
		}
		err := p.SendJob(opts)
		if err != nil {
			blog.Errorf("send project job to msg queue error, opts: %v, err: %v", opts, err)
			return
		}
	}
	blog.Infof("[producer] send project job success, count: %d", len(projectList))
}

// ClusterProducer is the function to produce cluster data job and send to message queue
func (p *Producer) ClusterProducer(dimension string) {
	cmConn, err := p.CMClient.GetClusterManagerConn()
	if err != nil {
		blog.Errorf("get cm conn error:%v", err)
		return
	}
	defer cmConn.Close()
	cliWithHeader := p.CMClient.NewGrpcClientWithHeader(p.ctx, cmConn)
	clusterList, err := p.resourceGetter.GetClusterIDList(cliWithHeader.Ctx, cliWithHeader.Cli)
	if err != nil || clusterList == nil {
		blog.Errorf("get clusterList error: %v", err)
		return
	}
	for _, cluster := range clusterList {
		opts := common.JobCommonOpts{
			ProjectID:   cluster.ProjectID,
			ClusterID:   cluster.ClusterID,
			ClusterType: cluster.ClusterType,
			CurrentTime: common.FormatTime(time.Now(), dimension),
			Dimension:   dimension,
			ObjectType:  common.ClusterType,
		}
		err := p.SendJob(opts)
		if err != nil {
			blog.Errorf("send cluster job to msg queue error, opts: %v, err: %v", opts, err)
			return
		}
	}
	blog.Infof("[producer] send cluster job success, count: %d", len(clusterList))
}

// NamespaceProducer is the function to produce namespace data job and send to message queue
func (p *Producer) NamespaceProducer(dimension string) {
	cmConn, err := p.CMClient.GetClusterManagerConn()
	if err != nil {
		blog.Errorf("get cm conn error:%v", err)
		return
	}
	defer cmConn.Close()
	cliWithHeader := p.CMClient.NewGrpcClientWithHeader(p.ctx, cmConn)
	namespaceList, err := p.resourceGetter.GetNamespaceList(cliWithHeader.Ctx, cliWithHeader.Cli, p.storageCli)
	if err != nil || namespaceList == nil {
		blog.Errorf("get namespace list error: %v", err)
		return
	}
	for _, namespace := range namespaceList {
		opts := common.JobCommonOpts{
			ClusterID:   namespace.ClusterID,
			ProjectID:   namespace.ProjectID,
			ClusterType: namespace.ClusterType,
			Namespace:   namespace.Name,
			CurrentTime: common.FormatTime(time.Now(), dimension),
			Dimension:   dimension,
			ObjectType:  common.NamespaceType,
		}
		err := p.SendJob(opts)
		if err != nil {
			blog.Errorf("send namespace job to msg queue error, opts: %v, err: %v", opts, err)
			return
		}
	}
	blog.Infof("[producer] send namespace job success, count: %d", len(namespaceList))
}

// WorkloadProducer is the function to produce workload data job and send to message queue
func (p *Producer) WorkloadProducer(dimension string) {
	cmConn, err := p.CMClient.GetClusterManagerConn()
	if err != nil {
		blog.Errorf("get cm conn error:%v", err)
		return
	}
	defer cmConn.Close()
	cliWithHeader := p.CMClient.NewGrpcClientWithHeader(p.ctx, cmConn)
	workloadList, err := p.resourceGetter.GetWorkloadList(cliWithHeader.Ctx, cliWithHeader.Cli, p.storageCli)
	if err != nil || workloadList == nil {
		blog.Errorf("get workload list error: %v", err)
		return
	}
	for _, workload := range workloadList {
		opts := common.JobCommonOpts{
			ProjectID:    workload.ProjectID,
			ClusterID:    workload.ClusterID,
			ClusterType:  workload.ClusterType,
			Namespace:    workload.Namespace,
			WorkloadType: workload.ResourceType,
			Name:         workload.Name,
			CurrentTime:  common.FormatTime(time.Now(), dimension),
			Dimension:    dimension,
			ObjectType:   common.WorkloadType,
		}
		err := p.SendJob(opts)
		if err != nil {
			blog.Errorf("send workload job to msg queue error, opts: %v, err: %v", opts, err)
			return
		}
	}
	blog.Infof("[producer] send workload job success, count: %d", len(workloadList))
}

// SendJob is the function to send data job to msg queue
func (p *Producer) SendJob(opts common.JobCommonOpts) error {
	dataJob := datajob.DataJob{Opts: opts}
	msg := &broker.Message{Header: map[string]string{
		"resourceType": common.DataJobQueue,
		"clusterId":    "dataManager",
	}}
	err := codec.EncJson(dataJob, &msg.Body)
	if err != nil {
		blog.Errorf("transfer dataJob to msg body error, dataJob: %v, error: %v", dataJob, err)
		return err
	}
	err = p.msgQueue.Publish(msg)
	if err != nil {
		blog.Errorf("send message error: %v", err)
		return err
	}
	return nil
}
