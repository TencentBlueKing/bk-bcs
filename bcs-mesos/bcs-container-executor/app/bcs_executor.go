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
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/util"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/mesosproto/mesos"
	bcstype "github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/schetypes"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-container-executor/container"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-container-executor/container/cni"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-container-executor/container/cnm"
	exec "github.com/Tencent/bk-bcs/bcs-mesos/bcs-container-executor/executor"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-container-executor/extendedresource"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-container-executor/healthcheck"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-container-executor/logs"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-container-executor/network"
	cninet "github.com/Tencent/bk-bcs/bcs-mesos/bcs-container-executor/network/cni"
	cnmnet "github.com/Tencent/bk-bcs/bcs-mesos/bcs-container-executor/network/cnm"

	"github.com/golang/protobuf/proto"
	"golang.org/x/net/context"
)

//ContainerCheckTicker duration for checking Running Container Status
const ContainerCheckTicker = 1

const (
	// ContainerNamePrefix docker container name prefix
	ContainerNamePrefix = "bcs-container-"
	// ContainerCoreFilePrefix container corefile directory
	ContainerCoreFilePrefix = "/data/corefile"
	// ExecutorStatus_NOTRUNNING not running state
	ExecutorStatus_NOTRUNNING = "NotRunning"
	// ExecutorStatus_LAUNCHING launching state
	ExecutorStatus_LAUNCHING = "Launching"
	// ExecutorStatus_RUNNING running state
	ExecutorStatus_RUNNING = "Running"
	// ExecutorStatus_KILLING killing state
	ExecutorStatus_KILLING = "Killing"
	// ExecutorStatus_SHUTDOWN shutdown state
	ExecutorStatus_SHUTDOWN = "Shutdown"
)

//NewBcsExecutor create Executor instance
func NewBcsExecutor(flag *CommandFlags) exec.Executor {
	//create container runtime client
	docker := container.NewDockerContainer(flag.DockerSocket, flag.User, flag.Passwd)
	if docker == nil {
		logs.Errorln("BcsExecutor create docker client failed")
		return nil
	}
	//create stop channel
	eCxt, eCancel := context.WithCancel(context.Background())
	bcsExecutor := &BcsExecutor{
		flag:      flag,
		status:    ExecutorStatus_NOTRUNNING,
		exeCxt:    eCxt,
		exeCancel: eCancel,
		podStatus: container.PodStatus_UNKNOWN,
		container: docker,
		launched:  false,
		tasks: &BcsTaskInfo{
			ContainerRef:  make(map[string]string),
			TaskRef:       make(map[string]string),
			TaskInfo:      make(map[string]*mesos.TaskInfo),
			ContainerInfo: make(map[string]*container.BcsContainerInfo),
		},
		messages:               make(map[int64]*bcstype.BcsMessage),
		extendedResourceDriver: extendedresource.NewDriver(flag.ExtendedResourceDir),
	}
	//create network manager for executor
	createNetManager(bcsExecutor, bcsExecutor.flag)
	if bcsExecutor.netManager == nil {
		logs.Errorf("BcsExecutor init cni network manager failed, NO CNI config deployed")
		return nil
	}
	if err := bcsExecutor.netManager.Init(); err != nil {
		logs.Errorf("BcsExecutor init network manager failed: %s\n", err.Error())
		return nil
	}
	return bcsExecutor
}

func createNetManager(bcsExecutor *BcsExecutor, flag *CommandFlags) {
	if flag.NetworkMode == "cni" {
		//cni pod
		bcsExecutor.netManager = cninet.NewNetManager(flag.CNIPluginDir+"/bin", flag.CNIPluginDir+"/conf")
	} else {
		//docker pod
		bcsExecutor.netManager = cnmnet.NewNetManager()
	}
}

func createPod(executor *BcsExecutor, flag *CommandFlags,
	containerTasks []*container.BcsContainerTask, podEvent *container.PodEventHandler) {
	if flag.NetworkMode == "cni" {
		//cni pod
		executor.podInst = cni.NewPod(executor.container, containerTasks, podEvent,
			flag.NetworkImage, executor.extendedResourceDriver)
	} else {
		//docker pod
		executor.podInst = cnm.NewPod(executor.container, containerTasks, podEvent,
			executor.extendedResourceDriver)
	}
}

//BcsExecutor implement interface MesosContainerExecutor
type BcsExecutor struct {
	flag                   *CommandFlags       //command line flags
	exeCxt                 context.Context     //exit context
	exeCancel              context.CancelFunc  //exit cancel func
	netManager             network.NetManager  //network operation interface
	podInst                container.Pod       //pod interface
	podStatus              container.PodStatus //pod status
	driver                 exec.ExecutorDriver //ExecutorDriver
	status                 string              //flag for task statuss
	container              container.Container //container operation tool
	launched               bool                //when true, executor already launched task
	exeLock                sync.RWMutex        //lock for tasks & monitors
	tasks                  *BcsTaskInfo        //taskinfo cache, key is TaskName
	messages               map[int64]*bcstype.BcsMessage
	extendedResourceDriver *extendedresource.Driver
}

//Stop send stop signal
func (executor *BcsExecutor) Stop() {
	executor.exeCancel()
}

//SetDriver driver injection
func (executor *BcsExecutor) SetDriver(driver exec.ExecutorDriver) {
	executor.driver = driver
}

//Registered call by ExecutorDriver when receiving Subscribed from mesos slave
func (executor *BcsExecutor) Registered(driver exec.ExecutorDriver, execInfo *mesos.ExecutorInfo, fwinfo *mesos.FrameworkInfo, slaveInfo *mesos.AgentInfo) {
	executor.driver = driver
	logs.Infoln("Registered Executor on slave ", slaveInfo.GetHostname())
}

//Reregistered call by ExecutorDriver when receiving Subscribed from mesos slave
func (executor *BcsExecutor) Reregistered(driver exec.ExecutorDriver, slaveInfo *mesos.AgentInfo) {
	logs.Infoln("Re-registered Executor on slave ", slaveInfo.GetHostname())
}

//Disconnected call by ExecutorDriver if cnnection to mesos slave broken
func (executor *BcsExecutor) Disconnected(driver exec.ExecutorDriver) {
	logs.Infoln("Executor disconnected. Waiting for Reregistered.")
}

