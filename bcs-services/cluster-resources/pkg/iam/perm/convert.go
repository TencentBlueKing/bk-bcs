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

package perm

import (
	bkiam "github.com/TencentBlueKing/iam-go-sdk"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
)

// staticIAMPath 避免 ResourceNode.Rp 为空时 BuildResourceNode panic。
type staticIAMPath string

// BuildIAMPath xxx
func (p staticIAMPath) BuildIAMPath() string {
	return string(p)
}

func convertResourceNode(n bkiam.ResourceNode) iam.ResourceNode {
	attr := n.Attribute
	if attr == nil {
		attr = map[string]interface{}{}
	}
	path := ""
	if v, ok := attr[string(iam.BkIAMPath)].(string); ok {
		path = v
	}
	return iam.ResourceNode{
		System:    n.System,
		RType:     n.Type,
		RInstance: n.ID,
		Rp:        staticIAMPath(path),
		Attr:      attr,
	}
}

func convertResourceNodes(nodes []bkiam.ResourceNode) []iam.ResourceNode {
	out := make([]iam.ResourceNode, 0, len(nodes))
	for _, n := range nodes {
		out = append(out, convertResourceNode(n))
	}
	return out
}

func convertApplicationAction(a bkiam.ApplicationAction) iam.ApplicationAction {
	return iam.ApplicationAction{
		ActionID:         a.ID,
		RelatedResources: a.RelatedResourceTypes,
	}
}

func convertApplicationActions(reqs []ActionResourcesRequest) []iam.ApplicationAction {
	out := make([]iam.ApplicationAction, 0, len(reqs))
	for i := range reqs {
		out = append(out, convertApplicationAction(reqs[i].ToAction()))
	}
	return out
}
