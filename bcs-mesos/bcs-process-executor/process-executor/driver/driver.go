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

package driver

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/mesosproto/mesos"
	bcstype "github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/schetypes"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-process-executor/process-executor/executor"
	protoExec "github.com/Tencent/bk-bcs/bcs-mesos/bcs-process-executor/process-executor/protobuf/executor"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-process-executor/process-executor/types"

	"github.com/golang/protobuf/proto"
	"github.com/mesos/mesos-go/api/v0/upid"
)

//DriverEnv The following environment variables are set by the agent that can be
//    used by the executor upon startup:
//MESOS_FRAMEWORK_ID: FrameworkID of the scheduler needed as part of the SUBSCRIBE call.
//MESOS_EXECUTOR_ID: ExecutorID of the executor needed as part of the SUBSCRIBE call.
//MESOS_DIRECTORY: Path to the working directory for the executor on the host filesystem(deprecated).
//MESOS_SANDBOX: Path to the mapped sandbox inside of the container (determined by the
//    agent flag sandbox_directory) for either mesos container with image or docker container.
//    For the case of command task without image specified, it is the path to the sandbox
//    on the host filesystem, which is identical to MESOS_DIRECTORY. MESOS_DIRECTORY
//    is always the sandbox on the host filesystem.
//MESOS_AGENT_ENDPOINT: agent endpoint i.e. ip:port to be used by the executor to connect
//    to the agent.
//MESOS_CHECKPOINT: If set to true, denotes that framework has checkpointing enabled.
//MESOS_EXECUTOR_SHUTDOWN_GRACE_PERIOD: Amount of time the agent would wait for an
//    executor to shut down (e.g., 60secs, 3mins etc.) after sending a SHUTDOWN event.
//If MESOS_CHECKPOINT is set i.e. when framework checkpointing is enabled, the following
//    additional variables are also set that can be used by the executor for retrying
//    upon a disconnection with the agent:
//MESOS_RECOVERY_TIMEOUT: The total duration that the executor should spend retrying
//    before shutting itself down when it is disconnected from the agent (e.g., 15mins,
//    5secs etc.). This is configurable at agent startup via the flag --recovery_timeout.
//MESOS_SUBSCRIPTION_BACKOFF_MAX: The maximum backoff duration to be used by the executor
//    between two retries when disconnected (e.g., 250ms, 1mins etc.). This is configurable
//    at agent startup via the flag --executor_reregistration_timeout.
type DriverEnv struct {
	MesosSlavePID            string //agent slave pid
	MesosSlaveID             string //agent slave uniq id
	MesosAgentEndpoint       string //agent ip:port endpoint to connect to the agent
	MesosFrameworkID         string //frameworkid from agent
	MesosExecutorID          string //exector id from agent
	SSLEnabled               bool   //true is agent enable https
	MesosSandBox             string //Path to the mapped sandbox inside of the container
	MesosCheckpoint          bool   //If set to true, denotes that framework has checkpointing enabled
	MesosRecoveryTimeout     int    //The total duration that the executor should spend retrying before shutting it self down when it is disconnected from the agent
	MesosSubscriptionBackoff int    //The maximum backoff duration between two retries when disconnected
	MesosShutdownGracePeriod int    //Amount of time the agent would wait for an executor to shut down (e.g., 60secs, 3mins etc.) after sending a SHUTDOWN event
}

