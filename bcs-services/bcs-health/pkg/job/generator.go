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

package job

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-health/pkg/zk"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-health/util"

	"golang.org/x/net/context"
)

type JobInterf interface {
	WatchSlaveJobs(ctx context.Context, slave *util.Slave, writeCh chan *util.Job)
	ListSlaveJobs(zones util.Zones) []*util.Job
}

func NewJobController(zkAddrs []string) (JobInterf, error) {
	jc := new(JobController)
	jc.jobPool = newJobPool()
	baseBranch := fmt.Sprintf("%s/%s", types.BCS_SERV_BASEPATH, types.BCS_MODULE_LOADBALANCE)
	z, err := zk.NewZkWatcher(zkAddrs, baseBranch, jc.jobPool)
	if err != nil {
		return nil, err
	}
	if err := z.Run(); err != nil {
		return nil, err
	}
	jc.zkWatcher = z
	jc.jobPool.SyncEvents()
	return jc, nil
}

type JobController struct {
	zkWatcher *zk.ZkWatcher
	jobPool   *jobPool
}

func (j *JobController) JobExist(job *util.Job) bool {
	return j.jobPool.JobExist(job)
}

func (j *JobController) WatchSlaveJobs(ctx context.Context, slave *util.Slave, writeCh chan *util.Job) {
	name := fmt.Sprintf("%s-%d", slave.IP, time.Now().Unix())
	j.jobPool.AddSubscribers(slave.Zones, name, writeCh)

	go func() {
		select {
		case <-ctx.Done():
			j.jobPool.DeleteSubscribers(slave.Zones, name)
			if ctx.Err() != nil {
				blog.Errorf("slave[ClusterName:%s, IP:%s] watch jobs failed. err: %v",
					slave.SlaveClusterName, slave.IP, ctx.Err())
				return
			}
			blog.Infof("slave[ClusterName:%s, IP:%s] stop watch jobs.", slave.SlaveClusterName, slave.IP)
		}
	}()
	return
}

func (j *JobController) ListSlaveJobs(zones util.Zones) []*util.Job {
	return j.jobPool.GetZoneJobs(zones)
}

const default_event_chan_length = 500

func newJobPool() *jobPool {
	return &jobPool{
		subscribers: make(map[util.Zone]map[string]chan *util.Job),
		eventChan:   make(chan *util.Job, default_event_chan_length),
		pool:        make(map[util.Zone]map[string]*util.Job),
	}
}

type jobPool struct {
	locker      sync.Mutex
	subscribers map[util.Zone]map[string]chan *util.Job
	eventChan   chan *util.Job
	// format: map[zone]map[job.Protocol:job.Url]Job
	pool map[util.Zone]map[string]*util.Job
}

func (j *jobPool) OnAddLeaf(branch, leaf, value string) {
	job, err := formatJob(value, util.AddAction)
	if err != nil {
		blog.Errorf("received zk add event, handle failed, branch: %s, leaf: %s, value: %s, err: %v",
			branch, leaf, value, err)
		return
	}
	j.addJob(job)
	blog.Warnf("*add* new health job, branch: %s, leaf: %s, value: %v", branch, leaf, value)
}

func (j *jobPool) OnUpdateLeaf(branch, leaf, oldvalue, newvalue string) {
	job, err := formatJob(newvalue, util.UpdateAction)
	if err != nil {
		blog.Errorf("received zk update event, handle failed, branch: %s, leaf: %s, value: %s, err: %v",
			branch, leaf, newvalue, err)
		return
	}
	j.updateJob(job)
	blog.Warnf("*update* health job, branch: %s, leaf: %s, value: %v", branch, leaf, newvalue)
}

func (j *jobPool) OnDeleteLeaf(branch, leaf, value string) {
	job, err := formatJob(value, util.DeleteAction)
	if err != nil {
		blog.Errorf("received zk delete event, handle failed, branch: %s, leaf: %s, value: %s, err: %v",
			branch, leaf, value, err)
		return
	}
	j.deleteJob(job)
	blog.Warnf("*delete* health job, branch: %s, leaf: %s, value: %v", branch, leaf, value)
}

