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

package groupdefaultjobecho

import (
	"bk-bscp/cmd/middle-services/bscp-patcher/modules/ctm"
	"bk-bscp/pkg/common"
	"bk-bscp/pkg/logger"
)

// Job is echo job.
type Job struct {
	// Name job name.
	Name string

	// job inner sequence id.
	seq string
}

// GetName return the cron job name.
func (job *Job) GetName() string {
	return job.Name
}

// NeedRun is the func to decide if the cron job should run in this moment.
func (job *Job) NeedRun() bool {
	needRun := ctm.GetController().NeedRun(job.Name)
	if !needRun {
		return false
	}
	// need run now, reset a new sequence id.
	job.seq = common.Sequence()
	return true
}

// BeforeRun echo job before run func.
func (job *Job) BeforeRun() error {
	logger.Infof("cron job[%s|%s] execute BeforeRun success", job.Name, job.seq)
	return nil
}

// Run echo job run func.
func (job *Job) Run() error {
	// NOTE: do your job logics base on ctm controller, viper, smgr.
	// viper := ctm.GetViper()
	// smgr := ctm.GetShardingDBManager()

	logger.Infof("cron job[%s|%s] execute Run success", job.Name, job.seq)
	return nil
}

// AfterRun echo job after run func.
func (job *Job) AfterRun() error {
	logger.Infof("cron job[%s|%s] execute AfterRun success", job.Name, job.seq)
	ctm.GetController().Done(job.Name)
	return nil
}
