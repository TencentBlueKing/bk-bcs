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

package executor

import (
	"bk-bcs/bcs-common/common/blog"
	proc_daemon "bk-bcs/bcs-mesos/bcs-process-executor/process-executor/proc-daemon"
	"bk-bcs/bcs-mesos/bcs-process-executor/process-executor/types"
	"bk-bcs/bcs-mesos/bcs-scheduler/src/mesosproto/mesos"
	bcstype "bk-bcs/bcs-mesos/bcs-scheduler/src/types"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/user"
	"strconv"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/pborman/uuid"
)

const (
	//report task status time
	ReportTaskStatusPeriod = 30 //seconds
)

type bcsExecutor struct {
	sync.RWMutex

	tasks         map[string]*types.ProcessTaskInfo //key: taskinfo id
	status        types.ExecutorStatus
	callbackFuncs map[types.CallbackFuncType]interface{}
	procDaemon    proc_daemon.ProcDaemon

	ackUpdates   map[string]*mesos.TaskStatus //key: taskinfo id
	updatesLocks sync.RWMutex

	cxt    context.Context    //context for cancel
	cancel context.CancelFunc //function for cancel

	isAskedShutdown bool //scheduler shutdown the exeutor, and is true
}

//NewExecutor
func NewExecutor(cxt context.Context) Executor {
	executor := &bcsExecutor{
		tasks:         make(map[string]*types.ProcessTaskInfo),
		callbackFuncs: make(map[types.CallbackFuncType]interface{}),
		status:        types.ExecutorStatusUnknown,
		procDaemon:    proc_daemon.NewDaemon(),
		ackUpdates:    make(map[string]*mesos.TaskStatus),
	}

	executor.cxt, executor.cancel = context.WithCancel(cxt)
	return executor
}

//RegisterCallbackFunc
func (e *bcsExecutor) RegisterCallbackFunc(funcType types.CallbackFuncType, fun interface{}) {
	e.callbackFuncs[funcType] = fun
}

// GetExecutorStatus
func (e *bcsExecutor) GetExecutorStatus() types.ExecutorStatus {
	return e.status
}

//LaunchTaskgroup
func (e *bcsExecutor) LaunchTaskgroup(taskgroup *mesos.TaskGroupInfo) {
	var err error

	e.Lock()
	if e.status != types.ExecutorStatusUnknown {
		blog.Errorf("bcsExecutor current status %s, not receive LaunchTaskgroup", string(e.status))
		e.Unlock()
		return
	}
	blog.Infof("bcsExecutor launch taskgroup start")
	e.status = types.ExecutorStatusLaunching
	e.Unlock()

	//craete process taskinfo
	for _, task := range taskgroup.GetTasks() {
		proc, err := createProcessTaskinfo(task)
		if err != nil {
			blog.Errorf("Launch task %s failed, update task status TASK_ERROR", task.GetTaskId().GetValue())
			e.updateTaskStatus(task.GetTaskId().GetValue(), types.TaskStatusError, err.Error())
			e.status = types.ExecutorStatusFinish
			return
		}

		e.tasks[proc.TaskId] = proc
	}

	for id, task := range e.tasks {
		//write local file
		for _, file := range task.LocalFiles {
			err = writeLocalFile(file)
			if err != nil {
				blog.Errorf("Launch task %s write file %s error %s, update task status TASK_FAILED", task.TaskId, file.To, err.Error())
				e.updateTaskStatus(id, types.TaskStatusFailed, err.Error())
				go e.innerShutdown()
				return
			}
		}

		blog.Infof("start create process task %s", id)
		err = e.procDaemon.CreateProcess(task.ProcInfo)
		if err != nil {
			blog.Errorf("Launch task %s error %s, update task status TASK_FAILED", task.TaskId, err.Error())
			e.updateTaskStatus(id, types.TaskStatusFailed, err.Error())
			go e.innerShutdown()
			return
		}
	}
	e.status = types.ExecutorStatusRunning
	blog.Infof("bcsExecutor launch taskgroup done")
	go e.loopInspectTasks()
}

