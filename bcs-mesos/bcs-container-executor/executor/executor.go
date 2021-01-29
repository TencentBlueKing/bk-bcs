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
	"encoding/base64"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/mesosproto/mesos"
	conn "github.com/Tencent/bk-bcs/bcs-mesos/bcs-container-executor/connection"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-container-executor/logs"
	exec "github.com/Tencent/bk-bcs/bcs-mesos/bcs-container-executor/mesos/executor"

	"github.com/golang/protobuf/proto"
	"github.com/mesos/mesos-go/api/v0/upid"
	"github.com/pborman/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
	"golang.org/x/net/context"
	//"github.com/gogo/protobuf/test/fuzztests"
)

const (
	DefaultMetricsTextFile = "/data/bcs/export_data"
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

//GetAllEnvs get all info from environment
func (ee *DriverEnv) GetAllEnvs() error {
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
		//get MESOS_RECOVERY_TIMEOUT & MESOS_SUBSCRIPTION_BACKOFF_MAX
		ee.MesosRecoveryTimeout, _ = strconv.Atoi(os.Getenv("MESOS_RECOVERY_TIMEOUT"))
		ee.MesosSubscriptionBackoff, _ = strconv.Atoi(os.Getenv("MESOS_SUBSCRIPTION_BACKOFF_MAX"))
	}
	ee.MesosShutdownGracePeriod, _ = strconv.Atoi(os.Getenv("MESOS_EXECUTOR_SHUTDOWN_GRACE_PERIOD"))
	return nil
}

//DriverConfig hold all custom info for ExecutorDriver
type DriverConfig struct {
	Executor Executor //Executor interface
}

//BcsExecutorDriver BCS implementation for ExecutorDriver
type BcsExecutorDriver struct {
	lock              sync.RWMutex               //lock for status/data
	executor          Executor                   //custom executor
	exeEnv            *DriverEnv                 //executor environment required
	status            mesos.Status               //driver status
	connected         bool                       //flag for connection
	stateReConnected  bool                       //flag for re-connecting
	reConCxt          context.Context            //context for reconnection
	frameworkID       *mesos.FrameworkID         //scheduler frameworkid from environment
	agentID           *mesos.AgentID             //mesos slave ID form environment
	agentPID          *upid.UPID                 //mesos slave upid for identify
	executorID        *mesos.ExecutorID          //self executor id from environment
	connection        conn.Connection            //network connection to mesos slave
	tasks             map[string]*mesos.TaskInfo //key is uuid, task info map, send to slave when reregistered
	updates           *exec.Call_Update          //key is uuid, call_Update send to mesos slave when reregistered
	currentTaskStatus map[string]*mesos.TaskState
	stopCxt           context.Context    //context for cancel
	canceler          context.CancelFunc //function for cancel
}

//NewExecutorDriver create BcsExecutorDriver with ExecutorConfig
func NewExecutorDriver(bcsExe Executor) ExecutorDriver {
	//parse all config item from environment
	envs := new(DriverEnv)
	envErr := envs.GetAllEnvs()
	if envErr != nil {
		fmt.Fprintf(os.Stderr, "Get environments failed, %s\n", envErr.Error())
		return nil
	}
	slaveUpid, err := upid.Parse(envs.MesosSlavePID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Can't not parse mesos slave upid: %s\n", envs.MesosSlavePID)
		return nil
	}
	rootCxt, rootCancel := context.WithCancel(context.Background())
	stopCxt, _ := context.WithCancel(rootCxt)
	reCon, _ := context.WithCancel(rootCxt)
	return &BcsExecutorDriver{
		executor:         bcsExe,
		exeEnv:           envs,
		status:           mesos.Status_DRIVER_NOT_STARTED,
		connected:        false,
		stateReConnected: false,
		reConCxt:         reCon,
		frameworkID:      &mesos.FrameworkID{Value: proto.String(envs.MesosFrameworkID)},
		agentID:          &mesos.AgentID{Value: proto.String(envs.MesosSlaveID)},
		agentPID:         slaveUpid,
		executorID:       &mesos.ExecutorID{Value: proto.String(envs.MesosExecutorID)},
		tasks:            make(map[string]*mesos.TaskInfo),
		//updates:          make([]*exec.Call_Update,0),
		currentTaskStatus: make(map[string]*mesos.TaskState),
		stopCxt:           stopCxt,
		canceler:          rootCancel,
	}
}