//LaunchTask launch a task, create taskgroup for this TaskInfo
//all launchTask request will be handled by LaunchTaskGroup
func (executor *BcsExecutor) LaunchTask(driver exec.ExecutorDriver, taskInfo *mesos.TaskInfo) {
	taskGroup := &mesos.TaskGroupInfo{}
	logs.Infof("BcsExecutor asked to launch task %s, construct group to handle\n", taskInfo.GetTaskId().GetValue())
	taskGroup.Tasks = append(taskGroup.Tasks, taskInfo)
	executor.LaunchTaskGroup(driver, taskGroup)
}

//LaunchTaskGroup Invoked when a task has been launched on this executor (initiated
//via SchedulerDriver.LaunchTasks). Note that this task can be realized
//with a goroutine, an external process, or some simple computation, however,
//no other callbacks will be invoked on this executor until this
//callback has returned.
func (executor *BcsExecutor) LaunchTaskGroup(driver exec.ExecutorDriver, taskGroup *mesos.TaskGroupInfo) {
	logs.Infof("BcsExecutor asked to launch %d tasks.\n", len(taskGroup.Tasks))
	if executor.status != ExecutorStatus_NOTRUNNING {
		logs.Errorln("BcsExecutor is already launch task before. Drop taskGroup")
		executor.updateTaskGroup(driver, taskGroup, mesos.TaskState_TASK_DROPPED, "executor is launched, drop incomming task")
		return
	}

	//begin to launch
	executor.exeLock.Lock()
	defer executor.exeLock.Unlock()
	executor.status = ExecutorStatus_LAUNCHING
	//construct BcsContainerTask to create Pod
	var containerTasks []*container.BcsContainerTask
	for _, taskInfo := range taskGroup.GetTasks() {
		by, _ := json.Marshal(taskInfo)
		logs.Infof("Launch Task %s with data %s.\n", taskInfo.GetName(), string(by))

		taskID := taskInfo.GetTaskId()
		if t := executor.tasks.GetTask(taskID.GetValue()); t != nil {
			executor.updateTaskGroup(driver, taskGroup, mesos.TaskState_TASK_DROPPED, "Duplicated TaskInfo ID from Slave")
			executor.status = ExecutorStatus_SHUTDOWN
			driver.Stop()
			return
		}
		if taskInfo.GetContainer() == nil {
			logs.Errorf("Task %s Launch failed, Empty ContainerInfo in TaskInfo\n", taskID.GetValue())
			executor.updateTaskGroup(driver, taskGroup, mesos.TaskState_TASK_ERROR, "Lost Container info in TaskInfo")
			executor.status = ExecutorStatus_SHUTDOWN
			driver.Stop()
			return
		}
		//parse TaskInfo.Data to bcstypes.DataClass
		dataClass, dataErr := executor.handleDataClass(taskInfo)
		if dataErr != nil {
			logs.Errorf("Launch Task %s failed, decode BcsDataClass err: %s\n", taskID.GetValue(), dataErr.Error())
			executor.updateTaskGroup(driver, taskGroup, mesos.TaskState_TASK_FAILED, dataErr.Error())
			executor.status = ExecutorStatus_SHUTDOWN
			driver.Stop()
			return
		}
		mesosContainer := taskInfo.GetContainer()
		dockerInfo := mesosContainer.GetDocker()
		//create ContainerName for starting Docker Container
		/*agentID := taskInfo.GetAgentId().GetValue()
		randID := uuid.NewUUID().String()*/
		containerTask := new(container.BcsContainerTask)
		containerTask.PublishAllPorts = false
		containerTask.OOMKillDisabled = false
		containerTask.AutoRemove = true
		containerTask.KillPolicy = 1
		//containerTask.Name = ContainerNamePrefix + agentID + "." + randID
		containerTask.Name = ContainerNamePrefix + taskID.GetValue()
		containerTask.TaskId = taskID.GetValue()
		containerTask.HostName = mesosContainer.GetHostname()
		//Setting grace exit for every container
		if taskInfo.GetKillPolicy() != nil && taskInfo.GetKillPolicy().GetGracePeriod() != nil {
			gracePeriod := int(taskInfo.GetKillPolicy().GetGracePeriod().GetNanoseconds() / 1000000000)
			if gracePeriod > containerTask.KillPolicy {
				containerTask.KillPolicy = int(gracePeriod)
				logs.Infof("Task %s Get Kill Policy %d seconds\n", taskID.GetValue(), containerTask.KillPolicy)
			} else {
				logs.Infof("Task %s Get default Kill Policy %d seconds\n", taskID.GetValue(), containerTask.KillPolicy)
			}
		}
		//setting command & Env
		executor.commandEnvSetting(containerTask, taskInfo)
		//parse custom paramter for docker
		executor.dockerParameterSetting(containerTask, dockerInfo)
		logs.Infof("Create container task with image %s", containerTask.Image)
		//custom setting for container
		if customErr := executor.customSettingContainer(containerTask, dataClass); customErr != nil {
			logs.Errorf("Launch Task %s failed, setting custom feature err: %s\n", taskID.GetValue(), customErr.Error())
			executor.updateTaskGroup(driver, taskGroup, mesos.TaskState_TASK_FAILED, customErr.Error())
			executor.status = ExecutorStatus_SHUTDOWN
			driver.Stop()
			return
		}
		executor.tasks.SetTaskInfo(taskInfo)
		containerInfo := new(container.BcsContainerInfo)
		containerInfo.Name = containerTask.Name
		reflectErr := executor.tasks.SetContainerWithTaskID(taskID.GetValue(), containerInfo)
		logs.Infof("Task %s create container %s info success. recored in cache: %v", taskID.GetValue(), containerTask.Name, reflectErr)
		//copy ports info from taskInfo
		portMaps := dockerInfo.GetPortMappings()
		for _, ports := range portMaps {
			p := container.BcsPort{
				Protocol:      strings.ToLower(ports.GetProtocol()),
				ContainerPort: strconv.Itoa(int(ports.GetContainerPort())),
				HostPort:      strconv.Itoa(int(ports.GetHostPort())),
			}
			containerInfo.Ports = append(containerInfo.Ports, p)
		}
		//setting attach network interface for Container
		switch dockerInfo.GetNetwork() {
		case mesos.ContainerInfo_DockerInfo_USER:
			//user define network, get setting info from Parameters.X.Key == net
			logs.Infof("Get user define network: %s", containerTask.NetworkName)
			if executor.flag.NetworkMode == "cnm" && !containerTask.PublishAllPorts {
				//check Container Port publish
				executor.portMappingSetting(containerTask, taskInfo)
			}
		//done(developerJim): setting PortMapping, only useful when use docker network mode #bridge#
		case mesos.ContainerInfo_DockerInfo_BRIDGE:
			//check all ports publish flag, setting in dockerParamterSetting()
			if !containerTask.PublishAllPorts {
				//check Container Port publish
				executor.portMappingSetting(containerTask, taskInfo)
			}
			containerTask.NetworkName = strings.ToLower(dockerInfo.GetNetwork().String())
		default:
			//other netowrk: HOST, NONE, overwrite old value
			containerTask.NetworkName = strings.ToLower(dockerInfo.GetNetwork().String())
		}
		//setting volumes
		executor.volumeSetting(containerTask, taskInfo)
		//setting health check
		if err := executor.healthcheckSetting(containerTask, taskInfo); err != nil {
			logs.Errorf("Create healcheck for image: %s failed, %s\n", containerTask.Image, err.Error())
			executor.updateTaskGroup(driver, taskGroup, mesos.TaskState_TASK_FAILED, "create healcheck failed, "+err.Error())
			executor.status = ExecutorStatus_SHUTDOWN
			driver.Stop()
			return
		}
		//adding
		containerTasks = append(containerTasks, containerTask)
	}
	//tasks are ready, create Pod now
	podEvent := &container.PodEventHandler{
		PreStart: executor.preContainerStartEventCallback,
	}

	createPod(executor, executor.flag, containerTasks, podEvent)

	executor.podInst.SetPodID(executor.driver.ExecutorID())
	if initErr := executor.podInst.Init(); initErr != nil {
		executor.updateTaskGroup(driver, taskGroup, mesos.TaskState_TASK_FAILED, "Pod init failed: "+executor.podInst.GetMessage())
		executor.status = ExecutorStatus_SHUTDOWN
		driver.Stop()
		return
	}

	//create recovery for release ip address
	defer func() {
		if err := recover(); err != nil && executor.podInst != nil {
			logs.Errorf("Lauch panic: %v\n", err)
			executor.podInst.Stop(1)
			executor.updateTaskGroup(driver, taskGroup, mesos.TaskState_TASK_FAILED, "bcs executor down")
			executor.netManager.TearDownPod(executor.podInst)
			executor.podInst.Finit()
			driver.Stop()
		}
	}()

	// When the start time of the container is too long (for example: pull images is too long),
	// continuously send the starting status to ensure the normal start of the container
	stopCh := make(chan struct{})
	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-stopCh:
				logs.Infof("stop send task status starting")
				return

			case <-ticker.C:
				executor.updateTaskGroup(driver, taskGroup, mesos.TaskState_TASK_STARTING, "taskgroup is starting")
			}
		}
	}()

	//go executor.watchStartingTask()
	logs.Infof("BcsExecutor init pod success.")

	if setupErr := executor.netManager.SetUpPod(executor.podInst); setupErr != nil {
		close(stopCh)
		executor.updateTaskGroup(driver, taskGroup, mesos.TaskState_TASK_FAILED, "Pod Setup failed: "+setupErr.Error())
		executor.status = ExecutorStatus_SHUTDOWN
		executor.podInst.Finit()
		driver.Stop()
		return
	}
	logs.Infof("BcsExecutor Setup pod network success.")

	if startErr := executor.podInst.Start(); startErr != nil {
		close(stopCh)
		executor.updateTaskGroup(driver, taskGroup, mesos.TaskState_TASK_FAILED, "Pod Start failed: "+startErr.Error())
		executor.status = ExecutorStatus_SHUTDOWN
		executor.netManager.TearDownPod(executor.podInst)
		executor.podInst.Finit()
		driver.Stop()
		return
	}
	close(stopCh)
	logs.Infof("BcsExecutor start pod success. update local container info, ready to watch Pod status")

	executor.podStatus = container.PodStatus_STARTING
	executor.updateTaskGroup(driver, taskGroup, mesos.TaskState_TASK_STARTING, "Pod is starting")
	//start success, Get container info and reply to scheduler
	//that container is starting
	containerInfos := executor.podInst.GetContainers()
	//setting containerInfos<==>taskInfo
	for _, conInfo := range containerInfos {
		oldContainer := executor.tasks.GetContainer(conInfo.Name)
		if oldContainer == nil {
			logs.Errorf("BcsExecutor Lost container %s in cache\n", conInfo.Name)
			continue
		}
		oldContainer.Update(conInfo)
	}
	//watching all containers
	go executor.monitorPod()
	return
}

