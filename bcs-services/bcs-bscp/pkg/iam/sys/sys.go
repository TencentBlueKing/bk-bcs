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

// Package sys NOTES
package sys

import (
	"context"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/errf"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/iam/client"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
)

// Sys iam system related operate.
type Sys struct {
	client *client.Client
}

// NewSys create sys to iam sys related operate.
func NewSys(client *client.Client) (*Sys, error) {
	if client == nil {
		return nil, errf.New(errf.InvalidParameter, "client is nil")
	}

	sys := &Sys{
		client: client,
	}
	return sys, nil
}

// GetSystemToken get system token from iam, used to validate if request is from iam.
func (s *Sys) GetSystemToken(ctx context.Context) (string, error) {
	return s.client.GetSystemToken(ctx)
}

// Register register auth model to iam.
func (s *Sys) Register(ctx context.Context, host string) error {
	system, err := s.registerSystem(ctx, host)
	if err != nil {
		return err
	}

	// Note: 如果更新的资源依赖新增的其他资源，会存在问题。如更新的实例视图，依赖新增的资源类型。
	removedResTypeMap, newResTypes, err := s.classResType(ctx, system.ResourceTypes)
	if err != nil {
		return err
	}

	removedInstSelectionMap, newInstSelections, err := s.classInstSelection(ctx, system.InstanceSelections)
	if err != nil {
		return err
	}

	removedResActionMap, newResActions, err := s.classAction(ctx, system.Actions)
	if err != nil {
		return err
	}

	// 因为资源间的依赖关系，删除和更新的顺序为 1.Action 2.InstanceSelection 3.ResourceType
	// 因为资源间的依赖关系，新建的顺序则反过来为 1.ResourceType 2.InstanceSelection 3.Action
	// ActionGroup依赖于Action，该资源的增删操作始终放在最后
	// 先删除资源，再新增资源，因为实例视图的名称在系统中是唯一的，如果不先删，同样名称的实例视图将创建失败

	if err = s.deleteRemovedModel(ctx, removedResActionMap, removedInstSelectionMap, removedResTypeMap); err != nil {
		return err
	}

	if err := s.registerNewModel(ctx, newResTypes, newInstSelections, newResActions); err != nil {
		return err
	}

	if err := s.registerActionGroup(ctx, system.ActionGroups); err != nil {
		return err
	}

	if err := s.registerResCreatorAction(ctx, system.ResourceCreatorActions); err != nil {
		return err
	}

	if err := s.registerCommonAction(ctx, system.CommonActions); err != nil {
		return err
	}

	return nil
}

// registerCommonAction register or update common actions
func (s *Sys) registerCommonAction(ctx context.Context, actions []client.CommonAction) error {
	commonActions := GenerateCommonActions()
	if len(actions) == 0 {
		if len(commonActions) != 0 {
			if err := s.client.RegisterCommonActions(ctx, commonActions); err != nil {
				logs.Errorf("register common actions failed, common actions: %v, err: %v", commonActions, err)
				return err
			}
		}
	} else {
		if err := s.client.UpdateCommonActions(ctx, commonActions); err != nil {
			logs.Errorf("update common actions failed, common actions: %v, err: %v", commonActions, err)
			return err
		}
	}

	return nil
}

// registerResCreatorAction register or update resource creator actions.
func (s *Sys) registerResCreatorAction(ctx context.Context, actions client.ResourceCreatorActions) error {
	resourceCreatorActions := GenerateResourceCreatorActions()
	if len(actions.Config) == 0 {
		if len(resourceCreatorActions.Config) != 0 {
			if err := s.client.RegisterResourceCreatorActions(ctx, resourceCreatorActions); err != nil {
				logs.Errorf("register resource creator actions failed, resource creator actions: %v, err: %v",
					resourceCreatorActions, err)
				return err
			}
		}
	} else {
		if err := s.client.UpdateResourceCreatorActions(ctx, resourceCreatorActions); err != nil {
			logs.Errorf("update resource creator actions failed, resource creator actions: %v, err: %v",
				resourceCreatorActions, err)
			return err
		}
	}

	return nil
}

