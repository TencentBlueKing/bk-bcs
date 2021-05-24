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

package cni

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	comtypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-common/common/util"
	bcstypes "github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/schetypes"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-container-executor/container"
	devicepluginmanager "github.com/Tencent/bk-bcs/bcs-mesos/bcs-container-executor/devicepluginmanager"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-container-executor/extendedresource"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-container-executor/healthcheck"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-container-executor/logs"
	exeutil "github.com/Tencent/bk-bcs/bcs-mesos/bcs-container-executor/util"

	//"github.com/pborman/uuid"
	schedTypes "github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/schetypes"

	"golang.org/x/net/context"
)

const (
	defaultNetworkImage = "bcs/network"
	defaultNetworkTag   = "latest"
	//defaultPodNamePrex         = "bcs_pod"
	defaultNetworkNamePrex     = "bcs-network-"
	defaultPodNetnsTemplate    = "/proc/%s/ns/net"
	defaultDockerTypeKey       = "io.mesos.docker.type"
	defaultDockerTypeContainer = "container"
	defaultDockerTypeNetwork   = "network"
)

const (
	//defaultGracefullExit    = 1  //graceful exit for container
	defaultPodWatchInterval = 1 //interval for watch container's status
	//defaultErrTolerate      = 60 //time for determining report container runtime failed
)

const (
	ContainerStatusAbnormal = 1
)

const (
	PodNetEgressLimitName = "CTNR_NET_EGRESS_RATE"
)

const (
	BcsNetworkContainer     = "BcsNetworkContainer"
	BcsNetworkContainerTrue = "true"
	BcsNetworkMode          = "networkMode"
)

//NewPod create CNIPod instance with container interaface and container info
func NewPod(operator container.Container, tasks []*container.BcsContainerTask,
	handler *container.PodEventHandler, netImage string,
	extendedResourceDriver *extendedresource.Driver) container.Pod {

	if len(tasks) == 0 {
		logs.Errorf("Create CNIPod error, Container tasks are 0")
		return nil
	}
	taskMap := make(map[string]*container.BcsContainerTask)
	for _, task := range tasks {
		taskMap[task.Name] = task
	}
	rCxt, rCancel := context.WithCancel(context.Background())
	pod := &CNIPod{
		healthyChanged:   false,
		healthy:          true,
		injected:         false,
		cniIPAddr:        tasks[0].NetworkIPAddr,
		status:           container.PodStatus_UNKNOWN,
		podCxt:           rCxt,
		podCancel:        rCancel,
		events:           handler,
		networkName:      tasks[0].NetworkName,
		conClient:        operator,
		conTasks:         taskMap,
		networkTaskId:    tasks[0].TaskId,
		runningContainer: make(map[string]*container.BcsContainerInfo),
		resourceManager: devicepluginmanager.NewResourceManager(
			devicepluginmanager.NewDevicePluginManager(),
			extendedResourceDriver),
	}
	if len(tasks[0].NetworkIPAddr) != 0 {
		//ip injected by executor
		pod.injected = true
	}

	if netImage != "" {
		pod.netImage = netImage
	} else {
		pod.netImage = fmt.Sprintf("%s:%s", defaultNetworkImage, defaultNetworkTag)
	}

	pod.NetLimit = tasks[0].NetLimit
	return pod
}

//CNIPod combination for mutiple container
type CNIPod struct {
	healthyChanged   bool                //if healthy changed
	healthy          bool                //pod healthy status
	podID            string              //pod id
	namespace        string              //namespace from scheduler
	injected         bool                //ip source
	cniIPAddr        string              //ip address from cni tools
	cniHostName      string              //host name for network infrastructure
	netns            string              //netns path string
	status           container.PodStatus //status for all container
	exitCode         int
	message          string                                 //status message
	podCxt           context.Context                        //pod root context
	podCancel        context.CancelFunc                     //pod stop func
	lock             sync.Mutex                             //lock for monitor & query
	events           *container.PodEventHandler             //pod event changed callback collection
	networkName      string                                 //name for network
	netTask          *container.BcsContainerTask            //network container task, using in Init stage
	conClient        container.Container                    //container operator interface
	conTasks         map[string]*container.BcsContainerTask //task for running containers, key is taskID
	runningContainer map[string]*container.BcsContainerInfo //running container ID list for monitor
	//container network flow limit args
	NetLimit      *comtypes.NetLimit
	networkTaskId string
	netImage      string
	//device plugin manager
	resourceManager *devicepluginmanager.ResourceManager
}

