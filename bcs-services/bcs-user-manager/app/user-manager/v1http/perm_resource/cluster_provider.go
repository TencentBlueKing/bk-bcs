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

package perm_resource

import (
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/pkg/cmanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/pkg/iam"
)

// NewClusterResourceProvider cluster provider
func NewClusterResourceProvider(manager *cmanager.ClusterManagerClient, resourceType iam.TypeID) Provider {
	return &clusterResourceProvider{
		manager:      manager,
		resourceType: resourceType,
	}
}

type clusterResourceProvider struct {
	manager      *cmanager.ClusterManagerClient
	resourceType iam.TypeID
}

// ListAttr list resource attr
func (crp *clusterResourceProvider) ListAttr() ([]AttrResource, error) {
	if crp == nil {
		return nil, ErrServerNotInit
	}
	attrs := []AttrResource{}
	return attrs, nil
}

// ListAttrValue list resource attr value for certain attr
func (crp *clusterResourceProvider) ListAttrValue(filter *ListAttrValueFilter, page Page) (*ListAttrValueResult, error) {
	if crp == nil {
		return nil, ErrServerNotInit
	}
	attrValues := &ListAttrValueResult{}
	return attrValues, nil
}

// ListInstance for list cluster instance
func (crp *clusterResourceProvider) ListInstance(filter *ListInstanceFilter, page Page) (*ListInstanceResult, error) {
	if crp == nil {
		return nil, ErrServerNotInit
	}
	resources := make([]InstanceResource, 0)
	return &ListInstanceResult{
		Count:   0,
		Results: resources,
	}, nil
}

func (crp *clusterResourceProvider) SearchInstance(filter *SearchInstanceFilter, page Page) (*ListInstanceResult, error) {
	if crp == nil {
		return nil, ErrServerNotInit
	}

	resources := make([]InstanceResource, 0)
	return &ListInstanceResult{
		Count:   int64(0),
		Results: resources,
	}, nil
}

// FetchInstanceInfo Attrs only support "_bk_iam_path_"
func (crp *clusterResourceProvider) FetchInstanceInfo(filter *FetchInstanceInfoFilter) ([]map[string]interface{}, error) {
	if crp == nil {
		return nil, ErrServerNotInit
	}

	instanceInfo := []map[string]interface{}{}
	if len(filter.IDs) == 0 {
		return nil, fmt.Errorf("FetchInstanceInfoFilter instance IDs is null")
	}

	// get instance attr values
	for _, clusterID := range filter.IDs {
		instance := make(map[string]interface{})
		instance[idField] = clusterID
		// traverse attr values and set
		instance[IamPathKey] = []string{}

		instanceInfo = append(instanceInfo, instance)
	}

	return instanceInfo, nil
}

// ListInstanceByPolicy for list instance by policy
func (crp *clusterResourceProvider) ListInstanceByPolicy(filter *ListInstanceByPolicyFilter, page Page) (*ListInstanceResult, error) {
	if crp == nil {
		return nil, ErrServerNotInit
	}

	return &ListInstanceResult{
		Count:   0,
		Results: nil,
	}, nil
}

func (crp *clusterResourceProvider) getClusterAttrValues() []string {
	return []string{IamPathKey}
}
