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

package client

import (
	"errors"
	"fmt"
	"sync"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/iam/sdk/operator"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/tools"
)

const (
	// RequestIDHeader iam rid header key.
	RequestIDHeader = "X-Request-Id"
	bkapiAuthHeader = "X-Bkapi-Authorization"

	// BkIAMMaxPageSize blueking iam max page size.
	BkIAMMaxPageSize = 1000

	// SystemIDIAM iam system id.
	SystemIDIAM = "bk_iam"

	// IamPathKey is the key to describe the auth path that this resource need to auth.
	// only if the path is matched one of the use's auth policy, then a use's
	// have this resource's operate authorize.
	IamPathKey = "_bk_iam_path_"
	// IamIDKey defines the iam id key
	IamIDKey = "id"
)

// Config auth center config.
type Config struct {
	// blueking's auth center addresses
	Address []string
	// app code is used for authorize used.
	AppCode string
	// app secret is used for authorized
	AppSecret string
	// the system id that bscp used in auth center.
	// default value: bk-bscp
	SystemID string
	// TLS is http TLS config
	TLS *tools.TLSConfig
}

// iamDiscovery used to iam discovery.
type iamDiscovery struct {
	servers []string
	index   int
	sync.Mutex
}

// GetServers get iam server host.
func (s *iamDiscovery) GetServers() ([]string, error) {
	s.Lock()
	defer s.Unlock()

	num := len(s.servers)
	if num == 0 {
		return []string{}, errors.New("oops, there is no server can be used")
	}

	if s.index < num-1 {
		s.index++
		return append(s.servers[s.index-1:], s.servers[:s.index-1]...), nil
	}

	s.index = 0
	return append(s.servers[num-1:], s.servers[:num-1]...), nil
}

// AuthError is auth error.
type AuthError struct {
	RequestID string
	Reason    error
}

// Error return auth error string.
func (a *AuthError) Error() string {
	if len(a.RequestID) == 0 {
		return a.Reason.Error()
	}
	return fmt.Sprintf("iam request id: %s, err: %s", a.RequestID, a.Reason.Error())
}

// BaseResponse is http base response.
type BaseResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// System system api param.
type System struct {
	ID                 string     `json:"id,omitempty"`
	Name               string     `json:"name,omitempty"`
	EnglishName        string     `json:"name_en,omitempty"`
	Description        string     `json:"description,omitempty"`
	EnglishDescription string     `json:"description_en,omitempty"`
	Clients            string     `json:"clients,omitempty"`
	ProviderConfig     *SysConfig `json:"provider_config"`
}

// SystemQueryField is system query field for searching system info.
type SystemQueryField string

// SystemQueryField is system query field for searching system info.
const (
	FieldBaseInfo               SystemQueryField = "base_info"
	FieldResourceTypes          SystemQueryField = "resource_types"
	FieldActions                SystemQueryField = "actions"
	FieldActionGroups           SystemQueryField = "action_groups"
	FieldInstanceSelections     SystemQueryField = "instance_selections"
	FieldResourceCreatorActions SystemQueryField = "resource_creator_actions"
	FieldCommonActions          SystemQueryField = "common_actions"
)

// SysConfig is update system config request param.
type SysConfig struct {
	Host string `json:"host,omitempty"`
	Auth string `json:"auth,omitempty"`
}

// SystemResp is get system info response param.
type SystemResp struct {
	BaseResponse
	Data RegisteredSystemInfo `json:"data"`
}

// RegisteredSystemInfo is get system info response detail.
type RegisteredSystemInfo struct {
	BaseInfo               System                 `json:"base_info"`
	ResourceTypes          []ResourceType         `json:"resource_types"`
	Actions                []ResourceAction       `json:"actions"`
	ActionGroups           []ActionGroup          `json:"action_groups"`
	InstanceSelections     []InstanceSelection    `json:"instance_selections"`
	ResourceCreatorActions ResourceCreatorActions `json:"resource_creator_actions"`
	CommonActions          []CommonAction         `json:"common_actions"`
}

// TypeID is iam resource type.
type TypeID string

