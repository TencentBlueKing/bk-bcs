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

package proc_daemon

import (
	"fmt"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-process-executor/process-executor/types"
	"k8s.io/kubernetes/staging/src/k8s.io/apimachinery/pkg/util/json"
)

const (
	ProcessDaemonEndpoint = "/var/run/process.sock"
)

type daemon struct {
	cli *HttpConnection
}

func NewDaemon() ProcDaemon {
	return &daemon{
		cli: NewHttpConnection(ProcessDaemonEndpoint),
	}
}

func (d *daemon) CreateProcess(processInfo *types.ProcessInfo) error {
	by, _ := json.Marshal(processInfo)
	_, err := d.cli.requestProcessDaemon("POST", "/process", by)
	if err != nil {
		blog.Errorf("daemon create process %s error %s", processInfo.Id, err.Error())
		return err
	}

	return nil
}

func (d *daemon) InspectProcessStatus(procId string) (*types.ProcessStatusInfo, error) {
	by, err := d.cli.requestProcessDaemon("GET", fmt.Sprintf("/process/%s/status", procId), nil)
	if err != nil {
		blog.Errorf("daemon inspect process %s error %s", procId, err.Error())
		return nil, err
	}

	var status *types.ProcessStatusInfo
	err = json.Unmarshal(by, &status)
	if err != nil {
		blog.Errorf("Unmarshal data %s to types.ProcessStatusInfo error %s", string(by), err.Error())
		return nil, err
	}

	return status, nil
}

func (d *daemon) StopProcess(procId string, timeout int) error {
	_, err := d.cli.requestProcessDaemon("PUT", fmt.Sprintf("/process/%s/stop/%d", procId, timeout), nil)
	if err != nil {
		blog.Errorf("daemon stop process %s error %s", procId, err.Error())
		return err
	}

	return nil
}

func (d *daemon) DeleteProcess(procId string) error {
	_, err := d.cli.requestProcessDaemon("DELETE", fmt.Sprintf("/process/%s", procId), nil)
	if err != nil {
		blog.Errorf("daemon delete process %s error %s", procId, err.Error())
		return err
	}

	return nil
}

func (d *daemon) ReloadProcess(procId string) error {
	return nil
}

func (d *daemon) RestartProcess(procId string) error {
	return nil
}
