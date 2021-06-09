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

package cnm

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/util"
	schedTypes "github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/schetypes"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-container-executor/container"
	devicepluginmanager "github.com/Tencent/bk-bcs/bcs-mesos/bcs-container-executor/devicepluginmanager"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-container-executor/extendedresource"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-container-executor/healthcheck"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-container-executor/logs"

	"golang.org/x/net/context"
)

const (
	defaultPodNetnsTemplate = "/proc/%s/ns/net"
)

const (
	//defaultGracefullExit    = 1  //graceful exit for container
	defaultPodWatchInterval = 1 //interval for watch container's status
	//defaultErrTolerate      = 60 //time for determining report container runtime failed
)

const (
	// ContainerStatusAbnormal abnormal definition
	ContainerStatusAbnormal = 1
)

//NewPod create CNIPod instance with container interaface and container info
//CNM pod implementation is defferent with cni pod.
func NewPod(operator container.Container, tasks []*container.BcsContainerTask,
	handler *container.PodEventHandler, extendedResourceDriver *extendedresource.Driver) container.Pod {
	if len(tasks) == 0 {
		logs.Errorf("Create DockerPod error, Container tasks are 0")
		return nil
	}
	taskMap := make(map[string]*container.BcsContainerTask)
	if len(tasks) > 1 {
		for _, task := range tasks[1:] {
			taskMap[task.Name] = task
		}
	}
	rCxt, rCancel := context.WithCancel(context.Background())
	pod := &DockerPod{
		healthy:          true,
		status:           container.PodStatus_UNKNOWN,
		podCxt:           rCxt,
		podCancel:        rCancel,
		events:           handler,
		netTask:          tasks[0],
		conClient:        operator,
		conTasks:         taskMap,
		runningContainer: make(map[string]*container.BcsContainerInfo),
		resourceManager: devicepluginmanager.NewResourceManager(
			devicepluginmanager.NewDevicePluginManager(),
			extendedResourceDriver),
	}
	return pod
}

//DockerPod combination for mutiple container
type DockerPod struct {
	healthy          bool                                   //healthy status
	podID            string                                 //podID
	namespace        string                                 //namespace from scheduler
	cnmIPAddr        string                                 //ip address from cni tools
	netns            string                                 //netns path string
	status           container.PodStatus                    //status for all container
	exitCode         int                                    // status docker exitcode
	message          string                                 //status message
	podCxt           context.Context                        //pod root context
	podCancel        context.CancelFunc                     //pod stop func
	lock             sync.Mutex                             //lock for monitor & query
	events           *container.PodEventHandler             //pod event changed callback collection
	netTask          *container.BcsContainerTask            //network container task, using in Init stage
	conClient        container.Container                    //container operator interface
	conTasks         map[string]*container.BcsContainerTask //task for running containers, key is taskID
	runningContainer map[string]*container.BcsContainerInfo //running container Name list for monitor
	//device plugin manager
	resourceManager *devicepluginmanager.ResourceManager
}

//IsHealthy check pod is healthy
func (p *DockerPod) IsHealthy() bool {
	return p.healthy
}

//GetContainerTasks get local container task
func (p *DockerPod) GetContainerTasks() map[string]*container.BcsContainerTask {
	return p.conTasks
}

//Injection check ip address is injected
func (p *DockerPod) Injection() bool {
	return false
}

//SetIPAddr set ip addr for pod
func (p *DockerPod) SetIPAddr(ip string) {
	//not implemented
}

//GetIPAddr get pod ip address if exist
func (p *DockerPod) GetIPAddr() string {
	return ""
}

//SetPodID set pod id if needed
func (p *DockerPod) SetPodID(ID string) {
	p.podID = ID
}

//GetNamespace get pod namespace
func (p *DockerPod) GetNamespace() string {
	return p.namespace
}

//GetNetns netns path, format like /proc/$pid/net/netns
func (p *DockerPod) GetNetns() string {
	if p.netTask == nil {
		return ""
	}
	return fmt.Sprintf(defaultPodNetnsTemplate, strconv.Itoa(p.netTask.RuntimeConf.Pid))
}

