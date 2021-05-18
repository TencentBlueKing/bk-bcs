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

package app

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	commtypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	bcstype "github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/schetypes"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-container-executor/container"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-container-executor/logs"

	"github.com/golang/protobuf/proto"
)

/*
 * FrameworkMessage handler
 * frameworkMessageEnvironmentUpdate(deprecated, no effect)
 * frameworkMessageFileUpload
 * frameworkMessageSignalExecute
**/

//frameworkMessageEnvironmentUpdate update Environment info in Container Runtime
func (executor *BcsExecutor) frameworkMessageEnvironmentUpdate(taskID string, env *bcstype.Msg_Env) error {
	//todo(developerJim): Env can not update after container started.
	//             this function must delete after bcs-scheduler clean this feature.
	executor.exeLock.Lock()
	defer executor.exeLock.Unlock()
	//get containerID list from local cache
	var containerList []string
	var err error
	if taskID == "" {
		//get all container
		containerList = executor.tasks.GetAllContainerID()
	} else {
		containerInfo := executor.tasks.GetContainerByTaskID(taskID)
		if containerInfo == nil {
			err = fmt.Errorf("task %s container not found", taskID)
			return err
		}
		containerList = append(containerList, containerInfo.ID)
	}
	if len(containerList) == 0 {
		err = fmt.Errorf("container list is empty")
		return err
	}
	//create Environment update shell command
	command := []string{"/bin/sh", "-c"}
	envshell := "\"export " + *env.Name + "=" + *env.Value + "\""
	command = append(command, envshell)

	for _, ID := range containerList {
		if runErr := executor.container.RunCommand(ID, command); runErr != nil {
			err = fmt.Errorf("Update Environment %s failed: %s", envshell, runErr.Error())
		}
	}

	return err
}

//frameworkMessageFileUpload upload file to Running Container
func (executor *BcsExecutor) frameworkMessageFileUpload(taskID string, fileInfo *bcstype.Msg_LocalFile) error {
	executor.exeLock.Lock()
	defer executor.exeLock.Unlock()
	logs.Infof("Executor Get FrameworkMessage LocalFile. File: %s", *fileInfo.To)
	//get container
	var containerList []string
	var err error
	if taskID == "" {
		//get all container
		containerList = executor.tasks.GetAllContainerID()
	} else {
		containerInfo := executor.tasks.GetContainerByTaskID(taskID)
		if containerInfo == nil {
			err = fmt.Errorf("task %s container not found", taskID)
			return err
		}

		containerList = append(containerList, containerInfo.ID)
	}
	if len(containerList) == 0 {
		err = fmt.Errorf("container list is empty")
		return err
	}

	//Upload file to Container
	for _, ID := range containerList {
		if copyErr := executor.copyFileToContainer(ID, fileInfo); copyErr != nil {
			err = fmt.Errorf("FrameworkMessage LocalFile copy err: %s", copyErr.Error())
			return err
		}
	}
	logs.Infoln(fmt.Sprintf("FrameworkMessage LocalFile %s copy success", *fileInfo.To))

	return nil
}

func (executor *BcsExecutor) frameworkMessageRemoteFile(taskID string, remote *bcstype.Msg_Remote) error {
	logs.Infof("Executor get framework RemoteFile %s for %s", *remote.From, *remote.To)
	local, err := executor.downloadRemoteFile(remote)
	if err != nil {
		logs.Errorf("framework message [RemoteFile] handle err, %s, do nothing", err)
		return err
	}
	logs.Infoln("framework message [RemoteFile] change to LocalFile")
	return executor.frameworkMessageFileUpload(taskID, local)
}

