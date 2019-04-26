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

package manager

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"bk-bcs/bcs-common/common/blog"
	commtype "bk-bcs/bcs-common/common/types"
	"bk-bcs/bcs-mesos/bcs-check/bcscheck/config"
	"bk-bcs/bcs-mesos/bcs-check/bcscheck/report"
	"bk-bcs/bcs-mesos/bcs-check/bcscheck/report/scheReporter"
	"bk-bcs/bcs-mesos/bcs-check/bcscheck/types"

	"golang.org/x/net/context"
)

const (
	//MaxDataQueueLength
	MaxDataQueueLength = 1024
)

type manager struct {
	lock sync.RWMutex

	conf config.HealthCheckConfig

	dataQueue chan *types.HealthSyncData

	healthChecks map[string]*healthCheckManager

	reporters map[types.HealthReporterMode]report.Reporter

	reportDataQueue chan *types.HealthCheck

	cxt    context.Context
	cancel context.CancelFunc
}

func NewManager(rootCxt context.Context, conf config.HealthCheckConfig) Manager {
	cxt, cancel := context.WithCancel(rootCxt)

	m := &manager{
		conf:            conf,
		cxt:             cxt,
		cancel:          cancel,
		dataQueue:       make(chan *types.HealthSyncData, MaxDataQueueLength),
		reportDataQueue: make(chan *types.HealthCheck, MaxDataQueueLength),
		healthChecks:    make(map[string]*healthCheckManager),
	}

	return m
}

func (m *manager) initReporter() {
	m.reporters = make(map[types.HealthReporterMode]report.Reporter)

	//init scheduler reporter
	reporter := scheReporter.NewScheReporter(m.cxt, m.conf)
	m.reporters[types.HealthReporterModeScheduler] = reporter

}

func (m *manager) Run() {
	go m.handleDataQueue()
	go m.loopReportDatas()
	go m.runReporters()
}

func (m *manager) Stop() {
	m.cancel()
}

func (m *manager) runReporters() {
	m.initReporter()

	for _, reporter := range m.reporters {
		go reporter.Run()
	}
}

func (m *manager) Sync(data *types.HealthSyncData) error {
	m.dataQueue <- data

	return nil
}

func (m *manager) syncReporterData(data *types.HealthCheck) error {
	m.reportDataQueue <- data

	return nil
}

func (m *manager) handleDataQueue() {
	tick := time.NewTicker(time.Second * 60)

	var err error

	for {

		select {
		case <-tick.C:
			blog.V(3).Info("Manager handle data queue")

		case <-m.cxt.Done():
			blog.Warn("Manager stop handle data queue")
			return

		case data := <-m.dataQueue:
			blog.V(3).Infof("Manager handleDataQueue action %s checkId %s", data.Action, data.HealthCheck.ID)

			err = m.handleHealthSyncData(data)
			if err != nil {
				blog.Error("Manager handleDataQueue error %s", err.Error())
			}

		}

	}
}

func (m *manager) handleHealthSyncData(data *types.HealthSyncData) error {
	var err error

	switch data.Action {
	// if add
	case types.SyncDataActionAdd:
		err = m.add(data.HealthCheck)

	// if delete
	case types.SyncDataActionDelete:
		err = m.delete(data.HealthCheck)

	// if update
	case types.SyncDataActionUpdate:
		err = m.update(data.HealthCheck)

	default:
		err = fmt.Errorf("HealthSyncData action %s is invalid", data.Action)
	}

	return err
}

func (m *manager) add(data *types.HealthCheck) error {
	by, _ := json.Marshal(data)
	blog.Info("Manager Add HealthCheck %s operation %s,for details: %s", data.ID, string(data.Operation), string(by))

	m.lock.Lock()
	defer m.lock.Unlock()

	_, ok := m.healthChecks[data.ID]
	if ok {
		return fmt.Errorf("health check %s exist", data.ID)
	}

	healthCheck := newHealthCheckManager(m.cxt, m, *data)
	healthCheck.start()

	m.healthChecks[data.ID] = healthCheck

	return nil
}

func (m *manager) delete(data *types.HealthCheck) error {
	blog.Info("Manager Delete HealthCheck %s operation %s", data.ID, string(data.Operation))

	m.lock.Lock()
	defer m.lock.Unlock()

	healthCheck, ok := m.healthChecks[data.ID]
	if !ok {
		blog.Warn("healthCheck %s not exist", data.ID)
	}

	delete(m.healthChecks, data.ID)

	healthCheck.stop()
	healthCheck = nil

	return nil
}

func (m *manager) update(data *types.HealthCheck) error {
	by, _ := json.Marshal(data)
	blog.Info("Manager Update HealthCheck %s operation %s,for details: %s", data.ID, string(data.Operation), string(by))

	m.lock.Lock()
	defer m.lock.Unlock()

	healthCheck, ok := m.healthChecks[data.ID]
	if !ok {
		return fmt.Errorf("health check %s not exist", data.ID)
	}

	err := healthCheck.update(*data)
	if err != nil {
		blog.Errorf("manager healthCheck update %s error %s", data.ID, err.Error())
	}

	return nil
}

func (m *manager) report(data *types.HealthCheck) error {

	switch data.TaskMode {
	case types.HealthTaskModeMesos:
		data.ReporterMode = types.HealthReporterModeScheduler

	default:
		err := fmt.Errorf("data %s TaskMode %s is invalid", data.ID, data.TaskMode)
		return err
	}

	reporter, ok := m.reporters[data.ReporterMode]
	if !ok {
		return fmt.Errorf("ReporterMode %s not found", data.ReporterMode)
	}

	result := commtype.HealthCheckResult{
		ID:      data.OriginID,
		Type:    data.Type,
		Message: data.Message,
	}

	if data.Status == types.HealthCheckStatusRunning {
		result.Status = true
	} else {
		result.Status = false
	}

	switch data.Type {
	// if http
	case commtype.BcsHealthCheckType_HTTP, commtype.BcsHealthCheckType_REMOTEHTTP:
		httpd := commtype.HttpHealthCheck{
			Port:   data.Http.Port,
			Scheme: data.Http.Scheme,
			Path:   data.Http.Path,
		}

		result.Http = &httpd

		// if tcp
	case commtype.BcsHealthCheckType_TCP, commtype.BcsHealthCheckType_REMOTETCP:
		tcpd := commtype.TcpHealthCheck{
			Port: data.Tcp.Port,
		}

		result.Tcp = &tcpd

	default:
		return fmt.Errorf("HealthCheck ID %s type %s is invalid", data.ID, string(data.Type))
	}

	return reporter.Sync(&result)
}

func (m *manager) loopReportDatas() {

	var err error

	for {

		select {
		case <-m.cxt.Done():
			blog.Warn("Manager stop loopReportDatas")
			return

		case data := <-m.reportDataQueue:
			err = m.report(data)
			if err != nil {
				blog.Errorf("report data %s error %s", data.ID, err.Error())
			}

		}

	}
}
