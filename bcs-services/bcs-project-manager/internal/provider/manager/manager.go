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

// Package manager xxx
package manager

import (
	"sync"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/provider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/provider/internal"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/provider/utils"
)

var (
	providerManagerLock sync.RWMutex
	providerManagerOnce sync.Once
)

var (
	quotaMgrs    map[string]provider.QuotaManager
	capacityMgrs map[string]provider.CapacityManager // nolint
	// taskMgrs     map[string]provider.TaskManager
	validateMgrs map[string]provider.ValidateManager
)

func init() {
	providerManagerOnce.Do(func() {
		quotaMgrs = make(map[string]provider.QuotaManager)
		// taskMgrs = make(map[string]provider.TaskManager)
		capacityMgrs = make(map[string]provider.CapacityManager)
		validateMgrs = make(map[string]provider.ValidateManager)
	})
}

// RegisterQuotaMgrs register all quota managers
func RegisterQuotaMgrs() {
	initQuotaManager(internal.ProviderName, internal.NewQuotaManager())
}

// RegisterValidateMgrs register all validate managers
func RegisterValidateMgrs() {
	initValidateManager(internal.ProviderName, internal.NewValidateManager())
}

// GetQuotaManager for quota manager
func GetQuotaManager(provider string) (provider.QuotaManager, error) {
	providerManagerLock.RLock()
	defer providerManagerLock.RUnlock()

	mgr, ok := quotaMgrs[provider]
	if !ok {
		return nil, utils.NewNoProviderError(provider)
	}
	return mgr, nil
}

// GetValidateManager for validate manager
func GetValidateManager(provider string) (provider.ValidateManager, error) {
	providerManagerLock.RLock()
	defer providerManagerLock.RUnlock()
	mgr, ok := validateMgrs[provider]
	if !ok {
		return nil, utils.NewNoProviderError(provider)
	}
	return mgr, nil
}

// initQuotaManager for quota manager init
func initQuotaManager(provider string, quota provider.QuotaManager) {
	providerManagerLock.Lock()
	defer providerManagerLock.Unlock()
	quotaMgrs[provider] = quota
}

// initValidateManager for validate manager init
func initValidateManager(provider string, validate provider.ValidateManager) {
	providerManagerLock.Lock()
	defer providerManagerLock.Unlock()
	validateMgrs[provider] = validate
}

/*
// RegisterTaskMgrs register all task managers
func RegisterTaskMgrs() {
	initTaskManager(internal.ProviderName, internal.NewTaskManager())
}

// initTaskManager for async task manager init
func initTaskManager(provider string, t provider.TaskManager) {
	providerManagerLock.Lock()
	defer providerManagerLock.Unlock()
	taskMgrs[provider] = t
}

// GetTaskManager for task manager
func GetTaskManager(provider string) (provider.TaskManager, error) {
	providerManagerLock.RLock()
	defer providerManagerLock.RUnlock()
	mgr, ok := taskMgrs[provider]
	if !ok {
		return nil, utils.NewNoProviderError(provider)
	}
	return mgr, nil
}
*/