//GetNetworkName netns name, like cni name or docker network name
func (p *DockerPod) GetNetworkName() string {
	return p.netTask.NetworkName
}

//GetPid get pod pid, in CNIPod, pid come from network container
func (p *DockerPod) GetPid() int {
	if p.netTask == nil {
		return 0
	}
	return p.netTask.RuntimeConf.Pid
}

//GetContainerID get pod container id, in CNIPod,
//container id is network container id
func (p *DockerPod) GetContainerID() string {
	if p.netTask == nil {
		return ""
	}
	return p.netTask.RuntimeConf.ID
}

//GetPodID get pod id
func (p *DockerPod) GetPodID() string {
	return p.podID
}

//GetNetStatus return network status
func (p *DockerPod) GetNetStatus() string {
	return ""
}

//GetPodStatus get pod status. see @PodStatus
func (p *DockerPod) GetPodStatus() container.PodStatus {
	return p.status
}

//GetMessage get pod status message
func (p *DockerPod) GetMessage() string {
	return p.message
}

//Init init pod infrastructure, in cnipod, we need start container by
//@defaultNetworkImage first. we will create network depend on this
//cotainer. all other containers will share network info with network container
func (p *DockerPod) Init() error {
	p.status = container.PodStatus_INIT
	p.message = "Pod is initing"

	envHost := container.BcsKV{
		Key:   "BCS_CONTAINER_IP",
		Value: util.GetIPAddress(),
	}
	p.netTask.Env = append(p.netTask.Env, envHost)
	//assignment for environments
	container.EnvOperCopy(p.netTask)

	cleanExtendedResourceFunc := func() {
		if p.netTask != nil {
			for _, ex := range p.netTask.ExtendedResources {
				if err := p.resourceManager.ReleaseExtendedResources(ex.Name, p.netTask.TaskId); err != nil {
					// do not break
					logs.Errorf("release extended resources %v failed, err %s", ex, err.Error())
				}
			}
		}
	}

	var extendedErr error
	//if task contains extended resources, need connect device plugin to allocate resources
	for _, ex := range p.netTask.ExtendedResources {
		logs.Infof("task %s contains extended resource %s, then allocate it", p.netTask.TaskId, ex.Name)
		envs, err := p.resourceManager.ApplyExtendedResources(ex, p.netTask.TaskId)
		if err != nil {
			logs.Errorf("apply extended resource failed, err %s", err.Error())
			extendedErr = err
			break
		}
		logs.Infof("add env %v for task %s", envs, p.netTask.TaskId)

		//append response docker envs to task.envs
		for k, v := range envs {
			kv := container.BcsKV{
				Key:   k,
				Value: v,
			}
			p.netTask.Env = append(p.netTask.Env, kv)
		}
	}
	//if allocate extended resource failed, then return and exit
	if extendedErr != nil {
		logs.Errorf(extendedErr.Error())
		p.status = container.PodStatus_FAILED
		p.message = extendedErr.Error()
		cleanExtendedResourceFunc()
		return extendedErr
	}

	//fix(developerJim): all containers in pod can not create PortMappings separately,
	// so we need to copy all PortMappings from other containers to Network
	// container, network container applies all PortMappings with docker
	p.copyPortMappings()

	//step: creating network container
	var createErr error
	if p.netTask.RuntimeConf, createErr = p.conClient.CreateContainer(p.netTask.Name, p.netTask); createErr != nil {
		logs.Errorf("DockerPod init failed in Creating master container. err: %s\n", createErr.Error())
		p.status = container.PodStatus_FAILED
		p.message = createErr.Error()
		cleanExtendedResourceFunc()
		return createErr
	}
	p.netTask.RuntimeConf.Status = container.ContainerStatus_CREATED
	p.netTask.RuntimeConf.Message = "container is created"
	p.netTask.RuntimeConf.Resource = p.netTask.Resource

	logs.Infof("task %s cpu %f memory %f", p.netTask.Name, p.netTask.Resource.Cpus, p.netTask.Resource.Mem)

	//setting preStart
	if p.events != nil && p.events.PreStart != nil {
		if preErr := p.events.PreStart(p.netTask); preErr != nil {
			p.conClient.RemoveContainer(p.netTask.Name, true)
			p.netTask.RuntimeConf.Status = container.ContainerStatus_EXITED
			p.netTask.RuntimeConf.Message = "container PreSetting failed: " + preErr.Error()
			p.status = container.PodStatus_FAILED
			p.message = preErr.Error()
			cleanExtendedResourceFunc()
			return preErr
		}
	}

	if err := p.conClient.StartContainer(p.netTask.RuntimeConf.ID); err != nil {
		logs.Errorln("DockerPod init failed in Starting master container, err: ", err.Error())
		p.conClient.RemoveContainer(p.netTask.RuntimeConf.ID, true)
		p.status = container.PodStatus_FAILED
		p.message = err.Error()
		p.netTask.RuntimeConf.Status = container.ContainerStatus_EXITED
		p.netTask.RuntimeConf.Message = "container start failed: " + err.Error()
		cleanExtendedResourceFunc()
		return err
	}
	//todo(developerJim): is it useful to check status? or just waiting for containerMonitor
	info, conErr := p.conClient.InspectContainer(p.netTask.RuntimeConf.ID)
	if conErr != nil {
		logs.Errorln("DockerPod init failed in inspecting master container, err: ", conErr.Error())
		p.status = container.PodStatus_FAILED
		p.message = conErr.Error()
		cleanExtendedResourceFunc()
		return conErr
	}
	if info.Status != container.ContainerStatus_RUNNING {
		logs.Errorf("DockerPod init stage failed, inspectContainer %s, but %s needed\n", info.Status, container.ContainerStatus_RUNNING)
		p.status = container.PodStatus_FAILED
		p.message = "docker pod master container init failed"
		cleanExtendedResourceFunc()
		return fmt.Errorf("docker pod master container init failed")
	}
	p.cnmIPAddr = info.IPAddress
	p.netTask.RuntimeConf.Message = "container is starting"
	p.netTask.RuntimeConf.NodeAddress = util.GetIPAddress()
	p.netTask.RuntimeConf.IPAddress = info.IPAddress
	p.netTask.RuntimeConf.NetworkMode = info.NetworkMode
	p.runningContainer[p.netTask.RuntimeConf.Name] = p.netTask.RuntimeConf
	logs.Infof("DockerPod treat container [%s] net container, ip: %s\n", p.netTask.RuntimeConf.Name, p.cnmIPAddr)
	return nil
}