// ResourceType describe resource type defined and registered to iam.
type ResourceType struct {
	// unique id
	ID TypeID `json:"id"`
	// unique name
	Name   string `json:"name"`
	NameEn string `json:"name_en"`
	// unique description
	Description    string         `json:"description"`
	DescriptionEn  string         `json:"description_en"`
	Parents        []Parent       `json:"parents"`
	ProviderConfig ResourceConfig `json:"provider_config"`
	Version        int64          `json:"version"`
}

// ResourceConfig is the permission center calls the configuration file of the query resource instance interface.
type ResourceConfig struct {
	// the url to get this resource.
	Path string `json:"path"`
}

// Parent is the direct superior of the resource type, which can have multiple direct superiors,
// can be the resource type of one's own system or the resource type of other system, can be an
// empty list, does not allow repetition, and the data is only used for display on the permission center product.
type Parent struct {
	// only one value for bscp.
	// default value: bk-bscp
	SystemID   string `json:"system_id"`
	ResourceID TypeID `json:"id"`
}

// ActionType is iam action type.
type ActionType string

// ActionID is iam action id.
type ActionID string

// ResourceAction iam action related request param.
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

// SelectionMode 选择类型, 资源在权限中心产品上配置权限时的作用范围
type SelectionMode string

const (
	// modeInstance 仅可选择实例, 默认值
	modeInstance SelectionMode = "instance" //nolint:unused
	// modeAttribute 仅可配置属性, 此时instance_selections配置不生效
	modeAttribute SelectionMode = "attribute" //nolint:unused
	// modeAll 可以同时选择实例和配置属性
	modeAll SelectionMode = "all" //nolint:unused
)

// RelateResourceType the order of operating objects, resource type list and list must be
// consistent with the order of product display and authentication verification. If the
// operation does not need to be associated with a resource instance, it can be empty
// here. Note that this is an orderly list!
type RelateResourceType struct {
	SystemID           string                     `json:"system_id"`
	ID                 TypeID                     `json:"id"`
	NameAlias          string                     `json:"name_alias"`
	NameAliasEn        string                     `json:"name_alias_en"`
	Scope              *Scope                     `json:"scope"`
	SelectionMode      SelectionMode              `json:"selection_mode"`
	InstanceSelections []RelatedInstanceSelection `json:"related_instance_selections"`
}

// Scope is optional to limit the operation's selection of the resource
type Scope struct {
	Op      string         `json:"op"`
	Content []ScopeContent `json:"content"`
}

// ScopeContent is scope strategy content.
type ScopeContent struct {
	Op    string `json:"op"`
	Field string `json:"field"`
	Value string `json:"value"`
}

// RelatedInstanceSelection is the associated instance view, that is, the selection of resources
// when configuring permissions on the permission Center product; you can configure the instance
// view of this system, or you can configure the instance view of other systems
type RelatedInstanceSelection struct {
	ID       InstanceSelectionID `json:"id"`
	SystemID string              `json:"system_id"`
	// if true, then this selected instance with not be calculated to calculate the auth.
	// as is will be ignored, the only usage for this selection is to support a convenient
	// way for user to find it's resource instances.
	IgnoreAuthPath bool `json:"ignore_iam_path"`
}

// ActionGroup groups related resource actions to make action selection more organized
type ActionGroup struct {
	// must be a unique name in the whole system.
	Name      string         `json:"name"`
	NameEn    string         `json:"name_en"`
	SubGroups []ActionGroup  `json:"sub_groups,omitempty"`
	Actions   []ActionWithID `json:"actions,omitempty"`
}

// ActionWithID is action id.
type ActionWithID struct {
	ID ActionID `json:"id"`
}

// InstanceSelectionID is iam instance selection id.
type InstanceSelectionID string

// InstanceSelection is instance selection request param.
type InstanceSelection struct {
	// unique
	ID InstanceSelectionID `json:"id"`
	// unique
	Name string `json:"name"`
	// unique
	NameEn            string          `json:"name_en"`
	ResourceTypeChain []ResourceChain `json:"resource_type_chain"`
}

