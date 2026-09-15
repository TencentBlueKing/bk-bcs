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

package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIgnoredIsTerminalAndSucceeded(t *testing.T) {
	assert.True(t, IsTaskTerminal(TaskStatusIgnored))
	assert.True(t, IsTaskSucceeded(TaskStatusIgnored))
	assert.True(t, IsTaskSucceeded(TaskStatusSuccess))

	assert.False(t, IsTaskSucceeded(TaskStatusFailure))
	assert.False(t, IsTaskSucceeded(TaskStatusTimeout))
	assert.False(t, IsTaskSucceeded(TaskStatusRevoked))
	assert.False(t, IsTaskSucceeded(TaskStatusRunning))

	// IGNORED 视同成功, 不应被当作待下发状态而被重新投递
	assert.NotContains(t, PendingTaskStatus, TaskStatusIgnored)
}

func TestIgnoredStepIsCompleted(t *testing.T) {
	step := NewStep("step1", "hello")
	assert.False(t, step.IsCompleted())

	// 幂等忽略的步骤视同完成, 且不受重试次数影响
	step.SetStatus(TaskStatusIgnored)
	step.MaxRetries = 3
	assert.True(t, step.IsCompleted())
}
