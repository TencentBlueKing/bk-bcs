/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * 	http://opensource.org/licenses/MIT
 *
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package perm

import (
	bkiam "github.com/TencentBlueKing/iam-go-sdk"

	conf "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/config"
)

// ActionResourcesRequest ...
type ActionResourcesRequest struct {
	ActionID    string
	ResType     string
	ResIDs      []string
	ParentChain []IAMRes
}

// ToAction ...
func (r *ActionResourcesRequest) ToAction() bkiam.ApplicationAction {
	parentChainNodes := r.getParentChainNodes()
	instances := []bkiam.ApplicationResourceInstance{}
	for _, resID := range r.ResIDs {
		inst := append(parentChainNodes[:], bkiam.ApplicationResourceNode{Type: r.ResType, ID: resID})
		instances = append(instances, inst)
	}
	types := []bkiam.ApplicationRelatedResourceType{
		{SystemID: conf.G.IAM.SystemID, Type: r.ResType, Instances: instances},
	}
	return bkiam.ApplicationAction{ID: r.ActionID, RelatedResourceTypes: types}
}

func (r *ActionResourcesRequest) getParentChainNodes() []bkiam.ApplicationResourceNode {
	nodes := []bkiam.ApplicationResourceNode{}
	for _, p := range r.ParentChain {
		nodes = append(nodes, bkiam.ApplicationResourceNode{Type: p.ResType, ID: p.ResID})
	}
	return nodes
}
