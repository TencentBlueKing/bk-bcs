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

// Package pbas provides auth server core protocol struct and convert functions.
package pbas

import (
	"bscp.io/pkg/iam/client"
	"bscp.io/pkg/iam/meta"
)

// UserInfo convert pb UserInfo to meta type UserInfo
func (m *UserInfo) UserInfo() *meta.UserInfo {
	if m == nil {
		return nil
	}

	return &meta.UserInfo{
		UserName: m.UserName,
	}
}

// PbUserInfo convert meta UserInfo to pb UserInfo
func PbUserInfo(m *meta.UserInfo) *UserInfo {
	if m == nil {
		return nil
	}

	return &UserInfo{
		UserName: m.UserName,
	}
}

// ResourceAttribute convert pb ResourceAttribute to meta type ResourceAttribute
func (m *ResourceAttribute) ResourceAttribute() *meta.ResourceAttribute {
	if m == nil {
		return nil
	}

	return &meta.ResourceAttribute{
		Basic: m.Basic.Basic(),
		BizID: m.BizId,
	}
}

// ResourceAttributes convert pb ResourceAttribute array to pb ResourceAttribute array
func ResourceAttributes(resourceAttributes []*ResourceAttribute) []*meta.ResourceAttribute {
	result := make([]*meta.ResourceAttribute, len(resourceAttributes))

	if len(resourceAttributes) == 0 {
		return result
	}

	for index, resourceAttribute := range resourceAttributes {
		result[index] = resourceAttribute.ResourceAttribute()
	}

	return result
}

// PbResourceAttribute convert meta ResourceAttribute to pb ResourceAttribute
func PbResourceAttribute(m *meta.ResourceAttribute) *ResourceAttribute {
	if m == nil {
		return nil
	}

	return &ResourceAttribute{
		Basic: PbBasic(m.Basic),
		BizId: m.BizID,
	}
}

// PbResourceAttributes convert meta ResourceAttribute array to pb ResourceAttribute array
func PbResourceAttributes(resourceAttributes []*meta.ResourceAttribute) []*ResourceAttribute {
	result := make([]*ResourceAttribute, len(resourceAttributes))

	if len(resourceAttributes) == 0 {
		return result
	}

	for index, resourceAttribute := range resourceAttributes {
		result[index] = PbResourceAttribute(resourceAttribute)
	}

	return result
}

// Basic convert pb Basic to meta type Basic
func (m *Basic) Basic() meta.Basic {
	return meta.Basic{
		Type:       meta.ResourceType(m.Type),
		Action:     meta.Action(m.Action),
		ResourceID: m.ResourceId,
	}
}

// PbBasic convert meta Basic to pb Basic
func PbBasic(m meta.Basic) *Basic {
	return &Basic{
		Type:       string(m.Type),
		Action:     string(m.Action),
		ResourceId: m.ResourceID,
	}
}

// BasicDetails convert pb BasicDetail array to pb BasicDetail array
func BasicDetails(perm *IamPermission) []meta.BasicDetail {
	result := make([]meta.BasicDetail, 0)

	for _, action := range perm.Actions {
		for _, resourceType := range action.RelatedResourceTypes {
			for _, instance := range resourceType.Instances {
				for _, i := range instance.Instances {
					result = append(result, meta.BasicDetail{
						TypeName:     resourceType.TypeName,
						ActionName:   action.Name,
						ResourceName: i.Name,
					})
				}
			}
		}
	}

	return result
}

// Decision convert pb Decision to meta type Decision
func (m *Decision) Decision() *meta.Decision {
	if m == nil {
		return nil
	}

	return &meta.Decision{
		Resource:   m.Resource.ResourceAttribute(),
		Authorized: m.Authorized,
	}
}

// Decisions convert pb Decision array to pb Decision array
func Decisions(decisions []*Decision) []*meta.Decision {
	result := make([]*meta.Decision, len(decisions))

	if len(decisions) == 0 {
		return result
	}

	for index, decision := range decisions {
		result[index] = decision.Decision()
	}

	return result
}

// PbDecision convert meta Decision to pb Decision
func PbDecision(m *meta.Decision) *Decision {
	if m == nil {
		return nil
	}

	return &Decision{
		Resource:   PbResourceAttribute(m.Resource),
		Authorized: m.Authorized,
	}
}

