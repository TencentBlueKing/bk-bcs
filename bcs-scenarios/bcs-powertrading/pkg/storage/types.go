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

// Package storage xxx
package storage

import (
	"time"

	"github.com/google/uuid"

	powertrading "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-powertrading/proto"
)

const (
	// TaskRunning running
	TaskRunning = "Running"
	// TaskFinished finished
	TaskFinished = "Finished"
	// TaskFailed failed
	TaskFailed = "Failed"
	// TaskWaiting waiting
	TaskWaiting = "Waiting"
)

const (
	// MemoryCheck memory check
	MemoryCheck = "memoryCheck"
	// BkOpsTaskCheck task check
	BkOpsTaskCheck = "bksopsCheck"
	// ClusterCheck cluster check
	ClusterCheck = "clusterCheck"
	// ImportedCheck imported check
	ImportedCheck = "importedCheck"
	// BkOpsTaskClean task clean
	BkOpsTaskClean = "bksopsClean"
	// BusinessCheck business check
	BusinessCheck = "businessCheck"
)

const (
	// CheckTask check task
	CheckTask = "machineCheck"
	// CleanTask clean task
	CleanTask = "machineClean"
)

const (
	// MachineCheckSuccess success
	MachineCheckSuccess = "success"
	// MachineNeedClean need clean
	MachineNeedClean = "needClean"
	// MachineCheckFailure failure
	MachineCheckFailure = "failure"
)

// MachineTask machine task
type MachineTask struct {
	TaskID       string                         `json:"taskId" bson:"taskID"`
	Message      string                         `json:"message" bson:"message"`
	CurrentStep  string                         `json:"currentStep" bson:"currentStep"`
	Status       string                         `json:"status" bson:"status"`
	BusinessID   string                         `json:"businessID" bson:"businessID"`
	IPList       []string                       `json:"ipList" bson:"ipList"`
	Source       string                         `json:"source" bson:"source"`
	DevicePoolID string                         `json:"devicePoolID" bson:"devicePoolID"`
	UpdateTime   time.Time                      `json:"updatedTime" bson:"updatedTime"`
	CreateTime   time.Time                      `json:"createTime" bson:"createTime"`
	RetryTimes   int                            `json:"retryTimes" bson:"retryTimes"`
	Detail       map[string]*TaskDetail         `json:"detail" bson:"detail"`
	Type         string                         `json:"type" bson:"type"`
	Summary      map[string]map[string][]string `json:"summary" bson:"summary"`
}

// TaskDetail task detail
type TaskDetail struct {
	Message      string                                      `json:"message" bson:"message"`
	BksOpsTaskID string                                      `json:"bksOpsTaskID" bson:"bksOpsTaskID"`
	JobID        string                                      `json:"jobID" bson:"jobID"`
	Status       string                                      `json:"status" bson:"status"`
	IPList       []string                                    `json:"IPList" bson:"IPList"`
	DetailList   map[string]*powertrading.MachineTestMessage `json:"detailList" bson:"detailList"`
}

// InitNewCleanMachineTask init new clean task
func InitNewCleanMachineTask() *MachineTask {
	task := &MachineTask{
		TaskID:       uuid.New().String(),
		Status:       TaskRunning,
		CurrentStep:  BkOpsTaskClean,
		DevicePoolID: "",
		UpdateTime:   time.Now(),
		CreateTime:   time.Now(),
		Detail:       make(map[string]*TaskDetail),
		Type:         CleanTask,
		Summary:      make(map[string]map[string][]string),
	}
	task.Detail[BkOpsTaskClean] = &TaskDetail{
		Status:     TaskRunning,
		DetailList: make(map[string]*powertrading.MachineTestMessage),
	}
	task.Detail[BkOpsTaskCheck] = &TaskDetail{
		Status:     TaskWaiting,
		DetailList: make(map[string]*powertrading.MachineTestMessage),
	}
	task.Summary[MachineCheckSuccess] = make(map[string][]string, 0)
	task.Summary[MachineCheckFailure] = make(map[string][]string, 0)
	task.Summary[MachineNeedClean] = make(map[string][]string, 0)
	return task
}

