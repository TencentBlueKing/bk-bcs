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
	"strings"

	comtypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	bcstypes "github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/schetypes"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-container-executor/healthcheck"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-container-executor/logs"
)

//BcsPort port service for container port reflection
type BcsPort struct {
	Name          string `json:"name,omitempty"`
	ContainerPort string `json:"containerPort,omitempty"`
	HostPort      string `json:"hostPort,omitempty"`
	Protocol      string `json:"protocol,omitempty"`
	HostIP        string `json:"hostIP,omitempty"` //use for host has multiple ip address
}

//BcsKV key/value structure for anywhere necessary
type BcsKV struct {
	Key   string
	Value string
}

//BcsVolume volume info from mesos to container
type BcsVolume struct {
	ReadOnly      bool
	HostPath      string
	ContainerPath string
}

//BcsContainerTask task info for running container
type BcsContainerTask struct {
	Name              string                       //container name
	Image             string                       //container image
	HostName          string                       //container hostname
	Hosts             []string                     //host:ip pair for /etc/hosts in container
	Command           string                       //container command
	Args              []string                     //args for command
	Env               []BcsKV                      //environments
	Volums            []BcsVolume                  //host path mount
	NetworkName       string                       //container network name
	NetworkIPAddr     string                       //container ip address request
	ForcePullImage    bool                         //pull image every time
	OOMKillDisabled   bool                         //OOM kill feature, default is true
	AutoRemove        bool                         //remove container when exit, default false
	Ulimits           []BcsKV                      //ulimit for docker parameter
	ShmSize           int64                        //docker hostconfig shm size, 1 = 1B
	Privileged        bool                         //setting container privileged
	PublishAllPorts   bool                         //publish all ports in container
	PortBindings      map[string]BcsPort           //port for container reflection, only useful for docker bridge
	Labels            []BcsKV                      //label for container
	Resource          *bcstypes.Resource           //container resource request
	LimitResource     *bcstypes.Resource           // container resource limit
	ExtendedResources []*comtypes.ExtendedResource //extended resources
	BcsMessages       []*bcstypes.BcsMessage       //bcs define message
	RuntimeConf       *BcsContainerInfo            //container runtime info
	HealthCheck       healthcheck.Checker          //for health check
	KillPolicy        int                          //kill policy timeout, unit is seconds
	//container network flow limit args
	NetLimit *comtypes.NetLimit
	TaskId   string

	Ipc string //IPC namespace to use

	/*Healthy                 bool `json:"Healthy,omitempty"` //Container healthy
	IsChecked               bool `json:",omitempty"`        //is health check
	ConsecutiveFailureTimes int  `json:",omitempty"`        //consecutive failure times*/
}

//EnvOperCopy environment $ operation support
func EnvOperCopy(task *BcsContainerTask) {
	tmp := make(map[string]string)
	for _, e := range task.Env {
		tmp[e.Key] = e.Value
	}
	for index, e := range task.Env {
		if strings.HasPrefix(e.Value, "$") {
			key := e.Value[1:]
			if dest, ok := tmp[key]; ok {
				e.Value = dest
				task.Env[index] = e
				logs.Infof("BcsContainerTask setting %s=%s, vlaue comes from %s\n", e.Key, e.Value, key)
			}
		}
	}
}
