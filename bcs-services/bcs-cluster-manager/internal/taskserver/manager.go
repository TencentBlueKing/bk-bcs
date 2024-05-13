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

// Package taskserver xxx
package taskserver

import (
	"context"
	"fmt"
	"sync"
	"time"

	machinery "github.com/RichardKnop/machinery/v2"
	"github.com/RichardKnop/machinery/v2/backends/mongo"
	"github.com/RichardKnop/machinery/v2/brokers/amqp"
	"github.com/RichardKnop/machinery/v2/config"
	"github.com/RichardKnop/machinery/v2/locks/eager"
	"github.com/RichardKnop/machinery/v2/tasks"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	cmongo "github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers/mongo"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	localtask "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/util"
)

var taskSvc *TaskServer
var once sync.Once

// GetTaskServer create task manager implementation
func GetTaskServer() *TaskServer {
	once.Do(func() {
		cxt, cancel := context.WithCancel(context.Background())
		taskSvc = &TaskServer{
			cxt:    cxt,
			cancel: cancel,
			lock:   &sync.Mutex{},
		}
	})
	return taskSvc
}

// TaskServer server for go-machinery
type TaskServer struct {
	brokerOption  *options.BrokerConfig
	backendOption *cmongo.Options

	cxt    context.Context
	cancel context.CancelFunc
	lock   sync.Locker
	server *machinery.Server
	worker *machinery.Worker
}

// Init register all background task, init server
func (ts *TaskServer) Init(opt *options.BrokerConfig, backend *cmongo.Options) error {
	if opt == nil || backend == nil {
		blog.Errorf("TaskServer lost Broker or backend Config")
		return fmt.Errorf("lost broker/backend configuration")
	}
	ts.brokerOption = opt
	ts.backendOption = backend
	if err := ts.validateOption(); err != nil {
		blog.Errorf("taskserver validate broker/backend Option failed, %s", err.Error())
		return err
	}
	// init server & worker
	if err := ts.initServer(); err != nil {
		blog.Errorf("task server init go-machinery server failed, %s", err.Error())
		return err
	}
	if err := ts.initWorker(); err != nil {
		blog.Errorf("task server init go-machinery worker failed, %s", err.Error())
		return err
	}
	return nil
}

// Run running server & worker
func (ts *TaskServer) Run() error {
	return nil
}

// Stop running
func (ts *TaskServer) Stop() {
	ts.cancel()
}

// Dispatch dispatch task to worker
func (ts *TaskServer) Dispatch(task *proto.Task) error {
	// store all task information and then dispatch to remote worker
	if err := validateTask(task); err != nil {
		blog.Errorf("task %s/%s is not validate, %s", task.TaskID, task.TaskType, err.Error())
		return err
	}
	// create all task to signature
	blog.Infof("task %s/%s with steps %v ready to dispatch worker", task.TaskID, task.TaskType, task.StepSequence)
	var signatures []*tasks.Signature
	for _, stepName := range task.StepSequence {
		step := task.Steps[stepName]
		signature := &tasks.Signature{
			UUID: fmt.Sprintf("task-%s-%s", task.TaskID, stepName),
			Name: step.TaskMethod,
			// two parameters: taskID, stepName
			Args:                        []tasks.Arg{{Type: "string", Value: task.TaskID}, {Type: "string", Value: stepName}},
			IgnoreWhenTaskNotRegistered: true,
		}
		signatures = append(signatures, signature)
	}
	ts.lock.Lock()
	defer ts.lock.Unlock()

	// sending to workers
	chain, _ := tasks.NewChain(signatures...)
	// NOCC:vet/vet(设计如此)
	cxt, cancelCtx := context.WithCancel(ts.cxt)
	defer cancelCtx()
	asyncResult, err := ts.server.SendChainWithContext(cxt, chain)
	if err != nil {
		// try to re-send tasks with back-off strategy in for loop?
		blog.Errorf("sending task %s to worker failed: %s", task.TaskID, err.Error())
		return err
	}

	go func(t *proto.Task, c *tasks.Chain) { // nolint
		// check async results
		for retry := 3; retry > 0; retry-- {
			results, err := asyncResult.Get(time.Second * 5)
			if err != nil {
				blog.Errorf("tracing task %s result failed, %s. retry %d", t.TaskID, err.Error(), retry)
				continue
			}
			// check results
			blog.Infof("tracing task %s result %s", t.TaskID, tasks.HumanReadableResults(results))
		}
	}(task, chain)

	return nil
}

func (ts *TaskServer) validateOption() error {
	if len(ts.backendOption.Username) == 0 || len(ts.backendOption.Password) == 0 {
		return fmt.Errorf("backend lost username or password")
	}
	return nil
}

// init server
func (ts *TaskServer) initServer() error {
	mongoCli, err := util.NewMongoCli(ts.backendOption)
	if err != nil {
		return err
	}

	config := &config.Config{
		Broker:          ts.brokerOption.QueueAddress,
		DefaultQueue:    ts.brokerOption.Exchange,
		ResultsExpireIn: 3600 * 48,
		MongoDB: &config.MongoDBConfig{
			Client:   mongoCli,
			Database: ts.backendOption.Database,
		},
		AMQP: &config.AMQPConfig{
			Exchange:      ts.brokerOption.Exchange,
			ExchangeType:  "direct",
			BindingKey:    ts.brokerOption.Exchange,
			PrefetchCount: 50,
		},
	}
	broker := amqp.New(config)
	backend, err := mongo.New(config)
	if err != nil {
		blog.Errorf("task server init mongo backend failed, %s", err.Error())
		return err
	}
	lock := eager.New()
	ts.server = machinery.NewServer(config, broker, backend, lock)

	// get all cloud actions for registry
	allTasks := make(map[string]interface{})
	for _, mgr := range cloudprovider.GetAllTaskManager() {
		actions := mgr.GetAllTask()
		for name, task := range actions {
			if _, ok := allTasks[name]; ok {
				blog.Errorf("taskserver init failed, task %s duplicated", name)
				return fmt.Errorf("task %s duplicated", name)
			}
			allTasks[name] = task
		}
	}

	// register common actions
	for name, action := range localtask.RegisterCommonActions() {
		if _, ok := allTasks[name]; ok {
			blog.Errorf("taskserver init failed, task %s duplicated", name)
			return fmt.Errorf("task %s duplicated", name)
		}
		allTasks[name] = action
	}
	if err = ts.server.RegisterTasks(allTasks); err != nil {
		blog.Errorf("task server register tasks failed, %s", err.Error())
		return err
	}
	return nil
}

// init worker
func (ts *TaskServer) initWorker() error {
	ts.worker = ts.server.NewWorker("", 100)
	// int all kinds handler, here we inject some custom code for error handling,
	// start and end of task hooks, useful for metrics
	errorHandler := func(err error) {
		blog.Errorf("task error handler: %s", err)
	}
	preTaskHandler := func(signature *tasks.Signature) {
		blog.Infof("start task handler for: %s", signature.Name)
	}
	postTaskHandler := func(signature *tasks.Signature) {
		blog.Infof("end task handler for: %s", signature.Name)
	}
	ts.worker.SetPostTaskHandler(postTaskHandler)
	ts.worker.SetErrorHandler(errorHandler)
	ts.worker.SetPreTaskHandler(preTaskHandler)

	// start worker
	go func() {
		if err := ts.worker.Launch(); err != nil {
			errMsg := fmt.Sprintf("task server worker launch failed, %s", err.Error())
			blog.Errorf(errMsg)
			return
		}
	}()

	return nil
}
