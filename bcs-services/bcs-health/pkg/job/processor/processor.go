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

package processor

import (
	"encoding/json"
	"time"

	"bk-bcs/bcs-common/common/blog"
	"bk-bcs/bcs-services/bcs-health/pkg/role"
	"bk-bcs/bcs-services/bcs-health/util"

	"bk-bcs/bcs-services/bcs-health/pkg/alarm/utils"
	etcdc "github.com/coreos/etcd/client"
)

const job_expire_seconds = 60
const alarm_convergence_seconds = 30 * 60

type JobProcessor interface {
	WriteJob(j *util.Job) error
}

func NewJobProcessor(localIP string, rootPath string, etcdCli etcdc.KeysAPI, alarmSet utils.AlarmFactory, role role.RoleInterface) (JobProcessor, error) {
	jdb, err := NewJobDb(rootPath, etcdCli)
	if err != nil {
		return nil, err
	}
	p := new(Processor)
	p.localIP = localIP
	p.jobDB = jdb
	p.policyC = NewPolicyCenter()
	p.sendTo = alarmSet
	p.role = role
	go p.syncAlarms()
	return p, nil
}

type Processor struct {
	localIP string
	jobDB   JobDBInterf
	policyC PolicyCenterInterf
	sendTo  utils.AlarmFactory
	role    role.RoleInterface
}

func (p *Processor) WriteJob(j *util.Job) error {
	return p.jobDB.WriteJob(j)
}

func (p *Processor) syncAlarms() {
	ticker := time.Tick(1 * time.Second)
	//ticker := time.Tick(30 * time.Second)
	for {
		select {
		case <-ticker:
			blog.V(5).Infof("start to sync alarms.")
		}
		if !p.role.IsMaster() {
			continue
		}

		jobs := p.jobDB.ListJobs()
		alarms := p.policyC.GetAlarm(jobs)
		for _, a := range alarms {
			if err := p.sendTo.SendAlarm(a, p.localIP); err != nil {
				js, _ := json.Marshal(a)
				blog.Errorf("send alarm failed. alarm: %s, err: %v", string(js), err)
				continue
			}
		}
	}
}