func (executor *BcsExecutor) watchStartingTask() {
	for {
		//time.Sleep(time.Minute)
		if executor.podStatus != container.PodStatus_INIT &&
			executor.podStatus != container.PodStatus_UNKNOWN {
			logs.Infof("pod status %s, then return\n", executor.podStatus)
			return
		}

		for _, task := range executor.podInst.GetContainerTasks() {

			taskinfo := executor.tasks.TaskInfo[task.TaskId]
			update := &mesos.TaskStatus{
				TaskId:  taskinfo.GetTaskId(),
				State:   mesos.TaskState_TASK_STARTING.Enum(),
				Message: proto.String(fmt.Sprintf("task is starting")),
				Source:  mesos.TaskStatus_SOURCE_EXECUTOR.Enum(),
			}
			executor.driver.SendStatusUpdate(update)
		}
	}
}

//KillTask kill task by taskId
func (executor *BcsExecutor) KillTask(driver exec.ExecutorDriver, taskID *mesos.TaskID) {
	logs.Infof("Kill task %s\n", taskID.GetValue())
	executor.exeLock.Lock()
	defer executor.exeLock.Unlock()

	if executor.status == ExecutorStatus_KILLING {
		//Executor is under killing status
		logs.Infoln("BcsExecutor is under killing. Drop killMessage")
		return
	}
	if executor.status == ExecutorStatus_SHUTDOWN {
		logs.Infoln("BcsExecutor is under Shutdown. Drop KillMessage")
		return
	}
	//search kill task
	container := executor.tasks.GetContainerByTaskID(taskID.GetValue())
	if container == nil {
		logs.Errorf("Can not find Task %s by ID, task lost\n", taskID.GetValue())
		lost := &mesos.TaskStatus{
			TaskId:  taskID,
			State:   mesos.TaskState_TASK_LOST.Enum(),
			Message: proto.String("Task lost in Executor"),
		}
		driver.SendStatusUpdate(lost)
		return
	}
	executor.status = ExecutorStatus_KILLING
	//task found, because we support pod feature,
	//if one task was killed, all others need to kill either
	taskGroup := executor.tasks.GetTaskGroup()
	executor.Stop()
	executor.updateTaskGroup(driver, taskGroup, mesos.TaskState_TASK_KILLING, "task is under killing")
	//ready to kill TaskGroup
	executor.killTaskGroup()

	//reply task status
	//todo(developerJim): maybe we need check killing results for task
	//executor need to handle if task is under TaskState_TASK_UNREACHABLE
	//because docker container is down
	executor.updateTaskGroup(driver, taskGroup, mesos.TaskState_TASK_FINISHED, "task was killed")
	executor.tasks.Clean()
}