//Finit dockerpod finit, nothing to be release.
//keep it empty
func (p *DockerPod) Finit() error {
	for _, task := range p.conTasks {
		for _, ex := range task.ExtendedResources {
			if err := p.resourceManager.ReleaseExtendedResources(ex.Name, task.TaskId); err != nil {
				// do not break
				logs.Errorf("release extended resources %v failed, err %s", ex, err.Error())
			}
		}
	}
	if p.netTask != nil {
		for _, ex := range p.netTask.ExtendedResources {
			if err := p.resourceManager.ReleaseExtendedResources(ex.Name, p.netTask.TaskId); err != nil {
				// do not break
				logs.Errorf("release extended resources %v failed, err %s", ex, err.Error())
			}
		}
	}
	return nil
}

//Start Starting all containers
func (p *DockerPod) Start() error {
	p.status = container.PodStatus_STARTING
	p.message = "Pod is starting"
	//add pod ipaddr to ENV
	envHost := container.BcsKV{
		Key:   "BCS_CONTAINER_IP",
		Value: util.GetIPAddress(),
	}

	logs.Infof("docker pod start container...")

	for name, task := range p.conTasks {
		//create container attach to network infrastructure
		task.NetworkName = "container:" + p.GetContainerID()
		task.RuntimeConf = &container.BcsContainerInfo{
			Name: name,
		}
		task.Env = append(task.Env, envHost)
		//assignment for environments
		container.EnvOperCopy(task)
		var extendedErr error
		//if task contains extended resources, need connect device plugin to allocate resources
		for _, ex := range task.ExtendedResources {
			logs.Infof("task %s contains extended resource %s, then allocate it", task.TaskId, ex.Name)
			envs, err := p.resourceManager.ApplyExtendedResources(ex, p.netTask.TaskId)
			if err != nil {
				logs.Errorf("apply extended resource failed, err %s", err.Error())
				extendedErr = err
				break
			}
			logs.Infof("add env %v for task %s", envs, task.TaskId)

			//append response docker envs to task.envs
			for k, v := range envs {
				kv := container.BcsKV{
					Key:   k,
					Value: v,
				}
				task.Env = append(task.Env, kv)
			}
		}

		//if allocate extended resource failed, then return and exit
		if extendedErr != nil {
			logs.Errorf(extendedErr.Error())
			task.RuntimeConf.Status = container.ContainerStatus_EXITED
			task.RuntimeConf.Message = extendedErr.Error()
			p.startFailedStop(extendedErr)
			return extendedErr
		}
		createdInst, createErr := p.conClient.CreateContainer(name, task)
		if createErr != nil {
			logs.Errorf("DockerPod create %s with name %s failed, err: %s\n", task.Image, name, createErr.Error())
			task.RuntimeConf.Status = container.ContainerStatus_EXITED
			task.RuntimeConf.Message = createErr.Error()
			p.startFailedStop(createErr)
			return createErr
		}
		task.RuntimeConf.ID = createdInst.ID
		task.RuntimeConf.NodeAddress = util.GetIPAddress()
		task.RuntimeConf.IPAddress = p.cnmIPAddr
		task.RuntimeConf.Status = container.ContainerStatus_CREATED
		task.RuntimeConf.Message = "container created"
		task.RuntimeConf.Resource = task.Resource

		logs.Infof("task %s cpu %f mem %f", task.TaskId, task.RuntimeConf.Resource.Cpus, task.RuntimeConf.Resource.Mem)

		//crate success, event callback before start
		if p.events != nil && p.events.PreStart != nil {
			preErr := p.events.PreStart(task)
			if preErr != nil {
				logs.Errorf("DockerPod PreStart setting container %s err: %s\n", task.RuntimeConf.ID, preErr.Error())
				p.conClient.RemoveContainer(task.RuntimeConf.ID, true)
				task.RuntimeConf.Status = container.ContainerStatus_EXITED
				task.RuntimeConf.Message = preErr.Error()
				p.startFailedStop(createErr)
				return preErr
			}
		}
		//ready to starting
		if err := p.conClient.StartContainer(task.RuntimeConf.Name); err != nil {
			logs.Errorf("DockerPod Start %s with name %s failed, err: %s\n", task.Image, name, err.Error())
			task.RuntimeConf.Status = container.ContainerStatus_EXITED
			task.RuntimeConf.Message = err.Error()
			p.conClient.RemoveContainer(task.RuntimeConf.Name, true)
			p.startFailedStop(err)
			return err
		}
		task.RuntimeConf.Message = "container is starting"
		if p.events != nil && p.events.PostStart != nil {
			p.events.PostStart(task)
		}
		p.runningContainer[task.RuntimeConf.Name] = task.RuntimeConf
		logs.Infof("Pod add container %s in running container.\n", task.RuntimeConf.Name)
	}
	p.conTasks[p.netTask.Name] = p.netTask
	//all container starting, start containerWatch
	watchCxt, _ := context.WithCancel(p.podCxt)
	go p.containersWatch(watchCxt)
	return nil
}