//getAllEnvs get all info from environment
func (ee *DriverEnv) getAllEnvs() error {
	ee.MesosSlavePID = os.Getenv("MESOS_SLAVE_PID")
	if ee.MesosSlavePID == "" {
		return fmt.Errorf("Expecting MESOS_SLAVE_PID to be set in the environment")
	}

	ee.MesosSlaveID = os.Getenv("MESOS_SLAVE_ID")
	ee.MesosFrameworkID = os.Getenv("MESOS_FRAMEWORK_ID")
	if ee.MesosFrameworkID == "" {
		return fmt.Errorf("Expecting MESOS_FRAMEWORK_ID to be set in the environment")
	}

	ee.MesosExecutorID = os.Getenv("MESOS_EXECUTOR_ID")
	if ee.MesosExecutorID == "" {
		return fmt.Errorf("Expecting MESOS_EXECUTOR_ID to be set in the environment")
	}

	value := os.Getenv("SSL_ENABLED")
	if value == "1" || value == "true" {
		ee.SSLEnabled = true
	}

	ee.MesosSandBox = os.Getenv("MESOS_SANDBOX")
	if ee.MesosSandBox == "" {
		return fmt.Errorf("Expecting MESOS_SANDBOX to be set in the environment")
	}

	ee.MesosAgentEndpoint = os.Getenv("MESOS_AGENT_ENDPOINT")
	if ee.MesosAgentEndpoint == "" {
		return fmt.Errorf("Expecting MESOS_AGENT_ENDPOINT to be set in the environment")
	}

	ee.MesosCheckpoint = false
	checkPoint := os.Getenv("MESOS_CHECKPOINT")
	if checkPoint == "1" || checkPoint == "true" {
		ee.MesosCheckpoint = true
		ee.MesosRecoveryTimeout, _ = strconv.Atoi(os.Getenv("MESOS_RECOVERY_TIMEOUT"))
		ee.MesosSubscriptionBackoff, _ = strconv.Atoi(os.Getenv("MESOS_SUBSCRIPTION_BACKOFF_MAX"))
	}

	ee.MesosShutdownGracePeriod, _ = strconv.Atoi(os.Getenv("MESOS_EXECUTOR_SHUTDOWN_GRACE_PERIOD"))
	return nil
}

const (
	SlaveUri = "/api/v1/executor"

	KeepAliveConn    = true
	NotKeepAliveConn = false
)

//ExecutorDriver BCS implementation for ExecutorDriver
type ExecutorDriver struct {
	sync.RWMutex
	executor  executor.Executor //custom executor
	status    mesos.Status      //driver status
	conn      *HttpConnection
	connected bool //flag for connection

	slaveEndpoint string
	slaveUri      string

	frameworkID *mesos.FrameworkID //scheduler frameworkid from environment
	agentID     *mesos.AgentID     //mesos slave ID form environment
	agentPID    *upid.UPID         //mesos slave upid for identify
	executorID  *mesos.ExecutorID  //self executor id from environment
	exeEnv      *DriverEnv         //executor environment required

	cxt    context.Context    //context for cancel
	cancel context.CancelFunc //function for cancel
}

func NewExecutorDriver(cxt context.Context, bcsExec executor.Executor) (*ExecutorDriver, error) {
	envs := &DriverEnv{}
	err := envs.getAllEnvs()
	if err != nil {
		blog.Errorf("ExecutorDriver get envs error %s", err.Error())
		return nil, err
	}

	slaveUpid, err := upid.Parse(envs.MesosSlavePID)
	if err != nil {
		blog.Errorf("parse mesosslave pid %s error %s", envs.MesosSlavePID, err.Error())
		return nil, err
	}

	driver := &ExecutorDriver{
		executor:      bcsExec,
		exeEnv:        envs,
		status:        mesos.Status_DRIVER_NOT_STARTED,
		frameworkID:   &mesos.FrameworkID{Value: proto.String(envs.MesosFrameworkID)},
		agentID:       &mesos.AgentID{Value: proto.String(envs.MesosSlaveID)},
		agentPID:      slaveUpid,
		executorID:    &mesos.ExecutorID{Value: proto.String(envs.MesosExecutorID)},
		slaveEndpoint: fmt.Sprintf("http://%s", envs.MesosAgentEndpoint),
		slaveUri:      SlaveUri,
	}

	driver.conn = NewHttpConnection(driver.slaveEndpoint, driver.slaveUri)
	driver.cxt, driver.cancel = context.WithCancel(cxt)

	return driver, nil
}

func (driver *ExecutorDriver) Start() {
	err := driver.subscribe()
	if err != nil {
		blog.Errorf("subcribe mesos slave error %s, and exit", err.Error())
		os.Exit(1)
	}

	//register callback function update task status
	updateFunc := types.UpdateTaskFunc(driver.UpdateTaskStatus)
	driver.executor.RegisterCallbackFunc(types.CallbackFuncUpdateTask, updateFunc)
}

//subscribe send subscribe message to mesos slave
func (driver *ExecutorDriver) subscribe() error {
	subscribe := new(protoExec.Call_Subscribe)
	call := &protoExec.Call{
		ExecutorId:  driver.executorID,
		FrameworkId: driver.frameworkID,
		Type:        protoExec.Call_SUBSCRIBE.Enum(),
		Subscribe:   subscribe,
	}

	resp, err := driver.conn.Send(call, KeepAliveConn)
	if err != nil {
		return err
	}
	go driver.recvLoop(resp)

	return nil
}