//FrameworkMessage receiving message from scheduler framework
func (executor *BcsExecutor) FrameworkMessage(driver exec.ExecutorDriver, msg string) {
	//parse to BcsMessage
	var bcsMessage bcstype.BcsMessage
	if err := json.Unmarshal([]byte(msg), &bcsMessage); err != nil {
		logs.Errorf("parse Framework Message err: %s, origin message: %s", err.Error(), msg)
		return
	}

	var err error
	switch *bcsMessage.Type {
	case bcstype.Msg_LOCALFILE:
		err = executor.frameworkMessageFileUpload(bcsMessage.TaskID.GetValue(), bcsMessage.Local)
	case bcstype.Msg_REMOTE:
		err = executor.frameworkMessageRemoteFile(bcsMessage.TaskID.GetValue(), bcsMessage.Remote)
	case bcstype.Msg_SIGNAL:
		err = executor.frameworkMessageSignalExecute(bcsMessage.TaskID.GetValue(), bcsMessage.Sig)
	case bcstype.Msg_ENV:
		//executor.frameworkMessageEnvironmentUpdate(bcsMessage.TaskID.GetValue(), bcsMessage.Env)
		err = fmt.Errorf("executor can not support env change when container is running")
		logs.Errorln("executor can not support ENV change when container is running.")
	case bcstype.Msg_UPDATE_TASK:
		err = executor.frameworkMessageUpdateResources(bcsMessage.UpdateTaskResources)
	case bcstype.Msg_COMMIT_TASK:
		err = executor.frameworkMessageCommitTask(bcsMessage.CommitTask)
	case bcstype.Msg_Req_COMMAND_TASK:
		go executor.frameworkMessageCommandTask(bcsMessage.RequestCommandTask)
		return

	default:
		logs.Errorf("Get unknown message type %d in frameworkMessage", *bcsMessage.Type)
	}

	if err != nil {
		bcsMessage.Status = bcstype.Msg_Status_Failed
		bcsMessage.Message = err.Error()
		logs.Errorf("handler message error %s", err.Error())
	} else {
		bcsMessage.Status = bcstype.Msg_Status_Success
	}
	bcsMessage.CompleteTime = time.Now().Unix()

	executor.exeLock.Lock()
	executor.messages[bcsMessage.Id] = &bcsMessage
	executor.exeLock.Unlock()
}

//Shutdown Executor Shutdown & exit
func (executor *BcsExecutor) Shutdown(driver exec.ExecutorDriver) {
	logs.Infoln("Shutting down the executor")
	//Executor asked to stop, just killing all Container and quit
	executor.exeLock.Lock()
	defer executor.exeLock.Unlock()
	executor.Stop()
	if executor.status == ExecutorStatus_RUNNING || executor.status == ExecutorStatus_LAUNCHING {
		logs.Infoln("Executor is under RUNNING or LAUNCHING, prepared to clean all tasks")
		//clean TaskGroup & running containers
		taskGroup := executor.tasks.GetTaskGroup()
		if taskGroup == nil {
			logs.Errorln("BcsExecutor Get no taskGroup info when shutdown!")
			executor.status = ExecutorStatus_SHUTDOWN
			return
		}
		executor.updateTaskGroup(driver, taskGroup, mesos.TaskState_TASK_KILLING, "task was killing")
		executor.killTaskGroup()
		//todo(developerJim): maybe we need check killing results for task
		//executor need to handle if task is under TaskState_TASK_UNREACHABLE
		//because docker container is down
		executor.updateTaskGroup(driver, taskGroup, mesos.TaskState_TASK_FINISHED, "task was shutdown")
		executor.tasks.Clean()
	} else {
		logs.Infof("BcsExecutor is under %s, no container need to kill\n", executor.status)
	}
	executor.status = ExecutorStatus_SHUTDOWN
}

//Error Executor receiving Error message from mesos slave.
func (executor *BcsExecutor) Error(driver exec.ExecutorDriver, err string) {
	logs.Infoln("Got error message:", err)
}

