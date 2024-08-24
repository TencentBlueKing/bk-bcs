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
	ibackend "github.com/RichardKnop/machinery/v2/backends/iface"
	ibroker "github.com/RichardKnop/machinery/v2/brokers/iface"
	"github.com/RichardKnop/machinery/v2/config"
	ilock "github.com/RichardKnop/machinery/v2/locks/iface"
	"github.com/RichardKnop/machinery/v2/log"
	"github.com/RichardKnop/machinery/v2/tasks"

	irevoker "github.com/Tencent/bk-bcs/bcs-common/common/task/brokers/iface"
	istep "github.com/Tencent/bk-bcs/bcs-common/common/task/steps/iface"
	istore "github.com/Tencent/bk-bcs/bcs-common/common/task/stores/iface"
	"github.com/Tencent/bk-bcs/bcs-common/common/task/types"
)

const (
	// DefaultWorkerConcurrency default worker concurrency
	DefaultWorkerConcurrency = 10
)

// BrokerConfig config for go-machinery broker

// TaskManager manager for task server
type TaskManager struct { // nolint
	moduleName string
	lock       sync.Locker
	server     *machinery.Server
	worker     *machinery.Worker

	workerNum         int
	stepExecutors     map[istep.StepName]istep.StepExecutor
	callbackExecutors map[istep.CallbackName]istep.CallbackExecutor
	cfg               *ManagerConfig
	store             istore.Store

	ctx    context.Context
	cancel context.CancelFunc
}

// ManagerConfig options for manager
type ManagerConfig struct {
	ModuleName   string
	WorkerName   string
	WorkerNum    int
	Broker       ibroker.Broker
	Backend      ibackend.Backend
	Lock         ilock.Lock
	Store        istore.Store
	ServerConfig *config.Config
}

// NewTaskManager create new manager
func NewTaskManager() *TaskManager {
	ctx, cancel := context.WithCancel(context.Background())

	m := &TaskManager{
		ctx:           ctx,
		cancel:        cancel,
		lock:          &sync.Mutex{},
		workerNum:     DefaultWorkerConcurrency,
		stepExecutors: istep.GetRegisters(), // get all step workers
		cfg:           &ManagerConfig{},
	}
	return m
}

// Init init machinery server and worker
func (m *TaskManager) Init(cfg *ManagerConfig) error {
	err := m.validate(cfg)
	if err != nil {
		return err
	}

	if cfg.ServerConfig == nil {
		cfg.ServerConfig = &config.Config{
			ResultsExpireIn: 3600 * 48,
			NoUnixSignals:   true,
		}
	}
	m.cfg = cfg
	m.store = cfg.Store

	if m.stepExecutors == nil {
		m.stepExecutors = make(map[istep.StepName]istep.StepExecutor)
	}

	m.callbackExecutors = istep.GetCallbackRegisters()

	m.moduleName = cfg.ModuleName
	if cfg.WorkerNum != 0 {
		m.workerNum = cfg.WorkerNum
	}

	if err := m.initGlobalStorage(); err != nil {
		return err
	}

	if err := m.initServer(); err != nil {
		return err
	}

	if err := m.initWorker(cfg.WorkerName, cfg.WorkerNum); err != nil {
		return err
	}

	return nil
}

func (m *TaskManager) initGlobalStorage() error {
	globalStorage = m.store
	return nil
}

func (m *TaskManager) validate(c *ManagerConfig) error {
	// module name check
	if c.ModuleName == "" {
		return fmt.Errorf("module name is empty")
	}

	return nil
}

func (m *TaskManager) initServer() error {
	m.server = machinery.NewServer(m.cfg.ServerConfig, m.cfg.Broker, m.cfg.Backend, m.cfg.Lock)

	return nil
}