//Stop stop all containers
func (p *DockerPod) Stop(graceExit int) {
	p.status = container.PodStatus_KILLING
	p.message = "Pod is killing all containers"
	p.lock.Lock()
	defer p.lock.Unlock()
	p.podCancel()
	logs.Infof("DockerPod prepare to stop %d running containers\n", len(p.runningContainer))
	for name := range p.runningContainer {
		//preStop event
		task := p.conTasks[name]
		if p.events != nil && p.events.PreStop != nil {
			p.events.PreStop(task)
		}
		if task.HealthCheck != nil {
			task.HealthCheck.Stop()
		}

		if err := p.conClient.StopContainer(name, task.KillPolicy); err != nil {
			logs.Errorf("DockerPod stop container %s failed: %s\n", name, err.Error())
			//todo(developerJim): if container daemon connection broken, maybe try again later
			continue
		}

		if task.AutoRemove {
			p.conClient.RemoveContainer(name, true)
		}

		logs.Infof("DockerPod stop container %s success.\n", task.RuntimeConf.Name)
		task.RuntimeConf.Status = container.ContainerStatus_EXITED
		task.RuntimeConf.Message = "container asked to exit"
		//PostStop event
		if p.events != nil && p.events.PostStop != nil {
			p.events.PostStop(task)
		}
	}

	p.status = container.PodStatus_KILLED
	p.message = "Pod is killed"
	logs.Infof("DockerPod stop %d running containers done\n", len(p.runningContainer))
}

