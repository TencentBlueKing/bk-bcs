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

package cases

import (
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
)

// ResourceManager struct of resource manager, which is to save created data ids.
// Its purpose is to conveniently obtain the ID of the resource created earlier.
type ResourceManager struct {
	// App key is app mode, value is []uint32 of app id
	App map[table.AppMode][]uint32
	// Hook key is app id, value is []uint32 of hook id
	Hook map[uint32][]uint32
	// ConfigItem key is app id, value is []uint32 of config item id
	ConfigItem map[uint32][]uint32
	// Content key is config item id, value is content id
	Content map[uint32]uint32
	// Commit key is content id, value is commit id
	Commit map[uint32]uint32
	// Release key is app id, value is release id
	Release map[uint32]uint32
	// StrategySet key is app id, value is strategy set id
	StrategySet map[uint32]uint32
	// Strategies key is strategy set id, value is strategy id
	Strategies map[uint32][]uint32
	// AppToStrategy key is app id, value is strategy id
	AppToStrategy map[uint32][]uint32
	// Publish key is app id, value is publish strategy id
	Publish map[uint32][]uint32
	// Instance key is app id, value is publish instance id
	Instance map[uint32][]uint32
}

// NewResourceManager initial resource manager
func NewResourceManager() *ResourceManager {
	rm := &ResourceManager{}
	rm.App = make(map[table.AppMode][]uint32)
	rm.App[table.Normal] = make([]uint32, 0)
	rm.App[table.Namespace] = make([]uint32, 0)
	rm.Hook = make(map[uint32][]uint32)
	rm.ConfigItem = make(map[uint32][]uint32)
	rm.Content = make(map[uint32]uint32)
	rm.Commit = make(map[uint32]uint32)
	rm.Release = make(map[uint32]uint32)
	rm.StrategySet = make(map[uint32]uint32)
	rm.Strategies = make(map[uint32][]uint32)
	rm.AppToStrategy = make(map[uint32][]uint32)
	rm.Publish = make(map[uint32][]uint32)
	rm.Instance = make(map[uint32][]uint32)
	return rm
}

// AddApp add a created app id
func (rm *ResourceManager) AddApp(mode table.AppMode, appId uint32) {
	rm.App[mode] = append(rm.App[mode], appId)
}

// GetApp get an app resource id
func (rm *ResourceManager) GetApp(mode table.AppMode) uint32 {
	return rm.App[mode][0]
}

// DeleteApp delete an app resource id
func (rm *ResourceManager) DeleteApp(mode table.AppMode, id uint32) {
	rm.App[mode] = deleteId(rm.App[mode], id)
}

// AddHook add a created hook id
func (rm *ResourceManager) AddHook(appId, hookId uint32) {
	if _, ok := rm.Hook[appId]; !ok {
		rm.Hook[appId] = make([]uint32, 0)
	}
	rm.Hook[appId] = append(rm.Hook[appId], hookId)
}

// GetHook get a created hook id
func (rm *ResourceManager) GetHook(appId uint32) uint32 {
	if len(rm.Hook[appId]) == 0 {
		return 0
	}
	return rm.Hook[appId][0]
}

// DeleteHook delete a created hook id
func (rm *ResourceManager) DeleteHook(appId uint32, ciId uint32) {
	rm.Hook[appId] = deleteId(rm.Hook[appId], ciId)
}

// AddConfigItem add a created config item id
func (rm *ResourceManager) AddConfigItem(appId, configItemId uint32) {
	if _, ok := rm.ConfigItem[appId]; !ok {
		rm.ConfigItem[appId] = make([]uint32, 0)
	}
	rm.ConfigItem[appId] = append(rm.ConfigItem[appId], configItemId)
}

// GetConfigItem get a created config item id
func (rm *ResourceManager) GetConfigItem(appId uint32) uint32 {
	if len(rm.ConfigItem[appId]) == 0 {
		return 0
	}
	return rm.ConfigItem[appId][0]
}

// DeleteConfigItem delete a created config item id
func (rm *ResourceManager) DeleteConfigItem(appId uint32, ciId uint32) {
	rm.ConfigItem[appId] = deleteId(rm.ConfigItem[appId], ciId)
}

// AddContent add a created content id
func (rm *ResourceManager) AddContent(configItemId, contentId uint32) {
	rm.Content[configItemId] = contentId
}

// GetContent get a created content id
func (rm *ResourceManager) GetContent(configItemId uint32) uint32 {
	value, ok := rm.Content[configItemId]
	if ok {
		return value
	}

	return 0
}