//ExecutorID get ExecutorID from mesos slave
func (driver *BcsExecutorDriver) ExecutorID() string {
	return driver.exeEnv.MesosExecutorID
}

//Start the executor driver. This needs to be called before any
//other driver calls are made.
func (driver *BcsExecutorDriver) Start() (mesos.Status, error) {
	fmt.Fprintln(os.Stdout, "Starting BcsExecutorDriver...")
	if driver.status != mesos.Status_DRIVER_NOT_STARTED {
		return driver.status, fmt.Errorf("Unable start driver, expecting %s, but got %s", mesos.Status_DRIVER_NOT_STARTED, driver.status)
	}
	//create connection for driver
	//driver.connection = conn.NewFakeConnection()
	driver.connection = conn.NewConnection()
	//install all message handler, only one handler can execute each time
	var handlerLock sync.Mutex
	//subscribe
	driver.connection.Install(exec.Event_SUBSCRIBED, func(from *upid.UPID, event *exec.Event) {
		subs := event.GetSubscribed()
		//handlerLock.Lock()
		//defer handlerLock.Unlock()
		driver.subscribed(from, subs)
		executorSlaveConnection.Set(1)
	})
	//launch
	driver.connection.Install(exec.Event_LAUNCH, func(from *upid.UPID, event *exec.Event) {
		launch := event.GetLaunch()
		handlerLock.Lock()
		defer handlerLock.Unlock()
		driver.runTask(from, launch)
	})
	//launchgroup
	driver.connection.Install(exec.Event_LAUNCH_GROUP, func(from *upid.UPID, event *exec.Event) {
		launchGroup := event.GetLaunchGroup()
		handlerLock.Lock()
		defer handlerLock.Unlock()
		driver.runTaskGroup(from, launchGroup)
	})
	//kill
	driver.connection.Install(exec.Event_KILL, func(from *upid.UPID, event *exec.Event) {
		kill := event.GetKill()
		handlerLock.Lock()
		defer handlerLock.Unlock()
		driver.killTask(from, kill)
	})
	//framework message
	driver.connection.Install(exec.Event_MESSAGE, func(from *upid.UPID, event *exec.Event) {
		ack := event.GetMessage()
		handlerLock.Lock()
		defer handlerLock.Unlock()
		driver.frameworkMessage(from, ack)
	})
	//acknowledged
	driver.connection.Install(exec.Event_ACKNOWLEDGED, func(from *upid.UPID, event *exec.Event) {
		ack := event.GetAcknowledged()
		handlerLock.Lock()
		defer handlerLock.Unlock()
		driver.acknowledgementMessage(from, ack)
	})
	//shutdown
	driver.connection.Install(exec.Event_SHUTDOWN, func(from *upid.UPID, event *exec.Event) {
		//no message body for shutdown
		handlerLock.Lock()
		defer handlerLock.Unlock()
		driver.shutdown(from, event)
	})
	//error message
	driver.connection.Install(exec.Event_ERROR, func(from *upid.UPID, event *exec.Event) {
		err := event.GetError()
		handlerLock.Lock()
		defer handlerLock.Unlock()
		driver.frameworkError(from, err)
	})
	//http connection close callback
	driver.connection.Install(conn.Event_CONNECTION_CLOSE, func(from *upid.UPID, event *exec.Event) {
		//from & event all nil, do not use them
		handlerLock.Lock()
		defer handlerLock.Unlock()
		driver.connected = false
		driver.stateReConnected = true
		//create goroutine reconnect
		go driver.reconnectLoop()
		executorSlaveConnection.Set(0)
		//subscribe success, setting state in reConnected

	})
	//fix by developerJim, 2016-12-31
	//check http or https
	if driver.exeEnv.SSLEnabled {
		//todo(developerJim): add SSL INFO for connection
		return driver.status, fmt.Errorf("HTTPS Unimplementation")
	}
	if err := driver.connection.Start("http://"+driver.exeEnv.MesosAgentEndpoint, "/api/v1/executor"); err != nil {
		fmt.Fprintf(os.Stderr, "BcsExecutorDriver starting connection to slave failed: %s\n", err.Error())
		return driver.status, err
	}

	//todo(developerJim): get upid from connection

	//ready to send subcribe message to mesos slave
	if err := driver.subscribe(); err != nil {
		fmt.Fprintf(os.Stderr, "BcsExecutorDriver send Call_Subscribe message failed: %s\n", err.Error())
		return driver.status, err
	}
	//sending subscribe message success, wait launch or launchGroup
	driver.status = mesos.Status_DRIVER_RUNNING
	fmt.Fprintf(os.Stdout, "BcsExecutorDriver starting with ExecutorID: %s\n", driver.executorID.GetValue())
	//handle prometheus metrics to text file
	go driver.metricsToText()

	return driver.status, nil
}

