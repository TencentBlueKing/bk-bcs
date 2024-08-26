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

package v1

import (
	"fmt"
	"hash/fnv"
	"math/rand"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
)

const (
	// WorkflowFinalizer defines the finalizer for workflow object
	WorkflowFinalizer = "finalizer.gitopsworkflow.bkbcs.tencent.com"
	// WorkflowLabelProject defines the workflow's project
	WorkflowLabelProject = "gitopsworkflow.bkbcs.tencent.com/project"
	// WorkflowAnnotationCreateUser defines the user who created
	WorkflowAnnotationCreateUser = "gitopsworkflow.bkbcs.tencent.com/create-user"
	// WorkflowAnnotationUpdateUser defines the user who updated
	WorkflowAnnotationUpdateUser = "gitopsworkflow.bkbcs.tencent.com/update-user"

	// HistoryLabelWorkflow defines the workflow name label
	HistoryLabelWorkflow = "gitopsworkflow.bkbcs.tencent.com/workflow"
	// HistoryAnnotationWorkflow defines the parent workflow of history
	HistoryAnnotationWorkflow = "gitopsworkflow.bkbcs.tencent.com/workflow"

	// EngineBKDevOps defines the engine of blueking devops
	EngineBKDevOps = "bkdevops"
)

// SecretName defines the secret suffix
func SecretName(prefix string, str string) string {
	h := fnv.New32a()
	_, err := h.Write([]byte(str))
	if err != nil {
		blog.Warnf("write secret failed: %s", err.Error())
	}
	return fmt.Sprintf("%s-%v", prefix, h.Sum32())
}

// RandomNum return the random num
// nolint
func RandomNum() int64 {
	return rand.New(rand.NewSource(time.Now().UnixNano())).Int63n(1000000)
}
