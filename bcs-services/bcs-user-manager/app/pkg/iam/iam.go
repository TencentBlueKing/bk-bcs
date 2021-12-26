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
 *
 */

package iam

import (
	"encoding/base64"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/models"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/storages/sqlstore"
)

const (
	defaultTimeout = time.Second * 10
)

// Iam client for system model register
type Iam struct {
	client *iamModelServer
}

// NewIam init IAM for register permission model
func NewIam(config *AuthConfig) (*Iam, error) {
	iamModelServer := newIamModelServer(config)
	if iamModelServer == nil {
		blog.Errorf("NewIamModelServer failed")
		return nil, fmt.Errorf("NewIamModelServer failed")
	}

	return &Iam{client: iamModelServer}, nil
}

// RegisterSystem register model system
func (i *Iam) RegisterSystem(config *SysConfig) error {
	if i == nil {
		return ErrInitServerFail
	}

	// register system info
	systemInfo, err := i.client.GetSystemInfo(defaultTimeout)
	if err != nil && err != ErrNotFound {
		blog.Errorf("get system info failed; %v", err)
		return err
	}

	// if not registered, will be register system
	if err == ErrNotFound {
		sys := System{
			ID:          SystemIDBKBCS,
			Name:        SystemNameBKBCS,
			EnglishName: SystemNameBKBCSEn,
			Clients:     SystemIDBKBCS,
			ProviderConfig: &SysConfig{
				Host:    config.Host,
				Auth:    config.Auth,
				Healthz: config.Healthz,
			},
		}
		err = i.client.RegisterSystem(defaultTimeout, sys)
		if err != nil {
			blog.Errorf("register system failed: %v", err)
			return err
		}

		blog.Infof("register new system successful: %+v", sys)
	} else if systemInfo.Data.BaseInfo.ProviderConfig == nil || systemInfo.Data.BaseInfo.ProviderConfig.IsDifferentConfig(config) {
		// system no providerConfig or provider is different form origin, update system provider config
		if err = i.client.UpdateSystemConfig(defaultTimeout, config); err != nil {
			blog.Errorf("update system config[%+v] failed; %v", config, err)
			return err
		}

		blog.Infof("update system provider config successful; %+v", config)
	}

	// resource type
	existResourceType := map[TypeID]bool{}
	removeResourceType := map[TypeID]struct{}{}
	for _, resourceType := range systemInfo.Data.ResourceTypes {
		existResourceType[resourceType.ID] = true
		removeResourceType[resourceType.ID] = struct{}{}
	}

	newResourceType := make([]ResourceType, 0)
	for _, resourceType := range GenerateResourceTypes() {
		delete(removeResourceType, resourceType.ID)

		// current resource type registered, update it. else register it
		if existResourceType[resourceType.ID] {
			if err = i.client.UpdateResourceTypes(defaultTimeout, resourceType); err != nil {
				blog.Errorf("update resource type %s failed: %v, input resource[%v]", resourceType.ID, err, resourceType)
			}
		} else {
			newResourceType = append(newResourceType, resourceType)
		}
	}

	if len(newResourceType) > 0 {
		err = i.client.RegisterResourceTypes(defaultTimeout, newResourceType)
		if err != nil {
			blog.Errorf("register resource types[%+v] failed: %v", newResourceType, err)
			return err
		}

		blog.Infof("register new resource types[%v] successful", newResourceType)
	}

	// instance selection
	existInstanceSelection := map[InstanceSelectionID]bool{}
	removeInstanceSelection := map[InstanceSelectionID]struct{}{}
	for _, instanceSelection := range systemInfo.Data.InstanceSelections {
		existInstanceSelection[instanceSelection.ID] = true
		removeInstanceSelection[instanceSelection.ID] = struct{}{}
	}

	newInstanceSelections := make([]InstanceSelection, 0)
	for _, instanceSelection := range GenerateInstanceSelections() {
		delete(removeInstanceSelection, instanceSelection.ID)

		if existInstanceSelection[instanceSelection.ID] {
			err = i.client.UpdateInstanceSelection(defaultTimeout, instanceSelection)
			if err != nil {
				blog.Errorf("update instance selection[%s] failed; %v", instanceSelection.ID, err)
				return err
			}
		} else {
			newInstanceSelections = append(newInstanceSelections, instanceSelection)
		}
	}
	if len(newInstanceSelections) > 0 {
		err = i.client.CreateInstanceSelection(defaultTimeout, newInstanceSelections)
		if err != nil {
			blog.Errorf("register new instance selection[%+v] failed; %v", newInstanceSelections, err)
			return err
		}

		blog.Infof("register new instance selections successful")
	}

	// resource action
	existResourceAction := map[ActionID]bool{}
	removeResourceAction := map[ActionID]struct{}{}
	for _, action := range systemInfo.Data.Actions {
		existResourceAction[action.ID] = true
		removeResourceAction[action.ID] = struct{}{}
	}

	newResourceActions := make([]ResourceAction, 0)
	for _, action := range GenerateActions() {
		delete(removeResourceAction, action.ID)

		if existResourceAction[action.ID] {
			err = i.client.UpdateAction(defaultTimeout, action)
			if err != nil {
				blog.Errorf("update resource action[%s] failed; %v", action.ID, err)
				return err
			}
		} else {
			newResourceActions = append(newResourceActions, action)
		}
	}
	if len(newResourceActions) > 0 {
		err = i.client.CreateAction(defaultTimeout, newResourceActions)
		if err != nil {
			blog.Errorf("register new resource action[%+v] failed; %v", newInstanceSelections, err)
			return err
		}

		blog.Infof("register new resource actions successful")
	}

	// remove needless resource
	if actionLen := len(removeResourceAction); actionLen > 0 {
		removeActionIDs := make([]ActionID, actionLen)
		for actionID := range removeResourceAction {
			removeActionIDs = append(removeActionIDs, actionID)
		}

		err = i.client.DeleteAction(defaultTimeout, removeActionIDs)
		if err != nil {
			blog.Errorf("delete resource actions[%+v] failed: %v", removeActionIDs, err)
			return err
		}

		blog.Infof("remove actionIDs[%+v] successful", removeActionIDs)
	}

	if selectionLen := len(removeInstanceSelection); selectionLen > 0 {
		removeSelectionIDs := make([]InstanceSelectionID, selectionLen)
		for selectionID := range removeInstanceSelection {
			removeSelectionIDs = append(removeSelectionIDs, selectionID)
		}

		err = i.client.DeleteInstanceSelection(defaultTimeout, removeSelectionIDs)
		if err != nil {
			blog.Errorf("delete instance selections[%+v] failed: %v", removeSelectionIDs, err)
			return err
		}

		blog.Infof("remove instance selections[%+v] successful", removeSelectionIDs)
	}

	if resourceLen := len(removeResourceType); resourceLen > 0 {
		removeTypeIDs := make([]TypeID, resourceLen)
		for typeID := range removeResourceType {
			removeTypeIDs = append(removeTypeIDs, typeID)
		}

		err = i.client.DeleteResourceTypes(defaultTimeout, removeTypeIDs)
		if err != nil {
			blog.Errorf("delete resource type[%+v] failed: %v", removeTypeIDs, err)
			return err
		}

		blog.Infof("remove resource typeIDs[%+v] successful", removeTypeIDs)
	}

	// register or update action groups
	actionGroups := GenerateActionGroups()
	if len(systemInfo.Data.ActionGroups) == 0 {
		err = i.client.RegisterActionGroup(defaultTimeout, actionGroups)
		if err != nil {
			blog.Errorf("register action groups[%+v] failed: %v", actionGroups, err)
			return err
		}

		blog.Infof("register action groups[%+v] successful", actionGroups)
	} else {
		err = i.client.UpdateActionGroups(defaultTimeout, actionGroups)
		if err != nil {
			blog.Errorf("update action groups[%+v] failed: %v", actionGroups, err)
			return err
		}

		blog.Infof("update action groups[%+v] successful", actionGroups)
	}

	// register resource creator action or register common actions if need

	// register iam call user-manager permission
	err = i.RegisterUserAuthForIAM()
	if err != nil {
		blog.Errorf("RegisterUserAuthForIAM failed: %v", err)
		return err
	}

	blog.Infof("system %s register successful", SystemIDBKBCS)
	return nil
}