//IsHealthy check pod is healthy
func (p *CNIPod) IsHealthy() bool {
	return p.healthy
}

func (p *CNIPod) GetContainerTasks() map[string]*container.BcsContainerTask {
	return p.conTasks
}

//Injection check ip address is injected
func (p *CNIPod) Injection() bool {
	return p.injected
}

//SetIPAddr set ip addr for pod
func (p *CNIPod) SetIPAddr(ip string) {
	p.cniIPAddr = ip
}

//GetIPAddr get pod ip address if exist
func (p *CNIPod) GetIPAddr() string {
	return p.cniIPAddr
}

//SetPodID set pod id if needed
func (p *CNIPod) SetPodID(ID string) {
	p.podID = ID
}

//GetNamespace get pod namespace
func (p *CNIPod) GetNamespace() string {
	return p.namespace
}

//GetNetns netns path, format like /proc/$pid/net/netns
func (p *CNIPod) GetNetns() string {
	if p.netTask == nil {
		return ""
	}
	return fmt.Sprintf(defaultPodNetnsTemplate, strconv.Itoa(p.netTask.RuntimeConf.Pid))
}

//GetNetworkName netns name, like cni name or docker network name
func (p *CNIPod) GetNetworkName() string {
	return p.networkName
}

//GetPid get pod pid, in CNIPod, pid come from network container
func (p *CNIPod) GetPid() int {
	if p.netTask == nil {
		return 0
	}
	return p.netTask.RuntimeConf.Pid
}

//GetContainerID get pod container id, in CNIPod,
//container id is network container id
func (p *CNIPod) GetContainerID() string {
	if p.netTask == nil {
		return ""
	}
	return p.netTask.RuntimeConf.ID
}

//GetPodID get pod id
func (p *CNIPod) GetPodID() string {
	return p.podID
}

//GetNetStatus return network status
func (p *CNIPod) GetNetStatus() string {
	return ""
}

//GetPodStatus get pod status. see @PodStatus
func (p *CNIPod) GetPodStatus() container.PodStatus {
	return p.status
}

//GetMessage get pod status message
func (p *CNIPod) GetMessage() string {
	return p.message
}

