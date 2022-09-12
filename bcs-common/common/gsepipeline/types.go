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

package gsepipeline

import (
	"time"

	"github.com/docker/engine-api/types"
)

// Storage save data to gse interface
type Storage interface {
	AddStats(msg LogMsg) error
}

type eventActor struct {
	ID         string            `json:"ID"`
	Attributes map[string]string `json:"Attributes"`
}

// EventJSON event json
type EventJSON struct {
	Status   string     `json:"status"`
	ID       string     `json:"id"`
	From     string     `json:"from"`
	Type     string     `json:"Type"`
	Action   string     `json:"Action"`
	Actor    eventActor `json:"Actor"`
	Time     uint64     `json:"time"`
	TimeNano uint64     `json:"timeNano"`

	line string
}

// ContainerMount definition for container mount
type ContainerMount struct {
	Source      string
	Destination string
}

// ContainerReference Container reference contains enough information to uniquely identify a container
type ContainerReference struct {
	// The container id
	ID string `json:"id,omitempty"`

	CreateTime time.Time `json:"created,omitempty"`

	// The absolute name of the container. This is unique on the machine.
	Name string `json:"name,omitempty"`

	Image     string `json:"image,omitempty"`
	IPAddress string `json:"host,omitempty"`

	ContainerRootDirector string `json:"container_root,omitempty"`

	// Other names by which the container is known within a certain namespace.
	// This is unique within that namespace.
	Aliases []string `json:"aliases,omitempty"`

	// Namespace under which the aliases of a container are unique.
	// An example of a namespace is "docker" for Docker containers.
	Namespace string `json:"namespace,omitempty"`

	Labels map[string]string `json:"labels,omitempty"`

	ContainerInfo types.ContainerJSON `json:"-"`
}

// LogMsg log object
type LogMsg struct {
	FileName     string              `json:"filename,omitempty"`
	Timestamp    time.Time           `json:"timestamp"`
	Stream       string              `json:"stream"`
	Log          interface{}         `json:"log"`
	DataID       uint64              `json:"dataid"`
	ContainerRef *ContainerReference `json:"container_info,omitempty"`
}
