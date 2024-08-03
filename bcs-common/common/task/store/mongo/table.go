package mongo

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
	Name   string            `json:"name" bson:"name"`
	Alias  string            `json:"alias" bson:"alias"`
	Params map[string]string `json:"params" bson:"params"`
	// step extras for string json, need client step to parse
	Extras              string `json:"extras" bson:"extras"`
	Status              string `json:"status" bson:"status"`
	Message             string `json:"message" bson:"message"`
	SkipOnFailed        bool   `json:"skipOnFailed" bson:"skipOnFailed"`
	RetryCount          uint32 `json:"retryCount" bson:"retryCount"`
	Start               string `json:"start" bson:"start"`
	End                 string `json:"end" bson:"end"`
	ExecutionTime       uint32 `json:"executionTime" bson:"executionTime"`
	MaxExecutionSeconds uint32 `json:"maxExecutionSeconds" bson:"maxExecutionSeconds"`
	LastUpdate          string `json:"lastUpdate" bson:"lastUpdate"`
}