//Init init pod infrastructure, in cnipod, we need start container by
//@defaultNetworkImage first. we will create network depend on this
//cotainer. all other containers will share network info with network container
func (p *CNIPod) Init() error {
	p.status = container.PodStatus_INIT
	p.message = "Pod is initing"
	//step 1: construct container task for network container

	for _, task := range p.conTasks {
		if task.HostName != "" {
			p.cniHostName = task.HostName
			break
		}
	}

	if p.cniHostName == "" {
		p.cniHostName = "pod-" + exeutil.RandomString(12)
	}

	p.netTask = &container.BcsContainerTask{
		//Name:           defaultPodNamePrex + "-" + uuid.NewUUID().String(),
		Name:           defaultNetworkNamePrex + p.networkTaskId,
		Image:          p.netImage,
		ForcePullImage: false,
		Privileged:     false,
		NetworkName:    "none",
		HostName:       p.cniHostName,
	}
	p.netTask.Resource = &bcstypes.Resource{
		Cpus: 1,
	}

	netflag := container.BcsKV{
		Key:   BcsNetworkContainer,
		Value: BcsNetworkContainerTrue,
	}
	netMode := container.BcsKV{
		Key:   BcsNetworkMode,
		Value: p.GetNetworkName(),
	}
	dockerType := container.BcsKV{
		Key:   defaultDockerTypeKey,
		Value: defaultDockerTypeNetwork,
	}
	p.netTask.Labels = []container.BcsKV{netflag, netMode, dockerType}

	for _, task := range p.conTasks {
		p.netTask.Labels = append(p.netTask.Labels, task.Labels...)
	}

	//step 2: starting network container
	var createErr error
	if p.networkName == "host" {
		//for network infrastructure, CNI tools set
		//nothing in `host` mode, otherwise, keep
		//`none` and wait setting from CNI
		p.netTask.NetworkName = p.networkName
	}
	logs.Infof("CNIPod ready to init network infra, image: %s, name: %s, network name: %s\n", p.netTask.Image, p.netTask.Name, p.netTask.NetworkName)
	if p.netTask.RuntimeConf, createErr = p.conClient.CreateContainer(p.netTask.Name, p.netTask); createErr != nil {
		logs.Errorf("CNIPod init failed in Creating network infrastructure container. err: %s\n", createErr.Error())
		p.status = container.PodStatus_FAILED
		p.message = createErr.Error()
		return createErr
	}
	p.netTask.RuntimeConf.Status = container.ContainerStatus_CREATED
	if err := p.conClient.StartContainer(p.netTask.RuntimeConf.ID); err != nil {
		logs.Errorln("CNIPod init failed in Starting network infrastructure container, err: ", err.Error())
		p.status = container.PodStatus_FAILED
		p.message = err.Error()
		p.conClient.RemoveContainer(p.netTask.RuntimeConf.ID, true)
		return err
	}
	info, conErr := p.conClient.InspectContainer(p.netTask.RuntimeConf.ID)
	if conErr != nil {
		logs.Errorln("CNIPod init failed in inspecting network infrastructure container, err: ", conErr.Error())
		p.status = container.PodStatus_FAILED
		p.message = conErr.Error()
		return conErr
	}
	if info.Status != container.ContainerStatus_RUNNING {
		logs.Errorf("CNIPod init stage failed, inspectContainer %s, but %s needed\n", info.Status, container.ContainerStatus_RUNNING)
		p.status = container.PodStatus_FAILED
		p.message = "cni pod network infrastructure init failed"
		return fmt.Errorf("cni pod network infrastructure init failed")
	}
	p.netTask.RuntimeConf.Update(info)
	//note: now Pod do not need to watch network infrastructure container status
	//network infrastructure is manager by NetworkManager
	//todo(developerJim): network infrastructure is under risk if somebody kill it manually
	//             Pod must recored network info in infrastructure, if network infrastructure
	//             exits(under normal conditions, it's impossible), NetworkManager
	//             will release network info with CNI API
	return nil
}

//Finit cni pod finit, close network infrastructure
func (p *CNIPod) Finit() error {
	if p.netTask != nil && p.netTask.RuntimeConf != nil && p.netTask.RuntimeConf.ID != "" {
		//kill network infrastructure
		if err := p.conClient.StopContainer(p.netTask.RuntimeConf.ID, 1); err != nil {
			//todo(developerJim): stop network container failed. need to try again
			//or check network container status
			return err
		}
		p.conClient.RemoveContainer(p.netTask.RuntimeConf.ID, true)
		logs.Infof("CNIPod finit network container %s success\n", p.netTask.RuntimeConf.ID)
		return nil
	}
	for _, task := range p.conTasks {
		for _, ex := range task.ExtendedResources {
			if err := p.resourceManager.ReleaseExtendedResources(ex.Name, task.TaskId); err != nil {
				// do not break
				logs.Errorf("release extended resources %v failed, err %s", ex, err.Error())
			}
		}
	}
	logs.Infoln("CNIPod nothing can be finit.")
	return nil
}

