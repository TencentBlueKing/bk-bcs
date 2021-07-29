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

package types

import (
	"strings"
	"time"
)

const (
	ContainerStatus_PAUSED     = "paused"
	ContainerStatus_RESTARTING = "restarting"
	ContainerStatus_RUNNING    = "running"
	ContainerStatus_DEAD       = "dead"
	ContainerStatus_CREATED    = "created"
	ContainerStatus_EXITED     = "exited"
)

// BcsPort port service for container port reflection
type BcsPort struct {
	Name          string `json:"name,omitempty"`
	ContainerPort string `json:"containerPort,omitempty"`
	HostPort      string `json:"hostPort,omitempty"`
	Protocol      string `json:"protocol,omitempty"`
	HostIP        string `json:"hostIP,omitempty"` //use for host has multiple ip address
}

// BcsImage image info from hub
type BcsImage struct {
	ID         string   //image id
	Repository []string //repository, including tag
	Created    int64    //create time
	Size       int64    //image size
}

// BcsContainerInfo only for BcsExecutor
type BcsContainerInfo struct {
	ID                      string      `json:"ID,omitempty"`          //container ID
	Name                    string      `json:"Name,omitempty"`        //container name
	Pid                     int         `json:"Pid,omitempty"`         //container pid
	StartAt                 time.Time   `json:"StartAt,omitempty"`     //startting time
	FinishAt                time.Time   `json:"FinishAt,omitempty"`    //Exit time
	Status                  string      `json:"Status,omitempty"`      //status string, paused, restarting, running, dead, created, exited
	Healthy                 bool        `json:"Healthy,omitempty"`     //Container healthy
	IsChecked               bool        `json:",omitempty"`            //is health check
	ConsecutiveFailureTimes int         `json:",omitempty"`            //consecutive failure times
	ExitCode                int         `json:"ExitCode,omitempty"`    //container exit code
	Hostname                string      `json:"Hostname,omitempty"`    //container host name
	NetworkMode             string      `json:"NetworkMode,omitempty"` //Network mode for container
	IPAddress               string      `json:"IPAddress,omitempty"`   //Contaienr IP address
	NodeAddress             string      `json:"NodeAddress,omitempty"` //node host address
	Ports                   []BcsPort   `json:"Ports,omitempty"`       //ports info for report
	Message                 string      `json:"Message,omitempty"`     //status message for container
	Resource                *Resource   `json:"Resource,omitempty"`
	BcsMessage              *BcsMessage `json:",omitempty"`
	OOMKilled               bool        `json:"OOMKilled,omitempty"` //container exited, whether oom
}

// Update data from other info
func (info *BcsContainerInfo) Update(other *BcsContainerInfo) {
	if info.Name != other.Name {
		return
	}
	info.ID = other.ID
	info.Pid = other.Pid
	info.StartAt = other.StartAt
	info.FinishAt = other.FinishAt
	info.Healthy = other.Healthy
	info.Status = other.Status
	info.ExitCode = other.ExitCode
	info.Hostname = other.Hostname
	info.IsChecked = other.IsChecked
	info.ConsecutiveFailureTimes = other.ConsecutiveFailureTimes
	info.OOMKilled = other.OOMKilled
	if strings.Contains(other.NetworkMode, "container:") {
		info.NetworkMode = "user"
	} else {
		info.NetworkMode = other.NetworkMode
	}
	if other.IPAddress != "" {
		info.IPAddress = other.IPAddress
	}
	if other.NodeAddress != "" {
		info.NodeAddress = other.NodeAddress
	}
	if other.Message != "" {
		info.Message = other.Message
	}
	if other.Resource != nil {
		info.Resource = other.Resource
	}
	if other.BcsMessage != nil {
		info.BcsMessage = other.BcsMessage
	}
}
