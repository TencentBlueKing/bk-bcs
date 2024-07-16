package mysql

import (
	"time"

	"gorm.io/gorm"
)

type BaseModel struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type TaskRecords struct {
	BaseModel
	TaskType            string            `json:"taskType" gorm:"taskType"`
	TaskName            string            `json:"taskName" gorm:"taskName"`
	CurrentStep         string            `json:"currentStep" gorm:"currentStep"`
	StepSequence        []string          `json:"stepSequence" gorm:"stepSequence"`
	Steps               map[string]int64  `json:"steps" gorm:"steps"`
	CallBackFuncName    string            `json:"callBackFuncName" gorm:"callBackFuncName"`
	CommonParams        map[string]string `json:"commonParams" gorm:"commonParams"`
	ExtraJson           string            `json:"extraJson" gorm:"extraJson"`
	Status              string            `json:"status" gorm:"status"`
	Message             string            `json:"message" gorm:"message"`
	ForceTerminate      bool              `json:"forceTerminate" gorm:"forceTerminate"`
	Start               string            `json:"start" gorm:"start"`
	End                 string            `json:"end" gorm:"end"`
	ExecutionTime       uint32            `json:"executionTime" gorm:"executionTime"`
	MaxExecutionSeconds uint32            `json:"maxExecutionSeconds" gorm:"maxExecutionSeconds"`
	Creator             string            `json:"creator" gorm:"creator"`
	LastUpdate          string            `json:"lastUpdate" gorm:"lastUpdate"`
	Updater             string            `json:"updater" gorm:"updater"`
}

type StepRecords struct {
	BaseModel
	Name                string            `json:"name" gorm:"name"`
	Alias               string            `json:"alias" gorm:"alias"`
	Input               map[string]string `json:"input" gorm:"input"`
	Output              map[string]string `json:"output" gorm:"output"`
	Extras              string            `json:"extras" gorm:"extras"`
	Status              string            `json:"status" gorm:"status"`
	Message             string            `json:"message" gorm:"message"`
	SkipOnFailed        bool              `json:"skipOnFailed" gorm:"skipOnFailed"`
	RetryCount          uint32            `json:"retryCount" gorm:"retryCount"`
	Start               string            `json:"start" gorm:"start"`
	End                 string            `json:"end" gorm:"end"`
	ExecutionTime       uint32            `json:"executionTime" gorm:"executionTime"`
	MaxExecutionSeconds uint32            `json:"maxExecutionSeconds" gorm:"maxExecutionSeconds"`
	LastUpdate          string            `json:"lastUpdate" gorm:"lastUpdate"`
}