//Start Starting all containers
func (p *CNIPod) Start() error {
	p.status = container.PodStatus_STARTING
	p.message = "Pod is starting"
	for name, task := range p.conTasks {
		//create container attach to network infrastructure
		task.NetworkName = "container:" + p.GetContainerID()
		task.RuntimeConf = &container.BcsContainerInfo{
			Name: name,
		}

		if task.Labels == nil {
			task.Labels = make([]container.BcsKV, 0)
		}

		dockerType := container.BcsKV{
			Key:   defaultDockerTypeKey,
			Value: defaultDockerTypeContainer,
		}
		task.Labels = append(task.Labels, dockerType)

		if p.networkName != "host" && p.networkName != "none" {
			//todo(developerJim): docker api has bug for setting hostname & extracthosts after v1.9, so
			//when network is not setting host or none, create custom /etc/hosts file for cni pod container binding.
			//we need to fix this issue someday
			etcFile, err := p.createEtcHosts(p.cniIPAddr, p.cniHostName)
			if err != nil {
				task.RuntimeConf.Status = container.ContainerStatus_EXITED
				task.RuntimeConf.Message = fmt.Sprintf("pod create /etc/hosts failed, %s", err.Error())
				p.startFailedStop(err)
				return fmt.Errorf("pod create /etc/hosts failed, %s", err.Error())
			}
			hostsBinding := container.BcsVolume{
				HostPath:      etcFile,
				ContainerPath: "/etc/hosts",
				ReadOnly:      true,
			}
			task.Volums = append(task.Volums, hostsBinding)
		}
		//add pod ipaddr to ENV
		envHost := container.BcsKV{
			Key:   "BCS_CONTAINER_IP",
			Value: p.cniIPAddr,
		}
		task.Env = append(task.Env, envHost)
		//assignment all env before create
		container.EnvOperCopy(task)

		hostname := task.HostName
		task.HostName = ""
		var extendedErr error
		//if task contains extended resources, need connect device plugin to allocate resources
		for _, ex := range task.ExtendedResources {
			logs.Infof("task %s contains extended resource %s, then allocate it", task.TaskId, ex.Name)
			envs, err := p.resourceManager.ApplyExtendedResources(ex, task.TaskId)
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

		createInst, createErr := p.conClient.CreateContainer(name, task)
		if createErr != nil {
			logs.Errorf("CNIPod create %s with name %s failed, err: %s\n", task.Image, name, createErr.Error())
			task.RuntimeConf.Status = container.ContainerStatus_EXITED
			task.RuntimeConf.Message = createErr.Error()
			p.startFailedStop(createErr)
			return createErr
		}
		task.HostName = hostname

		task.RuntimeConf.ID = createInst.ID
		task.RuntimeConf.NodeAddress = util.GetIPAddress()
		task.RuntimeConf.IPAddress = p.cniIPAddr
		task.RuntimeConf.Status = container.ContainerStatus_CREATED
		task.RuntimeConf.Message = "container created"
		//crate success, event callback before start
		if p.events != nil && p.events.PreStart != nil {
			preErr := p.events.PreStart(task)
			if preErr != nil {
				logs.Errorf("Pod PreStart setting container %s err: %s\n", task.RuntimeConf.ID, preErr.Error())
				p.conClient.RemoveContainer(task.RuntimeConf.ID, true)
				task.RuntimeConf.Status = container.ContainerStatus_EXITED
				task.RuntimeConf.Message = preErr.Error()
				p.startFailedStop(createErr)
				return preErr
			}
		}
		//ready to starting
		if err := p.conClient.StartContainer(task.RuntimeConf.Name); err != nil {
			logs.Errorf("CNIPod Start %s with name %s failed, err: %s\n", task.Image, name, err.Error())
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
		//starting health
		logs.Infof("Pod add container %s in running container.\n", task.RuntimeConf.Name)
	}
	//all container starting, start containerWatch
	watchCxt, _ := context.WithCancel(p.podCxt)
	go p.containersWatch(watchCxt)
	return nil
}

//Stop stop all containers
func (p *CNIPod) Stop(graceExit int) {
	p.status = container.PodStatus_KILLING
	p.message = "Pod is killing all containers"
	p.lock.Lock()
	defer p.lock.Unlock()
	p.podCancel()
	logs.Infof("CNIPod prepare to stop %d running containers\n", len(p.runningContainer))
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
			logs.Errorf("CNIPod stop container %s failed: %s\n", name, err.Error())
			//todo(developerJim): if container daemon connection broken, maybe try again later
			continue
		}
		if task.AutoRemove {
			p.conClient.RemoveContainer(name, true)
		}
		task.RuntimeConf.Status = container.ContainerStatus_EXITED
		task.RuntimeConf.Message = "container asked to exit"
		//PostStop event
		if p.events != nil && p.events.PostStop != nil {
			p.events.PostStop(task)
		}
		delete(p.runningContainer, name)
	}
	p.status = container.PodStatus_KILLED
	p.message = "Pod is killed"
	logs.Infof("CNIPod stop %d running containers done\n", len(p.runningContainer))
}

//GetContainers get all running container's info
func (p *CNIPod) GetContainers() []*container.BcsContainerInfo {
	p.lock.Lock()
	defer p.lock.Unlock()
	var infos []*container.BcsContainerInfo
	for _, task := range p.conTasks {
		infos = append(infos, task.RuntimeConf)
	}
	return infos
}