//handle executor metrics to textfile
//module nodeexport report the textfile to prometheus
func (driver *BcsExecutorDriver) metricsToText() {
	err := os.MkdirAll(DefaultMetricsTextFile, 0755)
	if err != nil {
		blog.Errorf("mkdir dir %s error %s", DefaultMetricsTextFile, err.Error())
	}

	for {
		time.Sleep(time.Minute)

		mfs, err := prometheus.DefaultGatherer.Gather()
		if err != nil {
			blog.Errorf("prometheus gather %s", err.Error())
			continue
		}

		fPath := fmt.Sprintf("%s/%s.prom", DefaultMetricsTextFile, driver.executorID.GetValue())
		f, err := os.OpenFile(fPath, os.O_CREATE|os.O_RDWR, 0664)
		if err != nil {
			blog.Errorf("openfile %s error %s", fPath, err.Error())
			continue
		}
		f.Truncate(0)

		for _, mf := range mfs {
			_, err = expfmt.MetricFamilyToText(f, mf)
			if err != nil {
				blog.Errorf("file %s MetricFamilyToText error %s", fPath, err.Error())
			}
		}

		//close file
		f.Close()
	}

}

//Stop the executor driver.
//executor will exited
func (driver *BcsExecutorDriver) Stop() (mesos.Status, error) {
	fmt.Fprintln(os.Stdout, "Stop ExecutorDriver")
	if driver.status != mesos.Status_DRIVER_RUNNING {
		return driver.status, fmt.Errorf("Unable Stop, status is Not RUNNING")
	}
	//ready to stop connection
	if driver.connected {
		logs.Infoln("ExecutorDriver is under connection, wait slave reply acknowledged")
		//check all update info acknowledged
		checkTick := time.NewTicker(500 * time.Microsecond)
		defer checkTick.Stop()
		timeoutTick := time.NewTicker(5 * time.Second)
		defer timeoutTick.Stop()
		for driver.updates != nil && driver.connected {
			//if connection lost, no need to wait acknowledgement
			select {
			case <-timeoutTick.C:
				fmt.Fprintln(os.Stdout, "ExecutorDriver wait acknowledgement from slave timeout(5 seconds)")
				goto StopConnection
			case <-checkTick.C:
				if driver.updates != nil && driver.connected {
					fmt.Fprintln(os.Stdout, "ExecutorDriver accepts all acknowledgement, ready to exit")
					goto StopConnection
				}
			}
		}
	StopConnection:
		fmt.Fprintln(os.Stdout, "ExecutorDriver is stopping Connection...")
		driver.connection.Stop(true)
		driver.connected = false
	}
	logs.Infoln("ExecutorDriver connection to slave handle done, Stop flow done")
	driver.canceler()
	driver.status = mesos.Status_DRIVER_STOPPED
	return driver.status, nil
}

//Abort the driver so that no more callbacks can be made to the
//executor. The semantics of abort and stop have deliberately been
//separated so that code can detect an aborted driver (i.e., via
//the return status of ExecutorDriver.Join, see below), and
//instantiate and start another driver if desired (from within the
//same process ... although this functionality is currently not
//supported for executors).
func (driver *BcsExecutorDriver) Abort() (mesos.Status, error) {
	fmt.Fprintln(os.Stdout, "Abort ExecutorDriver")
	if driver.status != mesos.Status_DRIVER_RUNNING {
		return driver.status, fmt.Errorf("Unable Abort, status is Not RUNNING")
	}
	//ready to stop connection
	if driver.connected {
		fmt.Fprintln(os.Stdout, "ExecutorDriver is stopping Connection...")
		driver.connection.Stop(true)
		driver.connected = false
	}
	driver.canceler()
	driver.status = mesos.Status_DRIVER_ABORTED
	return driver.status, nil
}

