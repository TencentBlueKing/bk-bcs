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

package iam

import (
	"github.com/TencentBlueKing/iam-go-sdk"
)

// AttributeKey for attr key type
type AttributeKey string

const (
	// BkIAMPath bk_iam_path
	BkIAMPath AttributeKey = "_bk_iam_path_"
)

// PathBuildIAM interface for build IAMPath
type PathBuildIAM interface {
	BuildIAMPath() string
}

// ResourceNode build resource node
type ResourceNode struct {
	System    string
	RType     string
	RInstance string
	Rp        PathBuildIAM
	Attr      map[string]interface{}
}

// BuildResourceNode build iam resourceNode
func (rn ResourceNode) BuildResourceNode() iam.ResourceNode {
	attr := map[string]interface{}{}
	bkPath := rn.Rp.BuildIAMPath()

	if bkPath != "" {
		attr[string(BkIAMPath)] = bkPath
	}

	for k, v := range rn.Attr {
		attr[k] = v
	}

	resourceNode := iam.NewResourceNode(rn.System, rn.RType, rn.RInstance, attr)

	return resourceNode
}

// PermissionRequest xxx
type PermissionRequest struct {
	SystemID string
	UserName string
}

func (pr PermissionRequest) validate() bool { // nolint
	if pr.SystemID == "" || pr.UserName == "" {
		return false
	}

	return true
}

// MakeRequestWithoutResources make request for no resources
func (pr PermissionRequest) MakeRequestWithoutResources(actionID string) iam.Request {

	subject := iam.Subject{
		Type: "user",
		ID:   pr.UserName,
	}

	action := iam.Action{ID: actionID}
	return iam.NewRequest(pr.SystemID, subject, action, nil)
}

// MakeRequestWithResources make request for signal action signal resource
func (pr PermissionRequest) MakeRequestWithResources(actionID string, nodes []ResourceNode) iam.Request {
	subject := iam.Subject{
		Type: "user",
		ID:   pr.UserName,
	}

	action := iam.Action{ID: actionID}

	iamNodes := make([]iam.ResourceNode, 0)
	for i := range nodes {
		iamNodes = append(iamNodes, nodes[i].BuildResourceNode())
	}
	return iam.NewRequest(pr.SystemID, subject, action, iamNodes)
}

// MakeRequestMultiActionResources make request for multi actions and signal resource
func (pr PermissionRequest) MakeRequestMultiActionResources(actions []string,
	nodes []ResourceNode) iam.MultiActionRequest {
	subject := iam.Subject{
		Type: "user",
		ID:   pr.UserName,
	}

	multiAction := make([]iam.Action, 0)
	for i := range actions {
		multiAction = append(multiAction, iam.Action{ID: actions[i]})
	}

	iamNodes := make([]iam.ResourceNode, 0)
	for i := range nodes {
		iamNodes = append(iamNodes, nodes[i].BuildResourceNode())
	}

	return iam.NewMultiActionRequest(pr.SystemID, subject, multiAction, iamNodes)
}

// MakeReqMultiActionsWithoutRes make request for multi actions and no resource
func (pr PermissionRequest) MakeReqMultiActionsWithoutRes(actions []string) iam.MultiActionRequest {
	subject := iam.Subject{
		Type: "user",
		ID:   pr.UserName,
	}

	multiAction := make([]iam.Action, 0)
	for i := range actions {
		multiAction = append(multiAction, iam.Action{ID: actions[i]})
	}

	return iam.NewMultiActionRequest(pr.SystemID, subject, multiAction, nil)
}

// RelatedResourceNode xxx
type RelatedResourceNode struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

// RelatedResourceLevel related resource level
type RelatedResourceLevel struct {
	Nodes []RelatedResourceNode
}

// BuildInstance build application resource instance
func (rrl RelatedResourceLevel) BuildInstance() iam.ApplicationResourceInstance {
	nodeList := make([]iam.ApplicationResourceNode, 0)

	for i := range rrl.Nodes {
		nodeList = append(nodeList, iam.ApplicationResourceNode{
			Type: rrl.Nodes[i].Type,
			ID:   rrl.Nodes[i].ID,
		})
	}
	return nodeList
}