//GetContainerByName get container info by container name
func (p *CNIPod) GetContainerByName(name string) *container.BcsContainerInfo {
	if info, ok := p.conTasks[name]; ok {
		return info.RuntimeConf
	}
	return nil
}

//UploadFile upload source file to container(name)
func (p *CNIPod) UploadFile(name, source, dest string) error {
	_, ok := p.conTasks[name]
	if !ok {
		return fmt.Errorf("CNIPod Get no container named: %s", name)
	}
	return p.conClient.UploadToContainer(name, source, dest)
}

//Execute execute command in container(name)
func (p *CNIPod) Execute(name string, command []string) error {
	_, ok := p.conTasks[name]
	if !ok {
		return fmt.Errorf("CNIPod Get no container named: %s", name)
	}
	return p.conClient.RunCommand(name, command)
}

/////////////////////////////////////////////////
//   Inner Methods
/////////////////////////////////////////////////
func (p *CNIPod) startFailedStop(err error) {
	logs.Infof("CNIPod in start failed stop, runningContainers: %d\n", len(p.runningContainer))
	for name := range p.runningContainer {
		logs.Infof("DockerPod stop container %s in startFailedStop\n", name)
		//todo(developerJim): check container real status
		task, ok := p.conTasks[name]
		if !ok {
			logs.Errorf("CNIPod lost previous running container %s info\n", name)
			continue
		}
		p.conClient.StopContainer(name, task.KillPolicy)
		if task.AutoRemove {
			p.conClient.RemoveContainer(name, true)
		}
		task.RuntimeConf.Status = container.ContainerStatus_EXITED
		task.RuntimeConf.Message = "container exit because other container in pod failed"
		delete(p.runningContainer, name)
	}
	p.status = container.PodStatus_FAILED
	p.message = err.Error()
	logs.Infoln("CNIPod startFailed stop end.")
}