// PbDecisions convert meta Decision array to pb Decision array
func PbDecisions(decisions []*meta.Decision) []*Decision {
	result := make([]*Decision, len(decisions))

	if len(decisions) == 0 {
		return result
	}

	for index, decision := range decisions {
		result[index] = PbDecision(decision)
	}

	return result
}

// PbIamPermission convert meta IamPermission to pb IamPermission
func PbIamPermission(m *meta.IamPermission) *IamPermission {
	if m == nil {
		return nil
	}

	return &IamPermission{
		SystemId:   m.SystemID,
		SystemName: m.SystemName,
		Actions:    PbIamActions(m.Actions),
	}
}

// PbIamAction convert meta IamAction to pb IamAction
func PbIamAction(m *meta.IamAction) *IamAction {
	if m == nil {
		return nil
	}

	return &IamAction{
		Id:                   m.ID,
		Name:                 m.Name,
		RelatedResourceTypes: PbIamResourceTypes(m.RelatedResourceTypes),
	}
}

// PbIamActions convert meta IamAction array to pb IamAction array
func PbIamActions(actions []*meta.IamAction) []*IamAction {
	result := make([]*IamAction, len(actions))

	if len(actions) == 0 {
		return result
	}

	for index, action := range actions {
		result[index] = PbIamAction(action)
	}

	return result
}

// PbIamResourceType convert meta IamResourceType to pb IamResourceType
func PbIamResourceType(m *meta.IamResourceType) *IamResourceType {
	if m == nil {
		return nil
	}

	return &IamResourceType{
		SystemId:   m.SystemID,
		SystemName: m.SystemName,
		Type:       m.Type,
		TypeName:   m.TypeName,
		Instances:  PbIamResourceInstancesArr(m.Instances),
		Attributes: PbIamResourceAttributes(m.Attributes),
	}
}

// PbIamResourceTypes convert meta IamResourceType array to pb IamResourceType array
func PbIamResourceTypes(resourceTypes []*meta.IamResourceType) []*IamResourceType {
	result := make([]*IamResourceType, len(resourceTypes))

	if len(resourceTypes) == 0 {
		return result
	}

	for index, resType := range resourceTypes {
		result[index] = PbIamResourceType(resType)
	}

	return result
}

// PbIamResourceInstance convert meta IamResourceInstance to pb IamResourceInstance
func PbIamResourceInstance(m *meta.IamResourceInstance) *IamResourceInstance {
	if m == nil {
		return nil
	}

	return &IamResourceInstance{
		Type:     m.Type,
		TypeName: m.TypeName,
		Id:       m.ID,
		Name:     m.Name,
	}
}

// PbIamResourceInstancesArr convert meta IamResourceInstance 2-D array to pb IamResourceInstances array
func PbIamResourceInstancesArr(resourceInstancesArr [][]*meta.IamResourceInstance) []*IamResourceInstances {
	result := make([]*IamResourceInstances, len(resourceInstancesArr))

	if len(resourceInstancesArr) == 0 {
		return result
	}

	for index, resourceInstances := range resourceInstancesArr {
		instances := make([]*IamResourceInstance, len(resourceInstances))
		for idx, resourceInstance := range resourceInstances {
			instances[idx] = PbIamResourceInstance(resourceInstance)
		}
		result[index] = &IamResourceInstances{
			Instances: instances,
		}
	}

	return result
}

// PbIamResourceAttribute convert meta IamResourceAttribute to pb IamResourceAttribute
func PbIamResourceAttribute(m *meta.IamResourceAttribute) *IamResourceAttribute {
	if m == nil {
		return nil
	}

	return &IamResourceAttribute{
		Id:     m.ID,
		Values: PbIamResourceAttributeValues(m.Values),
	}
}

// PbIamResourceAttributes convert meta IamResourceAttribute array to pb IamResourceAttribute array
func PbIamResourceAttributes(attributes []*meta.IamResourceAttribute) []*IamResourceAttribute {
	result := make([]*IamResourceAttribute, len(attributes))

	if len(attributes) == 0 {
		return result
	}

	for index, attribute := range attributes {
		result[index] = PbIamResourceAttribute(attribute)
	}

	return result
}

