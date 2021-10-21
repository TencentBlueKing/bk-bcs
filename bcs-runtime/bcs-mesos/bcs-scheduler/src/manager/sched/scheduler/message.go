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

package scheduler

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/mesosproto/mesos"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/mesosproto/sched"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/schetypes"

	"github.com/golang/protobuf/proto"
)

//SendMessage send msg by scheduler to executor, msg is handled by master with MESSAGE call
func (s *Scheduler) SendMessage(taskGroup *types.TaskGroup, msg []byte) (*http.Response, error) {
	blog.Info("Send %d msg(%s) to executor %s", len(msg), msg, taskGroup.ID)
	call := &sched.Call{
		FrameworkId: s.framework.GetId(),
		Type:        sched.Call_MESSAGE.Enum(),
		Message: &sched.Call_Message{
			AgentId:    &mesos.AgentID{Value: proto.String(taskGroup.AgentID)},
			ExecutorId: &mesos.ExecutorID{Value: proto.String(taskGroup.ID)},
			Data:       []byte(base64.StdEncoding.EncodeToString(msg)),
		},
	}

	return s.send(call)
}

//SendBcsMessage send bcs message to TaskGroup
func (s *Scheduler) SendBcsMessage(taskGroup *types.TaskGroup, bcsMsg *types.BcsMessage) (*types.BcsMessage, error) {
	if taskGroup.Status != types.TASKGROUP_STATUS_RUNNING {
		return nil, fmt.Errorf("taskgroup %s must be running", taskGroup.ID)
	}

	bcsMsg.Id = time.Now().Unix()
	bcsMsg.Status = types.Msg_Status_Staging
	bcsMsg.CreateTime = time.Now().Unix()
	bcsMsg.TaskGroupId = taskGroup.ID

	msgStr, err := json.Marshal(bcsMsg)
	if err != nil {
		blog.Error("Send msg to executor %s fail for %s", taskGroup.ID, err.Error())
		return nil, err
	}

	resp, err := s.SendMessage(taskGroup, msgStr)
	if err != nil {
		return nil, fmt.Errorf("Send msg to executor %s fail for %s", taskGroup.ID, err.Error())
	}

	if resp != nil && resp.StatusCode != http.StatusAccepted {
		err = fmt.Errorf("Send msg to executor %s fail for status code %d received", taskGroup.ID, resp.StatusCode)
		bcsMsg.Status = types.Msg_Status_Failed
		bcsMsg.Message = err.Error()
	}

	//if taskGroup.BcsMessages==nil {
	//	taskGroup.BcsMessages = make(map[int64]*types.BcsMessage)
	//}
	//taskGroup.BcsMessages[bcsMsg.Id] = bcsMsg

	taskGroup.BcsEventMsg = bcsMsg

	//save taskGroup into zk, in this function, task will alse be saved
	if err = s.store.SaveTaskGroup(taskGroup); err != nil {
		blog.Error("status report: save taskgroup: %s information into db failed! err:%s", taskGroup.ID, err.Error())
	}

	return bcsMsg, err
}

//SendLocalFile send local file to executor
func (s *Scheduler) SendLocalFile(taskGroup *types.TaskGroup, ctxBase64, to, right, user string) (*types.BcsMessage, error) {
	msg := &types.Msg_LocalFile{
		To:     proto.String(to),
		Right:  proto.String(right),
		User:   proto.String(user),
		Base64: proto.String(ctxBase64),
	}

	//TODO specify a specific Task with TaskID that is not used now
	bcsMsg := &types.BcsMessage{
		Type:  types.Msg_LOCALFILE.Enum(),
		Local: msg,
		//TaskID : null,
	}

	return s.SendBcsMessage(taskGroup, bcsMsg)
}

//SendRemoteFile send remote file to executor
func (s *Scheduler) SendRemoteFile(taskGroup *types.TaskGroup, from, to, right, user string) (*types.BcsMessage, error) {
	msg := &types.Msg_Remote{
		To:    proto.String(to),
		Right: proto.String(right),
		User:  proto.String(user),
		From:  proto.String(from),
	}

	//TODO specify a specific Task with TaskID that is not used now
	bcsMsg := &types.BcsMessage{
		Type:   types.Msg_REMOTE.Enum(),
		Remote: msg,
		//TaskID : null,
	}

	return s.SendBcsMessage(taskGroup, bcsMsg)
}

//SendSignal send any user specifyed signal to the executor
func (s *Scheduler) SendSignal(taskGroup *types.TaskGroup, signal uint32) (*types.BcsMessage, error) {
	msg := &types.Msg_Signal{
		Signal: proto.Uint32(signal),
	}

	//TODO specify a specific Task with TaskID that is not used now
	bcsMsg := &types.BcsMessage{
		Type: types.Msg_SIGNAL.Enum(),
		Sig:  msg,
		//TaskID : null,
	}

	return s.SendBcsMessage(taskGroup, bcsMsg)
}

//SendEnv send env to the executor, name is the env value key,
//replace indicates whether to replace an existing one if it is exist already
//if replace is false, addition or creation is the default behavior
func (s *Scheduler) SendEnv(taskGroup *types.TaskGroup, name, value string /*replace bool*/) (*types.BcsMessage, error) {
	msg := &types.Msg_Env{
		Name:  proto.String(name),
		Value: proto.String(value),
		//Rep:   replace,
	}

	//TODO specify a specific Task with TaskID that is not used now
	bcsMsg := &types.BcsMessage{
		Type: types.Msg_ENV.Enum(),
		Env:  msg,
		//TaskID : null,
	}

	return s.SendBcsMessage(taskGroup, bcsMsg)
}

//ProcessCommandMessage handle response bcs message
func (s *Scheduler) ProcessCommandMessage(bcsMsg *types.BcsMessage) {

	if bcsMsg.ResponseCommandTask == nil {
		blog.Error("procss command message, but data empty")
		return
	}
	s.store.LockCommand(bcsMsg.ResponseCommandTask.ID)
	defer s.store.UnLockCommand(bcsMsg.ResponseCommandTask.ID)

	cmdID := bcsMsg.ResponseCommandTask.ID
	taskID := bcsMsg.ResponseCommandTask.TaskId
	blog.Info("procss command message: command(%s), task(%s)", cmdID, taskID)

	command, err := s.store.FetchCommand(cmdID)
	if err != nil {
		blog.Warn("get command(%s) err: %s", cmdID)
		return
	}

	exist := false
	for _, taskGroup := range command.Status.Taskgroups {
		for _, task := range taskGroup.Tasks {
			if taskID == task.TaskId {
				task.Status = bcsMsg.ResponseCommandTask.Status
				task.Message = bcsMsg.ResponseCommandTask.Message
				task.CommInspect = bcsMsg.ResponseCommandTask.CommInspect
				blog.Info("update command(%s) task(%s:%s:%s)", cmdID, taskID, task.Status, task.Message)
				exist = true
				break
			}
		}

		if exist {
			break
		}
	}

	if exist {
		err := s.store.SaveCommand(command)
		if err != nil {
			blog.Error("process command message: command(%s), task(%s) update failed %s", cmdID, taskID, err.Error())
		} else {
			blog.Infof("process command message: command(%s), task(%s) updated", cmdID, taskID)
		}

	} else {
		blog.Error("process command message: command(%s), task(%s) not exist", cmdID, taskID)
	}

	return
}
