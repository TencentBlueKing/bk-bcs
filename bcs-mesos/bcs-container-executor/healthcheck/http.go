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

package healthcheck

import (
	"fmt"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-container-executor/logs"
	"net/http"
	"strconv"
	"time"

	"golang.org/x/net/context"
)

//NewHTTPChecker create http checker
func NewHTTPChecker(container, schema string, port int, path string, mechanism *TimeMechanism, notify FailureNotify) (Checker, error) {
	if port <= 1 {
		return nil, fmt.Errorf("http checker port is invalid")
	}
	if mechanism.IntervalSeconds <= mechanism.TimeoutSeconds {
		return nil, fmt.Errorf("Interval Seconds must larger than Timeout Seconds")
	}
	cxt, cancel := context.WithCancel(context.Background())
	checker := &HTTPChecker{
		CheckerStat: CheckerStat{
			Failures:            0,
			Ticks:               0,
			ConsecutiveFailures: 0,
			Healthy:             true,
		},
		container: container,
		schema:    schema,
		port:      port,
		path:      path,
		isPause:   false,
		cxt:       cxt,
		canceler:  cancel,
		mechanism: mechanism,
		notify:    notify,
	}
	//default schema
	if len(schema) == 0 {
		checker.schema = "http"
	}
	return checker, nil
}

//HTTPChecker check for http protocol health check
type HTTPChecker struct {
	CheckerStat
	container string             //container id
	schema    string             //http or https
	ipaddr    string             //ip address to check
	port      int                //port to check
	path      string             //path for http request
	isPause   bool               //pause checker
	cxt       context.Context    //context to control exit
	canceler  context.CancelFunc //canceler
	mechanism *TimeMechanism     //time config for checker
	notify    FailureNotify      //failure callback
}

func (check *HTTPChecker) IsStarting() bool {
	return check.Started
}

//SetHost setting host / ipaddress
func (check *HTTPChecker) SetHost(host string) {
	check.ipaddr = host
}

//Start start checker
func (check *HTTPChecker) Start() {
	check.Started = true
	check.StartPoint = time.Now()
	time.Sleep(time.Duration(int64(check.mechanism.GracePeriodSeconds)) * time.Second)
	check.check()

	tick := time.NewTicker(time.Duration(int64(check.mechanism.IntervalSeconds)) * time.Second)
	defer tick.Stop()
	for {
		select {
		case <-check.cxt.Done():
			logs.Infof("HTTPChecker %s://%s:%d%s ask to exit", check.schema, check.ipaddr, check.port, check.path)
			return
		case check.LastCheck = <-tick.C:
			if check.isPause {
				continue
			}
			check.check()
		}
	}
}

func (check *HTTPChecker) check() {
	check.Ticks++
	healthy := check.ReCheck()
	//notGrace := int(check.LastCheck.Unix()-check.StartPoint.Unix()) > check.mechanism.GracePeriodSeconds
	//if !healthy && notGrace {
	if !healthy {
		check.LastFailure = check.LastCheck
		check.ConsecutiveFailures++
		check.Healthy = false
		logs.Infof("HTTPChecker %s://%s:%d%s become **Unhealthy**", check.schema, check.ipaddr, check.port, check.path)
		if check.notify != nil {
			check.notify(check)
		}
	} else {
		check.Healthy = true
		check.ConsecutiveFailures = 0
	}
}

//Stop stop checker
func (check *HTTPChecker) Stop() {
	check.canceler()
}

//ReCheck ask checker to check
func (check *HTTPChecker) ReCheck() bool {
	client := &http.Client{
		Timeout: time.Second * time.Duration(int64(check.mechanism.TimeoutSeconds)),
	}
	dest := check.schema + "://" + check.ipaddr + ":" + strconv.Itoa(check.port) + check.path
	response, err := client.Get(dest)
	if err != nil {
		return false
	}
	defer response.Body.Close()
	if response.StatusCode >= 200 && response.StatusCode < 400 {
		return true
	}
	return false
}

//Pause pause check
func (check *HTTPChecker) Pause() error {
	check.isPause = true
	return nil
}

//Resume arouse checker
func (check *HTTPChecker) Resume() error {
	check.isPause = false
	return nil
}

//Name get check name
func (check *HTTPChecker) Name() string {
	return HttpHealthcheck
}

//Relation checker relative to container
func (check *HTTPChecker) Relation() string {
	return check.container
}
