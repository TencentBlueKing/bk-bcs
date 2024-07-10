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

// Package task is a package for task management
package task

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/RichardKnop/machinery/v2"
	"github.com/RichardKnop/machinery/v2/backends/mongo"
	"github.com/RichardKnop/machinery/v2/brokers/amqp"
	"github.com/RichardKnop/machinery/v2/config"
	"github.com/RichardKnop/machinery/v2/locks/eager"
	"github.com/RichardKnop/machinery/v2/tasks"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/task/store"
	"github.com/Tencent/bk-bcs/bcs-common/common/task/types"
	bcsmongo "github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers/mongo"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
)

const (
	// DefaultWorkerConcurrency default worker concurrency
	DefaultWorkerConcurrency = 10
)

// BrokerConfig config for go-machinery broker
type BrokerConfig struct {
	QueueAddress string `json:"address"`
	Exchange     string `json:"exchange"`
}

// TaskManager manager for task server
type TaskManager struct { // nolint
	moduleName string
	lock       sync.Locker
	server     *machinery.Server
	worker     *machinery.Worker

	brokerConfig *BrokerConfig
	mongoConfig  *bcsmongo.Options

	workerNum     int
	stepWorkers   map[string]StepWorkerInterface
	callBackFuncs map[string]CallbackInterface

	ctx    context.Context
	cancel context.CancelFunc
}

// ManagerConfig options for manager
type ManagerConfig struct {
	ModuleName  string
	StepWorkers []StepWorkerInterface
	CallBacks   []CallbackInterface
	WorkerNum   int
	Broker      *BrokerConfig
	Backend     *bcsmongo.Options
}

// NewTaskManager create new manager
func NewTaskManager() *TaskManager {
	ctx, cancel := context.WithCancel(context.Background())

	m := &TaskManager{
		ctx:       ctx,
		cancel:    cancel,
		lock:      &sync.Mutex{},
		workerNum: DefaultWorkerConcurrency,
	}
	return m
}

// Init init machinery server and worker
func (m *TaskManager) Init(cfg *ManagerConfig) error {
	err := m.validate(cfg)
	if err != nil {
		return err
	}

	if m.stepWorkers == nil {
		m.stepWorkers = make(map[string]StepWorkerInterface)
	}

	if m.callBackFuncs == nil {
		m.callBackFuncs = make(map[string]CallbackInterface)
	}

	m.brokerConfig = cfg.Broker
	m.mongoConfig = cfg.Backend
	m.moduleName = cfg.ModuleName
	if cfg.WorkerNum != 0 {
		m.workerNum = cfg.WorkerNum
	}

	// save step workers and check duplicate
	for _, w := range cfg.StepWorkers {
		if _, ok := m.stepWorkers[w.GetName()]; ok {
			return fmt.Errorf("step [%s] already exists", w.GetName())
		}
		m.stepWorkers[w.GetName()] = w
	}

	// save callbacks and check duplicate
	for _, c := range cfg.CallBacks {
		if _, ok := m.callBackFuncs[c.GetName()]; ok {
			return fmt.Errorf("callback func [%s] already exists", c.GetName())
		}
		m.callBackFuncs[c.GetName()] = c
	}

	if err := m.initGlobalStorage(); err != nil {
		return err
	}

	if err := m.initServer(); err != nil {
		return err
	}

	if err := m.initWorker(cfg.WorkerNum); err != nil {
		return err
	}

	return nil
}

func (m *TaskManager) validate(c *ManagerConfig) error {
	// module name check
	if c.ModuleName == "" {
		return fmt.Errorf("module name is empty")
	}

	// step worker check
	if c.StepWorkers == nil || len(c.StepWorkers) == 0 {
		return fmt.Errorf("step worker is empty")
	}

	// broker config check
	if c.Broker == nil || c.Broker.Exchange == "" || c.Broker.QueueAddress == "" {
		return fmt.Errorf("broker config is empty")
	}

	if c.Backend == nil {
		return fmt.Errorf("backend config is empty")
	}

	return nil
}

func (m *TaskManager) initGlobalStorage() error {
	mongoDB, err := bcsmongo.NewDB(m.mongoConfig)
	if err != nil {
		return fmt.Errorf("init mongo db failed, err %s", err.Error())
	}
	if err = mongoDB.Ping(); err != nil {
		return fmt.Errorf("ping mongo db failed, err %s", err.Error())
	}

	if m.moduleName == "" {
		return fmt.Errorf("module name is empty")
	}
	modelSet := store.NewModelSet(mongoDB, m.moduleName)

	globalStorage = modelSet
	return nil
}

