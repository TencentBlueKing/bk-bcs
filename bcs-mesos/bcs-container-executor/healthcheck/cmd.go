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
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-container-executor/logs"

	dockerclient "github.com/fsouza/go-dockerclient"
	"golang.org/x/net/context"
)

//NewCommandChecker create http checker
func NewCommandChecker(cmd string, endpoint string, mechanism *TimeMechanism) (Checker, error) {
	if mechanism.IntervalSeconds <= mechanism.TimeoutSeconds {
		return nil, fmt.Errorf("Interval Seconds must larger than Timeout Seconds")
	}
	cxt, cancel := context.WithCancel(context.Background())
	checker := &CommandChecker{
		CheckerStat: CheckerStat{
			Failures:            0,
			Ticks:               0,
			ConsecutiveFailures: 0,
			Healthy:             true,
		},
		cmd:       cmd,
		endpoint:  endpoint,
		isPause:   false,
		cxt:       cxt,
		canceler:  cancel,
		mechanism: mechanism,
	}
	client, err := dockerclient.NewClient(endpoint)
	if err != nil {
		return nil, err
	}
	checker.client = client

	return checker, nil
}

//CommandChecker check for http protocol health check
type CommandChecker struct {
	CheckerStat
	containerId string //container id
	cmd         string //command value

	endpoint string //docker.sock
	client   *dockerclient.Client

	isPause   bool               //pause checker
	cxt       context.Context    //context to control exit
	canceler  context.CancelFunc //canceler
	mechanism *TimeMechanism     //time config for checker
}

func (check *CommandChecker) IsStarting() bool {
	return check.Started
}

//SetHost setting container_id
func (check *CommandChecker) SetHost(containerid string) {
	check.containerId = containerid
}

//Start start checker
func (check *CommandChecker) Start() {
	check.Started = true
	check.StartPoint = time.Now()
	time.Sleep(time.Duration(int64(check.mechanism.GracePeriodSeconds)) * time.Second)
	check.check()

	tick := time.NewTicker(time.Duration(int64(check.mechanism.IntervalSeconds)) * time.Second)
	defer tick.Stop()
	for {
		select {
		case <-check.cxt.Done():
			logs.Infof("CommandChecker container %s command %s ask to exit", check.containerId, check.cmd)
			return
		case check.LastCheck = <-tick.C:
			if check.isPause {
				continue
			}
			check.check()
		}
	}
}

func (check *CommandChecker) check() {
	check.Ticks++
	healthy := check.ReCheck()
	//notGrace := int(check.LastCheck.Unix()-check.StartPoint.Unix()) > check.mechanism.GracePeriodSeconds
	//if !healthy && notGrace {
	if !healthy {
		check.LastFailure = check.LastCheck
		check.ConsecutiveFailures++
		check.Healthy = false
		logs.Infof("CommandChecker container %s command %s become **Unhealthy**", check.containerId, check.cmd)
	} else {
		check.Healthy = true
		check.ConsecutiveFailures = 0
	}
}

//Stop stop checker
func (check *CommandChecker) Stop() {
	check.canceler()
}

//ReCheck ask checker to check
func (check *CommandChecker) ReCheck() bool {
	cmds := strings.Split(strings.TrimSpace(check.cmd), " ")
	//create exec with command
	createOpt := dockerclient.CreateExecOptions{
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          false,
		Cmd:          cmds,
		Container:    check.containerId,
	}
	logs.Infof("create command healthcheck(%v)", cmds)
	exeInst, err := check.client.CreateExec(createOpt)
	if err != nil {
		logs.Errorf("CommandChecker CreateExec container %s command %s error %s",
			check.containerId, check.cmd, err.Error())
		return false
	}
	//start exec
	var outBuf bytes.Buffer
	var errBuf bytes.Buffer
	startOpt := dockerclient.StartExecOptions{
		OutputStream: &outBuf,
		ErrorStream:  &errBuf,
	}
	err = check.client.StartExec(exeInst.ID, startOpt)
	if err != nil {
		logs.Errorf("CommandChecker StartExec container %s command %s error %s",
			check.containerId, check.cmd, err.Error())
		return false
	}

	ins, err := check.client.InspectExec(exeInst.ID)
	if err != nil {
		logs.Errorf("CommandChecker InspectExec container %s command %s error %s",
			check.containerId, check.cmd, err.Error())
		return false
	}

	if ins.ExitCode == 0 {
		return true
	}
	logs.Errorf("CommandChecker exec container %s command %s failed: stdout(%s), stderr(%s)",
		check.containerId, check.cmd, string(outBuf.Bytes()), string(errBuf.Bytes()))

	return false
}

//Pause pause check
func (check *CommandChecker) Pause() error {
	check.isPause = true
	return nil
}

//Resume arouse checker
func (check *CommandChecker) Resume() error {
	check.isPause = false
	return nil
}

//Name get check name
func (check *CommandChecker) Name() string {
	return CommandHealthcheck
}

//Relation checker relative to container
func (check *CommandChecker) Relation() string {
	return check.containerId
}