//frameworkMessageSignalExecute send signal to process in Container
func (executor *BcsExecutor) frameworkMessageSignalExecute(taskID string, singalInfo *bcstype.Msg_Signal) error {
	executor.exeLock.Lock()
	defer executor.exeLock.Unlock()
	logs.Infof(fmt.Sprintf("Executor get framework signal %d for %v", singalInfo.Signal, singalInfo.ProcessName))
	//get containerID list from local cache
	var containerList []string
	var err error
	if taskID == "" {
		//get all container
		containerList = executor.tasks.GetAllContainerID()
	} else {
		containerInfo := executor.tasks.GetContainerByTaskID(taskID)
		if containerInfo == nil {
			err = fmt.Errorf("task %s container not found", taskID)
			return err
		}
		containerList = append(containerList, containerInfo.ID)
	}
	if len(containerList) == 0 {
		err = fmt.Errorf("container list is empty")
		return err
	}
	//create Environment update shell command
	command := []string{"/bin/sh", "-c"}
	killSignal := strconv.Itoa(int(*singalInfo.Signal))
	kill := "\"killall -" + killSignal + *singalInfo.ProcessName + "\""
	command = append(command, kill)

	for _, ID := range containerList {
		logs.Infof("execute shell command ##%s## in container %s", command, ID)
		if runErr := executor.container.RunCommand(ID, command); runErr != nil {
			err = fmt.Errorf("Sending  %s failed: %s", command, runErr.Error())
		}
	}

	return err
}

func (executor *BcsExecutor) frameworkMessageUpdateResources(msg *bcstype.Msg_UpdateTaskResources) error {
	executor.exeLock.Lock()
	defer executor.exeLock.Unlock()

	var err error
	for _, update := range msg.Resources {
		logs.Infof("Executor get framework update taskid  %s resources cpu %f mem %f", *update.TaskId, *update.Cpu, *update.Mem)

		container := executor.tasks.GetContainerByTaskID(*update.TaskId)
		if container == nil {
			err = fmt.Errorf("task %s container not found", *update.TaskId)
			return err
		}

		err = executor.podInst.UpdateResources(container.ID, update)
		if err != nil {
			err = fmt.Errorf("update taskid %s resource error %s", *update.TaskId, err.Error())
			return err
		}

		logs.Infof("update taskid %s resource success", *update.TaskId)
	}

	logs.Infof("handler message update task resources done")
	return nil
}

func (executor *BcsExecutor) frameworkMessageCommitTask(msg *bcstype.Msg_CommitTask) error {
	executor.exeLock.Lock()
	defer executor.exeLock.Unlock()

	var err error
	for _, commit := range msg.Tasks {
		logs.Infof("Executor get framework commit taskid %s image %s", *commit.TaskId, *commit.Image)

		container := executor.tasks.GetContainerByTaskID(*commit.TaskId)
		if container == nil {
			err = fmt.Errorf("task %s container not found", *commit.TaskId)
			return err
		}

		err = executor.podInst.CommitImage(container.ID, *commit.Image)
		if err != nil {
			err = fmt.Errorf("commit taskid %s image %s error %s", *commit.TaskId, *commit.Image, err.Error())
			return err
		}

		logs.Infof("update taskid %s image %s success", *commit.TaskId, *commit.Image)

	}

	logs.Infof("handler message commit task image done")
	return err
}

func (executor *BcsExecutor) frameworkMessageCommandTask(msg *bcstype.RequestCommandTask) {
	var err error
	var resp *bcstype.ResponseCommandTask
	container := executor.tasks.GetContainerByTaskID(msg.TaskId)
	if container == nil {
		logs.Errorf("task %s container not found", msg.TaskId)
		resp = &bcstype.ResponseCommandTask{
			ID:          msg.ID,
			TaskId:      msg.TaskId,
			ContainerId: msg.ContainerId,
			Status:      commtypes.TaskCommandStatusFailed,
			Message:     fmt.Sprintf("task %s container not found", msg.TaskId),
		}
	} else {
		msg.ContainerId = container.ID
		resp, _ = executor.container.RunCommandV2(msg)
	}

	bcsMsg := &bcstype.BcsMessage{
		Type:                bcstype.Msg_Res_COMMAND_TASK.Enum(),
		ResponseCommandTask: resp,
	}
	by, _ := json.Marshal(bcsMsg)
	_, err = executor.driver.SendFrameworkMessage(string(by))
	if err != nil {
		logs.Errorf("send framework message error %s", err.Error())
	} else {
		logs.Infof("handler message command %s task %s done", msg.ID, msg.TaskId)
	}
}