//monitorContainer monitor Container by Name
//if container is lost, or stopCh is closed, monitor exit.
func (executor *BcsExecutor) monitorPod() {
	defer func() {
		if err := recover(); err != nil {
			logs.Errorf("%v\n", err)
			executor.podInst.Stop(1)
			executor.netManager.TearDownPod(executor.podInst)
			executor.podInst.Finit()
			executor.tasks.Clean()
			executor.Stop()
			executor.driver.Stop()
		}
	}()

	tick := time.NewTicker(1 * time.Second)
	defer tick.Stop()
	reporting := 0
	for {
		select {
		case <-executor.exeCxt.Done():
			logs.Infoln("Monitor for Pod quit.")
			return
		case <-tick.C:
			executor.exeLock.Lock()
			//logs.Infof("ready to check Container status, now: %s\n", time.Now())
			if executor.status == ExecutorStatus_KILLING || executor.status == ExecutorStatus_SHUTDOWN {
				//executor ready to exit
				logs.Infoln("BcsExecutor status changed to exit/shutdown, no Pod monitor yet")
				executor.exeLock.Unlock()
				return
			}

			var changed bool
			podNewStatus := executor.podInst.GetPodStatus()
			if executor.podStatus != podNewStatus {
				logs.Infof("Pod status change from %s to %s\n", executor.podStatus, podNewStatus)
				changed = true
				executor.podStatus = podNewStatus
			}
			switch executor.podStatus {
			case container.PodStatus_UNKNOWN, container.PodStatus_INIT:
				//error status,
				logs.Errorf("BcsExecutor Got error pod status: %s, wait for fix\n", executor.podStatus)
			case container.PodStatus_STARTING, container.PodStatus_KILLING:
				logs.Infof("BcsExecutor get pod under %s, wait for next tick\n", executor.podStatus)
			case container.PodStatus_RUNNING:
				//newStatus := executor.podInst.IsHealthy()
				//changed := (newStatus != lastHealthStatus)
				var message *bcstype.BcsMessage
				for _, msg := range executor.messages {
					message = msg
					break
				}

				if message != nil {
					delete(executor.messages, message.Id)
				}

				for _, task := range executor.podInst.GetContainerTasks() {
					var healthyChanged bool
					if task.HealthCheck != nil {
						if task.RuntimeConf.Healthy != task.HealthCheck.IsHealthy() || task.RuntimeConf.IsChecked != task.HealthCheck.IsTicks() ||
							task.RuntimeConf.ConsecutiveFailureTimes != task.HealthCheck.ConsecutiveFailure() {
							healthyChanged = true
							task.RuntimeConf.Healthy = task.HealthCheck.IsHealthy()
							task.RuntimeConf.IsChecked = task.HealthCheck.IsTicks()
							task.RuntimeConf.ConsecutiveFailureTimes = task.HealthCheck.ConsecutiveFailure()
						}
					}
					taskRunning := true
					if task.RuntimeConf.Status != container.ContainerStatus_RUNNING && task.RuntimeConf.Status != container.ContainerStatus_PAUSED {
						taskRunning = false
					}

					if reporting%300 == 0 || changed || healthyChanged || message != nil || !taskRunning {
						//report data every 30 seconds or pod healthy status changed
						executor.status = ExecutorStatus_RUNNING
						logs.Infof("all task is Running, healthy: %t, isChecked: %t, ConsecutiveFailureTimes: %d"+
							" report task status\n", task.RuntimeConf.Healthy, task.RuntimeConf.IsChecked, task.RuntimeConf.ConsecutiveFailureTimes)
						//get all container update and report to scheduler
						///for _, info := range containers {
						info := task.RuntimeConf
						info.BcsMessage = message
						localInfo := executor.tasks.GetContainer(info.Name)
						localInfo.Update(info)
						infoby, _ := json.Marshal(localInfo)
						taskInfo := executor.tasks.GetTaskByContainerID(info.Name)

						//if changed, send task status update
						//if task status!=running, then send status update
						if changed || !taskRunning {
							update := &mesos.TaskStatus{
								TaskId:  taskInfo.GetTaskId(),
								State:   mesos.TaskState_TASK_RUNNING.Enum(),
								Message: proto.String(localInfo.Message),
								Source:  mesos.TaskStatus_SOURCE_EXECUTOR.Enum(),
								Healthy: proto.Bool(task.RuntimeConf.Healthy),
							}
							update.Data = infoby
							executor.driver.SendStatusUpdate(update)
							//if running, send message for task status update
						} else {
							bcsMsg := &bcstype.BcsMessage{
								Id:         time.Now().UnixNano(),
								TaskID:     taskInfo.GetTaskId(),
								Type:       bcstype.Msg_TASK_STATUS_UPDATE.Enum(),
								TaskStatus: infoby,
							}
							by, _ := json.Marshal(bcsMsg)
							_, err := executor.driver.SendFrameworkMessage(string(by))
							if err != nil {
								logs.Errorf("send framework message error %s", err.Error())
							}
						}
					}
				}

				reporting++
			case container.PodStatus_FAILED, container.PodStatus_KILLED:
				logs.Errorf("BcsExecutor Get Pos status %s, pod is down. ready to clean\n", executor.podStatus)
				//report container task status, clean all task info and exit
				containers := executor.podInst.GetContainers()
				for _, info := range containers {
					localInfo := executor.tasks.GetContainer(info.Name)
					localInfo.Update(info)
					taskInfo := executor.tasks.GetTaskByContainerID(info.Name)
					update := &mesos.TaskStatus{
						TaskId:  taskInfo.GetTaskId(),
						State:   mesos.TaskState_TASK_FAILED.Enum(),
						Message: proto.String(info.Message),
						Source:  mesos.TaskStatus_SOURCE_EXECUTOR.Enum(),
					}
					update.Data, _ = json.Marshal(localInfo)
					executor.driver.SendStatusUpdate(update)
				}
				logs.Errorf("BcsExecutor reports containers in pod done, teardown pod and clean task info\n")
				if err := executor.netManager.TearDownPod(executor.podInst); err != nil {
					logs.Errorf("BcsExecutor teardown pod err: %s\n", err.Error())
					//todo(developerJim): NetManager using NetService interface to release ip address
				}
				executor.podInst.Finit()
				executor.tasks.Clean()
				executor.Stop()
				executor.driver.Stop()
				executor.exeLock.Unlock()
				return
			case container.PodStatus_FINISH:
				logs.Errorf("BcsExecutor Get Pos status %s, pod is down. ready to clean\n", executor.podStatus)
				//report container task status, clean all task info and exit
				containers := executor.podInst.GetContainers()
				for _, info := range containers {
					localInfo := executor.tasks.GetContainer(info.Name)
					localInfo.Update(info)
					taskInfo := executor.tasks.GetTaskByContainerID(info.Name)
					update := &mesos.TaskStatus{
						TaskId:  taskInfo.GetTaskId(),
						State:   mesos.TaskState_TASK_FINISHED.Enum(),
						Message: proto.String(info.Message),
						Source:  mesos.TaskStatus_SOURCE_EXECUTOR.Enum(),
					}
					update.Data, _ = json.Marshal(localInfo)
					executor.driver.SendStatusUpdate(update)
				}
				logs.Errorf("BcsExecutor reports containers in pod done, teardown pod and clean task info\n")
				if err := executor.netManager.TearDownPod(executor.podInst); err != nil {
					logs.Errorf("BcsExecutor teardown pod err: %s\n", err.Error())
					//todo(developerJim): NetManager using NetService interface to release ip address
				}
				executor.podInst.Finit()
				executor.tasks.Clean()
				executor.Stop()
				executor.driver.Stop()
				executor.exeLock.Unlock()
				return
			} //end switch
			executor.exeLock.Unlock()
		} //end of select
	} //end of for
}

func (executor *BcsExecutor) unHealthyNotify(checker healthcheck.Checker) {

}

//killTaskGroup kill all record TaskInfo
func (executor *BcsExecutor) killTaskGroup() {
	//stop Pod and TearDownPod network configuration
	executor.podInst.Stop(1)
	if err := executor.netManager.TearDownPod(executor.podInst); err != nil {
		logs.Errorf("BcsExecutor teardown pod with %s failed, err: %s", executor.podInst.GetNetworkName(), err.Error())
		//todo(developerJim): if reason of failure is CNI execute error,
		//ip address info must be released forcibly
	}
	//ignore error now
	executor.podInst.Finit()
	logs.Infoln("BcsExecutor killed all containers in runtime")
}