// RelatedResourceType xxx
type RelatedResourceType struct {
	SystemID string // systemID
	RType    string // system resource type
}

// BuildRelatedResource build level instances
func (rrt RelatedResourceType) BuildRelatedResource(
	instances []iam.ApplicationResourceInstance) iam.ApplicationRelatedResourceType {
	return iam.ApplicationRelatedResourceType{
		SystemID:  rrt.SystemID,
		Type:      rrt.RType,
		Instances: instances,
	}
}

// ApplicationRequest xxx
type ApplicationRequest struct {
	SystemID string
}

// BuildApplication build application
func (ar ApplicationRequest) BuildApplication(relatedResources []ApplicationAction) iam.Application {
	actions := make([]iam.ApplicationAction, 0)
	for i := range relatedResources {
		actions = append(actions, iam.ApplicationAction{
			ID:                   relatedResources[i].ActionID,
			RelatedResourceTypes: relatedResources[i].RelatedResources,
		})
	}

	return iam.NewApplication(ar.SystemID, actions)
}

// ApplicationAction application action
type ApplicationAction struct {
	ActionID         string
	RelatedResources []iam.ApplicationRelatedResourceType
}

// BuildRelatedResourceTypes : instanceList for same resourceType resource
func BuildRelatedResourceTypes(systemID, resourceType string,
	instanceList []iam.ApplicationResourceInstance) iam.ApplicationRelatedResourceType {
	return iam.ApplicationRelatedResourceType{
		SystemID:  systemID,
		Type:      resourceType,
		Instances: instanceList,
	}
}

// Instance for instance resource level,
// for example: {"type": "project", "id": "123456"}, {"type": "cluster", "id": "BCS-K8S-25000"}
type Instance struct {
	ResourceType string
	ResourceID   string
}

// BuildResourceInstance generate ApplicationResourceInstance
func BuildResourceInstance(instances []Instance) iam.ApplicationResourceInstance {
	ari := iam.ApplicationResourceInstance{}

	for i := range instances {
		ari = append(ari, iam.ApplicationResourceNode{
			Type: instances[i].ResourceType,
			ID:   instances[i].ResourceID,
		})
	}

	return ari
}

// GradeManagerRequest grade manager request
type GradeManagerRequest struct {
	System              string               `json:"system"`
	Name                string               `json:"name"`
	Description         string               `json:"description"`
	Members             []string             `json:"members"`
	AuthorizationScopes []AuthorizationScope `json:"authorization_scopes"`
	SubjectScopes       []iam.Subject        `json:"subject_scopes"`
}

// LevelResource level resource
type LevelResource struct {
	Type string `json:"type"`
	ID   string `json:"id"`
	Name string `json:"name"`
}

// BuildAuthorizationScope 同一实例视图资源授权范围
func BuildAuthorizationScope(resourceType TypeID, actions []ActionID,
	resourceLevel []LevelResource) AuthorizationScope {
	iamActions := make([]iam.Action, 0)
	for i := range actions {
		iamActions = append(iamActions, iam.Action{ID: string(actions[i])})
	}

	paths := make([]Path, 0)
	for i := range resourceLevel {
		paths = append(paths, Path{
			System: SystemIDBKBCS,
			Type:   resourceLevel[i].Type,
			ID:     resourceLevel[i].ID,
			Name:   resourceLevel[i].Name,
		})
	}

	// "resources\": [\"This field may not be null.\"]
	return AuthorizationScope{
		System:  SystemIDBKBCS,
		Actions: iamActions,
		Resources: func() []ResourcePath {
			if len(paths) == 0 {
				return []ResourcePath{}
			}
			return []ResourcePath{
				{
					System: SystemIDBKBCS,
					RType:  string(resourceType),
					Paths:  [][]Path{paths},
				},
			}
		}(),
	}
}