// ResourceChain is resource type info.
type ResourceChain struct {
	SystemID string `json:"system_id"`
	ID       TypeID `json:"id"`
}

// RscTypeAndID resource type with id, used to represent resource layer from root to leaf
type RscTypeAndID struct {
	ResourceType TypeID `json:"resource_type"`
	ResourceID   string `json:"resource_id,omitempty"`
}

// Resource iam resource, system is resource's iam system id, type is resource type,
// resource id and attribute are used for filtering
type Resource struct {
	System    string                 `json:"system"`
	Type      TypeID                 `json:"type"`
	ID        string                 `json:"id,omitempty"`
	Attribute map[string]interface{} `json:"attribute,omitempty"`
}

// ResourceCreatorActions is specifies resource creation actions' related actions,
// that resource creator will have permissions.
type ResourceCreatorActions struct {
	Config []ResourceCreatorAction `json:"config"`
}

// ResourceCreatorAction is action corresponding to the source type that can be
// authorized to the creator at the time of creation
type ResourceCreatorAction struct {
	ResourceID       TypeID                  `json:"id"`
	Actions          []CreatorRelatedAction  `json:"actions"`
	SubResourceTypes []ResourceCreatorAction `json:"sub_resource_types,omitempty"`
}

// CreatorRelatedAction is list of Action that can authorize the creator when the corresponding resource is created
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

// ListPoliciesParams list iam policies parameter
type ListPoliciesParams struct {
	ActionID  ActionID
	Page      int64
	PageSize  int64
	Timestamp int64
}

// ListPoliciesResp list iam policies response
type ListPoliciesResp struct {
	BaseResponse
	Data *ListPoliciesData `json:"data"`
}

// ListPoliciesData list policy data, which represents iam policies
type ListPoliciesData struct {
	Metadata PolicyMetadata `json:"metadata"`
	Count    int64          `json:"count"`
	Results  []PolicyResult `json:"results"`
}

// PolicyMetadata iam policy metadata
type PolicyMetadata struct {
	System    string       `json:"system"`
	Action    ActionWithID `json:"action"`
	Timestamp int64        `json:"timestamp"`
}

// PolicyResult iam policy result
type PolicyResult struct {
	Version string        `json:"version"`
	ID      int64         `json:"id"`
	Subject PolicySubject `json:"subject"`
	// Expression *operator.Policy `json:"expression"`
	ExpiredAt int64 `json:"expired_at"`
}

// PolicySubject policy subject, which represents user or user group for now
type PolicySubject struct {
	Type string `json:"type"`
	ID   string `json:"id"`
	Name string `json:"name"`
}

// GetPolicyOption defines options to get policy
type GetPolicyOption AuthOptions

// GetPolicyResp define get policy response data.
type GetPolicyResp struct {
	BaseResponse `json:",inline"`
	Data         *operator.Policy `json:"data"`
}

// ListPolicyOptions defines options to list a user's policy.
type ListPolicyOptions struct {
	System    string     `json:"system"`
	Subject   Subject    `json:"subject"`
	Actions   []Action   `json:"actions"`
	Resources []Resource `json:"resources"`
}

// ListPolicyResp defines response data to list policy.
type ListPolicyResp struct {
	BaseResponse `json:",inline"`
	Data         []*ActionPolicy `json:"data"`
}

// GetPolicyByExtResOption defines options to get policy by external resource, eg. cmdb biz
type GetPolicyByExtResOption struct {
	AuthOptions
	ExtResources []ExtResource `json:"ext_resources"`
}

// ExtResource represents external resources for filtering
type ExtResource struct {
	System string   `json:"system"`
	Type   TypeID   `json:"type"`
	IDs    []string `json:"ids"`
}

// GetPolicyByExtResResp define get policy by external resource response data.
type GetPolicyByExtResResp struct {
	BaseResponse `json:",inline"`
	Data         *GetPolicyByExtResResult `json:"data"`
}

// GetPolicyByExtResResult  get a user's policy by external resource result.
type GetPolicyByExtResResult struct {
	Expression   *operator.Policy      `json:"expression"`
	ExtResources []ExtResourceInstance `json:"ext_resources"`
}