//GetContainers get all running container's info
func (p *DockerPod) GetContainers() []*container.BcsContainerInfo {
	p.lock.Lock()
	defer p.lock.Unlock()
	var infos []*container.BcsContainerInfo
	for _, task := range p.conTasks {
		infos = append(infos, task.RuntimeConf)
	}
	return infos
}

//GetContainerByName get container info by container name
func (p *DockerPod) GetContainerByName(name string) *container.BcsContainerInfo {
	if info, ok := p.conTasks[name]; ok {
		return info.RuntimeConf
	}
	return nil
}

//UploadFile upload source file to container(name)
func (p *DockerPod) UploadFile(name, source, dest string) error {
	_, ok := p.conTasks[name]
	if !ok {
		return fmt.Errorf("DockerPod Get no container named: %s", name)
	}
	return p.conClient.UploadToContainer(name, source, dest)
}

//Execute execute command in container(name)
func (p *DockerPod) Execute(name string, command []string) error {
	_, ok := p.conTasks[name]
	if !ok {
		return fmt.Errorf("DockerPod Get no container named: %s", name)
	}
	return p.conClient.RunCommand(name, command)
}

/////////////////////////////////////////////////
//   Inner Methods
/////////////////////////////////////////////////
func (p *DockerPod) startFailedStop(err error) {
	logs.Infof("DockerPod in start failed stop, runningContainers: %d\n", len(p.runningContainer))
	for name := range p.runningContainer {
		logs.Infof("DockerPod stop container %s in startFailedStop\n", name)
		//todo(developerJim): check container real status
		var task *container.BcsContainerTask
		var ok bool
		task, ok = p.conTasks[name]
		if !ok {
			if name != p.netTask.Name {
				logs.Errorf("DockerPod lost previous starting container %s info\n", name)
				continue
			}
			task = p.netTask
		}
		p.conClient.StopContainer(name, task.KillPolicy)
		if task.AutoRemove {
			p.conClient.RemoveContainer(name, true)
		}
		task.RuntimeConf.Status = container.ContainerStatus_EXITED
		task.RuntimeConf.Message = "container exit because other container in pod failed"
		if p.netTask.Name == name {
			p.netTask.RuntimeConf.Status = container.ContainerStatus_EXITED
			p.netTask.RuntimeConf.Message = "main container exit because other container in pod failed"
		}
		delete(p.runningContainer, name)
	}
	p.status = container.PodStatus_FAILED
	p.message = err.Error()
	logs.Infoln("DockerPod startFailed stop end.")
}