func (m *TaskManager) initServer() error {
	mongoCli, err := newMongoCli(m.mongoConfig)
	if err != nil {
		return err
	}

	serverConfig := &config.Config{
		Broker:          m.brokerConfig.QueueAddress,
		DefaultQueue:    m.brokerConfig.Exchange,
		ResultsExpireIn: 3600 * 48,
		MongoDB: &config.MongoDBConfig{
			Client:   mongoCli,
			Database: m.mongoConfig.Database,
		},
		AMQP: &config.AMQPConfig{
			Exchange:      m.brokerConfig.Exchange,
			ExchangeType:  "direct",
			BindingKey:    m.brokerConfig.Exchange,
			PrefetchCount: 50,
		},
	}
	broker := amqp.New(serverConfig)
	backend, err := mongo.New(serverConfig)
	if err != nil {
		return fmt.Errorf("task server init mongo backend failed, %s", err.Error())
	}
	lock := eager.New()
	m.server = machinery.NewServer(serverConfig, broker, backend, lock)

	return nil
}

// register step workers and init workers
func (m *TaskManager) initWorker(workerNum int) error {
	// register all workers
	if err := m.registerStepWorkers(); err != nil {
		return fmt.Errorf("register workers failed, err: %s", err.Error())
	}

	m.worker = m.server.NewWorker("", workerNum)

	preTaskHandler := func(signature *tasks.Signature) {
		blog.Infof("start task handler for: %s", signature.Name)
	}
	postTaskHandler := func(signature *tasks.Signature) {
		blog.Infof("end task handler for: %s", signature.Name)
	}
	errorHandler := func(err error) {
		blog.Infof("task error handler: %s", err)
	}

	m.worker.SetPreTaskHandler(preTaskHandler)
	m.worker.SetPostTaskHandler(postTaskHandler)
	m.worker.SetErrorHandler(errorHandler)

	return nil
}

// Run start worker
func (m *TaskManager) Run() {
	// start worker
	go func() {
		if err := m.worker.Launch(); err != nil {
			errMsg := fmt.Sprintf("task server worker launch failed: %s", err.Error())
			blog.Infof(errMsg)
			return
		}
	}()
}

// GetTaskWithID get task by taskid
func (m *TaskManager) GetTaskWithID(ctx context.Context, taskId string) (*types.Task, error) {
	return GetGlobalStorage().GetTask(ctx, taskId)
}

// ListTask return tasks with conditions
func (m *TaskManager) ListTask(ctx context.Context, cond *operator.Condition,
	opt *store.ListOption) ([]types.Task, error) {
	return GetGlobalStorage().ListTask(ctx, cond, opt)
}

// UpdateTask update task
// ! warning: modify task status will cause task status not consistent
func (m *TaskManager) UpdateTask(ctx context.Context, task *types.Task) error {
	return GetGlobalStorage().UpdateTask(ctx, task)
}

// PatchTaskInfo update task info
// ! warning: modify task status will cause task status not consistent
func (m *TaskManager) PatchTaskInfo(ctx context.Context, taskID string, patchs map[string]interface{}) error {
	return GetGlobalStorage().PatchTask(ctx, taskID, patchs)
}

// RetryAll reset status to running and dispatch all tasks
func (m *TaskManager) RetryAll(task *types.Task) error {
	task.SetStatus(types.TaskStatusRunning)
	task.SetMessage("task retrying")

	if err := GetGlobalStorage().UpdateTask(context.Background(), task); err != nil {
		return err
	}
	return m.dispatchAt(task, "")
}

// RetryAt reset status to running and dispatch tasks which begin with stepName
func (m *TaskManager) RetryAt(task *types.Task, stepName string) error {
	task.SetStatus(types.TaskStatusRunning)
	task.SetMessage("task retrying")

	if err := GetGlobalStorage().UpdateTask(context.Background(), task); err != nil {
		return err
	}
	return m.dispatchAt(task, stepName)
}

// Dispatch dispatch task
func (m *TaskManager) Dispatch(task *types.Task) error {
	if err := GetGlobalStorage().CreateTask(context.Background(), task); err != nil {
		fmt.Println(err)
		return err
	}

	return m.dispatchAt(task, "")
}

func (m *TaskManager) transTaskToSignature(task *types.Task, stepNameBegin string) ([]*tasks.Signature, error) {
	var signatures []*tasks.Signature

	for _, stepName := range task.StepSequence {
		_, ok := task.Steps[stepName]
		if !ok {
			blog.Errorf("task[%s] transTaskToSignature failed: %s not exist", stepName)
			return nil, fmt.Errorf("%s not exist", stepName)
		}

		// skip steps which before begin step, empty str not skip any steps
		if stepName != "" && stepNameBegin != "" && stepName != stepNameBegin {
			continue
		}

		// build signature from step
		signature := &tasks.Signature{
			UUID: fmt.Sprintf("%s-%s", task.TaskID, stepName),
			Name: stepName,
			// two parameters: taskID, stepName
			Args: []tasks.Arg{
				{
					Type:  "string",
					Value: task.GetTaskID(),
				},
				{
					Type:  "string",
					Value: stepName,
				},
			},
			IgnoreWhenTaskNotRegistered: true,
		}

		signatures = append(signatures, signature)
	}

	return signatures, nil
}