// register step workers and init workers
func (m *TaskManager) initWorker(workerName string, workerNum int) error {
	// register all workers
	if err := m.registerStepWorkers(); err != nil {
		return fmt.Errorf("register workers failed, err: %s", err.Error())
	}

	m.worker = m.server.NewWorker(workerName, workerNum)

	preTaskHandler := func(signature *tasks.Signature) {
		log.INFO.Printf("start task[%s] handler for: %s", signature.UUID, signature.Name)
	}
	postTaskHandler := func(signature *tasks.Signature) {
		log.INFO.Printf("end task[%s] handler for: %s", signature.UUID, signature.Name)
	}
	errorHandler := func(err error) {
		log.INFO.Printf("task error handler: %s", err)
	}

	m.worker.SetPreTaskHandler(preTaskHandler)
	m.worker.SetPostTaskHandler(postTaskHandler)
	m.worker.SetErrorHandler(errorHandler)

	return nil
}

// Run start worker
func (m *TaskManager) Run() error {
	return m.worker.Launch()
}

// GetTaskWithID get task by taskid
func (m *TaskManager) GetTaskWithID(ctx context.Context, taskId string) (*types.Task, error) {
	return GetGlobalStorage().GetTask(ctx, taskId)
}

// UpdateTask update task
// ! warning: modify task status will cause task status not consistent
func (m *TaskManager) UpdateTask(ctx context.Context, task *types.Task) error {
	return GetGlobalStorage().UpdateTask(ctx, task)
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

// Revoke revoke the task
func (m *TaskManager) Revoke(task *types.Task) error {
	revoker, ok := m.cfg.Broker.(irevoker.Revoker)
	if !ok {
		return types.ErrNotImplemented
	}

	task.SetStatus(types.TaskStatusRevoke)
	task.SetMessage("task revoked")

	if err := GetGlobalStorage().UpdateTask(context.Background(), task); err != nil {
		return err
	}
	return revoker.Revoke(context.Background(), task.TaskID)
}

// Dispatch dispatch task
func (m *TaskManager) Dispatch(task *types.Task) error {
	if err := task.Validate(); err != nil {
		return err
	}

	if err := GetGlobalStorage().CreateTask(context.Background(), task); err != nil {
		return err
	}

	return m.dispatchAt(task, "")
}

func (m *TaskManager) transTaskToSignature(task *types.Task, stepNameBegin string) []*tasks.Signature {
	var signatures []*tasks.Signature

	for _, step := range task.Steps {
		// skip steps which before begin step, empty str not skip any steps
		if step.Name != "" && stepNameBegin != "" && step.Name != stepNameBegin {
			continue
		}

		// build signature from step
		signature := &tasks.Signature{
			UUID: fmt.Sprintf("%s-%s", task.TaskID, step.Name),
			Name: step.Executor,
			ETA:  step.ETA,
			// two parameters: taskID, stepName
			Args: []tasks.Arg{
				{
					Name:  "task_id",
					Type:  "string",
					Value: task.GetTaskID(),
				},
				{
					Name:  "step_name",
					Type:  "string",
					Value: step.Name,
				},
			},

			IgnoreWhenTaskNotRegistered: true,
		}

		signatures = append(signatures, signature)
	}

	return signatures
}

// dispatchAt task to machinery
func (m *TaskManager) dispatchAt(task *types.Task, stepNameBegin string) error {
	signatures := m.transTaskToSignature(task, stepNameBegin)

	m.lock.Lock()
	defer m.lock.Unlock()

	// sending to workers
	chain, err := tasks.NewChain(signatures...)
	if err != nil {
		log.ERROR.Printf("taskManager[%s] DispatchChainTask NewChain failed: %v", task.GetTaskID(), err)
		return err
	}

	// send chain to machinery & ctx for tracing
	_, err = m.server.SendChainWithContext(context.Background(), chain)
	if err != nil {
		return fmt.Errorf("send chain to machinery failed: %s", err.Error())
	}

	return nil
}

// registerStepWorkers build machinery workers for all step worker
func (m *TaskManager) registerStepWorkers() error {
	allTasks := make(map[string]interface{}, 0)
	for stepName := range m.stepExecutors {
		name := string(stepName)
		if _, ok := allTasks[name]; ok {
			return fmt.Errorf("task %s already exists", name)
		}
		allTasks[name] = m.doWork
	}
	err := m.server.RegisterTasks(allTasks)
	return err
}

// doWork machinery 通用处理函数
func (m *TaskManager) doWork(taskID string, stepName string) error { // nolint
	defer RecoverPrintStack(fmt.Sprintf("%s-%s", taskID, stepName))

	log.INFO.Printf("start to execute task[%s] stepName[%s]", taskID, stepName)

	state, err := m.getTaskState(taskID, stepName)
	if err != nil {
		log.ERROR.Printf("task[%s] stepName[%s] getTaskState failed: %v",
			taskID, stepName, err)
		return err
	}

	// step executed success
	if state.step == nil {
		log.INFO.Printf("task[%s] stepName[%s] already exec successful && skip",
			taskID, stepName)
		return nil
	}

	step := state.step
	stepExecutor, ok := m.stepExecutors[istep.StepName(step.Executor)]
	if !ok {
		log.ERROR.Printf("task[%s] stepName[%s] executor[%s] not found", taskID, stepName, state.step.Executor)
		return fmt.Errorf("step executor[%s] not found", step.Executor)
	}

	start := time.Now()
	// step timeout
	stepCtx, stepCancel := GetTimeOutCtx(m.ctx, step.MaxExecutionSeconds)
	defer stepCancel()

	// task revoke
	revokeCtx := context.TODO()
	if m.cfg != nil && m.cfg.Broker != nil {
		if revoker, ok := m.cfg.Broker.(irevoker.Revoker); ok {
			revokeCtx = revoker.RevokeCtx(taskID)
		}
	}

	// task timeout
	t := state.task.GetStartTime()
	taskCtx, taskCancel := GetDeadlineCtx(m.ctx, &t, state.task.MaxExecutionSeconds)
	defer taskCancel()

	tmpCh := make(chan error, 1)
	go func() {
		// call step worker
		execCtx := istep.NewContext(stepCtx, GetGlobalStorage(), state.GetTask(), step)
		tmpCh <- stepExecutor.Execute(execCtx)
	}()

	select {
	case retErr := <-tmpCh:
		log.INFO.Printf("task %s step %s errLocal: %v", taskID, stepName, retErr)

		// update task & step status
		if retErr == nil {
			state.updateStepSuccess(start)
			return nil
		}
		state.updateStepFailure(start, retErr, false)

		if step.GetSkipOnFailed() {
			return nil
		}

		if step.GetRetryCount() < step.MaxRetries {
			retryIn := time.Second * time.Duration(retryNext(int(step.GetRetryCount())))
			log.INFO.Printf("retry task %s step %s, retried=%d, maxRetries=%d, retryIn=%s",
				taskID, step.GetName(), step.GetRetryCount(), step.MaxRetries, retryIn)
			return tasks.NewErrRetryTaskLater(retErr.Error(), retryIn)
		}

		return retErr

	case <-stepCtx.Done():
		retErr := fmt.Errorf("task %s step %s timeout", taskID, step.GetName())
		state.updateStepFailure(start, retErr, false)

		if step.GetSkipOnFailed() {
			return nil
		}

		if step.GetRetryCount() < step.MaxRetries {
			retryIn := time.Second * time.Duration(retryNext(int(step.GetRetryCount())))
			log.INFO.Printf("retry task %s step %s, retried=%d, maxRetries=%d, retryIn=%s",
				taskID, step.GetName(), step.GetRetryCount(), step.MaxRetries, retryIn)
			return tasks.NewErrRetryTaskLater("some error", retryIn)
		}

		return retErr

	case <-revokeCtx.Done():
		// task revoke
		retErr := fmt.Errorf("task %s step %s has been revoke", taskID, step.GetName())
		state.updateStepFailure(start, retErr, false)

		// 取消指令, 不再重试
		return retErr

	case <-taskCtx.Done():
		// task timeOut
		retErr := fmt.Errorf("task %s exec timeout", taskID)
		state.updateStepFailure(start, retErr, true)

		// 整个任务结束
		return retErr
	}
}

// Stop running
func (m *TaskManager) Stop() {
	// should set NoUnixSignals
	m.worker.Quit()
	m.cancel()
}
