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
}

// BuildResourceNode build iam resourceNode
func (rn ResourceNode) BuildResourceNode() iam.ResourceNode {
	attr := map[string]interface{}{}
	bkPath := rn.Rp.BuildIAMPath()

	if bkPath != "" {
		attr[string(BkIAMPath)] = bkPath
	}

	resourceNode := iam.NewResourceNode(rn.System, rn.RType, rn.RInstance, attr)

	return resourceNode
}

// PermissionRequest xxx
type PermissionRequest struct {
	SystemID string
	UserName string
}

func (pr PermissionRequest) validate() bool {
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
func (pr PermissionRequest) MakeRequestMultiActionResources(actions []string, nodes []ResourceNode) iam.MultiActionRequest {
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
func (rrt RelatedResourceType) BuildRelatedResource(instances []iam.ApplicationResourceInstance) iam.ApplicationRelatedResourceType {
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

// BuildRelatedResourceTypes: instanceList for same resourceType resource
func BuildRelatedResourceTypes(systemID, resourceType string, instanceList []iam.ApplicationResourceInstance) iam.ApplicationRelatedResourceType {
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