//updateTaskGroup update all task status in TaskGroup
func (executor *BcsExecutor) updateTaskGroup(driver exec.ExecutorDriver, taskGroup *mesos.TaskGroupInfo, status mesos.TaskState, message string) {
	for _, task := range taskGroup.GetTasks() {
		err := executor.updateTaskStatus(driver, task, status, message)
		if err != nil {
			logs.Errorf("Update task %s status failed: %s\n", task.GetTaskId().GetValue(), err.Error())
		}
	}
}

//updateTaskStatus update task with status and message
func (executor *BcsExecutor) updateTaskStatus(driver exec.ExecutorDriver, taskInfo *mesos.TaskInfo, status mesos.TaskState, message string) error {
	logs.Infof("Task %s Update Status %s, Message: %s\n", taskInfo.GetTaskId().GetValue(), status.String(), message)
	update := &mesos.TaskStatus{
		TaskId:  taskInfo.GetTaskId(),
		State:   status.Enum(),
		Message: proto.String(message),
		Source:  mesos.TaskStatus_SOURCE_EXECUTOR.Enum(),
	}
	container := executor.tasks.GetContainerByTaskID(taskInfo.GetTaskId().GetValue())
	if container != nil {
		update.Data, _ = json.Marshal(container)
	} else {
		logs.Infof("Task %s update without container status.\n", taskInfo.GetTaskId().GetValue())
	}
	_, err := driver.SendStatusUpdate(update)
	return err
}

//handleDataClass parse TaskInfo.Data to bcstype.DataClass
func (executor *BcsExecutor) handleDataClass(taskInfo *mesos.TaskInfo) (*bcstype.DataClass, error) {
	base64Bytes := taskInfo.GetData()
	if base64Bytes == nil || len(base64Bytes) == 0 {
		//no data for this task
		logs.Errorln("TaskInfo.Data is empty in Task ", taskInfo.GetTaskId().GetValue())
		return nil, fmt.Errorf("TaskInfo.Data is empty")
	}
	//base64 decode
	dataBytes, err := base64.StdEncoding.DecodeString(string(base64Bytes))
	if err != nil {
		logs.Errorf("decode TaskInfo.Data failed: %s\n", err.Error())
		return nil, fmt.Errorf("base64 decode taskinfo.data failed: " + err.Error())
	}

	logs.Infof("handle dataclass %s", string(dataBytes))

	var data bcstype.DataClass
	//parse json to message list
	if jsonErr := json.Unmarshal(dataBytes, &data); jsonErr != nil {
		logs.Errorf("decode TaskInfo.Data json failed: %s\n", jsonErr.Error())
		return nil, fmt.Errorf("json decode taskinfo.data failed: " + jsonErr.Error())
	}
	return &data, nil
}

//customSettingContainer setting container info before createContainer
func (executor *BcsExecutor) customSettingContainer(taskInfo *container.BcsContainerTask, dataClass *bcstype.DataClass) error {
	//setting sandbox path,
	// volumn := container.BcsVolume{
	// 	ReadOnly:      true,
	// 	HostPath:      os.Getenv("MESOS_SANDBOX"),
	// 	ContainerPath: executor.flag.MappedDirectory,
	// }
	// taskInfo.Volums = append(taskInfo.Volums, volumn)
	//setting corefile volume
	IDList := strings.Split(executor.driver.ExecutorID(), ".")
	hostCore := filepath.Join(ContainerCoreFilePrefix, strings.Join(IDList[0:len(IDList)-1], "."), IDList[len(IDList)-1])
	os.MkdirAll(hostCore, 0777)
	os.Chmod(hostCore, 0777)
	coreVol := container.BcsVolume{
		ReadOnly:      true,
		HostPath:      hostCore,
		ContainerPath: "/data/corefile",
	}
	taskInfo.Volums = append(taskInfo.Volums, coreVol)
	//setting default Environments
	stName := container.BcsKV{
		Key:   "BCS_ETH_NAME",
		Value: "eth1",
	}
	taskInfo.Env = append(taskInfo.Env, stName)
	cniName := container.BcsKV{
		Key:   "BCS_CNI_NAME",
		Value: "eth1",
	}
	taskInfo.Env = append(taskInfo.Env, cniName)
	//setting default Environments
	// caliName := container.BcsKV{
	// 	Key:   "BCS_CALICO_NAME",
	// 	Value: "cali0",
	// }
	// taskInfo.Env = append(taskInfo.Env, caliName)
	//setting mesos slave host ip
	ipAddr := container.BcsKV{
		Key:   "BCS_NODE_IP",
		Value: util.GetIPAddress(),
	}
	taskInfo.Env = append(taskInfo.Env, ipAddr)
	podID := container.BcsKV{
		Key:   "BCS_POD_ID",
		Value: executor.driver.ExecutorID(),
	}
	taskInfo.Env = append(taskInfo.Env, podID)
	//check dataClass for bcs define info
	if dataClass.Msgs != nil && len(dataClass.Msgs) > 0 {
		for _, item := range dataClass.Msgs {
			switch *item.Type {
			case bcstype.Msg_SECRET:
				//done(developerJim): check Secret.Type for Environment or File
				//only support Environments default
				if *item.Secret.Type == bcstype.Secret_Env {
					data, err := base64.StdEncoding.DecodeString(*item.Secret.Value)
					if err != nil {
						return fmt.Errorf("Decode secret %s to Environment failed, %s", *item.Secret.Value, err.Error())
					}
					secretEnv := container.BcsKV{
						Key:   *item.Secret.Name,
						Value: string(data),
					}
					logs.Infof("Additional adding Secrets [%s=%s] to Environment in customSetting\n", secretEnv.Key, secretEnv.Value)
					taskInfo.Env = append(taskInfo.Env, secretEnv)
				} else if *item.Secret.Type == bcstype.Secret_File {
					secretFile := new(bcstype.Msg_LocalFile)
					secretFile.To = proto.String(*item.Secret.Name)
					secretFile.User = proto.String("root")
					secretFile.Right = proto.String("r")
					secretFile.Base64 = item.Secret.Value
					secretItem := new(bcstype.BcsMessage)
					secretItem.Type = bcstype.Msg_LOCALFILE.Enum()
					secretItem.Local = secretFile
					logs.Infof("Setting secret %s to file\n", *item.Secret.Name)
					taskInfo.BcsMessages = append(taskInfo.BcsMessages, secretItem)
				}
			case bcstype.Msg_REMOTE:
				if err := executor.dataClassRemote(taskInfo, item.Remote); err != nil {
					return err
				}
			case bcstype.Msg_ENV_REMOTE:
				if err := executor.dataClassRemoteEnv(taskInfo, item.EnvRemote); err != nil {
					return err
				}
			case bcstype.Msg_ENV:
				//base64 decoding
				contentBytes, err := base64.StdEncoding.DecodeString(*item.Env.Value)
				if err != nil {
					logs.Errorf("decode bcs custom Environment Err: %s\n", err.Error())
					return err
				}
				value := string(contentBytes)
				customEnv := container.BcsKV{
					Key:   *item.Env.Name,
					Value: value,
				}
				logs.Infof("custom Msg [%s=%s] to Environment in customSetting\n", *item.Env.Name, value)
				taskInfo.Env = append(taskInfo.Env, customEnv)
			case bcstype.Msg_LOCALFILE:
				taskInfo.BcsMessages = append(taskInfo.BcsMessages, item)
			}
		}
	}
	//resource info
	taskInfo.Resource = dataClass.Resources
	taskInfo.LimitResource = dataClass.LimitResources
	taskInfo.ExtendedResources = dataClass.ExtendedResources
	taskInfo.NetLimit = dataClass.NetLimit
	return nil

}

