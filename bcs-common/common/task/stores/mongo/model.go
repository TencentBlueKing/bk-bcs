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

package mongo

import "time"

// Task mongodb task model
type Task struct {
	TaskIndex           string            `bson:"taskIndex"`
	TaskIndexType       string            `bson:"taskIndexType"`
	TaskID              string            `bson:"taskID"`
	TaskType            string            `bson:"taskType"`
	TaskName            string            `bson:"taskName"`
	CurrentStep         string            `bson:"currentStep"`
	Steps               []*Step           `bson:"steps"`
	CallbackName        string            `bson:"callbackName"`
	CallbackResult      string            `bson:"callbackResult"`
	CallbackMessage     string            `bson:"callbackMessage"`
	CommonParams        map[string]string `bson:"commonParams"`
	CommonPayload       string            `bson:"commonPayload"`
	Status              string            `bson:"status"`
	Message             string            `bson:"message"`
	ExecutionTime       uint32            `bson:"executionTime"`
	MaxExecutionSeconds uint32            `bson:"maxExecutionSeconds"`
	Creator             string            `bson:"creator"`
	Updater             string            `bson:"updater"`
	Start               time.Time         `bson:"start"`
	End                 time.Time         `bson:"end"`
	CreatedAt           time.Time         `bson:"createdAt"`
	LastUpdate          time.Time         `bson:"lastUpdate"`
}

// Step mongodb step model
type Step struct {
	Name                string            `bson:"name"`
	Alias               string            `bson:"alias"`
	Executor            string            `bson:"executor"`
	Params              map[string]string `bson:"params"`
	Payload             string            `bson:"payload"`
	Status              string            `bson:"status"`
	Message             string            `bson:"message"`
	ETA                 *time.Time        `bson:"eta"` // 延迟执行时间(Estimated Time of Arrival)
	SkipOnFailed        bool              `bson:"skipOnFailed"`
	RetryCount          uint32            `bson:"retryCount"`
	MaxRetries          uint32            `bson:"maxRetries"`
	ExecutionTime       uint32            `bson:"executionTime"`
	MaxExecutionSeconds uint32            `bson:"maxExecutionSeconds"`
	Start               time.Time         `bson:"start"`
	End                 time.Time         `bson:"end"`
	LastUpdate          time.Time         `bson:"lastUpdate"`
}
