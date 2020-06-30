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
	"fmt"
	"net"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	commtype "github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-check/bcscheck/types"

	"golang.org/x/net/context"
)

const (
	//FailedRetryTimes
	FailedRetryTimes           = 3
	FailedRetryIntervalSeconds = 1
)

type healthCheckManager struct {
	lock sync.RWMutex

	id           string
	taskMode     types.HealthTaskMode
	reporterMode types.HealthReporterMode

	healthCheck *types.HealthCheck

	operation types.HealthCheckOperation // operation, like running, stopped

	status     types.HealthCheckStatus //status, like running, failed
	lastStatus types.HealthCheckStatus

	running    bool
	isReporter bool
	isCheck    bool

	httpCli *HttpClient

	manager *manager

	cxt    context.Context
	cancel context.CancelFunc
}

func newHealthCheckManager(rootCxt context.Context, m *manager, healthCheck types.HealthCheck) *healthCheckManager {
	cxt, cancel := context.WithCancel(rootCxt)

	h := &healthCheckManager{
		id:           healthCheck.ID,
		taskMode:     healthCheck.TaskMode,
		reporterMode: healthCheck.ReporterMode,
		healthCheck:  &healthCheck,
		status:       types.HealthCheckStatusFailed,
		lastStatus:   types.HealthCheckStatusFailed,
		operation:    healthCheck.Operation,
		manager:      m,
		cxt:          cxt,
		cancel:       cancel,
	}

	if h.operation == types.HealthCheckOperationRunning {
		h.running = true
	} else {
		h.running = false
	}

	return h
}

func (h *healthCheckManager) getOuterHealthCheck() types.HealthCheck {
	h.lock.Lock()
	health := *h.healthCheck
	health.Status = h.status
	h.lock.Unlock()

	return health
}

func (h *healthCheckManager) start() {
	go h.runHealthCheck()
	go h.reportHealthCheckStatus()
}

func (h *healthCheckManager) stop() {
	if h.running {
		h.cancel()
	}
}

func (h *healthCheckManager) update(healthCheck types.HealthCheck) error {
	//stop health check endpoint
	h.stop()

	h.lock.Lock()

	h.healthCheck = &healthCheck
	h.operation = healthCheck.Operation

	if h.operation == types.HealthCheckOperationRunning {
		h.running = true
	} else {
		h.running = false
	}

	h.lock.Unlock()

	h.start()
	return nil
}

func (h *healthCheckManager) reportHealthCheckStatus() {

	if !h.running {
		blog.Info("reportHealthCheckStatus healthCheckManager %s operation %s", h.id, types.HealthCheckOperationStopped)
		return
	}

	tick := time.NewTicker(time.Second * 180)

	for {

		select {
		case <-h.cxt.Done():
			blog.Info("healthCheckManager %s stop reportHealthCheckStatus", h.id)
			return

		case <-tick.C:
			health := h.getOuterHealthCheck()
			err := h.manager.syncReporterData(&health)
			if err != nil {
				blog.Errorf("sync report data error %s", err.Error())
			}

		default:
			if h.isCheck && (!h.isReporter || h.status != h.lastStatus) {
				h.isReporter = true
				h.lastStatus = h.status
				health := h.getOuterHealthCheck()
				err := h.manager.syncReporterData(&health)
				if err != nil {
					blog.Errorf("sync report data error %s", err.Error())
				}
			}
		}

		time.Sleep(time.Second)

	}
}

func (h *healthCheckManager) runHealthCheck() {

	if !h.running {
		blog.Info("runHealthCheck healthCheckManager %s operation %s", h.id, types.HealthCheckOperationStopped)
		return
	}

	_, err := h.checkoutHealthCheckParas()
	if err != nil {
		blog.Errorf(err.Error())
		return
	}

	blog.Info("healthCheckManager %s health check running", h.id)

	health := h.getInnerHealthCheck()

	if health.GracePeriodSeconds > 0 {
		time.Sleep(time.Second * time.Duration(health.GracePeriodSeconds))
	}

	intervalS := health.IntervalSeconds
	duration := time.Second * time.Duration(intervalS)

	tick := time.NewTicker(duration)

	for {
		select {
		case <-h.cxt.Done():
			blog.Info("healthCheckManager %s stop runHealthCheck", h.id)
			return

		case <-tick.C:
			go h.checkEndpoint()

		}
	}

}