// InitNewCheckMachineTask init new check task
func InitNewCheckMachineTask() *MachineTask {
	task := &MachineTask{
		TaskID:       uuid.New().String(),
		Status:       TaskRunning,
		CurrentStep:  BusinessCheck,
		DevicePoolID: "",
		UpdateTime:   time.Now(),
		CreateTime:   time.Now(),
		Detail:       make(map[string]*TaskDetail),
		Type:         CheckTask,
		Summary:      make(map[string]map[string][]string),
	}
	task.Detail[BusinessCheck] = &TaskDetail{
		Status:     TaskRunning,
		DetailList: make(map[string]*powertrading.MachineTestMessage),
	}
	task.Detail[ClusterCheck] = &TaskDetail{
		Status:     TaskWaiting,
		DetailList: make(map[string]*powertrading.MachineTestMessage),
	}
	task.Detail[ImportedCheck] = &TaskDetail{
		Status:     TaskWaiting,
		DetailList: make(map[string]*powertrading.MachineTestMessage),
	}
	task.Detail[MemoryCheck] = &TaskDetail{
		Status:     TaskWaiting,
		DetailList: make(map[string]*powertrading.MachineTestMessage),
	}
	task.Detail[BkOpsTaskCheck] = &TaskDetail{
		Status:     TaskWaiting,
		DetailList: make(map[string]*powertrading.MachineTestMessage),
	}
	task.Summary[MachineCheckSuccess] = make(map[string][]string, 0)
	task.Summary[MachineCheckFailure] = make(map[string][]string, 0)
	task.Summary[MachineNeedClean] = make(map[string][]string, 0)
	return task
}

// MachineSpecification machine detail
type MachineSpecification struct {
	TotalMem   float64
	MemPercent float64
	TotalCPU   float64
	CpuPercent float64
}

// DeviceOperationData device operation data struct
type DeviceOperationData struct {
	DeviceID                string    `json:"deviceID" bson:"deviceID"`
	AssetID                 string    `json:"assetID" bson:"assetID"`
	InnerIP                 string    `json:"innerIP" bson:"innerIP"`
	PoolID                  string    `json:"poolID" bson:"poolID"`
	PoolName                string    `json:"poolName" bson:"poolName"`
	DeviceBusinessID        string    `json:"deviceBusinessID" bson:"deviceBusinessID"`
	RealBusinessID          string    `json:"realBusinessID" bson:"realBusinessID"`
	InstanceType            string    `json:"instanceType" bson:"instanceType"`
	BusinessName            string    `json:"businessName" bson:"businessName"`
	ConsumerID              string    `json:"consumerID" bson:"consumerID"`
	ShouldConsumedClusterID string    `json:"shouldConsumedClusterID" bson:"shouldConsumedClusterID"`
	ShouldConsumedNodeGroup string    `json:"shouldConsumedNodeGroup" bson:"shouldConsumedNodeGroup"`
	Source                  string    `json:"source" bson:"source"`
	RealClusterID           string    `json:"realClusterID" bson:"realClusterID"`
	RealNodeGroup           string    `json:"realNodeGroup" bson:"realNodeGroup"`
	NodeStatus              string    `json:"nodeStatus"`
	DeviceStatus            string    `json:"deviceStatus" bson:"deviceStatus"`
	BusinessCheck           bool      `json:"businessCheck" bson:"businessCheck"`
	ConsumeCheck            bool      `json:"consumeCheck" bson:"consumeCheck"`
	Message                 string    `json:"message" bson:"message"`
	CheckTime               time.Time `json:"checkTime" bson:"checkTime"`
	RecordTime              string    `json:"recordTime" bson:"recordTime"`
}
