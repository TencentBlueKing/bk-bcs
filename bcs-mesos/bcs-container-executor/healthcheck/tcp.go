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
	"net"
	"strconv"
	"time"

	"golang.org/x/net/context"
)

//NewTCPChecker create TCP checker
func NewTCPChecker(container string, port int, mechanism *TimeMechanism, notify FailureNotify) (Checker, error) {
	if mechanism.IntervalSeconds <= mechanism.TimeoutSeconds {
		return nil, fmt.Errorf("Interval Seconds must larger than Timeout Seconds")
	}
	if port < 1 {
		return nil, fmt.Errorf("port is invalid")
	}
	cxt, cancel := context.WithCancel(context.Background())
	check := &TCPChecker{
		CheckerStat: CheckerStat{
			Failures:            0,
			Ticks:               0,
			ConsecutiveFailures: 0,
			Healthy:             true,
		},
		container: container,
		port:      port,
		isPause:   false,
		cxt:       cxt,
		canceler:  cancel,
		mechanism: mechanism,
	}
	return check, nil
}

//TCPChecker create check for tcp port check
type TCPChecker struct {
	CheckerStat
	container string             //name for container
	ipaddr    string             //ip address to check
	port      int                //port to check
	isPause   bool               //status for pause
	cxt       context.Context    //context to control exit
	canceler  context.CancelFunc //canceler
	mechanism *TimeMechanism     //time config for checker
	notify    FailureNotify      //callback when unhealthy
}

func (check *TCPChecker) IsStarting() bool {
	return check.Started
}

//SetHost setting host / ipaddress
func (check *TCPChecker) SetHost(host string) {
	check.ipaddr = host
}

//Start start checker, must running in indivisual goroutine
func (check *TCPChecker) Start() {
	check.Started = true
	check.StartPoint = time.Now()
	time.Sleep(time.Duration(int64(check.mechanism.GracePeriodSeconds)) * time.Second)
	check.check()

	tick := time.NewTicker(time.Duration(int64(check.mechanism.IntervalSeconds)) * time.Second)
	defer tick.Stop()
	for {
		select {
		case <-check.cxt.Done():
			logs.Infof("TCPChecker ###%s:%d### ask to exit", check.ipaddr, check.port)
			return
		case check.LastCheck = <-tick.C:
			if check.isPause {
				continue
			}
			check.check()
		}
	}
}

func (check *TCPChecker) check() {
	check.Ticks++
	healthy := check.ReCheck()
	/*notGrace := int(check.LastCheck.Unix()-check.StartPoint.Unix()) > check.mechanism.GracePeriodSeconds
	if !healthy && notGrace {*/
	if !healthy {
		check.LastFailure = check.LastCheck
		check.ConsecutiveFailures++
		check.Healthy = false
		logs.Infof("TCPChecker %s:%d become **Unhealthy**", check.ipaddr, check.port)
		if check.notify != nil {
			check.notify(check)
		}
	} else {
		check.Healthy = true
		check.ConsecutiveFailures = 0
	}
}

//Stop stop checker
func (check *TCPChecker) Stop() {
	check.canceler()
}

//ReCheck ask checker to check
func (check *TCPChecker) ReCheck() bool {
	dest := check.ipaddr + ":" + strconv.Itoa(check.port)
	con, err := net.DialTimeout("tcp", dest, time.Duration(int64(check.mechanism.TimeoutSeconds))*time.Second)
	if err != nil {
		return false
	}
	defer con.Close()
	return true
}

//Pause pause check
func (check *TCPChecker) Pause() error {
	check.isPause = true
	return nil
}

//Resume arouse checker
func (check *TCPChecker) Resume() error {
	check.isPause = false
	return nil
}

//Name get check name
func (check *TCPChecker) Name() string {
	return TcpHealthcheck
}

//Relation checker relative to container
func (check *TCPChecker) Relation() string {
	return check.container
}