//Join Waits for the driver to be stopped or aborted, possibly
//blocking the calling goroutine indefinitely. The return status of
//this function can be used to determine if the driver was aborted
//(see package mesos for a description of Status).
func (driver *BcsExecutorDriver) Join() (mesos.Status, error) {
	fmt.Println("join ExecutorDriver, wait for ExecutorDriver stop")
	if driver.status != mesos.Status_DRIVER_RUNNING {
		return driver.status, fmt.Errorf("Unable Abort, status is Not RUNNING")
	}
	//wait for stop signal
	select {
	case <-driver.stopCxt.Done():
		fmt.Fprintf(os.Stderr, "ExecutorDriver exit in join...\n")
		return driver.status, nil
	}
}

//Run Starts and immediately joins (i.e., blocks on) the driver.
func (driver *BcsExecutorDriver) Run() (mesos.Status, error) {
	status, err := driver.Start()
	if err != nil {
		return driver.Stop()
	}
	if status != mesos.Status_DRIVER_RUNNING {
		return status, fmt.Errorf("Unable to Run, expect status %s, but got %s", mesos.Status_DRIVER_RUNNING, status)
	}
	return driver.Join()
}

//SendStatusUpdate a status update to the framework scheduler, retrying as
//necessary until an acknowledgement has been received or the
//executor is terminated (in which case, a TASK_LOST status update
//will be sent). See Scheduler.StatusUpdate for more information
//about status update acknowledgements.
func (driver *BcsExecutorDriver) SendStatusUpdate(taskStatus *mesos.TaskStatus) (mesos.Status, error) {
	if driver.status != mesos.Status_DRIVER_RUNNING {
		fmt.Fprintf(os.Stderr, "Unable to SendStatusUpdate, expecting Status %s, but got %s", mesos.Status_DRIVER_RUNNING, driver.status)
		return driver.status, fmt.Errorf("ExecutorDriver status Not RUNNING")
	}
	if taskStatus.GetState() == mesos.TaskState_TASK_STAGING {
		err := fmt.Errorf("Executor is Not Allowed to send Staging Task")
		fmt.Fprintf(os.Stderr, "Send Error: %s, ExecutorDriver Abort\n", err.Error())
		if _, stopErr := driver.Abort(); stopErr != nil {
			fmt.Fprintf(os.Stderr, "ExecutorDriver Abort in SendStatusUpdate failed: %s", stopErr.Error())
		}
		return driver.status, err
	}
	//setting TaskStatus attributes
	ID := uuid.NewUUID()
	now := float64(time.Now().Unix())
	taskStatus.Timestamp = proto.Float64(now)
	taskStatus.AgentId = driver.agentID
	taskStatus.ExecutorId = driver.executorID
	taskStatus.Uuid = ID
	callUpdate := &exec.Call_Update{
		Status: taskStatus,
	}
	//fmt.Fprintf(os.Stdout, "ExecutorDriver send TaskStatus update %s\n", callUpdate.String())
	//create Call
	call := &exec.Call{
		FrameworkId: driver.frameworkID,
		ExecutorId:  driver.executorID,
		Type:        exec.Call_UPDATE.Enum(),
		Update:      callUpdate,
	}
	fmt.Fprintf(os.Stdout, "ExecutorDriver send task %s, UUID: %s\n", taskStatus.GetTaskId().GetValue(), ID.String())

	/*driver.lock.Lock()
	driver.updates = callUpdate
	driver.lock.Unlock()*/
	//send message to slave
	if err := driver.connection.Send(call, false); err != nil {
		logs.Errorf("ExecutorDriver send Call_Update failed: %s\n", err.Error())
		return driver.status, err
	}
	//executorID = taskgroupid
	taskgroupReportTotal.WithLabelValues(driver.executorID.GetValue()).Inc()

	return driver.status, nil
}

//SendFrameworkMessage send a message to the framework scheduler. These messages are
//best effort; do not expect a framework message to be
//retransmitted in any reliable fashion.
func (driver *BcsExecutorDriver) SendFrameworkMessage(data string) (mesos.Status, error) {
	logs.Infof("Sending Framework message: %s", data)
	if driver.status != mesos.Status_DRIVER_RUNNING {
		fmt.Fprintf(os.Stderr, "Unable to SendFramworkMessage, expecting status %s, but Got %s\n", mesos.Status_DRIVER_RUNNING, driver.status)
		return driver.status, fmt.Errorf("ExecutorDriver is Not Running")
	}
	//create Message
	call := &exec.Call{
		FrameworkId: driver.frameworkID,
		ExecutorId:  driver.executorID,
		Type:        exec.Call_MESSAGE.Enum(),
		Message: &exec.Call_Message{
			Data: []byte(data),
		},
	}
	if err := driver.connection.Send(call, false); err != nil {
		fmt.Fprintf(os.Stderr, "ExecutorDriver send Call_Message failed: %s\n", err.Error())
		return driver.status, err
	}
	return driver.status, nil
}