//runningFailedStop failed stop for running container
func (p *DockerPod) runningFailedStop(err error) {
	logs.Infof("DockerPod in running failed stop, runningContainers: %d\n", len(p.runningContainer))
	for name := range p.runningContainer {
		logs.Infof("DockerPod stop container %s in runningFailedStop\n", name)
		//todo(developerJim): check container real status
		task, ok := p.conTasks[name]
		if !ok {
			logs.Errorf("DockerPod lost previous running container %s info\n", name)
			continue
		}
		if task.HealthCheck != nil {
			task.HealthCheck.Stop()
		}
		p.conClient.StopContainer(name, task.KillPolicy)
		/*if task.AutoRemove {
			p.conClient.RemoveContainer(name, true)
		}*/
		task.RuntimeConf.Status = container.ContainerStatus_EXITED
		task.RuntimeConf.Message = fmt.Sprintf("container exit because other container exited in pod")

		delete(p.runningContainer, name)
	}

	p.status = container.PodStatus_FAILED
	p.message = err.Error()
	logs.Infoln("DockerPod runningFailed stop end.")
}

//contianerWatch tick for watch container status
func (p *DockerPod) containersWatch(cxt context.Context) {
	//check Pod status, watch only effective under Pod_STARTING
	if p.status != container.PodStatus_STARTING {
		logs.Errorf("DockerPod status Error, request %s, but got %s, DockerPod Container watch exit\n", container.PodStatus_STARTING, p.status)
		return
	}

	tick := time.NewTicker(defaultPodWatchInterval * time.Second)
	defer tick.Stop()
	//total := defaultErrTolerate * len(p.runningContainer)
	for {
		select {
		case <-cxt.Done():
			logs.Infof("DockerPod ask to stop, DockerPod Container watch exit.")
			return
		case <-tick.C:
			if err := p.containerCheck(); err != nil {
				return
			}
		} //end select
	}
}