func (driver *ExecutorDriver) subscribed(from *upid.UPID, pbMsg *protoExec.Event_Subscribed) {
	driver.Lock()
	blog.Infof("ExecutorDriver subscribed success")
	driver.connected = true
	driver.Unlock()
}

func (driver *ExecutorDriver) launchTaskgroup(from *upid.UPID, pbMsg *protoExec.Event_LaunchGroup) {
	blog.Infof("ExecutorDriver launch taskgroup start...")
	taskgroup := pbMsg.GetTaskGroup()
	driver.executor.LaunchTaskgroup(taskgroup)
}

//recvLoop recv message from mesos
func (driver *ExecutorDriver) recvLoop(response *http.Response) {
	defer response.Body.Close()

	var err error
	jsonDecoder := json.NewDecoder(NewReader(response.Body))
	for {
		select {
		case <-driver.cxt.Done():
			blog.Infof("stop recvLoop mesos response")
			return
		default:
			event := new(protoExec.Event)
			err = jsonDecoder.Decode(event)
			if err != nil {
				driver.Lock()
				driver.connected = false
				go driver.loopSubcribedMesosSlave()
				driver.Unlock()
				return
			}

			switch event.GetType() {
			case protoExec.Event_SUBSCRIBED:
				driver.subscribed(nil, event.GetSubscribed())

			case protoExec.Event_UNKNOWN:
				blog.Infof("recv event type UNKNOWN")

			case protoExec.Event_LAUNCH_GROUP:
				driver.launchTaskgroup(nil, event.GetLaunchGroup())

			case protoExec.Event_SHUTDOWN:
				blog.Infof("driver receive event shutdown")
				driver.executor.Shutdown()

			case protoExec.Event_ACKNOWLEDGED:
				driver.acknowledged(nil, event.GetAcknowledged())

			default:
				blog.Errorf("event type %s is invalid", event.GetType().String())
			}
		}

	}
}

func (driver *ExecutorDriver) handleMessage(from *upid.UPID, pbMsg *protoExec.Event_Message) {
	data, err := base64.StdEncoding.DecodeString(string(pbMsg.GetData()))
	if err != nil {
		blog.Errorf("base64 DecodeString Message error %s", err.Error())
		return
	}

	//parse to BcsMessage
	var bcsMessage bcstype.BcsMessage
	if err := json.Unmarshal([]byte(data), &bcsMessage); err != nil {
		blog.Errorf("unmarshal data %s to BcsMessage error %s", data, err.Error())
		return
	}

	switch *bcsMessage.Type {
	case bcstype.Msg_RELOAD_TASK:
		err = driver.executor.ReloadTasks()

	case bcstype.Msg_RESTART_TASK:
		err = driver.executor.RestartTasks()

	default:
		blog.Errorf("message type %d is invalid", *bcsMessage.Type)
	}

	if err != nil {
		blog.Errorf("handle bcs message type %s error %s", *bcsMessage.Type, err.Error())
	}

	return
}

func (driver *ExecutorDriver) acknowledged(from *upid.UPID, pbMsg *protoExec.Event_Acknowledged) {
	//blog.Infof("get slave acknowledged taskid %s uuid %s",pbMsg.GetTaskId().GetValue(),string(pbMsg.GetUuid()))
	driver.executor.AckTaskStatusMessage(pbMsg.GetTaskId().GetValue(), pbMsg.GetUuid())
}

func (driver *ExecutorDriver) loopSubcribedMesosSlave() {
	blog.Infof("loop subcribe mesos slave")

	ticker := time.NewTicker(time.Second)
	for {
		select {
		case <-driver.cxt.Done():
			blog.Infof("stop loop subcribe mesos slave")
			return

		case <-ticker.C:
			err := driver.subscribe()
			if err != nil {
				blog.Errorf("subcribe mesos slave error %s", err.Error())
			} else {
				blog.Infof("subcribe mesos slave success")
				return
			}
		}
	}
}

func (driver *ExecutorDriver) UpdateTaskStatus(status *mesos.TaskStatus) error {
	status.AgentId = driver.agentID
	status.ExecutorId = driver.executorID
	callUpdate := &protoExec.Call_Update{
		Status: status,
	}

	call := &protoExec.Call{
		FrameworkId: driver.frameworkID,
		ExecutorId:  driver.executorID,
		Type:        protoExec.Call_UPDATE.Enum(),
		Update:      callUpdate,
	}

	_, err := driver.conn.Send(call, NotKeepAliveConn)
	return err
}