//subscribe send subscribe message to mesos slave
//check tasks & updates info, if these two map are not
//empty, ExecutorDriver must be disconnected with mesos slave,
//TaskInfo & TaskStatus will consider Unacknowledged, combine
//all info to Call_Subscribe
func (driver *BcsExecutorDriver) subscribe() error {
	subscribe := new(exec.Call_Subscribe)
	driver.lock.Lock()
	if len(driver.tasks) != 0 {
		for _, value := range driver.tasks {
			subscribe.UnacknowledgedTasks = append(subscribe.UnacknowledgedTasks, value)
		}
	}
	/*if driver.updates!=nil {
		subscribe.UnacknowledgedUpdates = append(subscribe.UnacknowledgedUpdates, driver.updates)
	}*/

	driver.lock.Unlock()
	call := &exec.Call{
		ExecutorId:  driver.executorID,
		FrameworkId: driver.frameworkID,
		Type:        exec.Call_SUBSCRIBE.Enum(),
		Subscribe:   subscribe,
	}

	if err := driver.connection.Send(call, true); err != nil {
		return err
	}
	fmt.Fprintln(os.Stderr, "ExecutorDriver is waiting callback...")
	return nil
}

/*
 * all register functions
 * only one callback handler can be call each time, see BcsExecutorDriver.Start()
 */

func (driver *BcsExecutorDriver) subscribed(from *upid.UPID, pbMsg *exec.Event_Subscribed) {
	agentInfo := pbMsg.GetAgentInfo()
	if driver.status == mesos.Status_DRIVER_ABORTED {
		fmt.Fprintf(os.Stdout, "ignoring subscribed message from slave %s because aborted\n", agentInfo.GetHostname())
		return
	}
	if driver.isStopped() {
		fmt.Fprintf(os.Stdout, "ignoring subcribed message from slave %s because stopped\n", agentInfo.GetHostname())
		return
	}
	//todo(developerJim): check frameworkID & ExecutorID are equal locals
	driver.connected = true
	fmt.Fprintf(os.Stdout, "ExecutorDriver registered with slave %s/%s success\n", agentInfo.GetHostname(), agentInfo.GetId().GetValue())
	executorInfo := pbMsg.GetExecutorInfo()
	frameworkInfo := pbMsg.GetFrameworkInfo()
	if driver.stateReConnected {
		driver.stateReConnected = false
		driver.executor.Reregistered(driver, agentInfo)
	} else {
		driver.executor.Registered(driver, executorInfo, frameworkInfo, agentInfo)
	}
}

//reconnect backoff strategy implementation for MESOS_RECOVERY_TIMEOUT, If it is
//not able to establish a subscription with the agent within this duration, it should gracefully exit.
func (driver *BcsExecutorDriver) reconnect(from *upid.UPID, pbMsg proto.Message) {
	fmt.Fprintln(os.Stderr, "Reconnect backoff strategy is Not Implemented!")
}

//runTask running one task
func (driver *BcsExecutorDriver) runTask(from *upid.UPID, pbMsg *exec.Event_Launch) {
	if driver.status == mesos.Status_DRIVER_ABORTED {
		fmt.Fprintln(os.Stdout, "Ignore Launch message from slave because ExecutorDriver Abort")
		return
	}
	if driver.status == mesos.Status_DRIVER_STOPPED {
		fmt.Fprintln(os.Stdout, "Ignore Launch message from slave because ExecutorDriver Stop.")
		return
	}
	task := pbMsg.GetTask()
	taskID := task.GetTaskId()
	agentID := task.GetAgentId()
	fmt.Fprintf(os.Stdout, "BcsExecutorDriver get task %s from slave %s.\n", taskID.GetValue(), agentID.GetValue())
	//check taskInfo is duplicated
	if _, exist := driver.tasks[taskID.GetValue()]; exist {
		fmt.Fprintf(os.Stderr, "BcsExecutorDriver get duplicated task from slave %s, Executor Exit\n", agentID.GetValue())
		os.Exit(255)
	}
	//recored and launch task
	driver.tasks[taskID.GetValue()] = task
	driver.executor.LaunchTask(driver, task)
}