// GlobalSubjectUser all global user
var GlobalSubjectUser = iam.Subject{
	Type: "*",
	ID:   "*",
}

// AuthorizationScope authorization scope
type AuthorizationScope struct {
	System    string         `json:"system"`
	Actions   []iam.Action   `json:"actions"`
	Resources []ResourcePath `json:"resources"`
}

// ResourcePath xxx
type ResourcePath struct {
	System string `json:"system"`
	RType  string `json:"type"`
	// Paths 批量资源拓扑，某种资源可能属于不同的拓扑
	Paths [][]Path `json:"paths"`
}

// Path 拓扑层级
type Path struct {
	System string `json:"system"`
	Type   string `json:"type"`
	ID     string `json:"id"`
	Name   string `json:"name"`
}

// GradeManagerResponse grade manager response
type GradeManagerResponse struct {
	BaseResponse
	Data GradeManagerID `json:"data"`
}

// GradeManagerID return ID
type GradeManagerID struct {
	ID uint64 `json:"id"`
}

// CreateUserGroupRequest create user group request
type CreateUserGroupRequest struct {
	Groups []UserGroup `json:"groups"`
}

// UserGroup xxx
type UserGroup struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// CreateUserGroupResponse create user group response
type CreateUserGroupResponse struct {
	BaseResponse
	Data []uint64 `json:"data"`
}

// UserType user type
type UserType string

// String to string
func (ut UserType) String() string {
	return string(ut)
}

var (
	// User user type
	User UserType = "user"
	// Department department type
	Department UserType = "department"
)

// BuildUserSubject build userType subject
func BuildUserSubject(name string) iam.Subject {
	return iam.Subject{
		Type: User.String(),
		ID:   name,
	}
}

// BuildDepartmentSubject build departmentType subject
func BuildDepartmentSubject(name string) iam.Subject {
	return iam.Subject{
		Type: Department.String(),
		ID:   name,
	}
}

// AddGroupMemberRequest add user group member request
type AddGroupMemberRequest struct {
	Members   []iam.Subject `json:"members"`
	ExpiredAt int           `json:"expired_at"`
}

// AddGroupMemberResponse add user group member response
type AddGroupMemberResponse struct {
	BaseResponse
	Data struct{} `json:"data"`
}

// DeleteGroupMemberRequest delete user group member request
type DeleteGroupMemberRequest struct {
	// Type: User or Department
	Type string `json:"type"`
	// IDs: users or departmentID
	IDs []string `json:"ids"`
}

// ResourceCreatorActionRequest request
type ResourceCreatorActionRequest struct {
	System       string     `json:"system"`
	ResourceType string     `json:"type"`
	ResourceID   string     `json:"id"`
	ResourceName string     `json:"name"`
	Creator      string     `json:"creator"`
	Ancestors    []Ancestor `json:"ancestors,omitempty"`
}

func buildResourceCreatorActionRequest(resource ResourceCreator, ancestors []Ancestor) *ResourceCreatorActionRequest {
	request := &ResourceCreatorActionRequest{
		System:       SystemIDBKBCS,
		ResourceType: resource.ResourceType,
		ResourceID:   resource.ResourceID,
		ResourceName: resource.ResourceName,
		Creator:      resource.Creator,
	}
	if len(ancestors) > 0 {
		request.Ancestors = ancestors
	}

	return request
}

// ResourceCreatorActionResponse response
type ResourceCreatorActionResponse struct {
	BaseResponse
	Data []ActionPolicy `json:"data"`
}

// ActionPolicy creator被授权对应的Action和策略ID列表
type ActionPolicy struct {
	Action struct {
		ID string `json:"id"`
	} `json:"action"`
	PolicyID int `json:"policy_id"`
}

// ResourceCreator resource info
type ResourceCreator struct {
	ResourceType string `json:"type"`
	ResourceID   string `json:"id"`
	ResourceName string `json:"name"`
	Creator      string `json:"creator"`
}

// Ancestor resource level parent
type Ancestor struct {
	System string `json:"system"`
	Type   string `json:"type"`
	ID     string `json:"id"`
}