//loopInspectTasks
func (e *bcsExecutor) loopInspectTasks() {
	var inspectNum uint64

	for {
		ticker := time.NewTicker(time.Second)

		select {
		case <-e.cxt.Done():
			blog.Infof("stop loopInspectTasks")
			return

		case <-ticker.C:
			inspectNum++
		}

		for taskid, task := range e.tasks {
			status, err := e.procDaemon.InspectProcessStatus(taskid)
			if err != nil {
				blog.Errorf("inspect process %s status error %s", taskid, err.Error())
				continue
			}

			oldStatus := task.Status
			task.Status, err = e.getTaskStatusFromProcessStatus(status.Status)
			if err != nil {
				blog.Errorf(fmt.Sprintf("process %s %s", taskid, err.Error()))
				continue
			}

			var isUpdate bool
			if oldStatus != task.Status {
				isUpdate = true
				blog.Infof("process %s status %s change to %s", taskid, string(oldStatus), status.Status)
			}
			if inspectNum%ReportTaskStatusPeriod == 0 {
				isUpdate = true
				blog.Infof("update process %s status %s", taskid, status.Status)
			}

			if !isUpdate {
				continue
			}

			e.updateTaskStatus(taskid, task.Status, status.Message)

			if task.Status == types.TaskStatusFinish || task.Status == types.TaskStatusFailed {

				blog.Infof("task %s status %s, and shutdown executor", task.TaskId, task.Status)
				go e.innerShutdown()
				return
			}
		}
	}
}

//getTaskStatusFromProcessStatus
func (e *bcsExecutor) getTaskStatusFromProcessStatus(status types.ProcessStatusType) (types.TaskStatus, error) {
	switch status {
	//process status staging
	//show executor recieve the tasks
	case types.ProcessStatusStaging:
		return types.TaskStatusStaging, nil
	//process status starting
	//show executor check tasks valid, starting it
	case types.ProcessStatusStarting:
		return types.TaskStatusStarting, nil
	//process status running
	//started tasks, process running
	case types.ProcessStatusRunning:
		return types.TaskStatusRunning, nil
	//process status stopping
	//executor recieve shutdown command
	//then stop process
	case types.ProcessStatusStopping:
		return types.TaskStatusKilling, nil
	//process status stopped
	//executor stopped process
	case types.ProcessStatusStopped:
		if e.isAskedShutdown {
			return types.TaskStatusFinish, nil
		}
		return types.TaskStatusFailed, nil

	default:
		return "", fmt.Errorf("status %s is invalid", status)
	}
}

//Shutdown
//recieve the shutdown command
//will stop all process tasks
//exit 0
func (e *bcsExecutor) Shutdown() {
	e.isAskedShutdown = true

	//shutdown
	e.innerShutdown()
}

//innerShutdown
func (e *bcsExecutor) innerShutdown() {
	e.Lock()
	if e.status == types.ExecutorStatusShutdown || e.status == types.ExecutorStatusFinish {
		e.Unlock()
		return
	}
	blog.Infof("shut down bcs executor")
	e.status = types.ExecutorStatusShutdown
	e.Unlock()

	for {
		time.Sleep(time.Second)

		var downNum int
		for taskid, task := range e.tasks {
			status, err := e.procDaemon.InspectProcessStatus(taskid)
			if err != nil {
				blog.Errorf("inspect process %s status error %s", taskid, err.Error())
				break
			}

			oldStatus := task.Status
			task.Status, err = e.getTaskStatusFromProcessStatus(status.Status)

			if oldStatus != task.Status {
				e.updateTaskStatus(taskid, task.Status, status.Message)
			}

			//if process status starting or running, then stop it
			if task.Status == types.TaskStatusStarting || task.Status == types.TaskStatusRunning {
				blog.Infof("stop process %s status %s", task.TaskId, task.Status)
				err = e.procDaemon.StopProcess(taskid, task.ProcInfo.StopTimeout)
				if err != nil {
					blog.Errorf("stop process %s error %s", taskid, err.Error())
				}
			}

			if task.Status == types.TaskStatusStaging || task.Status == types.TaskStatusFailed ||
				task.Status == types.TaskStatusFinish {

				downNum++
				//delete process info
				err = e.procDaemon.DeleteProcess(taskid)
				if err != nil {
					blog.Errorf("delete process %s error %s", taskid, err.Error())
				}
			}
		}

		if downNum == len(e.tasks) {
			//wait for mesos slave acknowledge task status message
			e.waitForAckAndExit()
			blog.Infof("executor's tasks all are down, then executor will exit")
			e.status = types.ExecutorStatusFinish
			return
		}
	}
}