//runTaskGroup running TaskGroup from slave
func (driver *BcsExecutorDriver) runTaskGroup(from *upid.UPID, pbMsg *exec.Event_LaunchGroup) {
	if driver.status == mesos.Status_DRIVER_ABORTED {
		fmt.Fprintln(os.Stdout, "Ignore LaunchTasks message from slave because ExecutorDriver Abort")
		return
	}
	if driver.status == mesos.Status_DRIVER_STOPPED {
		fmt.Fprintln(os.Stdout, "Ignore LaunchTasks message from slave because ExecutorDriver Stop.")
		return
	}
	taskGroup := pbMsg.GetTaskGroup()
	tasks := taskGroup.GetTasks()
	if len(tasks) == 0 {
		//Error, No tasks in LaunchGroup, ExecutorDriver exit
		fmt.Fprintln(os.Stderr, "ExecutorDriver Get 0 task in LaunchGroup message.")
		os.Exit(255)
	}
	agentID := tasks[0].GetAgentId()
	fmt.Fprintf(os.Stdout, "BcsExecutorDriver get %d tasks from salve %s\n", len(tasks), agentID.GetValue())
	//check existence in local taskInfo
	for _, task := range tasks {
		if _, exist := driver.tasks[task.GetTaskId().GetValue()]; exist {
			fmt.Fprintf(os.Stderr, "BcsExecutorDriver get duplicated TaskId %s from slave\n", task.GetTaskId())
			os.Exit(255)
		}
		//update TaskInfo in local
		driver.tasks[task.GetTaskId().GetValue()] = task
	}
	//ready to post Executor
	driver.executor.LaunchTaskGroup(driver, taskGroup)
}

//kill task
func (driver *BcsExecutorDriver) killTask(from *upid.UPID, pbMsg *exec.Event_Kill) {
	if driver.status == mesos.Status_DRIVER_ABORTED {
		fmt.Fprintln(os.Stdout, "Ignore Kill message from slave because ExecutorDriver Abort")
		return
	}
	if driver.status == mesos.Status_DRIVER_STOPPED {
		fmt.Fprintln(os.Stdout, "Ignore Kill message from slave because ExecutorDriver Stop")
		return
	}
	taskID := pbMsg.GetTaskId()
	fmt.Fprintf(os.Stdout, "BcsExecutorDriver ready to kill task %s\n", taskID.GetValue())
	driver.executor.KillTask(driver, taskID)
}

//acknowledgementMessage acknowledge task status for scheduler
func (driver *BcsExecutorDriver) acknowledgementMessage(from *upid.UPID, pbMsg *exec.Event_Acknowledged) {
	if driver.status == mesos.Status_DRIVER_ABORTED {
		fmt.Fprintln(os.Stdout, "Ignore Acknowledged message from slave because ExecutorDriver Abort")
		return
	}
	taskID := pbMsg.GetTaskId()
	updateUuID := uuid.UUID(pbMsg.GetUuid())
	if driver.status == mesos.Status_DRIVER_STOPPED {
		fmt.Fprintf(
			os.Stdout,
			"Ignore Acknowledged taskId %s, uuid %s from slave because ExecutorDriver Stop",
			taskID.GetValue(),
			updateUuID.String(),
		)
		return
	}
	driver.lock.Lock()
	delete(driver.tasks, taskID.GetValue())
	driver.lock.Unlock()

	fmt.Fprintf(os.Stdout, "ExecutorDriver get acknowledgement from slave, taskId %s, uuid %s\n", taskID.GetValue(), updateUuID.String())
	taskgroupAckTotal.WithLabelValues(driver.executorID.GetValue()).Inc()

	//clean local Unacknowledged info with taskId & uuid
	//todo(developerJim): how to handle if missing TaskInfo & uuid in local map
	/*driver.lock.Lock()
	defer driver.lock.Unlock()
	delete(driver.tasks, taskID.GetValue())
	//delete(driver.updates, updateUuID.String())
	//No need to notify Executor
	uid := uuid.UUID(driver.updates.Status.GetUuid())

	if uid.String() == updateUuID.String(){
		driver.updates = nil
	}*/

	return
}