func (executor *BcsExecutor) dockerParameterSetting(containerTask *container.BcsContainerTask, dockerInfo *mesos.ContainerInfo_DockerInfo) {
	//parse parameter, we can add parameter if needed
	containerTask.Privileged = dockerInfo.GetPrivileged()
	containerTask.ForcePullImage = dockerInfo.GetForcePullImage()
	containerTask.Image = dockerInfo.GetImage()

	for _, param := range dockerInfo.GetParameters() {
		switch param.GetKey() {
		case "net":
			containerTask.NetworkName = param.GetValue()
		case "P":
			containerTask.PublishAllPorts = true
		case "oom-kill-disable":
			if param.GetValue() == "true" {
				containerTask.OOMKillDisabled = true
				logs.Infoln("docker parameter cumstom setting oom-kill-disable: true")
			}
		case "rm":
			if param.GetValue() == "false" {
				containerTask.AutoRemove = false
				logs.Infoln("docker parameter cumstom setting --rm: false")
			}
		case "ulimit":
			data := strings.Split(param.GetValue(), "=")
			if len(data) != 2 {
				logs.Errorf("Executor Get error ulimit setting, string: %s, length: %d", param.GetValue(), len(data))
				continue
			}
			_, err := strconv.Atoi(data[1])
			if err != nil {
				logs.Errorf("Executor get error value from ulimit, string:%s , err: %s", param.GetValue(), err.Error())
				continue
			}
			kv := container.BcsKV{
				Key:   data[0],
				Value: data[1],
			}
			logs.Infof("Executor setting ulimit %s", param.GetValue())
			containerTask.Ulimits = append(containerTask.Ulimits, kv)
		case "shm-size":
			shm, err := strconv.Atoi(param.GetValue())
			if err != nil {
				logs.Errorf("strconv.Atoi param key %s value %s error %s", param.GetKey(), param.GetValue(), err.Error())
				continue
			}

			//MB -> B
			containerTask.ShmSize = int64(1024 * 1024 * shm)

		case "ipc":
			containerTask.Ipc = string(param.GetValue())

		case "ip":
			containerTask.NetworkIPAddr = param.GetValue()

		default:
			blog.Errorf("param key %s is invalid", param.GetKey())
		}
	}
}

func (executor *BcsExecutor) commandEnvSetting(containerTask *container.BcsContainerTask, taskInfo *mesos.TaskInfo) {
	command := taskInfo.GetCommand()
	if command != nil {
		containerTask.Command = command.GetValue()
		containerTask.Args = command.GetArguments()
		//setting Env
		envs := command.GetEnvironment()
		if envs != nil {
			for _, env := range envs.GetVariables() {
				item := container.BcsKV{
					Key:   env.GetName(),
					Value: env.GetValue(),
				}
				containerTask.Env = append(containerTask.Env, item)
			}
		}
	}
	if taskInfo.GetLabels() != nil {
		//setting TaskInfo.Labels into Environment for container
		//if labels is setting by scheduler
		labels := taskInfo.GetLabels()
		for _, label := range labels.GetLabels() {
			labelEnv := container.BcsKV{
				Key:   label.GetKey(),
				Value: label.GetValue(),
			}
			logs.Infof("import lable [%s=%s] to docker label", label.GetKey(), label.GetValue())
			//containerTask.Env = append(containerTask.Env, labelEnv)
			containerTask.Labels = append(containerTask.Labels, labelEnv)

			//todo(developerJim): add custom label parse here
			if label.GetKey() == "io.tencent.bcs.netsvc.requestip" && label.GetValue() != "" {
				containerTask.NetworkIPAddr = label.GetValue()
			}
		}
	}
}

func (executor *BcsExecutor) healthcheckSetting(containerTask *container.BcsContainerTask, taskInfo *mesos.TaskInfo) error {
	health := taskInfo.GetHealthCheck()
	if health != nil {
		tm := &healthcheck.TimeMechanism{
			IntervalSeconds:     int(health.GetIntervalSeconds()),
			TimeoutSeconds:      int(health.GetTimeoutSeconds()),
			ConsecutiveFailures: int(health.GetConsecutiveFailures()),
			GracePeriodSeconds:  int(health.GetGracePeriodSeconds()),
		}
		var checker healthcheck.Checker
		var checkErr error
		switch health.GetType() {
		case mesos.HealthCheck_HTTP:
			httpChecker := health.GetHttp()
			checker, checkErr = healthcheck.NewHTTPChecker(containerTask.Name, httpChecker.GetScheme(), int(httpChecker.GetPort()), httpChecker.GetPath(), tm, nil)
		case mesos.HealthCheck_TCP:
			tcpChecker := health.GetTcp()
			checker, checkErr = healthcheck.NewTCPChecker(containerTask.Name, int(tcpChecker.GetPort()), tm, nil)
		case mesos.HealthCheck_COMMAND:
			cmdChcker := health.GetCommand()
			checker, checkErr = healthcheck.NewCommandChecker(cmdChcker.GetValue(), executor.flag.DockerSocket, tm)
		default:
			checkErr = fmt.Errorf("Get Unsupported health check type %d", health.GetType())
		}
		if checkErr != nil {
			return checkErr
		}
		containerTask.HealthCheck = checker
	}
	return nil
}