//runningFailedStop failed stop for running container
func (p *CNIPod) runningFailedStop(err error) {
	logs.Infof("CNIPod in running failed stop, runningContainers: %d\n", len(p.runningContainer))
	for name := range p.runningContainer {
		logs.Infof("CNIPod stop container %s in runningFailedStop\n", name)
		//todo(developerJim): check container real status
		task, ok := p.conTasks[name]
		if !ok {
			logs.Errorf("CNIPod lost previous running container %s info\n", name)
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
	logs.Infoln("CNIPod runningFailed stop end.")
}

//contianerWatch tick for watch container status
func (p *CNIPod) containersWatch(cxt context.Context) {
	//check Pod status, watch only effective under Pod_STARTING
	if p.status != container.PodStatus_STARTING {
		logs.Errorf("CNIPod status Error, request %s, but got %s, CNIPod Container watch exit\n", container.PodStatus_STARTING, p.status)
		return
	}

	tick := time.NewTicker(defaultPodWatchInterval * time.Second)
	defer tick.Stop()
	for {
		select {
		case <-cxt.Done():
			logs.Infof("CNIPod ask to stop, CNIPod Container watch exit.")
			return
		case <-tick.C:
			if err := p.containerCheck(); err != nil {
				logs.Infof("Pod watch for container exit: %s", err.Error())
				return
			}
		} //end select
	}
}

func (p *CNIPod) containerCheck() error {
	running := 0
	healthyCount := 0
	p.lock.Lock()
	defer p.lock.Unlock()

	tolerance := 0
	for name := range p.runningContainer {
		info, err := p.conClient.InspectContainer(name)
		if err != nil {
			//inspect error
			tolerance++
			logs.Errorf("CNIPod Inspect info from container runtime Err: %s, #########wait for next tick, tolerance: %d#########\n", err.Error(), tolerance)
			continue
		}

		info.Healthy = true
		task := p.conTasks[name]
		if task.RuntimeConf.Status != info.Status {
			//status changed
			logs.Infof("CNIPod container %s status become %s from %s\n", name, info.Status, task.RuntimeConf.Status)
			task.RuntimeConf.Update(info)
			switch task.RuntimeConf.Status {
			case container.ContainerStatus_CREATED, container.ContainerStatus_RESTARTING:
				//todo(developerJim): update task status with TaskState_TASK_STARTING
				logs.Infof("Impossible status [%s] for CNIPod Container status watch. wait next tick\n", task.RuntimeConf.Status)
			case container.ContainerStatus_RUNNING, container.ContainerStatus_PAUSED:
				//update status
				task.RuntimeConf.Message = "container is running, healthy status unkown"
				if task.HealthCheck != nil && !task.HealthCheck.IsStarting() {
					//health check starting when Status become RUNNING
					logs.Infof("container [%s] is running, healthy status unkown, starting HealthyChecker with ip: %s\n", task.RuntimeConf.Name, p.cniIPAddr)
					if task.HealthCheck.Name() == healthcheck.CommandHealthcheck {
						task.HealthCheck.SetHost(task.RuntimeConf.ID)
					} else {
						task.HealthCheck.SetHost(p.cniIPAddr)
					}
					go task.HealthCheck.Start()
				}
				running++
				if running == len(p.runningContainer) && p.status != container.PodStatus_RUNNING {
					p.status = container.PodStatus_RUNNING
					p.message = "Pod is running, but healthy status is unkown"
					logs.Infof("Pod is running, but healthy status is unkown\n")
				}
			case container.ContainerStatus_EXITED, container.ContainerStatus_DEAD:
				//one container down, update dead container and then KILL all left
				logs.Infof("CNIPod Get container %s #%s#, ready clean all containers\n", task.RuntimeConf.Name, task.RuntimeConf.Status)
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
				logs.Infof("CNIPod Get container %s #%s#, ready clean all containers\n", task.RuntimeConf.Name, task.RuntimeConf.Status)
				task.RuntimeConf.Message = "container is paused"
				if task.HealthCheck != nil {
					task.HealthCheck.Stop()
				}

				// i think status paused is unnormal
				p.exitCode = ContainerStatusAbnormal

				//stop running container
				p.runningFailedStop(fmt.Errorf("Pod failed because container %s paused", name))
				return fmt.Errorf("container is down")*/
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
	return nil
}

func (p *CNIPod) createEtcHosts(ip, hostname string) (string, error) {
	hostsFile := filepath.Join(os.Getenv("MESOS_SANDBOX"), "bcs-etc-hosts")
	if _, err := os.Stat(hostsFile); os.IsExist(err) {
		logs.Infof("CNIPod check bcs-etc-hosts exist, skip\n")
		return hostsFile, nil
	}
	var buffer bytes.Buffer
	buffer.WriteString("#bcs-container-executor hosts file\n")
	buffer.WriteString("127.0.0.1\tlocalhost\n")                      // ipv4 localhost
	buffer.WriteString("::1\tlocalhost ip6-localhost ip6-loopback\n") // ipv6 localhost
	buffer.WriteString("fe00::0\tip6-localnet\n")
	buffer.WriteString("fe00::0\tip6-mcastprefix\n")
	buffer.WriteString("fe00::1\tip6-allnodes\n")
	buffer.WriteString("fe00::2\tip6-allrouters\n")
	buffer.WriteString(fmt.Sprintf("%s\t%s\n", ip, hostname))
	if err := ioutil.WriteFile(hostsFile, buffer.Bytes(), 0644); err != nil {
		logs.Errorf("CNIPod write bcs-etc-hosts file, err: %v\n", err)
		return "", err
	}
	return hostsFile, nil
}

func (p *CNIPod) GetNetArgs() [][2]string {
	args := make([][2]string, 0)

	if p.NetLimit != nil && p.NetLimit.EgressLimit > 0 {
		// EgressLimit Mbps * 1024 to Kbps
		val := strconv.Itoa(p.NetLimit.EgressLimit * 1024)
		logs.Infof("CNIPod set netlimit EgressLimit %sKbps", val)

		netEgressLimit := [2]string{PodNetEgressLimitName, val}
		args = append(args, netEgressLimit)
	}

	args = append(args, [2]string{"IgnoreUnknown", "true"})

	return args
}

// UpdateResources update resources of containers
func (p *CNIPod) UpdateResources(id string, resource *schedTypes.TaskResources) error {
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

func (p *CNIPod) CommitImage(id, image string) error {
	return p.conClient.CommitImage(id, image)
}
