/*
Tencent is pleased to support the open source community by making Blueking Container Service available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package cron

import (
	"errors"
	"log"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/spf13/viper"

	"bk-bscp/pkg/common"
)

const (
	// defaultScheduleInterval is default schedule interval.
	defaultScheduleInterval = time.Second
)

// Job is an interface for submitted cron jobs.
type Job interface {
	// NeedRun is the func to decide if the job should run in this moment.
	NeedRun() bool

	// BeforeRun is func executed before Run func.
	BeforeRun() error

	// Run is the main func of one job.
	Run() error

	// AfterRun is func executed after Run func executed success.
	AfterRun() error
}

// runner is job runner with job name and job cron entry id.
type runner struct {
	name string
	id   cron.EntryID
	spec string
	job  Job
}

// Run runs target job with BeforeRun/Run/AfterRun
// different hooks running in order.
func (r *runner) Run() {
	if r.job == nil {
		return
	}

	if r.job.NeedRun() {
		// BeforeRun/Run/AfterRun funcs are single hooks, no mater the
		// pre result the next func must run.
		r.job.BeforeRun()
		r.job.Run()
		r.job.AfterRun()
	}
}

// Cron keeps track of any number of entries, invoking the associated func as
// specified by the schedule. It may be started, stopped, and the entries may
// be inspected while running. One Cron is one job runner, you can add a job to It
// with specs which accepts this spec: https://en.wikipedia.org/wiki/Cron
type Cron struct {
	viper *viper.Viper
	cron  *cron.Cron

	// cron name.
	name string

	// running flag.
	once      sync.Once
	isRunning bool
	isStop    bool

	// registered jobs, job unique name -> runner.
	registeredJobs   map[string]*runner
	registeredJobsMu sync.RWMutex

	// enabled jobs, job unique name -> runner.
	enabledJobs   map[string]*runner
	enabledJobsMu sync.RWMutex
}

// New creates and returns a new Cron object base on the options.
func New(name, cronFile string) (*Cron, error) {
	viper := viper.New()
	viper.SetConfigFile(cronFile)

	fd, _ := common.CreateFile(cronFile)
	fd.Close()

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}
	viper.WatchConfig()

	if len(name) == 0 {
		name = "default"
	}

	return &Cron{
		viper:          viper,
		cron:           cron.New(),
		name:           name,
		registeredJobs: make(map[string]*runner, 0),
		enabledJobs:    make(map[string]*runner, 0),
	}, nil
}

// RegisterJob registers a new Job with target name.
func (c *Cron) RegisterJob(name string, job Job) error {
	if len(name) == 0 {
		return errors.New("empty job name")
	}
	if job == nil {
		return errors.New("invalid job struct")
	}

	c.registeredJobsMu.Lock()
	defer c.registeredJobsMu.Unlock()

	if _, exists := c.registeredJobs[name]; exists {
		return errors.New("job with the target name is already exists")
	}
	c.registeredJobs[name] = &runner{name: name, job: job}

	return nil
}

// UnregisterJob unregisters a job with target name.
func (c *Cron) UnregisterJob(name string) {
	c.registeredJobsMu.Lock()
	defer c.registeredJobsMu.Unlock()
	delete(c.registeredJobs, name)
}

// GetJobSpec returns the job spec of target job name.
func (c *Cron) GetJobSpec(name string) string {
	return c.viper.GetString(name)
}

func (c *Cron) schedule() {
	// stop and delete jobs which not registered or registered without spec.
	c.enabledJobsMu.Lock()
	for name, runner := range c.enabledJobs {
		c.registeredJobsMu.RLock()
		_, exists := c.registeredJobs[name]
		c.registeredJobsMu.RUnlock()

		if !exists {
			// stop and delete not registered job.
			log.Printf("cron[%s] stop and delete not registered job[%s]", c.name, name)
			c.cron.Remove(runner.id)
			delete(c.enabledJobs, name)
			continue
		}

		jobSpec := c.GetJobSpec(name)
		if len(jobSpec) == 0 {
			// stop and delete empty spec job.
			log.Printf("cron[%s] stop and delete registered empty spec job[%s]", c.name, name)
			c.cron.Remove(runner.id)
			delete(c.enabledJobs, name)
			continue
		}

		// job spec maybe changed.
		if jobSpec != runner.spec {
			// restart the runner.
			log.Printf("cron[%s] spec changed old[%s] new[%s], restart job[%s]", c.name, runner.spec, jobSpec, name)
			c.cron.Remove(runner.id)

			id, err := c.cron.AddJob(jobSpec, runner)
			if err != nil {
				// restart job failed.
				log.Printf("cron[%s] restart job[%s], failed: %+v", c.name, name, err)
				continue
			}
			runner.id = id
			runner.spec = jobSpec
			log.Printf("cron[%s] restart job[%s] success", c.name, name)
		}
	}
	c.enabledJobsMu.Unlock()

	// start registered jobs which could find valid cron spec.
	c.registeredJobsMu.RLock()
	defer c.registeredJobsMu.RUnlock()

	for name, runner := range c.registeredJobs {
		// job spec accepts this spec: https://en.wikipedia.org/wiki/Cron,
		// should add the cron spec in viper.
		jobSpec := c.GetJobSpec(name)
		if len(jobSpec) == 0 {
			// empty job spec, not need to run this job.
			continue
		}

		c.enabledJobsMu.Lock()
		if _, exists := c.enabledJobs[name]; exists {
			// already exists.
			c.enabledJobsMu.Unlock()
			continue
		}

		id, err := c.cron.AddJob(jobSpec, runner)
		if err != nil {
			// start job failed, try next round.
			log.Printf("cron[%s] start new job[%s], failed: %+v", c.name, name, err)
			c.enabledJobsMu.Unlock()
			continue
		}
		log.Printf("cron[%s] start new job[%s] success", c.name, name)

		runner.id = id
		runner.spec = jobSpec
		c.enabledJobs[name] = runner
		c.enabledJobsMu.Unlock()
	}
}

// Start runs the cron scheduler in its own goroutine, or no-op if already started.
func (c *Cron) Start() {
	if c.isStop {
		return
	}
	c.cron.Start()
	c.isRunning = true
	log.Printf("cron[%s] running now", c.name)

	// job schedule func, only run once for one cron.
	scheduleFunc := func() {
		for {
			<-time.After(defaultScheduleInterval)

			if c.isStop {
				log.Printf("cron[%s] stop now", c.name)
				return
			}

			if c.isRunning {
				c.schedule()
			}
		}
	}
	go c.once.Do(scheduleFunc)
}

// Pause pauses the cron scheduler if it is running; otherwise it does nothing.
func (c *Cron) Pause() {
	c.isRunning = false
	<-c.cron.Stop().Done()
}

// Stop stops the cron scheduler if it is running; otherwise it does nothing.
// If the Cron is stoped, it can not start again.
func (c *Cron) Stop() {
	c.isRunning = false
	c.isStop = true
	<-c.cron.Stop().Done()
}
