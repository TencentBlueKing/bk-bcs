# BCS异步任务流框架

## 背景

* 主要为了解决项目中复杂场景(集群管理任务、资源管理、联邦集群管理任务等)的分布式并发任务处理及任务编排场景，通过统一的框架实现松耦合、易扩展的特性的任务管理系统

## 方案

## 技术方案

* 基于 [go-machinery](https://github.com/RichardKnop/machinery) 的 [Workflows](https://github.com/RichardKnop/machinery?tab=readme-ov-file#workflows) 的 **Chains** 模式，通过上层任务抽象，实现异步任务框架处理
* 依赖：消息队列rabbitmmq、数据库mongo

### 任务框架

### 支持的组件

* Brokers: etcd
* Locks：etcd
* Backends: etcd

### 任务框架实现功能

* 基于协程级别的轻量级任务执行
* 支持水平扩展，提升任务处理并发量
* 支持任务流处理
* 支持子任务变量共享
* 支持任务从当前失败节点重试
* 支持任务取消
* 支持任务跳过失败子任务
* 支持指定任务节点运行
* 支持自定义的任务回调机制
* 可扩展的变量渲染
* 子任务超时控制、任务超时控制机制

### 任务模型

抽象任务结构如下所示，Task是主任务，Step是工作流子任务，通过StepSequence控制执行的顺序。

```
// Task task definition
type Task struct {
	// index for task, client should set this field
	TaskIndex string `json:"taskIndex" bson:"taskIndex"`
	TaskID    string `json:"taskId" bson:"taskId"`
	TaskType  string `json:"taskType" bson:"taskType"`
	TaskName  string `json:"taskName" bson:"taskName"`
	// steps and params
	CurrentStep      string            `json:"currentStep" bson:"currentStep"`
	StepSequence     []string          `json:"stepSequence" bson:"stepSequence"`
	Steps            map[string]*Step  `json:"steps" bson:"steps"`
	CallBackFuncName string            `json:"callBackFuncName" bson:"callBackFuncName"`
	CommonParams     map[string]string `json:"commonParams" bson:"commonParams"`
	ExtraJson        string            `json:"extraJson" bson:"extraJson"`

	Status              string `json:"status" bson:"status"`
	Message             string `json:"message" bson:"message"`
	ForceTerminate      bool   `json:"forceTerminate" bson:"forceTerminate"`
	Start               string `json:"start" bson:"start"`
	End                 string `json:"end" bson:"end"`
	ExecutionTime       uint32 `json:"executionTime" bson:"executionTime"`
	MaxExecutionSeconds uint32 `json:"maxExecutionSeconds" bson:"maxExecutionSeconds"`
	Creator             string `json:"creator" bson:"creator"`
	LastUpdate          string `json:"lastUpdate" bson:"lastUpdate"`
	Updater             string `json:"updater" bson:"updater"`
}

// Step step definition
type Step struct {
	Name     string            `json:"name" bson:"name"`
	Method   string            `json:"method" bson:"method"`
	StepName string            `json:"stepname" bson:"stepname"`
	Params   map[string]string `json:"params" bson:"params"`
	// step extras for string json, need client step to parse
	Extras       string `json:"extras" bson:"extras"`
	Status       string `json:"status" bson:"status"`
	Message      string `json:"message" bson:"message"`
	SkipOnFailed bool   `json:"skipOnFailed" bson:"skipOnFailed"`
	RetryCount   uint32 `json:"retryCount" bson:"retryCount"`

	Start               string `json:"start" bson:"start"`
	End                 string `json:"end" bson:"end"`
	ExecutionTime       uint32 `json:"executionTime" bson:"executionTime"`
	MaxExecutionSeconds uint32 `json:"maxExecutionSeconds" bson:"maxExecutionSeconds"`
	LastUpdate          string `json:"lastUpdate" bson:"lastUpdate"`
}

// StepWorkerInterface that client must implement
type StepWorkerInterface interface {
	GetMethod() string
	DoWork(task *types.Task) error
}

// CallbackInterface that client must implement
type CallbackInterface interface {
	GetName() string
	Callback(isSuccess bool, task *types.Task)
}

// TaskMgr build task
type TaskMgr interface {
	Name() string
	Type() string
	BuildTask(info types.TaskInfo, opts ...types.TaskOption) (*types.Task, error)
	Steps(defineSteps []StepMgr) []*types.Step
}

// StepMgr build step
type StepMgr interface {
	Name() string
	GetMethod() string
	BuildStep(kvs []KeyValue, opts ...types.StepOption) *types.Step
	DoWork(task *types.Task) error
}
```
* Task 是主任务，控制子任务执行顺序以及子任务的执行参数，并存储子任务共享变量，同时负责子任务的切换以及主任务的状态更新。
* Step 是工作流子任务，进一步抽象是 接口StepWorkerInterface，实现重要的业务逻辑。而通过对接口StepWorkerInterface抽象封装来实现任务切换和子任务状态更新
* StepMgr 为构建step子任务 以及 step子任务的业务逻辑执行体
* TaskMgr 为构建task任务
* CallbackInterface 注册回调方法

### 示例代码

接入框架的示例代码可参考 task 目录下的 example 例子