/*
 * dataClass message handler
 * dataClassRemote
**/

const (
	defaultHTTPRequestTimeout = 120
)

//dataClassRemote handle bcs_remote message, download data from remote http url, push to created container
func (executor *BcsExecutor) dataClassRemote(taskInfo *container.BcsContainerTask, remote *bcstype.Msg_Remote) error {
	logs.Infof("BcsExecutor download %s from remote", *remote.From)
	local, err := executor.downloadRemoteFile(remote)
	if err != nil {
		return err
	}
	item := new(bcstype.BcsMessage)
	item.Type = bcstype.Msg_LOCALFILE.Enum()
	item.Local = local
	taskInfo.BcsMessages = append(taskInfo.BcsMessages, item)
	return nil
}

//dataClassRemoteEnv handle Msg_EnvRemote message, download data from remote http url, push to created container
func (executor *BcsExecutor) dataClassRemoteEnv(taskInfo *container.BcsContainerTask, remote *bcstype.Msg_EnvRemote) error {
	logs.Infof("BcsExecutor download Environment %s from %s", *remote.Name, *remote.From)

	//download content from remote http url.
	client := http.Client{
		Timeout: time.Duration(3 * time.Second),
	}
	response, err := client.Get(*remote.From)
	if err != nil {
		logs.Errorf("BcsExecutor download %s err: %s", *remote.From, err.Error())
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		logs.Errorf("BcsExecutor http get %s error, code: %d", *remote.From, response.StatusCode)
		return fmt.Errorf("http get %s response %s", *remote.From, response.Status)
	}
	data, readErr := ioutil.ReadAll(response.Body)
	if readErr != nil {
		logs.Errorf("BcsExecutor read %s reqeust err: %s", *remote.From, readErr.Error())
		return readErr
	}
	//convert data to base64
	item := container.BcsKV{
		Key:   *remote.Name,
		Value: string(data),
	}
	taskInfo.Env = append(taskInfo.Env, item)
	return nil
}

//downloadRemoteFile download remote file, change to local one
func (executor *BcsExecutor) downloadRemoteFile(remote *bcstype.Msg_Remote) (*bcstype.Msg_LocalFile, error) {
	local := new(bcstype.Msg_LocalFile)
	local.To = proto.String(*remote.To)
	local.User = proto.String(*remote.User)
	local.Right = proto.String(*remote.Right)

	//download content from remote http url.
	client := http.Client{
		Timeout: time.Duration(defaultHTTPRequestTimeout * time.Second),
	}
	request, reqErr := http.NewRequest("GET", *remote.From, nil)
	if reqErr != nil {
		logs.Errorf("BcsExecutor create http request for %s err, %s", *remote.From, reqErr.Error())
		return nil, reqErr
	}
	if len(*remote.RemoteUser) != 0 {
		request.SetBasicAuth(*remote.RemoteUser, *remote.RemotePasswd)
	}
	response, err := client.Do(request)
	if err != nil {
		logs.Errorf("BcsExecutor download %s err: %s", *remote.From, err.Error())
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		logs.Errorf("BcsExecutor http get %s error, code: %d", *remote.From, response.StatusCode)
		return nil, fmt.Errorf("http get %s response %s", *remote.From, response.Status)
	}
	data, readErr := ioutil.ReadAll(response.Body)
	if readErr != nil {
		logs.Errorf("BcsExecutor read %s reqeust err: %s", *remote.From, readErr.Error())
		return nil, readErr
	}
	//convert data to base64
	local.Base64 = proto.String(base64.StdEncoding.EncodeToString(data))
	logs.Infof("BcsExecutor download %s success", *remote.From)
	return local, nil
}
