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

package collector

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"bk-bcs/bcs-common/common/blog"
	"bk-bcs/bcs-services/bcs-health/pkg/client"
	"bk-bcs/bcs-services/bcs-health/pkg/role"
	"bk-bcs/bcs-services/bcs-health/util"

	"bk-bcs/bcs-common/common/statistic"
	"golang.org/x/net/context"
)

type CollectResult interface {
	Collect(*util.Job, error)
}

func NewJobCollector(slave *util.Slave, cli client.ClientInterface, role role.RoleInterface) *JobCollector {
	jc := &JobCollector{
		slave: slave,
		taskPool: &taskPool{
			tasks: make(map[string]*task),
		},
		role: role,
		cli:  cli,
	}

	return jc
}

type JobCollector struct {
	slave    *util.Slave
	taskPool *taskPool
	cli      client.ClientInterface
	role     role.RoleInterface
}

func (j *JobCollector) Run() {
	go j.watch()
	go j.startSync()
}

func (j *JobCollector) Collect(job *util.Job, err error) {
	job.Action = util.HandledAction
	job.Status = &util.JobStatus{
		SlaveInfo:  j.slave,
		Success:    true,
		Message:    "success",
		FinishedAt: time.Now().Unix(),
	}
	if err != nil {
		job.Status.Success = false
		job.Status.Message = err.Error()
	}

	if err := j.cli.ReportJobs(job); err != nil {
		blog.Errorf("report job[%s] status failed. err: %v", job.Name(), err)
		return
	}
	blog.V(4).Infof("report job[%s] status[%t] success.", job.Name(), job.Status.Success)
	return
}

func (j *JobCollector) watch() {
	jobChan := make(chan *util.Job)
	errChan := make(chan error)

	watchJobID := "watch_job_id"
retry:
	blog.Infof("---> starting watch jobs from health master.")

	// first, do sync.
	j.doSync()

	closeChan := make(chan struct{})
	if err := j.cli.WatchJobs(j.slave, jobChan, errChan); nil != err {
		statistic.Set(watchJobID, fmt.Errorf("watch jobs failed. err: %v", err))
		blog.Errorf("watch jobs failed. err: %v", err)
		time.Sleep(1 * time.Second)
		goto retry
	}
	statistic.Reset(watchJobID)

	go func() {
		for {
			select {
			case <-closeChan:
				blog.Errorf("received signal to stop watch jobs from master health.")
				return
			case job := <-jobChan:
				switch job.Action {
				case util.AddAction:
					j.AddTask(job)

				case util.UpdateAction:
					j.UpdateTask(job)

				case util.DeleteAction:
					j.DeleteTask(job)

				default:
					j.UpdateTask(job)
				}
			}
		}
	}()

	select {
	case err := <-errChan:
		blog.Errorf("watch jobs from master health, stop now, because of error occurred, err: %v", err)
		close(closeChan)
		blog.Warnf("*** try to re-watch master again. ***")
		goto retry
	}
}

func (j *JobCollector) startSync() {
	blog.Infof("start sync jobs from master.")
	j.doSync()
	ticker := time.Tick(30 * time.Second)
	for {
		select {
		case <-ticker:
		}
		j.doSync()
	}
}

func (j *JobCollector) doSync() {
	syncJobID := "sync_job_id"
	newJobs, err := j.cli.ListJobs(j.slave)
	if err != nil {
		statistic.Set(syncJobID, fmt.Errorf("sync jobs from master failed. err:%v", err))
		blog.Errorf("sync jobs from master failed. err:%v", err)
		return
	}
	statistic.Reset(syncJobID)

	newJobsMapper := make(map[string]*util.Job)
	for _, job := range newJobs {
		if len(job.Zone) == 0 || len(job.Url) == 0 || len(job.Protocol) == 0 {
			blog.Errorf("watch jobs, but got a invalid job: [%s]", job.Name())
			continue
		}
		newJobsMapper[getKey(job)] = job
	}

	runningJobsmapper := make(map[string]*util.Job)
	for _, job := range j.taskPool.ListJobs() {
		runningJobsmapper[getKey(job)] = job
	}

	for key, job := range newJobsMapper {
		if _, exist := runningJobsmapper[key]; !exist {
			blog.Warnf("sync jobs, find a *new* job: %s", job.Name())
			j.AddTask(job)
		}
	}

	for key, job := range runningJobsmapper {
		if _, exist := newJobsMapper[key]; !exist {
			blog.Warnf("sync jobs, find a *redundant* job: %s", job.Name())
			j.DeleteTask(job)
		}
	}

	blog.V(5).Info("sync jobs from master finished.")
}

type task struct {
	ctx        context.Context
	cancelFunc context.CancelFunc
	job        *util.Job
}

