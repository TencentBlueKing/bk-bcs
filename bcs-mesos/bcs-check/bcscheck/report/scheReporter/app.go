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

package scheReporter

import (
	"encoding/json"
	"fmt"
	"runtime"
	"strconv"
	"time"

	rd "bk-bcs/bcs-common/common/RegisterDiscover"
	"bk-bcs/bcs-common/common/blog"
	"bk-bcs/bcs-common/common/http/httpclient"
	commtype "bk-bcs/bcs-common/common/types"
	"bk-bcs/bcs-mesos/bcs-check/bcscheck/config"
	"bk-bcs/bcs-mesos/bcs-check/bcscheck/report"

	"golang.org/x/net/context"
)

const (
	MaxDataQueueLength = 1024
)

type scheReporter struct {
	conf config.HealthCheckConfig

	currScheduler string

	dataQueue chan *commtype.HealthCheckResult

	cli *httpclient.HttpClient

	cxt    context.Context
	cancel context.CancelFunc
}

func NewScheReporter(rootCxt context.Context, conf config.HealthCheckConfig) report.Reporter {
	cxt, cancel := context.WithCancel(rootCxt)

	r := &scheReporter{
		cxt:       cxt,
		cancel:    cancel,
		conf:      conf,
		dataQueue: make(chan *commtype.HealthCheckResult, MaxDataQueueLength),
	}

	r.init()

	return r
}

func (r *scheReporter) init() {
	r.initCli()
}

func (r *scheReporter) initCli() {
	r.cli = httpclient.NewHttpClient()

	if r.conf.ClientCert.IsSSL {
		r.cli.SetTlsVerity(r.conf.ClientCert.CAFile, r.conf.ClientCert.CertFile, r.conf.ClientCert.KeyFile,
			r.conf.ClientCert.CertPasswd)
	}

	r.cli.SetHeader("Content-Type", "application/json")
	r.cli.SetHeader("Accept", "application/json")
}

func (r *scheReporter) Run() {
	go r.handleDataQueue()
	go r.discvScheduler()
}

func (r *scheReporter) Stop() {
	r.cancel()
}

func (r *scheReporter) Sync(data *commtype.HealthCheckResult) error {
	blog.V(3).Infof("scheReporter rev data %s status %t", data.ID, data.Status)

	r.dataQueue <- data

	return nil
}

func (r *scheReporter) handleDataQueue() {

	tick := time.NewTicker(time.Second * 10)

	var err error

	for {

		select {
		case <-tick.C:
			blog.V(3).Info("TaskManager handle data queue")

		case <-r.cxt.Done():
			blog.Warn("TaskManager stop handle data queue")
			return

		case data := <-r.dataQueue:
			err = r.handleData(data)
			if err != nil {
				blog.Error("scheReporter handleData ID %s error %s", data.ID, err.Error())
			} else {
				blog.Infof("scheReporter handleData ID %s status %t success", data.ID, data.Status)
			}

		}

	}
}

func (r *scheReporter) handleData(data *commtype.HealthCheckResult) error {
	uri := fmt.Sprintf("/healthcheck")
	by, _ := json.Marshal(data)

	_, err := r.requestSchedulerV1("POST", uri, by)
	return err
}

func (r *scheReporter) discvScheduler() {
	blog.Infof("healtchcheck begin to discover scheduler from (%s), curr goroutine num(%d)", r.conf.SchedDiscvSvr, runtime.NumGoroutine())

	MesosDiscv := r.conf.SchedDiscvSvr
	regDiscv := rd.NewRegDiscover(MesosDiscv)
	if regDiscv == nil {
		blog.Errorf("new scheduler discover(%s) return nil", MesosDiscv)
		time.Sleep(3 * time.Second)
		go r.discvScheduler()
		return
	}

	blog.Infof("new scheduler discover(%s) succ, current goroutine num(%d)", MesosDiscv, runtime.NumGoroutine())

	err := regDiscv.Start()
	if err != nil {
		blog.Errorf("scheduler discover start error(%s)", err.Error())
		time.Sleep(3 * time.Second)
		go r.discvScheduler()
		return
	}

	blog.Infof("scheduler discover start succ, current goroutine num(%d)", runtime.NumGoroutine())

	discvPath := commtype.BCS_SERV_BASEPATH + "/" + commtype.BCS_MODULE_SCHEDULER
	discvMesosEvent, err := regDiscv.DiscoverService(discvPath)
	if err != nil {
		blog.Errorf("watch scheduler under (%s: %s) error(%s)", MesosDiscv, discvPath, err.Error())
		regDiscv.Stop()
		time.Sleep(3 * time.Second)
		go r.discvScheduler()
		return
	}

	blog.Infof("watch scheduler under (%s: %s), current goroutine num(%d)", MesosDiscv, discvPath, runtime.NumGoroutine())

	tick := time.NewTicker(180 * time.Second)
	for {
		select {
		case <-tick.C:
			blog.Infof("scheduler discover(%s:%s), curr scheduler:%s", MesosDiscv, discvPath, r.currScheduler)

		case <-r.cxt.Done():
			blog.Infof("scheReporter stop discover scheduler")
			return

		case event := <-discvMesosEvent:
			blog.Infof("discover event for scheduler")
			if event.Err != nil {
				blog.Errorf("get scheduler discover event err:%s", event.Err.Error())
				regDiscv.Stop()
				time.Sleep(3 * time.Second)
				go r.discvScheduler()
				return
			}

			currSched := ""
			blog.Infof("get scheduler node num(%d)", len(event.Server))

			for i, server := range event.Server {
				blog.Infof("get scheduler: server[%d]: %s %s", i, event.Key, server)

				var serverInfo commtype.SchedulerServInfo

				if err = json.Unmarshal([]byte(server), &serverInfo); err != nil {
					blog.Errorf("fail to unmarshal scheduler(%s), err:%s", string(server), err.Error())
				}

				if i == 0 {
					currSched = serverInfo.ServerInfo.Scheme + "://" + serverInfo.ServerInfo.IP + ":" + strconv.Itoa(int(serverInfo.ServerInfo.Port))
				}
			}

			if currSched != r.currScheduler {
				blog.Infof("scheduler changed(%s-->%s)", r.currScheduler, currSched)
				r.currScheduler = currSched
			}
		} // select
	} // for
}

func (r *scheReporter) requestSchedulerV1(method, uri string, data []byte) ([]byte, error) {
	if r.currScheduler == "" {
		return nil, fmt.Errorf("there is no scheduler")
	}

	uri = fmt.Sprintf("%s/v1%s", r.currScheduler, uri)

	blog.V(3).Infof("request uri %s data %s", uri, string(data))

	var by []byte
	var err error

	switch method {
	case "GET":
		by, err = r.cli.GET(uri, nil, data)

	case "POST":
		by, err = r.cli.POST(uri, nil, data)

	case "DELETE":
		by, err = r.cli.DELETE(uri, nil, data)

	case "PUT":
		by, err = r.cli.PUT(uri, nil, data)

	default:
		err = fmt.Errorf("uri %s method %s is invalid", uri, method)
	}

	return by, err
}

func (r *scheReporter) IsHealthy() (bool, error) {
	if r.currScheduler == "" {
		return false, fmt.Errorf("not found scheduler")
	}

	return true, nil
}