//ReloadTasks
//command task reload command
func (e *bcsExecutor) ReloadTasks() error {
	for _, task := range e.tasks {
		blog.Infof("reload task %s start...", task.TaskId)
		err := e.procDaemon.ReloadProcess(task.TaskId)
		if err != nil {
			blog.Errorf("reload process %s error %s", task.TaskId, err.Error())
			return err
		}
	}

	return nil
}

// RestartTasks
//restart process tasks
func (e *bcsExecutor) RestartTasks() error {
	for _, task := range e.tasks {
		blog.Infof("reload task %s start...", task.TaskId)
		err := e.procDaemon.RestartProcess(task.TaskId)
		if err != nil {
			blog.Errorf("reload process %s error %s", task.TaskId, err.Error())
			return err
		}
	}

	return nil
}

//waitForAckAndExit
func (e *bcsExecutor) waitForAckAndExit() {
	if len(e.ackUpdates) == 0 {
		blog.Infof("bcsExecutor ack updates message is empty, and exit")
		return
	}

	//check all update info acknowledged
	checkTick := time.NewTicker(500 * time.Microsecond)
	timeoutTick := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-timeoutTick.C:
			blog.Infof("bcsExecutor wait acknowledgement from slave timeout(5 seconds), and exit")
			return
		case <-checkTick.C:
			if len(e.ackUpdates) == 0 {
				blog.Infof("bcsExecutor ack updates message is empty, and exit")
				return
			}
		}
	}
}

// AckTaskStatusMessage
func (e *bcsExecutor) AckTaskStatusMessage(taskId string, uid []byte) {
	e.updatesLocks.Lock()
	defer e.updatesLocks.Unlock()
	ouid := uuid.UUID(uid)

	blog.Infof("get slave acknowledged taskid %s uuid %s", taskId, ouid.String())

	status, ok := e.ackUpdates[taskId]
	if !ok {
		return
	}

	if uuid.UUID(status.Uuid).String() == ouid.String() {
		delete(e.ackUpdates, taskId)
	}
}

// updateTaskStatus
// ticker report tasks status
func (e *bcsExecutor) updateTaskStatus(taskId string, status types.TaskStatus, msg string) {
	var state mesos.TaskState

	switch status {
	case types.TaskStatusStarting:
		state = mesos.TaskState_TASK_STARTING

	case types.TaskStatusRunning:
		state = mesos.TaskState_TASK_RUNNING

	case types.TaskStatusKilling:
		state = mesos.TaskState_TASK_KILLING

	case types.TaskStatusFinish:
		state = mesos.TaskState_TASK_FINISHED

	case types.TaskStatusFailed:
		state = mesos.TaskState_TASK_FAILED

	case types.TaskStatusError:
		state = mesos.TaskState_TASK_ERROR

	default:
		blog.Errorf("task %s status %s is invalid", taskId, string(status))
		return
	}

	update := &mesos.TaskStatus{
		TaskId:  &mesos.TaskID{Value: proto.String(taskId)},
		State:   state.Enum(),
		Message: proto.String(msg),
		Source:  mesos.TaskStatus_SOURCE_EXECUTOR.Enum(),
	}

	ID := uuid.NewUUID()
	now := float64(time.Now().Unix())
	update.Timestamp = proto.Float64(now)
	update.Uuid = ID

	Func, ok := e.callbackFuncs[types.CallbackFuncUpdateTask]
	if !ok {
		blog.Errorf("CallbackFuncUpdateTask not found")
		return
	}

	blog.Infof("update task %s status %s uuid %s msg %s", taskId, state.String(), ID.String(), msg)

	e.updatesLocks.Lock()
	e.ackUpdates[taskId] = update
	e.updatesLocks.Unlock()

	updateFunc := Func.(types.UpdateTaskFunc)
	err := updateFunc(update)
	if err != nil {
		blog.Errorf("update task %s status %s msg %s error %s", taskId, state.String(), msg, err.Error())
	}
	return
}

