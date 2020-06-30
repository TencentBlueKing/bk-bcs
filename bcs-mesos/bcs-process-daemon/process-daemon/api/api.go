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

package api

import (
	"encoding/json"
	"fmt"
	"github.com/Tencent/bk-bcs/bcs-common/common"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	bhttp "github.com/Tencent/bk-bcs/bcs-common/common/http"
	"github.com/Tencent/bk-bcs/bcs-common/common/http/httpserver"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-process-daemon/process-daemon/manager"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-process-executor/process-executor/types"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"

	restful "github.com/emicklei/go-restful"
)

type Router struct {
	backend manager.Manager
	actions []*httpserver.Action

	Variables map[string]string
}

// NewRouter return api router
func NewRouter(b manager.Manager) *Router {
	r := &Router{
		backend: b,
		actions: make([]*httpserver.Action, 0),
	}

	r.initRoutes()
	r.initVariables()
	return r
}

func (r *Router) initVariables() {
	workspace := r.backend.GetConfig().WorkspaceDir

	r.Variables = map[string]string{
		"${work_base_dir}": filepath.Join(workspace, "work_dir"),
		"${run_base_dir}":  filepath.Join(workspace, "run_dir"),
	}
}

//get http api routing table information, and use it to register http client
func (r *Router) GetActions() []*httpserver.Action {
	return r.actions
}

func createResponeData(err error, code int, msg string, data interface{}) []byte {
	var rpyErr error
	if err != nil {
		rpyErr = bhttp.InternalError(code, msg)
	} else {
		rpyErr = fmt.Errorf(bhttp.GetRespone(common.BcsSuccess, common.BcsSuccessStr, data))
	}

	blog.V(3).Infof("createRespone: %s", rpyErr.Error())

	return []byte(rpyErr.Error())
}

func (r *Router) initRoutes() {
	//launch container webconsole proxy
	r.actions = append(r.actions, httpserver.NewAction("POST", "/process", nil, r.createProcess))
	r.actions = append(r.actions, httpserver.NewAction("GET", "/process/{id}/status", nil, r.inspectProcessStatus))
	r.actions = append(r.actions, httpserver.NewAction("PUT", "/process/{id}/stop/{timeout}", nil, r.stopProcess))
	r.actions = append(r.actions, httpserver.NewAction("DELETE", "/process/{id}", nil, r.deleteProcess))
}

func (r *Router) createProcess(req *restful.Request, resp *restful.Response) {
	by, err := ioutil.ReadAll(req.Request.Body)
	if err != nil {
		blog.Errorf("Router create process read request body error %s",
			err.Error())
		data := createResponeData(err, common.BcsErrDaemonCreateProcessFailed,
			err.Error(), nil)
		resp.Write(data)
		return
	}

	blog.Infof("Router create process body(%s)", string(by))
	by = r.parseVariable(by)

	blog.Infof("%s", string(by))

	var processInfo *types.ProcessInfo
	err = json.Unmarshal(by, &processInfo)
	if err != nil {
		blog.Errorf("Router create process body(%s) unmarshal error %s",
			string(by), err.Error())
		data := createResponeData(err, common.BcsErrDaemonCreateProcessFailed,
			err.Error(), nil)
		resp.Write(data)
		return
	}

	err = r.backend.CreateProcess(processInfo)
	if err != nil {
		data := createResponeData(err, common.BcsErrDaemonCreateProcessFailed,
			err.Error(), nil)
		resp.Write(data)
		return
	}

	data := createResponeData(nil, 0, "", nil)
	resp.Write(data)
	blog.Infof("Router create process %s success", processInfo.Id)

	return
}

func (r *Router) inspectProcessStatus(req *restful.Request, resp *restful.Response) {
	processId := req.PathParameter("id")
	blog.V(3).Infof("Router inspect process %s status", processId)

	status, err := r.backend.InspectProcessStatus(processId)
	if err != nil {
		data := createResponeData(err, common.BcsErrDaemonInspectProcessFailed,
			err.Error(), nil)
		resp.Write(data)
		return
	}

	data := createResponeData(nil, 0, "", status)
	resp.Write(data)
	blog.V(3).Infof("Router inspect process %s status success", processId)

	return
}

func (r *Router) stopProcess(req *restful.Request, resp *restful.Response) {
	processId := req.PathParameter("id")
	timeout := req.PathParameter("timeout")
	blog.Infof("Router stop process %s timeout %s", processId, timeout)

	time, err := strconv.Atoi(timeout)
	if err != nil {
		blog.Errorf("Router process %s timeout %s, err: %s", processId, timeout, err.Error())
		data := createResponeData(err, common.BcsErrDaemonStopProcessFailed,
			err.Error(), nil)
		resp.Write(data)
		return
	}

	err = r.backend.StopProcess(processId, time)
	if err != nil {
		data := createResponeData(err, common.BcsErrDaemonStopProcessFailed,
			err.Error(), nil)
		resp.Write(data)
		return
	}

	data := createResponeData(nil, 0, "", nil)
	resp.Write(data)
	blog.Infof("Router stop process %s success", processId)

	return
}

func (r *Router) deleteProcess(req *restful.Request, resp *restful.Response) {
	processId := req.PathParameter("id")
	blog.Infof("Router delete process %s", processId)

	err := r.backend.DeleteProcess(processId)
	if err != nil {
		data := createResponeData(err, common.BcsErrDaemonDeleteProcessFailed,
			err.Error(), nil)
		resp.Write(data)
		return
	}

	data := createResponeData(nil, 0, "", nil)
	resp.Write(data)
	blog.Infof("Router delete process %s success", processId)

	return
}

//replace variables
func (r *Router) parseVariable(by []byte) []byte {
	str := string(by)

	for k, v := range r.Variables {
		str = strings.Replace(str, k, v, -1)
	}

	return []byte(str)
}
