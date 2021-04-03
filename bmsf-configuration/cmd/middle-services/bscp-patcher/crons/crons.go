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

package crons

import (
	"bk-bscp/cmd/middle-services/bscp-patcher/crons/group-default.job-echo"
	"bk-bscp/cmd/middle-services/bscp-patcher/modules/ctm"
)

var (
	// all registered cron jobs.
	jobs = []ctm.Job{}
)

// register one crontab job.
func register(job ctm.Job) {
	jobs = append(jobs, job)
}

// CronJobs returns all registered crontab jobs.
func CronJobs() []ctm.Job {
	return jobs
}

// register crons here.
func init() {
	// group-default.job-echo
	register(&groupdefaultjobecho.Job{Name: "group-default.job-echo"})

	// TODO add your cron job pkg here.
}