// createProcessTaskinfo
func createProcessTaskinfo(task *mesos.TaskInfo) (*types.ProcessTaskInfo, error) {
	if task.GetCommand() == nil {
		err := fmt.Errorf("task %s is not command task", task.GetTaskId().GetValue())
		return nil, err
	}

	by, _ := json.Marshal(task)
	blog.Infof("taskinfo %s", string(by))

	processTask := &types.ProcessTaskInfo{
		TaskId: task.GetTaskId().GetValue(),
		Status: types.TaskStatusStaging,
		ProcInfo: &types.ProcessInfo{
			Id:   task.GetTaskId().GetValue(),
			Envs: make([]string, 0),
			Argv: make([]string, 0),
			Uris: make([]*types.Uri, 0),
		},
	}

	dc, err := getBcsDataClass(task)
	if err != nil {
		return nil, err
	}

	//set process task parameters by dataclass
	setProcessTaskParams(processTask, dc)

	//setting Env
	envs := task.GetCommand().GetEnvironment()
	if envs != nil {
		for _, env := range envs.GetVariables() {
			item := fmt.Sprintf("%s=%s", env.GetName(), env.GetValue())
			processTask.ProcInfo.Envs = append(processTask.ProcInfo.Envs, item)
		}
	}

	processTask.ProcInfo.Resource = dc.Resources
	processTask.ProcInfo.StartCmd = dc.ProcInfo.StartCmd
	processTask.ProcInfo.StopCmd = dc.ProcInfo.StopCmd
	processTask.ProcInfo.StopTimeout = int(task.KillPolicy.GetGracePeriod().GetNanoseconds() / 1000 / 1000 / 1000)
	processTask.ProcInfo.Resource = dc.Resources
	processTask.ProcInfo.WorkDir = dc.ProcInfo.WorkPath
	processTask.ProcInfo.PidFile = dc.ProcInfo.PidFile
	processTask.ProcInfo.ProcessName = dc.ProcInfo.ProcName
	processTask.ProcInfo.ExecutorId = task.Executor.GetExecutorId().GetValue()
	processTask.ProcInfo.KillCmd = dc.ProcInfo.KillCmd
	processTask.ProcInfo.StartGracePeriod = int64(dc.ProcInfo.StartGracePeriod)

	for _, uri := range dc.ProcInfo.Uris {
		u := &types.Uri{
			Value:     uri.Value,
			User:      uri.User,
			Pwd:       uri.Pwd,
			OutputDir: uri.OutputDir,
		}
		processTask.ProcInfo.Uris = append(processTask.ProcInfo.Uris, u)
	}

	by, _ = json.Marshal(processTask)
	blog.Infof("launch process task %s", string(by))
	return processTask, nil
}

//set dataclass messages to processtaskinfo, include: secret info, envs, local file, remote file
func setProcessTaskParams(processInfo *types.ProcessTaskInfo, dataClass *bcstype.DataClass) error {
	if dataClass.Msgs == nil || len(dataClass.Msgs) == 0 {
		return nil
	}

	for _, item := range dataClass.Msgs {
		switch *item.Type {
		case bcstype.Msg_SECRET:
			data, err := base64.StdEncoding.DecodeString(*item.Secret.Value)
			if err != nil {
				return fmt.Errorf("Decode secret value %s error %s", *item.Secret.Value, err.Error())
			}

			switch *item.Secret.Type {
			case bcstype.Secret_Env:
				secretEnv := fmt.Sprintf("%s=%s", *item.Secret.Name, string(data))
				blog.Infof("Additional adding Secrets %s to Environment in customSetting", secretEnv)
				processInfo.ProcInfo.Envs = append(processInfo.ProcInfo.Envs, secretEnv)

			case bcstype.Secret_File:
				secretFile := &types.LocalFile{
					To:    *item.Secret.Name,
					User:  "root",
					Right: "r",
					Value: string(data),
				}

				blog.Infof("Setting secret %s to file", secretFile.To)
				processInfo.LocalFiles = append(processInfo.LocalFiles, secretFile)
			}

		case bcstype.Msg_REMOTE:
			value, err := downloadRemoteFile(*item.Remote.From, *item.Remote.RemoteUser, *item.Remote.RemotePasswd)
			if err != nil {
				return err
			}
			file := &types.LocalFile{
				To:    *item.Remote.To,
				Right: *item.Remote.Right,
				User:  *item.Remote.User,
				Value: value,
			}
			blog.Infof("setting remote file %s to %s file", *item.Remote.From, file.To)
			processInfo.LocalFiles = append(processInfo.LocalFiles, file)

		case bcstype.Msg_ENV:
			//base64 decoding
			value, err := base64.StdEncoding.DecodeString(*item.Env.Value)
			if err != nil {
				blog.Errorf("decode bcs custom Environment error %s", err.Error())
				return err
			}
			env := fmt.Sprintf("%s=%s", *item.Env.Name, string(value))
			blog.Infof("custom Msg [%s=%s] to Environment in customSetting", *item.Env.Name, value)
			processInfo.ProcInfo.Envs = append(processInfo.ProcInfo.Envs, env)

		case bcstype.Msg_LOCALFILE:
			//base64 decoding
			value, err := base64.StdEncoding.DecodeString(*item.Local.Base64)
			if err != nil {
				blog.Errorf("decode bcs localfile Base64 %s error %s", *item.Local.Base64, err.Error())
				return err
			}

			file := &types.LocalFile{
				To:    *item.Local.To,
				Right: *item.Local.Right,
				User:  *item.Local.User,
				Value: string(value),
			}

			blog.Infof("setting local file %s", file.To)
			processInfo.LocalFiles = append(processInfo.LocalFiles, file)
		}
	}

	return nil
}