func (executor *BcsExecutor) volumeSetting(containerTask *container.BcsContainerTask, taskInfo *mesos.TaskInfo) {
	mesosContainer := taskInfo.GetContainer()
	if mesosContainer.GetVolumes() != nil && len(mesosContainer.GetVolumes()) > 0 {
		for _, volume := range mesosContainer.GetVolumes() {
			//done(developerJim): setting HostPath to BcsStorage if hostpath empty
			bv := container.BcsVolume{
				ReadOnly:      true,
				HostPath:      volume.GetHostPath(),
				ContainerPath: volume.GetContainerPath(),
			}
			if len(volume.GetHostPath()) == 0 {
				var custompath []string
				custompath = append(custompath, os.Getenv("MESOS_SANDBOX"))
				custompath = append(custompath, "storage")
				custompath = append(custompath, volume.GetContainerPath())
				bv.HostPath = strings.Join(custompath, "/")
				os.MkdirAll(bv.HostPath, 0755)
				logs.Infof("Setting indivisual path %s\n", bv.HostPath)
			}
			//feature: support BCS_POD_ID in host path to make volumn
			// difference with same application
			if strings.Contains(bv.HostPath, "$BCS_POD_ID") {
				bv.HostPath = strings.Replace(bv.HostPath, "$BCS_POD_ID", executor.driver.ExecutorID(), -1)
			}
			if volume.GetMode() == mesos.Volume_RW {
				bv.ReadOnly = false
			}
			containerTask.Volums = append(containerTask.Volums, bv)
		}
	}
}

func (executor *BcsExecutor) portMappingSetting(containerTask *container.BcsContainerTask, taskInfo *mesos.TaskInfo) {
	docker := taskInfo.GetContainer().GetDocker()
	portMaps := docker.GetPortMappings()
	containerTask.PortBindings = make(map[string]container.BcsPort)
	if portMaps != nil && len(portMaps) > 0 {
		for _, ports := range portMaps {
			if ports.GetHostPort() == 0 {
				logs.Errorln("Executor Get 0 Host port in TaskInfo. drop HostPort info")
				continue
			}
			logs.Infof("portMapping info Protocol %s HostPort %d ContainerPort %d.", ports.GetProtocol(), ports.GetHostPort(), ports.GetContainerPort())
			p := container.BcsPort{
				Protocol:      strings.ToLower(ports.GetProtocol()),
				ContainerPort: strconv.Itoa(int(ports.GetContainerPort())),
				HostPort:      strconv.Itoa(int(ports.GetHostPort())),
			}
			var key string
			if p.Protocol == "http" || p.Protocol == "https" {
				key = p.ContainerPort + "/tcp"
			} else {
				key = p.ContainerPort + "/" + p.Protocol
			}
			containerTask.PortBindings[key] = p
			logs.Infof("Executor setting portMapping %s###%v", key, p)
		}
	}
}

//postSettingContainer setting container info after createContainer but before startContainer
func (executor *BcsExecutor) preContainerStartEventCallback(containerTask *container.BcsContainerTask) error {
	//(containerID string, dataClass *bcstype.DataClass
	//setting config file
	info := containerTask.RuntimeConf
	if info == nil {
		logs.Errorln("BcsExecutor callback for preStarting error, BcsContainerInfo is lost in Runtime")
		return fmt.Errorf("BcsContainerInfo is lost in Runtime")
	}
	if info.Status != container.ContainerStatus_CREATED {
		logs.Errorf("BcsExecutor callback in preStarting event Error, container %s status need #created#, but got %s\n", info.Name, info.Status)
		return fmt.Errorf("container status error")
	}
	//self define message for setting
	if containerTask.BcsMessages != nil {
		for _, item := range containerTask.BcsMessages {
			switch *item.Type {
			//upload file to created container
			case bcstype.Msg_LOCALFILE:
				if copyErr := executor.copyFileToContainer(containerTask.RuntimeConf.ID, item.Local); copyErr != nil {
					return copyErr
				}
			} //end switch
		} //end for
	} //end bcs message
	return nil
}

//copyFileToContainer copy Msg_LocalFile to contianer by ContianerID
func (executor *BcsExecutor) copyFileToContainer(containerID string, fileInfo *bcstype.Msg_LocalFile) error {
	if containerID == "" {
		return fmt.Errorf("Container ID lost when copy file")
	}
	logs.Infof("Executor ready to copy File: %s to container %s\n", *fileInfo.To, containerID)
	//save base64 content to SANDBOX directory
	fileName := filepath.Base(*fileInfo.To)
	sourceFile := filepath.Join(os.Getenv("MESOS_SANDBOX"), fileName)
	fobj, err := os.Create(sourceFile)
	defer fobj.Close()
	if err != nil {
		return fmt.Errorf("Create File %s in SandBox failed: %s", sourceFile, err.Error())
	}
	if fileInfo.Right != nil && *fileInfo.Right == "rw" {
		fobj.Chmod(0666)
	}
	if fileInfo.User != nil && *fileInfo.User != "" && *fileInfo.User != "root" {
		//if use is not root and not empty,
		//executor will set file owner within system
		if u, uErr := user.Lookup(*fileInfo.User); uErr == nil {
			uid, _ := strconv.Atoi(u.Uid)
			gid, _ := strconv.Atoi(u.Gid)
			fobj.Chown(uid, gid)
		}
	}
	fileBytes, decodeErr := base64.StdEncoding.DecodeString(*fileInfo.Base64)
	if decodeErr != nil {
		return fmt.Errorf("Decode base64 content failed: %s", decodeErr.Error())
	}
	if _, wErr := fobj.Write(fileBytes); wErr != nil {
		return fmt.Errorf("Write File %s in Sandbox failed: %s", sourceFile, wErr.Error())
	}
	if uploadErr := executor.container.UploadToContainer(containerID, sourceFile, *fileInfo.To); uploadErr != nil {
		return fmt.Errorf("Upload File %s to Container %s failed: %s", sourceFile, containerID, uploadErr.Error())
	}
	logs.Infof("Executor copy File: %s to container %s success.\n", *fileInfo.To, containerID)
	//clean mapped directory cache file
	if mvErr := os.Remove(sourceFile); mvErr != nil {
		logs.Errorf("remove mapped derectory file %s err: %s", sourceFile, mvErr.Error())
	}
	return nil
}