// registerActionGroup register or update action group.
func (s *Sys) registerActionGroup(ctx context.Context, groups []client.ActionGroup) error {
	actionGroups := GenerateStaticActionGroups()
	if len(actionGroups) != 0 {
		if len(groups) == 0 {
			if err := s.client.RegisterActionGroups(ctx, actionGroups); err != nil {
				logs.Errorf("register action groups failed, action groups: %v, err: %v", actionGroups, err)
				return err
			}

		} else {
			if err := s.client.UpdateActionGroups(ctx, actionGroups); err != nil {
				logs.Errorf("update action groups failed, action groups: %v, err: %v", actionGroups, err)
				return err
			}
		}
	}

	return nil
}

// registerNewModel register new auth model.
func (s *Sys) registerNewModel(ctx context.Context, newResTypes []client.ResourceType,
	newInstSelections []client.InstanceSelection, newResActions []client.ResourceAction) error {
	if len(newResTypes) > 0 {
		if err := s.client.RegisterResourcesTypes(ctx, newResTypes); err != nil {
			logs.Errorf("register resource types failed, types: %v, err: %v", newResTypes, err)
			return err
		}
	}

	if len(newInstSelections) > 0 {
		if err := s.client.RegisterInstanceSelections(ctx, newInstSelections); err != nil {
			logs.Errorf("register instance selections failed, types: %v, err: %v", newInstSelections, err)
			return err
		}
	}

	if len(newResActions) > 0 {
		if err := s.client.RegisterActions(ctx, newResActions); err != nil {
			logs.Errorf("register resource actions failed, actions: %v, err: %v", newResActions, err)
			return err
		}
	}

	return nil
}

// deleteRemovedModel delete removed auth model.
func (s *Sys) deleteRemovedModel(ctx context.Context, resActions map[client.ActionID]struct{},
	selections map[client.InstanceSelectionID]struct{}, resTypes map[client.TypeID]struct{}) error {

	if len(resActions) > 0 {
		removedResourceActionIDs := make([]client.ActionID, len(resActions))
		idx := 0
		// before deleting action, the dependent action policies must be deleted
		for resourceActionID := range resActions {
			if err := s.client.DeleteActionPolicies(ctx, resourceActionID); err != nil {
				logs.Errorf("delete action policies failed, id: %s, err: %v", resourceActionID, err)
				return err
			}

			removedResourceActionIDs[idx] = resourceActionID
			idx++
		}
		if err := s.client.DeleteActions(ctx, removedResourceActionIDs); err != nil {
			logs.Errorf("delete resource actions failed, actions: %v, err: %v", removedResourceActionIDs, err)
			return err
		}
	}

	if len(selections) > 0 {
		removedInstanceSelectionIDs := make([]client.InstanceSelectionID, len(selections))
		idx := 0
		for resourceActionID := range selections {
			removedInstanceSelectionIDs[idx] = resourceActionID
			idx++
		}
		if err := s.client.DeleteInstanceSelections(ctx, removedInstanceSelectionIDs); err != nil {
			logs.Errorf("delete instance selections failed, selections: %v, err: %v",
				removedInstanceSelectionIDs, err)
			return err
		}
	}

	if len(resTypes) > 0 {
		removedResourceTypeIDs := make([]client.TypeID, len(resTypes))
		idx := 0
		for resourceType := range resTypes {
			removedResourceTypeIDs[idx] = resourceType
			idx++
		}
		if err := s.client.DeleteResourcesTypes(ctx, removedResourceTypeIDs); err != nil {
			logs.Errorf("delete resource types failed, types: %v, err: %v", removedResourceTypeIDs, err)
			return err
		}
	}

	return nil
}

// classAction class action to removed and new create.
func (s *Sys) classAction(ctx context.Context, actions []client.ResourceAction) (
	map[client.ActionID]struct{}, []client.ResourceAction, error) {

	old := make(map[client.ActionID]bool)
	removed := make(map[client.ActionID]struct{})
	news := make([]client.ResourceAction, 0)
	for _, resourceAction := range actions {
		old[resourceAction.ID] = true
		removed[resourceAction.ID] = struct{}{}
	}

	for _, resourceAction := range GenerateStaticActions() {
		// registered resource action exist in current resource actions, should not be removed
		delete(removed, resourceAction.ID)
		// if current resource action is registered, update it, or else register it
		if old[resourceAction.ID] {
			if err := s.client.UpdateAction(ctx, resourceAction); err != nil {
				logs.Errorf("update resource action failed, id: %s, action: %v, err: %v",
					resourceAction.ID, resourceAction, err)
				return nil, nil, err
			}
		} else {
			news = append(news, resourceAction)
		}
	}

	return removed, news, nil
}