// getBcsDataClass
func getBcsDataClass(taskInfo *mesos.TaskInfo) (*bcstype.DataClass, error) {
	data := taskInfo.GetData()
	if data == nil || len(data) == 0 {
		err := fmt.Errorf("task %s data is empty", taskInfo.GetTaskId().GetValue())
		blog.Errorf(err.Error())
		return nil, err
	}

	by, err := base64.StdEncoding.DecodeString(string(data))
	if err != nil {
		blog.Errorf("task %s decode data error %s", taskInfo.GetTaskId().GetValue(), err.Error())
		return nil, err
	}

	blog.Infof("process %s dataclass %s", taskInfo.GetTaskId().GetValue(), string(by))
	var dc *bcstype.DataClass
	err = json.Unmarshal(by, &dc)
	if err != nil {
		blog.Errorf("task %s Unmarshal data %s to DataClass error: %s", taskInfo.GetTaskId().GetValue(), string(by), err.Error())
		return nil, err
	}

	return dc, nil
}

//downloadRemoteFile download remote file, change to local one
func downloadRemoteFile(from, user, pwd string) (string, error) {
	//download content from remote http url.
	client := http.Client{
		Timeout: time.Duration(120 * time.Second),
	}

	request, err := http.NewRequest("GET", from, nil)
	if err != nil {
		blog.Errorf("http NewRequest %s error %s", from, err.Error())
		return "", err
	}

	if len(user) != 0 {
		request.SetBasicAuth(user, pwd)
	}

	response, err := client.Do(request)
	if err != nil {
		blog.Errorf("download remote file %s error %s", from, err.Error())
		return "", err
	}
	defer response.Body.Close()

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		blog.Errorf("read url %s response body error %s", from, err.Error())
		return "", err
	}

	if response.StatusCode != http.StatusOK {
		blog.Errorf("request url %s response statuscode %d body %s", from, response.StatusCode, string(data))
		return "", fmt.Errorf("request url %s response statuscode %d body %s", from, response.StatusCode, string(data))
	}

	return string(data), nil
}

// writeLocalFile
func writeLocalFile(localFile *types.LocalFile) error {
	file, err := os.Create(localFile.To)
	if err != nil {
		blog.Errorf("create file %s error %s", localFile.To, err.Error())
		return err
	}
	defer file.Close()

	blog.Infof("write local file %s user %s", localFile.To, localFile.User)
	//write file value
	_, err = file.Write([]byte(localFile.Value))
	if err != nil {
		blog.Errorf("write local file %s error %s", localFile.To, err.Error())
		return err
	}

	//set file user
	if localFile.User != "root" {
		u, err := user.Lookup(localFile.User)
		if err != nil {
			blog.Errorf("write localfile %s lookup user %s eror %s", localFile.To, localFile.User, err.Error())
		} else {
			uid, _ := strconv.Atoi(u.Uid)
			gid, _ := strconv.Atoi(u.Gid)
			err = file.Chown(uid, gid)
			if err != nil {
				blog.Errorf("chown local file %s user %s error %s", localFile.To, localFile.User, err.Error())
			}
		}
	}

	//set file chmod
	if localFile.Right == "rw" {
		err = file.Chmod(0664)
	} else {
		err = file.Chmod(0444)
	}

	if err != nil {
		blog.Errorf("chmod local file %s error %s", localFile.To, err.Error())
	}

	return nil
}