func (j *jobPool) AddSubscribers(zones util.Zones, name string, writeCh chan *util.Job) {
	j.locker.Lock()
	defer j.locker.Unlock()

	if zones.IsAllZone() {
		if _, exist := j.subscribers[util.AllZones]; !exist {
			j.subscribers[util.AllZones] = make(map[string]chan *util.Job)
		}
		j.subscribers[util.AllZones][name] = writeCh
	} else {
		for _, zone := range zones {
			if _, exist := j.subscribers[zone]; !exist {
				j.subscribers[zone] = make(map[string]chan *util.Job)
			}
			j.subscribers[zone][name] = writeCh
		}
	}
}

func (j *jobPool) DeleteSubscribers(zones util.Zones, name string) {
	j.locker.Lock()
	defer j.locker.Unlock()

	if zones.IsAllZone() {
		delete(j.subscribers[util.AllZones], name)
		return
	}
	for _, z := range zones {
		delete(j.subscribers[z], name)
	}
	return
}

func (j *jobPool) SyncEvents() {
	blog.Infof("start sync events to health slaves by watch.")
	go func() {
		for job := range j.eventChan {
			do := func() {
				j.locker.Lock()
				defer j.locker.Unlock()

				for zone, users := range j.subscribers {
					if zone.IsAllZone() || job.Zone == zone {
						for name, ch := range users {
							ch <- job
							blog.V(4).Infof("broadcast subscribers[%s] job event, action:%s, zone:%s, url: %s",
								name, job.Action, job.Zone, string(job.Protocol)+":"+job.Url)
						}
						continue
					}
				}
			}

			do()
		}
	}()
}

func (j *jobPool) JobExist(job *util.Job) bool {
	if _, exist := j.pool[job.Zone]; !exist {
		return false
	}
	key := jobKey(job)
	_, exist := j.pool[job.Zone][key]
	return exist
}

func (j *jobPool) addJob(job *util.Job) {
	j.locker.Lock()
	defer j.locker.Unlock()
	key := jobKey(job)
	if _, exist := j.pool[job.Zone]; !exist {
		j.pool[job.Zone] = make(map[string]*util.Job)
	}
	j.pool[job.Zone][key] = job
	j.launchEvent(job)
}

func (j *jobPool) updateJob(job *util.Job) {
	j.addJob(job)
}

func (j *jobPool) deleteJob(job *util.Job) {
	j.locker.Lock()
	defer j.locker.Unlock()
	key := jobKey(job)
	delete(j.pool[job.Zone], key)
	j.launchEvent(job)
}

func (j *jobPool) launchEvent(job *util.Job) {
	blog.V(5).Infof("job pool launch a event, job: %s", job.String())
	if len(j.eventChan) == default_event_chan_length {
		<-j.eventChan
		blog.Warnf("event chan is overflow, drop the oldest one. job:%s", job.String())
	}
	j.eventChan <- job
}

func (j *jobPool) GetZoneJobs(zones util.Zones) []*util.Job {
	j.locker.Lock()
	defer j.locker.Unlock()
	jobs := make([]*util.Job, 0)

	if zones.IsAllZone() {
		for _, js := range j.pool {
			for _, j := range js {
				jobs = append(jobs, j)
			}
		}
		return jobs
	}

	for _, z := range zones {
		for zone, js := range j.pool {
			if z == zone {
				for _, j := range js {
					jobs = append(jobs, j)
				}
			}
		}
	}
	return jobs
}

func jobKey(job *util.Job) string {
	if job == nil {
		return ""
	}
	return fmt.Sprintf("%s:%s", job.Protocol, job.Url)
}

func formatJob(value string, act util.ActionType) (*util.Job, error) {
	s := types.ServerInfo{}
	if err := json.Unmarshal([]byte(value), &s); nil != err {
		return nil, err
	}
	url := fmt.Sprintf("%s:%d", s.IP, s.Port)
	return &util.Job{
		Module:   "Loadbalance",
		Action:   act,
		Zone:     util.Zone(s.Cluster),
		Protocol: util.TCP,
		Url:      url,
	}, nil
}
