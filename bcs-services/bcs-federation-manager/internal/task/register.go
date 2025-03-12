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

package task

import (
	"github.com/RichardKnop/logging"

	"github.com/Tencent/bk-bcs/bcs-common/common/task"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/task/logger"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/task/steps"
)

// RegisterSteps return all steps to be registered
func RegisterSteps() []task.StepWorkerInterface {
	regSteps := make([]task.StepWorkerInterface, 0)
	regSteps = append(regSteps, steps.GetAllSteps()...)

	return regSteps
}

// RegisterCallbacks return all callbacks to be registered
func RegisterCallbacks() []task.CallbackInterface {
	callbacks := make([]task.CallbackInterface, 0)
	callbacks = append(callbacks, steps.GetAllCallbacks()...)
	return callbacks
}

// NewLogger create a logger
func NewLogger() logging.LoggerInterface {
	return logger.NewTaskLogger()
}