// RegisterUserAuthForIAM register perm model to permissionSystem
func (i *Iam) RegisterUserAuthForIAM() error {
	if i == nil {
		return ErrInitServerFail
	}

	systemToken, err := i.client.GetSystemToken(defaultTimeout)
	if err != nil {
		return err
	}
	// base64 encode
	token := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", SystemIDIAM, systemToken)))

	var userID uint
	user := &models.BcsUser{
		Name: SystemIDIAM,
	}
	userInDb := sqlstore.GetUserByCondition(user)
	// if not exist create it
	if userInDb == nil {
		user.UserType = sqlstore.PlainUser
		user.UserToken = token
		user.ExpiresAt = time.Now().Add(sqlstore.AdminSaasUserExpiredTime)

		err = sqlstore.CreateUser(user)
		if err != nil {
			blog.Errorf("RegisterUserAuthForIAM CreateUser[%v] failed: %v", user, err)
			return err
		}
		time.Sleep(time.Millisecond * 500)

		user := &models.BcsUser{
			Name: SystemIDIAM,
		}
		newUser := sqlstore.GetUserByCondition(user)
		if newUser == nil {
			return fmt.Errorf("RegisterUserAuthForIAM GetUserByCondition failed")
		}

		userID = newUser.ID
	} else {
		userID = userInDb.ID
	}

	// grant permission
	roleInDb := sqlstore.GetRole("manager")
	if roleInDb == nil {
		return fmt.Errorf("RegisterUserAuthForIAM GetRole[%s] failed", "manager")
	}

	userResourceRole := &models.BcsUserResourceRole{
		UserId:       userID,
		ResourceType: "usermanager",
		Resource:     "*",
		RoleId:       roleInDb.ID,
	}
	urrInDb := sqlstore.GetUrrByCondition(userResourceRole)
	if urrInDb != nil {
		blog.Infof("user[%s] exist permission[%v]", userID, urrInDb)
		return nil
	}

	err = sqlstore.CreateUserResourceRole(userResourceRole)
	if err != nil {
		blog.Infof("RegisterUserAuthForIAM CreateUserResourceRole[%v] successful", userResourceRole)
		return err
	}

	return nil
}
