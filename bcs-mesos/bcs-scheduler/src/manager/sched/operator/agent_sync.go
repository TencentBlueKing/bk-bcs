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

package operator

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	master "github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/mesosproto/mesos/master"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/schetypes"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-scheduler/src/manager/sched/client"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-scheduler/src/manager/store"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-scheduler/src/util"

	"github.com/golang/protobuf/proto"
)

// interval for synchronize agents from mesos master, seconds
const AGENT_SYNC_INTERVAL = 240

// Operate manager control message
type OperatorMsg struct {
	MsgType string
}

// Operate manager
type OperatorMgr struct {
	store store.Store
	// operator client to mesos master
	operatorClient *client.Client
	// at present, just for exit msg
	OperatorMsgQueue chan *OperatorMsg

	openCheck bool
}

// Create operate manager
func CreateOperatorMgr(store store.Store, client *client.Client) (*OperatorMgr, error) {

	mgr := &OperatorMgr{
		store:          store,
		operatorClient: client,
		openCheck:      false,
	}

	// create msg queue for events
	mgr.OperatorMsgQueue = make(chan *OperatorMsg, 32)

	return mgr, nil
}

func (mgr *OperatorMgr) stop() {
	blog.V(3).Infof("update agents: operatorMgr Stop...")
	close(mgr.OperatorMsgQueue)
}

// Send control message to operate manager
func (mgr *OperatorMgr) SendMsg(msg *OperatorMsg) error {

	blog.V(3).Infof("update agents: send an msg to operatorMgr")

	select {
	case mgr.OperatorMsgQueue <- msg:
	default:
		blog.Error("update agents: send an msg to operator manager fail")
		return fmt.Errorf("update agents: operator manager is busy now")
	}

	return nil
}

// the main loop for Operator manager
func OperatorManage(mgr *OperatorMgr) {
	blog.Info("update agents: goroutine start ...")

	blog.V(3).Infof("update agents: to sync agents sync mesos master to DB")
	mgr.UpdateMesosAgents()

	for {
		select {
		case req := <-mgr.OperatorMsgQueue:
			// at present, only exit msg
			blog.Info("update agents: receive msg (%s)", req.MsgType)
			if req.MsgType == "opencheck" {
				mgr.openCheck = true
			} else if req.MsgType == "closecheck" {
				mgr.openCheck = false
			} else if req.MsgType == "stop" {
				mgr.stop()
				blog.Info("update agents: goroutine finish!")
				return
			}
		case <-time.After(time.Second * time.Duration(AGENT_SYNC_INTERVAL)):
			blog.Info("update agents: to sync agents from mesos master to DB")
			mgr.UpdateMesosAgents()
		}
	}
}

func (mgr *OperatorMgr) UpdateMesosAgents() {
	blog.Info("update agents: begin")
	call := &master.Call{
		Type: master.Call_GET_AGENTS.Enum(),
	}

	req, err := proto.Marshal(call)
	if err != nil {
		blog.Error("update agents: query agentInfo proto.Marshal err: %s", err.Error())
		return
	}

	if mgr.operatorClient == nil {
		blog.Error("update agents: mgr.operatorClient is nil")
		return
	}

	resp, err := mgr.operatorClient.Send(req)
	if err != nil {
		blog.Error("update agents: query agentInfo Send err: %s", err.Error())
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		blog.Error("update agents: query agentInfo unexpected response statusCode: %d", resp.StatusCode)
		return
	}

	var response master.Response
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		blog.Error("update agents: Decode response failed: %s", err)
		return
	}

	blog.V(3).Infof("update agents: response msg type(%d)", response.GetType())
	agentInfo := response.GetGetAgents()
	if agentInfo == nil {
		blog.Warn("update agents: response Agents == nil")
		return
	}

	currSyncNum := len(agentInfo.Agents)
	blog.Info("update agents: current mesos agents count(%d)", currSyncNum)

	currTime := time.Now().Unix()

	var agent types.Agent
	for index, oneAgent := range agentInfo.Agents {
		innerIP, _ := util.GetMesosAgentInnerIP(oneAgent.GetAgentInfo().GetAttributes())
		if innerIP == "" {
			blog.Errorf("mesos agent(%s) don't have InnerIP attribute", oneAgent.GetAgentInfo().GetHostname())
			continue
		}

		blog.Infof("update agents: ===>agent[%d]: name(%s), info(%s)", index, innerIP, oneAgent.String())
		dbAgent, dbErr := mgr.store.FetchAgent(innerIP)
		if dbAgent == nil && dbErr == store.ErrNoFound {
			blog.Infof("update agents: new agent(%s) come to online", oneAgent.GetAgentInfo().GetHostname())
		}

		if dbAgent != nil {
			if reflect.DeepEqual(dbAgent.AgentInfo, oneAgent) {
				blog.Infof("new agent (%s) info no change", oneAgent.GetAgentInfo().GetHostname())
				continue
			}
		}

		agent.Key = innerIP
		agent.AgentInfo = oneAgent
		agent.LastSyncTime = currTime
		err = mgr.store.SaveAgent(&agent)
		if err != nil {
			blog.Error("update agents: save agent(%s) to db err:%s", innerIP, err.Error())
		} else {
			blog.Infof("update agents: save agent(%s) to db succ", innerIP)
		}
	}

	blog.Info("update agents: done ==> sync time(%d), mesos num(%d) ", currTime, currSyncNum)
	return
}
