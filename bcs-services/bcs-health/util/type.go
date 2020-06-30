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

package util

import (
	"encoding/json"
	"fmt"
	"github.com/Tencent/bk-bcs/bcs-common/common/types"
)

type Server struct {
	BindAddr string `json:"bind"`
	Port     uint   `json:"port"`
	TLS      `json:"serverTLS,inline"`
}

type TLS struct {
	CaFile   string `json:"ca-file"`
	CertFile string `json:"cert-file"`
	KeyFile  string `json:"key-file"`
	PassWord string `json:"-"`
}

// a zone contains a collection of jobs that this node will do.
// a zone value could be a cluster-id or a self-defined string which is set by user.
type Zone string

const AllZones Zone = "all-bcs-health-zones"

func (z Zone) IsAllZone() bool {
	if z == AllZones {
		return true
	}
	return false
}

type Zones []Zone

func (z Zones) IsAllZone() bool {
	for _, zone := range z {
		if zone.IsAllZone() {
			return true
		}
	}
	return false
}

// health slave's meta info
type Slave struct {
	// the name of the slave cluster, must be unique among all clusters.
	// can be the value of cluster-id or others.
	SlaveClusterName string `json:"slaveClusterName"`
	// containers a collection of zone. it determines the jobs which is get from master.
	Zones Zones `json:"zones"`
	// details about this slave.
	types.ServerInfo `json:",inline"`
}

type Protocol string

const (
	HTTP Protocol = "http"
	TCP  Protocol = "tcp"
)

type ActionType string

const (
	// this job is added now
	AddAction ActionType = "add"
	// this job is updated.
	UpdateAction ActionType = "update"
	// this job is deleted.
	DeleteAction ActionType = "delete"
	// this job's action(above) is already handled.
	HandledAction ActionType = "handled"
)

// job contains all the info which is needed during the check.
type Job struct {
	// which module this job belongs to.
	Module string `json:"module"`
	// job actions, include: add, update, delete, handled.
	Action ActionType `json:"action"`
	// zone that this job belongs to.
	Zone Zone `json:"zone"`
	// Protocol that this job will use.
	Protocol Protocol `json:"protocol"`
	// url of the checked point, ip:port
	Url string `json:"url"`
	// the result of this job.
	Status *JobStatus `json:"status,omitempty"`
}

func (j *Job) Name() string {
	return fmt.Sprintf("%s::%s::%s", j.Zone, j.Protocol, j.Url)
}

func (j *Job) String() string {
	js, _ := json.Marshal(j)
	return string(js)
}

func (j *Job) DeepCopy() *Job {
	job := &Job{
		Module:   j.Module,
		Action:   j.Action,
		Zone:     j.Zone,
		Protocol: j.Protocol,
		Url:      j.Url,
	}

	if j.Status != nil {
		job.Status = &JobStatus{
			SlaveInfo:  j.Status.SlaveInfo,
			Success:    j.Status.Success,
			Message:    j.Status.Message,
			FinishedAt: j.Status.FinishedAt,
		}
	}

	return job
}

type JobStatus struct {
	// slave infos that do this job
	SlaveInfo *Slave `json:"slaveInfo,omitempty"`
	// where the job is success or not.
	Success bool `json:"success,omitempty"`
	// record the check result when failed.
	Message string `json:"message,omitempty"`
	// time of the job is done.
	FinishedAt int64 `json:"finishedAt,omitempty"`
}

type SvrResponse struct {
	Error error  `json:"error"`
	Jobs  []*Job `json:"jobs,omitempty"`
}
