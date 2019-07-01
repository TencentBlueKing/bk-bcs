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
	"time"
)

type ReplicaControllerStatus string

const (
	RC_Staging   ReplicaControllerStatus = "Staging"
	RC_Deploying ReplicaControllerStatus = "Deploying"
	RC_Running   ReplicaControllerStatus = "Running"
	RC_Operating ReplicaControllerStatus = "Operating"
	RC_Finish    ReplicaControllerStatus = "Finish"
	RC_Error     ReplicaControllerStatus = "Error"
)

// BcsReplicaControllerStatus define ReplicaController status
type BcsReplicaControllerStatus struct {
	ObjectMeta `json:"metadata"`
	// Instance is the number of requested instance
	Instance int `json:"instance"`
	// BuildedInstance is the number of actual instance
	BuildedInstance int `json:"buildedInstance"`
	// RunningInstance is the number of running status instance
	RunningInstance int `json:"runningInstance"`
	// CreateTime is the date and time of begining to create ReplicaController
	CreateTime time.Time `json:"createTime"`
	// LastUpdateTime is the date and time when ReplicaController status change
	LastUpdateTime time.Time `json:"lastUpdateTime,omitempty"`
	// ReportTime is the date and time when report ReplicaController status
	ReportTime time.Time `json:"reportTime,omitempty"`
	// Status is the status of ReplicaController
	Status ReplicaControllerStatus `json:"status"`
	// LastStatus is the last status of ReplicaController
	LastStatus ReplicaControllerStatus `json:"lastStatus,omitempty"`
	// Message indicating details about why the ReplicaController is in this status
	Message string `json:"message,omitempty"`
	// Pods the index of the pod which is created by this ReplicaController
	Pods []*BcsPodIndex `json:"pods"`
	// added  20181011, add for differentiate process/application
	Kind BcsDataType `json:"kind"`
}

type PodStatus string

const (
	Pod_Staging  PodStatus = "Staging"
	Pod_Starting PodStatus = "Starting"
	Pod_Running  PodStatus = "Running"
	Pod_Error    PodStatus = "Error"
	Pod_Killing  PodStatus = "Killing"
	Pod_Killed   PodStatus = "Killed"
	Pod_Failed   PodStatus = "Failed"
	Pod_Finish   PodStatus = "Finish"
)

type BcsPodIndex struct {
	Name string `json:"name"`
}

//BcsPodStatus define pod status
type BcsPodStatus struct {
	ObjectMeta `json:"metadata"`
	// RcName the rc name who create this pod. if this pod is created by pod, RcName is empty
	RcName string `json:"rcname,omitempty"`
	// Status is the status of this pod
	Status PodStatus `json:"status,omitempty"`
	// LastStatus is the last status of this pod
	LastStatus PodStatus `json:"lastStatus,omitempty"`
	// HostIP is the ip address of the host where this pod run
	HostIP string `json:"hostIP,omitempty"`
	// HostName is the hostname where this pod run
	HostName string `json:"hostName"`
	// PodIP is the pod address of this pod
	PodIP string `json:"podIP,omitempty"`
	// Message indicating details about why the pod is in this status
	Message string `json:"message,omitempty"`
	// StartTime is date and time when this pod start
	StartTime time.Time `json:"startTime,omitempty"`
	// LastStatus is the last status of this pod
	LastUpdateTime time.Time `json:"lastUpdateTime,omitempty"`
	// ReportTime is the date and time when report pod status
	ReportTime time.Time `json:"reportTime,omitempty"`
	// KillPolicy
	KillPolicy *KillPolicy `json:"killPolicy,omitempty"`
	// RestartPolicy
	RestartPolicy *RestartPolicy `json:"restartPolicy,omitempty"`
	// ContainerStatus is the container status
	ContainerStatuses []*BcsContainerStatus `json:"containerStatuses,omitempty"`
	//bcs message
	BcsMessage string `json:"bcsMessage,omitempty"`
	// added  20181011, add for differentiate process/application
	Kind BcsDataType `json:"kind"`
}

// BcsHealthCheckStatus define health status
type BcsHealthCheckStatus struct {
	// healthcheck type
	Type BcsHealthCheckType `json:"type"`
	// the health check result, ture is ok, false is not good
	Result bool `json:"result"`
	// the health check message. e.g: if health is not good, the message record the reason
	Message string `json:"message,omitempty"`
}

type ContainerStatus string

// There are the valid statuses of container
const (
	Container_Staging  ContainerStatus = "Staging"
	Container_Starting ContainerStatus = "Starting"
	Container_Running  ContainerStatus = "Running"
	Container_Killing  ContainerStatus = "Killing"
	Container_Killed   ContainerStatus = "Killed"
	Container_Finish   ContainerStatus = "Finish"
	Container_Failed   ContainerStatus = "Failed"
	Container_Error    ContainerStatus = "Error"
)

//BcsContainerStatus define container status
type BcsContainerStatus struct {
	// container name
	Name string `json:"name"`
	// container id
	ContainerID  string `json:"containerID"`
	RestartCount int32  `json:"restartCount,omitempty"`
	// Status is the status of this container
	Status ContainerStatus `json:"status,omitempty"`
	// Status is the last status of this container``
	LastStatus ContainerStatus `json:"lastStatus,omitempty"`
	// exit code
	TerminateExitCode int `json:"exitcode,omitempty"`
	// Image is this container image
	Image string `json:"image"`
	// Message indicating details about why this container`` is in this status
	Message string `json:"message,omitempty"`
	// StartTime is date and time when this container start
	StartTime time.Time `json:"startTime,omitempty"`
	// LastStatus is the last status of Container
	LastUpdateTime time.Time `json:"lastUpdateTime,omitempty"`
	// FinishTime is the finish date and time
	FinishTime time.Time `json:"finishTime,omitempty"`
	// health check status
	HealthCheckStatus []*BcsHealthCheckStatus `json:"healCheckStatus,omitempty"`
	// ports
	Ports []ContainerPort `json:"containerPort,omitempty"`
	//command
	Command string `json:"command,omitempty"`
	// arguments
	Args []string `json:"args,omitempty"`
	//volumes
	Volumes []Volume `json:"volumes,omitempty"`
	// network
	Network string `json:"networkMode,omitempty"`
	// labels
	Labels map[string]string `json:"labels,omitempty"`
	// Resources
	Resources ResourceRequirements `json:"resources,omitempty"`
	//envs
	Env map[string]string `json:"env,omitempty"`
}

type BcsDeploymentStatus struct {
	ObjectMeta `json:"metadata"`
	Instance   int32 `json:"instance"`

	// Total number of non-terminated pods targeted by this deployment that have the desired template spec.
	UpdatedInstance int32 `json:"updatedInstance"`

	// Total number of ready pods targeted by this deployment.
	ReadyInstance int32 `json:"readyInstance"`

	// Total number of available pods (ready for at least minReadySeconds) targeted by this deployment.
	AvailableInstance int32 `json:"availableInstance"`

	// Total number of unavailable pods targeted by this deployment.
	UnavailableInstance int32 `json:"unavailableInstance"`

	Status []*BcsReplicaControllerStatus `json:"deploymentStatus"`
	// ReportTime is the date and time when report status
	ReportTime time.Time `json:"reportTime,omitempty"`
}