func (h *healthCheckManager) checkoutHealthCheckParas() (bool, error) {
	health := h.healthCheck

	if health.IntervalSeconds <= health.TimeoutSeconds {
		return false, fmt.Errorf("checkoutHealthCheckParas ID %s IntervalSeconds %d must be "+
			"less than TimeoutSeconds %d", health.ID, health.IntervalSeconds, health.TimeoutSeconds)
	}

	var ipAddr string
	var port int32

	switch health.Type {
	case commtype.BcsHealthCheckType_HTTP, commtype.BcsHealthCheckType_REMOTEHTTP:
		if health.Http == nil {
			return false, fmt.Errorf("checkoutHealthCheckParas ID %s Http can't be nil", health.ID)
		}

		ipAddr = health.Http.Ip
		port = health.Http.Port

		if health.Http.Scheme != "http" && health.Http.Scheme != "https" {
			return false, fmt.Errorf("checkoutHealthCheckParas ID %s http scheme %s is invalid", health.ID, health.Http.Scheme)
		}

	case commtype.BcsHealthCheckType_TCP, commtype.BcsHealthCheckType_REMOTETCP:
		if health.Tcp == nil {
			return false, fmt.Errorf("checkoutHealthCheckParas ID %s Tcp can't be nil", health.ID)
		}

		ipAddr = health.Tcp.Ip
		port = health.Tcp.Port

	default:
		return false, fmt.Errorf("checkoutHealthCheckParas ID %s type %s is invalid", health.ID, health.Type)
	}

	reg, _ := regexp.Compile(`((2[0-4]\d|25[0-5]|[01]?\d\d?)\.){3}(2[0-4]\d|25[0-5]|[01]?\d\d?)`)

	if !reg.MatchString(ipAddr) {
		return false, fmt.Errorf("checkoutHealthCheckParas ID %s ipAddr %s is invalid", health.ID, ipAddr)
	}

	if port <= 0 || port > 65535 {
		return false, fmt.Errorf("checkoutHealthCheckParas ID %s port %d is invalid", health.ID, port)
	}

	return true, nil
}

func (h *healthCheckManager) getInnerHealthCheck() *types.HealthCheck {
	h.lock.Lock()
	health := h.healthCheck
	h.lock.Unlock()

	return health
}

func (h *healthCheckManager) checkEndpoint() {
	health := h.getInnerHealthCheck()

	var ok bool
	var message string

	// if check failed,then retry 2 times
	for i := 0; i < FailedRetryTimes; i++ {
		switch health.Type {
		// if http
		case commtype.BcsHealthCheckType_HTTP, commtype.BcsHealthCheckType_REMOTEHTTP:
			ok, message = h.checkHttp()

			// if tcp
		case commtype.BcsHealthCheckType_TCP, commtype.BcsHealthCheckType_REMOTETCP:
			ok, message = h.checkTcp()

		default:
			blog.Error("checkEndpoint ID %s type %s is invalid", health.ID, string(health.Type))
			return
		}

		//if ok ,then think the endpoint is ok
		if ok {
			break
		}

		time.Sleep(time.Second * FailedRetryIntervalSeconds)
	}

	if ok {
		h.status = types.HealthCheckStatusRunning
	} else {
		h.status = types.HealthCheckStatusFailed
	}

	h.isCheck = true
	health.Message = message

}

func (h *healthCheckManager) initHttpCli() {
	h.httpCli = NewHttpClient()
	health := h.getInnerHealthCheck()

	duration := time.Second * time.Duration(health.TimeoutSeconds)
	h.httpCli.SetTimeOut(duration)

	if health.Http != nil && health.Http.Headers != nil {
		for k, v := range health.Http.Headers {
			h.httpCli.SetHeader(k, v)
		}
	}
}

func (h *healthCheckManager) checkHttp() (bool, string) {
	if h.httpCli == nil {
		h.initHttpCli()
	}

	httpd := h.getInnerHealthCheck().Http

	httpP := strings.TrimLeft(httpd.Path, "/")

	url := fmt.Sprintf("%s://%s:%d/%s", httpd.Scheme, httpd.Ip, httpd.Port, httpP)

	message := fmt.Sprintf("check endpoint %s:%d %s ok", httpd.Ip, httpd.Port, httpd.Scheme)

	code, _, err := h.httpCli.GET(url, nil, nil)
	if err != nil {
		blog.Errorf("check endpoint %s:%d error %s", httpd.Ip, httpd.Port, err.Error())
		message = fmt.Sprintf("check endpoint %s:%d failed: %s", httpd.Ip, httpd.Port, err.Error())

		return false, message
	}

	if code >= 200 && code <= 399 {
		blog.V(3).Infof("check endpoint %s:%d %s success", httpd.Ip, httpd.Port, httpd.Scheme)
		return true, message
	}

	blog.Errorf("check endpoint %s:%d resp httpcode %d", httpd.Ip, httpd.Port, code)
	message = fmt.Sprintf("check endpoint %s:%d resp httpcode %d", httpd.Ip, httpd.Port, code)

	return false, message
}

func (h *healthCheckManager) checkTcp() (bool, string) {
	health := h.getInnerHealthCheck()
	tcpd := health.Tcp

	message := fmt.Sprintf("check endpoint %s:%d tcp ok", tcpd.Ip, tcpd.Port)

	duration := time.Second * time.Duration(health.TimeoutSeconds)
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", tcpd.Ip, tcpd.Port), duration)
	if err != nil {
		blog.Errorf("check endpoint %s:%d error %s", tcpd.Ip, tcpd.Port, err.Error())
		message = fmt.Sprintf("check endpoint %s:%d failed: %s", tcpd.Ip, tcpd.Port, err.Error())

		return false, message
	}

	err = conn.Close()
	if err != nil {
		blog.Errorf("close tcp conn error %s", err.Error())
	}

	blog.V(3).Infof("check endpoint %s:%d tcp success", tcpd.Ip, tcpd.Port)

	return true, message
}
