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

package container

import (
	schedTypes "github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/schetypes"
)

//PodStatus for imply container status in pod
type PodStatus string

const (
	PodStatus_UNKNOWN  PodStatus = "unknown"  //init status code
	PodStatus_INIT     PodStatus = "init"     //INIT pod, network infrastructure
	PodStatus_STARTING PodStatus = "starting" //pulling images, starting all containers
	PodStatus_RUNNING  PodStatus = "running"  //all containers are running
	PodStatus_FAILED   PodStatus = "failed"   //one container failed
	PodStatus_KILLING  PodStatus = "killing"  //get kill command, killing all container
	PodStatus_KILLED   PodStatus = "killed"   //all container killed
	PodStatus_FINISH   PodStatus = "finish"   //get shutdown command, close all container
)

//ConEventCB callback for every container status changed
type ConEventCB func(*BcsContainerTask) error

//PodEventCB pod event callback
type PodEventCB func(Pod) error

//PodEventHandler callback collection for pod status changed
type PodEventHandler struct {
	PreStart  ConEventCB //call before container start
	PostStart ConEventCB //call after container start success
	PreStop   ConEventCB //call before container stop, not including failed stop
	PostStop  ConEventCB //call after container stopped, not including failed stop
}

//Pod interface for bcs
type Pod interface {
	IsHealthy() bool                                  //check pod is healthy
	Injection() bool                                  //check ip address is injected
	SetIPAddr(ip string)                              //set Pod ip address
	GetIPAddr() string                                //get pod ip address if exist
	SetPodID(ID string)                               //set pod id if needed
	GetNamespace() string                             //get namespace, not used now
	GetNetns() string                                 //get netns path string
	GetNetworkName() string                           //get container network name
	GetNetArgs() [][2]string                          //get cni plugin args
	GetPid() int                                      //get network infrastructure container pid
	GetContainerID() string                           //get network infrastructure container id
	GetPodID() string                                 //get pod id
	GetNetStatus() string                             //get network ip address information
	GetPodStatus() PodStatus                          //pod status, see @PodStatus
	GetMessage() string                               //Get status message
	Init() error                                      //init pod, start network container
	Finit() error                                     //finit pod, release pod resource
	Start() error                                     //start all container
	Stop(graceExit int)                               //stop all container, graceExit
	UploadFile(name, source, dest string) error       //upload source file to container(name)
	Execute(name string, command []string) error      //execute command in container(name)
	GetContainers() []*BcsContainerInfo               //Get all containers info
	GetContainerByName(name string) *BcsContainerInfo //Get container by ID
	//update container resources limit
	//para1: container id
	//para2: resources,cpu mem
	UpdateResources(string, *schedTypes.TaskResources) error
	//commit image
	//para1: container id
	//para2: image name
	CommitImage(string, string) error
	GetContainerTasks() map[string]*BcsContainerTask
}