// dispatchAt task to machinery
func (m *TaskManager) dispatchAt(task *types.Task, stepNameBegin string) error {
	signatures, err := m.transTaskToSignature(task, stepNameBegin)
	if err != nil {
		blog.Errorf("dispatchAt task %s failed: %v", task.TaskID, err)
		return err
	}

	m.lock.Lock()
	defer m.lock.Unlock()

	// sending to workers
	chain, err := tasks.NewChain(signatures...)
	if err != nil {
		blog.Errorf("taskManager[%s] DispatchChainTask NewChain failed: %v", task.GetTaskID(), err)
		return err
	}

	// send chain to machinery & ctx for tracing
	asyncResult, err := m.server.SendChainWithContext(context.Background(), chain)
	if err != nil {
		return fmt.Errorf("send chain to machinery failed: %s", err.Error())
	}

	// get results
	go func(t *types.Task, _ *tasks.Chain) {
		// check async results
		for retry := 3; retry > 0; retry-- {
			results, err := asyncResult.Get(time.Second * 5)
			if err != nil {
				fmt.Printf("tracing task %s result failed, %s. retry %d\n", t.GetTaskID(), err.Error(), retry)
				continue
			}
			// check results
			blog.Infof("tracing task %s result %s", t.GetTaskID(), tasks.HumanReadableResults(results))
		}
	}(task, chain)

	return nil
}

// registerStepWorkers build machinery workers for all step worker
func (m *TaskManager) registerStepWorkers() error {
	allTasks := make(map[string]interface{}, 0)

	for stepName, stepWorker := range m.stepWorkers {
		do := stepWorker.DoWork

		t := func(taskID string, stepName string) error {
			defer RecoverPrintStack(fmt.Sprintf("%s-%s", taskID, stepName))

			blog.Infof("start to execute task[%s] stepName[%s]", taskID, stepName)

			start := time.Now()
			state, step, err := getTaskStateAndCurrentStep(taskID, stepName, m.callBackFuncs)
			if err != nil {
				blog.Errorf("task[%s] stepName[%s] getTaskStateAndCurrentStep failed: %v",
					taskID, stepName, err)
				return err
			}
			// step executed success
			if step == nil {
				blog.Infof("task[%s] stepName[%s] already exec successful && skip",
					taskID, stepName, err)
				return nil
			}

			// step timeout
			stepCtx, stepCancel := GetTimeOutCtx(m.ctx, step.MaxExecutionSeconds)
			defer stepCancel()
			// task timeout
			t, _ := state.task.GetStartTime()
			taskCtx, taskCancel := GetDeadlineCtx(m.ctx, &t, state.task.MaxExecutionSeconds)
			defer taskCancel()

			tmpCh := make(chan error, 1)
			go func() {
				defer RecoverPrintStack(fmt.Sprintf("%s-%s", taskID, stepName))
				// call step worker
				tmpCh <- do(state.task)
			}()

			select {
			case errLocal := <-tmpCh:
				blog.Infof("task %s step %s errLocal: %v", taskID, stepName, errLocal)

				// update task & step status
				if errLocal != nil {
					if err := state.updateStepFailure(start, step.GetName(), errLocal, false); err != nil {
						blog.Infof("task %s update step %s to failure failed: %s",
							taskID, step.GetName(), errLocal.Error())
					}
				} else {
					if err := state.updateStepSuccess(start, step.GetName()); err != nil {
						blog.Infof("task %s update step %s to success failed: %s",
							taskID, step.GetName(), err.Error())
					}
				}

				if errLocal != nil && !step.GetSkipOnFailed() {
					return errLocal
				}

				return nil
			case <-stepCtx.Done():
				retErr := fmt.Errorf("task %s step %s timeout", taskID, step.GetName())
				errLocal := state.updateStepFailure(start, step.GetName(), retErr, false)
				if errLocal != nil {
					blog.Infof("update step %s to failure failed: %s", step.GetName(), errLocal.Error())
				}
				if !step.GetSkipOnFailed() {
					return retErr
				}
				return nil
			case <-taskCtx.Done():
				// task timeOut
				retErr := fmt.Errorf("task %s exec timeout", taskID)
				errLocal := state.updateStepFailure(start, step.GetName(), retErr, true)
				if errLocal != nil {
					blog.Errorf("update step %s to failure failed: %s", step.GetName(), errLocal.Error())
				}

				return retErr
			}
		}

		if _, ok := allTasks[stepName]; ok {
			return fmt.Errorf("task %s already exists", stepName)
		}
		allTasks[stepName] = t
	}
	err := m.server.RegisterTasks(allTasks)
	return err
}

// Stop running
func (m *TaskManager) Stop() {
	m.worker.Quit()
	m.cancel()
}
