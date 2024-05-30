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
	driver "go.mongodb.org/mongo-driver/mongo"
	mopt "go.mongodb.org/mongo-driver/mongo/options"

	"github.com/Tencent/bk-bcs/bcs-common/common/task/store"
	types "github.com/Tencent/bk-bcs/bcs-common/common/task/types"
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

// Manager manager for task server
type Manager struct {
	moduleName string
	lock       sync.Locker
	server     *machinery.Server
	worker     *machinery.Worker

	brokerConfig *BrokerConfig
	mongoConfig  *bcsmongo.Options

	workerNum     int
	stepWorkers   map[string]StepWorkerInterface
	callBackFuncs map[string]CallbackInterface
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

// NewManager create new manager
func NewManager() *Manager {
	m := &Manager{
		lock:      &sync.Mutex{},
		workerNum: DefaultWorkerConcurrency,
	}
	return m
}

// Init init machinery server and worker
func (m *Manager) Init(cfg *ManagerConfig) error {
	err := m.validate(cfg)
	if err != nil {
		return err
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

func (m *Manager) validate(c *ManagerConfig) error {
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

func (m *Manager) initGlobalStorage() error {
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

func (m *Manager) initServer() error {
	mongoCli, err := newMongoCli(m.mongoConfig)
	if err != nil {
		return err
	}

	config := &config.Config{
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
	broker := amqp.New(config)
	backend, err := mongo.New(config)
	if err != nil {
		return fmt.Errorf("task server init mongo backend failed, %s", err.Error())
	}
	lock := eager.New()
	m.server = machinery.NewServer(config, broker, backend, lock)

	return nil
}

// register step workers and init workers
func (m *Manager) initWorker(workerNum int) error {
	// register all workers
	if err := m.registerStepWorkers(); err != nil {
		return fmt.Errorf("register workers failed, err: %s", err.Error())
	}

	m.worker = m.server.NewWorker("", workerNum)

	preTaskHandler := func(signature *tasks.Signature) {
		fmt.Printf("start task handler for: %s", signature.Name)
	}
	postTaskHandler := func(signature *tasks.Signature) {
		fmt.Printf("end task handler for: %s", signature.Name)
	}
	errorHandler := func(err error) {
		fmt.Printf("task error handler: %s", err)
	}

	m.worker.SetPreTaskHandler(preTaskHandler)
	m.worker.SetPostTaskHandler(postTaskHandler)
	m.worker.SetErrorHandler(errorHandler)

	return nil
}

// Run start worker
func (m *Manager) Run() {
	// start worker
	go func() {
		if err := m.worker.Launch(); err != nil {
			errMsg := fmt.Sprintf("task server worker launch failed, %s", err.Error())
			panic(errMsg)
		}
	}()
}

// GetTaskWithID get task by taskid
func (m *Manager) GetTaskWithID(ctx context.Context, taskid string) (*types.Task, error) {
	return getGlobalStorage().GetTask(ctx, taskid)
}

// ListTask return tasks with conditions
func (m *Manager) ListTask(ctx context.Context, cond *operator.Condition, opt *store.ListOption) ([]types.Task, error) {
	return getGlobalStorage().ListTask(ctx, cond, opt)
}

// UpdateTask update task
// ! warning: modify task status will cause task status not consistent
func (m *Manager) UpdateTask(ctx context.Context, task *types.Task) error {
	return getGlobalStorage().UpdateTask(ctx, task)
}

// PatchTaskInfo update task info
// ! warning: modify task status will cause task status not consistent
func (m *Manager) PatchTaskInfo(ctx context.Context, taskID string, patchs map[string]interface{}) error {
	// warning:
	return getGlobalStorage().PatchTask(ctx, taskID, patchs)
}

// RetryAll reset status to running and dispatch all tasks
func (m *Manager) RetryAll(task *types.Task) error {
	task.SetStatus(types.TaskStatusRunning)
	task.SetMessage("task retrying")

	if err := getGlobalStorage().UpdateTask(context.Background(), task); err != nil {
		return err
	}
	return m.dispatchAt(task, "")
}

// RetryAt reset status to running and dispatch tasks which begin with stepName
func (m *Manager) RetryAt(task *types.Task, stepName string) error {
	task.SetStatus(types.TaskStatusRunning)
	task.SetMessage("task retrying")

	if err := getGlobalStorage().UpdateTask(context.Background(), task); err != nil {
		return err
	}
	return m.dispatchAt(task, stepName)
}

// Dispatch dispatch task
func (m *Manager) Dispatch(task *types.Task) error {
	if err := getGlobalStorage().CreateTask(context.Background(), task); err != nil {
		return err
	}
	return m.dispatchAt(task, "")
}

// dispatchAt task to machinery
func (m *Manager) dispatchAt(task *types.Task, stepNameBegin string) error {
	var signatures []*tasks.Signature
	for _, stepName := range task.StepSequence {
		// skip steps which before begin step, empty str not skip any steps
		if stepName != "" && stepName != stepNameBegin {
			continue
		}
		signature := &tasks.Signature{
			UUID: fmt.Sprintf("task-%s-%s", task.GetTaskID(), stepName),
			Name: stepName,
			// two parameters: taskID, stepName
			Args:                        []tasks.Arg{{Type: "string", Value: task.GetTaskID()}, {Type: "string", Value: stepName}},
			IgnoreWhenTaskNotRegistered: true,
		}
		signatures = append(signatures, signature)
	}

	m.lock.Lock()
	defer m.lock.Unlock()

	// create chain
	chain, _ := tasks.NewChain(signatures...)

	ctx, cancelFunc := context.WithCancel(context.Background())
	if task.GetMaxExecutionSeconds() != time.Duration(0) {
		ctx, cancelFunc = context.WithTimeout(ctx, task.GetMaxExecutionSeconds())
	}
	defer cancelFunc()

	//send chain to machinery
	asyncResult, err := m.server.SendChainWithContext(ctx, chain)
	if err != nil {
		return fmt.Errorf("send chain to machinery failed: %s", err.Error())
	}

	// get results
	go func(t *types.Task, c *tasks.Chain) {
		// check async results
		for retry := 3; retry > 0; retry-- {
			results, err := asyncResult.Get(time.Second * 5)
			if err != nil {
				fmt.Printf("tracing task %s result failed, %s. retry %d", t.GetTaskID(), err.Error(), retry)
				continue
			}
			// check results
			fmt.Printf("tracing task %s result %s", t.GetTaskID(), tasks.HumanReadableResults(results))
		}
	}(task, chain)

	return nil
}

// registerStepWorkers build machinery workers for all step worker
func (m *Manager) registerStepWorkers() error {
	allTasks := make(map[string]interface{}, 0)
	for stepName, stepWorker := range m.stepWorkers {
		do := stepWorker.DoWork

		t := func(taskID string, stepName string) error {
			start := time.Now()
			state, step, err := m.getTaskStateAndCurrentStep(taskID, stepName)
			if err != nil {
				return err
			}
			// step executed success
			if step == nil {
				return nil
			}

			ctx, cancel := context.WithTimeout(context.Background(), step.GetMaxExecutionSeconds())
			defer cancel()
			stepDone := make(chan bool, 1)

			go func() {
				// call step worker
				if err = do(state); err != nil {
					if err := state.updateStepFailure(start, step.GetStepName(), err); err != nil {
						fmt.Printf("update step %s to failure failed: %s", step.GetStepName(), err.Error())
					}

					// step done
					stepDone <- true
					return
				}

				if err := state.updateStepSuccess(start, step.GetStepName()); err != nil {
					fmt.Printf("update step %s to success failed: %s", step.GetStepName(), err.Error())
				}

				// step done
				stepDone <- true
			}()

			select {
			case <-ctx.Done():
				retErr := fmt.Errorf("step %s timeout", step.GetStepName())
				if err := state.updateStepFailure(start, step.GetStepName(), retErr); err != nil {
					fmt.Printf("update step %s to failure failed: %s", step.GetStepName(), err.Error())
				}
				if !step.GetSkipOnFailed() {
					return retErr
				}
				return nil
			case <-stepDone:
				// step done
				if err != nil && !step.GetSkipOnFailed() {
					return err
				}
				return nil
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

// getTaskStateAndCurrentStep get task state and current step
func (m *Manager) getTaskStateAndCurrentStep(taskid, stepName string) (*State, *types.Step, error) {
	task, err := getGlobalStorage().GetTask(context.Background(), taskid)
	if err != nil {
		return nil, nil, fmt.Errorf("get task %s information failed, %s", taskid, err.Error())
	}

	if task.CommonParams == nil {
		task.CommonParams = make(map[string]string, 0)
	}

	state := NewState(task, stepName)
	if state.isTaskTerminated() {
		return nil, nil, fmt.Errorf("task %s is terminated, step %s skip", taskid, stepName)
	}
	step, err := state.isReadyToStep(stepName)
	if err != nil {
		return nil, nil, fmt.Errorf("task %s step %s is not ready, %s", taskid, stepName, err.Error())
	}
	if step == nil {
		// step successful and skip
		return state, nil, nil
	}

	// inject call back func
	if state.task.GetCallback() != "" {
		if callback, ok := m.callBackFuncs[state.task.GetCallback()]; ok {
			state.callBack = callback.Callback
		}
	}

	return state, nil, nil
}

func newMongoCli(opt *bcsmongo.Options) (*driver.Client, error) {
	credential := mopt.Credential{
		AuthMechanism: opt.AuthMechanism,
		AuthSource:    opt.AuthDatabase,
		Username:      opt.Username,
		Password:      opt.Password,
		PasswordSet:   true,
	}
	if len(credential.AuthMechanism) == 0 {
		credential.AuthMechanism = "SCRAM-SHA-256"
	}
	// construct mongo client options
	mCliOpt := &mopt.ClientOptions{
		Auth:  &credential,
		Hosts: opt.Hosts,
	}
	if opt.MaxPoolSize != 0 {
		mCliOpt.MaxPoolSize = &opt.MaxPoolSize
	}
	if opt.MinPoolSize != 0 {
		mCliOpt.MinPoolSize = &opt.MinPoolSize
	}
	var timeoutDuration time.Duration
	if opt.ConnectTimeoutSeconds != 0 {
		timeoutDuration = time.Duration(opt.ConnectTimeoutSeconds) * time.Second
	}
	mCliOpt.ConnectTimeout = &timeoutDuration

	// create mongo client
	mCli, err := driver.NewClient(mCliOpt)
	if err != nil {
		return nil, err
	}
	// connect to mongo
	if err = mCli.Connect(context.TODO()); err != nil {
		return nil, err
	}

	if err = mCli.Ping(context.TODO(), nil); err != nil {
		return nil, err
	}

	return mCli, nil
}
