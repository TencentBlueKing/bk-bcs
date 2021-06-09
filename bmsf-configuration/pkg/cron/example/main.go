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

package main

import (
	"log"
	"time"

	"bk-bscp/pkg/cron"
)

// build your own job object which has BeforeRun/Run/AfterRun funcs
// as a cron.Job interface.
type job struct {
	name string
}

// NeedRun returns if the job need to run in this moment.
func (j *job) NeedRun() bool {
	return true
}

// BeforeRun do some pre check logics.
func (j *job) BeforeRun() error {
	log.Printf("==== Job[%s] BeforeRun ====", j.name)
	return nil
}

// Run do the main logics.
func (j *job) Run() error {
	log.Printf("==== Job[%s] Run ====", j.name)
	return nil
}

// AfterRun do the finishing logics.
func (j *job) AfterRun() error {
	log.Printf("==== Job[%s] AfterRun ====", j.name)
	return nil
}

func main() {
	// your cron name that just for memorandum, if empty the 'default'
	// would be used.
	cronName := "mycron"

	// your cron config file, write your cron job specs, and you can make
	// them as different groups. Support json/yaml and normal formats.
	cronFile := "cron.yaml"

	// create your cron, one cron object could handle multi jobs.
	// Not need to create multi crons in your single server instance.
	cron, err := cron.New(cronName, cronFile)
	if err != nil {
		panic(err)
	}
	// start cron.
	cron.Start()

	// register a new job.
	job := &job{name: "group-data-check.job-check"}
	if err := cron.RegisterJob(job.name, job); err != nil {
		panic(err)
	}

	// debug, see the job running log.
	time.Sleep(5 * time.Second)

	// unregister the job.
	cron.UnregisterJob(job.name)

	// debug, see the job stop log.
	time.Sleep(5 * time.Second)

	// re-register the job, run again.
	if err := cron.RegisterJob(job.name, job); err != nil {
		panic(err)
	}

	// debug, see the job re-running log.
	time.Sleep(5 * time.Second)

	// pause the cron, all the jobs would stop.
	// Delet the cron.yaml spec if you want to stop just one target
	// job, or UnregisterJob it.
	cron.Pause()

	// debug, see cron jobs stop log.
	time.Sleep(5 * time.Second)

	// restart cron, all jobs would re-running.
	cron.Start()

	// debug, see the jobs re-running log.
	time.Sleep(5 * time.Second)

	// stop cron, and could not restart it again.
	cron.Stop()

	// debug, see the stop log.
	time.Sleep(5 * time.Second)

	// try restart the cron, and see the log, all the jobs could not re-running again.
	cron.Start()

	select {}
}