//receive framework message from scheduler
//the task complete is sync
func (driver *BcsExecutorDriver) frameworkMessage(from *upid.UPID, pbMsg *exec.Event_Message) {
	if driver.status == mesos.Status_DRIVER_ABORTED {
		fmt.Fprintln(os.Stdout, "Ignore Message info because ExecutorDriver Abort")
		return
	}
	if driver.status == mesos.Status_DRIVER_STOPPED {
		fmt.Fprintln(os.Stdout, "Ignore Message info because ExecutorDriver Stop")
		return
	}
	data, err := base64.StdEncoding.DecodeString(string(pbMsg.GetData()))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Decode base64 FrameworkMessage err: %s\n", err.Error())
		return
	}
	driver.executor.FrameworkMessage(driver, string(data))
}

//shutdown the executor, and killall containers
func (driver *BcsExecutorDriver) shutdown(from *upid.UPID, pbMsg *exec.Event) {
	if driver.status == mesos.Status_DRIVER_ABORTED {
		fmt.Fprintln(os.Stdout, "Ignore Shutdown info because ExecutorDriver Abort")
		return
	}
	if driver.status == mesos.Status_DRIVER_STOPPED {
		fmt.Fprintln(os.Stdout, "Ignore Shutdown info because ExecutorDriver Stop")
		return
	}
	fmt.Fprintln(os.Stdout, "BcsExecutorDriver asked to shutdown")
	driver.executor.Shutdown(driver)
	driver.Stop()
}

func (driver *BcsExecutorDriver) frameworkError(from *upid.UPID, pbMsg *exec.Event_Error) {
	if driver.status == mesos.Status_DRIVER_ABORTED {
		fmt.Fprintln(os.Stdout, "Ignore Error message because ExecutorDriver Abort")
		return
	}
	fmt.Fprintln(os.Stdout, "BcsExecutorDriver received Error message")
	driver.executor.Error(driver, pbMsg.GetMessage())
}

func (driver *BcsExecutorDriver) networkError(from *upid.UPID, pbMsg proto.Message) {
	fmt.Fprintln(os.Stderr, "ExecutorDriver Not Implemented")
}

//driver status access

//Status return driver status
func (driver *BcsExecutorDriver) Status() mesos.Status {
	driver.lock.RLock()
	defer driver.lock.RUnlock()
	return driver.status
}

//IsRunning check driver is running
func (driver *BcsExecutorDriver) IsRunning() bool {
	driver.lock.RLock()
	defer driver.lock.RUnlock()
	return driver.status == mesos.Status_DRIVER_RUNNING
}

//isStopped check driver is stopped
func (driver *BcsExecutorDriver) isStopped() bool {
	driver.lock.RLock()
	defer driver.lock.RUnlock()
	return driver.status == mesos.Status_DRIVER_STOPPED
}

//isStopped check driver is stopped
func (driver *BcsExecutorDriver) isConnected() bool {
	driver.lock.RLock()
	defer driver.lock.RUnlock()
	return driver.connected
}

//reconnectLoop create loop for reconnect to slave when long connection is down
func (driver *BcsExecutorDriver) reconnectLoop() {
	fmt.Fprintln(os.Stdout, "##################ExecutorDriver enter reconnecting loop, retry every 1 seconds#####################")
	tick := time.NewTicker(time.Second * 1)
	//todo(developerJim): add retry handler according MESOS_RECOVERY_TIMEOUT,
	//a suitable backoff strategy must be implemented later
	//now is only subscribe simply, actually, we need method reconnect
	for {
		select {
		case <-driver.reConCxt.Done():
			fmt.Fprintln(os.Stdout, "ExecutorDriver ask to exit in reconnection loop")
			return
		case now := <-tick.C:
			fmt.Fprintf(os.Stdout, "ExecutorDriver tick: %s, ready to resubscribed......\n", now.String())
			if !driver.stateReConnected {
				fmt.Fprintln(os.Stdout, "ExecutorDriver is not under reconnect status, close reconnection loop.")
				return
			}
			if err := driver.subscribe(); err != nil {
				fmt.Fprintf(os.Stderr, "ExecutorDriver send Call_Subscribe message in reconnection loop failed: %s, wait for next tick\n", err.Error())
			} else {
				fmt.Fprintln(os.Stdout, "ExecutorDriver Send Subscribe success, wait for reply in 5 seconds")
			}
		}
	}
}
