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
 */

// Package manager xxx
package manager

import (
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/encrypt"
	"github.com/Tencent/bk-bcs/bcs-common/common/task"
	bcsmongo "github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers/mongo"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/common/constant"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/logging"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/provider/common/callback"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/provider/common/steps"
	isteps "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/provider/internal/steps"
)

// globalTaskServer global task server
var globalTaskServer *task.TaskManager

// GetTaskServer get task server
func GetTaskServer() *task.TaskManager {
	return globalTaskServer
}

// RunTaskManager run task manager
func RunTaskManager() (*task.TaskManager, error) {
	btm := task.NewTaskManager()

	// 判断 password 是否加密，如果加密需要解密获取到原始数据
	password := config.GlobalConf.Mongo.Password
	if password != "" && config.GlobalConf.Mongo.Encrypted {
		realPwd, _ := encrypt.DesDecryptFromBase([]byte(password))
		password = string(realPwd)
	}

	taskCfg := &task.ManagerConfig{
		ModuleName: constant.ModuleName,
		WorkerNum:  config.GlobalConf.TaskConfig.WorkerCnt,
		Broker: &task.BrokerConfig{
			QueueAddress: config.GlobalConf.TaskConfig.QueueAddress,
			Exchange:     config.GlobalConf.TaskConfig.Exchange,
		},
		Backend: &bcsmongo.Options{
			Hosts:                 strings.Split(config.GlobalConf.Mongo.Address, ","),
			Replicaset:            config.GlobalConf.Mongo.Replicaset,
			ConnectTimeoutSeconds: int(config.GlobalConf.Mongo.ConnectTimeout),
			AuthDatabase:          config.GlobalConf.Mongo.AuthDatabase,
			Database:              config.GlobalConf.Mongo.Database,
			Username:              config.GlobalConf.Mongo.Username,
			Password:              password,
			MaxPoolSize:           uint64(config.GlobalConf.Mongo.MaxPoolSize),
			MinPoolSize:           uint64(config.GlobalConf.Mongo.MinPoolSize),
		},
	}
	// register step worker && callback
	taskCfg.StepWorkers = registerSteps()
	taskCfg.CallBacks = registerCallbacks()

	// init task manager
	err := btm.Init(taskCfg)
	if err != nil {
		logging.Error("init task manager failed, %s", err.Error())
		return nil, err
	}

	// run task manager
	btm.Run()

	globalTaskServer = btm

	return btm, nil
}

// registerSteps register all steps
func registerSteps() []task.StepWorkerInterface {
	stepList := make([]task.StepWorkerInterface, 0)

	// common steps
	stepList = append(stepList, steps.NewHelloStep())
	stepList = append(stepList, steps.NewSumStep())

	stepList = append(stepList, steps.NewItsmApproveStep())
	stepList = append(stepList, steps.NewWaitStep())
	stepList = append(stepList, steps.NewItsmSubmitStep())

	stepList = append(stepList, isteps.NewFederationQuotaStep())

	return stepList
}

// registerCallbacks register all callbacks
func registerCallbacks() []task.CallbackInterface {
	callbacks := make([]task.CallbackInterface, 0)

	callbacks = append(callbacks, callback.NewTestCallback())
	callbacks = append(callbacks, callback.NewQuotaCallback())

	return callbacks
}