// ExtResourceData represents external resource response data
type ExtResourceData struct {
	System    string                `json:"system"`
	Type      TypeID                `json:"type"`
	Instances []ExtResourceInstance `json:"instances"`
}

// ExtResourceInstance represents external resource instance response data
type ExtResourceInstance struct {
	ID        string                 `json:"id,omitempty"`
	Attribute map[string]interface{} `json:"attribute,omitempty"`
}

// AuthorizeList Defines the list structure of authorized instance ids.
// If the permission type is unlimited, the "IsAny" field is true and the "IDS" is empty.
// Otherwise, the "IsAny" field is false and the "Ids" is the specific resource instance id.
type AuthorizeList struct {
	// Ids is the authorized resource id list.
	Ids []string `json:"ids"`
	// IsAny = true means the user have all the permissions to access the resources.
	IsAny bool `json:"isAny"`
}

// Decision describes authorize decision, have already been authorized(true) or not(false)
type Decision struct {
	Authorized bool `json:"authorized"`
}

// AuthOptions describes an item to be authorized
type AuthOptions struct {
	System    string     `json:"system"`
	Subject   Subject    `json:"subject"`
	Action    Action     `json:"action"`
	Resources []Resource `json:"resources"`
}

// Validate the auth options is valid or not.
func (a AuthOptions) Validate() error {
	if len(a.System) == 0 {
		return errors.New("system is empty")
	}

	if len(a.Subject.Type) == 0 {
		return errors.New("subject.type is empty")
	}

	if len(a.Subject.ID) == 0 {
		return errors.New("subject.id is empty")
	}

	if len(a.Action.ID) == 0 {
		return errors.New("action.id is empty")
	}

	return nil
}

// AuthBatchOptions describes resource  items to be authorized
type AuthBatchOptions struct {
	System  string       `json:"system"`
	Subject Subject      `json:"subject"`
	Batch   []*AuthBatch `json:"batch"`
}

// Validate auth batch options is valid or not.
func (a AuthBatchOptions) Validate() error {
	if len(a.System) == 0 {
		return errors.New("system is empty")
	}

	if len(a.Subject.Type) == 0 {
		return errors.New("subject.type is empty")
	}

	if len(a.Subject.ID) == 0 {
		return errors.New("subject.id is empty")
	}

	if len(a.Batch) == 0 {
		return nil
	}

	for _, b := range a.Batch {
		if len(b.Action.ID) == 0 {
			return errors.New("empty action id")
		}
	}
	return nil
}

// AuthBatch defines auth batch options
type AuthBatch struct {
	Action    Action     `json:"action"`
	Resources []Resource `json:"resources"`
}

// Subject defines authorize resource type and instance id.
type Subject struct {
	Type TypeID `json:"type"`
	ID   string `json:"id"`
}

// Action defines the use's action, which must correspond to the registered action ids in iam.
type Action struct {
	ID string `json:"id"`
}

// ActionPolicy defines policy for a action.
type ActionPolicy struct {
	Action Action           `json:"action"`
	Policy *operator.Policy `json:"condition"`
}

// ListWithAttributes defines options to list resource instances with attribute.
type ListWithAttributes struct {
	Operator operator.OpType `json:"op"`
	// resource instance id list, this list is not required, it also
	// one of the query filter with Operator.
	IDList       []string           `json:"ids"`
	AttrPolicies []*operator.Policy `json:"attr_policies"`
	Type         TypeID             `json:"type"`
}

// GrantResourceCreatorActionAncestor defines resource creator action ancestor.
type GrantResourceCreatorActionAncestor struct {
	System string `json:"system"`
	Type   TypeID `json:"type"`
	ID     string `json:"id"`
}

// GrantResourceCreatorActionOption defines options to grant resource creator action.
type GrantResourceCreatorActionOption struct {
	System    string                               `json:"system"`
	Type      TypeID                               `json:"type"`
	ID        string                               `json:"id"`
	Name      string                               `json:"name"`
	Creator   string                               `json:"creator"`
	Ancestors []GrantResourceCreatorActionAncestor `json:"ancestors"`
}
