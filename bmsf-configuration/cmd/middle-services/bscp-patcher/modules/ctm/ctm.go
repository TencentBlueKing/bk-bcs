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

package ctm

import (
	"fmt"
	"sync"

	"github.com/spf13/viper"

	"bk-bscp/internal/dbsharding"
	"bk-bscp/pkg/cron"
	"bk-bscp/pkg/grpclb"
	"bk-bscp/pkg/logger"
)

// Job is the Interface which contains the methods that the cron job needs to implement.
type Job interface {
	// GetName return the cron job name.
	GetName() string

	// NeedRun is the func to decide if the cron job should run in this moment.
	NeedRun() bool

	// BeforeRun is func executed before Run func.
	BeforeRun() error

	// Run is the main func of one job.
	Run() error

	// AfterRun is func executed after Run func executed success.
	AfterRun() error
}

var defaultController *Controller

// Controller handles crontab job reentry.
type Controller struct {
	service *grpclb.Service

	// reentry cached crontab jobs running state for reentry.
	// crontab job name -> isRunning.
	reentry map[string]bool
	// used for manager running state mutexes of target jobs.
	reentryMu sync.RWMutex
}

// NeedRun returns if the target job could run in this moment.
func (c *Controller) NeedRun(jobName string) bool {
	isMaster, err := c.service.IsMaster()
	if err != nil || !isMaster {
		logger.V(4).Infof("cron[default] job[%s] check controller state, master[%+v], %+v", jobName, isMaster, err)
		return false
	}

	c.reentryMu.Lock()
	defer c.reentryMu.Unlock()

	isRunning, exists := c.reentry[jobName]
	if !exists || !isRunning {
		// mark running now.
		c.reentry[jobName] = true
		return true
	}
	return false
}

// Done marks running done state of target job.
func (c *Controller) Done(jobName string) {
	c.reentryMu.Lock()
	defer c.reentryMu.Unlock()

	// mark not running now.
	c.reentry[jobName] = false
}

var defaultCTM *CrontabManager

// CrontabManager is crontab jobs manager.
type CrontabManager struct {
	viper   *viper.Viper
	smgr    *dbsharding.ShardingManager
	service *grpclb.Service

	// crontab.
	cron *cron.Cron

	// controller.
	controller *Controller
}

// NewCrontabManager creates a crontab jobs manager instance.
func NewCrontabManager(viper *viper.Viper, smgr *dbsharding.ShardingManager, service *grpclb.Service) *CrontabManager {
	ctm := &CrontabManager{
		viper:   viper,
		smgr:    smgr,
		service: service,
	}

	controller := &Controller{
		service: service,
		reentry: make(map[string]bool, 0),
	}
	ctm.controller = controller

	// setup default crontab job manager and controller.
	defaultController = controller
	defaultCTM = ctm

	return ctm
}

// GetViper returns the viper.
func GetViper() *viper.Viper {
	return defaultCTM.GetViper()
}

// GetViper returns the viper of crontab manager.
func (ctm *CrontabManager) GetViper() *viper.Viper {
	if ctm == nil {
		return nil
	}
	return ctm.viper
}

// GetShardingDBManager returns the sharding database manager.
func GetShardingDBManager() *dbsharding.ShardingManager {
	return defaultCTM.GetShardingDBManager()
}

// GetShardingDBManager returns the sharding database manager of crontab manager.
func (ctm *CrontabManager) GetShardingDBManager() *dbsharding.ShardingManager {
	if ctm == nil {
		return nil
	}
	return ctm.smgr
}

// GetController returns the crontab job controller.
func GetController() *Controller {
	return defaultCTM.GetController()
}

// GetController returns the crontab job controller of crontab manager.
func (ctm *CrontabManager) GetController() *Controller {
	if ctm == nil {
		return nil
	}
	return ctm.controller
}

// Start starts the crontab job manager.
func (ctm *CrontabManager) Start(jobs []Job) error {
	// create crontab.
	cron, err := cron.New("", ctm.viper.GetString("server.cronFile"))
	if err != nil {
		return fmt.Errorf("create crontab failed, %+v", err)
	}
	ctm.cron = cron
	ctm.cron.Start()

	for _, job := range jobs {
		jobName := job.GetName()
		if err := ctm.cron.RegisterJob(jobName, job); err != nil {
			return fmt.Errorf("register job %s failed, %+v", jobName, err)
		}
		logger.Infof("cron[default] register job[%s] success", jobName)
	}

	return nil
}

// Stop stops the crontab manager.
func (ctm *CrontabManager) Stop() {
	if ctm.cron != nil {
		ctm.cron.Stop()
	}
}