// classInstSelection class instance selection to removed and new create.
func (s *Sys) classInstSelection(ctx context.Context, selections []client.InstanceSelection) (
	map[client.InstanceSelectionID]struct{}, []client.InstanceSelection, error) {

	old := make(map[client.InstanceSelectionID]bool)
	removed := make(map[client.InstanceSelectionID]struct{})
	news := make([]client.InstanceSelection, 0)
	for _, instanceSelection := range selections {
		old[instanceSelection.ID] = true
		removed[instanceSelection.ID] = struct{}{}
	}

	for _, resourceType := range GenerateStaticInstanceSelections() {
		// registered instance selection removed in current instance selections, should not be removed
		delete(removed, resourceType.ID)
		// if current instance selection is registered, update it, or else register it
		if old[resourceType.ID] {
			if err := s.client.UpdateInstanceSelection(ctx, resourceType); err != nil {
				logs.Errorf("update instance selection failed, id: %s, type: %v, err: %v",
					resourceType.ID, resourceType, err)
				return nil, nil, err
			}
		} else {
			news = append(news, resourceType)
		}
	}

	return removed, news, nil
}

// classResType class resource type to removed and new create.
func (s *Sys) classResType(ctx context.Context, resTypes []client.ResourceType) (
	map[client.TypeID]struct{}, []client.ResourceType, error) {

	old := make(map[client.TypeID]bool)
	removed := make(map[client.TypeID]struct{})
	news := make([]client.ResourceType, 0)
	for _, resourceType := range resTypes {
		old[resourceType.ID] = true
		removed[resourceType.ID] = struct{}{}
	}

	for _, resourceType := range GenerateStaticResourceTypes() {
		// registered resource type exist in current resource types, should not be removed
		delete(removed, resourceType.ID)
		// if current resource type is registered, update it, or else register it
		if old[resourceType.ID] {
			if err := s.client.UpdateResourcesType(ctx, resourceType); err != nil {
				logs.Errorf("update resource type failed, id: %s, type: %v, err: %v",
					resourceType.ID, resourceType, err)
				return nil, nil, err
			}
		} else {
			news = append(news, resourceType)
		}
	}

	return removed, news, nil
}

// registerSystem register or update system to iam.
func (s *Sys) registerSystem(ctx context.Context, host string) (*client.RegisteredSystemInfo, error) {
	resp, err := s.client.GetSystemInfo(ctx, []client.SystemQueryField{})
	if err != nil && err != client.ErrNotFound {
		logs.Errorf("get system info failed, err: %v", err)
		return nil, err
	}

	// if iam bscp system has not been registered, register system
	if err == client.ErrNotFound {
		sys := client.System{
			ID:          SystemIDBSCP,
			Name:        SystemNameBSCP,
			EnglishName: SystemNameBSCPEn,
			Clients:     SystemIDBSCP,
			ProviderConfig: &client.SysConfig{
				Host: host,
				Auth: "basic",
			},
		}

		if err = s.client.RegisterSystem(ctx, sys); err != nil {
			logs.Errorf("register system failed, system: %v, err: %v", sys, err)
			return nil, err
		}

		if logs.V(5) {
			logs.Infof("register new system succeed, system: %v", sys)
		}

	} else if resp.Data.BaseInfo.ProviderConfig == nil || resp.Data.BaseInfo.ProviderConfig.Host != host {
		// if iam registered bscp system has no ProviderConfig
		// or registered host config is different with current host config, update system host config
		if err = s.client.UpdateSystemConfig(ctx, &client.SysConfig{Host: host}); err != nil {
			logs.Errorf("update system host config failed, host: %s, err: %v", host, err)
			return nil, err
		}

		if resp.Data.BaseInfo.ProviderConfig == nil {
			if logs.V(5) {
				logs.Infof("update system host succeed, new: %s", host)
			}

		} else {
			if logs.V(5) {
				logs.Infof("update system host succeed, old: %s, new: %s", resp.Data.BaseInfo.
					ProviderConfig.Host, host)
			}

		}
	}

	return &resp.Data, nil
}
