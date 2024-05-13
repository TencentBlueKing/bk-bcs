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

package namespace

import (
	"fmt"

	bkiam "github.com/TencentBlueKing/iam-go-sdk"

	conf "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/config"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/iam/perm"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/constants"
)

// ResRequest xxx
type ResRequest struct {
	ProjectID string
	ClusterID string
	ResType   string
	Resource  string
}

// NewResRequest xxx
func NewResRequest(projectID, clusterID, resource string) *ResRequest {
	return &ResRequest{ProjectID: projectID, ClusterID: clusterID, ResType: perm.ResTypeNS, Resource: resource}
}

// MakeResources xxx
func (r *ResRequest) MakeResources(resIDs []string) []bkiam.ResourceNode {
	resources := []bkiam.ResourceNode{}
	for _, id := range resIDs {
		resources = append(resources, bkiam.ResourceNode{
			System:    conf.G.IAM.SystemID,
			Type:      r.ResType,
			ID:        id,
			Attribute: r.MakeAttribute(id),
		})
	}
	return resources
}

// MakeAttribute xxx
func (r *ResRequest) MakeAttribute(_ string) map[string]interface{} {
	attr := map[string]interface{}{
		"_bk_iam_path_": fmt.Sprintf("/project,%s/cluster,%s/", r.ProjectID, r.ClusterID),
	}
	if r.Resource != "" {
		attr[constants.AttrResourceType] = constants.GetResourceAttr(r.Resource)
	}
	return attr
}

// FormMap xxx
func (r *ResRequest) FormMap(m map[string]interface{}) perm.ResRequest {
	if projID, ok := m["ProjectID"]; ok {
		r.ProjectID = projID.(string)
	}
	if clusterID, ok := m["ClusterID"]; ok {
		r.ClusterID = clusterID.(string)
	}
	return r
}