func (p *DockerPod) containerCheck() error {
	running := 0
	healthyCount := 0
	tolerance := 0
	p.lock.Lock()
	defer p.lock.Unlock()

	for name := range p.runningContainer {
		info, err := p.conClient.InspectContainer(name)
		if err != nil {
			//todo(developerJim): inspect error, how to handle ?
			tolerance++
			logs.Errorf("DockerPod Inspect info from container runtime Err: %s, #########wait for next tick, tolerance: %d#########\n", err.Error(), tolerance)
			continue
		}

		info.Healthy = true
		//logs.Infof("DEBUG %+v\n", info)
		if p.cnmIPAddr == "" && info.IPAddress != "" {
			//setting cnm ip address again if get nothing in Init Stage
			logs.Infof("Pod ####recovery#### CNM ip address [%s] in running state\n", info.IPAddress)
			p.cnmIPAddr = info.IPAddress
		}
		task := p.conTasks[name]
		if task.RuntimeConf.Status != info.Status {
			//status changed
			logs.Infof("DockerPod container %s status become %s from %s", name, info.Status, task.RuntimeConf.Status)
			task.RuntimeConf.Update(info)
			switch task.RuntimeConf.Status {
			case container.ContainerStatus_CREATED, container.ContainerStatus_RESTARTING:
				//todo(developerJim): update task status with TaskState_TASK_STARTING
				logs.Infof("Impossible status [%s] for DockerPod Container status watch. wait next tick", task.RuntimeConf.Status)
			case container.ContainerStatus_RUNNING, container.ContainerStatus_PAUSED:
				//update status
				task.RuntimeConf.Message = "container is running, healthy status unkown"
				if task.HealthCheck != nil && !task.HealthCheck.IsStarting() {
					//health check starting when Status become RUNNING
					logs.Infof("container [%s] is running, healthy status unkown, starting HealthyChecker, ip: %s\n", task.RuntimeConf.Name, p.cnmIPAddr)
					if task.HealthCheck.Name() == healthcheck.CommandHealthcheck {
						task.HealthCheck.SetHost(task.RuntimeConf.ID)
					} else {
						task.HealthCheck.SetHost(p.cnmIPAddr)
					}

					go task.HealthCheck.Start()
				}
				running++
				if running == len(p.runningContainer) && p.status != container.PodStatus_RUNNING {
					logs.Infoln("DockerPod status is first changing to RUNNING")
					p.status = container.PodStatus_RUNNING
					p.message = "Pod is running, but healthy status is unkown"
					logs.Infof("Pod is running, but healthy status is unkown\n")
				}
			case container.ContainerStatus_EXITED, container.ContainerStatus_DEAD:
				//one container down, update dead container and then KILL all left
				logs.Infof("DockerPod Get container %s #%s#, ready clean all containers\n", task.RuntimeConf.Name, task.RuntimeConf.Status)
				if task.HealthCheck != nil {
					task.HealthCheck.Stop()
				}
				delete(p.runningContainer, name)

				p.exitCode = task.RuntimeConf.ExitCode
				task.RuntimeConf.Message = "The container exits with an exception and you need to look at the business log location problem"

				//stop running container
				p.runningFailedStop(fmt.Errorf("Pod failed because %s", task.RuntimeConf.Message))
				return fmt.Errorf(task.RuntimeConf.Message)

				/*case container.ContainerStatus_PAUSED:
				logs.Infof("DockerPod Get container %s #%s#, ready clean all containers\n", task.RuntimeConf.Name, task.RuntimeConf.Status)
				task.RuntimeConf.Message = "container is paused"
				if task.HealthCheck != nil {
					task.HealthCheck.Stop()
				}

				//think status paused is unnormal
				p.exitCode = ContainerStatusAbnormal

				//stop running container
				p.runningFailedStop(fmt.Errorf("Pod failed because container %s paused", name))
				return fmt.Errorf("container is paused")*/
			} //end of switch
		} //end if
		if task.HealthCheck != nil && task.HealthCheck.IsStarting() {
			if task.HealthCheck.IsHealthy() {
				task.RuntimeConf.Healthy = true
				task.RuntimeConf.Message = "container is running and healthy"
				healthyCount++
			} else {
				info.Healthy = false
				task.RuntimeConf.Healthy = false
				//pod become unhealthy
				p.healthy = false
				p.message = "Pod is running, but unhealthy"
				task.RuntimeConf.Message = "container is running, but unhealthy"
			}
		}
	} //end for
	if healthyCount == len(p.runningContainer) && p.status == container.PodStatus_RUNNING {
		p.status = container.PodStatus_RUNNING
		p.message = "Pod is running, all container is healthy"
	}

	//logs.Infof("check container done")
	return nil
}

func (p *DockerPod) copyPortMappings() {
	if p.netTask.PortBindings == nil {
		p.netTask.PortBindings = make(map[string]container.BcsPort)
	}
	//copy PortMappings from p.conTasks to p.netTask
	for _, task := range p.conTasks {
		if task.PortBindings != nil && len(task.PortBindings) > 0 {
			for key, value := range task.PortBindings {
				p.netTask.PortBindings[key] = value
				logs.Infof("copy PortMappings %s info to network container\n", key)
			}
			task.PortBindings = nil
		}
	}
}

//GetNetArgs implementation
func (p *DockerPod) GetNetArgs() [][2]string {
	return nil
}

//UpdateResources update CPU or MEM resource in runtime
func (p *DockerPod) UpdateResources(id string, resource *schedTypes.TaskResources) error {
	var exist bool
	var conTask *container.BcsContainerTask

	for _, con := range p.conTasks {
		if con.RuntimeConf.ID == id {
			exist = true
			conTask = con
			break
		}
	}

	if !exist {
		return fmt.Errorf("container id %s is invalid", id)
	}

	err := p.conClient.UpdateResources(id, resource)
	if err != nil {
		return err
	}

	conTask.RuntimeConf.Resource.Cpus = *resource.ReqCpu
	conTask.RuntimeConf.Resource.Mem = *resource.ReqMem
	return nil
}

// CommitImage image commit
func (p *DockerPod) CommitImage(id, image string) error {
	return p.conClient.CommitImage(id, image)
}