// AddCommit add a created commit id
func (rm *ResourceManager) AddCommit(contentId, commitId uint32) {
	rm.Commit[contentId] = commitId
}

// GetCommit get a created commit id
func (rm *ResourceManager) GetCommit(contentId uint32) uint32 {
	value, ok := rm.Commit[contentId]
	if ok {
		return value
	}

	return 0
}

// AddRelease add a created release id
func (rm *ResourceManager) AddRelease(appId, releaseId uint32) {
	rm.Release[appId] = releaseId
}

// GetRelease  get a created release id
func (rm *ResourceManager) GetRelease(appId uint32) uint32 {
	value, ok := rm.Release[appId]
	if ok {
		return value
	}

	return 0
}

// GetAppToRelease  get an app id and release id, when app has release
func (rm *ResourceManager) GetAppToRelease() (appId, relId uint32) {
	if len(rm.Release) == 0 {
		return 0, 0
	}
	for key := range rm.Release {
		appId = key
		relId = rm.Release[key]
		break
	}
	return appId, relId
}

// AddStrategySet  add a created strategy set id
func (rm *ResourceManager) AddStrategySet(appId, stgSetId uint32) {
	rm.StrategySet[appId] = stgSetId
}

// GetStrategySet  get a created strategy set id
func (rm *ResourceManager) GetStrategySet(appId uint32) uint32 {
	value, ok := rm.StrategySet[appId]
	if ok {
		return value
	}

	return 0
}

// GetAppToStrategySet  get an app id and strategy set id, when app has strategy set
func (rm *ResourceManager) GetAppToStrategySet() (appId, stgSetId uint32) {
	if len(rm.StrategySet) == 0 {
		return 0, 0
	}
	for key := range rm.StrategySet {
		appId = key
		stgSetId = rm.StrategySet[key]
		break
	}

	return appId, stgSetId
}

// DeleteStrategySet delete a created strategy set id
func (rm *ResourceManager) DeleteStrategySet(appId uint32) {
	delete(rm.StrategySet, appId)
}

// AddStrategy  add a created strategy id
func (rm *ResourceManager) AddStrategy(appId, stgSetId, stgId uint32) {
	rm.Strategies[stgSetId] = append(rm.Strategies[stgSetId], stgId)
	rm.AppToStrategy[appId] = append(rm.AppToStrategy[appId], stgId)
}

// GetStrategy get a created strategy id
func (rm *ResourceManager) GetStrategy(stgSetId uint32) uint32 {
	if len(rm.Strategies[stgSetId]) == 0 {
		return 0
	}
	return rm.Strategies[stgSetId][0]
}

// DeleteStrategy delete a created strategy id
func (rm *ResourceManager) DeleteStrategy(appId, stgSetId, stgId uint32) {
	rm.Strategies[stgSetId] = deleteId(rm.Strategies[stgSetId], stgId)
	rm.AppToStrategy[appId] = deleteId(rm.AppToStrategy[appId], stgId)
}

// GetAppToStrategy get an app id and stg id, when app has strategy
func (rm *ResourceManager) GetAppToStrategy() (appId, stgId uint32) {
	if len(rm.AppToStrategy) == 0 {
		return
	}
	for key := range rm.AppToStrategy {
		// find an app which has strategies
		if len(rm.AppToStrategy[key]) != 0 {
			return key, rm.AppToStrategy[key][0]
		}
	}

	return appId, stgId
}

// AddPublish  add a created publish strategy id
func (rm *ResourceManager) AddPublish(appId, publishId uint32) {
	rm.Publish[appId] = append(rm.Publish[appId], publishId)
}

// GetPublish  get a created publish id
func (rm *ResourceManager) GetPublish(appId uint32) uint32 {
	if len(rm.Publish[appId]) == 0 {
		return 0
	}
	return rm.Publish[appId][0]
}

// AddInstance  add a created publish instance id
func (rm *ResourceManager) AddInstance(appId, publishId uint32) {
	rm.Instance[appId] = append(rm.Instance[appId], publishId)
}

// GetInstance  get a created publish id
func (rm *ResourceManager) GetInstance(appId uint32) uint32 {
	if len(rm.Instance[appId]) == 0 {
		return 0
	}
	return rm.Instance[appId][0]
}

// deleteId delete the id specified in the id slice
func deleteId(ids []uint32, idToDelete uint32) []uint32 {
	if len(ids) == 0 {
		return ids
	}
	for index, id := range ids {
		if id != idToDelete {
			continue
		}
		if index == len(ids)-1 {
			return ids[:index]
		}
		return append(ids[:index], ids[index+1:]...)
	}
	return ids
}
