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
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	// ErrServerNotInit server not init
	ErrServerNotInit = errors.New("iam server not init")

	// defaultAllowTTL default cache time
	defaultAllowTTL = time.Minute * 1
)

const (
	// SystemUser systemUser
	SystemUser = "bk_bcs"
	// SystemIDBKBCS systemID
	SystemIDBKBCS = "bk_bcs_app"
	// IamAppURL permission system url
	IamAppURL = ""
)

// BaseResponse base response
type BaseResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// SystemResponse system response
type SystemResponse struct {
	BaseResponse
	Data interface{} `json:"data"`
}

// SystemToken xxx
type SystemToken struct {
	Token string `json:"token"`
}

// TokenResponse xxx
type TokenResponse struct {
	BaseResponse
	Data SystemToken `json:"data"`
}

// System data
type System struct {
	ID                 string     `json:"id,omitempty"`
	Name               string     `json:"name,omitempty"`
	EnglishName        string     `json:"name_en,omitempty"`
	Description        string     `json:"description,omitempty"`
	EnglishDescription string     `json:"description_en,omitempty"`
	Clients            string     `json:"clients,omitempty"`
	ProviderConfig     *SysConfig `json:"provider_config"`
}

// SysConfig system host and auth
type SysConfig struct {
	Host    string `json:"host,omitempty"`
	Auth    string `json:"auth,omitempty"`
	Healthz string `json:"healthz,omitempty"`
}

// IsDifferentConfig check system config
func (config *SysConfig) IsDifferentConfig(origin *SysConfig) bool {
	if strings.EqualFold(config.Host, origin.Host) && strings.EqualFold(config.Auth, origin.Auth) &&
		strings.EqualFold(config.Healthz, origin.Healthz) {
		return false
	}

	return true
}

// SystemResp for systemInfo response
type SystemResp struct {
	BaseResponse
	Data RegisteredSystemInfo `json:"data"`
}

// RegisteredSystemInfo system register info
type RegisteredSystemInfo struct {
	BaseInfo               System                 `json:"base_info"`
	ResourceTypes          []ResourceType         `json:"resource_types"`
	Actions                []ResourceAction       `json:"actions"`
	ActionGroups           []ActionGroup          `json:"action_groups"`
	InstanceSelections     []InstanceSelection    `json:"instance_selections"`
	ResourceCreatorActions ResourceCreatorActions `json:"resource_creator_actions"`
	CommonActions          []CommonAction         `json:"common_actions"`
}

// AuthError xxx
type AuthError struct {
	RequestID string
	Reason    error
}

// Error to string
func (a *AuthError) Error() string {
	if len(a.RequestID) == 0 {
		return a.Reason.Error()
	}

	return fmt.Sprintf("iam request id: %s, err: %s", a.RequestID, a.Reason.Error())
}

// TypeID for ResourceType
type TypeID string

// ResourceType describe resource type defined and registered to iam.
type ResourceType struct {
	ID             TypeID         `json:"id"`
	Name           string         `json:"name"`
	NameEn         string         `json:"name_en"`
	Description    string         `json:"description"`
	DescriptionEn  string         `json:"description_en"`
	Parents        []Parent       `json:"parents"`
	ProviderConfig ResourceConfig `json:"provider_config"`
	Version        int64          `json:"version"`
}

// Parent xxx
type Parent struct {
	SystemID   string `json:"system_id"`
	ResourceID TypeID `json:"id"`
}

// ResourceConfig callback path to get resource
type ResourceConfig struct {
	Path string `json:"path"`
}

// ActionType for register action's type
type ActionType string

const (
	// Create actionType
	Create ActionType = "create"
	// Delete actionType
	Delete ActionType = "delete"
	// View actionType
	View ActionType = "view"
	// Edit actionType
	Edit ActionType = "edit"
	// List actionType
	List ActionType = "list"
)

// ActionID xxx
type ActionID string

// String to string
func (aID ActionID) String() string {
	return string(aID)
}

// ResourceAction action
type ResourceAction struct {
	// must be a unique id in the whole system.
	ID ActionID `json:"id"`
	// must be a unique name in the whole system.
	Name                 string               `json:"name"`
	NameEn               string               `json:"name_en"`
	Type                 ActionType           `json:"type"`
	RelatedResourceTypes []RelateResourceType `json:"related_resource_types"`
	RelatedActions       []ActionID           `json:"related_actions"`
	Version              int                  `json:"version"`
}

// RelateResourceType action related resource
type RelateResourceType struct {
	SystemID           string                     `json:"system_id"`
	ID                 TypeID                     `json:"id"`
	NameAlias          string                     `json:"name_alias"`
	NameAliasEn        string                     `json:"name_alias_en"`
	SelectionMode      string                     `json:"selection_mode"`
	InstanceSelections []RelatedInstanceSelection `json:"related_instance_selections"`
}

// Scope xxx
type Scope struct {
	Op      string         `json:"op"`
	Content []ScopeContent `json:"content"`
}

// ScopeContent xxx
type ScopeContent struct {
	Op    string `json:"op"`
	Field string `json:"field"`
	Value string `json:"value"`
}

// ActionGroup related resource actions to make action selection more organized
type ActionGroup struct {
	// must be a unique name in the whole system.
	Name      string         `json:"name"`
	NameEn    string         `json:"name_en"`
	SubGroups []ActionGroup  `json:"sub_groups,omitempty"`
	Actions   []ActionWithID `json:"actions,omitempty"`
}

// ActionWithID action
type ActionWithID struct {
	ID ActionID `json:"id"`
}

// InstanceSelectionID instance selection
type InstanceSelectionID string

// InstanceSelection instance selection
type InstanceSelection struct {
	ID                InstanceSelectionID `json:"id"`
	Name              string              `json:"name"`
	NameEn            string              `json:"name_en"`
	ResourceTypeChain []ResourceChain     `json:"resource_type_chain"`
}

// ResourceChain resource level
type ResourceChain struct {
	SystemID string `json:"system_id"`
	ID       TypeID `json:"id"`
}

// RelatedInstanceSelection xxx
type RelatedInstanceSelection struct {
	SystemID string              `json:"system_id"`
	ID       InstanceSelectionID `json:"id"`
	// if true, then this selected instance with not be calculated to calculate the auth.
	// as is will be ignored, the only usage for this selection is to support a convenient
	// way for user to find it's resource instances.
	IgnoreAuthPath bool `json:"ignore_iam_path"`
}

// ResourceCreatorActions specifies resource creation actions' related actions that resource creator will have permissions to
type ResourceCreatorActions struct {
	Config []ResourceCreatorAction `json:"config"`
}

// ResourceCreatorAction creator actions
type ResourceCreatorAction struct {
	ResourceID       TypeID                  `json:"id"`
	Actions          []CreatorRelatedAction  `json:"actions"`
	SubResourceTypes []ResourceCreatorAction `json:"sub_resource_types,omitempty"`
}

// CreatorRelatedAction related action
type CreatorRelatedAction struct {
	ID         ActionID `json:"id"`
	IsRequired bool     `json:"required"`
}

// CommonAction specifies a common operation's related iam actions
type CommonAction struct {
	Name        string         `json:"name"`
	EnglishName string         `json:"name_en"`
	Actions     []ActionWithID `json:"actions"`
}