// PbIamResourceAttributeValue convert meta IamResourceAttributeValue to pb IamResourceAttributeValue
func PbIamResourceAttributeValue(m *meta.IamResourceAttributeValue) *IamResourceAttributeValue {
	if m == nil {
		return nil
	}

	return &IamResourceAttributeValue{
		Id: m.ID,
	}
}

// PbIamResourceAttributeValues convert meta IamResourceAttributeValue array to pb IamResourceAttributeValue array
func PbIamResourceAttributeValues(values []*meta.IamResourceAttributeValue) []*IamResourceAttributeValue {
	result := make([]*IamResourceAttributeValue, len(values))

	if len(values) == 0 {
		return result
	}

	for index, value := range values {
		result[index] = PbIamResourceAttributeValue(value)
	}

	return result
}

// GrantResourceCreatorAction convert pb GrantResourceCreatorActionReq to client GrantResourceCreatorActionOption
func GrantResourceCreatorAction(req *GrantResourceCreatorActionReq) *client.GrantResourceCreatorActionOption {
	return &client.GrantResourceCreatorActionOption{
		System:    req.System,
		Type:      client.TypeID(req.Type),
		ID:        req.Id,
		Name:      req.Name,
		Creator:   req.Creator,
		Ancestors: GrantResourceCreatorActionAncetors(req.Ancestors),
	}
}

// GrantResourceCreatorActionAncetors convert pb GrantResourceCreatorActionReq_Ancestor array
// to client GrantResourceCreatorActionAncestor array
//
//nolint:lll
func GrantResourceCreatorActionAncetors(ancetors []*GrantResourceCreatorActionReq_Ancestor) []client.GrantResourceCreatorActionAncestor {
	result := make([]client.GrantResourceCreatorActionAncestor, len(ancetors))

	if len(ancetors) == 0 {
		return result
	}

	for index, ancetor := range ancetors {
		result[index] = GrantResourceCreatorActionAncetor(ancetor)
	}

	return result
}

// GrantResourceCreatorActionAncetor convert pb GrantResourceCreatorActionReq_Ancestor
// to client GrantResourceCreatorActionAncestor
//
//nolint:lll
func GrantResourceCreatorActionAncetor(ancetor *GrantResourceCreatorActionReq_Ancestor) client.GrantResourceCreatorActionAncestor {
	return client.GrantResourceCreatorActionAncestor{
		System: ancetor.System,
		Type:   client.TypeID(ancetor.Type),
		ID:     ancetor.Id,
	}
}

// PbGrantResourceCreatorActionAncestor convert client GrantResourceCreatorActionAncestor
// to pb GrantResourceCreatorActionReq_Ancestor
//
//nolint:lll
func PbGrantResourceCreatorActionAncestor(ancetor client.GrantResourceCreatorActionAncestor) *GrantResourceCreatorActionReq_Ancestor {
	return &GrantResourceCreatorActionReq_Ancestor{
		System: ancetor.System,
		Type:   string(ancetor.Type),
		Id:     ancetor.ID,
	}
}

// PbGrantResourceCreatorActionAncestors convert client GrantResourceCreatorActionAncestor array
// to pb GrantResourceCreatorActionReq_Ancestor array
//
//nolint:lll
func PbGrantResourceCreatorActionAncestors(ancetors []client.GrantResourceCreatorActionAncestor) []*GrantResourceCreatorActionReq_Ancestor {
	result := make([]*GrantResourceCreatorActionReq_Ancestor, len(ancetors))

	if len(ancetors) == 0 {
		return result
	}

	for index, ancetor := range ancetors {
		result[index] = PbGrantResourceCreatorActionAncestor(ancetor)
	}

	return result
}

// PbGrantResourceCreatorActionOption convert client GrantResourceCreatorActionOption
// to pb GrantResourceCreatorActionReq
//
//nolint:lll
func PbGrantResourceCreatorActionOption(option *client.GrantResourceCreatorActionOption) *GrantResourceCreatorActionReq {
	return &GrantResourceCreatorActionReq{
		System:    option.System,
		Type:      string(option.Type),
		Id:        option.ID,
		Name:      option.Name,
		Creator:   option.Creator,
		Ancestors: PbGrantResourceCreatorActionAncestors(option.Ancestors),
	}
}
