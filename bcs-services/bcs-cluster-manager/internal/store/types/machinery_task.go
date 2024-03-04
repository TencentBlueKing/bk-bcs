/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package types

import "time"

const (
	// Pending state server insert task
	Pending = "PENDING"
	// Received state worker receive task
	Received = "RECEIVED"
	// Started state worker exec task
	Started = "STARTED"
	// Success state task success
	Success = "SUCCESS"
	// Failure state task failed
	Failure = "FAILURE"

	// FieldTaskName field task_name
	FieldTaskName = "task_name"
)

// TaskState machinery task state list
var TaskState = []string{Pending, Received, Started}

type Task struct {
	Id        string    `json:"id" bson:"_id"`
	CreatedAt time.Time `json:"createdAt" bson:"created_at"`
	State     string    `json:"state" bson:"state"`
	TaskName  string    `json:"taskName" bson:"task_name"`
	Results   []string  `json:"results" bson:"results"`
}