type taskPool struct {
	locker sync.Mutex
	tasks  map[string]*task
}

func (t *taskPool) ListJobs() []*util.Job {
	t.locker.Lock()
	defer t.locker.Unlock()
	jobs := make([]*util.Job, 0)
	for _, task := range t.tasks {
		jobs = append(jobs, task.job.DeepCopy())
	}
	return jobs
}

func (j *JobCollector) AddTask(job *util.Job) {
	blog.Infof("add *new* job task: %s", job.Name())
	j.taskPool.locker.Lock()
	defer j.taskPool.locker.Unlock()

	key := getKey(job)
	if _, exist := j.taskPool.tasks[key]; exist {
		blog.Infof("add *new* job task: %s, but already exist.", job.Name())
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	task := &task{
		ctx:        ctx,
		cancelFunc: cancel,
		job:        job,
	}

	j.taskPool.tasks[key] = task
	go j.doCheck(task)
}

func (j *JobCollector) DeleteTask(job *util.Job) {
	blog.Infof("prepare to *delete* task: %s", job.Name())
	j.taskPool.locker.Lock()
	defer j.taskPool.locker.Unlock()
	key := getKey(job)
	task, exist := j.taskPool.tasks[key]
	if !exist {
		blog.Infof("*delete* task: %s, but not exist.", job.Name())
		return
	}
	task.cancelFunc()
	delete(j.taskPool.tasks, key)
	blog.Infof("*delete* task[%s] success.", job.Name())
}

func (j *JobCollector) UpdateTask(job *util.Job) {
	blog.Infof("*update* task, current: %s", job.Name())
	j.taskPool.locker.Lock()
	defer j.taskPool.locker.Unlock()
	key := getKey(job)
	oldTask, exist := j.taskPool.tasks[key]
	if !exist {
		ctx, cancel := context.WithCancel(context.Background())
		newTask := &task{
			ctx:        ctx,
			cancelFunc: cancel,
			job:        job,
		}
		go j.doCheck(newTask)
		return
	}
	// delete the old task.
	oldTask.cancelFunc()
	delete(j.taskPool.tasks, key)

	// create a new task
	ctx, cancel := context.WithCancel(context.Background())
	newTask := &task{
		ctx:        ctx,
		cancelFunc: cancel,
		job:        job,
	}
	j.taskPool.tasks[key] = newTask
	go j.doCheck(newTask)
}

const check_delay_seconds = 3 * time.Second

func (j *JobCollector) doCheck(task *task) {

	for {
		select {
		case <-task.ctx.Done():
			blog.Infof("stop check job: %s", task.job.Name())
			return
		default:
		}
		if !j.role.IsMaster() {
			time.Sleep(3 * time.Second)
			continue
		}
		blog.V(4).Infof("do job：%s", task.job.Name())
		var err error
		switch task.job.Protocol {
		case util.TCP:
			err = j.checkTCP(task)
		case util.HTTP:
			err = j.checkHttp(task)
		default:
			blog.Errorf("unsupported job protocol: %s.", task.job.Protocol)
			return
		}
		job := task.job.DeepCopy()
		j.Collect(job, err)
		time.Sleep(check_delay_seconds)
	}
}

func (j *JobCollector) checkTCP(task *task) error {
	url := task.job.Url
	doCheck := func() error {
		conn, err := net.DialTimeout("tcp", url, 3*time.Second)
		if nil != err {
			return err
		}
		defer conn.Close()
		return nil
	}

	err := doCheck()
	if nil == err {
		return nil
	}
	blog.Errorf("health alarm check addr:%s failed for first time. err: %s", url, err)

	// do first back off with 3 seconds.
	time.Sleep(3 * time.Second)
	err = doCheck()
	if nil == err {
		return nil
	}
	blog.Errorf("health alarm check addr:%s failed with 1st backoff(3s). err: %s", url, err)

	// do first back off with 5 seconds.
	time.Sleep(5 * time.Second)
	err = doCheck()
	if nil == err {
		return nil
	}
	blog.Errorf("health alarm check addr:%s failed with 2nd backoff(5s). err: %s", url, err)
	return err
}

func (j *JobCollector) checkHttp(task *task) error {
	url := task.job.Url
	if !strings.HasPrefix(url, "http://") {
		url = fmt.Sprintf("http://%s", url)
	}
	cli := http.DefaultClient
	cli.Timeout = time.Duration(3 * time.Second)
	_, err := cli.Get(url)
	return err
}

func getKey(j *util.Job) string {
	if j == nil {
		return ""
	}
	return fmt.Sprintf("%s-%s-%s", j.Zone, j.Protocol, j.Url)
}
